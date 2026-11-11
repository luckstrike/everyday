// Package config reads the backend's settings from a JSON file and environment variables.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"everyday/backend/internal/weather"
)

// Config is everything the server needs to start.
type Config struct {
	// Mock serves made-up development data instead of live sources.
	Mock     bool
	Location weather.Location
	// GoogleCredentials is the credentials file path; empty means the default location.
	GoogleCredentials string
	// BillsFile is the bills file path; empty means the default location.
	BillsFile string
	// GoogleCalendars, when set, are the only Google calendars shown. When empty the
	// dashboard mirrors the calendars ticked in Google Calendar.
	GoogleCalendars []string
	// GoogleExclude names calendars, by ID or name, to leave out of the mirrored set.
	GoogleExclude []string
}

// file is the shape of the JSON config file. It holds no secrets, so it is safe to
// back up or share. Every field is optional.
type file struct {
	Location *struct {
		Name      string   `json:"name"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	} `json:"location"`
	Google struct {
		Calendars []string `json:"calendars"`
		Exclude   []string `json:"exclude"`
	} `json:"google"`
}

// DefaultFile is where the config file lives unless EVERYDAY_CONFIG says otherwise:
// everyday/config.json in the user's config directory. It is empty if that directory
// cannot be found.
func DefaultFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "everyday", "config.json")
}

// Load builds the configuration. Settings come from, in increasing priority,
// built-in defaults, the config file, and environment variables (read through
// getenv, which is os.Getenv in production). defaultFile is used when
// EVERYDAY_CONFIG is not set; a missing default file is fine, but a missing file
// named by EVERYDAY_CONFIG is an error.
//
//	EVERYDAY_CONFIG     path of the JSON config file
//	EVERYDAY_MOCK       "1"/"true" serves mock data instead of live sources
//	EVERYDAY_LOCATION   display name, e.g. "Austin, TX"
//	EVERYDAY_LAT        latitude in degrees
//	EVERYDAY_LON        longitude in degrees
//	EVERYDAY_GOOGLE_CREDENTIALS  path of the Google credentials file
//	EVERYDAY_BILLS      path of the bills file
//	EVERYDAY_GOOGLE_CALENDARS    comma-separated Google calendar IDs; replaces mirroring
//
// The three location variables are set together or not at all; with no location
// anywhere the location is Dallas, TX.
func Load(getenv func(string) string, defaultFile string) (Config, error) {
	cfg := Config{
		Location:          weather.Location{Name: "Dallas, TX", Latitude: 32.7767, Longitude: -96.797},
		GoogleCredentials: getenv("EVERYDAY_GOOGLE_CREDENTIALS"),
		BillsFile:         getenv("EVERYDAY_BILLS"),
	}
	if err := applyFile(&cfg, getenv("EVERYDAY_CONFIG"), defaultFile); err != nil {
		return Config{}, err
	}
	if err := applyEnv(&cfg, getenv); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyFile(cfg *Config, explicit, fallback string) error {
	path := explicit
	if path == "" {
		path = fallback
	}
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) && explicit == "" {
		return nil
	}
	if err != nil {
		return err
	}
	var f file
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields() // a misspelled setting should be an error, not silently ignored
	if err := decoder.Decode(&f); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if loc := f.Location; loc != nil {
		if loc.Name == "" || loc.Latitude == nil || loc.Longitude == nil {
			return fmt.Errorf("%s: location needs a name, latitude and longitude", path)
		}
		if !within(*loc.Latitude, 90) || !within(*loc.Longitude, 180) {
			return fmt.Errorf("%s: location latitude must be within ±90 and longitude within ±180", path)
		}
		cfg.Location = weather.Location{Name: loc.Name, Latitude: *loc.Latitude, Longitude: *loc.Longitude}
	}
	cfg.GoogleCalendars = trimmed(f.Google.Calendars)
	cfg.GoogleExclude = trimmed(f.Google.Exclude)
	return nil
}

func applyEnv(cfg *Config, getenv func(string) string) error {
	if v := getenv("EVERYDAY_GOOGLE_CALENDARS"); v != "" {
		cfg.GoogleCalendars = trimmed(strings.Split(v, ","))
		if len(cfg.GoogleCalendars) == 0 {
			return fmt.Errorf("EVERYDAY_GOOGLE_CALENDARS has no calendar IDs, got %q", v)
		}
	}

	if v := getenv("EVERYDAY_MOCK"); v != "" {
		mock, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("EVERYDAY_MOCK must be true or false, got %q", v)
		}
		cfg.Mock = mock
	}

	name, lat, lon := getenv("EVERYDAY_LOCATION"), getenv("EVERYDAY_LAT"), getenv("EVERYDAY_LON")
	if name == "" && lat == "" && lon == "" {
		return nil
	}
	if name == "" || lat == "" || lon == "" {
		return fmt.Errorf("set EVERYDAY_LOCATION, EVERYDAY_LAT and EVERYDAY_LON together")
	}
	latitude, err := parseDegrees("EVERYDAY_LAT", lat, 90)
	if err != nil {
		return err
	}
	longitude, err := parseDegrees("EVERYDAY_LON", lon, 180)
	if err != nil {
		return err
	}
	cfg.Location = weather.Location{Name: name, Latitude: latitude, Longitude: longitude}
	return nil
}

func parseDegrees(name, value string, limit float64) (float64, error) {
	degrees, err := strconv.ParseFloat(value, 64)
	if err != nil || !within(degrees, limit) {
		return 0, fmt.Errorf("%s must be a number between -%g and %g, got %q", name, limit, limit, value)
	}
	return degrees, nil
}

func within(v, limit float64) bool { return !math.IsNaN(v) && v >= -limit && v <= limit }

// trimmed drops blank entries and surrounding spaces.
func trimmed(in []string) []string {
	var out []string
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
