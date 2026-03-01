package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates on GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("checking for updates on GitHub...")

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get("https://api.github.com/repos/zulfikawr/vbrowser/releases/latest")
		if err != nil {
			log.Error().Err(err).Msg("failed to check for updates")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error().Int("status", resp.StatusCode).Msg("failed to check for updates")
			return
		}

		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			log.Error().Err(err).Msg("failed to parse release info")
			return
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		if latestVersion == Version {
			fmt.Printf("vbrowser is already up to date (v%s)\n", Version)
			return
		}

		// Simple semver check for new version
		// If we are ahead (current > latest), we don't need to update
		if isNewer(latestVersion, Version) {
			fmt.Printf("A new version is available: v%s (current: v%s)\n", latestVersion, Version)
			fmt.Printf("Release notes and download: %s\n", release.HTMLURL)
			fmt.Println("\nTo update, run:")
			fmt.Println("curl -sSL https://raw.githubusercontent.com/zulfikawr/vbrowser/main/scripts/install.sh | sudo bash")
		} else {
			fmt.Printf("vbrowser is up to date (v%s). Latest release is v%s\n", Version, latestVersion)
		}
	},
}

func isNewer(latest, current string) bool {
	lParts := strings.Split(latest, ".")
	cParts := strings.Split(current, ".")

	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		var l, c int
		if _, err := fmt.Sscanf(lParts[i], "%d", &l); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(cParts[i], "%d", &c); err != nil {
			continue
		}

		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return len(lParts) > len(cParts)
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
