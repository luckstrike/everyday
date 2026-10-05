package domain

import "time"

type Hour struct {
	Label       string `json:"label"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
}
type ForecastDay struct {
	Date          time.Time `json:"date"`
	High          int       `json:"high"`
	Low           int       `json:"low"`
	Condition     string    `json:"condition"`
	Precipitation int       `json:"precipitation"`
}
type LocalEvent struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Venue    string    `json:"venue"`
	StartsAt time.Time `json:"startsAt"`
	Category string    `json:"category"`
	Price    string    `json:"price"`
	Reason   string    `json:"reason"`
}
type Weather struct {
	Daily         []ForecastDay `json:"daily"`
	Location      string        `json:"location"`
	Temperature   int           `json:"temperature"`
	Condition     string        `json:"condition"`
	High          int           `json:"high"`
	Low           int           `json:"low"`
	Precipitation int           `json:"precipitation"`
	Tomorrow      string        `json:"tomorrow"`
	Hourly        []Hour        `json:"hourly"`
	Alert         string        `json:"alert"`
}
type Sports struct {
	League    string    `json:"league"`
	Home      string    `json:"home"`
	HomeCode  string    `json:"homeCode"`
	Away      string    `json:"away"`
	AwayCode  string    `json:"awayCode"`
	HomeScore int       `json:"homeScore"`
	AwayScore int       `json:"awayScore"`
	Status    string    `json:"status"`
	Detail    string    `json:"detail"`
	StartsAt  time.Time `json:"startsAt"`
}
type Bill struct {
	Name    string    `json:"name"`
	Amount  float64   `json:"amount"`
	DueAt   time.Time `json:"dueAt"`
	Autopay bool      `json:"autopay"`
}
type Finance struct {
	Bills []Bill `json:"bills"`
}
type CalendarEvent struct {
	Title    string    `json:"title"`
	Location string    `json:"location"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}
type AttentionEvent struct {
	ID        string    `json:"id"`
	Widget    string    `json:"widget"`
	Kind      string    `json:"kind"`
	StartsAt  time.Time `json:"startsAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Decay     bool      `json:"decay"`
}
type State struct {
	LocalEvents []LocalEvent     `json:"localEvents"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Mock        bool             `json:"mock"`
	Scenario    string           `json:"scenario"`
	Weather     Weather          `json:"weather"`
	Sports      Sports           `json:"sports"`
	Finance     Finance          `json:"finance"`
	Calendar    []CalendarEvent  `json:"calendar"`
	Events      []AttentionEvent `json:"events"`
}
type Widget struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	BasePriority  int    `json:"basePriority"`
	MinWidth      int    `json:"minWidth"`
	MinHeight     int    `json:"minHeight"`
	DefaultWidth  int    `json:"defaultWidth"`
	DefaultHeight int    `json:"defaultHeight"`
}
type Screen struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	BasePriority  int    `json:"basePriority"`
	CooldownMs    int    `json:"cooldownMs"`
	MinDurationMs int    `json:"minDurationMs"`
	MaxDurationMs int    `json:"maxDurationMs"`
}
type Config struct {
	Screens        []Screen `json:"screens"`
	PollIntervalMs int      `json:"pollIntervalMs"`
	Development    bool     `json:"development"`
}

func Widgets() []Widget {
	return []Widget{
		{"clock", "Time & date", 10, 3, 2, 7, 3},
		{"weather", "Weather", 20, 3, 2, 5, 3},
		{"sports", "Sports", 20, 3, 2, 4, 3},
		{"finance", "Upcoming bills", 25, 3, 2, 4, 3},
		{"calendar", "Calendar", 25, 3, 2, 4, 3},
		{"discover", "Around Dallas", 15, 3, 2, 4, 3},
	}
}
func Configuration(dev bool) Config {
	return Config{[]Screen{
		{"home", "Home", 10, 0, 10000, 20000},
		{"weather", "Weather", 20, 65000, 10000, 25000},
		{"sports", "Sports", 20, 65000, 10000, 25000},
		{"finance", "Finance", 25, 65000, 10000, 25000},
	}, 5000, dev}
}
