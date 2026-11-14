package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeocoderLooksUpCoordinatesAndDisambiguatesCities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("name") != "Paris, France" || query.Get("count") != "8" || query.Get("language") != "en" {
			t.Errorf("unexpected query: %v", query)
		}
		w.Write([]byte(`{"results":[
		  {"name":"Paris","admin1":"Île-de-France","country":"France","latitude":48.85,"longitude":2.35},
		  {"name":"Paris","admin1":"Texas","country":"United States","latitude":33.66,"longitude":-95.55},
		  {"name":"Missing coordinates"},
		  {"name":"Invalid","latitude":200,"longitude":0},
		  {"name":"Berlin","admin1":"Berlin","country":"Germany","latitude":52.52,"longitude":13.41}
		]}`))
	}))
	defer server.Close()
	locations, err := (&Geocoder{BaseURL: server.URL}).Search(context.Background(), "Paris, France")
	if err != nil || len(locations) != 3 {
		t.Fatalf("%+v, %v", locations, err)
	}
	if locations[0].Name != "Paris, Île-de-France, France" || locations[0].Latitude != 48.85 || locations[1].Name != "Paris, Texas, United States" || locations[2].Name != "Berlin, Germany" {
		t.Fatalf("normalized results: %+v", locations)
	}
}

func TestGeocoderHandlesNoResultsUpstreamErrorsAndCancellation(t *testing.T) {
	for _, body := range []string{`{}`, `{"results":[]}`, `{"results":null}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		locations, err := (&Geocoder{BaseURL: server.URL}).Search(context.Background(), "nothing")
		server.Close()
		if err != nil || locations == nil || len(locations) != 0 {
			t.Fatalf("%+v, %v", locations, err)
		}
	}
	for _, status := range []int{200, 429, 500} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status); w.Write([]byte(`{broken`)) }))
		_, err := (&Geocoder{BaseURL: server.URL}).Search(context.Background(), "Austin")
		server.Close()
		if err == nil {
			t.Fatalf("expected error for status %d", status)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (&Geocoder{}).Search(ctx, "Austin"); err == nil {
		t.Fatal("expected cancellation")
	}
}
