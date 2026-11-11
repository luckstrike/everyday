package gcal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultTokenURL = "https://oauth2.googleapis.com/token"

	// Scope is the narrowest one that can read events: view events, nothing else.
	Scope = "https://www.googleapis.com/auth/calendar.events.readonly"
)

// ErrInvalidGrant means Google no longer accepts the saved sign-in, for example
// because access was revoked. Signing in again fixes it.
var ErrInvalidGrant = errors.New("google rejected the saved sign-in (invalid_grant)")

// OAuth is a Google OAuth client of the "Desktop app" type.
type OAuth struct {
	ClientID     string
	ClientSecret string
	// AuthURL, TokenURL and HTTP are overridable for tests; zero values use Google.
	AuthURL  string
	TokenURL string
	HTTP     *http.Client
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Authorize signs the user in with Google's installed-app flow: it listens on a
// loopback port, hands open the URL to visit, waits for Google to redirect back
// with a code, and exchanges it for a long-lived refresh token. PKCE and a random
// state protect the code on its way back.
func (o *OAuth) Authorize(ctx context.Context, open func(authURL string)) (refreshToken string, err error) {
	verifier, err := randomString(32)
	if err != nil {
		return "", err
	}
	state, err := randomString(16)
	if err != nil {
		return "", err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	redirect := "http://" + listener.Addr().String()

	type outcome struct {
		code string
		err  error
	}
	done := make(chan outcome, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var out outcome
		switch {
		case q.Get("state") != state:
			out.err = errors.New("sign-in response had the wrong state; try again")
		case q.Get("error") != "":
			out.err = fmt.Errorf("google sign-in failed: %s", q.Get("error"))
		case q.Get("code") == "":
			out.err = errors.New("google sign-in returned no code")
		default:
			out.code = q.Get("code")
		}
		if out.err != nil {
			http.Error(w, out.err.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintln(w, "Signed in. You can close this tab and return to the terminal.")
		}
		select {
		case done <- out:
		default:
		}
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener)
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()

	challenge := sha256.Sum256([]byte(verifier))
	authURL := o.AuthURL
	if authURL == "" {
		authURL = defaultAuthURL
	}
	open(authURL + "?" + url.Values{
		"client_id":             {o.ClientID},
		"redirect_uri":          {redirect},
		"response_type":         {"code"},
		"scope":                 {Scope},
		"state":                 {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
		// offline asks for a refresh token; consent makes Google issue one even on a repeat sign-in.
		"access_type": {"offline"},
		"prompt":      {"consent"},
	}.Encode())

	var out outcome
	select {
	case out = <-done:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if out.err != nil {
		return "", out.err
	}
	token, err := o.tokenRequest(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {out.code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirect},
	})
	if err != nil {
		return "", err
	}
	if token.RefreshToken == "" {
		return "", errors.New("google returned no refresh token; remove this app at myaccount.google.com/permissions and sign in again")
	}
	return token.RefreshToken, nil
}

// tokenRequest posts to Google's token endpoint with the client's identity added.
func (o *OAuth) tokenRequest(ctx context.Context, form url.Values) (tokenResponse, error) {
	form.Set("client_id", o.ClientID)
	form.Set("client_secret", o.ClientSecret)
	endpoint := o.TokenURL
	if endpoint == "" {
		endpoint = defaultTokenURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := httpClient(o.HTTP).Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer res.Body.Close()
	var token tokenResponse
	// Google explains failures in the JSON body, so decode it whatever the status.
	decodeErr := json.NewDecoder(res.Body).Decode(&token)
	switch {
	case token.Error == "invalid_grant":
		return tokenResponse{}, fmt.Errorf("%w: %s", ErrInvalidGrant, token.ErrorDescription)
	case token.Error != "":
		return tokenResponse{}, fmt.Errorf("google token endpoint: %s: %s", token.Error, token.ErrorDescription)
	case res.StatusCode != http.StatusOK:
		return tokenResponse{}, fmt.Errorf("google token endpoint returned %s", res.Status)
	case decodeErr != nil:
		return tokenResponse{}, fmt.Errorf("decode google token response: %w", decodeErr)
	}
	return token, nil
}

// TokenProvider supplies a current access token for API calls.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// RefreshTokens trades a refresh token for short-lived access tokens, reusing
// each one until shortly before it expires.
type RefreshTokens struct {
	oauth        *OAuth
	refreshToken string
	now          func() time.Time

	mu      sync.Mutex
	access  string
	expires time.Time
}

func NewRefreshTokens(o *OAuth, refreshToken string) *RefreshTokens {
	return &RefreshTokens{oauth: o, refreshToken: refreshToken, now: time.Now}
}

func (r *RefreshTokens) Token(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.access != "" && r.now().Before(r.expires.Add(-time.Minute)) {
		return r.access, nil
	}
	token, err := r.oauth.tokenRequest(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {r.refreshToken},
	})
	if err != nil {
		if errors.Is(err, ErrInvalidGrant) {
			return "", fmt.Errorf("%w; run `make google-auth` to sign in again", err)
		}
		return "", err
	}
	if token.AccessToken == "" {
		return "", errors.New("google returned no access token")
	}
	r.access = token.AccessToken
	r.expires = r.now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return r.access, nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}
	return &http.Client{Timeout: 10 * time.Second}
}
