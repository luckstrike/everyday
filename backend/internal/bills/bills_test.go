package bills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

// saveTime is the moment bills are saved in tests, which stamps any bill without an added date.
var saveTime = time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)

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
		{Name: "Internet", Amount: 65, Day: 15},
		{Name: "Electricity", Amount: 84.5, Day: 15},
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
	if !got[3].DueAt.Equal(date(2026, 12, 1)) || got[0].Amount != 40 {
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
	  {"name": "Internet", "amount": 65.5, "day": 15, "paidFor": "2026-11-15"}
	]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name != "Rent" || entries[0].PaidFor != "" || entries[1].Amount != 65.5 || entries[1].PaidFor != "2026-11-15" {
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
		"not json":                `{oops`,
		"a bare list":             `[{"name": "Rent", "amount": 1, "day": 1}]`,
		"misspelled field":        `{"bills": [{"name": "Rent", "amout": 1850, "day": 1}]}`,
		"unknown section":         `{"bill": []}`,
		"no name":                 `{"bills": [{"amount": 10, "day": 1}]}`,
		"blank name":              `{"bills": [{"name": "  ", "amount": 10, "day": 1}]}`,
		"day zero":                `{"bills": [{"name": "Rent", "amount": 10, "day": 0}]}`,
		"day missing":             `{"bills": [{"name": "Rent", "amount": 10}]}`,
		"day 32":                  `{"bills": [{"name": "Rent", "amount": 10, "day": 32}]}`,
		"negative amount":         `{"bills": [{"name": "Rent", "amount": -5, "day": 1}]}`,
		"amount as text":          `{"bills": [{"name": "Rent", "amount": "1850", "day": 1}]}`,
		"paidFor not a date":      `{"bills": [{"name": "Rent", "amount": 1, "day": 1, "paidFor": "tomorrow"}]}`,
		"paidFor impossible date": `{"bills": [{"name": "Rent", "amount": 1, "day": 1, "paidFor": "2026-02-30"}]}`,
		"addedOn not a date":      `{"bills": [{"name": "Rent", "amount": 1, "day": 1, "addedOn": "yesterday"}]}`,
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
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1850, "day": 1}, {"name": "Phone", "amount": 40, "day": 20}]}`)
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
		{Name: "Internet", Amount: 65.5, Day: 15, PaidFor: "2026-11-15"},
	}, saveTime)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("stat: %v, %v", info, err)
	}
	// What was saved is what the dashboard serves, with no restart in between.
	got, status := source.Bills(now)
	// Internet is paid for the 15th, so the unpaid Rent comes first even though it is due later.
	if status != domain.SourceLive || len(got) != 2 || got[0].Name != "Rent" || got[0].Paid || got[1].Name != "Internet" || !got[1].Paid {
		t.Fatalf("bills after save: %+v, %q", got, status)
	}
	// And the editor sees them in file order, with names tidied.
	entries, problem := source.Entries(now)
	if problem != "" || len(entries) != 2 || entries[0].Name != "Rent" || entries[1].PaidFor != "2026-11-15" || entries[1].Amount != 65.5 {
		t.Fatalf("entries: %+v, %q", entries, problem)
	}
}

func TestSaveRejectsInvalidBillsWithoutTouchingTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save([]Entry{{Name: "Rent", Amount: 1850, Day: 1}}, saveTime); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	for name, bad := range map[string][]Entry{
		"blank name":      {{Name: " ", Amount: 1, Day: 1}},
		"day too small":   {{Name: "Rent", Amount: 1, Day: 0}},
		"day too large":   {{Name: "Rent", Amount: 1, Day: 32}},
		"negative amount": {{Name: "Rent", Amount: -1, Day: 1}},
	} {
		err := source.Save(bad, saveTime)
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

	entries, problem := source.Entries(now)
	if len(entries) != 0 || problem == "" || !strings.Contains(problem, "day") {
		t.Fatalf("a broken file should be reported, not hidden: %+v, %q", entries, problem)
	}
	if _, status := source.Bills(now); status != domain.SourceUnavailable {
		t.Fatalf("status %q", status)
	}
	// The editor is how someone recovers from a mistake in the file.
	if err := source.Save([]Entry{{Name: "Rent", Amount: 1850, Day: 1}}, saveTime); err != nil {
		t.Fatal(err)
	}
	if got, status := source.Bills(now); status != domain.SourceLive || len(got) != 1 {
		t.Fatalf("after replacing: %+v, %q", got, status)
	}
}

func TestEntriesWithNoFileAreEmptyAndNotAProblem(t *testing.T) {
	entries, problem := NewSource(filepath.Join(t.TempDir(), "none.json")).Entries(time.Now())
	if entries == nil || len(entries) != 0 || problem != "" {
		t.Fatalf("got %#v, %q", entries, problem)
	}
	if entries, problem := NewSource("").Entries(time.Now()); entries == nil || problem != "" {
		t.Fatalf("no path: %#v, %q", entries, problem)
	}
	if err := NewSource("").Save(nil, saveTime); err == nil || IsInvalid(err) {
		t.Fatalf("saving with no path is a setup error, not invalid bills: %v", err)
	}
}

func TestSavingNoBillsWritesAnEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save(nil, saveTime); err != nil {
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

func at(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 9, 0, 0, 0, time.Local)
}

func TestAPaidTickOnlyCountsForTheDueDateItWasMadeFor(t *testing.T) {
	entries := []Entry{{Name: "Electricity", Amount: 84.5, Day: 18, PaidFor: "2026-11-18"}}
	for _, c := range []struct {
		now  time.Time
		paid bool
		due  time.Time
		why  string
	}{
		{at(2026, 11, 14), true, date(2026, 11, 18), "paid ahead of the due date"},
		{at(2026, 11, 18), true, date(2026, 11, 18), "still paid on the day itself"},
		{at(2026, 11, 19), false, date(2026, 12, 18), "the cycle rolled over, so the tick has cleared itself"},
		{at(2026, 12, 18), false, date(2026, 12, 18), "next month's bill is not paid by last month's tick"},
	} {
		got := Upcoming(entries, c.now)
		if got[0].Paid != c.paid || !got[0].DueAt.Equal(c.due) {
			t.Errorf("%s: paid %v due %v, want paid %v due %v", c.why, got[0].Paid, got[0].DueAt.Format("2006-01-02"), c.paid, c.due.Format("2006-01-02"))
		}
	}
}

func TestBillsStillToPayComeBeforePaidOnes(t *testing.T) {
	now := at(2026, 11, 11)
	got := Upcoming([]Entry{
		{Name: "Internet", Amount: 65, Day: 12, PaidFor: "2026-11-12"}, // soonest, but paid
		{Name: "Rent", Amount: 1850, Day: 20},
		{Name: "Phone", Amount: 40, Day: 14},
		{Name: "Water", Amount: 30, Day: 13, PaidFor: "2026-11-13"}, // paid
	}, now)
	var names []string
	for _, b := range got {
		names = append(names, b.Name)
	}
	// What is left to pay first, soonest due; then what is already paid, soonest due.
	if strings.Join(names, ",") != "Phone,Rent,Internet,Water" {
		t.Fatalf("got %v", names)
	}
}

func TestTheEditorSeesTheCurrentDueDateAndPaidState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save([]Entry{
		{Name: "Electricity", Amount: 84.5, Day: 18, PaidFor: "2026-11-18"},
		{Name: "Rent", Amount: 1850, Day: 1},
	}, saveTime); err != nil {
		t.Fatal(err)
	}
	listed, problem := source.Entries(at(2026, 11, 14))
	if problem != "" || len(listed) != 2 {
		t.Fatalf("%+v, %q", listed, problem)
	}
	// In file order, each with the date it is counting toward.
	if listed[0].Name != "Electricity" || listed[0].Due != "2026-11-18" || !listed[0].Paid {
		t.Fatalf("electricity: %+v", listed[0])
	}
	if listed[1].Name != "Rent" || listed[1].Due != "2026-12-01" || listed[1].Paid {
		t.Fatalf("rent: %+v", listed[1])
	}
	// A week later the cycle has rolled, and the same tick no longer counts.
	listed, _ = source.Entries(at(2026, 11, 25))
	if listed[0].Due != "2026-12-18" || listed[0].Paid {
		t.Fatalf("after the cycle rolled: %+v", listed[0])
	}
}

func TestFilesWrittenWithTheRetiredAutopayFlagStillWork(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	writeBills(t, path, `{"bills": [
	  {"name": "Rent", "amount": 1900, "day": 1, "autopay": false},
	  {"name": "Internet", "amount": 60.53, "day": 28, "autopay": true}
	]}`)
	source := NewSource(path)
	now := at(2026, 11, 14)

	// The dashboard keeps working, ignoring the flag.
	got, status := source.Bills(now)
	if status != domain.SourceLive || len(got) != 2 {
		t.Fatalf("a file with the retired flag must still be live: %+v, %q", got, status)
	}
	listed, problem := source.Entries(now)
	if problem != "" || len(listed) != 2 || listed[1].Amount != 60.53 {
		t.Fatalf("editor: %+v, %q", listed, problem)
	}
	// Saving writes the bills back without it, so the file tidies itself.
	if err := source.Save([]Entry{listed[0].Entry, listed[1].Entry}, saveTime); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "autopay") {
		t.Fatalf("the retired flag should be gone after a save: %s", data)
	}
	if !strings.Contains(string(data), `"name": "Internet"`) {
		t.Fatalf("the bills themselves must survive: %s", data)
	}
}

func TestAnUnpaidBillWritesNoPaidForField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save([]Entry{{Name: "Rent", Amount: 1, Day: 1}, {Name: "Phone", Amount: 1, Day: 2, PaidFor: "2026-11-02"}}, saveTime); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "paidFor") != 1 {
		t.Fatalf("only the paid bill should carry paidFor: %s", data)
	}
}

func TestPrevDue(t *testing.T) {
	cases := []struct {
		name string
		day  int
		now  time.Time
		want time.Time
	}{
		{"earlier this month", 5, at(2026, 11, 11), date(2026, 11, 5)},
		{"today has not passed yet", 11, at(2026, 11, 11), date(2026, 10, 11)},
		{"later this month, so last month's", 20, at(2026, 11, 11), date(2026, 10, 20)},
		{"across the new year", 20, at(2027, 1, 5), date(2026, 12, 20)},
		{"the 31st in a 30-day month", 31, at(2026, 12, 5), date(2026, 11, 30)},
		{"the 31st, March looking back at February", 31, at(2026, 3, 2), date(2026, 2, 28)},
		{"the 29th, March looking back at a leap February", 29, at(2028, 3, 1), date(2028, 2, 29)},
	}
	for _, c := range cases {
		if got := PrevDue(c.day, c.now); !got.Equal(c.want) {
			t.Errorf("%s: PrevDue(%d, %v) = %v, want %v", c.name, c.day, c.now.Format("2006-01-02"), got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func first(e Entry, now time.Time) domain.Bill { return Upcoming([]Entry{e}, now)[0] }

func TestABillWithNoAddedDateIsNeverOverdue(t *testing.T) {
	got := first(Entry{Name: "Phone", Amount: 45, Day: 10}, at(2026, 11, 14))
	if got.Overdue || got.DaysLate != 0 || !got.DueAt.Equal(date(2026, 12, 10)) {
		t.Fatalf("an untracked bill whose day has passed just moves to next month: %+v", got)
	}
}

func TestAMissedDueDateMakesABillOverdueUntilItIsPaid(t *testing.T) {
	e := Entry{Name: "Phone", Amount: 45, Day: 10, AddedOn: "2026-11-01"}
	got := first(e, at(2026, 11, 14))
	if !got.Overdue || got.Paid || got.DaysLate != 4 || !got.DueAt.Equal(date(2026, 11, 10)) {
		t.Fatalf("overdue since the 10th, four days ago: %+v", got)
	}
	// Paying settles it, and the next due date takes over.
	e.PaidFor = "2026-11-10"
	got = first(e, at(2026, 11, 14))
	if got.Overdue || got.DaysLate != 0 || got.Paid || !got.DueAt.Equal(date(2026, 12, 10)) {
		t.Fatalf("after paying, next month's bill: %+v", got)
	}
}

func TestABillIsNotOverdueForADateBeforeItWasAdded(t *testing.T) {
	// Added on the 14th with a due day of the 10th: this month's 10th was never missed.
	got := first(Entry{Name: "New", Amount: 1, Day: 10, AddedOn: "2026-11-14"}, at(2026, 11, 14))
	if got.Overdue || !got.DueAt.Equal(date(2026, 12, 10)) {
		t.Fatalf("a new bill starts clean: %+v", got)
	}
}

func TestABillAddedOnItsDueDayIsOverdueOnlyOnceTheDayHasPassed(t *testing.T) {
	e := Entry{Name: "Phone", Amount: 45, Day: 10, AddedOn: "2026-11-10"}
	if got := first(e, at(2026, 11, 10)); got.Overdue || !got.DueAt.Equal(date(2026, 11, 10)) {
		t.Fatalf("still due today, not late: %+v", got)
	}
	if got := first(e, at(2026, 11, 11)); !got.Overdue || got.DaysLate != 1 {
		t.Fatalf("the next day it is one day late: %+v", got)
	}
}

func TestOneTickSettlesEveryMissedMonth(t *testing.T) {
	e := Entry{Name: "Phone", Amount: 45, Day: 10, AddedOn: "2026-09-01"}
	now := at(2026, 11, 14)

	got := first(e, now)
	// September, October and November were all missed; it is shown as late since the first.
	if !got.Overdue || !got.DueAt.Equal(date(2026, 9, 10)) || got.DaysLate != 65 {
		t.Fatalf("overdue since September: %+v", got)
	}
	// Ticking records the most recent missed date, which settles the earlier ones too.
	listed := list([]Entry{e}, now)[0]
	if listed.PayFor != "2026-11-10" {
		t.Fatalf("ticking should record the latest missed date, got %s", listed.PayFor)
	}
	e.PaidFor = listed.PayFor
	if got := first(e, now); got.Overdue || !got.DueAt.Equal(date(2026, 12, 10)) {
		t.Fatalf("one tick should clear all three: %+v", got)
	}
}

func TestPayingForALaterDateSettlesTheEarlierOnes(t *testing.T) {
	// Added in September, paid through October: only November is missed.
	e := Entry{Name: "Phone", Amount: 45, Day: 10, AddedOn: "2026-09-01", PaidFor: "2026-10-10"}
	got := first(e, at(2026, 11, 14))
	if !got.Overdue || !got.DueAt.Equal(date(2026, 11, 10)) || got.DaysLate != 4 {
		t.Fatalf("%+v", got)
	}
}

func TestPayingAheadOfTheDueDateIsNotOverdue(t *testing.T) {
	e := Entry{Name: "Electricity", Amount: 84.5, Day: 18, AddedOn: "2026-10-01", PaidFor: "2026-11-18"}
	got := first(e, at(2026, 11, 14))
	if got.Overdue || !got.Paid || !got.DueAt.Equal(date(2026, 11, 18)) {
		t.Fatalf("%+v", got)
	}
}

func TestUntickingGoesBackToThePreviousDateSoNothingPaidBecomesOverdue(t *testing.T) {
	// Paid ahead for the 18th, in a bill that has been tracked since October.
	e := Entry{Name: "Electricity", Amount: 84.5, Day: 18, AddedOn: "2026-10-01", PaidFor: "2026-11-18"}
	now := at(2026, 11, 14)
	listed := list([]Entry{e}, now)[0]
	if !listed.Paid || listed.UnpayTo != "2026-10-18" {
		t.Fatalf("%+v", listed)
	}
	// Forgetting the paid date altogether would make October, which was paid, look
	// missed. Going back to October's date does not.
	e.PaidFor = listed.UnpayTo
	got := first(e, now)
	if got.Overdue || got.Paid || !got.DueAt.Equal(date(2026, 11, 18)) {
		t.Fatalf("unticked: back to unpaid, not overdue: %+v", got)
	}
	e.PaidFor = ""
	if !first(e, now).Overdue {
		t.Fatal("this is the mistake the previous date avoids: with no date, October looks missed")
	}
}

func TestMissedDatesFollowShortMonths(t *testing.T) {
	// Due on the 31st, tracked from the 1st: November's date is the 30th.
	e := Entry{Name: "Loan", Amount: 100, Day: 31, AddedOn: "2026-11-01"}
	got := first(e, at(2026, 12, 2))
	if !got.Overdue || !got.DueAt.Equal(date(2026, 11, 30)) || got.DaysLate != 2 {
		t.Fatalf("%+v", got)
	}
}

func TestDaysLateCountCalendarDaysAcrossAClockChange(t *testing.T) {
	chicago, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Skip("no time zone data on this machine")
	}
	// Daylight saving ended on 1 November 2026, making that day 25 hours long.
	e := Entry{Name: "Phone", Amount: 45, Day: 30, AddedOn: "2026-10-01"}
	for _, hour := range []int{0, 1, 12, 23} {
		now := time.Date(2026, 11, 3, hour, 30, 0, 0, chicago)
		got := first(e, now)
		if !got.Overdue || got.DaysLate != 4 || got.DueAt.Format("2006-01-02") != "2026-10-30" {
			t.Errorf("at %02d:30: %+v", hour, got)
		}
	}
}

func TestMissedBillsLeadThenBillsStillToPayThenPaidOnes(t *testing.T) {
	now := at(2026, 11, 14)
	got := Upcoming([]Entry{
		{Name: "Paid", Amount: 1, Day: 15, PaidFor: "2026-11-15"},
		{Name: "Upcoming", Amount: 1, Day: 20},
		{Name: "Late4", Amount: 1, Day: 10, AddedOn: "2026-11-01"},
		{Name: "Late6", Amount: 1, Day: 8, AddedOn: "2026-11-01"},
	}, now)
	var names []string
	for _, b := range got {
		names = append(names, b.Name)
	}
	// The longest overdue first, then what is still to pay, then what is paid.
	if strings.Join(names, ",") != "Late6,Late4,Upcoming,Paid" {
		t.Fatalf("got %v", names)
	}
}

func TestTheEditorIsToldWhatToRecordForEachBill(t *testing.T) {
	now := at(2026, 11, 14)
	listed := list([]Entry{
		{Name: "Late", Amount: 1, Day: 10, AddedOn: "2026-11-01"},
		{Name: "Upcoming", Amount: 1, Day: 20},
	}, now)
	late, upcoming := listed[0], listed[1]
	if !late.Overdue || late.Due != "2026-11-10" || late.DaysLate != 4 || late.PayFor != "2026-11-10" || late.Paid {
		t.Fatalf("late: %+v", late)
	}
	if upcoming.Overdue || upcoming.Due != "2026-11-20" || upcoming.PayFor != "2026-11-20" || upcoming.UnpayTo != "2026-10-20" {
		t.Fatalf("upcoming: %+v", upcoming)
	}
}

func TestSavingStampsTheAddedDateOnlyWhereItIsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	source := NewSource(path)
	if err := source.Save([]Entry{
		{Name: "Old", Amount: 1, Day: 5, AddedOn: "2026-09-01"},
		{Name: "New", Amount: 1, Day: 6},
	}, saveTime); err != nil {
		t.Fatal(err)
	}
	listed, problem := source.Entries(saveTime)
	if problem != "" || len(listed) != 2 {
		t.Fatalf("%+v, %q", listed, problem)
	}
	if listed[0].AddedOn != "2026-09-01" {
		t.Fatalf("a bill keeps the date it was added, not today's: %q", listed[0].AddedOn)
	}
	if listed[1].AddedOn != "2026-11-11" {
		t.Fatalf("a bill with no date is stamped with the day it is saved: %q", listed[1].AddedOn)
	}
}

func TestAHandWrittenFileStartsBeingTrackedOnlyOnceItIsSaved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bills.json")
	writeBills(t, path, `{"bills": [{"name": "Rent", "amount": 1900, "day": 1}]}`)
	source := NewSource(path)
	now := at(2026, 11, 14)

	// Untracked: the 1st has passed, but nothing flashes overdue.
	got, _ := source.Bills(now)
	if got[0].Overdue {
		t.Fatal("an existing bill must not turn overdue just because tracking now exists")
	}
	// Saving it from Settings starts the tracking from that day, still without a false alarm.
	listed, _ := source.Entries(now)
	if err := source.Save([]Entry{listed[0].Entry}, now); err != nil {
		t.Fatal(err)
	}
	if got, _ := source.Bills(now); got[0].Overdue {
		t.Fatalf("tracking starts clean on the day it begins: %+v", got[0])
	}
	// But a due date that passes after that, unpaid, is missed.
	if got, _ := source.Bills(at(2026, 12, 2)); !got[0].Overdue || got[0].DaysLate != 1 {
		t.Fatalf("the 1st of December, missed: %+v", got[0])
	}
}

func TestABillWithNoAmountIsKeptAsDueAndLeavesTheAmountOutOfTheFile(t *testing.T) {
	// A credit card has no fixed amount: a file may leave it out, and it is valid.
	entries, err := Parse([]byte(`{"bills": [{"name": "Amex", "day": 20}]}`))
	if err != nil || len(entries) != 1 || entries[0].Amount != 0 {
		t.Fatalf("parse: %+v, %v", entries, err)
	}
	got := Upcoming(entries, at(2026, 11, 14))
	if len(got) != 1 || got[0].Amount != 0 || !got[0].DueAt.Equal(date(2026, 11, 20)) {
		t.Fatalf("upcoming: %+v", got)
	}

	// Saving it must not write "amount": 0, so the file stays what a person would write.
	path := filepath.Join(t.TempDir(), "bills.json")
	if err := NewSource(path).Save(entries, saveTime); err != nil {
		t.Fatal(err)
	}
	written, _ := os.ReadFile(path)
	if strings.Contains(string(written), "amount") {
		t.Fatalf("the file should not mention an amount: %s", written)
	}
}
