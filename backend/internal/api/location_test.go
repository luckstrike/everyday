package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/domain"
	"everyday/backend/internal/weather"
)

type fakeLocationStore struct {
	location weather.Location
	reason   string
	err      error
	saves    int
}

func (s *fakeLocationStore) Location() (weather.Location, string) { return s.location, s.reason }
func (s *fakeLocationStore) Save(location weather.Location) error {
	s.saves++
	if s.err != nil {
		return s.err
	}
	s.location = location
	return nil
}

type fakeLocationSearch struct {
	query string
	err   error
}

func (s *fakeLocationSearch) Search(ctx context.Context, query string) ([]weather.Location, error) {
	s.query = query
	return []weather.Location{{Name: "Austin, Texas, United States", Latitude: 30.27, Longitude: -97.74}}, s.err
}

func locationRequest(method, path, body string) *http.Request {
	request := editRequest(method, body)
	request.URL.Path = path
	return request
}

func TestLocationEndpointsSaveAndRefreshWithoutDependingOnBills(t *testing.T) {
	store := &fakeLocationStore{location: weather.Location{Name: "Dallas"}}
	var changed weather.Location
	changes := 0
	handler := New(Options{LocationStore: store, OnLocationChanged: func(location weather.Location) { changed = location; changes++ }})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, locationRequest("GET", "/api/location", ""))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"editable":true`) {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, locationRequest("PUT", "/api/location", `{"name":" Austin ","latitude":30.27,"longitude":-97.74}`))
	if response.Code != 200 || changes != 1 || store.saves != 1 || store.location.Name != "Austin" || changed != store.location {
		t.Fatalf("%d %s; changed %+v", response.Code, response.Body.String(), changed)
	}
}

func TestLocationSaveRejectsInvalidOrUntrustedChanges(t *testing.T) {
	valid := `{"name":"Austin","latitude":30.27,"longitude":-97.74}`
	for name, body := range map[string]string{
		"missing coordinates":    `{"name":"Austin"}`,
		"missing longitude":      `{"name":"Austin","latitude":30.27}`,
		"null coordinates":       `{"name":"Austin","latitude":null,"longitude":null}`,
		"empty name":             `{"name":" ","latitude":0,"longitude":0}`,
		"latitude out of range":  `{"name":"Austin","latitude":91,"longitude":0}`,
		"longitude out of range": `{"name":"Austin","latitude":0,"longitude":-181}`,
		"unknown field":          `{"name":"Austin","latitude":0,"longitude":0,"admin":true}`,
		"trailing JSON":          valid + `{}`,
		"malformed":              `{oops`,
	} {
		t.Run(name, func(t *testing.T) {
			store := &fakeLocationStore{}
			response := httptest.NewRecorder()
			handler := New(Options{LocationStore: store, OnLocationChanged: func(weather.Location) { t.Fatal("must not refresh") }})
			handler.ServeHTTP(response, locationRequest("PUT", "/api/location", body))
			if response.Code != 400 || store.saves != 0 {
				t.Fatalf("%d %s", response.Code, response.Body.String())
			}
		})
	}
	for _, kind := range []string{"origin", "host", "content-type", "body size"} {
		store := &fakeLocationStore{}
		request := locationRequest("PUT", "/api/location", valid)
		status := 403
		switch kind {
		case "origin":
			request.Header.Set("Origin", "https://evil.example")
		case "host":
			request.Host = "evil.example:8080"
			request.Header.Del("Origin")
		case "content-type":
			request.Header.Set("Content-Type", "text/plain")
			status = 415
		case "body size":
			request = locationRequest("PUT", "/api/location", `{"name":"`+strings.Repeat("x", 70<<10)+`","latitude":0,"longitude":0}`)
			status = 413
		}
		response := httptest.NewRecorder()
		New(Options{LocationStore: store, OnLocationChanged: func(weather.Location) { t.Fatal("must not refresh") }}).ServeHTTP(response, request)
		if response.Code != status || store.saves != 0 {
			t.Fatalf("%s: %d %s", kind, response.Code, response.Body.String())
		}
	}
}

func TestLocationSaveFailuresAndOverridesDoNotChangeWeather(t *testing.T) {
	for _, locked := range []bool{false, true} {
		store := &fakeLocationStore{location: weather.Location{Name: "Dallas"}, err: errors.New("disk failed")}
		status := 500
		if locked {
			store.reason = "Environment override"
			status = 409
		}
		response := httptest.NewRecorder()
		New(Options{LocationStore: store, OnLocationChanged: func(weather.Location) { t.Fatal("must not refresh") }}).ServeHTTP(response, locationRequest("PUT", "/api/location", `{"name":"Austin","latitude":30,"longitude":-97}`))
		if response.Code != status || store.location.Name != "Dallas" {
			t.Fatalf("%d %s", response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	New(Options{Mock: true, Location: "Dallas"}).ServeHTTP(response, locationRequest("GET", "/api/location", ""))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"editable":false`) || !strings.Contains(response.Body.String(), "sample data") {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
}

func TestLocationSearchValidatesQueryAndReportsUpstreamFailure(t *testing.T) {
	search := &fakeLocationSearch{}
	handler := New(Options{LocationSearch: search})
	for query, status := range map[string]int{"Austin%2C+TX": 200, "a": 400, "": 400, strings.Repeat("x", 101): 400} {
		request := httptest.NewRequest("GET", "http://localhost/api/locations?q="+query, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != status {
			t.Fatalf("query %q: %d %s", query, response.Code, response.Body.String())
		}
		if status == 200 && search.query != "Austin, TX" {
			t.Fatal("query not decoded")
		}
	}
	search.err = errors.New("upstream offline")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://localhost/api/locations?q=Austin", nil))
	if response.Code != 502 || !strings.Contains(response.Body.String(), "try again") {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://evil.example/api/locations?q=Austin", nil))
	if response.Code != 403 {
		t.Fatal("search must be loopback-only")
	}
}

func TestWeatherStateUsesNewLocationWhileWaitingWithoutOldConditions(t *testing.T) {
	service := weather.NewService(nil)
	service.SetLocation(weather.Location{Name: "Austin", Latitude: 30.27, Longitude: -97.74})
	server := &Server{opts: Options{Location: "Dallas", Weather: service}}
	state := server.build(time.Now())
	if state.Weather.Location != "Austin" || !state.Weather.Loading || state.Sources.Weather != domain.SourceUnavailable || state.Weather.Temperature != 0 || len(state.Weather.Daily) != 0 {
		t.Fatalf("%+v", state)
	}
	data, err := json.Marshal(state)
	if err != nil || strings.Contains(string(data), `"daily":null`) {
		t.Fatalf("%s, %v", data, err)
	}
}
