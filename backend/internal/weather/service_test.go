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

type blockedFetcher struct {
	started chan struct{}
	release chan struct{}
}

func (f *blockedFetcher) Fetch(context.Context) (domain.Weather, error) {
	close(f.started)
	<-f.release
	return domain.Weather{Location: "Dallas", Temperature: 70}, nil
}

func TestChangingLocationClearsForecastAndRejectsAnOldInflightResult(t *testing.T) {
	old := &blockedFetcher{started: make(chan struct{}), release: make(chan struct{})}
	service := NewService(old)
	done := make(chan struct{})
	go func() { service.refresh(context.Background(), start); close(done) }()
	<-old.started
	service.replaceFetcher(&fakeFetcher{weather: domain.Weather{Location: "Austin", Temperature: 80}}, "Austin")
	current, ok := service.Current(start)
	if ok || current.Location != "Austin" || !current.Loading {
		t.Fatalf("old weather not cleared: %+v, %v", current, ok)
	}
	close(old.release)
	<-done
	current, ok = service.Current(start)
	if ok || current.Location != "Austin" || !current.Loading {
		t.Fatalf("old request replaced the new location: %+v", current)
	}
	if err := service.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	current, ok = service.Current(start)
	if !ok || current.Location != "Austin" || current.Temperature != 80 || current.Loading {
		t.Fatalf("new weather: %+v, %v", current, ok)
	}
}

func TestLocationChangeWakesRefreshLoopAndFailureDoesNotRestoreOldWeather(t *testing.T) {
	service := NewService(&fakeFetcher{weather: domain.Weather{Location: "Dallas"}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { service.Run(ctx); close(done) }()
	waitFor := func(ready func() bool) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if ready() {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("timed out waiting for forecast change")
	}
	waitFor(func() bool { _, ok := service.Current(time.Now()); return ok })
	service.replaceFetcher(&fakeFetcher{err: errors.New("offline")}, "Austin")
	waitFor(func() bool {
		current, ok := service.Current(time.Now())
		return !ok && current.Location == "Austin" && !current.Loading
	})
	service.replaceFetcher(&fakeFetcher{weather: domain.Weather{Location: "Berlin", Temperature: 50}}, "Berlin")
	waitFor(func() bool { current, ok := service.Current(time.Now()); return ok && current.Location == "Berlin" })
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop failed to stop")
	}
}
