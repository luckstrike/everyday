package gcal

import (
	"context"
	"log"
	"sync"
	"time"

	"everyday/backend/internal/domain"
)

const (
	// RefreshInterval is how often events are refreshed after a success.
	RefreshInterval = 5 * time.Minute
	// RetryInterval is how soon to try again after a failed refresh.
	RetryInterval = time.Minute
	// MaxAge is how old the events may get before they are no longer shown.
	MaxAge = 30 * time.Minute

	fetchTimeout = 15 * time.Second
)

// Fetcher gets fresh events. Client is the real one; tests use fakes.
type Fetcher interface {
	Fetch(ctx context.Context) ([]domain.CalendarEvent, error)
}

// Service caches the latest events so the dashboard's frequent polling never
// reaches Google. It mirrors weather.Service; if a third source needs the same
// refresh-and-cache loop, that is the time to share it.
type Service struct {
	fetcher Fetcher

	mu        sync.RWMutex
	last      []domain.CalendarEvent
	fetchedAt time.Time
}

func NewService(fetcher Fetcher) *Service {
	return &Service{fetcher: fetcher}
}

// Run refreshes the events until ctx is cancelled. A failed refresh keeps the
// previous events and is retried sooner.
func (s *Service) Run(ctx context.Context) {
	for {
		wait := RefreshInterval
		if err := s.refresh(ctx, time.Now()); err != nil {
			log.Printf("calendar refresh failed: %v", err)
			wait = RetryInterval
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func (s *Service) refresh(ctx context.Context, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	events, err := s.fetcher.Fetch(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.last, s.fetchedAt = events, now
	s.mu.Unlock()
	return nil
}

// Current returns the cached events, or false when there are none yet or they
// are too old. An empty list with true means the calendar really is empty.
func (s *Service) Current(now time.Time) ([]domain.CalendarEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.fetchedAt.IsZero() || now.Sub(s.fetchedAt) > MaxAge {
		return nil, false
	}
	return s.last, true
}
