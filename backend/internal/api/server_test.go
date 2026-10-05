package api

import (
	"apartment-display/backend/internal/domain"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpoints(t *testing.T) {
	h := New(true)
	for _, path := range []string{"/api/state", "/api/config", "/api/widgets"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
	for _, body := range []string{`{"scenario":"severe"}`, `{"scenario":"normal"}`} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/api/scenario", strings.NewReader(body)))
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		var state domain.State
		json.Unmarshal(w.Body.Bytes(), &state)
		if strings.Contains(body, "severe") && (len(state.Events) != 1 || state.Weather.Alert == "") {
			t.Fatal("missing alert")
		}
	}
	for _, body := range []string{`{"scenario":"unknown"}`, `{`, `{"scenario":"normal","extra":1}`, `{"scenario":"normal"} {}`} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/api/scenario", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatalf("expected 400: %s", body)
		}
	}
}
func TestSimulationDisabledOutsideDevelopment(t *testing.T) {
	w := httptest.NewRecorder()
	New(false).ServeHTTP(w, httptest.NewRequest("POST", "/api/scenario", strings.NewReader(`{"scenario":"severe"}`)))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
