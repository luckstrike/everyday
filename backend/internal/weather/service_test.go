package weather

import (
	"context"
	"errors"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

type fakeFetcher struct {
	weather domain.Weather
	err     error
	calls   int
}

func (f *fakeFetcher) Fetch(context.Context) (domain.Weather, error) {
	f.calls++
	return f.weather, f.err
}

var start = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func TestNothingIsShownBeforeTheFirstSuccessfulFetch(t *testing.T) {
	s := NewService(&fakeFetcher{err: errors.New("offline")})
	if err := s.refresh(context.Background(), start); err == nil {
		t.Fatal("expected the fetch error")
	}
	if _, ok := s.Current(start); ok {
		t.Fatal("no forecast should be available")
	}
}

func TestFailedRefreshKeepsTheLastGoodForecast(t *testing.T) {
	f := &fakeFetcher{weather: domain.Weather{Temperature: 70}}
	s := NewService(f)
	if err := s.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	f.weather, f.err = domain.Weather{Temperature: 99}, errors.New("offline")
	if err := s.refresh(context.Background(), start.Add(RefreshInterval)); err == nil {
		t.Fatal("expected the fetch error")
	}
	w, ok := s.Current(start.Add(RefreshInterval))
	if !ok || w.Temperature != 70 {
		t.Fatalf("got %+v, %v; want the earlier forecast", w, ok)
	}
}

func TestForecastExpiresAfterMaxAge(t *testing.T) {
	s := NewService(&fakeFetcher{weather: domain.Weather{Temperature: 70}})
	if err := s.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Current(start.Add(MaxAge)); !ok {
		t.Fatal("a forecast exactly MaxAge old should still show")
	}
	if _, ok := s.Current(start.Add(MaxAge + time.Second)); ok {
		t.Fatal("a forecast older than MaxAge should not show")
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	f := &fakeFetcher{weather: domain.Weather{Temperature: 70}}
	s := NewService(f)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if _, ok := s.Current(time.Now()); ok {
			break
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop")
	}
	if _, ok := s.Current(time.Now()); !ok {
		t.Fatal("Run should have fetched immediately")
	}
}
