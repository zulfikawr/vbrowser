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

func (m *Manager) initPulseAudio() string {
	// 1. Kill any existing PulseAudio to start fresh
	_ = exec.Command("pulseaudio", "--kill").Run()
	time.Sleep(500 * time.Millisecond)

	// 2. Start PulseAudio with minimal settings
	_ = exec.Command("pulseaudio", "--start", "--exit-idle-time=-1").Run()
	time.Sleep(500 * time.Millisecond)

	// 3. Unload suspend module
	_ = exec.Command("pactl", "unload-module", "module-suspend-on-idle").Run()

	sinkName := fmt.Sprintf("vbrowser-%d", m.cfg.Display.DisplayNum)

	// 4. Load the null-sink
	_ = exec.Command("pactl", "load-module", "module-null-sink",
		fmt.Sprintf("sink_name=%s", sinkName),
		fmt.Sprintf("sink_properties=device.description=%s", sinkName)).Run()

	// 5. Force it as default
	_ = exec.Command("pactl", "set-default-sink", sinkName).Run()
	_ = exec.Command("pactl", "set-sink-mute", sinkName, "0").Run()
	_ = exec.Command("pactl", "set-sink-volume", sinkName, "100%").Run()

	return sinkName
}

// Start launches Xvfb (on Linux) and Chromium.
func (m *Manager) Start(chromiumPath string) error {
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
	}

	args := m.buildChromiumArgs()

	// Nuclear option: use a dedicated bash script to ensure environment is perfect
	m.chromiumCmd = exec.Command(chromiumPath, args...)
	m.chromiumCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		m.chromiumCmd.Env = append(os.Environ(),
			fmt.Sprintf("DISPLAY=:%d", m.cfg.Display.DisplayNum),
			fmt.Sprintf("PULSE_SINK=%s", sinkName),
			"PULSE_SERVER=unix:/run/user/1000/pulse/native",
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
		pgid, err := syscall.Getpgid(m.chromiumCmd.Process.Pid)
		if err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
		} else {
			_ = m.chromiumCmd.Process.Signal(syscall.SIGTERM)
		}
		_ = m.chromiumCmd.Wait()
	}

	m.stopXvfb()
	_ = exec.Command("pulseaudio", "--kill").Run()

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

// Restart stops and starts the browser with new settings.
func (m *Manager) Restart(chromiumPath string) error {
	log.Info().Msg("restarting browser manager for configuration change")
	if err := m.Stop(); err != nil {
		log.Warn().Err(err).Msg("failed to stop manager during restart")
	}
	return m.Start(chromiumPath)
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
	}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		args = append(args, fmt.Sprintf("--display=:%d", m.cfg.Display.DisplayNum))
	}

	args = append(args, m.cfg.Browser.ExtraArgs...)

	return args
}
