// Package bills reads the household's recurring bills from a small JSON file and
// works out when each is next due. There is no external source: the file is the
// source, and it is re-read whenever it changes, so edits show up without a restart.
package bills

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"everyday/backend/internal/domain"
)

// Entry is one bill as written in the file. Every bill repeats monthly.
type Entry struct {
	Name    string  `json:"name"`
	Amount  float64 `json:"amount"`
	Day     int     `json:"day"` // day of the month it is due, 1 to 31
	Autopay bool    `json:"autopay"`
}

// file is the shape of the bills file. It is an object, not a bare list, so
// other fields can be added later without breaking existing files.
type file struct {
	Bills []Entry `json:"bills"`
}

// DefaultFile is where the bills file lives unless EVERYDAY_BILLS says otherwise:
// everyday/bills.json in the user's config directory. It is empty if that
// directory cannot be found.
func DefaultFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "everyday", "bills.json")
}

// Parse reads and validates the contents of a bills file. A misspelled field is an
// error rather than silently ignored.
func Parse(data []byte) ([]Entry, error) {
	var f file
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&f); err != nil {
		return nil, err
	}
	for i, e := range f.Bills {
		switch {
		case strings.TrimSpace(e.Name) == "":
			return nil, fmt.Errorf("bill %d has no name", i+1)
		case e.Day < 1 || e.Day > 31:
			return nil, fmt.Errorf("bill %q: day must be 1 to 31, got %d", e.Name, e.Day)
		case math.IsNaN(e.Amount) || math.IsInf(e.Amount, 0) || e.Amount < 0:
			return nil, fmt.Errorf("bill %q: amount must be zero or more", e.Name)
		}
		f.Bills[i].Name = strings.TrimSpace(e.Name)
	}
	return f.Bills, nil
}

// NextDue is the date a bill due on the given day of the month is next due, counting
// today as still due. A bill due on the 31st falls on the last day of a shorter
// month. The date is local midnight, in now's time zone.
func NextDue(day int, now time.Time) time.Time {
	loc := now.Location()
	year, month, today := now.Date()
	due := dueIn(year, month, day, loc)
	if due.Before(time.Date(year, month, today, 0, 0, 0, 0, loc)) {
		due = dueIn(year, month+1, day, loc)
	}
	return due
}

// dueIn is the given day of a month, clamped to the month's length. time.Date
// normalizes an out-of-range month, so month+1 rolls into the next year.
func dueIn(year int, month time.Month, day int, loc *time.Location) time.Time {
	if last := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day(); day > last {
		day = last
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// Upcoming turns entries into bills with their next due dates, soonest first.
// The result is never nil, so it encodes as [] and not null.
func Upcoming(entries []Entry, now time.Time) []domain.Bill {
	upcoming := make([]domain.Bill, 0, len(entries))
	for _, e := range entries {
		upcoming = append(upcoming, domain.Bill{
			Name:    e.Name,
			Amount:  e.Amount,
			DueAt:   NextDue(e.Day, now),
			Autopay: e.Autopay,
		})
	}
	sort.SliceStable(upcoming, func(i, j int) bool {
		if !upcoming[i].DueAt.Equal(upcoming[j].DueAt) {
			return upcoming[i].DueAt.Before(upcoming[j].DueAt)
		}
		return upcoming[i].Name < upcoming[j].Name
	})
	return upcoming
}

// Source serves the bills in a file, re-reading it whenever it changes.
type Source struct {
	path string

	mu      sync.Mutex
	stamp   stamp // identifies the version of the file that entries and err came from
	entries []Entry
	err     error
}

// stamp changes whenever the file is edited.
type stamp struct {
	modTime time.Time
	size    int64
}

func NewSource(path string) *Source { return &Source{path: path} }

// Bills returns the bills due next, and how trustworthy they are:
//
//	unconfigured  there is no bills file yet
//	unavailable   the file exists but could not be read or is invalid
//	live          the file was read; an empty list means there are no bills
func (s *Source) Bills(now time.Time) ([]domain.Bill, domain.SourceStatus) {
	empty := []domain.Bill{}
	if s.path == "" {
		return empty, domain.SourceUnconfigured
	}
	info, err := os.Stat(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return empty, domain.SourceUnconfigured
	}
	if err != nil {
		log.Printf("bills file %s: %v", s.path, err)
		return empty, domain.SourceUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if current := (stamp{info.ModTime(), info.Size()}); current != s.stamp {
		s.stamp = current
		s.entries, s.err = s.load()
		if s.err != nil {
			// Logged once per edit, not on every poll.
			log.Printf("bills file %s: %v", s.path, s.err)
		}
	}
	if s.err != nil {
		return empty, domain.SourceUnavailable
	}
	return Upcoming(s.entries, now), domain.SourceLive
}

func (s *Source) load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}
