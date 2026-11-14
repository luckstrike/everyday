package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidLocation = errors.New("choose a location with a name and valid coordinates")

func (location Location) Validate() error {
	if strings.TrimSpace(location.Name) == "" || utf8.RuneCountInString(location.Name) > 200 ||
		math.IsNaN(location.Latitude) || math.IsNaN(location.Longitude) ||
		math.Abs(location.Latitude) > 90 || math.Abs(location.Longitude) > 180 {
		return ErrInvalidLocation
	}
	return nil
}

type Geocoder struct {
	BaseURL string
	HTTP    *http.Client
}

func (g *Geocoder) Search(ctx context.Context, query string) ([]Location, error) {
	base := g.BaseURL
	if base == "" {
		base = "https://geocoding-api.open-meteo.com/v1/search"
	}
	client := g.HTTP
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	params := url.Values{"name": {query}, "count": {"8"}, "language": {"en"}, "format": {"json"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("location search returned %d", response.StatusCode)
	}
	var body struct {
		Results []struct {
			Name      string   `json:"name"`
			Region    string   `json:"admin1"`
			Country   string   `json:"country"`
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return nil, err
	}
	locations := []Location{}
	for _, result := range body.Results {
		if strings.TrimSpace(result.Name) == "" || result.Latitude == nil || result.Longitude == nil {
			continue
		}
		parts := []string{result.Name}
		for _, part := range []string{result.Region, result.Country} {
			if part != "" && part != parts[len(parts)-1] {
				parts = append(parts, part)
			}
		}
		location := Location{Name: strings.Join(parts, ", "), Latitude: *result.Latitude, Longitude: *result.Longitude}
		if location.Validate() == nil {
			locations = append(locations, location)
		}
	}
	return locations, nil
}
