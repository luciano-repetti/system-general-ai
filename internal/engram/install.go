package engram

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// InstallOptions controls where engram is installed and which version is fetched.
type InstallOptions struct {
	// TargetDir is the directory where the engram binary will be placed.
	// Conventional values:
	//   Linux/macOS: $HOME/.local/bin
	//   Windows:     %LOCALAPPDATA%\system-general-ai\bin
	TargetDir string
}

// Install downloads the appropriate engram binary for this platform and places it in TargetDir.
// Returns the absolute path to the installed binary.
func Install(ctx context.Context, opts InstallOptions) (string, error) {
	if opts.TargetDir == "" {
		return "", fmt.Errorf("TargetDir is required")
	}
	if err := os.MkdirAll(opts.TargetDir, 0o755); err != nil {
		return "", fmt.Errorf("create target dir: %w", err)
	}

	rel, err := FetchLatest(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch release: %w", err)
	}

	asset, err := rel.SelectAsset()
	if err != nil {
		return "", err
	}

	archivePath, err := downloadToTemp(ctx, asset)
	if err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}
	defer os.Remove(archivePath)

	binaryName := "engram"
	if runtime.GOOS == "windows" {
		binaryName = "engram.exe"
	}
	destPath := filepath.Join(opts.TargetDir, binaryName)

	if err := extractBinary(archivePath, asset.Name, destPath); err != nil {
		return "", fmt.Errorf("extract binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0o755); err != nil {
			return "", fmt.Errorf("chmod: %w", err)
		}
	}

	return destPath, nil
}

func downloadToTemp(ctx context.Context, asset *Asset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.DownloadURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "engram-*-"+asset.Name)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

func extractBinary(archivePath, archiveName, destPath string) error {
	switch {
	case strings.HasSuffix(archiveName, ".tar.gz"), strings.HasSuffix(archiveName, ".tgz"):
		return extractFromTarGz(archivePath, destPath)
	case strings.HasSuffix(archiveName, ".zip"):
		return extractFromZip(archivePath, destPath)
	default:
		// Asset is the binary itself (no compression).
		return copyFile(archivePath, destPath)
	}
}

func extractFromTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if !looksLikeEngramBinary(hdr.Name) {
			continue
		}
		return writeFile(destPath, tr)
	}
	return fmt.Errorf("engram binary not found in tar")
}

func extractFromZip(archivePath, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !looksLikeEngramBinary(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeFile(destPath, rc)
		rc.Close()
		return err
	}
	return fmt.Errorf("engram binary not found in zip")
}

func looksLikeEngramBinary(name string) bool {
	base := filepath.Base(name)
	lower := strings.ToLower(base)
	return lower == "engram" || lower == "engram.exe"
}

func writeFile(dest string, src io.Reader) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	return writeFile(dest, in)
}

