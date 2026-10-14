package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsSearch(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("search_hotel_availability",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Find hotels with award availability in an area and date range, cheapest award first (rooms.aero GET /search). "+
				"Covers Hilton, Hyatt, IHG, Marriott, Choice, Wyndham and I Prefer. "+
				"Describe the area with a bounding box (sw_lat, sw_lng, ne_lat, ne_lng) or a location to geocode. "+
				"Each result embeds the hotel and its cheapest standard and suite rates; take a hotel.id to bulk_hotel_availability, then an availability id to get_hotel_availability_details.",
		),
		mcp.WithString("start_date", mcp.Required(),
			mcp.Description("Earliest check-in date, YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Required(),
			mcp.Description("Latest check-in date, YYYY-MM-DD. Must not be before start_date.")),
		mcp.WithNumber("sw_lat",
			mcp.Description("South-west corner latitude (-90 to 90). Provide all four corners, or location.")),
		mcp.WithNumber("sw_lng",
			mcp.Description("South-west corner longitude (-180 to 180).")),
		mcp.WithNumber("ne_lat",
			mcp.Description("North-east corner latitude; must be north of sw_lat.")),
		mcp.WithNumber("ne_lng",
			mcp.Description("North-east corner longitude; must be east of sw_lng.")),
		mcp.WithString("location",
			mcp.Description("A place to geocode instead of a bounding box, such as \"Tokyo\" or \"Maui\". Ignored when a complete box is given.")),
		mcp.WithNumber("nights",
			mcp.Description("Only stays of exactly this many nights (1-5). Omit for every stay length.")),
		mcp.WithString("source",
			mcp.Description("Loyalty program key(s), comma-delimited, from: hilton, hyatt, ihg, marriott, choice, wyndham, iprefer.")),
		mcp.WithNumber("max_points",
			mcp.Description("Only hotels with an award at or below this many points.")),
		mcp.WithNumber("min_cpp",
			mcp.Description("Only hotels whose best award is worth at least this many US cents per point.")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip. Request the next page with skip = skip + take.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (1-500, default 100).")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startDate, err := req.RequireString("start_date")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		endDate, err := req.RequireString("end_date")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		args := req.GetArguments()
		params := roomsaero.SearchParams{
			StartDate: startDate,
			EndDate:   endDate,
			SWLat:     floatPtrArg(args, "sw_lat"),
			SWLng:     floatPtrArg(args, "sw_lng"),
			NELat:     floatPtrArg(args, "ne_lat"),
			NELng:     floatPtrArg(args, "ne_lng"),
			Location:  req.GetString("location", ""),
			Nights:    intPtrArg(args, "nights"),
			Source:    req.GetString("source", ""),
			MaxPoints: intPtrArg(args, "max_points"),
			MinCPP:    floatPtrArg(args, "min_cpp"),
			Skip:      intPtrArg(args, "skip"),
			Take:      intPtrArg(args, "take"),
		}
		resp, err := c.Search(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
