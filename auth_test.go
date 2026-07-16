package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestCaptureViaServer drives the loopback callback server: it starts the
// listener, simulates the consent redirect with a GET, and checks the captured
// code. SEATS_MCP_NO_BROWSER keeps the test from opening a real browser.
func TestCaptureViaServer(t *testing.T) {
	t.Setenv("SEATS_MCP_NO_BROWSER", "1")

	// Reserve a free loopback port, then release it for the callback server.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	redirectURI := "http://" + addr + "/callback"

	type result struct {
		code string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		code, err := obtainCode(context.Background(), redirectURI, "http://consent.invalid", "the-state")
		done <- result{code, err}
	}()

	// Poll the callback until the server is up, then deliver the redirect.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case res := <-done:
			t.Fatalf("obtainCode returned early: code=%q err=%v", res.code, res.err)
		default:
		}
		resp, gerr := http.Get("http://" + addr + "/callback?code=the-code&state=the-state")
		if gerr == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("obtainCode: %v", res.err)
		}
		if res.code != "the-code" {
			t.Errorf("code = %q, want the-code", res.code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("obtainCode did not return")
	}
}

// TestCaptureViaServerStateMismatch ensures a forged state is rejected.
func TestCaptureViaServerStateMismatch(t *testing.T) {
	t.Setenv("SEATS_MCP_NO_BROWSER", "1")

	l, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := l.Addr().String()
	l.Close()
	redirectURI := "http://" + addr + "/callback"

	type result struct {
		code string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		code, err := obtainCode(context.Background(), redirectURI, "http://consent.invalid", "expected")
		done <- result{code, err}
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + addr + "/callback?code=c&state=forged")
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case res := <-done:
		if res.err == nil {
			t.Fatal("expected state mismatch error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("obtainCode did not return")
	}
}

func TestParsePastedCode(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		state   string
		want    string
		wantErr bool
	}{
		{"bare code", "abc123", "st", "abc123", false},
		{"full url", "https://app.example/cb?code=xyz&state=st", "st", "xyz", false},
		{"url without state", "https://app.example/cb?code=xyz", "st", "xyz", false},
		{"state mismatch", "https://app.example/cb?code=xyz&state=other", "st", "", true},
		{"url missing code", "https://app.example/cb?state=st", "st", "", true},
		{"empty", "", "st", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parsePastedCode(c.input, c.state)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestIsLoopback(t *testing.T) {
	cases := map[string]bool{
		"localhost":   true,
		"127.0.0.1":   true,
		"::1":         true,
		"0.0.0.0":     false,
		"example.com": false,
		"192.168.1.5": false,
	}
	for host, want := range cases {
		if got := isLoopback(host); got != want {
			t.Errorf("isLoopback(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestOAuthConfigFromEnv(t *testing.T) {
	t.Setenv("SEATS_OAUTH_CLIENT_ID", "cid")
	t.Setenv("SEATS_OAUTH_CLIENT_SECRET", "secret")
	t.Setenv("SEATS_OAUTH_REDIRECT_URI", "https://app/cb")
	t.Setenv("SEATS_OAUTH_BASE_URL", "https://example/oauth2")

	cfg := oauthConfigFromEnv()
	if cfg.ClientID != "cid" || cfg.ClientSecret != "secret" {
		t.Errorf("client creds = %+v", cfg)
	}
	if cfg.RedirectURI != "https://app/cb" || cfg.BaseURL != "https://example/oauth2" {
		t.Errorf("cfg = %+v", cfg)
	}
	if !cfg.Complete() {
		t.Error("expected Complete()")
	}
}

func TestTokenStoreEnvOverride(t *testing.T) {
	t.Setenv("SEATS_OAUTH_TOKEN_FILE", "/tmp/custom-token.json")
	store, err := tokenStore()
	if err != nil {
		t.Fatal(err)
	}
	if store.Path != "/tmp/custom-token.json" {
		t.Errorf("path = %q", store.Path)
	}
}
