package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"everyday/backend/internal/bills"
	"everyday/backend/internal/domain"
	"everyday/backend/internal/mock"
	"everyday/backend/internal/weather"
)

// maxEditBody caps what the editing endpoints will read. A bills list is a few
// hundred bytes; anything near this is a mistake or an attack.
const maxEditBody = 64 << 10

// WeatherSource provides the latest live forecast, or false when it has none.
type WeatherSource interface {
	Current(now time.Time) (domain.Weather, bool)
}

// CalendarSource provides the latest live events, or false when it has none.
// An empty list with true means the calendar really is empty.
type CalendarSource interface {
	Current(now time.Time) ([]domain.CalendarEvent, bool)
}

// BillsStore lets the settings panel read and replace the bills file. Entries says,
// as of now, which due date each bill is counting toward and whether it is paid.
type BillsStore interface {
	Entries(now time.Time) (bills []bills.Listed, problem string)
	Save(entries []bills.Entry, now time.Time) error
}

// FinanceSource provides the bills due next and says how far to trust them: live,
// unavailable (the source is broken), or unconfigured (there is none yet).
type FinanceSource interface {
	Bills(now time.Time) ([]domain.Bill, domain.SourceStatus)
}

type LocationStore interface {
	Location() (weather.Location, string)
	Save(weather.Location) error
}

type LocationSearch interface {
	Search(context.Context, string) ([]weather.Location, error)
}

type Options struct {
	// Mock serves made-up development data and ignores the other fields.
	Mock bool
	// Weather is the live forecast. Location labels the weather widget while
	// no forecast is available.
	Weather  WeatherSource
	Location string
	// Calendar provides live events. Nil means no calendar is connected.
	Calendar CalendarSource
	// Finance provides the bills. Nil means there is no bills source at all.
	Finance FinanceSource
	// BillsStore is what the settings panel edits. Nil, as in mock mode, means bills
	// cannot be edited and the editing endpoints say so.
	BillsStore        BillsStore
	LocationStore     LocationStore
	LocationSearch    LocationSearch
	OnLocationChanged func(weather.Location)
}

type Server struct {
	opts       Options
	anchor     time.Time
	locationMu sync.Mutex
}

func New(opts Options) http.Handler {
	s := &Server{opts: opts, anchor: time.Now()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.state)
	mux.HandleFunc("GET /api/bills", s.editing(s.opts.BillsStore != nil, s.getBills))
	mux.HandleFunc("PUT /api/bills", s.editing(s.opts.BillsStore != nil, s.putBills))
	mux.HandleFunc("GET /api/location", s.editing(true, s.getLocation))
	mux.HandleFunc("PUT /api/location", s.editing(s.opts.LocationStore != nil && s.opts.OnLocationChanged != nil, s.putLocation))
	mux.HandleFunc("GET /api/locations", s.editing(s.opts.LocationSearch != nil, s.searchLocations))
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) { respond(w, domain.Configuration()) })
	return mux
}
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// editing guards the endpoints that change things. There is no login yet, so they
// are restricted to what a person at this machine's own browser can do, which also
// shuts out another web page that tries to drive the dashboard from the same
// browser:
//
//   - the request must be addressed to a loopback name, which defeats DNS rebinding
//     (a hostile site pointing its own name at 127.0.0.1);
//   - a change must be JSON, which a plain cross-site form cannot send without a
//     preflight this server never approves;
//   - a request that names its origin must name this server.
//
// When real authentication arrives it replaces these, not the other way round.
func (s *Server) editing(available bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !available {
			respondError(w, http.StatusNotFound, "these settings cannot be edited here")
			return
		}
		if !loopbackHost(r.Host) {
			respondError(w, http.StatusForbidden, "editing is only available on this machine")
			return
		}
		if r.Method != http.MethodGet {
			if kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || kind != "application/json" {
				respondError(w, http.StatusUnsupportedMediaType, "send the changes as JSON")
				return
			}
			if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host && origin != "https://"+r.Host {
				respondError(w, http.StatusForbidden, "this request came from another site")
				return
			}
		}
		next(w, r)
	}
}

type locationBody struct {
	Location weather.Location `json:"location"`
	Editable bool             `json:"editable"`
	Reason   string           `json:"reason,omitempty"`
}

func (s *Server) getLocation(w http.ResponseWriter, r *http.Request) {
	if s.opts.LocationStore == nil {
		respond(w, locationBody{Location: weather.Location{Name: s.opts.Location}, Reason: "Location cannot be edited while sample data is showing."})
		return
	}
	location, reason := s.opts.LocationStore.Location()
	respond(w, locationBody{Location: location, Editable: reason == "" && s.opts.OnLocationChanged != nil, Reason: reason})
}

func (s *Server) searchLocations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if length := utf8.RuneCountInString(query); length < 2 || length > 100 {
		respondError(w, http.StatusBadRequest, "enter between 2 and 100 characters to search")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	locations, err := s.opts.LocationSearch.Search(ctx, query)
	if err != nil {
		log.Printf("location search: %v", err)
		respondError(w, http.StatusBadGateway, "City search is unavailable. Please try again.")
		return
	}
	if locations == nil {
		locations = []weather.Location{}
	}
	respond(w, map[string]any{"locations": locations})
}

func (s *Server) putLocation(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxEditBody)
	var body struct {
		Name      string   `json:"name"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			respondError(w, http.StatusRequestEntityTooLarge, "that is too much to save")
		} else {
			respondError(w, http.StatusBadRequest, "could not read the location")
		}
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || body.Latitude == nil || body.Longitude == nil {
		respondError(w, http.StatusBadRequest, "choose a location with valid coordinates")
		return
	}
	location := weather.Location{Name: strings.TrimSpace(body.Name), Latitude: *body.Latitude, Longitude: *body.Longitude}
	if err := location.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.locationMu.Lock()
	defer s.locationMu.Unlock()
	if _, reason := s.opts.LocationStore.Location(); reason != "" {
		respondError(w, http.StatusConflict, reason)
		return
	}
	if err := s.opts.LocationStore.Save(location); err != nil {
		log.Printf("saving location: %v", err)
		respondError(w, http.StatusInternalServerError, "The location could not be saved. Your previous location is unchanged.")
		return
	}
	s.opts.OnLocationChanged(location)
	s.getLocation(w, r)
}

// loopbackHost reports whether a Host header names this machine itself.
func loopbackHost(host string) bool {
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	host = strings.Trim(host, "[]")
	return host == "127.0.0.1" || host == "::1" || strings.EqualFold(host, "localhost")
}

type billsBody struct {
	Bills   []bills.Listed `json:"bills"`
	Problem string         `json:"problem,omitempty"`
}

func (s *Server) getBills(w http.ResponseWriter, r *http.Request) {
	entries, problem := s.opts.BillsStore.Entries(time.Now())
	respond(w, billsBody{Bills: entries, Problem: problem})
}

func (s *Server) putBills(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxEditBody)
	var body struct {
		Bills []bills.Entry `json:"bills"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			respondError(w, http.StatusRequestEntityTooLarge, "that is too much to save")
			return
		}
		respondError(w, http.StatusBadRequest, "could not read the bills: "+err.Error())
		return
	}
	if err := s.opts.BillsStore.Save(body.Bills, time.Now()); err != nil {
		if bills.IsInvalid(err) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("saving bills: %v", err)
		respondError(w, http.StatusInternalServerError, "the bills could not be saved")
		return
	}
	entries, problem := s.opts.BillsStore.Entries(time.Now())
	respond(w, billsBody{Bills: entries, Problem: problem})
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
		} else if weather.Location != "" {
			state.Weather.Location = weather.Location
			state.Weather.Loading = weather.Loading
		}
	}
	if s.opts.Calendar != nil {
		// A connected calendar that has not delivered yet is unavailable, not unconfigured.
		state.Sources.Calendar = domain.SourceUnavailable
		if events, ok := s.opts.Calendar.Current(now); ok {
			if events != nil {
				state.Calendar = events
			}
			state.Sources.Calendar = domain.SourceLive
		}
	}
	if s.opts.Finance != nil {
		bills, status := s.opts.Finance.Bills(now)
		if bills != nil {
			state.Finance.Bills = bills
		}
		state.Sources.Finance = status
	}
	return state
}
