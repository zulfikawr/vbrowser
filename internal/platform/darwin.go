// Package platform provides platform-specific functionality for vbrowser.
//go:build darwin
// +build darwin

package platform

import (
	"fmt"
	"os/exec"
	"syscall"
)

// StartXvfb is a stub for macOS (not needed).
func StartXvfb(displayNum, width, height, depth int) (*exec.Cmd, error) {
	return nil, fmt.Errorf("Xvfb not required on macOS")
}

// SetDeathSig is a no-op on macOS as Pdeathsig is not supported.
func SetDeathSig(attr *syscall.SysProcAttr) {}
