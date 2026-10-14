package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsAvailabilityDetails(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("get_hotel_availability_details",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"Retrieve one hotel availability with every room type, its points and cash price, and a booking link for that exact stay (rooms.aero GET /availability/{id}). "+
				"Use an availability id from bulk_hotel_availability. Rooms are returned cheapest award first.",
		),
		mcp.WithString("availability_id", mcp.Required(),
			mcp.Description("The id field of an Availability object from bulk_hotel_availability.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("availability_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		resp, err := c.AvailabilityDetail(ctx, id)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
