package gcal

import (
	"context"
	"errors"
	"testing"
	"time"

	"everyday/backend/internal/domain"
)

type fakeFetcher struct {
	events []domain.CalendarEvent
	err    error
}

func (f *fakeFetcher) Fetch(context.Context) ([]domain.CalendarEvent, error) { return f.events, f.err }

var start = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func TestNothingIsShownBeforeTheFirstSuccessfulFetch(t *testing.T) {
	s := NewService(&fakeFetcher{err: errors.New("offline")})
	if err := s.refresh(context.Background(), start); err == nil {
		t.Fatal("expected the fetch error")
	}
	if _, ok := s.Current(start); ok {
		t.Fatal("no events should be available")
	}
}

func TestAnEmptyCalendarIsStillLive(t *testing.T) {
	s := NewService(&fakeFetcher{events: []domain.CalendarEvent{}})
	if err := s.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	if events, ok := s.Current(start); !ok || len(events) != 0 {
		t.Fatalf("got %v, %v; want an empty but live calendar", events, ok)
	}
}

func TestFailedRefreshKeepsTheLastGoodEvents(t *testing.T) {
	f := &fakeFetcher{events: []domain.CalendarEvent{{Title: "Dinner"}}}
	s := NewService(f)
	if err := s.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	f.events, f.err = nil, errors.New("offline")
	if err := s.refresh(context.Background(), start.Add(RefreshInterval)); err == nil {
		t.Fatal("expected the fetch error")
	}
	events, ok := s.Current(start.Add(RefreshInterval))
	if !ok || len(events) != 1 || events[0].Title != "Dinner" {
		t.Fatalf("got %v, %v; want the earlier events", events, ok)
	}
}

func TestEventsExpireAfterMaxAge(t *testing.T) {
	s := NewService(&fakeFetcher{events: []domain.CalendarEvent{{Title: "Dinner"}}})
	if err := s.refresh(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Current(start.Add(MaxAge)); !ok {
		t.Fatal("events exactly MaxAge old should still show")
	}
	if _, ok := s.Current(start.Add(MaxAge + time.Second)); ok {
		t.Fatal("events older than MaxAge should not show")
	}
}
