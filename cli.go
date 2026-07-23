package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/David2024patton/itak-shield/auth"
	shielddb "github.com/David2024patton/itak-shield/db"
	"github.com/David2024patton/itak-shield/guard"
	"github.com/David2024patton/itak-shield/scanner"
	svc "github.com/David2024patton/itak-shield/service"
)

// handleSubcommand checks if the first non-flag argument is a known subcommand.
// Returns true if handled (caller should exit), false to continue normal startup.
func handleSubcommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	cmd := strings.ToLower(args[0])
	subArgs := args[1:]

	switch cmd {
	case "scan":
		runScan(subArgs)
		return true
	case "user":
		runUser(subArgs)
		return true
	case "token":
		runToken(subArgs)
		return true
	case "install":
		runInstall(subArgs)
		return true
	case "uninstall":
		runUninstall()
		return true
	case "status":
		runServiceStatus()
		return true
	case "backup":
		runBackup(subArgs)
		return true
	case "help":
		printSubcommandHelp()
		return true
	}

	return false
}

// printSubcommandHelp shows available subcommands.
func printSubcommandHelp() {
	fmt.Println(`iTaK Shield - AI Privacy Proxy

SUBCOMMANDS:
  scan                Scan stdin/file for PII and prompt injection
  user list           List all users
  user add            Add a new user
  user delete         Delete a user
  token generate      Generate a Shield API token
  token revoke        Revoke a token
  token list          List tokens for a user
  install             Install as system service (start on boot)
  uninstall           Remove system service
  status              Show service status
  backup create       Create a database backup
  backup list         List backups

FLAGS (proxy mode):
  --target URL        Upstream API (e.g. https://api.openai.com)
  --port N            Proxy port (default: random)
  --gui-port N        Dashboard port (default: random)
  --bind ADDR         Bind address (default: 127.0.0.1)
  --config FILE       YAML config file
  --verbose           Show redaction details
  --no-gui            Run without dashboard
  --version           Print version

EXAMPLES:
  itak-shield                              # Launch dashboard
  itak-shield --target https://api.openai.com
  itak-shield scan < email.txt
  itak-shield scan suspicious_file.txt
  itak-shield user add --name "Dave" --email "dave@co.com" --type business --provider openai
  itak-shield token generate --user dave123 --label "laptop"
  itak-shield install
  itak-shield backup create`)
}

// ─── scan ───────────────────────────────────────

func runScan(args []string) {
	var input io.Reader
	if len(args) > 0 && args[0] != "-" {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot open %q: %v\n", args[0], err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	// Read all input.
	data, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	text := string(data)

	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "No input to scan. Pipe text or provide a filename.")
		fmt.Fprintln(os.Stderr, "  echo 'my SSN is 123-45-6789' | itak-shield scan")
		fmt.Fprintln(os.Stderr, "  itak-shield scan email.txt")
		os.Exit(1)
	}

	// PII scan.
	s := scanner.New()
	hits := s.Scan(text)

	// Prompt injection scan.
	g := guard.NewInputGuard()
	guardResult := g.ScanInput(text, "user")

	fmt.Println("=== iTaK Shield Scan Results ===")
	fmt.Println()

	if len(hits) == 0 && !guardResult.Blocked {
		fmt.Println("  No threats detected.")
	}

	if len(hits) > 0 {
		fmt.Printf("  PII Detected: %d item(s)\n", len(hits))
		for _, h := range hits {
			masked := h.Value
			if len(masked) > 4 {
				masked = masked[:2] + "***" + masked[len(masked)-2:]
			}
			fmt.Printf("    %-15s %s\n", h.Type, masked)
		}
		fmt.Println()
	}

	if guardResult.Blocked {
		fmt.Printf("  Prompt Injection: BLOCKED (severity: %s)\n", guardResult.Severity.String())
		for _, r := range guardResult.Reasons {
			fmt.Printf("    - %s\n", r)
		}
		fmt.Println()
	}

	if len(hits) > 0 || guardResult.Blocked {
		os.Exit(2) // Non-zero exit for CI/scripting use.
	}
}

// ─── user ───────────────────────────────────────

func runUser(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: itak-shield user [list|add|delete]")
		return
	}

	dbPath := filepath.Join(".", "shield.db")
	database, err := shielddb.Open(dbPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	store := database

	switch args[0] {
	case "list":
		profiles, err := database.ListUserProfiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(profiles) == 0 {
			fmt.Println("No users.")
			return
		}
		fmt.Printf("%-12s %-20s %-10s %-10s %-12s %-10s\n", "ID", "NAME", "TYPE", "PROVIDER", "ACTIVE KEYS", "SPEND")
		fmt.Println(strings.Repeat("-", 78))
		for _, p := range profiles {
			fmt.Printf("%-12s %-20s %-10s %-10s %-12d $%.4f\n",
				truncate(p.ID, 12), truncate(p.Name, 20),
				p.Type, truncate(p.Provider, 10),
				p.ActiveTokens, p.EstimatedUSD)
		}

	case "add":
		name, email, group, userType, provider, upstreamKey := "", "", "default", "personal", "", ""
		rateLimit := 0
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--name":
				if i+1 < len(args) { name = args[i+1]; i++ }
			case "--email":
				if i+1 < len(args) { email = args[i+1]; i++ }
			case "--group":
				if i+1 < len(args) { group = args[i+1]; i++ }
			case "--type":
				if i+1 < len(args) { userType = args[i+1]; i++ }
			case "--provider":
				if i+1 < len(args) { provider = args[i+1]; i++ }
			case "--key":
				if i+1 < len(args) { upstreamKey = args[i+1]; i++ }
			case "--rate-limit":
				if i+1 < len(args) { fmt.Sscanf(args[i+1], "%d", &rateLimit); i++ }
			}
		}
		if name == "" {
			fmt.Println("Usage: itak-shield user add --name \"Dave\" [--email dave@co.com] [--type business] [--provider openai] [--key sk-...]")
			return
		}

		mgr := auth.New(store, "")
		user, err := mgr.CreateUser(name, email, group, rateLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if provider != "" || userType != "personal" {
			database.UpdateUserProfile(user.ID, userType, provider, upstreamKey)
		}
		token, _ := mgr.GenerateToken(user.ID, "default", nil)

		fmt.Printf("User created: %s (ID: %s)\n", user.Name, user.ID)
		if token != nil {
			fmt.Printf("Shield API Key: %s\n", token.Key)
			fmt.Println("Save this key - it will not be shown again.")
		}

	case "delete":
		userID := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "--id" && i+1 < len(args) {
				userID = args[i+1]; i++
			}
		}
		if userID == "" {
			fmt.Println("Usage: itak-shield user delete --id USER_ID")
			return
		}

		// Confirm.
		fmt.Printf("Delete user %q? Type 'yes' to confirm: ", userID)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(answer) != "yes" {
			fmt.Println("Cancelled.")
			return
		}

		mgr := auth.New(store, "")
		if err := mgr.DeleteUser(userID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("User deleted.")
	}
}

// ─── token ──────────────────────────────────────

func runToken(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: itak-shield token [generate|revoke|list]")
		return
	}

	dbPath := filepath.Join(".", "shield.db")
	database, err := shielddb.Open(dbPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	mgr := auth.New(database, "")

	switch args[0] {
	case "generate":
		userID, label := "", "default"
		var expiry *time.Time
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--user":
				if i+1 < len(args) { userID = args[i+1]; i++ }
			case "--label":
				if i+1 < len(args) { label = args[i+1]; i++ }
			case "--expires":
				if i+1 < len(args) {
					d, err := time.ParseDuration(args[i+1])
					if err == nil {
						t := time.Now().Add(d)
						expiry = &t
					}
					i++
				}
			}
		}
		if userID == "" {
			fmt.Println("Usage: itak-shield token generate --user USER_ID [--label laptop] [--expires 720h]")
			return
		}
		token, err := mgr.GenerateToken(userID, label, expiry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Token generated: %s\n", token.Key)
		fmt.Printf("Label: %s\n", token.Label)
		if token.ExpiresAt != nil {
			fmt.Printf("Expires: %s\n", token.ExpiresAt.Format(time.RFC3339))
		}

	case "revoke":
		userID, key := "", ""
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--user":
				if i+1 < len(args) { userID = args[i+1]; i++ }
			case "--key":
				if i+1 < len(args) { key = args[i+1]; i++ }
			}
		}
		if userID == "" || key == "" {
			fmt.Println("Usage: itak-shield token revoke --user USER_ID --key TOKEN_KEY")
			return
		}
		if err := mgr.RevokeToken(userID, key); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Token revoked.")

	case "list":
		userID := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "--user" && i+1 < len(args) {
				userID = args[i+1]; i++
			}
		}
		if userID == "" {
			fmt.Println("Usage: itak-shield token list --user USER_ID")
			return
		}
		user, err := mgr.GetUser(userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(user.Tokens) == 0 {
			fmt.Println("No tokens for this user.")
			return
		}
		fmt.Printf("%-32s %-12s %-8s %-20s\n", "KEY", "LABEL", "STATUS", "EXPIRES")
		fmt.Println(strings.Repeat("-", 76))
		for _, t := range user.Tokens {
			status := "active"
			if t.Revoked { status = "revoked" }
			expires := "never"
			if t.ExpiresAt != nil { expires = t.ExpiresAt.Format("2006-01-02") }
			fmt.Printf("%-32s %-12s %-8s %-20s\n", truncate(t.Key, 32), t.Label, status, expires)
		}
	}
}

// ─── install / uninstall / status ───────────────

func runInstall(args []string) {
	// Parse optional --config and --port flags from the install args.
	// These get baked into the service's binPath so the service starts
	// with the right configuration on boot.
	var configPath string
	var port int
	var target string
	var extraArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config":
			if i+1 < len(args) { configPath = args[i+1]; i++ }
		case "--port":
			if i+1 < len(args) { fmt.Sscanf(args[i+1], "%d", &port); i++ }
		case "--target":
			if i+1 < len(args) { target = args[i+1]; i++ }
		default:
			extraArgs = append(extraArgs, args[i])
		}
	}

	var svcArgs []string
	if configPath != "" {
		svcArgs = append(svcArgs, "--config", configPath)
	}
	if port > 0 {
		svcArgs = append(svcArgs, "--port", fmt.Sprintf("%d", port))
	}
	if target != "" {
		svcArgs = append(svcArgs, "--target", target)
	}
	svcArgs = append(svcArgs, "--no-gui")
	svcArgs = append(svcArgs, extraArgs...)

	err := svc.Install("", svcArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
		os.Exit(1)
	}
}

func runUninstall() {
	err := svc.Uninstall()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
		os.Exit(1)
	}
}

func runServiceStatus() {
	installed, running, err := svc.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Status check failed: %v\n", err)
		os.Exit(1)
	}

	result := map[string]interface{}{
		"installed": installed,
		"running":   running,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

// ─── backup ─────────────────────────────────────

func runBackup(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: itak-shield backup [create|list]")
		return
	}

	dbPath := filepath.Join(".", "shield.db")
	database, err := shielddb.Open(dbPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	switch args[0] {
	case "create":
		path, err := database.Backup("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			os.Exit(1)
		}
		database.PruneBackups("", 5)
		fmt.Printf("Backup created: %s\n", path)

	case "list":
		backups, err := database.ListBackups("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(backups) == 0 {
			fmt.Println("No backups.")
			return
		}
		fmt.Printf("%-40s %-10s %-20s\n", "NAME", "SIZE", "DATE")
		fmt.Println(strings.Repeat("-", 72))
		for _, b := range backups {
			fmt.Printf("%-40s %-10s %-20s\n",
				b.Name, formatBytes(b.Size), b.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
}

// ─── helpers ────────────────────────────────────

func truncate(s string, max int) string {
	if len(s) <= max { return s }
	return s[:max-3] + "..."
}

func formatBytes(b int64) string {
	if b < 1024 { return fmt.Sprintf("%d B", b) }
	if b < 1024*1024 { return fmt.Sprintf("%.1f KB", float64(b)/1024) }
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}
