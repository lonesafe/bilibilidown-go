package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"bilibilidown-go/internal/api"
	"bilibilidown-go/internal/config"
	"bilibilidown-go/internal/downloader"
	"bilibilidown-go/internal/parser"
	"bilibilidown-go/internal/util"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server is the HTTP server
type Server struct {
	api      *api.BilibiliAPI
	parser   *parser.InputParser
	dl       *downloader.Downloader
	hub      *Hub
}

// Hub manages WebSocket connections
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func newHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool)}
}

func (h *Hub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

func (h *Hub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) broadcast(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

// NewServer creates a new HTTP server
func NewServer() *Server {
	bapi := api.NewBilibiliAPI()
	return &Server{
		api:    bapi,
		parser: parser.NewInputParser(bapi),
		dl:     downloader.NewDownloader(),
		hub:    newHub(),
	}
}

// RegisterRoutes sets up all HTTP routes
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/parse", s.handleParse)
	mux.HandleFunc("/api/download", s.handleDownload)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/stop", s.handleStopTask)
	mux.HandleFunc("/api/login/status", s.handleLoginStatus)
	mux.HandleFunc("/api/login/cookie", s.handleSetCookie)
	mux.HandleFunc("/api/login/qrcode", s.handleQRCode)
	mux.HandleFunc("/api/login/qrcode/poll", s.handleQRCodePoll)
	mux.HandleFunc("/api/user", s.handleUserInfo)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/ws", s.handleWebSocket)

	// Serve frontend - try multiple possible locations
	distPath := findFrontendDist()
	if distPath != "" {
		fs := http.FileServer(http.Dir(distPath))
		mux.Handle("/", spaHandler(fs, distPath))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("前端未构建，请在 frontend 目录执行 npm run build"))
		})
	}

	return mux
}

// findFrontendDist locates the frontend dist directory
func findFrontendDist() string {
	candidates := []string{
		"./frontend/dist",
		"../frontend/dist",
	}
	// Also check relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "frontend", "dist"),
			filepath.Join(exeDir, "..", "frontend", "dist"),
		)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// spaHandler serves SPA (single page app) - falls back to index.html for unknown routes
func spaHandler(fs http.Handler, distPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try the file first
		fullPath := filepath.Join(distPath, r.URL.Path)
		if r.URL.Path != "/" {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				// Not a real file, serve index.html for SPA routing
				http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
				return
			}
		}
		fs.ServeHTTP(w, r)
	})
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "方法不允许")
		return
	}
	var req struct {
		URL        string `json:"url"`
		QN         int    `json:"qn"`
		FetchLinks bool   `json:"fetchLinks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "请求格式错误")
		return
	}

	video, err := s.parser.GetResult(req.URL, req.QN, req.FetchLinks)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, video)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "方法不允许")
		return
	}
	var req struct {
		AvID  string `json:"avId"`
		CID   int64  `json:"cId"`
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "请求格式错误")
		return
	}

	task := s.dl.AddTask(req.AvID, req.CID, req.Title, req.URL)
	writeJSON(w, map[string]string{"id": task.ID})

	// Broadcast task updates
	go s.broadcastTaskUpdates()
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.dl.GetAllTasks()
	result := make([]downloader.TaskSnapshot, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, t.Snapshot())
	}
	writeJSON(w, result)
}

func (s *Server) handleStopTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "方法不允许")
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "请求格式错误")
		return
	}
	s.dl.StopTask(req.ID)
	writeJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleLoginStatus(w http.ResponseWriter, r *http.Request) {
	loggedIn := s.api.IsLoggedIn()
	writeJSON(w, map[string]interface{}{"loggedIn": loggedIn})
}

func (s *Server) handleSetCookie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "方法不允许")
		return
	}
	var req struct {
		Cookie string `json:"cookie"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "请求格式错误")
		return
	}
	util.SetGlobalCookies(req.Cookie)
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleQRCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "方法不允许")
		return
	}
	result, err := s.api.GenQRCode()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleQRCodePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "方法不允许")
		return
	}
	var req struct {
		QrcodeKey string `json:"qrcodeKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "请求格式错误")
		return
	}
	result, err := s.api.PollQRCode(req.QrcodeKey)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	// If login successful, set cookies globally
	if result.Code == 0 && result.CookieStr != "" {
		util.SetGlobalCookies(result.CookieStr)
	}
	writeJSON(w, result)
}

func (s *Server) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	name, mid, avatar, err := s.api.GetUserInfo()
	if err != nil {
		writeError(w, 500, "获取用户信息失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{
		"name":   name,
		"mid":    mid,
		"avatar": avatar,
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, config.Get())
		return
	}
	if r.Method == http.MethodPost {
		var req config.Config
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "请求格式错误")
			return
		}
		cfg := config.Get()
		if req.SavePath != "" {
			cfg.SavePath = req.SavePath
		}
		if req.FFmpegPath != "" {
			cfg.FFmpegPath = req.FFmpegPath
		}
		if req.DownloadPoolSize > 0 {
			cfg.DownloadPoolSize = req.DownloadPoolSize
		}
		writeJSON(w, cfg)
		return
	}
	writeError(w, 405, "方法不允许")
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.hub.register(conn)
	defer s.hub.unregister(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) broadcastTaskUpdates() {
	// Periodically broadcast task status
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		tasks := s.dl.GetAllTasks()
		snapshots := make([]downloader.TaskSnapshot, 0, len(tasks))
		allDone := true
		for _, t := range tasks {
			snap := t.Snapshot()
			snapshots = append(snapshots, snap)
			if snap.Status == string(downloader.StatusDownloading) || snap.Status == string(downloader.StatusMerging) {
				allDone = false
			}
		}
		s.hub.broadcast(map[string]interface{}{"type": "tasks", "data": snapshots})
		if allDone {
			break
		}
	}
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ParseQN parses quality string to int
func ParseQN(qn string) int {
	qnMap := map[string]int{
		"8K":   127, "HDR": 126, "4K": 120, "1080P60": 116,
		"1080P+": 112, "1080P": 80, "720P60": 74, "720P": 64,
		"480P": 32, "360P": 16,
	}
	if v, ok := qnMap[strings.ToUpper(qn)]; ok {
		return v
	}
	n, _ := strconv.Atoi(qn)
	return n
}
