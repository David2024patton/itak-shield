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
	flag.Parse()

	if *showVersion {
		fmt.Printf("iTaK Shield v%s\n", version)
		os.Exit(0)
	}

	// ─── CLI Mode: --target flag provided ────────
	if *target != "" {
		runCLI(*target, *port, *verbose)
		return
	}

	// ─── GUI Mode: no --target flag ──────────────
	if *noGUI {
		fmt.Fprintln(os.Stderr, "Usage: itak-shield --target https://api.openai.com [--port 8080] [--verbose]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or run without flags to launch the interactive GUI.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	runGUI(*guiPort)
}

// runCLI starts the proxy in headless CLI mode (original behavior).
func runCLI(target string, port int, verbose bool) {
	if port == 0 {
		port = 10000 + rand.Intn(55535)
	}

	p, err := proxy.New(target, verbose)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           iTaK Shield v" + version + "                │")
	fmt.Println("│         Privacy-First LLM Proxy             │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│  Listening:  http://%s        │\n", addr)
	fmt.Printf("│  Upstream:   %-30s │\n", target)
	fmt.Printf("│  Verbose:    %-30v │\n", verbose)
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  All PII is redacted before leaving your    │")
	fmt.Println("│  machine. Token map lives in memory only.   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	log.Printf("iTaK Shield listening on %s -> %s", addr, target)

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
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	// Open the browser.
	go openBrowser(guiAddr)

	gui := server.NewGUI()
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
