// Package browser provides browser management for vbrowser.
package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
)

// FindSystemBrowser attempts to find an installed Chromium, Google Chrome, or Firefox binary.
// browserType can be "auto", "chrome", "chromium", or "firefox".
func FindSystemBrowser(browserType string) (string, error) {
	var names []string
	var paths []string

	chromeNames := []string{
		"google-chrome",
		"google-chrome-stable",
		"google-chrome-unstable",
	}
	chromiumNames := []string{
		"chromium",
		"chromium-browser",
	}
	firefoxNames := []string{
		"firefox",
		"firefox-esr",
	}

	switch runtime.GOOS {
	case "linux":
		if browserType == "chrome" {
			names = chromeNames
		} else if browserType == "chromium" {
			names = chromiumNames
		} else if browserType == "firefox" {
			names = firefoxNames
		} else {
			names = append(chromiumNames, append(chromeNames, firefoxNames...)...)
		}

		paths = []string{
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/firefox",
			"/snap/bin/chromium",
			"/snap/bin/firefox",
		}
	case "darwin":
		chromePaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		}
		chromiumPaths := []string{
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		firefoxPaths := []string{
			"/Applications/Firefox.app/Contents/MacOS/firefox",
		}

		if browserType == "chrome" {
			names = []string{"Google Chrome"}
			paths = chromePaths
		} else if browserType == "chromium" {
			names = []string{"Chromium"}
			paths = chromiumPaths
		} else if browserType == "firefox" {
			names = []string{"Firefox"}
			paths = firefoxPaths
		} else {
			names = []string{"Chromium", "Google Chrome", "Firefox"}
			paths = append(chromiumPaths, append(chromePaths, firefoxPaths...)...)
		}
	case "windows":
		names = []string{"chrome.exe", "chromium.exe", "firefox.exe"}
		localAppData := os.Getenv("LocalAppData")
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")

		chromePaths := []string{
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
		}
		chromiumPaths := []string{
			filepath.Join(programFiles, "Chromium", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Chromium", "Application", "chrome.exe"),
		}
		firefoxPaths := []string{
			filepath.Join(programFiles, "Mozilla Firefox", "firefox.exe"),
			filepath.Join(programFilesX86, "Mozilla Firefox", "firefox.exe"),
		}

		if browserType == "chrome" {
			paths = chromePaths
		} else if browserType == "chromium" {
			paths = chromiumPaths
		} else if browserType == "firefox" {
			paths = firefoxPaths
		} else {
			paths = append(chromiumPaths, append(chromePaths, firefoxPaths...)...)
		}
	}

	// 1. Try PATH
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			log.Debug().Str("name", name).Str("path", path).Msg("Found browser in PATH")
			return path, nil
		}
	}

	// 2. Try common absolute paths
	for _, path := range paths {
		if browserType != "auto" {
			lowerPath := strings.ToLower(path)
			if browserType == "chrome" && !strings.Contains(lowerPath, "chrome") {
				continue
			}
			if browserType == "chromium" && !strings.Contains(lowerPath, "chromium") {
				continue
			}
			if browserType == "firefox" && !strings.Contains(lowerPath, "firefox") {
				continue
			}
		}

		if _, err := os.Stat(path); err == nil {
			log.Debug().Str("path", path).Msg("Found browser in common path")
			return path, nil
		}
	}

	return "", fmt.Errorf("%s binary not found", browserType)
}

// FindAllBrowsers returns a list of all detected browser binaries on the system.
func FindAllBrowsers() []string {
	found := make(map[string]bool)
	var results []string

	chromeNames := []string{"google-chrome", "google-chrome-stable", "google-chrome-unstable"}
	chromiumNames := []string{"chromium", "chromium-browser"}
	firefoxNames := []string{"firefox", "firefox-esr"}
	allNames := append(chromiumNames, append(chromeNames, firefoxNames...)...)

	allPaths := []string{
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/firefox",
		"/snap/bin/chromium",
		"/snap/bin/firefox",
	}

	// Try PATH
	for _, name := range allNames {
		if path, err := exec.LookPath(name); err == nil {
			if !found[path] {
				found[path] = true
				results = append(results, path)
			}
		}
	}

	// Try common paths
	for _, path := range allPaths {
		if _, err := os.Stat(path); err == nil {
			if !found[path] {
				found[path] = true
				results = append(results, path)
			}
		}
	}

	return results
}

// GetInstallInstructions returns platform-specific instructions for installing a browser.
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "Please install a browser:\n" +
				"  - Chromium: sudo apt update && sudo apt install -y chromium-browser\n" +
				"  - Firefox:  sudo apt update && sudo apt install -y firefox"
		}
		return "Please install a browser:\n" +
			"  - Chromium (Ubuntu/Debian): sudo apt install chromium-browser\n" +
			"  - Google Chrome (x86_64 only): wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && sudo apt install ./google-chrome-stable_current_amd64.deb\n" +
			"  - Firefox: sudo apt install firefox"
	case "darwin":
		return "Please install Chromium, Google Chrome, or Firefox. You can use Homebrew: brew install --cask firefox"
	case "windows":
		return "Please install Google Chrome or Firefox."
	default:
		return "Please install a supported browser (Chrome, Chromium, or Firefox) and ensure it is in your PATH."
	}
}
