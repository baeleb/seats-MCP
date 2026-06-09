package mcptools

import (
	"context"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSearch(s *server.MCPServer, c *seatsaero.Client) {
	tool := mcp.NewTool("search_award_availability",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Search cached award availability across mileage programs (seats.aero GET /search). "+
				"Returns Availability summaries for one or more origin/destination pairs over a date range. "+
				"Use this for fast multi-program searches; pair with get_trips to fetch flight-level detail.",
		),
		mcp.WithString("origin_airport", mcp.Required(),
			mcp.Description("Origin IATA airport code(s). Comma-delimited for multiple, e.g. \"SFO,LAX\".")),
		mcp.WithString("destination_airport", mcp.Required(),
			mcp.Description("Destination IATA airport code(s). Comma-delimited for multiple, e.g. \"FRA,LHR\".")),
		mcp.WithString("start_date",
			mcp.Description("Earliest departure date, YYYY-MM-DD.")),
		mcp.WithString("end_date",
			mcp.Description("Latest departure date, YYYY-MM-DD.")),
		mcp.WithString("sources",
			mcp.Description("Comma-delimited mileage program keys, e.g. \"aeroplan,united\".")),
		mcp.WithString("cabins",
			mcp.Description("Comma-delimited cabins to require, from: economy, premium, business, first.")),
		mcp.WithString("carriers",
			mcp.Description("Comma-delimited airline codes to filter by, e.g. \"DL,AA\".")),
		mcp.WithString("order_by",
			mcp.Description("Sort order. \"lowest_mileage\" or empty for default (departure date + available cabins).")),
		mcp.WithBoolean("only_direct_flights",
			mcp.Description("If true, return only direct flights.")),
		mcp.WithBoolean("include_trips",
			mcp.Description("If true, embed trip-level detail in results. Slower; prefer get_trips for large result sets.")),
		mcp.WithBoolean("minify_trips",
			mcp.Description("With include_trips, reduce trip fields for performance.")),
		mcp.WithBoolean("include_filtered",
			mcp.Description("If true, include raw filtered-out results.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (10-1000, default 500).")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip. Prefer cursor for stable pagination.")),
		mcp.WithNumber("cursor",
			mcp.Description("Cursor from a previous response for stable pagination.")),
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
		args := req.GetArguments()
		params := seatsaero.SearchParams{
			OriginAirport:      origin,
			DestinationAirport: destination,
			StartDate:          req.GetString("start_date", ""),
			EndDate:            req.GetString("end_date", ""),
			Sources:            req.GetString("sources", ""),
			Cabins:             req.GetString("cabins", ""),
			Carriers:           req.GetString("carriers", ""),
			OrderBy:            req.GetString("order_by", ""),
			OnlyDirectFlights:  boolPtrArg(args, "only_direct_flights"),
			IncludeTrips:       boolPtrArg(args, "include_trips"),
			MinifyTrips:        boolPtrArg(args, "minify_trips"),
			IncludeFiltered:    boolPtrArg(args, "include_filtered"),
			Take:               intPtrArg(args, "take"),
			Skip:               intPtrArg(args, "skip"),
			Cursor:             intPtrArg(args, "cursor"),
		}
		resp, err := c.Search(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
