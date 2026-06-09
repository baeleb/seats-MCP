# seats-MCP

A [Model Context Protocol](https://modelcontextprotocol.io) server that wraps the [seats.aero Partner API](https://developers.seats.aero/reference/getting-started-p). Exposes award flight availability search as MCP tools so AI agents can query redemption availability across major airline mileage programs.

Transport: **stdio**. Auth: `SEATS_AERO_API_KEY` env var.

## Tools

| Tool | Endpoint | Use for |
| --- | --- | --- |
| `search_award_availability` | `GET /search` | Cached, multi-program search by route + date range. Fast. |
| `bulk_availability` | `GET /availability` | Single-program bulk fetch, paginated. Explore-style queries. |
| `get_trips` | `GET /trips/{id}` | Drill into a specific Availability for flight-level itineraries. |
| `live_search` | `POST /live` | Real-time search for one route/date/program. Slower (5-15s). |

## Build

```sh
go build -o bin/seats-mcp .
```

Requires Go 1.23+.

## Configure an MCP client

### Claude Desktop / Claude Code

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (Claude Desktop) or your Claude Code MCP config:

```json
{
  "mcpServers": {
    "seats-aero": {
      "command": "/absolute/path/to/seats-MCP/bin/seats-mcp",
      "env": {
        "SEATS_AERO_API_KEY": "your-partner-api-key"
      }
    }
  }
}
```

Restart the client; the four tools above should appear.

### MCP Inspector (manual testing)

```sh
SEATS_AERO_API_KEY=your-key npx @modelcontextprotocol/inspector ./bin/seats-mcp
```

## API key

Pro users self-generate keys from the API tab in [seats.aero settings](https://seats.aero/settings). Commercial users receive keys from seats.aero. Rate limit: **1,000 requests/day**, shared across all uses of the key.

## Notes

- IDs returned by `live_search` are ephemeral and cannot be passed to `get_trips`.
- Times in responses use the **local airport timezone**, not UTC.
- Not all mileage programs report seat counts or trip-level data — treat capability gaps as expected.
- The server logs only to stderr (stdout is the MCP transport).

See [`CLAUDE.md`](./CLAUDE.md) for the full API surface reference.
