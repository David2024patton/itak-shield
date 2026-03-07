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
	"github.com/David2024patton/itak-shield/auth"
	"github.com/David2024patton/itak-shield/cache"
	"github.com/David2024patton/itak-shield/config"
	"github.com/David2024patton/itak-shield/dlp"
	"github.com/David2024patton/itak-shield/proxy"
	"github.com/David2024patton/itak-shield/retry"
	"github.com/David2024patton/itak-shield/server"
	"github.com/David2024patton/itak-shield/spend"
)

var version = "0.2.0"

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

	// ─── Load Configuration ──────────────────
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

	// ─── CLI Mode: target provided ───────────
	if cfg.Target != "" {
		runCLI(cfg, *port)
		return
	}

	// ─── GUI Mode: no target ─────────────────
	if *noGUI {
		fmt.Fprintln(os.Stderr, "Usage: itak-shield --target https://api.openai.com [--port 20979] [--verbose]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or run without flags to launch the interactive GUI.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Enterprise: itak-shield --config shield.yaml")
		flag.PrintDefaults()
		os.Exit(1)
	}

	runGUI(*guiPort, *bind)
}

// buildProxyOptions constructs the full set of proxy options from config.
func buildProxyOptions(cfg *config.Config) ([]proxy.Option, *audit.Logger) {
	var opts []proxy.Option

	// Audit logger.
	var auditLogger *audit.Logger
	if cfg.Audit.Enabled {
		var err error
		auditLogger, err = audit.New(cfg.Audit.Path, cfg.Audit.MaxSizeMB, cfg.Audit.MaxFiles)
		if err != nil {
			log.Fatalf("Failed to initialize audit logger: %v", err)
		}
		opts = append(opts, proxy.WithAuditLogger(auditLogger))
		log.Printf("[iTaK Shield] Audit logging to %s", cfg.Audit.Path)
	}

	// Custom PII rules.
	for _, rule := range cfg.Rules.Custom {
		opts = append(opts, proxy.WithCustomRule(rule.Name, rule.Pattern))
	}
	for _, disabled := range cfg.Rules.Disabled {
		opts = append(opts, proxy.WithDisabledRule(disabled))
	}

	// Auth (virtual API keys + rate limiting).
	if cfg.Auth.Enabled && len(cfg.Auth.Keys) > 0 {
		entries := make([]auth.KeyEntry, len(cfg.Auth.Keys))
		for i, k := range cfg.Auth.Keys {
			entries[i] = auth.KeyEntry{
				Key:       k.Key,
				User:      k.User,
				Group:     k.Group,
				RateLimit: k.RateLimit,
			}
		}
		mgr := auth.NewFromEntries(entries, cfg.Auth.InjectKey)
		opts = append(opts, proxy.WithAuth(mgr))
		log.Printf("[iTaK Shield] Auth enabled: %d virtual API keys", len(entries))
	}

	// Cache.
	if cfg.Cache.Enabled {
		c := cache.New(cfg.Cache.MaxEntries, cfg.Cache.TTLSeconds)
		if c != nil {
			opts = append(opts, proxy.WithCache(c))
			log.Printf("[iTaK Shield] Response cache enabled: %d entries, %ds TTL", cfg.Cache.MaxEntries, cfg.Cache.TTLSeconds)
		}
	}

	// Retry + fallback.
	if cfg.Retry.Enabled {
		retryCfg := &retry.Config{
			MaxRetries:      cfg.Retry.MaxRetries,
			BackoffMs:       cfg.Retry.BackoffMs,
			FallbackTargets: cfg.Retry.FallbackTargets,
		}
		opts = append(opts, proxy.WithRetry(retryCfg))
		log.Printf("[iTaK Shield] Retry enabled: %d retries, %dms backoff, %d fallback targets",
			cfg.Retry.MaxRetries, cfg.Retry.BackoffMs, len(cfg.Retry.FallbackTargets))
	}

	// Spend tracking + budgets.
	if cfg.Spend.Enabled {
		pricing := spend.Pricing{
			Input:  cfg.Spend.Pricing.Input,
			Output: cfg.Spend.Pricing.Output,
		}
		tracker := spend.New(cfg.Spend.Budgets, pricing)
		opts = append(opts, proxy.WithSpend(tracker))
		log.Printf("[iTaK Shield] Spend tracking enabled: %d budget groups", len(cfg.Spend.Budgets))
	}

	// DLP policies.
	if len(cfg.DLP.Policies) > 0 {
		policy := dlp.New(cfg.DLP.Policies)
		if policy != nil {
			opts = append(opts, proxy.WithDLP(policy))
			blockCount := 0
			for _, action := range cfg.DLP.Policies {
				if strings.EqualFold(action, "block") {
					blockCount++
				}
			}
			log.Printf("[iTaK Shield] DLP enabled: %d policies (%d block, %d redact)",
				len(cfg.DLP.Policies), blockCount, len(cfg.DLP.Policies)-blockCount)
		}
	}

	return opts, auditLogger
}

// runCLI starts the proxy in headless CLI mode.
func runCLI(cfg *config.Config, port int) {
	if port == 0 {
		port = 10000 + rand.Intn(55535)
	}

	opts, auditLogger := buildProxyOptions(cfg)
	if auditLogger != nil {
		defer auditLogger.Close()
	}

	p, err := proxy.New(cfg.Target, cfg.Verbose, opts...)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	listenHost := cfg.Listen
	addr := fmt.Sprintf("%s:%d", listenHost, port)

	fmt.Println("+---------------------------------------------+")
	fmt.Println("|           iTaK Shield v" + version + "                |")
	fmt.Println("|         Privacy-First LLM Proxy             |")
	fmt.Println("+---------------------------------------------+")
	fmt.Printf("|  Listening:  http://%-24s |\n", addr)
	fmt.Printf("|  Upstream:   %-30s |\n", cfg.Target)
	fmt.Printf("|  Verbose:    %-30v |\n", cfg.Verbose)
	if cfg.Audit.Enabled {
		fmt.Printf("|  Audit Log:  %-30s |\n", cfg.Audit.Path)
	}
	if len(cfg.Rules.Custom) > 0 {
		fmt.Printf("|  Custom Rules: %-28d |\n", len(cfg.Rules.Custom))
	}
	if len(cfg.Rules.Disabled) > 0 {
		fmt.Printf("|  Disabled:   %-30s |\n", strings.Join(cfg.Rules.Disabled, ", "))
	}
	if cfg.Auth.Enabled {
		fmt.Printf("|  Auth Keys:  %-30d |\n", len(cfg.Auth.Keys))
	}
	if cfg.Cache.Enabled {
		fmt.Printf("|  Cache:      %-30s |\n", fmt.Sprintf("%d entries, %ds TTL", cfg.Cache.MaxEntries, cfg.Cache.TTLSeconds))
	}
	if cfg.Retry.Enabled {
		fmt.Printf("|  Retry:      %-30s |\n", fmt.Sprintf("%d retries, %d fallbacks", cfg.Retry.MaxRetries, len(cfg.Retry.FallbackTargets)))
	}
	if cfg.Spend.Enabled {
		fmt.Printf("|  Budgets:    %-30d |\n", len(cfg.Spend.Budgets))
	}
	if len(cfg.DLP.Policies) > 0 {
		fmt.Printf("|  DLP Rules:  %-30d |\n", len(cfg.DLP.Policies))
	}
	fmt.Println("+---------------------------------------------+")
	fmt.Println("|  All PII is redacted before leaving your    |")
	fmt.Println("|  machine. Token map lives in memory only.   |")
	fmt.Println("+---------------------------------------------+")
	fmt.Println()

	log.Printf("iTaK Shield listening on %s -> %s", addr, cfg.Target)

	if err := http.ListenAndServe(addr, p); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// runGUI starts the interactive web GUI.
func runGUI(guiPort int, bindAddr string) {
	if guiPort == 0 {
		guiPort = 10000 + rand.Intn(55535)
	}
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	guiAddr := fmt.Sprintf("http://%s:%d", bindAddr, guiPort)

	fmt.Println("+---------------------------------------------+")
	fmt.Println("|           iTaK Shield v" + version + "                |")
	fmt.Println("|         Privacy-First LLM Proxy             |")
	fmt.Println("+---------------------------------------------+")
	fmt.Printf("|  GUI:  %-37s |\n", guiAddr)
	fmt.Println("|                                             |")
	if bindAddr == "0.0.0.0" {
		fmt.Println("|  Network mode: accessible from other hosts  |")
		fmt.Println("|  Open the URL above in your browser.        |")
	} else {
		fmt.Println("|  Opening your browser...                    |")
		fmt.Println("|  If it doesn't open, visit the URL above.   |")
	}
	fmt.Println("+---------------------------------------------+")
	fmt.Println("|  For CLI mode, use:                         |")
	fmt.Println("|  itak-shield --target https://api.openai.com|")
	fmt.Println("|                                             |")
	fmt.Println("|  Enterprise: itak-shield --config shield.yml|")
	fmt.Println("+---------------------------------------------+")
	fmt.Println()

	// Open the browser (skip in network/Docker mode).
	if bindAddr != "0.0.0.0" {
		go openBrowser(fmt.Sprintf("http://127.0.0.1:%d", guiPort))
	}

	gui := server.NewGUI(version, bindAddr)
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
