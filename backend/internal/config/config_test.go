package config

import "testing"

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestDefaultsToLiveDataInDallas(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mock {
		t.Fatal("mock data must be opt-in")
	}
	if cfg.Location.Name != "Dallas, TX" || cfg.Location.Latitude != 32.7767 || cfg.Location.Longitude != -96.797 {
		t.Fatalf("location: %+v", cfg.Location)
	}
}

func TestMockIsOptIn(t *testing.T) {
	for value, want := range map[string]bool{"1": true, "true": true, "0": false, "false": false} {
		cfg, err := Load(env(map[string]string{"EVERYDAY_MOCK": value}))
		if err != nil || cfg.Mock != want {
			t.Fatalf("EVERYDAY_MOCK=%q: mock %v, err %v", value, cfg.Mock, err)
		}
	}
}

func TestCustomLocation(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"EVERYDAY_LOCATION": "Reykjavik, IS", "EVERYDAY_LAT": "64.1466", "EVERYDAY_LON": "-21.9426",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location.Name != "Reykjavik, IS" || cfg.Location.Latitude != 64.1466 || cfg.Location.Longitude != -21.9426 {
		t.Fatalf("location: %+v", cfg.Location)
	}
}

func TestRejectsBadSettings(t *testing.T) {
	cases := map[string]map[string]string{
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
			if _, err := Load(env(vars)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
