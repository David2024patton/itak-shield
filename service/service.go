package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Install registers iTaK Shield as a system service.
// On Windows: creates a Windows Service via sc.exe
// On Linux: generates a systemd unit file
// On macOS: generates a launchd plist
func Install(exePath string, args []string) error {
	if exePath == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot find executable path: %w", err)
		}
		exePath, _ = filepath.Abs(exe)
	}

	switch runtime.GOOS {
	case "windows":
		return installWindows(exePath, args)
	case "linux":
		return installLinux(exePath, args)
	case "darwin":
		return installDarwin(exePath, args)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// Uninstall removes the system service registration.
func Uninstall() error {
	switch runtime.GOOS {
	case "windows":
		return uninstallWindows()
	case "linux":
		return uninstallLinux()
	case "darwin":
		return uninstallDarwin()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// Status checks if the service is installed and running.
func Status() (installed bool, running bool, err error) {
	switch runtime.GOOS {
	case "windows":
		return statusWindows()
	case "linux":
		return statusLinux()
	case "darwin":
		return statusDarwin()
	default:
		return false, false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

const serviceName = "iTaKShield"
const serviceDisplayName = "iTaK Shield Privacy Proxy"
const serviceDescription = "AI privacy proxy that scans for PII, blocks prompt injections, and enforces data loss prevention policies."

// ─── Windows ────────────────────────────────────

func installWindows(exePath string, args []string) error {
	// Use sc.exe to create the service. This requires admin privileges.
	binPath := exePath
	if len(args) > 0 {
		binPath = exePath + " " + strings.Join(args, " ")
	}

	cmd := exec.Command("sc.exe", "create", serviceName,
		"binPath=", binPath,
		"DisplayName=", serviceDisplayName,
		"start=", "auto",
		"obj=", "LocalSystem",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc create failed: %s (%w)", string(output), err)
	}

	// Set description.
	descCmd := exec.Command("sc.exe", "description", serviceName, serviceDescription)
	descCmd.CombinedOutput()

	// Set recovery: restart on failure.
	recoverCmd := exec.Command("sc.exe", "failure", serviceName,
		"reset=", "86400",
		"actions=", "restart/60000/restart/60000/restart/60000",
	)
	recoverCmd.CombinedOutput()

	fmt.Println("[iTaK Shield] Windows service installed. Start with: sc start " + serviceName)
	return nil
}

func uninstallWindows() error {
	// Stop first.
	stopCmd := exec.Command("sc.exe", "stop", serviceName)
	stopCmd.CombinedOutput() // Ignore errors if not running.

	cmd := exec.Command("sc.exe", "delete", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc delete failed: %s (%w)", string(output), err)
	}
	fmt.Println("[iTaK Shield] Windows service removed.")
	return nil
}

func statusWindows() (bool, bool, error) {
	cmd := exec.Command("sc.exe", "query", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, false, nil // Not installed.
	}
	out := string(output)
	installed := strings.Contains(out, serviceName)
	running := strings.Contains(out, "RUNNING")
	return installed, running, nil
}

// ─── Linux (systemd) ────────────────────────────

const systemdUnitPath = "/etc/systemd/system/itak-shield.service"

func installLinux(exePath string, args []string) error {
	argsStr := strings.Join(args, " ")
	workDir := filepath.Dir(exePath)

	unit := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
ExecStart=%s %s
WorkingDirectory=%s
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, serviceDescription, exePath, argsStr, workDir)

	if err := os.WriteFile(systemdUnitPath, []byte(unit), 0644); err != nil {
		return fmt.Errorf("write unit file: %w (try running with sudo)", err)
	}

	// Reload systemd.
	exec.Command("systemctl", "daemon-reload").CombinedOutput()

	// Enable on boot.
	exec.Command("systemctl", "enable", "itak-shield.service").CombinedOutput()

	fmt.Println("[iTaK Shield] systemd service installed. Start with: sudo systemctl start itak-shield")
	return nil
}

func uninstallLinux() error {
	exec.Command("systemctl", "stop", "itak-shield.service").CombinedOutput()
	exec.Command("systemctl", "disable", "itak-shield.service").CombinedOutput()

	if err := os.Remove(systemdUnitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove unit file: %w", err)
	}

	exec.Command("systemctl", "daemon-reload").CombinedOutput()

	fmt.Println("[iTaK Shield] systemd service removed.")
	return nil
}

func statusLinux() (bool, bool, error) {
	if _, err := os.Stat(systemdUnitPath); os.IsNotExist(err) {
		return false, false, nil
	}

	cmd := exec.Command("systemctl", "is-active", "itak-shield.service")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, false, nil // Installed but not running.
	}
	running := strings.TrimSpace(string(output)) == "active"
	return true, running, nil
}

// ─── macOS (launchd) ────────────────────────────

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.itak.shield.plist")
}

func installDarwin(exePath string, args []string) error {
	plistPath := launchdPlistPath()
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return err
	}

	argsXML := ""
	for _, a := range args {
		argsXML += fmt.Sprintf("\n    <string>%s</string>", a)
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.itak.shield</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>%s
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>WorkingDirectory</key>
  <string>%s</string>
  <key>StandardOutPath</key>
  <string>/tmp/itak-shield.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/itak-shield.err</string>
</dict>
</plist>
`, exePath, argsXML, filepath.Dir(exePath))

	if err := os.WriteFile(plistPath, []byte(plist), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	// Load the service.
	exec.Command("launchctl", "load", plistPath).CombinedOutput()

	fmt.Println("[iTaK Shield] launchd agent installed. It will start on login.")
	return nil
}

func uninstallDarwin() error {
	plistPath := launchdPlistPath()

	exec.Command("launchctl", "unload", plistPath).CombinedOutput()

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}

	fmt.Println("[iTaK Shield] launchd agent removed.")
	return nil
}

func statusDarwin() (bool, bool, error) {
	plistPath := launchdPlistPath()
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return false, false, nil
	}

	cmd := exec.Command("launchctl", "list", "com.itak.shield")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, false, nil
	}
	running := strings.Contains(string(output), "com.itak.shield")
	return true, running, nil
}
