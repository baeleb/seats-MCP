package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsListHotels(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("list_hotels",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Browse the rooms.aero hotel directory, ordered by name (rooms.aero GET /hotels). "+
				"Each hotel carries lowest_award, the cheapest award seen on any upcoming date in the last two weeks. "+
				"Narrow by id, internal_id (program property code), program, location, brand or award category, or search by name. "+
				"To fetch one hotel, pass its id. Use the returned id with bulk_hotel_availability.",
		),
		mcp.WithString("id",
			mcp.Description("Only these hotels, by rooms.aero id, comma-delimited (at most 100).")),
		mcp.WithString("internal_id",
			mcp.Description("Only these hotels, by the program's own property code, comma-delimited (at most 100). Combine with source.")),
		mcp.WithString("source",
			mcp.Description("Loyalty program key(s), comma-delimited, from: hilton, hyatt, ihg, marriott, choice, wyndham, iprefer.")),
		mcp.WithString("country",
			mcp.Description("Only hotels in these countries by display name, comma-delimited, e.g. \"United States\", \"Japan\".")),
		mcp.WithString("state",
			mcp.Description("Only hotels in these states or regions, comma-delimited.")),
		mcp.WithString("city",
			mcp.Description("Only hotels in these cities, comma-delimited.")),
		mcp.WithString("brand",
			mcp.Description("Only these brand codes, comma-delimited.")),
		mcp.WithString("category",
			mcp.Description("Only these award categories, comma-delimited.")),
		mcp.WithString("search",
			mcp.Description("Case-insensitive substring match on the hotel name.")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip. Request the next page with skip = skip + take.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (1-1000, default 100).")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		params := roomsaero.HotelsParams{
			ID:         req.GetString("id", ""),
			InternalID: req.GetString("internal_id", ""),
			Source:     req.GetString("source", ""),
			Country:    req.GetString("country", ""),
			State:      req.GetString("state", ""),
			City:       req.GetString("city", ""),
			Brand:      req.GetString("brand", ""),
			Category:   req.GetString("category", ""),
			Search:     req.GetString("search", ""),
			Skip:       intPtrArg(args, "skip"),
			Take:       intPtrArg(args, "take"),
		}
		resp, err := c.Hotels(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
