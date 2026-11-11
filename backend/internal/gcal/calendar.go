package gcal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"everyday/backend/internal/domain"
)

const (
	defaultBaseURL = "https://www.googleapis.com/calendar/v3"

	// lookahead is how far ahead events are fetched, and maxEvents how many are kept.
	// The dashboard shows at most a few, so this is generous.
	lookahead = 14 * 24 * time.Hour
	maxEvents = 20

	// maxCalendarPages bounds how many pages of the calendar list are read.
	maxCalendarPages = 10
)

// Client reads upcoming events from Google calendars. It must not be copied.
type Client struct {
	Tokens TokenProvider
	// CalendarIDs, when set, are the only calendars read; "primary" is the account's
	// main calendar. When empty, the client mirrors the account: every calendar the
	// user has ticked in Google Calendar, holidays and shared calendars included.
	CalendarIDs []string
	// Exclude leaves calendars out of the mirrored set. An entry matches a calendar's
	// ID or its name, ignoring case. It has no effect when CalendarIDs is set.
	Exclude []string
	// BaseURL, HTTP and Now are overridable for tests; zero values use Google and the real clock.
	BaseURL string
	HTTP    *http.Client
	Now     func() time.Time

	mu    sync.Mutex
	shown string // names of the calendars last logged, so a change is logged once
}

type calendarRef struct {
	ID   string
	Name string
}

type calendarListItem struct {
	ID              string `json:"id"`
	Summary         string `json:"summary"`
	SummaryOverride string `json:"summaryOverride"` // the user's own rename, if any
	Selected        bool   `json:"selected"`        // ticked in Google Calendar's sidebar
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

// statusError is a non-200 answer from Google.
type statusError struct {
	Code int
	Text string
}

func (e *statusError) Error() string { return "google calendar returned " + e.Text }

// Fetch returns upcoming and in-progress events from the chosen calendars, merged
// into one list in start order. A calendar that fails is skipped with a logged
// warning so one bad calendar cannot blank the rest; only when every calendar
// fails is the whole fetch an error.
func (c *Client) Fetch(ctx context.Context) ([]domain.CalendarEvent, error) {
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	token, err := c.Tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	calendars, err := c.calendars(ctx, token)
	if err != nil {
		return nil, err
	}
	var events []domain.CalendarEvent
	var failures []error
	for _, calendar := range calendars {
		items, err := c.list(ctx, token, calendar.ID, now)
		if err != nil {
			failures = append(failures, fmt.Errorf("calendar %q: %w", calendar.Name, err))
			continue
		}
		for _, item := range items {
			if event, ok := normalize(item); ok {
				events = append(events, event)
			}
		}
	}
	if len(failures) > 0 {
		if len(failures) == len(calendars) {
			return nil, errors.Join(failures...)
		}
		for _, failure := range failures {
			log.Printf("skipping a Google calendar: %v", failure)
		}
	}
	return merge(events), nil
}

// calendars decides which calendars to read: the configured ones, or else every
// calendar ticked in Google Calendar minus the excluded ones.
func (c *Client) calendars(ctx context.Context, token string) ([]calendarRef, error) {
	var refs []calendarRef
	if len(c.CalendarIDs) > 0 {
		for _, id := range c.CalendarIDs {
			refs = append(refs, calendarRef{ID: id, Name: id})
		}
	} else {
		listed, err := c.listCalendars(ctx, token)
		if err != nil {
			return nil, err
		}
		for _, item := range listed {
			if item.Selected && !c.excluded(item) {
				refs = append(refs, calendarRef{ID: item.ID, Name: firstNonEmpty(item.SummaryOverride, item.Summary, item.ID)})
			}
		}
	}
	c.noteShown(refs)
	return refs, nil
}

func (c *Client) excluded(item calendarListItem) bool {
	for _, name := range c.Exclude {
		if strings.EqualFold(name, item.ID) || strings.EqualFold(name, item.Summary) || strings.EqualFold(name, item.SummaryOverride) {
			return true
		}
	}
	return false
}

// noteShown logs which calendars are on display whenever that set changes, so
// "why isn't my calendar showing?" can be answered from the server log.
func (c *Client) noteShown(refs []calendarRef) {
	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = ref.Name
	}
	joined := strings.Join(names, ", ")
	c.mu.Lock()
	changed := joined != c.shown
	c.shown = joined
	c.mu.Unlock()
	if changed {
		log.Printf("Showing events from %d Google calendar(s): %s", len(refs), joined)
	}
}

// listCalendars reads the user's calendar list: their own, shared ones, and
// subscribed ones such as holidays. Calendars the user can only see as busy/free
// are left out, since they carry no event titles.
func (c *Client) listCalendars(ctx context.Context, token string) ([]calendarListItem, error) {
	var all []calendarListItem
	pageToken := ""
	for page := 0; page < maxCalendarPages; page++ {
		query := url.Values{
			"minAccessRole": {"reader"},
			"maxResults":    {"250"},
			"fields":        {"nextPageToken,items(id,summary,summaryOverride,selected)"},
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		var body struct {
			NextPageToken string             `json:"nextPageToken"`
			Items         []calendarListItem `json:"items"`
		}
		if err := c.getJSON(ctx, token, c.base()+"/users/me/calendarList?"+query.Encode(), &body); err != nil {
			var status *statusError
			if errors.As(err, &status) && status.Code == http.StatusForbidden {
				return nil, fmt.Errorf("google denied access to your calendar list (%w); if you signed in before calendar listing was added, run `make google-auth` again", err)
			}
			return nil, fmt.Errorf("list calendars: %w", err)
		}
		all = append(all, body.Items...)
		if body.NextPageToken == "" {
			break
		}
		pageToken = body.NextPageToken
	}
	return all, nil
}

func (c *Client) list(ctx context.Context, token, calendarID string, now time.Time) ([]eventItem, error) {
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
	var body struct {
		Items []eventItem `json:"items"`
	}
	endpoint := c.base() + "/calendars/" + url.PathEscape(calendarID) + "/events?" + query.Encode()
	if err := c.getJSON(ctx, token, endpoint, &body); err != nil {
		return nil, err
	}
	return body.Items, nil
}

func (c *Client) base() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return defaultBaseURL
}

func (c *Client) getJSON(ctx context.Context, token, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := httpClient(c.HTTP).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return &statusError{Code: res.StatusCode, Text: res.Status}
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode google calendar response: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
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
// calendars, such as an invitation), and keeps the first maxEvents. All-day events
// come before timed ones that start at the same moment, so on any given day they
// lead the list. The result is never nil, so it encodes as [] and not null.
func merge(events []domain.CalendarEvent) []domain.CalendarEvent {
	sort.SliceStable(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if !a.StartsAt.Equal(b.StartsAt) {
			return a.StartsAt.Before(b.StartsAt)
		}
		if a.AllDay != b.AllDay {
			return a.AllDay
		}
		return a.Title < b.Title
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
