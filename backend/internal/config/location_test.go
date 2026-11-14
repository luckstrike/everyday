package config

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"everyday/backend/internal/weather"
)

var testLocation = weather.Location{Name: "Austin, Texas, United States", Latitude: 30.2672, Longitude: -97.7431}

func TestLocationSavePreservesCalendarSettingsAndSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"google":{"calendars":["home"],"exclude":["work"]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	getenv := func(string) string { return "" }
	store := NewLocationStore(getenv, path, weather.Location{Name: "Dallas"})
	if err := store.Save(testLocation); err != nil {
		t.Fatal(err)
	}
	current, reason := store.Location()
	if current != testLocation || reason != "" {
		t.Fatalf("saved location: %+v, %q", current, reason)
	}
	cfg, err := Load(getenv, path)
	if err != nil || cfg.Location != testLocation || len(cfg.GoogleCalendars) != 1 || cfg.GoogleCalendars[0] != "home" || len(cfg.GoogleExclude) != 1 || cfg.GoogleExclude[0] != "work" {
		t.Fatalf("reloaded config: %+v, %v", cfg, err)
	}
}

func TestLocationSaveCreatesDefaultConfigAndUsesExplicitConfigPath(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "nested", "config.json")
		getenv := func(key string) string {
			if explicit && key == "EVERYDAY_CONFIG" {
				return path
			}
			return ""
		}
		fallback := path
		if explicit {
			fallback = filepath.Join(t.TempDir(), "unused.json")
		}
		store := NewLocationStore(getenv, fallback, weather.Location{})
		if err := store.Save(testLocation); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLocationFailedSaveLeavesFileAndActiveLocationAlone(t *testing.T) {
	for _, contents := range []string{`{broken`, `null`, `[]`} {
		path := filepath.Join(t.TempDir(), "config.json")
		os.WriteFile(path, []byte(contents), 0o600)
		initial := weather.Location{Name: "Dallas"}
		store := NewLocationStore(func(string) string { return "" }, path, initial)
		if err := store.Save(testLocation); err == nil {
			t.Fatal("expected invalid file error")
		}
		current, _ := store.Location()
		data, _ := os.ReadFile(path)
		if current != initial || string(data) != contents {
			t.Fatal("a failed save changed settings")
		}
	}
	blocker := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocker, nil, 0o600)
	store := NewLocationStore(func(string) string { return "" }, filepath.Join(blocker, "config.json"), weather.Location{Name: "Dallas"})
	if err := store.Save(testLocation); err == nil {
		t.Fatal("expected write failure")
	}
	current, _ := store.Location()
	if current.Name != "Dallas" {
		t.Fatal("failed save changed active location")
	}
}

func TestLocationStoreRefusesInvalidAndEnvironmentLockedLocations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	getenv := func(key string) string {
		if key == "EVERYDAY_LOCATION" {
			return "Dallas"
		}
		return ""
	}
	store := NewLocationStore(getenv, path, weather.Location{Name: "Dallas"})
	if _, reason := store.Location(); reason == "" {
		t.Fatal("environment override should be explained")
	}
	if err := store.Save(testLocation); err == nil {
		t.Fatal("environment override should not be editable")
	}
	store = NewLocationStore(func(string) string { return "" }, path, weather.Location{})
	for _, invalid := range []weather.Location{
		{Name: " "}, {Name: "Bad", Latitude: 91}, {Name: "Bad", Longitude: -181}, {Name: "Bad", Latitude: math.NaN()}, {Name: "Bad", Longitude: math.Inf(1)},
	} {
		if err := store.Save(invalid); err == nil {
			t.Fatalf("accepted %+v", invalid)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("invalid settings should not create the file")
	}
}
