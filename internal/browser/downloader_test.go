package browser

import (
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform, err := DetectPlatform()
	if err != nil {
		t.Fatalf("DetectPlatform failed: %v", err)
	}

	switch runtime.GOOS {
	case "linux":
		if platform != "Linux_x64" {
			t.Errorf("expected Linux_x64, got %s", platform)
		}
	case "darwin":
		if platform != "Mac" && platform != "Mac_Arm" {
			t.Errorf("expected Mac or Mac_Arm, got %s", platform)
		}
	case "windows":
		if platform != "Win_x64" {
			t.Errorf("expected Win_x64, got %s", platform)
		}
	}
}

func TestGetZipName(t *testing.T) {
	tests := []struct {
		platform string
		expected string
	}{
		{"Linux_x64", "chrome-linux.zip"},
		{"Mac", "chrome-mac.zip"},
		{"Mac_Arm", "chrome-mac.zip"},
		{"Win_x64", "chrome-win.zip"},
	}

	for _, tt := range tests {
		result := getZipName(tt.platform)
		if result != tt.expected {
			t.Errorf("getZipName(%s) = %s, want %s", tt.platform, result, tt.expected)
		}
	}
}

func TestGetCachedRevision(t *testing.T) {
	tmpDir := t.TempDir()

	rev := GetCachedRevision(tmpDir)
	if rev != "" {
		t.Errorf("expected empty revision, got %s", rev)
	}
}
