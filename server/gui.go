package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/David2024patton/itak-shield/auth"
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
	bindAddr    string
	authMgr     *auth.Manager
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
	Running       bool            `json:"running"`
	Target        string          `json:"target"`
	Port          int             `json:"port"`
	Requests      int64           `json:"requests"`
	Redacted      int64           `json:"redacted"`
	UptimeSeconds int64           `json:"uptime_seconds"`
	RecentLogs    []LogEntry      `json:"recent_logs"`
	Features      map[string]bool `json:"features,omitempty"`
}

// NewGUI creates a new GUI server.
func NewGUI(version string, bindAddr string) *GUIServer {
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	// Initialize auth manager with persistent file store.
	store := auth.NewFileStore(filepath.Join(".", "shield-users.json"))
	authMgr := auth.New(store, "")

	return &GUIServer{version: version, bindAddr: bindAddr, authMgr: authMgr}
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
	mux.HandleFunc("/api/analytics", g.handleAnalytics)
	mux.HandleFunc("/api/providers", g.handleProviders)
	mux.HandleFunc("/healthz", g.handleHealth)

	// User & token management
	mux.HandleFunc("/api/users", g.handleUsers)
	mux.HandleFunc("/api/users/", g.handleUserByID)
	mux.HandleFunc("/api/tokens", g.handleTokens)
	mux.HandleFunc("/api/tokens/revoke", g.handleRevokeToken)

	// Static files (the embedded web UI)
	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle("/", fileServer)

	addr := fmt.Sprintf("%s:%d", g.bindAddr, guiPort)
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
		req.Port = 20979
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
	proxyAddr := fmt.Sprintf("%s:%d", g.bindAddr, req.Port)
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
		resp.Features = map[string]bool{
			"auth":  stats.Features.Auth,
			"cache": stats.Features.Cache,
			"retry": stats.Features.Retry,
			"spend": stats.Features.Spend,
			"dlp":   stats.Features.DLP,
		}
	}

	writeJSON(w, resp)
}

// handleAnalytics returns enterprise feature analytics data.
func (g *GUIServer) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running || g.proxy == nil {
		writeJSON(w, map[string]interface{}{"active": false})
		return
	}

	stats := g.proxy.LiveStats()
	result := map[string]interface{}{
		"active":   true,
		"features": stats.Features,
	}

	if stats.CacheStats != nil {
		result["cache"] = stats.CacheStats
	}
	if stats.SpendStats != nil {
		result["spend"] = stats.SpendStats
	}
	if stats.AuthUsers != nil {
		result["auth_users"] = stats.AuthUsers
	}

	writeJSON(w, result)
}

// handleProviders returns the list of preset providers.
func (g *GUIServer) handleProviders(w http.ResponseWriter, r *http.Request) {
	providers := []map[string]string{
		// Foundation Models
		{"id": "openai", "name": "OpenAI", "url": "https://api.openai.com"},
		{"id": "anthropic", "name": "Anthropic", "url": "https://api.anthropic.com"},
		{"id": "gemini", "name": "Google Gemini", "url": "https://generativelanguage.googleapis.com"},
		{"id": "xai", "name": "xAI (Grok)", "url": "https://api.x.ai"},
		{"id": "deepseek", "name": "DeepSeek", "url": "https://api.deepseek.com"},
		{"id": "mistral", "name": "Mistral AI", "url": "https://api.mistral.ai"},
		{"id": "cohere", "name": "Cohere", "url": "https://api.cohere.com"},
		{"id": "nvidia", "name": "NVIDIA NIM", "url": "https://integrate.api.nvidia.com"},
		{"id": "qwen", "name": "Qwen (Alibaba)", "url": "https://dashscope.aliyuncs.com/compatible-mode"},
		{"id": "kimi", "name": "Kimi (Moonshot)", "url": "https://api.moonshot.cn"},
		{"id": "zhipu", "name": "Zhipu AI (GLM)", "url": "https://open.bigmodel.cn/api/paas"},
		{"id": "meta", "name": "Meta AI (Llama)", "url": "https://api.llama.com"},

		// API Gateways
		{"id": "openrouter", "name": "OpenRouter", "url": "https://openrouter.ai/api"},
		{"id": "groq", "name": "Groq", "url": "https://api.groq.com/openai"},
		{"id": "together", "name": "Together AI", "url": "https://api.together.xyz"},
		{"id": "fireworks", "name": "Fireworks AI", "url": "https://api.fireworks.ai/inference"},
		{"id": "huggingface", "name": "Hugging Face", "url": "https://api-inference.huggingface.co"},
		{"id": "deepinfra", "name": "DeepInfra", "url": "https://api.deepinfra.com/v1/openai"},
		{"id": "siliconflow", "name": "SiliconFlow", "url": "https://api.siliconflow.cn"},

		// Specialized
		{"id": "perplexity", "name": "Perplexity", "url": "https://api.perplexity.ai"},
		{"id": "cerebras", "name": "Cerebras", "url": "https://api.cerebras.ai"},

		// Local / Self-Hosted
		{"id": "ollama", "name": "Ollama", "url": "http://localhost:11434"},
		{"id": "lmstudio", "name": "LM Studio", "url": "http://localhost:1234/v1"},
		{"id": "llamacpp", "name": "Llama.cpp", "url": "http://localhost:20979"},
		{"id": "vllm", "name": "vLLM", "url": "http://localhost:8000/v1"},
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

// ── User & Token Management Handlers ────────────────────────────

// handleUsers handles GET (list) and POST (create) for /api/users.
func (g *GUIServer) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users := g.authMgr.ListUsers()
		if users == nil {
			users = []auth.User{}
		}
		writeJSON(w, users)

	case http.MethodPost:
		var req struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			Group     string `json:"group"`
			RateLimit int    `json:"rate_limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "Invalid request body"})
			return
		}
		if req.Name == "" {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "Name is required"})
			return
		}
		if req.Group == "" {
			req.Group = "default"
		}

		user, err := g.authMgr.CreateUser(req.Name, req.Email, req.Group, req.RateLimit)
		if err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}

		log.Printf("[iTaK Shield] User created: %s (%s)", user.Name, user.ID)
		writeJSON(w, map[string]interface{}{"ok": true, "user": user})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleUserByID handles GET and DELETE for /api/users/{id}.
func (g *GUIServer) handleUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path: /api/users/{id}
	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if userID == "" {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "User ID required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := g.authMgr.GetUser(userID)
		if err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, user)

	case http.MethodDelete:
		if err := g.authMgr.DeleteUser(userID); err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		log.Printf("[iTaK Shield] User deleted: %s", userID)
		writeJSON(w, map[string]interface{}{"ok": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTokens generates a new API token for a user (POST).
func (g *GUIServer) handleTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    string `json:"user_id"`
		Label     string `json:"label"`
		ExpiresIn *int   `json:"expires_in"` // hours, nil = never
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "Invalid request body"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "user_id is required"})
		return
	}
	if req.Label == "" {
		req.Label = "api-key"
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &t
	}

	token, err := g.authMgr.GenerateToken(req.UserID, req.Label, expiresAt)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	log.Printf("[iTaK Shield] Token generated for user %s: %s", req.UserID, req.Label)
	writeJSON(w, map[string]interface{}{"ok": true, "token": token})
}

// handleRevokeToken revokes a specific token (POST).
func (g *GUIServer) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		TokenKey string `json:"token_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "Invalid request body"})
		return
	}
	if req.UserID == "" || req.TokenKey == "" {
		writeJSON(w, map[string]interface{}{"ok": false, "error": "user_id and token_key are required"})
		return
	}

	if err := g.authMgr.RevokeToken(req.UserID, req.TokenKey); err != nil {
		writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	log.Printf("[iTaK Shield] Token revoked for user %s", req.UserID)
	writeJSON(w, map[string]interface{}{"ok": true})
}
