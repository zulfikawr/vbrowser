package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/internal/platform"
)

// Manager manages the lifecycle of Chromium and Xvfb processes.
type Manager struct {
	cfg         *config.Config
	chromiumCmd *exec.Cmd
	xvfbCmd     *exec.Cmd
	chromiumPid int
}

// NewManager creates a new browser manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

// Start launches Xvfb (on Linux) and Chromium.
func (m *Manager) Start(chromiumPath string) error {
	// Start Xvfb on Linux
	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		xvfb, err := platform.StartXvfb(
			m.cfg.Display.DisplayNum,
			m.cfg.Browser.WindowWidth,
			m.cfg.Browser.WindowHeight,
			m.cfg.Display.Depth,
		)
		if err != nil {
			return fmt.Errorf("start xvfb: %w", err)
		}
		m.xvfbCmd = xvfb
	}

	// Build Chromium arguments
	args := m.buildChromiumArgs()

	// Start Chromium
	m.chromiumCmd = exec.Command(chromiumPath, args...)
	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		m.chromiumCmd.Env = append(os.Environ(),
			fmt.Sprintf("DISPLAY=:%d", m.cfg.Display.DisplayNum),
		)
	}

	if err := m.chromiumCmd.Start(); err != nil {
		m.stopXvfb()
		return fmt.Errorf("start chromium: %w", err)
	}

	m.chromiumPid = m.chromiumCmd.Process.Pid

	log.Info().
		Int("pid", m.chromiumPid).
		Str("path", chromiumPath).
		Msg("Chromium started")

	return nil
}

// Stop gracefully terminates Chromium and Xvfb.
func (m *Manager) Stop() error {
	if m.chromiumCmd != nil && m.chromiumCmd.Process != nil {
		log.Info().Int("pid", m.chromiumPid).Msg("stopping Chromium")
		if err := m.chromiumCmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Warn().Err(err).Msg("failed to send SIGTERM to Chromium")
		}
		if err := m.chromiumCmd.Wait(); err != nil {
			log.Debug().Err(err).Msg("chromium wait error")
		}
	}

	m.stopXvfb()

	return nil
}

// Pid returns the Chromium process PID.
func (m *Manager) Pid() int {
	return m.chromiumPid
}

// IsRunning checks if Chromium is still running.
func (m *Manager) IsRunning() bool {
	if m.chromiumCmd == nil || m.chromiumCmd.Process == nil {
		return false
	}
	err := m.chromiumCmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func (m *Manager) stopXvfb() {
	if m.xvfbCmd != nil && m.xvfbCmd.Process != nil {
		log.Info().Int("pid", m.xvfbCmd.Process.Pid).Msg("stopping Xvfb")
		if err := m.xvfbCmd.Process.Kill(); err != nil {
			log.Warn().Err(err).Msg("failed to kill Xvfb")
		}
		if err := m.xvfbCmd.Wait(); err != nil {
			log.Debug().Err(err).Msg("xvfb wait error")
		}
	}
}

func (m *Manager) buildChromiumArgs() []string {
	args := []string{
		"--remote-debugging-port=9222",
		"--remote-debugging-address=127.0.0.1",
		fmt.Sprintf("--user-data-dir=%s", m.cfg.Browser.ProfileDir),
		fmt.Sprintf("--window-size=%d,%d", m.cfg.Browser.WindowWidth, m.cfg.Browser.WindowHeight),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-translate",
		"--disable-infobars",
		"--disable-features=TranslateUI",
		"--password-store=basic",
		"--use-mock-keychain",
		"--start-maximized",
	}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		args = append(args, fmt.Sprintf("--display=:%d", m.cfg.Display.DisplayNum))
	}

	args = append(args, m.cfg.Browser.ExtraArgs...)

	return args
}
