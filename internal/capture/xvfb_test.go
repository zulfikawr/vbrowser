//go:build linux
// +build linux

package capture

import (
	"testing"
)

func TestNewXvfbCapturer(t *testing.T) {
	capturer, err := NewXvfbCapturer(99, 1920, 1080)
	if err != nil {
		t.Fatalf("NewXvfbCapturer failed: %v", err)
	}

	if capturer == nil {
		t.Fatal("capturer is nil")
	}

	if capturer.displayNum != 99 {
		t.Errorf("expected display 99, got %d", capturer.displayNum)
	}

	if capturer.width != 1920 {
		t.Errorf("expected width 1920, got %d", capturer.width)
	}

	if capturer.height != 1080 {
		t.Errorf("expected height 1080, got %d", capturer.height)
	}
}

func TestClose(t *testing.T) {
	capturer, err := NewXvfbCapturer(99, 1920, 1080)
	if err != nil {
		t.Fatalf("NewXvfbCapturer failed: %v", err)
	}

	if err := capturer.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
