// Package cmd provides the command-line interface for vbrowser.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/config"
)

var (
	cfgFile  string
	logLevel string
	cfg      *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "vbrowser",
	Short: "Self-hosted virtual browser",
	Long:  "A single Go binary that launches a browser instance (Chrome or Chromium) and streams it via WebRTC",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		setupLogging()

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if logLevel != "" {
			cfg.Logging.Level = logLevel
		}

		cfg.ApplyEnv()

		level, err := zerolog.ParseLevel(cfg.Logging.Level)
		if err != nil {
			level = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(level)

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultConfig := filepath.Join(homeDir, ".config/vbrowser/config.json")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfig, "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error)")
}

func setupLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
