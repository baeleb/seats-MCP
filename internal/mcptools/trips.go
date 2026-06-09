package mcptools

import (
	"context"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerGetTrips(s *server.MCPServer, c *seatsaero.Client) {
	tool := mcp.NewTool("get_trips",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Fetch detailed flight itineraries (Trips) for an Availability ID (seats.aero GET /trips/{id}). "+
				"Use IDs from search_award_availability or bulk_availability. IDs from live_search are NOT valid here.",
		),
		mcp.WithString("availability_id", mcp.Required(),
			mcp.Description("The ID field of an Availability object.")),
		mcp.WithBoolean("include_filtered",
			mcp.Description("If true, include expensive dynamically-priced trips that were filtered.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("availability_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		includeFiltered := req.GetBool("include_filtered", false)
		resp, err := c.Trips(ctx, id, includeFiltered)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
