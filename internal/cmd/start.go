package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/vbrowser/internal/browser"
	"github.com/zulfikawr/vbrowser/internal/process"
	"github.com/zulfikawr/vbrowser/pkg/server"
)

var (
	foreground bool
	port       int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the vbrowser daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !foreground {
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

			// Ensure log file is set if not provided
			if cfg.Logging.File == "" {
				homeDir, _ := os.UserHomeDir()
				cfg.Logging.File = filepath.Join(homeDir, ".local/share/vbrowser/vbrowser.log")
				if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0755); err != nil {
					return fmt.Errorf("create log directory: %w", err)
				}
			}

			// Prepare command to run in background
			executable, err := os.Executable()
			if err != nil {
				return fmt.Errorf("find executable: %w", err)
			}

			newArgs := []string{}
			for _, arg := range os.Args[1:] {
				if arg != "start" && arg != "-f" && arg != "--foreground" {
					newArgs = append(newArgs, arg)
				}
			}
			newArgs = append([]string{"start", "--foreground"}, newArgs...)

			daemon := exec.Command(executable, newArgs...)
			daemon.Stdout = nil
			daemon.Stderr = nil
			daemon.Stdin = nil
			daemon.Env = append(os.Environ(), fmt.Sprintf("VBROWSER_LOG_FILE=%s", cfg.Logging.File))

			if err := daemon.Start(); err != nil {
				return fmt.Errorf("daemonize: %w", err)
			}

			fmt.Printf("vbrowser starting in background (PID %d)\n", daemon.Process.Pid)
			fmt.Printf("Logs: %s\n", cfg.Logging.File)
			return nil
		}

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

		// Find Browser
		browserPath := cfg.Browser.BrowserPath
		if browserPath == "" {
			var err error
			browserPath, err = browser.FindSystemBrowser("auto")
			if err != nil {
				return fmt.Errorf("no browser found\n\n%s\n\nAlternatively, specify the path in config or VBROWSER_BROWSER_PATH env var\n", browser.GetInstallInstructions())
			}
		}

		log.Info().Str("browser", browserPath).Msg("using browser")

		// Log active configuration
		log.Info().
			Int("width", cfg.Browser.WindowWidth).
			Int("height", cfg.Browser.WindowHeight).
			Int("fps", cfg.Stream.TargetFPS).
			Int("bitrate_kbps", cfg.Stream.MaxBitrateKbps).
			Int("display", cfg.Display.DisplayNum).
			Msg("active configuration")

		// Write PID file
		currentPid := os.Getpid()
		if err := process.Write(cfg.PIDFile, currentPid); err != nil {
			return fmt.Errorf("write pid file: %w", err)
		}

		// Start browser manager
		mgr := browser.NewManager(cfg)
		if err := mgr.Start(browserPath); err != nil {
			if err := process.Remove(cfg.PIDFile); err != nil {
				log.Warn().Err(err).Msg("failed to remove pid file")
			}
			return fmt.Errorf("start browser: %w", err)
		}

		log.Info().
			Int("pid", currentPid).
			Str("url", fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)).
			Msg("vbrowser started")

		// Start HTTP server
		srv := server.New(cfg, mgr, cfgFile)
		if err := srv.Start(); err != nil {
			if err := mgr.Stop(); err != nil {
				log.Warn().Err(err).Msg("failed to stop manager")
			}
			if err := process.Remove(cfg.PIDFile); err != nil {
				log.Warn().Err(err).Msg("failed to remove pid file")
			}
			return fmt.Errorf("start server: %w", err)
		}

		// Handle shutdown signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info().Msg("shutting down")

		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Stop(ctx); err != nil {
			log.Warn().Err(err).Msg("failed to stop server")
		}
		if err := mgr.Stop(); err != nil {
			log.Warn().Err(err).Msg("failed to stop manager")
		}
		if err := process.Remove(cfg.PIDFile); err != nil {
			log.Warn().Err(err).Msg("failed to remove pid file")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "run in foreground")
	startCmd.Flags().IntVar(&port, "port", 0, "override server port")
}
