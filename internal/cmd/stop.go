package cmd

import (
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running vbrowser daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement daemon stop logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
