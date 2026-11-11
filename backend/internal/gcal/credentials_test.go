package gcal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialsRoundTripAndArePrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "google.json")
	want := Credentials{ClientID: "id", ClientSecret: "secret", RefreshToken: "refresh"}
	if err := SaveCredentials(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCredentials(path)
	if err != nil || got != want {
		t.Fatalf("got %+v, %v", got, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("credentials file mode is %v, want 0600", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestLoadCredentialsReportsMissingAndBrokenFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadCredentials(filepath.Join(dir, "none.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	broken := filepath.Join(dir, "broken.json")
	os.WriteFile(broken, []byte("{not json"), 0o600)
	if _, err := LoadCredentials(broken); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("broken file: %v", err)
	}
}

func TestReadClientFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		os.WriteFile(path, []byte(body), 0o600)
		return path
	}
	for name, body := range map[string]string{
		"google-download": `{"installed":{"client_id":"id","client_secret":"secret","auth_uri":"x"}}`,
		"flat":            `{"client_id":"id","client_secret":"secret"}`,
	} {
		id, secret, err := ReadClientFile(write(name, body))
		if err != nil || id != "id" || secret != "secret" {
			t.Fatalf("%s: %q %q %v", name, id, secret, err)
		}
	}
	if _, _, err := ReadClientFile(write("empty", `{"installed":{"client_id":"id"}}`)); err == nil {
		t.Fatal("a file without a secret should fail")
	}
	// A Web application client can't complete the loopback sign-in, so say so plainly.
	_, _, err := ReadClientFile(write("web", `{"web":{"client_id":"id","client_secret":"secret"}}`))
	if err == nil || !strings.Contains(err.Error(), "Desktop app") {
		t.Fatalf("web client: %v", err)
	}
}
