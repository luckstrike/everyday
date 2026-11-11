package api

import (
	"everyday/backend/internal/domain"
	"everyday/backend/internal/mock"
	"encoding/json"
	"net/http"
	"time"
)

type Server struct {
	anchor time.Time
}

func New() http.Handler {
	s := &Server{anchor: time.Now()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.state)
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) { respond(w, domain.Configuration()) })
	return mux
}
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}
func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	respond(w, mock.Build(s.anchor, time.Now()))
}
