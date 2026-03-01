package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vbrowser v%s\n", Version)
		fmt.Println("Built with Go and GStreamer")
		fmt.Print("https://github.com/zulfikawr/vbrowser\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
