package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/baeleb/seats-MCP/internal/seatsaero"
)

// loginTimeout bounds how long `auth login` waits for the consent redirect.
const loginTimeout = 5 * time.Minute

// oauthConfigFromEnv reads OAuth application credentials from the environment.
func oauthConfigFromEnv() seatsaero.OAuthConfig {
	return seatsaero.OAuthConfig{
		ClientID:     os.Getenv("SEATS_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("SEATS_OAUTH_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("SEATS_OAUTH_REDIRECT_URI"),
		BaseURL:      os.Getenv("SEATS_OAUTH_BASE_URL"),
	}
}

// tokenStore returns the configured token store: SEATS_OAUTH_TOKEN_FILE when
// set, otherwise the default per-user location.
func tokenStore() (seatsaero.FileTokenStore, error) {
	if p := os.Getenv("SEATS_OAUTH_TOKEN_FILE"); p != "" {
		return seatsaero.FileTokenStore{Path: p}, nil
	}
	p, err := seatsaero.DefaultTokenPath()
	if err != nil {
		return seatsaero.FileTokenStore{}, err
	}
	return seatsaero.FileTokenStore{Path: p}, nil
}

// runAuth dispatches `auth <subcommand>`.
func runAuth(args []string) error {
	sub := "help"
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "login":
		return runAuthLogin(context.Background())
	case "status", "whoami":
		return runAuthStatus(context.Background())
	case "logout":
		return runAuthLogout()
	default:
		fmt.Fprint(os.Stderr, authUsage)
		if sub != "help" && sub != "-h" && sub != "--help" {
			return fmt.Errorf("unknown subcommand %q", sub)
		}
		return nil
	}
}

const authUsage = `Manage seats.aero OAuth2 credentials.

Usage:
  seats-mcp auth login     Run the OAuth2 consent flow and store tokens
  seats-mcp auth status    Show the stored token and authenticated user
  seats-mcp auth logout    Delete the stored token

Required environment variables for login:
  SEATS_OAUTH_CLIENT_ID       Your registered application's client ID
  SEATS_OAUTH_CLIENT_SECRET   Your registered application's client secret
  SEATS_OAUTH_REDIRECT_URI    A redirect URI registered for the app

Optional:
  SEATS_OAUTH_TOKEN_FILE      Where to persist tokens (default: per-user config dir)
  SEATS_OAUTH_BASE_URL        Override the OAuth2 base URL (default: https://seats.aero/oauth2)
`

func runAuthLogin(ctx context.Context) error {
	cfg := oauthConfigFromEnv()
	if !cfg.Complete() {
		return fmt.Errorf("missing OAuth config; set SEATS_OAUTH_CLIENT_ID, SEATS_OAUTH_CLIENT_SECRET, and SEATS_OAUTH_REDIRECT_URI")
	}
	store, err := tokenStore()
	if err != nil {
		return err
	}
	ts := seatsaero.NewTokenSource(cfg, nil, store, nil)

	state, err := randomState()
	if err != nil {
		return err
	}
	authURL := ts.AuthCodeURL(state)

	ctx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()

	code, err := obtainCode(ctx, cfg.RedirectURI, authURL, state)
	if err != nil {
		return err
	}

	if _, err := ts.Exchange(ctx, code, state); err != nil {
		return fmt.Errorf("exchange authorization code: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Authorized. Token saved to %s\n", store.Path)

	if info, err := ts.UserInfo(ctx); err == nil {
		fmt.Fprintf(os.Stderr, "Signed in as %s <%s>\n", info.Name, info.Email)
	}
	return nil
}

// obtainCode runs the appropriate redirect-capture strategy. For an http
// loopback redirect URI it spins a local callback server; otherwise it falls
// back to having the user paste the redirected URL.
func obtainCode(ctx context.Context, redirectURI, authURL, state string) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	if u.Scheme == "http" && isLoopback(u.Hostname()) {
		return captureViaServer(ctx, u, authURL, state)
	}
	return captureViaPaste(authURL, state)
}

func captureViaServer(ctx context.Context, u *url.URL, authURL, state string) (string, error) {
	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)

	callbackPath := u.Path
	if callbackPath == "" {
		callbackPath = "/"
	}

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			http.Error(w, "Authorization failed: "+e, http.StatusBadRequest)
			results <- result{err: fmt.Errorf("authorization error: %s", e)}
			return
		}
		if q.Get("state") != state {
			http.Error(w, "State mismatch.", http.StatusBadRequest)
			results <- result{err: fmt.Errorf("state mismatch (possible CSRF)")}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing code.", http.StatusBadRequest)
			results <- result{err: fmt.Errorf("no code in redirect")}
			return
		}
		fmt.Fprintln(w, "Authorization complete. You can close this tab and return to the terminal.")
		results <- result{code: code}
	})

	ln, err := net.Listen("tcp", u.Host)
	if err != nil {
		return "", fmt.Errorf("listen on %s: %w (is the redirect port free?)", u.Host, err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	fmt.Fprintf(os.Stderr, "Open this URL to authorize:\n\n  %s\n\nWaiting for the redirect...\n", authURL)
	openBrowser(authURL)

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("timed out waiting for authorization: %w", ctx.Err())
	case res := <-results:
		return res.code, res.err
	}
}

func captureViaPaste(authURL, state string) (string, error) {
	fmt.Fprintf(os.Stderr, "Open this URL to authorize:\n\n  %s\n\n", authURL)
	openBrowser(authURL)
	fmt.Fprint(os.Stderr, "After authorizing, paste the full redirect URL (or just the code) here:\n> ")

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return parsePastedCode(strings.TrimSpace(line), state)
}

// parsePastedCode extracts the authorization code from user-pasted input, which
// may be either the full redirect URL or a bare code. When a URL carries a state
// param it must match the expected state.
func parsePastedCode(line, state string) (string, error) {
	if line == "" {
		return "", fmt.Errorf("no input provided")
	}
	// A pasted redirect URL carries code + state as query params; a bare code
	// has neither, so treat it as the code directly.
	if strings.Contains(line, "?") || strings.Contains(line, "://") {
		u, err := url.Parse(line)
		if err != nil {
			return "", fmt.Errorf("parse pasted URL: %w", err)
		}
		q := u.Query()
		if got := q.Get("state"); got != "" && got != state {
			return "", fmt.Errorf("state mismatch (possible CSRF)")
		}
		code := q.Get("code")
		if code == "" {
			return "", fmt.Errorf("no code found in pasted URL")
		}
		return code, nil
	}
	return line, nil
}

func runAuthStatus(ctx context.Context) error {
	cfg := oauthConfigFromEnv()
	store, err := tokenStore()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return err
	}
	if tok == nil {
		fmt.Fprintln(os.Stderr, "No stored OAuth token. Run `seats-mcp auth login`.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Token file: %s\n", store.Path)
	fmt.Fprintf(os.Stderr, "Scope:      %s\n", tok.Scope)
	if tok.Valid(time.Now()) {
		fmt.Fprintf(os.Stderr, "Status:     valid (expires %s)\n", tok.Expiry.Format(time.RFC3339))
	} else {
		fmt.Fprintf(os.Stderr, "Status:     expired (expired %s; will refresh on next use)\n", tok.Expiry.Format(time.RFC3339))
	}

	if !cfg.Complete() {
		fmt.Fprintln(os.Stderr, "(Set the SEATS_OAUTH_* env vars to refresh and fetch user info.)")
		return nil
	}
	ts := seatsaero.NewTokenSource(cfg, tok, store, nil)
	info, err := ts.UserInfo(ctx)
	if err != nil {
		return fmt.Errorf("fetch user info: %w", err)
	}
	fmt.Fprintf(os.Stderr, "User:       %s <%s>\n", info.Name, info.Email)
	if info.SubscriptionExpires > 0 {
		fmt.Fprintf(os.Stderr, "Pro until:  %s\n", time.Unix(info.SubscriptionExpires, 0).Format(time.RFC3339))
	}
	return nil
}

func runAuthLogout() error {
	store, err := tokenStore()
	if err != nil {
		return err
	}
	if err := store.Delete(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Removed stored token at %s\n", store.Path)
	return nil
}

// randomState returns a cryptographically random opaque state value.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// openBrowser best-effort opens url in the user's default browser. Failures are
// ignored — the URL is always printed for manual use. Set SEATS_MCP_NO_BROWSER
// to skip launching a browser (e.g. headless/remote sessions).
func openBrowser(url string) {
	if os.Getenv("SEATS_MCP_NO_BROWSER") != "" {
		return
	}
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}
