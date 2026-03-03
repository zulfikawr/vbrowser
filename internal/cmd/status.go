package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/process"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status and stream info",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := process.Read(cfg.PIDFile)
		if err != nil {
			fmt.Println("● vbrowser is not running")
			return nil
		}

		if !process.IsRunning(pid) {
			fmt.Println("● vbrowser is not running (stale pid file)")
			return nil
		}

		fmt.Printf("● vbrowser is running\n")
		fmt.Printf("  PID:        %d\n", pid)
		fmt.Printf("  URL:        http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("  PID file:   %s\n", cfg.PIDFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
