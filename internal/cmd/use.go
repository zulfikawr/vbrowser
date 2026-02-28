package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/browser"
)

var useCmd = &cobra.Command{
	Use:   "use [chrome|chromium]",
	Short: "Switch between Chrome and Chromium",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d\n", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		browserType := args[0]
		if browserType != "chrome" && browserType != "chromium" {
			return fmt.Errorf("invalid browser type: %s (must be 'chrome' or 'chromium')\n", browserType)
		}

		path, err := browser.FindSystemBrowser(browserType)
		if err != nil {
			return fmt.Errorf("%w\n\n%s\n", err, browser.GetInstallInstructions())
		}

		log.Info().Str("type", browserType).Str("path", path).Msg("Found browser")

		cfg.Browser.BrowserPath = path
		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		log.Info().Str("config", cfgFile).Msg("Browser preference saved")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
