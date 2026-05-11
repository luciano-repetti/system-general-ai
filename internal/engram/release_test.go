package engram

import (
	"runtime"
	"testing"
)

func TestSelectAsset_PicksByOSArch(t *testing.T) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	rel := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "engram_v1.0.0_linux_amd64.tar.gz"},
			{Name: "engram_v1.0.0_linux_arm64.tar.gz"},
			{Name: "engram_v1.0.0_darwin_amd64.tar.gz"},
			{Name: "engram_v1.0.0_darwin_arm64.tar.gz"},
			{Name: "engram_v1.0.0_windows_amd64.zip"},
			{Name: "engram_v1.0.0_windows_arm64.zip"},
		},
	}

	got, err := rel.SelectAsset()
	if err != nil {
		t.Fatalf("SelectAsset: %v", err)
	}

	wantSubstrs := []string{osName, archName}
	for _, want := range wantSubstrs {
		if !contains(got.Name, want) {
			t.Errorf("asset %q does not contain %q", got.Name, want)
		}
	}
}

func TestSelectAsset_FailsWhenNoMatch(t *testing.T) {
	rel := &Release{
		TagName: "v1.0.0",
		Assets:  []Asset{{Name: "engram_v1.0.0_solaris_sparc.tar.gz"}},
	}
	_, err := rel.SelectAsset()
	if err == nil {
		t.Fatal("expected error for unsupported platform, got nil")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
