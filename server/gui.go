package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/David2024patton/itak-shield/proxy"
)

// GUIServer serves the embedded web UI and manages the proxy lifecycle.
type GUIServer struct {
	mu          sync.Mutex
	proxy       *proxy.Server
	proxyAddr   string
	proxyTarget string
	proxyPort   int
	running     bool
	startTime   time.Time
	listener    net.Listener
	stopChan    chan struct{}
	version     string
}

// LogEntry represents a single activity log entry for the GUI.
type LogEntry struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// StartRequest is the JSON body for POST /api/start.
type StartRequest struct {
	Target  string `json:"target"`
	Port    int    `json:"port"`
	Verbose bool   `json:"verbose"`
}

// StatusResponse is the JSON body for GET /api/status.
type StatusResponse struct {
	Running       bool       `json:"running"`
	Target        string     `json:"target"`
	Port          int        `json:"port"`
	Requests      int64      `json:"requests"`
	Redacted      int64      `json:"redacted"`
	UptimeSeconds int64      `json:"uptime_seconds"`
	RecentLogs    []LogEntry `json:"recent_logs"`
}

// NewGUI creates a new GUI server.
func NewGUI(version string) *GUIServer {
	return &GUIServer{version: version}
}

// Serve starts the GUI server on the given port, serving the embedded web content.
func (g *GUIServer) Serve(webFS embed.FS, guiPort int) error {
	// Extract the web/ subdirectory from the embedded FS.
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/start", g.handleStart)
	mux.HandleFunc("/api/stop", g.handleStop)
	mux.HandleFunc("/api/status", g.handleStatus)
	mux.HandleFunc("/api/providers", g.handleProviders)
	mux.HandleFunc("/healthz", g.handleHealth)

	// Static files (the embedded web UI)
	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle("/", fileServer)

	addr := fmt.Sprintf("127.0.0.1:%d", guiPort)
	log.Printf("iTaK Shield GUI ready at http://%s", addr)

	return http.ListenAndServe(addr, mux)
}

// handleStart starts the privacy proxy with the given configuration.
func (g *GUIServer) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "Invalid request body"})
		return
	}

	if req.Target == "" {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "Target URL is required"})
		return
	}
	if req.Port < 1024 || req.Port > 65535 {
		req.Port = 8080
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Stop existing proxy if running.
	if g.running {
		g.stopLocked()
	}

	// Create new proxy.
	p, err := proxy.New(req.Target, req.Verbose)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	// Bind to the proxy port.
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", req.Port)
	ln, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": fmt.Sprintf("Port %d is already in use", req.Port)})
		return
	}

	g.proxy = p
	g.proxyAddr = proxyAddr
	g.proxyTarget = req.Target
	g.proxyPort = req.Port
	g.running = true
	g.startTime = time.Now()
	g.listener = ln
	g.stopChan = make(chan struct{})

	// Run the proxy in a goroutine.
	go func() {
		srv := &http.Server{Handler: p}
		go func() {
			<-g.stopChan
			srv.Close()
		}()
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[iTaK Shield] Proxy error: %v", err)
		}
		g.mu.Lock()
		g.running = false
		g.mu.Unlock()
	}()

	log.Printf("[iTaK Shield] Proxy started on %s -> %s", proxyAddr, req.Target)

	writeJSON(w, map[string]interface{}{
		"ok":   true,
		"addr": proxyAddr,
	})
}

// handleStop stops the running proxy.
func (g *GUIServer) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.running {
		g.stopLocked()
		log.Printf("[iTaK Shield] Proxy stopped")
	}

	writeJSON(w, map[string]interface{}{"ok": true})
}

// stopLocked stops the proxy. Must be called with g.mu held.
func (g *GUIServer) stopLocked() {
	if g.stopChan != nil {
		close(g.stopChan)
	}
	g.running = false
}

// handleStatus returns the current proxy status and stats.
func (g *GUIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()

	resp := StatusResponse{
		Running: g.running,
		Target:  g.proxyTarget,
		Port:    g.proxyPort,
	}

	if g.running && g.proxy != nil {
		stats := g.proxy.LiveStats()
		resp.Requests = stats.Requests
		resp.Redacted = stats.Redacted
		resp.UptimeSeconds = int64(time.Since(g.startTime).Seconds())
		resp.RecentLogs = convertLogs(stats.RecentLogs)
	}

	writeJSON(w, resp)
}

// handleProviders returns the list of preset providers.
func (g *GUIServer) handleProviders(w http.ResponseWriter, r *http.Request) {
	providers := []map[string]string{
		{"id": "openai", "name": "OpenAI", "url": "https://api.openai.com"},
		{"id": "anthropic", "name": "Anthropic", "url": "https://api.anthropic.com"},
		{"id": "gemini", "name": "Google Gemini", "url": "https://generativelanguage.googleapis.com"},
	}
	writeJSON(w, providers)
}

// handleHealth returns a health check response for load balancers and k8s probes.
func (g *GUIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	var uptimeSeconds int64
	if g.running {
		uptimeSeconds = int64(time.Since(g.startTime).Seconds())
	}
	g.mu.Unlock()

	writeJSON(w, map[string]interface{}{
		"status":         "ok",
		"version":        g.version,
		"uptime_seconds": uptimeSeconds,
	})
}

// convertLogs converts proxy log entries to GUI log entries.
func convertLogs(logs []proxy.LogEntry) []LogEntry {
	result := make([]LogEntry, 0, len(logs))
	for _, l := range logs {
		result = append(result, LogEntry{
			Time:    l.Time.Format("15:04:05"),
			Type:    l.Type,
			Message: l.Message,
		})
	}
	return result
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
