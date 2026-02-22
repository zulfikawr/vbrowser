package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/browser"
	"github.com/zulfikawr/vbrowser/internal/process"
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

		// Check if already running
		if pid, err := process.Read(cfg.PIDFile); err == nil {
			if process.IsRunning(pid) {
				return fmt.Errorf("vbrowser is already running (PID %d)", pid)
			}
			log.Warn().Int("pid", pid).Msg("stale pid file found, removing")
			if err := process.Remove(cfg.PIDFile); err != nil {
				return fmt.Errorf("remove stale pid file: %w", err)
			}
		}

		// Handle Chromium download
		chromiumPath := cfg.Browser.ChromiumPath
		if chromiumPath == "" && cfg.Browser.AutoDownload && !noDownload {
			if err := ensureChromium(); err != nil {
				return fmt.Errorf("ensure chromium: %w", err)
			}
			var err error
			chromiumPath, err = browser.GetChromiumPath(cfg.Browser.DownloadDir)
			if err != nil {
				return fmt.Errorf("get chromium path: %w", err)
			}
		}

		if chromiumPath == "" {
			return fmt.Errorf("chromium path not configured and auto-download disabled")
		}

		log.Info().Str("chromium", chromiumPath).Msg("using Chromium")

		// Write PID file
		currentPid := os.Getpid()
		if err := process.Write(cfg.PIDFile, currentPid); err != nil {
			return fmt.Errorf("write pid file: %w", err)
		}

		// Start browser manager
		mgr := browser.NewManager(cfg)
		if err := mgr.Start(chromiumPath); err != nil {
			if err := process.Remove(cfg.PIDFile); err != nil {
				log.Warn().Err(err).Msg("failed to remove pid file")
			}
			return fmt.Errorf("start browser: %w", err)
		}

		log.Info().
			Int("pid", currentPid).
			Str("url", fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)).
			Msg("vbrowser started")

		// TODO: Start HTTP server

		// Handle shutdown signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info().Msg("shutting down")
		if err := mgr.Stop(); err != nil {
			log.Warn().Err(err).Msg("failed to stop manager")
		}
		if err := process.Remove(cfg.PIDFile); err != nil {
			log.Warn().Err(err).Msg("failed to remove pid file")
		}

		return nil
	},
}

func ensureChromium() error {
	platform, err := browser.DetectPlatform()
	if err != nil {
		return err
	}

	cachedRev := browser.GetCachedRevision(cfg.Browser.DownloadDir)
	if cachedRev != "" {
		log.Info().Str("revision", cachedRev).Msg("Chromium already downloaded")
		return nil
	}

	log.Info().Msg("fetching latest Chromium revision")
	revision, err := browser.FetchLatestRevision(platform)
	if err != nil {
		return err
	}

	log.Info().Str("revision", revision).Msg("downloading Chromium")
	if err := browser.Download(platform, revision, cfg.Browser.DownloadDir); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "run in foreground")
	startCmd.Flags().IntVar(&port, "port", 0, "override server port")
	startCmd.Flags().BoolVar(&noDownload, "no-download", false, "skip Chromium auto-download")
}
