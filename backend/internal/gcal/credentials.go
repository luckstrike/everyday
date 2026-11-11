// Package gcal reads events from Google Calendar. It talks to Google's OAuth and
// Calendar REST endpoints directly, so it needs nothing beyond the standard library.
package gcal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials are everything needed to read a Google account's calendars. They
// are secrets: the file holding them is private to the user and never in the repo.
type Credentials struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	// RefreshToken is empty until the user has signed in once.
	RefreshToken string `json:"refreshToken,omitempty"`
}

// DefaultCredentialsPath is where credentials live unless configured otherwise,
// inside the operating system's per-user config directory.
func DefaultCredentialsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "everyday", "google.json"), nil
}

// LoadCredentials reads the credentials file. A missing file returns an error
// for which errors.Is(err, fs.ErrNotExist) is true.
func LoadCredentials(path string) (Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, err
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return Credentials{}, fmt.Errorf("read %s: %w", path, err)
	}
	return c, nil
}

// SaveCredentials writes the file readable by the current user only. It writes a
// temporary file first so a crash cannot leave a half-written credentials file.
func SaveCredentials(path string, c Credentials) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".google-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// ReadClientFile reads an OAuth client ID and secret from a JSON file. It accepts
// the file Google's console offers to download ({"installed": {...}}) as well as a
// flat {"client_id": ..., "client_secret": ...}, so the secret never has to be typed.
func ReadClientFile(path string) (clientID, clientSecret string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	type client struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	var file struct {
		client
		Installed *client `json:"installed"`
		Web       *client `json:"web"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return "", "", fmt.Errorf("read %s: %w", path, err)
	}
	c := file.client
	if file.Installed != nil {
		c = *file.Installed
	}
	if file.Web != nil && file.Installed == nil {
		// Google ties the client type to the credential, so this cannot be fixed by
		// editing the file: a web client only accepts pre-registered redirect URIs,
		// and sign-in redirects to a loopback port chosen at run time.
		return "", "", fmt.Errorf("%s is a Web application client; create a new OAuth client of type \"Desktop app\" in the Google console and use that one", path)
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		return "", "", fmt.Errorf("%s has no client_id and client_secret", path)
	}
	return c.ClientID, c.ClientSecret, nil
}
