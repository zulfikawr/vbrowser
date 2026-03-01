package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running vbrowser daemon and clean up resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 0. Try to stop systemd service if it exists and is running
		_ = exec.Command("systemctl", "--user", "stop", "vbrowser.service").Run()

		// 1. Try to stop via PID file first (graceful)
		pid, err := process.Read(cfg.PIDFile)
		if err == nil && process.IsRunning(pid) {
			log.Info().Int("pid", pid).Msg("stopping vbrowser daemon")
			if proc, err := os.FindProcess(pid); err == nil {
				_ = proc.Signal(syscall.SIGTERM)

				// Wait for process to exit
				exited := false
				for i := 0; i < 50; i++ { // Wait up to 5 seconds
					if !process.IsRunning(pid) {
						exited = true
						break
					}
					time.Sleep(100 * time.Millisecond)
				}

				if !exited {
					log.Warn().Int("pid", pid).Msg("vbrowser daemon did not exit gracefully, force killing")
					_ = proc.Kill()
				}
			}
		}

		// 2. Force kill any remaining vbrowser, browser or Xvfb processes
		log.Info().Msg("cleaning up orphaned processes...")
		currentPid := os.Getpid()
		// Kill all vbrowser processes except the current one
		_ = exec.Command("pkill", "-9", "--exclude-pids", fmt.Sprintf("%d", currentPid), "-f", "vbrowser").Run()
		// If current process is also named vbrowser, pkill might have killed us if not careful
		// But usually it's fine as pkill matches the command line.

		_ = exec.Command("pkill", "-9", "-f", "chrome").Run()
		_ = exec.Command("pkill", "-9", "-f", "chromium").Run()
		_ = exec.Command("pkill", "-9", "-f", "firefox").Run()
		_ = exec.Command("pkill", "-9", "-f", "Xvfb").Run()

		_ = currentPid // avoid unused var if needed

		// 3. Clean up stale X11 lock files
		displayNum := cfg.Display.DisplayNum
		lockFile := fmt.Sprintf("/tmp/.X%d-lock", displayNum)
		unixSocket := fmt.Sprintf("/tmp/.X11-unix/X%d", displayNum)

		if _, err := os.Stat(lockFile); err == nil {
			log.Info().Str("path", lockFile).Msg("removing stale X11 lock file")
			os.Remove(lockFile)
		}

		if _, err := os.Stat(unixSocket); err == nil {
			log.Info().Str("path", unixSocket).Msg("removing stale X11 unix socket")
			os.Remove(unixSocket)
		}

		// 4. Remove PID file unconditionally
		_ = process.Remove(cfg.PIDFile)

		log.Info().Msg("vbrowser cleanup complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
