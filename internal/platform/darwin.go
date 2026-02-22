// Package platform provides platform-specific functionality for vbrowser.
//go:build darwin
// +build darwin

package platform

import (
	"fmt"
	"os/exec"
)

// StartXvfb is a stub for macOS (not needed).
func StartXvfb(displayNum, width, height, depth int) (*exec.Cmd, error) {
	return nil, fmt.Errorf("Xvfb not required on macOS")
}
