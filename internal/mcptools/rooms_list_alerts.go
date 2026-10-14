package mcptools

import (
	"context"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoomsListAlerts(s *server.MCPServer, c *roomsaero.Client) {
	tool := mcp.NewTool("list_hotel_alerts",
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDescription(
			"List the hotel price alerts on your rooms.aero account, newest first (rooms.aero GET /alerts). "+
				"Alerts are created and deleted on the rooms.aero site; this endpoint is read-only and requires a Pro user's key. "+
				"One-off alerts that have already fired are hidden unless include_expired is true.",
		),
		mcp.WithBoolean("include_expired",
			mcp.Description("Include one-off alerts that have already fired.")),
		mcp.WithNumber("skip",
			mcp.Description("Number of results to skip. Request the next page with skip = skip + take.")),
		mcp.WithNumber("take",
			mcp.Description("Max results per page (1-500, default 100).")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		params := roomsaero.AlertsParams{
			IncludeExpired: boolPtrArg(args, "include_expired"),
			Skip:           intPtrArg(args, "skip"),
			Take:           intPtrArg(args, "take"),
		}
		resp, err := c.Alerts(ctx, params)
		if err != nil {
			return apiErrResult(err), nil
		}
		return jsonResult(resp)
	})
}
