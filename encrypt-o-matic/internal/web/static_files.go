package web

import (
	"io"
	"net/http"
	"path"
	"strings"
)

func (s *Server) registerStatic(mux *http.ServeMux, static http.FileSystem) {
	// Explicit routes for known assets (reliable across platforms).
	mux.HandleFunc("/assets/css/app.css", func(w http.ResponseWriter, r *http.Request) {
		serveEmbeddedFile(w, static, "css/app.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("/assets/js/edu-content.js", func(w http.ResponseWriter, r *http.Request) {
		serveEmbeddedFile(w, static, "js/edu-content.js", "application/javascript; charset=utf-8")
	})
	mux.HandleFunc("/assets/js/app.js", func(w http.ResponseWriter, r *http.Request) {
		serveEmbeddedFile(w, static, "js/app.js", "application/javascript; charset=utf-8")
	})

	// Fallback for any other static files under /assets/
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(static)))
}

func serveEmbeddedFile(w http.ResponseWriter, static http.FileSystem, name, contentType string) {
	name = path.Clean(strings.ReplaceAll(name, "\\", "/"))
	f, err := static.Open(name)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = io.Copy(w, f)
}
