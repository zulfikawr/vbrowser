// Package platform provides platform-specific functionality for vbrowser.
//go:build windows
// +build windows

package platform

import (
	"fmt"
	"os/exec"
	"syscall"
)

// StartXvfb is a stub for Windows (not needed).
func StartXvfb(displayNum, width, height, depth int) (*exec.Cmd, error) {
	return nil, fmt.Errorf("Xvfb not required on Windows")
}

// SetDeathSig is a no-op on Windows as Pdeathsig is not supported.
func SetDeathSig(attr *syscall.SysProcAttr) {}
