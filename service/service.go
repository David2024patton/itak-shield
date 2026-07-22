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

// installWindows tries to create a Windows Service via sc.exe. If the
// current process isn't admin, it writes the sc.exe commands to a temp
// .bat file and launches it with a UAC prompt (PowerShell -Verb RunAs),
// then waits for it to finish. If the user declines UAC or it fails, it
// falls back to the Startup-folder approach (no admin needed).
func installWindows(exePath string, args []string) error {
	binPath := exePath
	if len(args) > 0 {
		binPath = exePath + " " + strings.Join(args, " ")
	}

	// Try sc.exe directly. If we're admin this works; if not, it fails
	// with "access denied" and we escalate.
	cmd := exec.Command("sc.exe", "create", serviceName,
		"binPath=", binPath,
		"DisplayName=", serviceDisplayName,
		"start=", "auto",
		"obj=", "LocalSystem",
	)
	_, err := cmd.CombinedOutput()
	if err == nil {
		// Success — we're admin. Set description + recovery + start.
		exec.Command("sc.exe", "description", serviceName, serviceDescription).CombinedOutput()
		exec.Command("sc.exe", "failure", serviceName,
			"reset=", "86400",
			"actions=", "restart/60000/restart/60000/restart/60000",
		).CombinedOutput()
		exec.Command("sc.exe", "start", serviceName).CombinedOutput()
		fmt.Println("[iTaK Shield] Windows service installed and started.")
		return nil
	}

	// Not admin — try elevation via a temp .bat + UAC prompt.
	fmt.Println("[iTaK Shield] Requesting admin privileges (UAC prompt)...")
	if tryElevatedSC(exePath, binPath) {
		fmt.Println("[iTaK Shield] Windows service installed via UAC elevation.")
		return nil
	}

	// User declined UAC or it failed — fall back to Startup folder (no admin).
	fmt.Println("[iTaK Shield] Using Startup folder (no admin needed).")
	return installStartupFolder(exePath, args)
}

// tryElevatedSC writes the sc.exe commands to a temp .bat file and launches
// it elevated via PowerShell Start-Process -Verb RunAs (UAC prompt). Returns
// true if the service was created successfully.
func tryElevatedSC(exePath, binPath string) bool {
	// Write a temp batch file with all the sc.exe commands.
	tmpDir := os.Getenv("TEMP")
	if tmpDir == "" {
		tmpDir = "."
	}
	batPath := filepath.Join(tmpDir, "itak-shield-install.bat")
	resultPath := filepath.Join(tmpDir, "itak-shield-install-result.txt")

	// Clean up any previous result file.
	os.Remove(resultPath)

	bat := fmt.Sprintf(`@echo off
sc.exe create %s binPath= "%s" DisplayName= "%s" start= auto obj= LocalSystem
if %%errorlevel%% neq 0 (
  echo FAILED > "%s"
  exit /b 1
)
sc.exe description %s "%s"
sc.exe failure %s reset= 86400 actions= restart/60000/restart/60000/restart/60000
sc.exe start %s
echo OK > "%s"
`,
		serviceName, binPath, serviceDisplayName,
		resultPath,
		serviceName, serviceDescription,
		serviceName,
		serviceName,
		resultPath,
	)

	if err := os.WriteFile(batPath, []byte(bat), 0644); err != nil {
		return false
	}
	defer os.Remove(batPath)

	// Launch the bat elevated and wait for it.
	psCmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs -Wait", batPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	err := cmd.Run()
	if err != nil {
		return false
	}

	// Check the result file.
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return false
	}
	os.Remove(resultPath)
	return strings.TrimSpace(string(data)) == "OK"
}

// installStartupFolder creates a VBS launcher in the Windows Startup folder
// so Shield launches when the user logs in. Does NOT require admin privileges.
func installStartupFolder(exePath string, args []string) error {
	startupDir := getStartupFolderPath()
	if err := os.MkdirAll(startupDir, 0755); err != nil {
		return fmt.Errorf("cannot access Startup folder: %w", err)
	}

	vbsPath := filepath.Join(startupDir, "itak-shield.vbs")
	argsStr := ""
	for _, a := range args {
		argsStr += " \"" + a + "\""
	}
	vbs := fmt.Sprintf(`Set WshShell = CreateObject("WScript.Shell")
WshShell.Run """%s""%s""", 0, False
`, exePath, argsStr)

	if err := os.WriteFile(vbsPath, []byte(vbs), 0644); err != nil {
		return fmt.Errorf("cannot write to Startup folder: %w", err)
	}

	fmt.Println("[iTaK Shield] Auto-start installed via Startup folder: " + vbsPath)
	return nil
}

// getStartupFolderPath returns the user's Windows Startup folder.
func getStartupFolderPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
}

func uninstallWindows() error {
	// Try service uninstall (needs admin, but best-effort).
	stopCmd := exec.Command("sc.exe", "stop", serviceName)
	stopCmd.CombinedOutput()
	cmd := exec.Command("sc.exe", "delete", serviceName)
	cmd.CombinedOutput()

	// Remove Startup folder launcher (no admin needed).
	vbsPath := filepath.Join(getStartupFolderPath(), "itak-shield.vbs")
	os.Remove(vbsPath)

	fmt.Println("[iTaK Shield] Auto-start removed.")
	return nil
}

func statusWindows() (bool, bool, error) {
	// Check Windows Service.
	cmd := exec.Command("sc.exe", "query", serviceName)
	output, err := cmd.CombinedOutput()
	if err == nil {
		out := string(output)
		installed := strings.Contains(out, serviceName)
		running := strings.Contains(out, "RUNNING")
		if installed {
			return true, running, nil
		}
	}
	// Check Startup folder fallback.
	vbsPath := filepath.Join(getStartupFolderPath(), "itak-shield.vbs")
	if _, err := os.Stat(vbsPath); err == nil {
		return true, false, nil // Installed via startup folder.
	}
	return false, false, nil
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
