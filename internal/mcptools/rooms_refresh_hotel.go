package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsRefreshHotel(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("refresh_hotel",
		mcp.WithDescription(
			"Ask for a hotel's rates to be retrieved again ahead of schedule (rooms.aero POST /hotels/{id}/refresh). "+
				"New data typically appears within minutes; poll bulk_hotel_availability with the hotel_id and watch last_checked_at advance. "+
				"Each user may request one refresh every 24 hours, shared with the refresh button on the rooms.aero site.",
		),
		mcp.WithString("hotel_id", mcp.Required(),
			mcp.Description("The rooms.aero hotel id.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("hotel_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		resp, err := c.RefreshHotel(ctx, id)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
