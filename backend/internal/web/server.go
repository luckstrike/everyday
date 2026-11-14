package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

func New(api http.Handler, frontend fs.FS) (http.Handler, error) {
	index, err := frontend.Open("index.html")
	if err != nil {
		return nil, fmt.Errorf("open built frontend index.html (run make build first): %w", err)
	}
	info, err := index.Stat()
	index.Close()
	if err != nil {
		return nil, fmt.Errorf("read built frontend index.html: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("built frontend index.html must be a regular file")
	}
	files := http.FileServerFS(frontend)
	mux := http.NewServeMux()
	mux.Handle("/api", api)
	mux.Handle("/api/", api)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		if !fs.ValidPath(name) {
			http.NotFound(w, r)
			return
		}
		for _, part := range strings.Split(name, "/") {
			if strings.HasPrefix(part, ".") {
				http.NotFound(w, r)
				return
			}
		}
		info, err := fs.Stat(frontend, name)
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasSuffix(name, ".html") {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
	return mux, nil
}
