package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

type fakeWeather struct {
	weather domain.Weather
	ok      bool
}

func (f fakeWeather) Current(time.Time) (domain.Weather, bool) { return f.weather, f.ok }

func getState(t *testing.T, opts Options) (domain.State, string) {
	t.Helper()
	w := httptest.NewRecorder()
	New(opts).ServeHTTP(w, httptest.NewRequest("GET", "/api/state", nil))
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var state domain.State
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	return state, w.Body.String()
}

func TestEndpoints(t *testing.T) {
	h := New(Options{Mock: true})
	for _, path := range []string{"/api/state", "/api/config"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
}

func TestMockModeLabelsEverySourceAsMock(t *testing.T) {
	state, _ := getState(t, Options{Mock: true})
	want := domain.Sources{Weather: domain.SourceMock, Finance: domain.SourceMock, Calendar: domain.SourceMock}
	if state.Sources != want || len(state.Calendar) == 0 {
		t.Fatalf("sources %+v, %d events", state.Sources, len(state.Calendar))
	}
}

func TestLiveModeUsesTheWeatherSourceAndLeavesTheRestUnconfigured(t *testing.T) {
	live := domain.Weather{Location: "Austin, TX", Temperature: 71, Condition: "Clear skies"}
	state, _ := getState(t, Options{Weather: fakeWeather{live, true}, Location: "Austin, TX"})
	want := domain.Sources{Weather: domain.SourceLive, Finance: domain.SourceUnconfigured, Calendar: domain.SourceUnconfigured}
	if state.Sources != want {
		t.Fatalf("sources: %+v", state.Sources)
	}
	if state.Weather.Temperature != 71 || len(state.Calendar) != 0 || len(state.Finance.Bills) != 0 {
		t.Fatalf("state: %+v", state)
	}
}

func TestLiveModeWithoutAForecastIsUnavailableButStillValid(t *testing.T) {
	for name, source := range map[string]WeatherSource{"no forecast yet": fakeWeather{}, "no source": nil} {
		t.Run(name, func(t *testing.T) {
			state, body := getState(t, Options{Weather: source, Location: "Austin, TX"})
			if state.Sources.Weather != domain.SourceUnavailable {
				t.Fatalf("weather source: %q", state.Sources.Weather)
			}
			if state.Weather.Location != "Austin, TX" {
				t.Fatalf("location label: %q", state.Weather.Location)
			}
			// The frontend calls .slice() and .filter() on these, so null would crash it.
			for _, empty := range []string{`"daily":[]`, `"hourly":[]`, `"bills":[]`, `"calendar":[]`} {
				if !strings.Contains(body, empty) {
					t.Fatalf("missing %s in %s", empty, body)
				}
			}
		})
	}
}

type fakeCalendar struct {
	events []domain.CalendarEvent
	ok     bool
}

func (f fakeCalendar) Current(time.Time) ([]domain.CalendarEvent, bool) { return f.events, f.ok }

func TestConnectedCalendarIsLiveAndShowsItsEvents(t *testing.T) {
	events := []domain.CalendarEvent{{Title: "Dinner with Alex"}}
	state, _ := getState(t, Options{Calendar: fakeCalendar{events, true}})
	if state.Sources.Calendar != domain.SourceLive || len(state.Calendar) != 1 || state.Calendar[0].Title != "Dinner with Alex" {
		t.Fatalf("sources %+v, events %+v", state.Sources, state.Calendar)
	}
	// Finance is still unconfigured: connecting a calendar changes nothing else.
	if state.Sources.Finance != domain.SourceUnconfigured {
		t.Fatalf("finance: %q", state.Sources.Finance)
	}
}

func TestEmptyConnectedCalendarIsLiveAndEncodesAsAList(t *testing.T) {
	state, body := getState(t, Options{Calendar: fakeCalendar{nil, true}})
	if state.Sources.Calendar != domain.SourceLive {
		t.Fatalf("calendar: %q", state.Sources.Calendar)
	}
	if !strings.Contains(body, `"calendar":[]`) {
		t.Fatalf("a nil list would encode as null and crash the frontend: %s", body)
	}
}

func TestConnectedCalendarThatHasNotDeliveredIsUnavailable(t *testing.T) {
	state, _ := getState(t, Options{Calendar: fakeCalendar{}})
	if state.Sources.Calendar != domain.SourceUnavailable || len(state.Calendar) != 0 {
		t.Fatalf("sources %+v, events %+v", state.Sources, state.Calendar)
	}
}

func TestRemovedEndpointsAreGone(t *testing.T) {
	for _, path := range []string{"/api/widgets", "/api/scenario"} {
		w := httptest.NewRecorder()
		New(Options{Mock: true}).ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
}
