package weather

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

// A trimmed real Open-Meteo response for Dallas (utc_offset_seconds is -6h).
const forecastJSON = `{
  "utc_offset_seconds": -21600,
  "current": {"temperature_2m": 69.4, "weather_code": 0, "is_day": 1},
  "hourly": {
    "temperature_2m": [69.4, 73.2, 75.7, 77.4, 78.5],
    "weather_code": [0, 0, 0, 2, 3]
  },
  "daily": {
    "time": ["2026-11-11", "2026-11-12", "2026-11-13", "2026-11-14", "2026-11-15"],
    "weather_code": [2, 0, 3, 3, 2],
    "temperature_2m_max": [78.8, 76.2, 72.9, 64.2, 70.1],
    "temperature_2m_min": [59.5, 51.4, 48.7, 53.4, 52.2],
    "precipitation_probability_max": [1, 0, 0, 17, 10]
  }
}`

func serve(t *testing.T, status int, body string) (*Client, *http.Request) {
	t.Helper()
	var got http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(context.Background())
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	client := &Client{Location: Location{Name: "Dallas, TX", Latitude: 32.7767, Longitude: -96.797}, BaseURL: server.URL}
	return client, &got
}

func TestFetchNormalizesForecast(t *testing.T) {
	client, _ := serve(t, 200, forecastJSON)
	w, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.Location != "Dallas, TX" || w.Temperature != 69 || w.Condition != "Clear skies" {
		t.Fatalf("current conditions: %+v", w)
	}
	// 59.5 rounds away from zero to 60; today's numbers come from the first day.
	if w.High != 79 || w.Low != 60 || w.Precipitation != 1 {
		t.Fatalf("today: high %d low %d precipitation %d", w.High, w.Low, w.Precipitation)
	}
	if w.Tomorrow != "About the same. Clear skies with a high of 76°." {
		t.Fatalf("tomorrow: %q", w.Tomorrow)
	}
	if w.Alert != "" {
		t.Fatalf("alert: %q", w.Alert)
	}
	if len(w.Daily) != 5 || len(w.Hourly) != 5 {
		t.Fatalf("got %d days and %d hours", len(w.Daily), len(w.Hourly))
	}
	wantDate := time.Date(2026, 11, 12, 0, 0, 0, 0, time.FixedZone("", -6*3600))
	if !w.Daily[1].Date.Equal(wantDate) {
		t.Fatalf("second day is %v, want local midnight %v", w.Daily[1].Date, wantDate)
	}
	if w.Hourly[0].Label != "Now" || w.Hourly[2].Label != "+2h" || w.Hourly[3].Condition != "Partly cloudy" {
		t.Fatalf("hourly: %+v", w.Hourly)
	}
	// The fixture's current weather code is 0, clear sky.
	if w.Icon != domain.IconClear {
		t.Fatalf("current icon %q, want %q", w.Icon, domain.IconClear)
	}
	// The fixture's daily weather codes are 2, 0, 3, 3, 2.
	var icons []domain.WeatherIcon
	for _, day := range w.Daily {
		icons = append(icons, day.Icon)
	}
	want := []domain.WeatherIcon{domain.IconPartlyCloud, domain.IconClear, domain.IconCloudy, domain.IconCloudy, domain.IconPartlyCloud}
	if fmt.Sprint(icons) != fmt.Sprint(want) {
		t.Fatalf("daily icons %v, want %v", icons, want)
	}
}

func TestCurrentConditionsUseAMoonAtNight(t *testing.T) {
	day, night := 1, 0
	cases := []struct {
		name  string
		code  int
		isDay *int
		want  domain.WeatherIcon
	}{
		{"clear by day", 0, &day, domain.IconClear},
		{"clear at night", 0, &night, domain.IconClearNight},
		{"mostly clear at night", 1, &night, domain.IconClearNight},
		{"partly cloudy by day", 2, &day, domain.IconPartlyCloud},
		{"partly cloudy at night", 2, &night, domain.IconPartlyCloudNight},
		// Weather that hides the sky looks the same day and night.
		{"rain at night", 61, &night, domain.IconRain},
		{"overcast at night", 3, &night, domain.IconCloudy},
		{"storm at night", 95, &night, domain.IconStorm},
		// A response without the flag is treated as daytime, never as night.
		{"clear, day unknown", 0, nil, domain.IconClear},
	}
	for _, c := range cases {
		if got := currentIcon(c.code, c.isDay); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestFetchAtNightDrawsTheMoonOnlyForTheCurrentConditions(t *testing.T) {
	night := strings.Replace(forecastJSON, `"is_day": 1`, `"is_day": 0`, 1)
	client, req := serve(t, 200, night)
	w, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.Icon != domain.IconClearNight {
		t.Fatalf("current icon %q, want %q", w.Icon, domain.IconClearNight)
	}
	for i, day := range w.Daily {
		if day.Icon == domain.IconClearNight || day.Icon == domain.IconPartlyCloudNight {
			t.Fatalf("forecast day %d shows a night icon %q", i, day.Icon)
		}
	}
	if !strings.Contains(req.URL.Query().Get("current"), "is_day") {
		t.Fatalf("the request must ask for is_day: %v", req.URL.Query())
	}
}

func TestIconPicksAPictureForEveryWeatherCode(t *testing.T) {
	cases := map[domain.WeatherIcon][]int{
		domain.IconClear:       {0, 1},
		domain.IconPartlyCloud: {2},
		domain.IconCloudy:      {3},
		domain.IconFog:         {45, 48},
		domain.IconRain:        {51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82},
		domain.IconSnow:        {71, 73, 75, 77, 85, 86},
		domain.IconStorm:       {95, 96, 99},
	}
	for want, codes := range cases {
		for _, code := range codes {
			if got := icon(code); got != want {
				t.Errorf("code %d: got %q, want %q", code, got, want)
			}
		}
	}
	// A code Open-Meteo adds later should still get a sensible picture.
	if got := icon(1234); got != domain.IconCloudy {
		t.Errorf("unknown code: got %q, want a plain cloud", got)
	}
	// Every code with a name in condition() also has an explicit icon, not the fallback by accident.
	for code := 0; code <= 99; code++ {
		if condition(code) != "Unknown" && code != 3 && icon(code) == domain.IconCloudy {
			t.Errorf("code %d (%s) fell through to the cloud fallback", code, condition(code))
		}
	}
}

func TestFetchAsksForTheConfiguredLocationInFahrenheit(t *testing.T) {
	client, req := serve(t, 200, forecastJSON)
	if _, err := client.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	if q.Get("latitude") != "32.7767" || q.Get("longitude") != "-96.797" || q.Get("temperature_unit") != "fahrenheit" {
		t.Fatalf("query: %v", q)
	}
	if q.Get("forecast_days") != "5" {
		t.Fatalf("forecast_days: %q, want 5", q.Get("forecast_days"))
	}
}

func TestFetchRejectsBadResponses(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
	}{
		"server error": {500, `{}`},
		"not json":     {200, `<html>`},
		"empty":        {200, `{}`},
		"ragged daily": {200, `{"utc_offset_seconds":0,"hourly":{"temperature_2m":[1],"weather_code":[0]},
			"daily":{"time":["2026-11-11","2026-11-12"],"weather_code":[0],"temperature_2m_max":[1],"temperature_2m_min":[1],"precipitation_probability_max":[1]}}`},
		"bad date": {200, `{"utc_offset_seconds":0,"hourly":{"temperature_2m":[1],"weather_code":[0]},
			"daily":{"time":["tomorrow"],"weather_code":[0],"temperature_2m_max":[1],"temperature_2m_min":[1],"precipitation_probability_max":[1]}}`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			client, _ := serve(t, c.status, c.body)
			if _, err := client.Fetch(context.Background()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestTomorrowTrend(t *testing.T) {
	cases := []struct {
		today, next int
		want        string
	}{
		{70, 74, "Warmer. Rain with a high of 74°."},
		{70, 66, "Cooler. Rain with a high of 66°."},
		{70, 73, "About the same. Rain with a high of 73°."},
	}
	for _, c := range cases {
		r := response{UTCOffsetSeconds: 0}
		r.Hourly.Temperature, r.Hourly.Code = []float64{1}, []int{0}
		r.Daily.Time = []string{"2026-11-11", "2026-11-12"}
		r.Daily.Code = []int{0, 61}
		r.Daily.High = []float64{float64(c.today), float64(c.next)}
		r.Daily.Low = []float64{50, 50}
		r.Daily.Precipitation = []float64{0, 0}
		w, err := normalize(r, "Dallas, TX")
		if err != nil {
			t.Fatal(err)
		}
		if w.Tomorrow != c.want {
			t.Fatalf("today %d next %d: %q, want %q", c.today, c.next, w.Tomorrow, c.want)
		}
	}
}
