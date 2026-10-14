package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/baeleb/seats-MCP/internal/mcptools"
	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/baeleb/seats-MCP/internal/seatsaero"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "seats-aero"
	serverVersion = "0.1.0"
)

func main() {
	godotenv.Load()

	// Subcommands: `auth ...` manages OAuth credentials; no args runs the
	// MCP server (the default, since MCP clients launch the binary directly).
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "auth" {
		if err := runAuth(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "auth: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runServer() error {
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Build an OAuth TokenSource if credentials are configured and a token has
	// been saved. ResolveAuthorizer then prefers OAuth over the API key.
	var ts *seatsaero.TokenSource
	cfg := oauthConfigFromEnv()
	if cfg.Complete() {
		store, err := tokenStore()
		if err != nil {
			return fmt.Errorf("oauth token store: %w", err)
		}
		tok, err := store.Load()
		if err != nil {
			return fmt.Errorf("load oauth token: %w", err)
		}
		if tok != nil {
			ts = seatsaero.NewTokenSource(cfg, tok, store, httpClient)
		}
	}

	auth, mode, err := seatsaero.ResolveAuthorizer(os.Getenv("SEATS_AERO_API_KEY"), ts)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "seats-aero: authenticating via %s\n", mode)

	client := seatsaero.NewWithAuthorizer(auth, httpClient)

	// rooms.aero accepts the same credentials, so it shares the resolved
	// Authorizer (API key or OAuth token).
	roomsClient := roomsaero.NewWithAuthorizer(auth, httpClient)

	s := server.NewMCPServer(serverName, serverVersion)
	mcptools.Register(s, client)
	mcptools.RegisterRooms(s, roomsClient)

	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
