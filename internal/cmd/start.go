package cmd

import (
	"github.com/spf13/cobra"
)

var (
	foreground bool
	port       int
	noDownload bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the vbrowser daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if port > 0 {
			cfg.Server.Port = port
		}

		// TODO: Implement daemon start logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "run in foreground")
	startCmd.Flags().IntVar(&port, "port", 0, "override server port")
	startCmd.Flags().BoolVar(&noDownload, "no-download", false, "skip Chromium auto-download")
}
