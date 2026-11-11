package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"everyday/backend/internal/domain"
)

const (
	defaultBaseURL = "https://www.googleapis.com/calendar/v3"

	// lookahead is how far ahead events are fetched, and maxEvents how many are kept.
	// The dashboard shows at most a few, so this is generous.
	lookahead = 14 * 24 * time.Hour
	maxEvents = 20
)

// Client reads upcoming events from one or more Google calendars.
type Client struct {
	Tokens TokenProvider
	// CalendarIDs are Google calendar IDs; "primary" is the account's main calendar.
	CalendarIDs []string
	// BaseURL, HTTP and Now are overridable for tests; zero values use Google and the real clock.
	BaseURL string
	HTTP    *http.Client
	Now     func() time.Time
}

type eventTime struct {
	DateTime string `json:"dateTime"` // timed events: RFC 3339 with an offset
	Date     string `json:"date"`     // all-day events: a plain calendar date
}

type eventItem struct {
	Status    string    `json:"status"`
	Summary   string    `json:"summary"`
	Location  string    `json:"location"`
	Start     eventTime `json:"start"`
	End       eventTime `json:"end"`
	Attendees []struct {
		Self           bool   `json:"self"`
		ResponseStatus string `json:"responseStatus"`
	} `json:"attendees"`
}

// Fetch returns upcoming and in-progress events from every configured calendar,
// merged into one list in start order.
func (c *Client) Fetch(ctx context.Context) ([]domain.CalendarEvent, error) {
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	token, err := c.Tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	var events []domain.CalendarEvent
	for _, id := range c.CalendarIDs {
		items, err := c.list(ctx, token, id, now)
		if err != nil {
			return nil, fmt.Errorf("calendar %q: %w", id, err)
		}
		for _, item := range items {
			if event, ok := normalize(item); ok {
				events = append(events, event)
			}
		}
	}
	return merge(events), nil
}

func (c *Client) list(ctx context.Context, token, calendarID string, now time.Time) ([]eventItem, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	query := url.Values{
		// singleEvents makes Google expand recurring events into individual occurrences.
		"singleEvents": {"true"},
		"orderBy":      {"startTime"},
		// timeMin bounds an event's end, so events already under way are included.
		"timeMin":    {now.Format(time.RFC3339)},
		"timeMax":    {now.Add(lookahead).Format(time.RFC3339)},
		"maxResults": {"50"},
		"fields":     {"items(status,summary,location,start,end,attendees(self,responseStatus))"},
	}
	endpoint := base + "/calendars/" + url.PathEscape(calendarID) + "/events?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := httpClient(c.HTTP).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google calendar returned %s", res.Status)
	}
	var body struct {
		Items []eventItem `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode google calendar response: %w", err)
	}
	return body.Items, nil
}

// normalize converts one Google event, reporting false for events that should not
// be shown: cancelled ones, and ones the user has declined.
func normalize(item eventItem) (domain.CalendarEvent, bool) {
	if item.Status == "cancelled" {
		return domain.CalendarEvent{}, false
	}
	for _, a := range item.Attendees {
		if a.Self && a.ResponseStatus == "declined" {
			return domain.CalendarEvent{}, false
		}
	}
	event := domain.CalendarEvent{Title: item.Summary, Location: item.Location}
	if event.Title == "" {
		event.Title = "(No title)"
	}
	var err error
	if item.Start.Date != "" {
		// All-day events are dates, not instants; they mean the local day. Google's
		// end date is exclusive, so it is already the midnight after the last day.
		event.AllDay = true
		if event.StartsAt, err = time.ParseInLocation("2006-01-02", item.Start.Date, time.Local); err != nil {
			return domain.CalendarEvent{}, false
		}
		if event.EndsAt, err = time.ParseInLocation("2006-01-02", item.End.Date, time.Local); err != nil {
			return domain.CalendarEvent{}, false
		}
		return event, true
	}
	if event.StartsAt, err = time.Parse(time.RFC3339, item.Start.DateTime); err != nil {
		return domain.CalendarEvent{}, false
	}
	if event.EndsAt, err = time.Parse(time.RFC3339, item.End.DateTime); err != nil {
		return domain.CalendarEvent{}, false
	}
	return event, true
}

// merge orders events by start, drops duplicates (the same event on two
// calendars, such as an invitation), and keeps the first maxEvents. The result
// is never nil, so it encodes as [] and not null.
func merge(events []domain.CalendarEvent) []domain.CalendarEvent {
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].StartsAt.Equal(events[j].StartsAt) {
			return events[i].StartsAt.Before(events[j].StartsAt)
		}
		return events[i].Title < events[j].Title
	})
	type key struct {
		title      string
		start, end int64
	}
	seen := map[key]bool{}
	merged := make([]domain.CalendarEvent, 0, len(events))
	for _, e := range events {
		k := key{e.Title, e.StartsAt.UnixNano(), e.EndsAt.UnixNano()}
		if seen[k] {
			continue
		}
		seen[k] = true
		merged = append(merged, e)
	}
	if len(merged) > maxEvents {
		merged = merged[:maxEvents]
	}
	return merged
}
