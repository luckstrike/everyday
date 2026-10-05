package api

import (
	"apartment-display/backend/internal/domain"
	"apartment-display/backend/internal/mock"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	mu          sync.Mutex
	anchor      time.Time
	scenario    string
	started     time.Time
	development bool
}

func New(dev bool) http.Handler {
	now := time.Now()
	s := &Server{anchor: now, scenario: "normal", started: now, development: dev}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.state)
	mux.HandleFunc("GET /api/widgets", func(w http.ResponseWriter, r *http.Request) { respond(w, domain.Widgets()) })
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) { respond(w, domain.Configuration(dev)) })
	if dev {
		mux.HandleFunc("POST /api/scenario", s.simulate)
	}
	return mux
}
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}
func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	state := mock.Build(s.anchor, time.Now(), s.scenario, s.started)
	s.mu.Unlock()
	respond(w, state)
}
func (s *Server) simulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Scenario string `json:"scenario"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || !mock.ValidScenario(req.Scenario) {
		http.Error(w, "Unknown or invalid scenario", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Expected a single JSON object", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.scenario = req.Scenario
	s.started = time.Now()
	state := mock.Build(s.anchor, s.started, s.scenario, s.started)
	s.mu.Unlock()
	respond(w, state)
}
