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
type State struct {
	LocalEvents []LocalEvent    `json:"localEvents"`
	GeneratedAt time.Time       `json:"generatedAt"`
	Mock        bool            `json:"mock"`
	Weather     Weather         `json:"weather"`
	Sports      Sports          `json:"sports"`
	Finance     Finance         `json:"finance"`
	Calendar    []CalendarEvent `json:"calendar"`
}
type Config struct {
	PollIntervalMs int `json:"pollIntervalMs"`
}

func Configuration() Config {
	return Config{PollIntervalMs: 5000}
}
