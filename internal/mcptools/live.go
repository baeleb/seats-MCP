package mcptools

import (
	"context"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerLiveSearch(s *server.MCPServer, c *seatsaero.Client) {
	tool := mcp.NewTool("live_search",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Real-time award search for a specific route, date, and mileage program (seats.aero POST /live). "+
				"Slower than cached search (typically 5-15s). IDs in the response are ephemeral and cannot be passed to get_trips. "+
				"Failed live searches do not count against the daily quota.",
		),
		mcp.WithString("origin_airport", mcp.Required(),
			mcp.Description("Origin IATA airport code.")),
		mcp.WithString("destination_airport", mcp.Required(),
			mcp.Description("Destination IATA airport code.")),
		mcp.WithString("departure_date", mcp.Required(),
			mcp.Description("Departure date, YYYY-MM-DD.")),
		mcp.WithString("source", mcp.Required(),
			mcp.Description("Mileage program key, e.g. \"aeroplan\".")),
		mcp.WithNumber("seat_count",
			mcp.Description("Adult passengers (1-9). Default 1.")),
		mcp.WithBoolean("disable_filters",
			mcp.Description("If true, remove dynamic pricing and airport mismatch filters.")),
		mcp.WithBoolean("show_dynamic_pricing",
			mcp.Description("If true, remove dynamic pricing filters only.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		origin, err := req.RequireString("origin_airport")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		destination, err := req.RequireString("destination_airport")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		departureDate, err := req.RequireString("departure_date")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		source, err := req.RequireString("source")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		args := req.GetArguments()
		params := seatsaero.LiveParams{
			OriginAirport:      origin,
			DestinationAirport: destination,
			DepartureDate:      departureDate,
			Source:             source,
			SeatCount:          intPtrArg(args, "seat_count"),
			DisableFilters:     boolPtrArg(args, "disable_filters"),
			ShowDynamicPricing: boolPtrArg(args, "show_dynamic_pricing"),
		}
		resp, err := c.LiveSearch(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
