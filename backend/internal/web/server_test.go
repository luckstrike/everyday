package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"everyday/backend/internal/api"
	"everyday/backend/internal/config"
	"everyday/backend/internal/weather"
)

func frontendFiles() fs.FS {
	return fstest.MapFS{
		"index.html":         {Data: []byte(`<!doctype html><html><script type="module" src="/assets/app-123.js"></script></html>`)},
		"assets/app-123.js":  {Data: []byte(`console.log("dashboard")`)},
		"assets/app-123.css": {Data: []byte(`body { color: green; }`)},
		".env":               {Data: []byte("not public")},
		"assets/.hidden":     {Data: []byte("not public")},
		"api/state":          {Data: []byte("must not shadow the API")},
	}
}

func TestBuiltDashboardAndAssetsAreServedAlongsideTheAPI(t *testing.T) {
	handler, err := New(api.New(api.Options{Mock: true}), frontendFiles())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ path, contentType, text, cache string }{
		{"/", "text/html", "<!doctype html>", "no-store"},
		{"/assets/app-123.js", "javascript", `console.log("dashboard")`, "no-cache"},
		{"/assets/app-123.css", "text/css", "color: green", "no-cache"},
		{"/api/state", "application/json", `"generatedAt"`, "no-store"},
		{"/api/config", "application/json", `"pollIntervalMs"`, "no-store"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "http://localhost:8080"+entry.path, nil))
		if response.Code != 200 || !strings.Contains(response.Body.String(), entry.text) || !strings.Contains(response.Header().Get("Content-Type"), entry.contentType) || response.Header().Get("Cache-Control") != entry.cache {
			t.Errorf("%s: %d %v %s", entry.path, response.Code, response.Header(), response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("HEAD", "http://localhost:8080/assets/app-123.js", nil))
	if response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") == "" {
		t.Fatalf("HEAD: %d %v %s", response.Code, response.Header(), response.Body.String())
	}
}

func TestMissingFilesDirectoriesAndUnknownAPIRoutesNeverReturnTheApp(t *testing.T) {
	handler, err := New(api.New(api.Options{Mock: true}), frontendFiles())
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/missing.js", "/assets", "/assets/", "/.env", "/assets/.hidden", "/api", "/api/missing", "/settings/location"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "http://localhost:8080"+path, nil))
		if response.Code != 404 || strings.Contains(response.Body.String(), "<!doctype html>") || strings.Contains(response.Body.String(), "not public") {
			t.Errorf("%s: %d %s", path, response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("POST", "http://localhost:8080/", strings.NewReader("unexpected")))
	if response.Code != 405 || response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("POST: %d %v", response.Code, response.Header())
	}
}

func TestMissingOrInvalidFrontendFailsBeforeServing(t *testing.T) {
	for _, frontend := range []fs.FS{
		fstest.MapFS{},
		fstest.MapFS{"index.html/child": {Data: []byte("directory, not an HTML file")}},
	} {
		if _, err := New(http.NotFoundHandler(), frontend); err == nil {
			t.Fatal("expected a missing/invalid build error")
		}
	}
}

func TestSingleServerKeepsSameOriginLocationEditingAndSecurityChecks(t *testing.T) {
	initial := weather.Location{Name: "Dallas", Latitude: 32.7767, Longitude: -96.797}
	store := config.NewLocationStore(func(string) string { return "" }, filepath.Join(t.TempDir(), "config.json"), initial)
	service := weather.NewService(nil)
	service.SetLocation(initial)
	handler, err := New(api.New(api.Options{Location: initial.Name, Weather: service, LocationStore: store, OnLocationChanged: service.SetLocation}), frontendFiles())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct {
		host, origin, contentType string
		status                    int
	}{
		{"127.0.0.1:8080", "http://127.0.0.1:8080", "application/json", 200},
		{"localhost:8080", "http://evil.example", "application/json", 403},
		{"evil.example:8080", "http://evil.example:8080", "application/json", 403},
		{"localhost:8080", "http://localhost:8080", "text/plain", 415},
	} {
		request := httptest.NewRequest("PUT", "http://"+entry.host+"/api/location", strings.NewReader(`{"name":"Austin","latitude":30.27,"longitude":-97.74}`))
		request.Header.Set("Origin", entry.origin)
		request.Header.Set("Content-Type", entry.contentType)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != entry.status {
			t.Errorf("%+v: %d %s", entry, response.Code, response.Body.String())
		}
	}
	current, _ := store.Location()
	if current.Name != "Austin" {
		t.Fatalf("saved location: %+v", current)
	}
}
