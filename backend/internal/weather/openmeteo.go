// Package weather fetches a real forecast from Open-Meteo (no API key needed)
// and keeps the last good result so the dashboard never waits on the network.
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"everyday/backend/internal/domain"
)

const defaultBaseURL = "https://api.open-meteo.com/v1/forecast"

const (
	forecastDays  = 5
	forecastHours = 5
)

// Location is where the forecast is for. Name is only a display label.
type Location struct {
	Name      string
	Latitude  float64
	Longitude float64
}

// Client fetches the forecast for one location from Open-Meteo.
type Client struct {
	Location Location
	// BaseURL and HTTP are overridable for tests; zero values use Open-Meteo.
	BaseURL string
	HTTP    *http.Client
}

type response struct {
	UTCOffsetSeconds int `json:"utc_offset_seconds"`
	Current          struct {
		Temperature float64 `json:"temperature_2m"`
		Code        int     `json:"weather_code"`
		// IsDay is 1 in daylight and 0 at night; a pointer, so a response that omits it reads as daytime.
		IsDay *int `json:"is_day"`
	} `json:"current"`
	Hourly struct {
		Temperature []float64 `json:"temperature_2m"`
		Code        []int     `json:"weather_code"`
	} `json:"hourly"`
	Daily struct {
		Time          []string  `json:"time"`
		Code          []int     `json:"weather_code"`
		High          []float64 `json:"temperature_2m_max"`
		Low           []float64 `json:"temperature_2m_min"`
		Precipitation []float64 `json:"precipitation_probability_max"`
	} `json:"daily"`
}

// Fetch returns the current conditions and forecast, in degrees Fahrenheit.
func (c *Client) Fetch(ctx context.Context) (domain.Weather, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	query := url.Values{
		"latitude":         {strconv.FormatFloat(c.Location.Latitude, 'f', -1, 64)},
		"longitude":        {strconv.FormatFloat(c.Location.Longitude, 'f', -1, 64)},
		"current":          {"temperature_2m,weather_code,is_day"},
		"hourly":           {"temperature_2m,weather_code"},
		"daily":            {"weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max"},
		"temperature_unit": {"fahrenheit"},
		"timezone":         {"auto"},
		"forecast_days":    {strconv.Itoa(forecastDays)},
		"forecast_hours":   {strconv.Itoa(forecastHours)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+query.Encode(), nil)
	if err != nil {
		return domain.Weather{}, err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return domain.Weather{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return domain.Weather{}, fmt.Errorf("open-meteo returned %s", res.Status)
	}
	var body response
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return domain.Weather{}, fmt.Errorf("decode open-meteo response: %w", err)
	}
	return normalize(body, c.Location.Name)
}

// normalize turns Open-Meteo's column-oriented arrays into the dashboard's models.
func normalize(r response, location string) (domain.Weather, error) {
	d, h := r.Daily, r.Hourly
	days := len(d.Time)
	if days == 0 || len(d.Code) != days || len(d.High) != days || len(d.Low) != days || len(d.Precipitation) != days {
		return domain.Weather{}, fmt.Errorf("open-meteo daily forecast is missing or inconsistent")
	}
	if len(h.Temperature) == 0 || len(h.Code) != len(h.Temperature) {
		return domain.Weather{}, fmt.Errorf("open-meteo hourly forecast is missing or inconsistent")
	}
	// Daily dates are local calendar days, so anchor them at local midnight.
	zone := time.FixedZone("", r.UTCOffsetSeconds)
	daily := make([]domain.ForecastDay, days)
	for i := range daily {
		date, err := time.ParseInLocation("2006-01-02", d.Time[i], zone)
		if err != nil {
			return domain.Weather{}, fmt.Errorf("open-meteo daily date: %w", err)
		}
		daily[i] = domain.ForecastDay{
			Date:          date,
			High:          round(d.High[i]),
			Low:           round(d.Low[i]),
			Condition:     condition(d.Code[i]),
			Icon:          icon(d.Code[i]),
			Precipitation: round(d.Precipitation[i]),
		}
	}
	hourly := make([]domain.Hour, len(h.Temperature))
	for i := range hourly {
		label := "Now"
		if i > 0 {
			label = fmt.Sprintf("+%dh", i)
		}
		hourly[i] = domain.Hour{Label: label, Temperature: round(h.Temperature[i]), Condition: condition(h.Code[i])}
	}
	return domain.Weather{
		Location:      location,
		Temperature:   round(r.Current.Temperature),
		Condition:     condition(r.Current.Code),
		Icon:          currentIcon(r.Current.Code, r.Current.IsDay),
		High:          daily[0].High,
		Low:           daily[0].Low,
		Precipitation: daily[0].Precipitation,
		Tomorrow:      tomorrow(daily),
		Daily:         daily,
		Hourly:        hourly,
		Alert:         "", // Open-Meteo has no alerts; a later source can fill this in.
	}, nil
}

func round(v float64) int { return int(math.Round(v)) }

// tomorrow writes a one-line outlook such as "Warmer. Showers with a high of 70°."
func tomorrow(daily []domain.ForecastDay) string {
	if len(daily) < 2 {
		return ""
	}
	today, next := daily[0], daily[1]
	trend := "About the same."
	switch diff := next.High - today.High; {
	case diff >= 4:
		trend = "Warmer."
	case diff <= -4:
		trend = "Cooler."
	}
	return fmt.Sprintf("%s %s with a high of %d°.", trend, next.Condition, next.High)
}

// icon picks the picture for a WMO weather code. Like condition it reads the code,
// not the wording, and anything unrecognised falls back to a plain cloud.
func icon(code int) domain.WeatherIcon {
	switch code {
	case 0, 1:
		return domain.IconClear
	case 2:
		return domain.IconPartlyCloud
	case 45, 48:
		return domain.IconFog
	case 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82:
		return domain.IconRain
	case 71, 73, 75, 77, 85, 86:
		return domain.IconSnow
	case 95, 96, 99:
		return domain.IconStorm
	}
	return domain.IconCloudy
}

// currentIcon is the picture for the conditions right now. Those also depend on
// whether the sun is up: a clear or partly cloudy sky is drawn with a moon at
// night. Forecast days stay daytime pictures, since a day is not a moment.
func currentIcon(code int, isDay *int) domain.WeatherIcon {
	kind := icon(code)
	if isDay == nil || *isDay != 0 {
		return kind
	}
	switch kind {
	case domain.IconClear:
		return domain.IconClearNight
	case domain.IconPartlyCloud:
		return domain.IconPartlyCloudNight
	}
	return kind
}

// condition names a WMO weather code, the scheme Open-Meteo reports.
func condition(code int) string {
	switch code {
	case 0:
		return "Clear skies"
	case 1:
		return "Mostly clear"
	case 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45, 48:
		return "Fog"
	case 51, 53, 55:
		return "Drizzle"
	case 56, 57:
		return "Freezing drizzle"
	case 61, 63, 65:
		return "Rain"
	case 66, 67:
		return "Freezing rain"
	case 71, 73, 75, 77:
		return "Snow"
	case 80, 81, 82:
		return "Showers"
	case 85, 86:
		return "Snow showers"
	case 95:
		return "Thunderstorms"
	case 96, 99:
		return "Thunderstorms with hail"
	}
	return "Unknown"
}
