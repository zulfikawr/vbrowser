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

// FindSystemBrowser attempts to find an installed Chromium or Google Chrome binary.
// browserType can be "auto", "chrome", or "chromium".
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

	switch runtime.GOOS {
	case "linux":
		if browserType == "chrome" {
			names = chromeNames
		} else if browserType == "chromium" {
			names = chromiumNames
		} else {
			names = append(chromiumNames, chromeNames...)
		}

		paths = []string{
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/snap/bin/chromium",
		}
	case "darwin":
		chromePaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		}
		chromiumPaths := []string{
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}

		if browserType == "chrome" {
			names = []string{"Google Chrome"}
			paths = chromePaths
		} else if browserType == "chromium" {
			names = []string{"Chromium"}
			paths = chromiumPaths
		} else {
			names = []string{"Chromium", "Google Chrome"}
			paths = append(chromiumPaths, chromePaths...)
		}
	case "windows":
		names = []string{"chrome.exe", "chromium.exe"}
		localAppData := os.Getenv("LocalAppData")
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")

		chromePaths := []string{
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
		}
		// Common Chromium paths on Windows
		chromiumPaths := []string{
			filepath.Join(programFiles, "Chromium", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Chromium", "Application", "chrome.exe"),
		}

		if browserType == "chrome" {
			paths = chromePaths
		} else if browserType == "chromium" {
			paths = chromiumPaths
		} else {
			paths = append(chromiumPaths, chromePaths...)
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
			// If user specified chrome/chromium, make sure path matches
			lowerPath := strings.ToLower(path)
			if browserType == "chrome" && !strings.Contains(lowerPath, "chrome") {
				continue
			}
			if browserType == "chromium" && !strings.Contains(lowerPath, "chromium") {
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

// GetInstallInstructions returns platform-specific instructions for installing a browser.
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "Please install Chromium (Google Chrome is not available for ARM64 Linux):\n" +
				"  - sudo apt update && sudo apt install -y chromium-browser"
		}
		return "Please install a browser:\n" +
			"  - Chromium (Ubuntu/Debian): sudo apt install chromium-browser\n" +
			"  - Google Chrome (x86_64 only): wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && sudo apt install ./google-chrome-stable_current_amd64.deb"
	case "darwin":
		return "Please install Chromium or Google Chrome. You can use Homebrew: brew install --cask chromium"
	case "windows":
		return "Please install Google Chrome from https://www.google.com/chrome/"
	default:
		return "Please install Chromium or Google Chrome and ensure it is in your PATH."
	}
}
