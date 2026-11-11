package domain

import "time"

type Hour struct {
	Label       string `json:"label"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
}

// WeatherIcon names a kind of weather the frontend can draw. It is separate from the
// display text, so the UI never has to read words like "Showers" or "Sunny".
type WeatherIcon string

const (
	IconClear            WeatherIcon = "clear"
	IconClearNight       WeatherIcon = "clear-night"
	IconPartlyCloud      WeatherIcon = "partly-cloudy"
	IconPartlyCloudNight WeatherIcon = "partly-cloudy-night"
	IconCloudy           WeatherIcon = "cloudy"
	IconFog              WeatherIcon = "fog"
	IconRain             WeatherIcon = "rain"
	IconSnow             WeatherIcon = "snow"
	IconStorm            WeatherIcon = "storm"
)

type ForecastDay struct {
	Date          time.Time   `json:"date"`
	High          int         `json:"high"`
	Low           int         `json:"low"`
	Condition     string      `json:"condition"`
	Icon          WeatherIcon `json:"icon"`
	Precipitation int         `json:"precipitation"`
}
type Weather struct {
	Daily         []ForecastDay `json:"daily"`
	Location      string        `json:"location"`
	Temperature   int           `json:"temperature"`
	Condition     string        `json:"condition"`
	Icon          WeatherIcon   `json:"icon"`
	High          int           `json:"high"`
	Low           int           `json:"low"`
	Precipitation int           `json:"precipitation"`
	Tomorrow      string        `json:"tomorrow"`
	Hourly        []Hour        `json:"hourly"`
	Alert         string        `json:"alert"`
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
	// AllDay events cover whole local days; EndsAt is the midnight after the last day.
	AllDay bool `json:"allDay"`
}

// SourceStatus says where a section of State came from, so the UI can tell
// real data from placeholder data and from nothing at all.
type SourceStatus string

const (
	// SourceLive is data fetched from a real service.
	SourceLive SourceStatus = "live"
	// SourceMock is made-up development data.
	SourceMock SourceStatus = "mock"
	// SourceUnavailable means a source exists but has no usable data right now.
	SourceUnavailable SourceStatus = "unavailable"
	// SourceUnconfigured means nothing is hooked up for this section yet.
	SourceUnconfigured SourceStatus = "unconfigured"
)

type Sources struct {
	Weather  SourceStatus `json:"weather"`
	Finance  SourceStatus `json:"finance"`
	Calendar SourceStatus `json:"calendar"`
}
type State struct {
	GeneratedAt time.Time       `json:"generatedAt"`
	Sources     Sources         `json:"sources"`
	Weather     Weather         `json:"weather"`
	Finance     Finance         `json:"finance"`
	Calendar    []CalendarEvent `json:"calendar"`
}
type Config struct {
	PollIntervalMs int `json:"pollIntervalMs"`
}

func Configuration() Config {
	return Config{PollIntervalMs: 5000}
}
