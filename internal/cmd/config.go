package cmd

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/browser"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vbrowser configuration",
}

var configAuthCmd = &cobra.Command{
	Use:   "auth [on|off] [token]",
	Short: "Configure authentication",
	Long:  "Enable or disable password protection. If enabling, provide a token (password).",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status := args[0]
		if status == "on" {
			if len(args) < 2 {
				return fmt.Errorf("token is required when enabling auth")
			}
			cfg.Server.Auth.Enabled = true
			cfg.Server.Auth.Token = args[1]
			log.Info().Msg("Authentication enabled")
		} else if status == "off" {
			cfg.Server.Auth.Enabled = false
			log.Info().Msg("Authentication disabled")
		} else {
			return fmt.Errorf("invalid status: %s (must be 'on' or 'off')", status)
		}

		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		return nil
	},
}

var configPortCmd = &cobra.Command{
	Use:   "port [number]",
	Short: "Set server port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port number: %s", args[0])
		}

		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}

		cfg.Server.Port = port
		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		log.Info().Int("port", port).Msg("Server port updated")
		return nil
	},
}

var configBrowserCmd = &cobra.Command{
	Use:   "browser [chrome|chromium|firefox]",
	Short: "Set preferred browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browserType := args[0]
		if browserType != "chrome" && browserType != "chromium" && browserType != "firefox" {
			return fmt.Errorf("invalid browser type: %s (must be 'chrome', 'chromium', or 'firefox')", browserType)
		}

		path, err := browser.FindSystemBrowser(browserType)
		if err != nil {
			return fmt.Errorf("%w\n\n%s", err, browser.GetInstallInstructions())
		}

		cfg.Browser.BrowserPath = path
		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		log.Info().Str("type", browserType).Str("path", path).Msg("Browser preference updated")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAuthCmd)
	configCmd.AddCommand(configPortCmd)
	configCmd.AddCommand(configBrowserCmd)
}
