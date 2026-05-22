package engram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// releasesAPI points at the upstream Engram binary releases.
// Engram is the persistent-memory MCP server we install and configure.
// We don't fork or vendor it — we download the latest release artifact at install time.
const releasesAPI = "https://api.github.com/repos/Gentleman-Programming/engram/releases/latest"

// Release is the subset of the GitHub Releases payload we care about.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a single downloadable file attached to a release.
type Asset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

// FetchLatest queries GitHub's API for the latest engram release.
func FetchLatest(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "system-general-ai")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &rel, nil
}

// SelectAsset picks the asset matching the current OS/arch combination.
func (r *Release) SelectAsset() (*Asset, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	for i := range r.Assets {
		if matchesPlatform(r.Assets[i].Name, osName, archName) {
			return &r.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no engram asset for %s/%s in release %s", osName, archName, r.TagName)
}

func matchesPlatform(assetName, os, arch string) bool {
	name := strings.ToLower(assetName)
	if !strings.Contains(name, strings.ToLower(os)) {
		return false
	}
	if !strings.Contains(name, strings.ToLower(arch)) {
		return false
	}
	return true
}
