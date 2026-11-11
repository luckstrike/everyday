package weather

import (
	"context"
	"log"
	"sync"
	"time"

	"everyday/backend/internal/domain"
)

const (
	// RefreshInterval is how often a good forecast is refreshed.
	RefreshInterval = 15 * time.Minute
	// RetryInterval is how soon to try again after a failed refresh.
	RetryInterval = time.Minute
	// MaxAge is how old a forecast may get before it is no longer shown.
	// Showing hours-old weather as current would be worse than showing none.
	MaxAge = time.Hour

	fetchTimeout = 15 * time.Second
)

// Fetcher gets a fresh forecast. Client is the real one; tests use fakes.
type Fetcher interface {
	Fetch(ctx context.Context) (domain.Weather, error)
}

// Service caches the latest forecast. The frontend polls every few seconds, but
// the upstream service is only contacted on the refresh schedule.
type Service struct {
	fetcher Fetcher

	mu        sync.RWMutex
	last      domain.Weather
	fetchedAt time.Time
}

func NewService(fetcher Fetcher) *Service {
	return &Service{fetcher: fetcher}
}

// Run refreshes the forecast until ctx is cancelled. A failed refresh keeps the
// previous forecast and is retried sooner.
func (s *Service) Run(ctx context.Context) {
	for {
		wait := RefreshInterval
		if err := s.refresh(ctx, time.Now()); err != nil {
			log.Printf("weather refresh failed: %v", err)
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
	weather, err := s.fetcher.Fetch(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.last, s.fetchedAt = weather, now
	s.mu.Unlock()
	return nil
}

// Current returns the cached forecast, or false when there is none or it is too old.
func (s *Service) Current(now time.Time) (domain.Weather, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.fetchedAt.IsZero() || now.Sub(s.fetchedAt) > MaxAge {
		return domain.Weather{}, false
	}
	return s.last, true
}
