package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"sync"

	"everyday/backend/internal/fileutil"
	"everyday/backend/internal/weather"
)

type LocationStore struct {
	mu       sync.Mutex
	path     string
	location weather.Location
	reason   string
}

func NewLocationStore(getenv func(string) string, defaultFile string, location weather.Location) *LocationStore {
	path := getenv("EVERYDAY_CONFIG")
	if path == "" {
		path = defaultFile
	}
	store := &LocationStore{path: path, location: location}
	if getenv("EVERYDAY_LOCATION") != "" || getenv("EVERYDAY_LAT") != "" || getenv("EVERYDAY_LON") != "" {
		store.reason = "Location is set by environment variables. Remove EVERYDAY_LOCATION, EVERYDAY_LAT and EVERYDAY_LON, then restart to edit it here."
	} else if path == "" {
		store.reason = "A local config file is needed to save your location."
	}
	return store
}

func (s *LocationStore) Location() (weather.Location, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.location, s.reason
}

func (s *LocationStore) Save(location weather.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reason != "" {
		return errors.New(s.reason)
	}
	location.Name = strings.TrimSpace(location.Name)
	if err := location.Validate(); err != nil {
		return err
	}
	body := map[string]json.RawMessage{}
	data, err := os.ReadFile(s.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err == nil {
		if err := json.Unmarshal(data, &body); err != nil {
			return err
		}
		if body == nil {
			return errors.New("config file must contain an object")
		}
	}
	body["location"], err = json.Marshal(location)
	if err != nil {
		return err
	}
	data, err = json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	if err := fileutil.WriteAtomic(s.path, append(data, '\n'), 0o600); err != nil {
		return err
	}
	s.location = location
	return nil
}
