package gcal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

type staticToken string

func (s staticToken) Token(context.Context) (string, error) { return string(s), nil }

var calendarNow = time.Date(2026, 11, 11, 10, 0, 0, 0, time.FixedZone("", -6*3600))

const primaryEvents = `{"items":[
  {"summary":"Dinner with Alex","location":"At home",
   "start":{"dateTime":"2026-11-12T18:00:00-06:00"},"end":{"dateTime":"2026-11-12T20:00:00-06:00"},
   "attendees":[{"self":true,"responseStatus":"accepted"}]},
  {"summary":"Alex's birthday","start":{"date":"2026-11-15"},"end":{"date":"2026-11-16"}},
  {"status":"cancelled","summary":"Cancelled thing",
   "start":{"dateTime":"2026-11-12T11:00:00-06:00"},"end":{"dateTime":"2026-11-12T12:00:00-06:00"}},
  {"summary":"Declined thing",
   "start":{"dateTime":"2026-11-12T12:00:00-06:00"},"end":{"dateTime":"2026-11-12T13:00:00-06:00"},
   "attendees":[{"self":false,"responseStatus":"accepted"},{"self":true,"responseStatus":"declined"}]},
  {"start":{"dateTime":"2026-11-12T09:00:00-06:00"},"end":{"dateTime":"2026-11-12T09:30:00-06:00"}}
]}`

const familyEvents = `{"items":[
  {"summary":"Dinner with Alex",
   "start":{"dateTime":"2026-11-12T18:00:00-06:00"},"end":{"dateTime":"2026-11-12T20:00:00-06:00"}},
  {"summary":"Pick up the kids",
   "start":{"dateTime":"2026-11-12T15:00:00-06:00"},"end":{"dateTime":"2026-11-12T15:30:00-06:00"}}
]}`

func calendarClient(t *testing.T, handler http.HandlerFunc, ids ...string) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{Tokens: staticToken("tok"), CalendarIDs: ids, BaseURL: server.URL, Now: func() time.Time { return calendarNow }}
}

func TestFetchMergesCalendarsAndSkipsWhatShouldNotShow(t *testing.T) {
	requests := map[string]int{}
	client := calendarClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("authorization: %q", r.Header.Get("Authorization"))
		}
		q := r.URL.Query()
		if q.Get("singleEvents") != "true" || q.Get("orderBy") != "startTime" ||
			q.Get("timeMin") != calendarNow.Format(time.RFC3339) ||
			q.Get("timeMax") != calendarNow.Add(lookahead).Format(time.RFC3339) {
			t.Errorf("query: %v", q)
		}
		switch r.URL.Path {
		case "/calendars/primary/events":
			w.Write([]byte(primaryEvents))
		case "/calendars/family@example.com/events":
			w.Write([]byte(familyEvents))
		default:
			http.NotFound(w, r)
		}
	}, "primary", "family@example.com")

	events, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, e := range events {
		titles = append(titles, e.Title)
	}
	// Untitled (09:00), the kids (15:00), dinner once (18:00, on both calendars), then the all-day birthday.
	want := []string{"(No title)", "Pick up the kids", "Dinner with Alex", "Alex's birthday"}
	if fmt.Sprint(titles) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", titles, want)
	}
	if requests["/calendars/primary/events"] != 1 || requests["/calendars/family@example.com/events"] != 1 {
		t.Fatalf("requests: %v", requests)
	}

	dinner := events[2]
	if dinner.Location != "At home" || dinner.AllDay || !dinner.StartsAt.Equal(time.Date(2026, 11, 12, 18, 0, 0, 0, time.FixedZone("", -6*3600))) {
		t.Fatalf("dinner: %+v", dinner)
	}
	birthday := events[3]
	wantStart := time.Date(2026, 11, 15, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2026, 11, 16, 0, 0, 0, 0, time.Local)
	if !birthday.AllDay || !birthday.StartsAt.Equal(wantStart) || !birthday.EndsAt.Equal(wantEnd) {
		t.Fatalf("all-day event should span the local day: %+v", birthday)
	}
}

func TestFetchWithNoEventsIsAnEmptyListNotNull(t *testing.T) {
	client := calendarClient(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{}`)) }, "primary")
	events, err := client.Fetch(context.Background())
	if err != nil || events == nil || len(events) != 0 {
		t.Fatalf("got %#v, %v", events, err)
	}
}

func TestFetchFailsWhenEveryCalendarFails(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		"forbidden": func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusForbidden) },
		"not json":  func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html>")) },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := calendarClient(t, handler, "primary").Fetch(context.Background()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestAllDayEventsLeadTheirDay(t *testing.T) {
	midnight := time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local)
	allDay := domain.CalendarEvent{Title: "Veterans Day", AllDay: true, StartsAt: midnight, EndsAt: midnight.AddDate(0, 0, 1)}
	earlyTimed := domain.CalendarEvent{Title: "Aardvark check-in", StartsAt: midnight, EndsAt: midnight.Add(time.Hour)}
	laterTimed := domain.CalendarEvent{Title: "Dinner", StartsAt: midnight.Add(18 * time.Hour), EndsAt: midnight.Add(20 * time.Hour)}

	// The all-day event leads even against a timed event starting at the same
	// instant, whose title would otherwise sort first, and whatever the input order.
	for _, input := range [][]domain.CalendarEvent{
		{laterTimed, earlyTimed, allDay},
		{allDay, laterTimed, earlyTimed},
	} {
		merged := merge(append([]domain.CalendarEvent(nil), input...))
		if len(merged) != 3 || merged[0].Title != "Veterans Day" || merged[1].Title != "Aardvark check-in" || merged[2].Title != "Dinner" {
			t.Fatalf("got %+v", merged)
		}
	}
}

func TestMergeKeepsOnlyTheFirstEvents(t *testing.T) {
	var many []domain.CalendarEvent
	for i := 0; i < maxEvents+5; i++ {
		start := calendarNow.Add(time.Duration(i) * time.Hour)
		many = append(many, domain.CalendarEvent{Title: fmt.Sprint("event ", i), StartsAt: start, EndsAt: start.Add(time.Hour)})
	}
	merged := merge(many)
	if len(merged) != maxEvents || merged[0].Title != "event 0" {
		t.Fatalf("got %d events starting with %q", len(merged), merged[0].Title)
	}
}

// A calendar list as Google returns it: the user's own calendar, a subscribed
// holiday calendar, Birthdays, a renamed work calendar, and one that is unticked.
const calendarList = `{"items":[
  {"id":"me@example.com","summary":"Alex","selected":true},
  {"id":"en.usa#holiday@group.v.calendar.google.com","summary":"Holidays in United States","selected":true},
  {"id":"birthdays","summary":"Birthdays","selected":true},
  {"id":"work@example.com","summary":"Work","summaryOverride":"Day job","selected":true},
  {"id":"old@example.com","summary":"Old project","selected":false}
]}`

// mirrorServer serves a calendar list and one event per calendar, titled with the
// calendar's ID, and records which calendars had their events requested.
type mirrorServer struct {
	mu        sync.Mutex
	list      string
	listCalls int
	events    []string
	failing   map[string]bool // calendar IDs whose events return an error
}

func (m *mirrorServer) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if r.URL.Path == "/users/me/calendarList" {
			m.listCalls++
			if r.URL.Query().Get("minAccessRole") != "reader" {
				t.Errorf("calendar list should skip calendars that hide event details: %v", r.URL.Query())
			}
			w.Write([]byte(m.list))
			return
		}
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/calendars/"), "/events")
		m.events = append(m.events, id)
		if m.failing[id] {
			http.Error(w, "nope", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, `{"items":[{"summary":%q,"start":{"date":"2026-11-15"},"end":{"date":"2026-11-16"}}]}`, id)
	}
}

func (m *mirrorServer) client(t *testing.T) *Client {
	t.Helper()
	return calendarClient(t, m.handler(t))
}

func TestWithNoConfiguredCalendarsTheTickedOnesAreMirrored(t *testing.T) {
	m := &mirrorServer{list: calendarList}
	events, err := m.client(t).Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// Everything ticked in Google Calendar, holidays included; the unticked one is left out.
	want := "birthdays,en.usa#holiday@group.v.calendar.google.com,me@example.com,work@example.com"
	var titles []string
	for _, e := range events {
		titles = append(titles, e.Title)
	}
	got := strings.Join(sorted(titles), ",")
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if m.listCalls != 1 {
		t.Fatalf("calendar list requested %d times", m.listCalls)
	}
}

func TestExcludeMatchesIDsAndNamesIgnoringCase(t *testing.T) {
	m := &mirrorServer{list: calendarList}
	client := m.client(t)
	// By name, by the user's own rename, and by ID.
	client.Exclude = []string{"BIRTHDAYS", "day job", "en.usa#holiday@group.v.calendar.google.com"}
	if _, err := client.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(m.events, ","); got != "me@example.com" {
		t.Fatalf("read events from %s", got)
	}
	// The calendar's original name still matches after a rename.
	m2 := &mirrorServer{list: calendarList}
	client2 := m2.client(t)
	client2.Exclude = []string{"Work"}
	client2.Fetch(context.Background())
	if strings.Contains(strings.Join(m2.events, ","), "work@example.com") {
		t.Fatalf("the renamed calendar should still be excluded by its original name: %v", m2.events)
	}
}

func TestConfiguredCalendarsSkipTheListAndIgnoreExcludes(t *testing.T) {
	m := &mirrorServer{list: calendarList}
	client := m.client(t)
	client.CalendarIDs = []string{"birthdays"}
	client.Exclude = []string{"birthdays"}
	if _, err := client.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if m.listCalls != 0 || strings.Join(m.events, ",") != "birthdays" {
		t.Fatalf("list calls %d, events read from %v", m.listCalls, m.events)
	}
}

func TestLongCalendarListsAreReadPageByPage(t *testing.T) {
	var pages []string
	client := calendarClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/users/me/calendarList" && r.URL.Query().Get("pageToken") == "":
			w.Write([]byte(`{"nextPageToken":"p2","items":[{"id":"one","summary":"One","selected":true}]}`))
		case r.URL.Path == "/users/me/calendarList" && r.URL.Query().Get("pageToken") == "p2":
			w.Write([]byte(`{"items":[{"id":"two","summary":"Two","selected":true}]}`))
		default:
			pages = append(pages, r.URL.Path)
			w.Write([]byte(`{}`))
		}
	})
	if _, err := client.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if strings.Join(pages, ",") != "/calendars/one/events,/calendars/two/events" {
		t.Fatalf("read %v", pages)
	}
}

func TestNothingTickedIsAnEmptyButLiveCalendar(t *testing.T) {
	m := &mirrorServer{list: `{"items":[{"id":"old","summary":"Old","selected":false}]}`}
	events, err := m.client(t).Fetch(context.Background())
	if err != nil || events == nil || len(events) != 0 {
		t.Fatalf("got %#v, %v", events, err)
	}
}

func TestMissingCalendarListPermissionSaysHowToFixIt(t *testing.T) {
	client := calendarClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Request had insufficient authentication scopes.", http.StatusForbidden)
	})
	_, err := client.Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "make google-auth") {
		t.Fatalf("got %v", err)
	}
}

func TestOneBrokenCalendarIsSkippedNotFatal(t *testing.T) {
	m := &mirrorServer{list: calendarList, failing: map[string]bool{"birthdays": true}}
	events, err := m.client(t).Fetch(context.Background())
	if err != nil {
		t.Fatalf("one broken calendar must not fail the rest: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events, want the 3 from the working calendars", len(events))
	}
}

func TestEveryBrokenCalendarIsAnError(t *testing.T) {
	m := &mirrorServer{list: calendarList, failing: map[string]bool{
		"me@example.com": true, "en.usa#holiday@group.v.calendar.google.com": true,
		"birthdays": true, "work@example.com": true,
	}}
	if _, err := m.client(t).Fetch(context.Background()); err == nil {
		t.Fatal("expected an error when no calendar can be read")
	}
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
