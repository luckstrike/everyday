package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestEndpoints(t *testing.T) {
	h := New()
	for _, path := range []string{"/api/state", "/api/config", "/api/widgets"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
}
func TestScenarioEndpointIsGone(t *testing.T) {
	w := httptest.NewRecorder()
	New().ServeHTTP(w, httptest.NewRequest("POST", "/api/scenario", nil))
	if w.Code != 404 && w.Code != 405 {
		t.Fatal(w.Code)
	}
}
