package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show live vbrowser service logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := exec.Command("journalctl", "--user", "-u", "vbrowser.service", "-f", "-n", "100")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
