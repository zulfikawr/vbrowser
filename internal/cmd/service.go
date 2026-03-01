package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage vbrowser systemd user service",
}

const serviceTemplate = `[Unit]
Description=vbrowser - Self-hosted virtual browser
After=network.target

[Service]
ExecStart={{.Executable}} start --foreground --config {{.Config}}
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and start vbrowser as a systemd user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find executable: %w", err)
		}

		homeDir, _ := os.UserHomeDir()
		serviceDir := filepath.Join(homeDir, ".config/systemd/user")
		servicePath := filepath.Join(serviceDir, "vbrowser.service")

		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			return fmt.Errorf("create service dir: %w", err)
		}

		f, err := os.Create(servicePath)
		if err != nil {
			return fmt.Errorf("create service file: %w", err)
		}
		defer f.Close()

		tmpl := template.Must(template.New("service").Parse(serviceTemplate))
		err = tmpl.Execute(f, struct {
			Executable string
			Config     string
		}{
			Executable: executable,
			Config:     cfgFile,
		})
		if err != nil {
			return fmt.Errorf("generate service file: %w", err)
		}

		log.Info().Str("path", servicePath).Msg("Service file created")

		// Reload systemd and start service
		cmds := [][]string{
			{"systemctl", "--user", "daemon-reload"},
			{"systemctl", "--user", "enable", "vbrowser.service"},
			{"systemctl", "--user", "start", "vbrowser.service"},
		}

		for _, c := range cmds {
			if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
				return fmt.Errorf("failed to run %v: %w", c, err)
			}
		}

		log.Info().Msg("vbrowser service installed and started successfully")
		fmt.Println("\nTo check service status, run:")
		fmt.Println("systemctl --user status vbrowser.service")
		fmt.Println("\nTo view logs, run:")
		fmt.Println("journalctl --user -u vbrowser.service -f")
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Stop and uninstall vbrowser systemd user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmds := [][]string{
			{"systemctl", "--user", "stop", "vbrowser.service"},
			{"systemctl", "--user", "disable", "vbrowser.service"},
		}

		for _, c := range cmds {
			_ = exec.Command(c[0], c[1:]...).Run()
		}

		homeDir, _ := os.UserHomeDir()
		servicePath := filepath.Join(homeDir, ".config/systemd/user/vbrowser.service")
		if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove service file: %w", err)
		}

		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()

		log.Info().Msg("vbrowser service uninstalled successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
}
