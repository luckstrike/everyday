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

	mu         sync.RWMutex
	last       domain.Weather
	fetchedAt  time.Time
	generation uint64
	wake       chan struct{}
	cancel     context.CancelFunc
}

func NewService(fetcher Fetcher) *Service {
	return &Service{fetcher: fetcher, wake: make(chan struct{}, 1)}
}

func (s *Service) SetLocation(location Location) {
	s.replaceFetcher(&Client{Location: location}, location.Name)
}

func (s *Service) replaceFetcher(fetcher Fetcher, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetcher = fetcher
	s.generation++
	s.last = domain.Weather{Location: name, Loading: true}
	s.fetchedAt = time.Time{}
	if s.cancel != nil {
		s.cancel()
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
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
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.wake:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (s *Service) refresh(ctx context.Context, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	s.mu.Lock()
	fetcher, generation := s.fetcher, s.generation
	s.cancel = cancel
	select {
	case <-s.wake:
	default:
	}
	s.mu.Unlock()
	weather, err := fetcher.Fetch(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != s.generation {
		return nil
	}
	s.cancel = nil
	s.last.Loading = false
	if err == nil {
		s.last, s.fetchedAt = weather, now
	}
	return err
}

// Current returns the cached forecast, or false when there is none or it is too old.
func (s *Service) Current(now time.Time) (domain.Weather, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.fetchedAt.IsZero() || now.Sub(s.fetchedAt) > MaxAge {
		return domain.Weather{Location: s.last.Location, Loading: s.last.Loading}, false
	}
	return s.last, true
}
