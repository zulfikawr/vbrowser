// Package process provides process management utilities for vbrowser.
package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Write creates a PID file at the given path with the specified PID.
// Returns an error if the file already exists or cannot be created.
func Write(path string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create pid dir: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("pid file already exists: %s", path)
	}

	data := []byte(strconv.Itoa(pid) + "\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}

	return nil
}

// Read reads the PID from the file at the given path.
func Read(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read pid file: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("parse pid: %w", err)
	}

	return pid, nil
}

// Remove deletes the PID file at the given path.
func Remove(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove pid file: %w", err)
	}
	return nil
}

// IsRunning checks if a process with the given PID is running.
func IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
