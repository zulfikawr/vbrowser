package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	pid := 12345
	if err := Write(pidFile, pid); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	readPid, err := Read(pidFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if readPid != pid {
		t.Errorf("expected pid %d, got %d", pid, readPid)
	}
}

func TestWriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	if err := Write(pidFile, 12345); err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	err := Write(pidFile, 67890)
	if err == nil {
		t.Error("expected error when writing to existing pid file")
	}
}

func TestReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "nonexistent.pid")

	_, err := Read(pidFile)
	if err == nil {
		t.Error("expected error when reading non-existent file")
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	if err := Write(pidFile, 12345); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := Remove(pidFile); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("pid file should not exist after removal")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "nonexistent.pid")

	if err := Remove(pidFile); err != nil {
		t.Errorf("Remove should not error on non-existent file: %v", err)
	}
}

func TestIsRunning(t *testing.T) {
	currentPid := os.Getpid()
	if !IsRunning(currentPid) {
		t.Error("current process should be running")
	}

	if IsRunning(0) {
		t.Error("pid 0 should not be running")
	}

	if IsRunning(-1) {
		t.Error("negative pid should not be running")
	}

	if IsRunning(999999) {
		t.Error("non-existent pid should not be running")
	}
}
