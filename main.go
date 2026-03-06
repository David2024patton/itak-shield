package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/David2024patton/itak-shield/audit"
	"github.com/David2024patton/itak-shield/config"
	"github.com/David2024patton/itak-shield/proxy"
	"github.com/David2024patton/itak-shield/server"
)

var version = "0.1.0"

//go:embed web/*
var webFS embed.FS

func main() {
	target := flag.String("target", "", "Upstream API URL (e.g. https://api.openai.com)")
	port := flag.Int("port", 0, "Local port to listen on (default: random 5-digit port)")
	verbose := flag.Bool("verbose", false, "Log redaction details")
	showVersion := flag.Bool("version", false, "Print version and exit")
	guiPort := flag.Int("gui-port", 0, "Port for the GUI (default: random 5-digit port)")
	noGUI := flag.Bool("no-gui", false, "Disable GUI mode even when no --target is given")
	bind := flag.String("bind", "", "Bind address (default: 127.0.0.1, use 0.0.0.0 for network)")
	configPath := flag.String("config", "", "Path to YAML config file (optional)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("iTaK Shield v%s\n", version)
		os.Exit(0)
	}

	// ─── Load Configuration ──────────────────────
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// CLI flags override config file values.
	if *target != "" {
		cfg.Target = *target
	}
	if *bind != "" {
		cfg.Listen = *bind
	}
	if *verbose {
		cfg.Verbose = true
	}

	// ─── Initialize Audit Logger ─────────────────
	var auditLogger *audit.Logger
	if cfg.Audit.Enabled {
		auditLogger, err = audit.New(cfg.Audit.Path, cfg.Audit.MaxSizeMB, cfg.Audit.MaxFiles)
		if err != nil {
			log.Fatalf("Failed to initialize audit logger: %v", err)
		}
		defer auditLogger.Close()
		log.Printf("[iTaK Shield] Audit logging to %s", cfg.Audit.Path)
	}

	// ─── CLI Mode: target provided ───────────────
	if cfg.Target != "" {
		runCLI(cfg, auditLogger, *port)
		return
	}

	// ─── GUI Mode: no target ─────────────────────
	if *noGUI {
		fmt.Fprintln(os.Stderr, "Usage: itak-shield --target https://api.openai.com [--port 8080] [--verbose]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or run without flags to launch the interactive GUI.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Enterprise: itak-shield --config shield.yaml")
		flag.PrintDefaults()
		os.Exit(1)
	}

	runGUI(*guiPort)
}

// runCLI starts the proxy in headless CLI mode.
func runCLI(cfg *config.Config, auditLogger *audit.Logger, port int) {
	if port == 0 {
		port = 10000 + rand.Intn(55535)
	}

	// Build scanner options from config.
	var opts []proxy.Option
	for _, rule := range cfg.Rules.Custom {
		opts = append(opts, proxy.WithCustomRule(rule.Name, rule.Pattern))
	}
	for _, disabled := range cfg.Rules.Disabled {
		opts = append(opts, proxy.WithDisabledRule(disabled))
	}
	if auditLogger != nil {
		opts = append(opts, proxy.WithAuditLogger(auditLogger))
	}

	p, err := proxy.New(cfg.Target, cfg.Verbose, opts...)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	listenHost := cfg.Listen
	addr := fmt.Sprintf("%s:%d", listenHost, port)

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           iTaK Shield v" + version + "                │")
	fmt.Println("│         Privacy-First LLM Proxy             │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│  Listening:  http://%-24s │\n", addr)
	fmt.Printf("│  Upstream:   %-30s │\n", cfg.Target)
	fmt.Printf("│  Verbose:    %-30v │\n", cfg.Verbose)
	if cfg.Audit.Enabled {
		fmt.Printf("│  Audit Log:  %-30s │\n", cfg.Audit.Path)
	}
	if len(cfg.Rules.Custom) > 0 {
		fmt.Printf("│  Custom Rules: %-28d │\n", len(cfg.Rules.Custom))
	}
	if len(cfg.Rules.Disabled) > 0 {
		fmt.Printf("│  Disabled:   %-30s │\n", strings.Join(cfg.Rules.Disabled, ", "))
	}
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  All PII is redacted before leaving your    │")
	fmt.Println("│  machine. Token map lives in memory only.   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	log.Printf("iTaK Shield listening on %s -> %s", addr, cfg.Target)

	if err := http.ListenAndServe(addr, p); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// runGUI starts the interactive web GUI.
func runGUI(guiPort int) {
	if guiPort == 0 {
		guiPort = 10000 + rand.Intn(55535)
	}

	guiAddr := fmt.Sprintf("http://127.0.0.1:%d", guiPort)

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           iTaK Shield v" + version + "                │")
	fmt.Println("│         Privacy-First LLM Proxy             │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│  GUI:  %-37s │\n", guiAddr)
	fmt.Println("│                                             │")
	fmt.Println("│  Opening your browser...                    │")
	fmt.Println("│  If it doesn't open, visit the URL above.   │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  For CLI mode, use:                         │")
	fmt.Println("│  itak-shield --target https://api.openai.com│")
	fmt.Println("│                                             │")
	fmt.Println("│  Enterprise: itak-shield --config shield.yml│")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	// Open the browser.
	go openBrowser(guiAddr)

	gui := server.NewGUI(version)
	if err := gui.Serve(webFS, guiPort); err != nil {
		log.Fatalf("GUI server error: %v", err)
	}
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
