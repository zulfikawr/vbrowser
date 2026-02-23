// Package platform provides platform-specific functionality for vbrowser.
//go:build linux
// +build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// StartXvfb starts an Xvfb virtual display server.
func StartXvfb(displayNum, width, height, depth int) (*exec.Cmd, error) {
	if _, err := exec.LookPath("Xvfb"); err != nil {
		return nil, fmt.Errorf("Xvfb not found. Install with: sudo apt-get install xvfb")
	}

	screen := fmt.Sprintf("%dx%dx%d", width, height, depth)
	args := []string{
		fmt.Sprintf(":%d", displayNum),
		"-screen", "0", screen,
		"-ac",
		"-nolisten", "tcp",
		"+extension", "RANDR",
	}

	log.Debug().Msgf("Starting Xvfb with args: %v", args)
	cmd := exec.Command("Xvfb", args...)

	// Xvfb is extremely noisy on stderr with non-fatal xkbcomp warnings.
	// We'll silence its stderr to keep our vbrowser logs clean.
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start Xvfb: %w", err)
	}

	// Wait for X11 socket to be ready
	socketPath := fmt.Sprintf("/tmp/.X11-unix/X%d", displayNum)
	ready := false
	for i := 0; i < 50; i++ { // Wait up to 5 seconds
		if _, err := os.Stat(socketPath); err == nil {
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !ready {
		log.Warn().Str("socket", socketPath).Msg("Xvfb socket did not appear in time")
	}

	log.Info().
		Int("display", displayNum).
		Str("resolution", screen).
		Int("pid", cmd.Process.Pid).
		Msg("Xvfb started")

	return cmd, nil
}
