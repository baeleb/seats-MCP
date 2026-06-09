# seats-MCP

A [Model Context Protocol](https://modelcontextprotocol.io) server for the [seats.aero Partner API](https://developers.seats.aero/reference/getting-started-p). Exposes award flight availability as MCP tools so AI coding agents can search redemption availability across major airline mileage programs.

## Quickstart

### 1. Get an API key

You need a seats.aero account with a **Pro** or **Commercial** plan.

1. Sign up at [seats.aero](https://seats.aero) and subscribe to Pro.
2. Go to [seats.aero/settings](https://seats.aero/settings) and open the **API** tab.
3. Click **Generate API Key**. Copy the key — it is shown only once.

> **Note:** seats.aero may limit API key access by geographical location or other criteria at their sole discretion. Pro keys are for **non-commercial use only**; commercial use requires a written agreement with seats.aero.

Rate limit: **1,000 requests/day** shared across all uses of the key.

### 2. Build the server

```sh
git clone https://github.com/baeleb/seats-MCP.git
cd seats-MCP
go build -o bin/seats-mcp .
```

Requires Go 1.26+.

### 3. Set your API key

Choose one method to make the key available at runtime:

**Option A — Shell profile (persistent, simplest)**

Add this line to `~/.zshrc` (or `~/.bashrc`):

```sh
export SEATS_AERO_API_KEY="your-partner-api-key"
```

Then restart your shell or run `source ~/.zshrc`.

**Option B — Inline per-command**

```sh
SEATS_AERO_API_KEY="your-key" ./bin/seats-mcp
```

### 4. Add to your MCP client

#### OpenCode

Add to `~/.config/opencode/opencode.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "seats-aero": {
      "type": "local",
      "command": ["/path/to/seats-MCP/bin/seats-mcp"],
      "enabled": true,
      "environment": {
        "SEATS_AERO_API_KEY": "{env:SEATS_AERO_API_KEY}"
      }
    }
  }
}
```

Use an absolute path to the binary. The `{env:SEATS_AERO_API_KEY}` syntax pulls the value from your shell environment — no secret written to the config file.

Restart OpenCode. The four tools below will appear under the `seats-aero_` prefix.

#### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

#### Claude Code

Add to your Claude Code MCP config:

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

#### MCP Inspector (manual testing)

```sh
SEATS_AERO_API_KEY="your-key" npx @modelcontextprotocol/inspector ./bin/seats-mcp
```

## Tools

| Tool | Endpoint | Use for |
| --- | --- | --- |
| `search_award_availability` | `GET /search` | Cached, multi-program search by route + date range. Fast. |
| `bulk_availability` | `GET /availability` | Single-program bulk fetch, paginated. Explore-style queries. |
| `get_trips` | `GET /trips/{id}` | Drill into a specific Availability for flight-level itineraries. |
| `live_search` | `POST /live` | Real-time search for one route/date/program. Slower (5-15s). |

## Notes

- **Live search is not available to Pro users.** The `/live` endpoint requires a commercial agreement with seats.aero. If you are a Pro user, the `live_search` tool will return errors.
- IDs returned by `live_search` are ephemeral and cannot be passed to `get_trips`.
- Times in responses use the **local airport timezone**, not UTC.
- Not all mileage programs report seat counts or trip-level data — treat capability gaps as expected.
- The server logs only to stderr (stdout is the MCP transport).

See [`CLAUDE.md`](./CLAUDE.md) for the full API surface reference.

## License

GPLv3. See [LICENSE](./LICENSE).
