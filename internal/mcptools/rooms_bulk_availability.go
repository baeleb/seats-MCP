package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsBulkAvailability(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("bulk_hotel_availability",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Retrieve every upcoming date with a rate for a hotel program or a set of hotels (rooms.aero GET /availability). "+
				"Returns one object per hotel, check-in date and stay length (1-5 nights). Use for broad scans such as \"every Hyatt award in Japan under 20,000 points\". "+
				"Requires either source or hotel_id. Take an availability id to get_hotel_availability_details for room types and a booking link.",
		),
		mcp.WithString("source",
			mcp.Description("Loyalty program key(s), comma-delimited, from: hilton, hyatt, ihg, marriott, choice, wyndham, iprefer. Required unless hotel_id is given.")),
		mcp.WithString("hotel_id",
			mcp.Description("Only these hotels, comma-delimited (at most 50). Required unless source is given.")),
		mcp.WithString("start_date",
			mcp.Description("Earliest check-in date, YYYY-MM-DD.")),
		mcp.WithString("end_date",
			mcp.Description("Latest check-in date, YYYY-MM-DD.")),
		mcp.WithNumber("nights",
			mcp.Description("Only stays of exactly this many nights (1-5). Omit for every stay length.")),
		mcp.WithString("brand",
			mcp.Description("Only hotels with these brand codes, comma-delimited.")),
		mcp.WithString("category",
			mcp.Description("Only hotels in these award categories, comma-delimited.")),
		mcp.WithString("country",
			mcp.Description("Only hotels in these countries, comma-delimited.")),
		mcp.WithString("rate_type",
			mcp.Description("Which kind of rate a row must offer: award (points), cash, or any. Default award.")),
		mcp.WithString("room_type",
			mcp.Description("Which room class must offer it: standard, suite, or any. Default any. Also scopes max_points, min_cpp and order_by to that class.")),
		mcp.WithNumber("max_points",
			mcp.Description("Only rows with an award at or below this many points.")),
		mcp.WithNumber("min_cpp",
			mcp.Description("Only rows whose award is worth at least this many US cents per point.")),
		mcp.WithString("order_by",
			mcp.Description("Sort order: arrival_date (soonest), lowest_award (cheapest points), lowest_cash, or last_checked. Default arrival_date.")),
		mcp.WithBoolean("include_hotel",
			mcp.Description("Embed the hotel object and booking link on each row. Set false for a smaller payload when you already have the hotels.")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip. Request the next page with skip = skip + take.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (1-500, default 100).")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := req.GetString("source", "")
		hotelID := req.GetString("hotel_id", "")
		if source == "" && hotelID == "" {
			return mcp.NewToolResultError("provide either source or hotel_id"), nil
		}
		args := req.GetArguments()
		params := roomsaero.AvailabilityParams{
			Source:       source,
			HotelID:      hotelID,
			StartDate:    req.GetString("start_date", ""),
			EndDate:      req.GetString("end_date", ""),
			Nights:       intPtrArg(args, "nights"),
			Brand:        req.GetString("brand", ""),
			Category:     req.GetString("category", ""),
			Country:      req.GetString("country", ""),
			RateType:     req.GetString("rate_type", ""),
			RoomType:     req.GetString("room_type", ""),
			MaxPoints:    intPtrArg(args, "max_points"),
			MinCPP:       floatPtrArg(args, "min_cpp"),
			OrderBy:      req.GetString("order_by", ""),
			IncludeHotel: boolPtrArg(args, "include_hotel"),
			Skip:         intPtrArg(args, "skip"),
			Take:         intPtrArg(args, "take"),
		}
		resp, err := c.BulkAvailability(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
