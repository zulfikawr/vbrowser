package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running vbrowser daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := process.Read(cfg.PIDFile)
		if err != nil {
			return fmt.Errorf("vbrowser is not running (no pid file)")
		}

		if !process.IsRunning(pid) {
			log.Warn().Int("pid", pid).Msg("process not running, removing stale pid file")
			if err := process.Remove(cfg.PIDFile); err != nil {
				return fmt.Errorf("remove pid file: %w", err)
			}
			return fmt.Errorf("vbrowser is not running")
		}

		log.Info().Int("pid", pid).Msg("stopping vbrowser")

		proc, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("find process: %w", err)
		}

		if err := proc.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("send SIGTERM: %w", err)
		}

		if err := process.Remove(cfg.PIDFile); err != nil {
			return fmt.Errorf("remove pid file: %w", err)
		}

		log.Info().Msg("vbrowser stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
