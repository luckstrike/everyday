package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// load runs Load with no default config file, so a developer's real config can't leak into tests.
func load(vars map[string]string) (Config, error) { return Load(env(vars), "") }

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDefaultsToLiveDataInDallas(t *testing.T) {
	cfg, err := load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mock {
		t.Fatal("mock data must be opt-in")
	}
	if cfg.Location.Name != "Dallas, TX" || cfg.Location.Latitude != 32.7767 || cfg.Location.Longitude != -96.797 {
		t.Fatalf("location: %+v", cfg.Location)
	}
	if len(cfg.GoogleCalendars) != 0 || len(cfg.GoogleExclude) != 0 {
		t.Fatalf("by default every ticked calendar is mirrored: %+v", cfg)
	}
}

func TestMockIsOptIn(t *testing.T) {
	for value, want := range map[string]bool{"1": true, "true": true, "0": false, "false": false} {
		cfg, err := load(map[string]string{"EVERYDAY_MOCK": value})
		if err != nil || cfg.Mock != want {
			t.Fatalf("EVERYDAY_MOCK=%q: mock %v, err %v", value, cfg.Mock, err)
		}
	}
}

func TestCustomLocation(t *testing.T) {
	cfg, err := load(map[string]string{
		"EVERYDAY_LOCATION": "Reykjavik, IS", "EVERYDAY_LAT": "64.1466", "EVERYDAY_LON": "-21.9426",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location.Name != "Reykjavik, IS" || cfg.Location.Latitude != 64.1466 || cfg.Location.Longitude != -21.9426 {
		t.Fatalf("location: %+v", cfg.Location)
	}
}

func TestGoogleCalendarEnvironmentSettings(t *testing.T) {
	cfg, err := load(map[string]string{
		"EVERYDAY_GOOGLE_CALENDARS":   " primary, family@example.com ,,",
		"EVERYDAY_GOOGLE_CREDENTIALS": "/tmp/google.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.GoogleCalendars) != 2 || cfg.GoogleCalendars[1] != "family@example.com" || cfg.GoogleCredentials != "/tmp/google.json" {
		t.Fatalf("custom: %+v", cfg)
	}
}

func TestBillsFileComesFromTheEnvironment(t *testing.T) {
	cfg, err := load(nil)
	if err != nil || cfg.BillsFile != "" {
		t.Fatalf("default: %+v, %v", cfg, err)
	}
	cfg, err = load(map[string]string{"EVERYDAY_BILLS": "/tmp/bills.json"})
	if err != nil || cfg.BillsFile != "/tmp/bills.json" {
		t.Fatalf("custom: %+v, %v", cfg, err)
	}
}

func TestConfigFileSetsLocationAndCalendars(t *testing.T) {
	path := writeConfig(t, `{
	  "location": {"name": "Austin, TX", "latitude": 30.2672, "longitude": -97.7431},
	  "google": {"exclude": ["Birthdays", " Tasks "]}
	}`)
	cfg, err := load(map[string]string{"EVERYDAY_CONFIG": path})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location.Name != "Austin, TX" || cfg.Location.Latitude != 30.2672 || cfg.Location.Longitude != -97.7431 {
		t.Fatalf("location: %+v", cfg.Location)
	}
	if len(cfg.GoogleExclude) != 2 || cfg.GoogleExclude[1] != "Tasks" || len(cfg.GoogleCalendars) != 0 {
		t.Fatalf("google: %+v", cfg)
	}

	path = writeConfig(t, `{"google": {"calendars": ["primary", "work@example.com"]}}`)
	cfg, err = load(map[string]string{"EVERYDAY_CONFIG": path})
	if err != nil || len(cfg.GoogleCalendars) != 2 || cfg.Location.Name != "Dallas, TX" {
		t.Fatalf("explicit calendars with the default location: %+v, %v", cfg, err)
	}
}

func TestEnvironmentOverridesTheConfigFile(t *testing.T) {
	path := writeConfig(t, `{
	  "location": {"name": "Austin, TX", "latitude": 30.2672, "longitude": -97.7431},
	  "google": {"calendars": ["from-file"]}
	}`)
	cfg, err := load(map[string]string{
		"EVERYDAY_CONFIG":           path,
		"EVERYDAY_LOCATION":         "Reykjavik, IS",
		"EVERYDAY_LAT":              "64.1466",
		"EVERYDAY_LON":              "-21.9426",
		"EVERYDAY_GOOGLE_CALENDARS": "from-env",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location.Name != "Reykjavik, IS" || len(cfg.GoogleCalendars) != 1 || cfg.GoogleCalendars[0] != "from-env" {
		t.Fatalf("env should win: %+v", cfg)
	}
}

func TestADefaultConfigFileIsOptionalButANamedOneIsNot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.json")
	if _, err := Load(env(nil), missing); err != nil {
		t.Fatalf("a missing default file is fine: %v", err)
	}
	if _, err := Load(env(map[string]string{"EVERYDAY_CONFIG": missing}), ""); err == nil {
		t.Fatal("a missing file named by EVERYDAY_CONFIG should be an error")
	}
	path := writeConfig(t, `{"google": {"exclude": ["Birthdays"]}}`)
	if cfg, err := Load(env(nil), path); err != nil || len(cfg.GoogleExclude) != 1 {
		t.Fatalf("the default file should be read when present: %+v, %v", cfg, err)
	}
}

func TestRejectsBadConfigFiles(t *testing.T) {
	cases := map[string]string{
		"not json":          `{oops`,
		"misspelled key":    `{"google": {"exlcude": ["Birthdays"]}}`,
		"unknown section":   `{"calendar": {}}`,
		"location no name":  `{"location": {"latitude": 1, "longitude": 2}}`,
		"location no lon":   `{"location": {"name": "x", "latitude": 1}}`,
		"latitude too big":  `{"location": {"name": "x", "latitude": 91, "longitude": 0}}`,
		"longitude too big": `{"location": {"name": "x", "latitude": 0, "longitude": -181}}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			path := writeConfig(t, body)
			_, err := load(map[string]string{"EVERYDAY_CONFIG": path})
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), path) {
				t.Fatalf("the error should name the file: %v", err)
			}
		})
	}
}

func TestRejectsBadSettings(t *testing.T) {
	cases := map[string]map[string]string{
		"no calendar ids":   {"EVERYDAY_GOOGLE_CALENDARS": " , "},
		"bad mock":          {"EVERYDAY_MOCK": "maybe"},
		"name only":         {"EVERYDAY_LOCATION": "Austin, TX"},
		"coordinates only":  {"EVERYDAY_LAT": "30.27", "EVERYDAY_LON": "-97.74"},
		"missing longitude": {"EVERYDAY_LOCATION": "Austin, TX", "EVERYDAY_LAT": "30.27"},
		"latitude range":    {"EVERYDAY_LOCATION": "x", "EVERYDAY_LAT": "91", "EVERYDAY_LON": "0"},
		"longitude range":   {"EVERYDAY_LOCATION": "x", "EVERYDAY_LAT": "0", "EVERYDAY_LON": "-181"},
		"not a number":      {"EVERYDAY_LOCATION": "x", "EVERYDAY_LAT": "north", "EVERYDAY_LON": "0"},
	}
	for name, vars := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := load(vars); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
