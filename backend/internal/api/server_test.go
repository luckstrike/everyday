package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/bills"
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

type fakeFinance struct {
	bills  []domain.Bill
	status domain.SourceStatus
}

func (f fakeFinance) Bills(time.Time) ([]domain.Bill, domain.SourceStatus) { return f.bills, f.status }

func TestBillsFromTheFinanceSourceAreServedWithTheirStatus(t *testing.T) {
	bills := []domain.Bill{{Name: "Rent", Amount: 1850}}
	state, _ := getState(t, Options{Finance: fakeFinance{bills, domain.SourceLive}})
	if state.Sources.Finance != domain.SourceLive || len(state.Finance.Bills) != 1 || state.Finance.Bills[0].Name != "Rent" {
		t.Fatalf("sources %+v, bills %+v", state.Sources, state.Finance.Bills)
	}
}

func TestFinanceStatusesPassThroughAndKeepListsAsLists(t *testing.T) {
	for _, status := range []domain.SourceStatus{domain.SourceUnconfigured, domain.SourceUnavailable} {
		state, body := getState(t, Options{Finance: fakeFinance{nil, status}})
		if state.Sources.Finance != status {
			t.Fatalf("finance status %q, want %q", state.Sources.Finance, status)
		}
		if !strings.Contains(body, `"bills":[]`) {
			t.Fatalf("a nil list would encode as null and crash the frontend: %s", body)
		}
	}
}

// fakeStore is an in-memory bills store for the editing endpoints.
type fakeStore struct {
	entries []bills.Entry
	problem string
	saveErr error
	saves   int
}

func (f *fakeStore) Entries(now time.Time) ([]bills.Listed, string) {
	listed := []bills.Listed{}
	for _, e := range f.entries {
		listed = append(listed, bills.Listed{Entry: e, Due: "2026-11-18"})
	}
	return listed, f.problem
}

func (f *fakeStore) Save(entries []bills.Entry) error {
	f.saves++
	if f.saveErr != nil {
		return f.saveErr
	}
	f.entries = entries
	return nil
}

// editRequest builds a request to the bills endpoints the way the dashboard's own
// page would send it: addressed to loopback, with that page as its origin.
func editRequest(method, body string) *http.Request {
	req := httptest.NewRequest(method, "/api/bills", strings.NewReader(body))
	req.Host = "127.0.0.1:5173"
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:5173")
	}
	return req
}

func serveEdit(store BillsStore, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	New(Options{BillsStore: store}).ServeHTTP(w, req)
	return w
}

func TestEditorReadsTheBillsAndAnyProblemWithTheFile(t *testing.T) {
	store := &fakeStore{entries: []bills.Entry{{Name: "Rent", Amount: 1850, Day: 1}}}
	w := serveEdit(store, editRequest("GET", ""))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Rent"`) || strings.Contains(w.Body.String(), "problem") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}

	broken := &fakeStore{problem: `bill "Rent": day must be 1 to 31, got 99`}
	w = serveEdit(broken, editRequest("GET", ""))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "day must be 1 to 31") || !strings.Contains(w.Body.String(), `"bills":[]`) {
		t.Fatalf("a broken file is reported, with an empty list: %d %s", w.Code, w.Body.String())
	}
}

func TestEditorSavesAndReturnsWhatWasSaved(t *testing.T) {
	store := &fakeStore{}
	w := serveEdit(store, editRequest("PUT", `{"bills":[{"name":"Rent","amount":1850,"day":1},{"name":"Phone","amount":40,"day":11,"paidFor":"2026-11-11"}]}`))
	if w.Code != 200 || store.saves != 1 || len(store.entries) != 2 || store.entries[1].PaidFor != "2026-11-11" {
		t.Fatalf("%d %s, saves %d, entries %+v", w.Code, w.Body.String(), store.saves, store.entries)
	}
	if !strings.Contains(w.Body.String(), `"name":"Phone"`) {
		t.Fatalf("response should carry the saved bills: %s", w.Body.String())
	}
	// Saving an empty list is how the last bill is deleted.
	w = serveEdit(store, editRequest("PUT", `{"bills":[]}`))
	if w.Code != 200 || len(store.entries) != 0 {
		t.Fatalf("empty save: %d %s, %+v", w.Code, w.Body.String(), store.entries)
	}
}

func TestEditorExplainsInvalidBillsWithoutHidingServerFailures(t *testing.T) {
	_, invalidErr := bills.Validate([]bills.Entry{{Name: "Rent", Amount: 1, Day: 99}})
	invalid := &fakeStore{saveErr: invalidErr}
	w := serveEdit(invalid, editRequest("PUT", `{"bills":[{"name":"Rent","amount":1,"day":99}]}`))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "day must be 1 to 31") {
		t.Fatalf("invalid bills are the sender's to fix: %d %s", w.Code, w.Body.String())
	}

	broken := &fakeStore{saveErr: errors.New("open /secret/path: permission denied")}
	w = serveEdit(broken, editRequest("PUT", `{"bills":[]}`))
	if w.Code != 500 || strings.Contains(w.Body.String(), "/secret/path") {
		t.Fatalf("server failures must not leak paths or details: %d %s", w.Code, w.Body.String())
	}
}

func TestEditorRejectsRequestsItCannotTrust(t *testing.T) {
	cases := map[string]struct {
		build func() *http.Request
		want  int
	}{
		"another site's origin": {func() *http.Request {
			r := editRequest("PUT", `{"bills":[]}`)
			r.Header.Set("Origin", "http://evil.example")
			return r
		}, 403},
		"a form post instead of JSON": {func() *http.Request {
			r := editRequest("PUT", `bills=1`)
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r
		}, 415},
		"plain text instead of JSON": {func() *http.Request {
			r := editRequest("PUT", `{"bills":[]}`)
			r.Header.Set("Content-Type", "text/plain")
			return r
		}, 415},
		"a hostile name pointing at this machine": {func() *http.Request {
			r := editRequest("PUT", `{"bills":[]}`)
			r.Host = "evil.example:8080"
			r.Header.Set("Origin", "http://evil.example:8080")
			return r
		}, 403},
		"reading from a non-loopback host": {func() *http.Request {
			r := editRequest("GET", "")
			r.Host = "dashboard.example.com"
			return r
		}, 403},
		"an unknown field": {func() *http.Request {
			return editRequest("PUT", `{"bills":[],"admin":true}`)
		}, 400},
		"malformed JSON": {func() *http.Request {
			return editRequest("PUT", `{oops`)
		}, 400},
		"a body that is far too large": {func() *http.Request {
			return editRequest("PUT", `{"bills":[{"name":"`+strings.Repeat("x", 70<<10)+`","amount":1,"day":1}]}`)
		}, 413},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			store := &fakeStore{}
			w := serveEdit(store, c.build())
			if w.Code != c.want {
				t.Fatalf("status %d, want %d: %s", w.Code, c.want, w.Body.String())
			}
			if store.saves != 0 {
				t.Fatal("a rejected request must not reach the store")
			}
		})
	}
}

func TestEditorAcceptsLocalhostNamesAndNoOriginHeader(t *testing.T) {
	for _, host := range []string{"localhost:5173", "127.0.0.1", "[::1]:8080", "LOCALHOST"} {
		r := editRequest("PUT", `{"bills":[]}`)
		r.Host = host
		r.Header.Del("Origin")
		if w := serveEdit(&fakeStore{}, r); w.Code != 200 {
			t.Errorf("host %q: %d %s", host, w.Code, w.Body.String())
		}
	}
}

func TestEditorIsUnavailableWhenThereIsNoStore(t *testing.T) {
	w := httptest.NewRecorder()
	New(Options{Mock: true}).ServeHTTP(w, editRequest("GET", ""))
	if w.Code != 404 || !strings.Contains(w.Body.String(), "error") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
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
