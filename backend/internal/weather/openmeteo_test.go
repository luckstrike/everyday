package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A trimmed real Open-Meteo response for Dallas (utc_offset_seconds is -6h).
const forecastJSON = `{
  "utc_offset_seconds": -21600,
  "current": {"temperature_2m": 69.4, "weather_code": 0},
  "hourly": {
    "temperature_2m": [69.4, 73.2, 75.7, 77.4, 78.5],
    "weather_code": [0, 0, 0, 2, 3]
  },
  "daily": {
    "time": ["2026-11-11", "2026-11-12", "2026-11-13", "2026-11-14"],
    "weather_code": [2, 0, 3, 3],
    "temperature_2m_max": [78.8, 76.2, 72.9, 64.2],
    "temperature_2m_min": [59.5, 51.4, 48.7, 53.4],
    "precipitation_probability_max": [1, 0, 0, 17]
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
	if len(w.Daily) != 4 || len(w.Hourly) != 5 {
		t.Fatalf("got %d days and %d hours", len(w.Daily), len(w.Hourly))
	}
	wantDate := time.Date(2026, 11, 12, 0, 0, 0, 0, time.FixedZone("", -6*3600))
	if !w.Daily[1].Date.Equal(wantDate) {
		t.Fatalf("second day is %v, want local midnight %v", w.Daily[1].Date, wantDate)
	}
	if w.Hourly[0].Label != "Now" || w.Hourly[2].Label != "+2h" || w.Hourly[3].Condition != "Partly cloudy" {
		t.Fatalf("hourly: %+v", w.Hourly)
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
