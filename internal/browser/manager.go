package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// getEffectiveProfileDir returns a browser-specific subdirectory to avoid profile corruption
func (m *Manager) getEffectiveProfileDir() string {
	baseDir := m.cfg.Browser.ProfileDir
	if m.isFirefox() {
		// Firefox Snap on Ubuntu cannot access hidden directories like .local
		// We'll use a more standard path for the firefox profile if it's linux
		if runtime.GOOS == "linux" {
			home, _ := os.UserHomeDir()
			return filepath.Join(home, "vbrowser_firefox_profile")
		}
		return filepath.Join(baseDir, "firefox")
	}
	return filepath.Join(baseDir, "chrome")
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

	if m.isFirefox() {
		if err := m.initFirefoxProfile(); err != nil {
			log.Warn().Err(err).Msg("failed to initialize firefox profile")
		}
	}

	args := m.buildBrowserArgs()

	m.browserCmd = exec.Command(browserPath, args...)
	m.browserCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Set environment variables
	env := os.Environ()
	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		env = append(env,
			fmt.Sprintf("DISPLAY=:%d", m.cfg.Display.DisplayNum),
			fmt.Sprintf("PULSE_SINK=%s", sinkName),
		)
	}
	if m.isFirefox() {
		env = append(env, "MOZ_NO_REMOTE=1")
	}
	m.browserCmd.Env = env

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

func (m *Manager) isFirefox() bool {
	lowerPath := strings.ToLower(m.browserPath)
	return strings.Contains(lowerPath, "firefox")
}

func (m *Manager) initFirefoxProfile() error {
	profileDir := m.getEffectiveProfileDir()
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}

	// Remove stale lock files that cause "Firefox is already running" errors
	os.Remove(filepath.Join(profileDir, "lock"))
	os.Remove(filepath.Join(profileDir, ".parentlock"))
	os.Remove(filepath.Join(profileDir, "parent.lock"))

	prefsPath := filepath.Join(profileDir, "prefs.js")
	prefs := []string{
		"user_pref(\"browser.shell.checkDefaultBrowser\", false);",
		"user_pref(\"browser.startup.homepage_welcome_url\", \"\");",
		"user_pref(\"browser.startup.homepage_welcome_url.additional\", \"\");",
		"user_pref(\"devtools.chrome.enabled\", true);",
		"user_pref(\"devtools.debugger.prompt-connection\", false);",
		"user_pref(\"devtools.debugger.remote-enabled\", true);",
		"user_pref(\"toolkit.telemetry.reportingpolicy.firstRunCheck\", false);",
		"user_pref(\"trailhead.firstrun.branches\", \"none\");",
		"user_pref(\"datareporting.policy.dataSubmissionEnabled\", false);",
		"user_pref(\"browser.tabs.remote.autostart\", true);",
		"user_pref(\"focusmanager.testmode\", true);",
		"user_pref(\"browser.tabs.unloadOnLowMemory\", false);",
		"user_pref(\"browser.sessionstore.interval\", 600000);",
		fmt.Sprintf("user_pref(\"width\", %d);", m.cfg.Browser.WindowWidth),
		fmt.Sprintf("user_pref(\"height\", %d);", m.cfg.Browser.WindowHeight),
	}

	return os.WriteFile(prefsPath, []byte(strings.Join(prefs, "\n")), 0644)
}

func (m *Manager) buildBrowserArgs() []string {
	if m.isFirefox() {
		return m.buildFirefoxArgs()
	}
	return m.buildChromiumArgs()
}

func (m *Manager) buildFirefoxArgs() []string {
	args := []string{
		"--remote-debugging-port=9222",
		"--profile", m.getEffectiveProfileDir(),
		"--no-remote",
		"--new-instance",
		"-width", fmt.Sprintf("%d", m.cfg.Browser.WindowWidth),
		"-height", fmt.Sprintf("%d", m.cfg.Browser.WindowHeight),
	}

	args = append(args, m.cfg.Browser.ExtraArgs...)
	return args
}

func (m *Manager) buildChromiumArgs() []string {
	args := []string{
		"--remote-debugging-port=9222",
		"--remote-debugging-address=127.0.0.1",
		fmt.Sprintf("--user-data-dir=%s", m.getEffectiveProfileDir()),
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
		"--disable-background-timer-throttling",
		"--disable-renderer-backgrounding",
		"--disable-backgrounding-occluded-windows",
		"--disable-ipc-flooding-protection",
		"--disable-background-networking",
		"--disable-hang-monitor",
		"--disable-breakpad",
		"--disable-client-side-phishing-detection",
		"--disable-component-update",
		"--disable-default-apps",
		"--disable-domain-reliability",
		"--disable-sync",
		"--metrics-recording-only",
		"--no-first-run",
		"--safebrowsing-disable-auto-update",
		"--password-store=basic",
		"--force-color-profile=srgb",
	}

	if runtime.GOOS == "linux" && m.cfg.Display.VirtualDisplay {
		args = append(args, fmt.Sprintf("--display=:%d", m.cfg.Display.DisplayNum))
	}

	args = append(args, m.cfg.Browser.ExtraArgs...)

	return args
}
