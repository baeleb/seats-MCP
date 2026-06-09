package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/calebhro/seats-MCP/internal/mcptools"
	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "seats-aero"
	serverVersion = "0.1.0"
)

func main() {
	apiKey := os.Getenv("SEATS_AERO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "SEATS_AERO_API_KEY is required")
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	client := seatsaero.New(apiKey, httpClient)

	s := server.NewMCPServer(serverName, serverVersion)
	mcptools.Register(s, client)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
