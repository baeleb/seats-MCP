package seatsaero

import (
	"context"
	"errors"
	"net/http"
)

// ErrNoCredentials is returned by ResolveAuthorizer when neither an OAuth token
// nor an API key is available.
var ErrNoCredentials = errors.New("no credentials: set SEATS_AERO_API_KEY or run `seats-mcp auth login`")

// Authorizer attaches authentication to an outgoing Partner API request.
//
// Two schemes are supported:
//   - APIKeyAuthorizer  — the static Partner-Authorization header (Pro keys).
//   - OAuthAuthorizer    — an OAuth2 bearer token, refreshed on demand, for
//     consumer-facing integrations that connect a user's seats.aero account.
type Authorizer interface {
	// Apply mutates req to carry valid credentials. It may perform network I/O
	// (e.g. refreshing an expired token), so ctx is threaded through.
	Apply(ctx context.Context, req *http.Request) error
}

// APIKeyAuthorizer authenticates with a static Partner API key sent in the
// Partner-Authorization header.
type APIKeyAuthorizer struct {
	Key string
}

func (a APIKeyAuthorizer) Apply(_ context.Context, req *http.Request) error {
	req.Header.Set("Partner-Authorization", a.Key)
	return nil
}

// OAuthAuthorizer authenticates with an OAuth2 access token sent as
// "Authorization: Bearer <token>". The token is fetched (and refreshed when
// expired) from the wrapped TokenSource on every request.
type OAuthAuthorizer struct {
	Source *TokenSource
}

func (a OAuthAuthorizer) Apply(ctx context.Context, req *http.Request) error {
	tok, err := a.Source.Token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	return nil
}

// ResolveAuthorizer selects how the server authenticates, preferring OAuth when
// a usable token is present and falling back to the API key otherwise.
//
// A TokenSource is "usable" when it holds a token that is either still valid or
// carries a refresh token (so it can be renewed). ts may be nil. The returned
// string is a short description suitable for a startup log line.
func ResolveAuthorizer(apiKey string, ts *TokenSource) (Authorizer, string, error) {
	if ts != nil {
		if tok := ts.Peek(); tok != nil && (tok.Valid(ts.now()) || tok.RefreshToken != "") {
			return OAuthAuthorizer{Source: ts}, "oauth", nil
		}
	}
	if apiKey != "" {
		return APIKeyAuthorizer{Key: apiKey}, "api_key", nil
	}
	return nil, "", ErrNoCredentials
}
