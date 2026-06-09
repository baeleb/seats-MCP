// Package mcptools wraps the seats.aero client as MCP tools.
package mcptools

import (
	"encoding/json"
	"errors"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register adds all seats.aero tools to the given MCP server.
func Register(s *server.MCPServer, c *seatsaero.Client) {
	registerSearch(s, c)
	registerBulkAvailability(s, c)
	registerGetTrips(s, c)
	registerLiveSearch(s, c)
}

// jsonResult marshals v and wraps it in a text tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("encode result: %v", err), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

// apiErrResult turns a seatsaero error into an MCP tool-error result.
// The Go error return is reserved for transport failures, so we always return nil here.
func apiErrResult(err error) *mcp.CallToolResult {
	if err == nil {
		return mcp.NewToolResultError("unknown error")
	}
	var apiErr *seatsaero.APIError
	if errors.As(err, &apiErr) {
		return mcp.NewToolResultErrorf("seats.aero %d: %s", apiErr.Status, apiErr.Body)
	}
	return mcp.NewToolResultError(err.Error())
}

// boolPtrArg reads an optional bool arg by checking key presence (not value).
// Returns nil when the key was not supplied.
func boolPtrArg(args map[string]any, key string) *bool {
	v, ok := args[key]
	if !ok {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

// intPtrArg reads an optional int arg by checking key presence.
func intPtrArg(args map[string]any, key string) *int {
	v, ok := args[key]
	if !ok {
		return nil
	}
	switch n := v.(type) {
	case float64:
		i := int(n)
		return &i
	case int:
		return &n
	}
	return nil
}

