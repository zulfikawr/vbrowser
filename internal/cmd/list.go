package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/browser"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all detected browser binaries",
	Run: func(cmd *cobra.Command, args []string) {
		browsers := browser.FindAllBrowsers()
		if len(browsers) == 0 {
			fmt.Println("No browsers detected.")
			fmt.Println(browser.GetInstallInstructions())
			return
		}

		fmt.Println("Detected browsers:")
		for _, b := range browsers {
			if b == cfg.Browser.BrowserPath {
				fmt.Printf("* %s (currently used)\n", b)
			} else {
				fmt.Printf("  %s\n", b)
			}
		}
		fmt.Println("\nTo switch browser, use: vbrowser config browser <type>")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
