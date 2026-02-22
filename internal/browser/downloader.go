// Package browser provides Chromium browser management for vbrowser.
package browser

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

const (
	baseURL = "https://commondatastorage.googleapis.com/chromium-browser-snapshots"
)

// DetectPlatform returns the Chromium platform string for the current OS/arch.
func DetectPlatform() (string, error) {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "Linux_x64", nil
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "Mac_Arm", nil
		}
		return "Mac", nil
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "Win_x64", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

// FetchLatestRevision fetches the latest Chromium revision for the given platform.
func FetchLatestRevision(platform string) (string, error) {
	url := fmt.Sprintf("%s/%s/LAST_CHANGE", baseURL, platform)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch revision: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch revision: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read revision: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// Download downloads and extracts Chromium to the destination directory.
func Download(platform, revision, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	zipName := getZipName(platform)
	url := fmt.Sprintf("%s/%s/%s/%s", baseURL, platform, revision, zipName)

	log.Info().Str("url", url).Msg("downloading Chromium")

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: status %d", resp.StatusCode)
	}

	tmpFile := filepath.Join(destDir, zipName)
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	if _, err := io.Copy(io.MultiWriter(out, bar), resp.Body); err != nil {
		out.Close()
		return fmt.Errorf("download: %w", err)
	}
	out.Close()

	log.Info().Msg("extracting Chromium")
	if err := extractZip(tmpFile, destDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if err := os.WriteFile(filepath.Join(destDir, ".revision"), []byte(revision), 0644); err != nil {
		return fmt.Errorf("write revision: %w", err)
	}

	log.Info().Str("revision", revision).Msg("Chromium downloaded successfully")
	return nil
}

// GetChromiumPath returns the path to the Chromium binary in the given directory.
func GetChromiumPath(destDir string) (string, error) {
	var paths []string

	switch runtime.GOOS {
	case "linux":
		paths = []string{
			filepath.Join(destDir, "chrome-linux", "chrome"),
			filepath.Join(destDir, "chrome"),
		}
	case "darwin":
		paths = []string{
			filepath.Join(destDir, "chrome-mac", "Chromium.app", "Contents", "MacOS", "Chromium"),
			filepath.Join(destDir, "Chromium.app", "Contents", "MacOS", "Chromium"),
		}
	case "windows":
		paths = []string{
			filepath.Join(destDir, "chrome-win", "chrome.exe"),
			filepath.Join(destDir, "chrome.exe"),
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("chromium binary not found in %s", destDir)
}

// GetCachedRevision returns the cached revision number, or empty string if not cached.
func GetCachedRevision(destDir string) string {
	data, err := os.ReadFile(filepath.Join(destDir, ".revision"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func getZipName(platform string) string {
	switch platform {
	case "Linux_x64":
		return "chrome-linux.zip"
	case "Mac", "Mac_Arm":
		return "chrome-mac.zip"
	case "Win_x64":
		return "chrome-win.zip"
	default:
		return "chrome.zip"
	}
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
