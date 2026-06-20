package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static
var staticFS embed.FS

// Server is the local HTTP dashboard server.
type Server struct {
	addr            string
	lastTestsPassed int
	lastTestsTotal  int
	lastTestRun     string
}

// NewServer creates a dashboard server bound to addr (e.g. ":8080").
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// Start runs the HTTP server (blocks).
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/dashboard", s.handleDashboard)
	mux.HandleFunc("/api/files", s.handleFiles)
	mux.HandleFunc("/api/files/", s.handleFileDetail)
	mux.HandleFunc("/api/encrypt", s.handleEncrypt)
	mux.HandleFunc("/api/decrypt", s.handleDecrypt)
	mux.HandleFunc("/api/verify-password", s.handleVerifyPassword)
	mux.HandleFunc("/api/setup-password", s.handleSetupPassword)
	mux.HandleFunc("/api/debug", s.handleDebug)
	mux.HandleFunc("/api/browse", s.handleBrowse)
	mux.HandleFunc("/api/jobs/", s.handleJob)
	mux.HandleFunc("/api/reviewer", s.handleReviewer)
	mux.HandleFunc("/api/reviewer/run-tests", s.handleRunTests)

	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	s.registerStatic(mux, http.FS(static))

	mux.HandleFunc("/", s.handleSPA)

	fmt.Printf("Encrypt-O-Matic Web UI running at http://localhost%s\n", s.addr)
	fmt.Println("Educational dashboard only — CLI remains the primary interface.")
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "UI not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
