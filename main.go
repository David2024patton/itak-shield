package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/David2024patton/itak-shield/proxy"
)

var version = "0.1.0"

func main() {
	target := flag.String("target", "", "Upstream API URL (e.g. https://api.openai.com)")
	port := flag.Int("port", 0, "Local port to listen on (default: random 5-digit port)")
	verbose := flag.Bool("verbose", false, "Log redaction details")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("GOProxy v%s\n", version)
		os.Exit(0)
	}

	if *target == "" {
		fmt.Fprintln(os.Stderr, "Usage: itak-shield --target https://api.openai.com [--port 8080] [--verbose]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "iTaK Shield sits between your AI tools and cloud APIs,")
		fmt.Fprintln(os.Stderr, "replacing sensitive data with safe placeholders before")
		fmt.Fprintln(os.Stderr, "it ever leaves your machine.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Use random 5-digit port if not specified.
	if *port == 0 {
		*port = 10000 + rand.Intn(55535)
	}

	p, err := proxy.New(*target, *verbose)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           iTaK Shield v" + version + "                │")
	fmt.Println("│         Privacy-First LLM Proxy             │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│  Listening:  http://%s        │\n", addr)
	fmt.Printf("│  Upstream:   %-30s │\n", *target)
	fmt.Printf("│  Verbose:    %-30v │\n", *verbose)
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  All PII is redacted before leaving your    │")
	fmt.Println("│  machine. Token map lives in memory only.   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	log.Printf("iTaK Shield listening on %s -> %s", addr, *target)

	if err := http.ListenAndServe(addr, p); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
