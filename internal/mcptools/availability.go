package mcptools

import (
	"context"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBulkAvailability(s *server.MCPServer, c *seatsaero.Client) {
	tool := mcp.NewTool("bulk_availability",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Bulk-fetch Availability objects for a single mileage program (seats.aero GET /availability). "+
				"Use this for Explore-style queries over a region or date range. Returns large result sets paginated by cursor.",
		),
		mcp.WithString("source", mcp.Required(),
			mcp.Description("Mileage program key, e.g. \"aeroplan\", \"united\", \"delta\".")),
		mcp.WithString("cabin",
			mcp.Description("Filter by cabin. One of: economy, premium, business, first.")),
		mcp.WithString("start_date",
			mcp.Description("Earliest departure date, YYYY-MM-DD.")),
		mcp.WithString("end_date",
			mcp.Description("Latest departure date, YYYY-MM-DD.")),
		mcp.WithString("origin_region",
			mcp.Description("Filter by origin region: North America, South America, Africa, Asia, Europe, Oceania.")),
		mcp.WithString("destination_region",
			mcp.Description("Filter by destination region: North America, South America, Africa, Asia, Europe, Oceania.")),
		mcp.WithBoolean("include_filtered",
			mcp.Description("If true, include raw filtered-out results.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (10-1000, default 500).")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip.")),
		mcp.WithNumber("cursor",
			mcp.Description("Cursor from a previous response for stable pagination.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source, err := req.RequireString("source")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		args := req.GetArguments()
		params := seatsaero.AvailabilityParams{
			Source:            source,
			Cabin:             req.GetString("cabin", ""),
			StartDate:         req.GetString("start_date", ""),
			EndDate:           req.GetString("end_date", ""),
			OriginRegion:      req.GetString("origin_region", ""),
			DestinationRegion: req.GetString("destination_region", ""),
			IncludeFiltered:   boolPtrArg(args, "include_filtered"),
			Take:              intPtrArg(args, "take"),
			Skip:              intPtrArg(args, "skip"),
			Cursor:            intPtrArg(args, "cursor"),
		}
		resp, err := c.BulkAvailability(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
