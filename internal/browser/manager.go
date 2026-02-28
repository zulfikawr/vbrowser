package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/vbrowser/internal/config"
	"github.com/zulfikawr/vbrowser/internal/platform"
	"github.com/zulfikawr/vbrowser/pkg/xorg"
)

// Manager manages the lifecycle of browser and Xvfb processes.
type Manager struct {
	cfg         *config.Config
	browserCmd  *exec.Cmd
	xvfbCmd     *exec.Cmd
	browserPid  int
	browserPath string
}

// NewManager creates a new browser manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

func (m *Manager) initPulseAudio() string {
	_ = exec.Command("pulseaudio", "--kill").Run()
	time.Sleep(500 * time.Millisecond)

	_ = exec.Command("pulseaudio", "--start", "--exit-idle-time=-1").Run()
	time.Sleep(500 * time.Millisecond)

	_ = exec.Command("pactl", "unload-module", "module-suspend-on-idle").Run()

	sinkName := fmt.Sprintf("vbrowser-%d", m.cfg.Display.DisplayNum)

	_ = exec.Command("pactl", "load-module", "module-null-sink",
		fmt.Sprintf("sink_name=%s", sinkName),
		fmt.Sprintf("sink_properties=device.description=%s", sinkName)).Run()

	_ = exec.Command("pactl", "set-default-sink", sinkName).Run()
	_ = exec.Command("pactl", "set-sink-mute", sinkName, "0").Run()
	_ = exec.Command("pactl", "set-sink-volume", sinkName, "100%").Run()

	return sinkName
}

// Start launches Xvfb (on Linux) and browser.
func (m *Manager) Start(browserPath string) error {
	m.browserPath = browserPath
	sinkName := m.initPulseAudio()

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

		// Open X11 display for native input handling
		displayStr := fmt.Sprintf(":%d", m.cfg.Display.DisplayNum)

		// Set DISPLAY env variable so C.XOpenDisplay(NULL) or C.XOpenDisplay(":99") works
		os.Setenv("DISPLAY", displayStr)

		// Wait a bit for Xvfb to be fully ready before opening the display
		success := false
		for i := 0; i < 20; i++ {
			if xorg.DisplayOpen(displayStr) {
				success = true
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		if !success {
			log.Error().Str("display", displayStr).Msg("failed to open display for xorg after 20 retries")
		}
	}

	args := m.buildBrowserArgs()

	// Nuclear option: use a dedicated bash script to ensure environment is perfect
	m.browserCmd = exec.Command(browserPath, args...)
	m.browserCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		m.browserCmd.Env = append(os.Environ(),
			fmt.Sprintf("DISPLAY=:%d", m.cfg.Display.DisplayNum),
			fmt.Sprintf("PULSE_SINK=%s", sinkName),
		)
	}

	if err := m.browserCmd.Start(); err != nil {
		m.stopXvfb()
		return fmt.Errorf("start browser: %w", err)
	}

	m.browserPid = m.browserCmd.Process.Pid

	log.Info().
		Int("pid", m.browserPid).
		Str("path", browserPath).
		Msg("browser started")

	return nil
}

// Stop gracefully terminates browser and Xvfb.
func (m *Manager) Stop() error {
	if m.browserCmd != nil && m.browserCmd.Process != nil {
		log.Info().Int("pid", m.browserPid).Msg("stopping browser")
		pgid, err := syscall.Getpgid(m.browserCmd.Process.Pid)
		if err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
		} else {
			_ = m.browserCmd.Process.Signal(syscall.SIGTERM)
		}
		_ = m.browserCmd.Wait()
	}

	xorg.DisplayClose()
	m.stopXvfb()
	_ = exec.Command("pulseaudio", "--kill").Run()

	return nil
}

// Pid returns the browser process PID.
func (m *Manager) Pid() int {
	return m.browserPid
}

// IsRunning checks if browser is still running.
func (m *Manager) IsRunning() bool {
	if m.browserCmd == nil || m.browserCmd.Process == nil {
		return false
	}
	err := m.browserCmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// Restart stops and starts the browser with new settings.
func (m *Manager) Restart(browserPath string) error {
	log.Info().Msg("restarting browser manager for configuration change")
	if err := m.Stop(); err != nil {
		log.Warn().Err(err).Msg("failed to stop manager during restart")
	}
	if browserPath == "" {
		browserPath = m.browserPath
	}
	return m.Start(browserPath)
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

func (m *Manager) buildBrowserArgs() []string {
	args := []string{
		"--remote-debugging-port=9222",
		"--remote-debugging-address=127.0.0.1",
		fmt.Sprintf("--user-data-dir=%s", m.cfg.Browser.ProfileDir),
		fmt.Sprintf("--window-size=%d,%d", m.cfg.Browser.WindowWidth, m.cfg.Browser.WindowHeight),
		fmt.Sprintf("--window-position=%d,%d", 0, 0),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-translate",
		"--disable-infobars",
		"--disable-features=TranslateUI",
		"--password-store=basic",
		"--use-mock-keychain",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-gpu",
		"--no-zygote",
		"--autoplay-policy=no-user-gesture-required",
		"--alsa-check-close-timeout=0",
		"--disable-audio-output-resampling",
		"--audio-buffer-size=4096",
		"--disable-features=AudioServiceSandbox",
		"--force-device-scale-factor=1",
		// Performance optimizations for low-latency streaming
		"--disable-background-timer-throttling",
		"--disable-renderer-backgrounding",
		"--disable-backgrounding-occluded-windows",
		"--force-color-profile=srgb",
	}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		args = append(args, fmt.Sprintf("--display=:%d", m.cfg.Display.DisplayNum))
	}

	args = append(args, m.cfg.Browser.ExtraArgs...)

	return args
}
