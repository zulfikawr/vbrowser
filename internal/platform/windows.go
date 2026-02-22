// Package platform provides platform-specific functionality for vbrowser.
//go:build windows
// +build windows

package platform

import (
	"fmt"
	"os/exec"
)

// StartXvfb is a stub for Windows (not needed).
func StartXvfb(displayNum, width, height, depth int) (*exec.Cmd, error) {
	return nil, fmt.Errorf("Xvfb not required on Windows")
}
