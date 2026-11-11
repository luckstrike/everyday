package gcal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestFetchFailsWhenAnyCalendarFails(t *testing.T) {
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
