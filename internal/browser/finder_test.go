package browser

import (
	"os"
	"testing"
)

func TestFindSystemBrowser(t *testing.T) {
	// Test auto-discovery
	path, err := FindSystemBrowser("auto")
	if err != nil {
		t.Logf("No browser found on this system: %v", err)
	} else {
		if path == "" {
			t.Error("FindSystemBrowser('auto') returned empty path without error")
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("FindSystemBrowser('auto') returned non-existent path: %s", path)
		}
	}

	// Test specific browsers (may fail if not installed, which is OK for a unit test on arbitrary systems)
	_, _ = FindSystemBrowser("chrome")
	_, _ = FindSystemBrowser("chromium")
}

func TestGetInstallInstructions(t *testing.T) {
	instr := GetInstallInstructions()
	if instr == "" {
		t.Error("GetInstallInstructions returned empty string")
	}
}
