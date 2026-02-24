package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("vbrowser v0.1.0")
		fmt.Println("Built with Go and GStreamer")
		fmt.Println("https://github.com/zulfikawr/vbrowser")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
