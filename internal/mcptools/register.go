// Package mcptools wraps the seats.aero and rooms.aero clients as MCP tools.
package mcptools

import (
	"encoding/json"
	"errors"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/baeleb/seats-MCP/internal/seatsaero"
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

// RegisterRooms adds all rooms.aero (hotel award) tools to the given MCP server.
func RegisterRooms(s *server.MCPServer, c *roomsaero.Client) {
	registerRoomsSearch(s, c)
	registerRoomsBulkAvailability(s, c)
	registerRoomsAvailabilityDetails(s, c)
	registerRoomsListHotels(s, c)
	registerRoomsRefreshHotel(s, c)
	registerRoomsListAlerts(s, c)
}

// jsonResult marshals v and wraps it in a text tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("encode result: %v", err), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

// apiErrResult turns a seatsaero or roomsaero error into an MCP tool-error
// result. The Go error return is reserved for transport failures, so callers
// always return nil alongside this.
func apiErrResult(err error) *mcp.CallToolResult {
	if err == nil {
		return mcp.NewToolResultError("unknown error")
	}
	var seatsErr *seatsaero.APIError
	if errors.As(err, &seatsErr) {
		return mcp.NewToolResultErrorf("seats.aero %d: %s", seatsErr.Status, seatsErr.Body)
	}
	var roomsErr *roomsaero.APIError
	if errors.As(err, &roomsErr) {
		return mcp.NewToolResultError(roomsErr.Error())
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

// floatPtrArg reads an optional float arg by checking key presence.
func floatPtrArg(args map[string]any, key string) *float64 {
	v, ok := args[key]
	if !ok {
		return nil
	}
	switch n := v.(type) {
	case float64:
		return &n
	case int:
		f := float64(n)
		return &f
	}
	return nil
}
