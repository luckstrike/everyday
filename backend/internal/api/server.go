package api

import (
	"encoding/json"
	"net/http"
	"time"

	"everyday/backend/internal/domain"
	"everyday/backend/internal/mock"
)

// WeatherSource provides the latest live forecast, or false when it has none.
type WeatherSource interface {
	Current(now time.Time) (domain.Weather, bool)
}

type Options struct {
	// Mock serves made-up development data and ignores the other fields.
	Mock bool
	// Weather is the live forecast. Location labels the weather widget while
	// no forecast is available.
	Weather  WeatherSource
	Location string
}

type Server struct {
	opts   Options
	anchor time.Time
}

func New(opts Options) http.Handler {
	s := &Server{opts: opts, anchor: time.Now()}
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
	respond(w, s.build(time.Now()))
}

// build assembles the normalized state. Sections without a live source are
// empty and say so in Sources, which the frontend uses for its empty states.
// Lists are empty slices, never nil, so they encode as [] and not null.
func (s *Server) build(now time.Time) domain.State {
	if s.opts.Mock {
		return mock.Build(s.anchor, now)
	}
	state := domain.State{
		GeneratedAt: now,
		Sources: domain.Sources{
			Weather:  domain.SourceUnavailable,
			Finance:  domain.SourceUnconfigured,
			Calendar: domain.SourceUnconfigured,
		},
		Weather: domain.Weather{
			Location: s.opts.Location,
			Daily:    []domain.ForecastDay{},
			Hourly:   []domain.Hour{},
		},
		Finance:  domain.Finance{Bills: []domain.Bill{}},
		Calendar: []domain.CalendarEvent{},
	}
	if s.opts.Weather != nil {
		if weather, ok := s.opts.Weather.Current(now); ok {
			state.Weather = weather
			state.Sources.Weather = domain.SourceLive
		}
	}
	return state
}
