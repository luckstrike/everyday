package gcal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeGoogle stands in for Google's token endpoint and records what it was sent.
type fakeGoogle struct {
	mu       sync.Mutex
	requests []url.Values
	respond  func(form url.Values) (status int, body string)
	server   *httptest.Server
}

func newFakeGoogle(t *testing.T, respond func(form url.Values) (int, string)) *fakeGoogle {
	t.Helper()
	g := &fakeGoogle{respond: respond}
	g.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		g.mu.Lock()
		g.requests = append(g.requests, r.PostForm)
		g.mu.Unlock()
		status, body := g.respond(r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(g.server.Close)
	return g
}

func (g *fakeGoogle) oauth() *OAuth {
	return &OAuth{ClientID: "client-id", ClientSecret: "client-secret", TokenURL: g.server.URL}
}

func (g *fakeGoogle) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.requests)
}

const tokenOK = `{"access_token":"access-1","refresh_token":"refresh-1","expires_in":3600}`

// authorize runs the sign-in flow, with a fake browser that redirects back to the
// local server using the given query (the state is filled in when it is "STATE").
func authorize(t *testing.T, g *fakeGoogle, redirectQuery string) (string, url.Values, error) {
	t.Helper()
	var authParams url.Values
	token, err := g.oauth().Authorize(context.Background(), func(authURL string) {
		u, err := url.Parse(authURL)
		if err != nil {
			t.Error(err)
			return
		}
		authParams = u.Query()
		query := strings.ReplaceAll(redirectQuery, "STATE", authParams.Get("state"))
		res, err := http.Get(authParams.Get("redirect_uri") + "/?" + query)
		if err != nil {
			t.Error(err)
			return
		}
		res.Body.Close()
	})
	return token, authParams, err
}

func TestAuthorizeSignsInWithPKCE(t *testing.T) {
	g := newFakeGoogle(t, func(url.Values) (int, string) { return 200, tokenOK })
	token, auth, err := authorize(t, g, "code=the-code&state=STATE")
	if err != nil || token != "refresh-1" {
		t.Fatalf("token %q, err %v", token, err)
	}
	if auth.Get("scope") != Scope || auth.Get("access_type") != "offline" || auth.Get("client_id") != "client-id" || auth.Get("code_challenge_method") != "S256" {
		t.Fatalf("authorization request: %v", auth)
	}
	if !strings.HasPrefix(auth.Get("redirect_uri"), "http://127.0.0.1:") {
		t.Fatalf("redirect must be loopback: %q", auth.Get("redirect_uri"))
	}
	if g.count() != 1 {
		t.Fatalf("token requests: %d", g.count())
	}
	form := g.requests[0]
	sum := sha256.Sum256([]byte(form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != auth.Get("code_challenge") {
		t.Fatal("the code verifier does not match the challenge that was sent")
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "the-code" ||
		form.Get("client_secret") != "client-secret" || form.Get("redirect_uri") != auth.Get("redirect_uri") {
		t.Fatalf("exchange request: %v", form)
	}
}

func TestAuthorizeRejectsBadRedirects(t *testing.T) {
	for name, query := range map[string]string{
		"wrong state":     "code=the-code&state=forged",
		"user denied":     "error=access_denied&state=STATE",
		"missing code":    "state=STATE",
		"denied no state": "error=access_denied",
	} {
		t.Run(name, func(t *testing.T) {
			g := newFakeGoogle(t, func(url.Values) (int, string) { return 200, tokenOK })
			if _, _, err := authorize(t, g, query); err == nil {
				t.Fatal("expected an error")
			}
			if g.count() != 0 {
				t.Fatal("the code must not be exchanged after a bad redirect")
			}
		})
	}
}

func TestAuthorizeNeedsARefreshToken(t *testing.T) {
	g := newFakeGoogle(t, func(url.Values) (int, string) { return 200, `{"access_token":"a","expires_in":3600}` })
	if _, _, err := authorize(t, g, "code=c&state=STATE"); err == nil || !strings.Contains(err.Error(), "refresh token") {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeStopsWhenTheUserNeverApproves(t *testing.T) {
	g := newFakeGoogle(t, func(url.Values) (int, string) { return 200, tokenOK })
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := g.oauth().Authorize(ctx, func(string) {}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
}

func TestRefreshTokensReuseAnAccessTokenUntilItNearsExpiry(t *testing.T) {
	g := newFakeGoogle(t, func(url.Values) (int, string) { return 200, `{"access_token":"access-1","expires_in":3600}` })
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	tokens := NewRefreshTokens(g.oauth(), "saved-refresh")
	tokens.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if token, err := tokens.Token(context.Background()); err != nil || token != "access-1" {
			t.Fatalf("token %q, err %v", token, err)
		}
	}
	if g.count() != 1 {
		t.Fatalf("%d token requests, want the first token reused", g.count())
	}
	if form := g.requests[0]; form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "saved-refresh" || form.Get("client_id") != "client-id" {
		t.Fatalf("refresh request: %v", form)
	}

	now = now.Add(59 * time.Minute) // within a minute of expiry, so it is replaced
	if _, err := tokens.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if g.count() != 2 {
		t.Fatalf("%d token requests, want a refresh near expiry", g.count())
	}
}

func TestRevokedSignInTellsTheUserHowToFixIt(t *testing.T) {
	g := newFakeGoogle(t, func(url.Values) (int, string) {
		return 400, `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`
	})
	_, err := NewRefreshTokens(g.oauth(), "old").Token(context.Background())
	if !errors.Is(err, ErrInvalidGrant) || !strings.Contains(err.Error(), "make google-auth") {
		t.Fatalf("got %v", err)
	}
}

func TestOtherTokenFailuresAreErrors(t *testing.T) {
	for name, c := range map[string]struct {
		status int
		body   string
	}{
		"oauth error": {400, `{"error":"invalid_client"}`},
		"server down": {503, `unavailable`},
		"no token":    {200, `{"expires_in":3600}`},
	} {
		t.Run(name, func(t *testing.T) {
			g := newFakeGoogle(t, func(url.Values) (int, string) { return c.status, c.body })
			if _, err := NewRefreshTokens(g.oauth(), "r").Token(context.Background()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
