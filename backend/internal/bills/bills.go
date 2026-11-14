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
	"everyday/backend/internal/fileutil"
)

// Entry is one bill as written in the file. Every bill repeats monthly.
type Entry struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Day    int     `json:"day"` // day of the month it is due, 1 to 31
	// PaidFor is the due date (YYYY-MM-DD) this bill was last marked paid for. A bill
	// counts as paid only while that is the due date it is currently counting toward,
	// so a tick clears itself when the next month's cycle begins.
	PaidFor string `json:"paidFor,omitempty"`
}

// file is the shape of the bills file. It is an object, not a bare list, so
// other fields can be added later without breaking existing files.
type file struct {
	Bills []Entry `json:"bills"`
}

// readFile is what is accepted when reading. It adds one setting that is read and
// then dropped: bills used to carry an autopay flag, and files written back then
// must keep working. Saving writes the entries out without it.
type readFile struct {
	Bills []struct {
		Entry
		Autopay *bool `json:"autopay"` // retired; ignored
	} `json:"bills"`
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

// invalidError is a problem with the content of the bills, as opposed to a problem
// reading or writing the file. Callers can tell the two apart with IsInvalid.
type invalidError string

func (e invalidError) Error() string { return string(e) }

// IsInvalid reports whether err says the bills themselves are wrong, so the person
// who supplied them can fix it.
func IsInvalid(err error) bool {
	var invalid invalidError
	return errors.As(err, &invalid)
}

// Parse reads and validates the contents of a bills file. A misspelled field is an
// error rather than silently ignored.
func Parse(data []byte) ([]Entry, error) {
	var f readFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&f); err != nil {
		return nil, invalidError(err.Error())
	}
	entries := make([]Entry, len(f.Bills))
	for i, b := range f.Bills {
		entries[i] = b.Entry
	}
	return Validate(entries)
}

// dateKey writes a date the way PaidFor does.
func dateKey(t time.Time) string { return t.Format("2006-01-02") }

// Validate checks entries and returns them tidied (names trimmed). It does not
// change the slice it is given.
func Validate(entries []Entry) ([]Entry, error) {
	tidy := make([]Entry, len(entries))
	for i, e := range entries {
		switch {
		case strings.TrimSpace(e.Name) == "":
			return nil, invalidError(fmt.Sprintf("bill %d has no name", i+1))
		case e.Day < 1 || e.Day > 31:
			return nil, invalidError(fmt.Sprintf("bill %q: day must be 1 to 31, got %d", e.Name, e.Day))
		case math.IsNaN(e.Amount) || math.IsInf(e.Amount, 0) || e.Amount < 0:
			return nil, invalidError(fmt.Sprintf("bill %q: amount must be zero or more", e.Name))
		}
		if e.PaidFor != "" {
			if _, err := time.Parse("2006-01-02", e.PaidFor); err != nil {
				return nil, invalidError(fmt.Sprintf("bill %q: paidFor must be a date like 2026-11-18, got %q", e.Name, e.PaidFor))
			}
		}
		e.Name = strings.TrimSpace(e.Name)
		tidy[i] = e
	}
	return tidy, nil
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

// Upcoming turns entries into bills with their next due dates. Bills still to pay
// come first, soonest due first, so a display with little room shows what is left to
// do; bills already paid for their current due date follow, in the same order. The
// result is never nil, so it encodes as [] and not null.
func Upcoming(entries []Entry, now time.Time) []domain.Bill {
	upcoming := make([]domain.Bill, 0, len(entries))
	for _, e := range entries {
		due := NextDue(e.Day, now)
		upcoming = append(upcoming, domain.Bill{
			Name:   e.Name,
			Amount: e.Amount,
			DueAt:  due,
			Paid:   e.PaidFor == dateKey(due),
		})
	}
	sort.SliceStable(upcoming, func(i, j int) bool {
		a, b := upcoming[i], upcoming[j]
		if a.Paid != b.Paid {
			return !a.Paid
		}
		if !a.DueAt.Equal(b.DueAt) {
			return a.DueAt.Before(b.DueAt)
		}
		return a.Name < b.Name
	})
	return upcoming
}

// Listed is a bill as the editor sees it: what is in the file, plus the due date it
// is counting toward right now and whether that one has been paid.
type Listed struct {
	Entry
	Due  string `json:"due"` // YYYY-MM-DD
	Paid bool   `json:"paid"`
}

func list(entries []Entry, now time.Time) []Listed {
	listed := make([]Listed, len(entries))
	for i, e := range entries {
		due := dateKey(NextDue(e.Day, now))
		listed[i] = Listed{Entry: e, Due: due, Paid: e.PaidFor == due}
	}
	return listed
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

// Entries returns the bills as written in the file, in file order, for editing, each with the due date it is counting toward at now and whether that one is paid. A
// missing file has no entries and no problem. A file that cannot be read or is
// invalid has no entries and a problem describing why, so the editor can say so and
// still let the person replace it.
func (s *Source) Entries(now time.Time) (bills []Listed, problem string) {
	if s.path == "" {
		return []Listed{}, ""
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return []Listed{}, ""
	}
	if err != nil {
		return []Listed{}, err.Error()
	}
	entries, err := Parse(data)
	if err != nil {
		return []Listed{}, err.Error()
	}
	return list(entries, now), ""
}

// Save replaces the bills file with entries, after checking them. Invalid entries
// are rejected without touching the file (see IsInvalid). The write is atomic, so
// the dashboard never reads a half-written file, and the running Source picks the
// change up like any other edit to the file.
func (s *Source) Save(entries []Entry) error {
	if s.path == "" {
		return errors.New("there is nowhere to save bills: no bills file is configured")
	}
	tidy, err := Validate(entries)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(file{Bills: tidy}, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteAtomic(s.path, append(data, '\n'), 0o600)
}

func (s *Source) load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}
