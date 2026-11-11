package mock

import (
	"testing"
	"time"
)

func TestMockDatesStayStableWhilePolling(t *testing.T) {
	anchor := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	first := Build(anchor, anchor)
	later := Build(anchor, anchor.Add(time.Hour))
	if first.Calendar[0].StartsAt != later.Calendar[0].StartsAt {
		t.Fatal("mock date moved during polling")
	}
}

func TestDailyForecast(t *testing.T) {
	anchor := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	state := Build(anchor, anchor)
	if len(state.Weather.Daily) != 5 {
		t.Fatal("missing forecast")
	}
	for i, day := range state.Weather.Daily {
		if !day.Date.Equal(anchor.AddDate(0, 0, i)) || day.High < day.Low {
			t.Fatal("invalid daily forecast")
		}
	}
}

func TestEveryMockForecastDayHasAnIcon(t *testing.T) {
	anchor := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for i, day := range Build(anchor, anchor).Weather.Daily {
		if day.Icon == "" {
			t.Fatalf("mock day %d (%s) has no icon", i, day.Condition)
		}
	}
}
