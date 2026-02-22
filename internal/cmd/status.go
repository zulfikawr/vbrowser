package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status and stream info",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement status logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
