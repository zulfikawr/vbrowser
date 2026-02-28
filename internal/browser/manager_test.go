package browser

import (
	"testing"

	"github.com/zulfikawr/vbrowser/internal/config"
)

func TestNewManager(t *testing.T) {
	cfg := config.Defaults()
	mgr := NewManager(cfg)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.cfg != cfg {
		t.Error("config not set correctly")
	}
}

func TestBuildBrowserArgs(t *testing.T) {
	cfg := config.Defaults()
	cfg.Browser.ExtraArgs = []string{"--test-flag"}
	mgr := NewManager(cfg)

	args := mgr.buildBrowserArgs()

	if len(args) == 0 {
		t.Fatal("no args generated")
	}

	found := false
	for _, arg := range args {
		if arg == "--remote-debugging-port=9222" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected --remote-debugging-port=9222 in args")
	}

	foundExtra := false
	for _, arg := range args {
		if arg == "--test-flag" {
			foundExtra = true
			break
		}
	}

	if !foundExtra {
		t.Error("extra args not included")
	}
}

func TestPid(t *testing.T) {
	cfg := config.Defaults()
	mgr := NewManager(cfg)

	if mgr.Pid() != 0 {
		t.Error("expected pid 0 before start")
	}
}

func TestIsRunning(t *testing.T) {
	cfg := config.Defaults()
	mgr := NewManager(cfg)

	if mgr.IsRunning() {
		t.Error("should not be running before start")
	}
}
