// Package config reads the backend's settings from environment variables.
package config

import (
	"fmt"
	"strconv"

	"everyday/backend/internal/weather"
)

// Config is everything the server needs to start.
type Config struct {
	// Mock serves made-up development data instead of live sources.
	Mock     bool
	Location weather.Location
}

// Load reads the settings through getenv (os.Getenv in production).
//
//	EVERYDAY_MOCK       "1"/"true" serves mock data instead of live sources
//	EVERYDAY_LOCATION   display name, e.g. "Austin, TX"
//	EVERYDAY_LAT        latitude in degrees
//	EVERYDAY_LON        longitude in degrees
//
// The three location variables are set together or not at all; with none set
// the location is Dallas, TX.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{Location: weather.Location{Name: "Dallas, TX", Latitude: 32.7767, Longitude: -96.797}}

	if v := getenv("EVERYDAY_MOCK"); v != "" {
		mock, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("EVERYDAY_MOCK must be true or false, got %q", v)
		}
		cfg.Mock = mock
	}

	name, lat, lon := getenv("EVERYDAY_LOCATION"), getenv("EVERYDAY_LAT"), getenv("EVERYDAY_LON")
	if name == "" && lat == "" && lon == "" {
		return cfg, nil
	}
	if name == "" || lat == "" || lon == "" {
		return Config{}, fmt.Errorf("set EVERYDAY_LOCATION, EVERYDAY_LAT and EVERYDAY_LON together")
	}
	latitude, err := parseDegrees("EVERYDAY_LAT", lat, 90)
	if err != nil {
		return Config{}, err
	}
	longitude, err := parseDegrees("EVERYDAY_LON", lon, 180)
	if err != nil {
		return Config{}, err
	}
	cfg.Location = weather.Location{Name: name, Latitude: latitude, Longitude: longitude}
	return cfg, nil
}

func parseDegrees(name, value string, limit float64) (float64, error) {
	degrees, err := strconv.ParseFloat(value, 64)
	if err != nil || degrees < -limit || degrees > limit {
		return 0, fmt.Errorf("%s must be a number between -%g and %g, got %q", name, limit, limit, value)
	}
	return degrees, nil
}
