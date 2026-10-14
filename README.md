# seats-MCP

> Your coding agent now has frequent flyer status.

Ask your AI to find award flights and hotel stays — right from the terminal, inside your editor, or mid-refactor. seats-MCP connects AI coding agents to the [seats.aero](https://developers.seats.aero/reference/getting-started-p) and [rooms.aero](https://developers.seats.aero/reference/rooms-getting-started) Partner APIs, giving them instant access to award availability across 20+ airline mileage programs and seven hotel loyalty programs (Hilton, Hyatt, IHG, Marriott, Choice, Wyndham and I Prefer).

**Example:** "Find me a business class seat from LAX to Tokyo on June 15th."

Your agent searches Aeroplan, United, Delta, Alaska, Emirates, and more — all in one call — and comes back with mileage costs, seat counts, airlines, and dates. No browser tabs. No point-and-click calendar hunting. Just ask.

Built on the [Model Context Protocol](https://modelcontextprotocol.io). stdio transport.

## Quickstart

### 1. Get an API key

You need a seats.aero account with a **Pro** or **Commercial** plan.

1. Sign up at [seats.aero](https://seats.aero) and subscribe to Pro.
2. Go to [seats.aero/settings](https://seats.aero/settings) and open the **API** tab.
3. Copy your API key. Keep this secret! 

> **Note:** seats.aero may limit API key access by geographical location or other criteria at their sole discretion. Pro keys are for **non-commercial use only**; commercial use requires a written agreement with seats.aero.

Rate limit: **1,000 requests/day** shared across all uses of the key.

### 2. Get the binary

**Option A — Download (recommended)**

Go to [Releases](https://github.com/baeleb/seats-MCP/releases) and download the binary for your platform, then make it executable:

```sh
chmod +x seats-mcp_darwin_arm64
```

| Platform | File |
|---|---|
| macOS Apple Silicon | `seats-mcp_darwin_arm64` |
| macOS Intel | `seats-mcp_darwin_amd64` |
| Linux x86_64 | `seats-mcp_linux_amd64` |
| Linux ARM | `seats-mcp_linux_arm64` |
| Windows x86_64 | `seats-mcp_windows_amd64.exe` |

**Option B — Build from source**

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

### 3b. (Alternative) Authenticate with OAuth2

Instead of a static API key, the server can connect a user's seats.aero Pro account via OAuth2 — useful for consumer-facing integrations. The server **prefers OAuth** when OAuth credentials and a saved token are present, and otherwise falls back to the API key, so the two can coexist.

1. Register an application at [seats.aero](https://seats.aero) to obtain a **client ID** and **client secret**. seats.aero requires the **redirect URI to be a public HTTPS URL** (loopback `http://localhost` URIs are not accepted), and it must share a domain with the Website URL you register. Any HTTPS page that simply loads works — e.g. your project's GitHub repo or a GitHub Pages site — because you complete the flow by pasting the redirected URL back into the CLI; the page itself doesn't need to handle the code.

2. Set the OAuth environment variables. `SEATS_OAUTH_REDIRECT_URI` must match the registered URI character-for-character:

   ```sh
   export SEATS_OAUTH_CLIENT_ID="your-client-id"
   export SEATS_OAUTH_CLIENT_SECRET="your-client-secret"
   export SEATS_OAUTH_REDIRECT_URI="https://github.com/you/your-repo"
   ```

3. Run the consent flow once:

   ```sh
   ./bin/seats-mcp auth login
   ```

   This prints (and opens) the consent URL. After you authorize, seats.aero redirects your browser to your HTTPS redirect URI with `?code=…&state=…` appended. Copy the **full URL from the address bar** and paste it at the prompt — the server extracts the code, exchanges it, and saves the tokens. (You can paste just the `code` value instead if you prefer. Set `SEATS_MCP_NO_BROWSER=1` to skip auto-opening the browser on headless machines.)

   Tokens are stored at `<user-config-dir>/seats-mcp/token.json` (override with `SEATS_OAUTH_TOKEN_FILE`). The server refreshes the access token automatically as it expires.

   ```sh
   ./bin/seats-mcp auth status    # show the stored token + signed-in user
   ./bin/seats-mcp auth logout     # delete the stored token
   ```

The OAuth access token is sent as `Partner-Authorization: Bearer <token>` on every Partner API request — the same header as an API key, prefixed with `Bearer`.

> **Loopback callback (optional):** if seats.aero ever permits an `http://localhost:PORT/callback` redirect URI for your app, `auth login` will instead spin up a temporary local listener and capture the redirect automatically — no pasting needed. The CLI picks this path on its own when the redirect URI is a loopback address.

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

Restart OpenCode. The tools below will appear under the `seats-aero_` prefix.

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

Flights (seats.aero):

| Tool | Endpoint | Use for |
| --- | --- | --- |
| `search_award_availability` | `GET /search` | Cached, multi-program search by route + date range. Fast. |
| `bulk_availability` | `GET /availability` | Single-program bulk fetch, paginated. Explore-style queries. |
| `get_trips` | `GET /trips/{id}` | Drill into a specific Availability for flight-level itineraries. |
| `live_search` | `POST /live` | Real-time search for one route/date/program. Slower (5-15s). |

Hotels (rooms.aero):

| Tool | Endpoint | Use for |
| --- | --- | --- |
| `search_hotel_availability` | `GET /search` | Find hotels with award availability in an area + date range, cheapest first. |
| `bulk_hotel_availability` | `GET /availability` | Every upcoming date with a rate for a program or a set of hotels. |
| `get_hotel_availability_details` | `GET /availability/{id}` | Room types with points/cash prices and a booking link for one stay. |
| `list_hotels` | `GET /hotels` | Browse the hotel directory by program, location, brand or name. |
| `refresh_hotel` | `POST /hotels/{id}/refresh` | Request a fresh rate check (one per user every 24h). |
| `list_hotel_alerts` | `GET /alerts` | List the price alerts on your rooms.aero account (read-only). |

## Notes

- **Flights and hotels share the same credentials.** Your seats.aero API key (or OAuth token) authenticates both the seats.aero and rooms.aero APIs. rooms.aero keeps its own separate quota of 1,000 calls/day, so hotel calls never count against your seats.aero quota (or vice versa).
- **Live search is not available to Pro users.** The `/live` endpoint requires a commercial agreement with seats.aero. If you are a Pro user, the `live_search` tool will return errors.
- IDs returned by `live_search` are ephemeral and cannot be passed to `get_trips`.
- Times in responses use the **local airport timezone**, not UTC.
- Not all mileage programs report seat counts or trip-level data — treat capability gaps as expected.
- The server logs only to stderr (stdout is the MCP transport).

See [`AGENTS.md`](./AGENTS.md) for the full API surface reference. (A [`CLAUDE.md`](./CLAUDE.md) symlink is kept for Claude Code, which reads that name.)

## License

GPLv3. See [LICENSE](./LICENSE).
