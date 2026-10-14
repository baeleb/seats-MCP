package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/baeleb/seats-MCP/internal/seatsaero"
)

// TestSharedAuthorizerSendsPartnerHeader locks in the credential contract of
// the runServer wiring: one seatsaero.OAuthAuthorizer is shared by both
// clients, and an OAuth token must reach both Partner APIs as
// "Partner-Authorization: Bearer <token>" — the plain Authorization header
// belongs to the OAuth2 /userinfo endpoint only.
func TestSharedAuthorizerSendsPartnerHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Partner-Authorization"); got != "Bearer seats:ota:test" {
			t.Errorf("%s: Partner-Authorization = %q, want %q", r.URL.Path, got, "Bearer seats:ota:test")
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("%s: unexpected Authorization header %q", r.URL.Path, got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ts := seatsaero.NewTokenSource(seatsaero.OAuthConfig{}, &seatsaero.Token{
		AccessToken: "seats:ota:test",
		Expiry:      time.Now().Add(time.Hour),
	}, nil, srv.Client())
	auth := seatsaero.OAuthAuthorizer{Source: ts}

	seats := seatsaero.NewWithAuthorizerBase(auth, srv.Client(), srv.URL)
	rooms := roomsaero.NewWithAuthorizerBase(auth, srv.Client(), srv.URL)

	if _, err := seats.Search(context.Background(), seatsaero.SearchParams{
		OriginAirport: "JFK", DestinationAirport: "NRT",
	}); err != nil {
		t.Fatalf("seats.aero search: %v", err)
	}
	if _, err := rooms.Hotels(context.Background(), roomsaero.HotelsParams{Source: "hyatt"}); err != nil {
		t.Fatalf("rooms.aero hotels: %v", err)
	}
}
