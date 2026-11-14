package bills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func TestNextDue(t *testing.T) {
	noon := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 12, 30, 0, 0, time.Local) }
	cases := []struct {
		name string
		day  int
		now  time.Time
		want time.Time
	}{
		{"later this month", 20, noon(2026, 11, 11), date(2026, 11, 20)},
		{"due today still counts", 11, noon(2026, 11, 11), date(2026, 11, 11)},
		{"due today, late in the evening", 11, time.Date(2026, 11, 11, 23, 59, 0, 0, time.Local), date(2026, 11, 11)},
		{"yesterday's day rolls to next month", 10, noon(2026, 11, 11), date(2026, 12, 10)},
		{"the 1st, once the 1st has passed", 1, noon(2026, 11, 2), date(2026, 12, 1)},
		{"rolls over the year", 5, noon(2026, 12, 20), date(2027, 1, 5)},
		{"the 31st in a 30-day month", 31, noon(2026, 11, 11), date(2026, 11, 30)},
		{"the 31st after November's last day passes", 31, noon(2026, 11, 30).Add(24 * time.Hour), date(2026, 12, 31)},
		{"the 31st falls on the end of February", 31, noon(2026, 2, 10), date(2026, 2, 28)},
		{"the 29th in a leap February", 29, noon(2028, 2, 10), date(2028, 2, 29)},
		{"the 29th in a normal February", 29, noon(2026, 2, 10), date(2026, 2, 28)},
		{"passed the clamped end of month", 31, noon(2026, 2, 28).Add(24 * time.Hour), date(2026, 3, 31)},
		{"the 30th after a 28-day February clamp", 30, noon(2026, 1, 31), date(2026, 2, 28)},
	}
	for _, c := range cases {
		if got := NextDue(c.day, c.now); !got.Equal(c.want) {
			t.Errorf("%s: NextDue(%d, %v) = %v, want %v", c.name, c.day, c.now.Format("2006-01-02 15:04"), got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func TestUpcomingOrdersBySoonestDueThenName(t *testing.T) {
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)
	got := Upcoming([]Entry{
		{Name: "Rent", Amount: 1850, Day: 1},
		{Name: "Internet", Amount: 65, Day: 15, Autopay: true},
		{Name: "Electricity", Amount: 84.5, Day: 15, Autopay: true},
		{Name: "Phone", Amount: 40, Day: 11},
	}, now)
	var names []string
	for _, b := range got {
		names = append(names, b.Name)
	}
	// Phone is due today, the two on the 15th tie and sort by name, rent is next month.
	if strings.Join(names, ",") != "Phone,Electricity,Internet,Rent" {
		t.Fatalf("got %v", names)
	}
	if !got[3].DueAt.Equal(date(2026, 12, 1)) || got[0].Amount != 40 || !got[1].Autopay {
		t.Fatalf("details: %+v", got)
	}
}

func TestUpcomingWithNoBillsIsAnEmptyListNotNull(t *testing.T) {
	if got := Upcoming(nil, time.Now()); got == nil || len(got) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestParse(t *testing.T) {
	entries, err := Parse([]byte(`{"bills": [
	  {"name": " Rent ", "amount": 1850, "day": 1},
	  {"name": "Internet", "amount": 65.5, "day": 15, "autopay": true}
	]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name != "Rent" || entries[0].Autopay || entries[1].Amount != 65.5 || !entries[1].Autopay {
		t.Fatalf("entries: %+v", entries)
	}
	if entries, err := Parse([]byte(`{"bills": []}`)); err != nil || len(entries) != 0 {
		t.Fatalf("an empty list is fine: %v, %v", entries, err)
	}
	if entries, err := Parse([]byte(`{}`)); err != nil || len(entries) != 0 {
		t.Fatalf("no bills key is fine: %v, %v", entries, err)
	}
}

func TestParseRejectsMistakes(t *testing.T) {
	cases := map[string]string{
		"not json":         `{oops`,
		"a bare list":      `[{"name": "Rent", "amount": 1, "day": 1}]`,
		"misspelled field": `{"bills": [{"name": "Rent", "amout": 1850, "day": 1}]}`,
		"unknown section":  `{"bill": []}`,
		"no name":          `{"bills": [{"amount": 10, "day": 1}]}`,
		"blank name":       `{"bills": [{"name": "  ", "amount": 10, "day": 1}]}`,
		"day zero":         `{"bills": [{"name": "Rent", "amount": 10, "day": 0}]}`,
		"day missing":      `{"bills": [{"name": "Rent", "amount": 10}]}`,
		"day 32":           `{"bills": [{"name": "Rent", "amount": 10, "day": 32}]}`,
		"negative amount":  `{"bills": [{"name": "Rent", "amount": -5, "day": 1}]}`,
		"amount as text":   `{"bills": [{"name": "Rent", "amount": "1850", "day": 1}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(body)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func writeBills(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	// File systems stamp modification times coarsely, so make each edit clearly newer.
	edited := time.Now().Add(time.Duration(len(body)) * time.Hour)
	if err := os.Chtimes(path, edited, edited); err != nil {
		t.Fatal(err)
	}
}

func TestSourceWithNoFileIsUnconfigured(t *testing.T) {
	source := NewSource(filepath.Join(t.TempDir(), "bills.json"))
	got, status := source.Bills(time.Now())
	if status != domain.SourceUnconfigured || got == nil || len(got) != 0 {
		t.Fatalf("got %#v, %q", got, status)
	}
	if _, status := NewSource("").Bills(time.Now()); status != domain.SourceUnconfigured {
		t.Fatalf("no path at all: %q", status)
	}
}

func TestSourceReadsTheFileAndFollowsEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)

	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}]}`)
	got, status := source.Bills(now)
	if status != domain.SourceLive || len(got) != 1 || got[0].Name != "Rent" {
		t.Fatalf("first read: %+v, %q", got, status)
	}

	// Editing the file changes what is served, with no restart.
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}, {"name": "Phone", "amount": 40, "day": 20, "autopay": true}]}`)
	got, status = source.Bills(now)
	if status != domain.SourceLive || len(got) != 2 || got[0].Name != "Phone" {
		t.Fatalf("after edit: %+v, %q", got, status)
	}

	// An empty file is a real answer: there are simply no bills.
	writeBills(t, path, `{"bills": []}`)
	got, status = source.Bills(now)
	if status != domain.SourceLive || got == nil || len(got) != 0 {
		t.Fatalf("empty: %#v, %q", got, status)
	}
}

func TestDueDatesFollowTheClockNotTheEdit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}]}`)
	source := NewSource(path)
	before, _ := source.Bills(time.Date(2026, 11, 30, 9, 0, 0, 0, time.Local))
	after, _ := source.Bills(time.Date(2026, 12, 1, 9, 0, 0, 0, time.Local))
	if !before[0].DueAt.Equal(date(2026, 12, 1)) || !after[0].DueAt.Equal(date(2026, 12, 1)) {
		t.Fatalf("rent due %v then %v; it is still due on the 1st", before[0].DueAt, after[0].DueAt)
	}
	later, _ := source.Bills(time.Date(2026, 12, 2, 9, 0, 0, 0, time.Local))
	if !later[0].DueAt.Equal(date(2027, 1, 1)) {
		t.Fatalf("the day after, rent rolls to next month: %v", later[0].DueAt)
	}
}

func TestSourceReportsAnInvalidFileAndRecovers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)

	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}]}`)
	if _, status := source.Bills(now); status != domain.SourceLive {
		t.Fatalf("valid file: %q", status)
	}

	// A bad edit must show up as unavailable, not as stale bills that hide the mistake.
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 99}]}`)
	got, status := source.Bills(now)
	if status != domain.SourceUnavailable || got == nil || len(got) != 0 {
		t.Fatalf("invalid file: %#v, %q", got, status)
	}
	// Asking again without another edit stays unavailable.
	if _, status := source.Bills(now); status != domain.SourceUnavailable {
		t.Fatalf("still invalid: %q", status)
	}

	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}]}`)
	if got, status := source.Bills(now); status != domain.SourceLive || len(got) != 1 {
		t.Fatalf("after fixing: %+v, %q", got, status)
	}
}

func TestDeletingTheFileMakesItUnconfiguredAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}]}`)
	if _, status := source.Bills(time.Now()); status != domain.SourceLive {
		t.Fatal("expected live")
	}
	os.Remove(path)
	if _, status := source.Bills(time.Now()); status != domain.SourceUnconfigured {
		t.Fatalf("after delete: %q", status)
	}
}

func TestSaveWritesAPrivateFileThatTheSourceReadsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "bills.json")
	source := NewSource(path)
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)

	err := source.Save([]Entry{
		{Name: "  Rent ", Amount: 1850, Day: 1},
		{Name: "Internet", Amount: 65.5, Day: 15, Autopay: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("stat: %v, %v", info, err)
	}
	// What was saved is what the dashboard serves, with no restart in between.
	got, status := source.Bills(now)
	if status != domain.SourceLive || len(got) != 2 || got[0].Name != "Internet" || got[1].Name != "Rent" {
		t.Fatalf("bills after save: %+v, %q", got, status)
	}
	// And the editor sees them in file order, with names tidied.
	entries, problem := source.Entries()
	if problem != "" || len(entries) != 2 || entries[0].Name != "Rent" || !entries[1].Autopay || entries[1].Amount != 65.5 {
		t.Fatalf("entries: %+v, %q", entries, problem)
	}
}

func TestSaveRejectsInvalidBillsWithoutTouchingTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save([]Entry{{Name: "Rent", Amount: 1850, Day: 1}}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	for name, bad := range map[string][]Entry{
		"blank name":      {{Name: " ", Amount: 1, Day: 1}},
		"day too small":   {{Name: "Rent", Amount: 1, Day: 0}},
		"day too large":   {{Name: "Rent", Amount: 1, Day: 32}},
		"negative amount": {{Name: "Rent", Amount: -1, Day: 1}},
	} {
		err := source.Save(bad)
		if err == nil || !IsInvalid(err) {
			t.Fatalf("%s: want an invalid-bills error, got %v", name, err)
		}
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("a rejected save must leave the file exactly as it was")
	}
}

func TestSaveCanReplaceABrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	writeBills(t, path, `{"bills": [{"name": "Rent", "day": 99}]}`)
	source := NewSource(path)
	now := time.Now()

	entries, problem := source.Entries()
	if len(entries) != 0 || problem == "" || !strings.Contains(problem, "day") {
		t.Fatalf("a broken file should be reported, not hidden: %+v, %q", entries, problem)
	}
	if _, status := source.Bills(now); status != domain.SourceUnavailable {
		t.Fatalf("status %q", status)
	}
	// The editor is how someone recovers from a mistake in the file.
	if err := source.Save([]Entry{{Name: "Rent", Amount: 1850, Day: 1}}); err != nil {
		t.Fatal(err)
	}
	if got, status := source.Bills(now); status != domain.SourceLive || len(got) != 1 {
		t.Fatalf("after replacing: %+v, %q", got, status)
	}
}

func TestEntriesWithNoFileAreEmptyAndNotAProblem(t *testing.T) {
	entries, problem := NewSource(filepath.Join(t.TempDir(), "none.json")).Entries()
	if entries == nil || len(entries) != 0 || problem != "" {
		t.Fatalf("got %#v, %q", entries, problem)
	}
	if entries, problem := NewSource("").Entries(); entries == nil || problem != "" {
		t.Fatalf("no path: %#v, %q", entries, problem)
	}
	if err := NewSource("").Save(nil); err == nil || IsInvalid(err) {
		t.Fatalf("saving with no path is a setup error, not invalid bills: %v", err)
	}
}

func TestSavingNoBillsWritesAnEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save(nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"bills": []`) {
		t.Fatalf("expected an empty list, not null: %s", data)
	}
	if _, status := source.Bills(time.Now()); status != domain.SourceLive {
		t.Fatalf("status %q", status)
	}
}

func TestIsInvalidSeparatesBadBillsFromOtherFailures(t *testing.T) {
	if _, err := Parse([]byte(`{oops`)); !IsInvalid(err) {
		t.Fatalf("malformed JSON is a content problem: %v", err)
	}
	if IsInvalid(os.ErrPermission) {
		t.Fatal("a permissions error is not an invalid-bills error")
	}
}
