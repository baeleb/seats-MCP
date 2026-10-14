# CLAUDE.md

Guidance for coding agents working with this repository.

## Build

```sh
go build -o bin/seats-mcp .
```

Requires Go 1.26+.

## Test

```sh
go test ./...          # CI runs this with -race -cover
```

Tests use the standard `testing` package. Partner API responses are stubbed with `httptest.NewServer` (point clients at it via `NewWithBase`/`NewWithAuthorizerBase`), and MCP tools are exercised end-to-end through `s.HandleMessage`. Tests live alongside source as `*_test.go`.

## Architecture

```
main.go                        Entry point — wires up auth → seats.aero + rooms.aero clients → MCP server (default), or dispatches `auth` subcommand
auth.go                        `auth login|status|logout` CLI: OAuth2 consent flow (loopback callback / paste fallback), env→config, token store
internal/
  seatsaero/                   Typed API client for seats.aero Partner API
    client.go                  HTTP client core: doJSON() applies an Authorizer, handles errors, JSON encoding
    auth.go                    Authorizer interface; APIKeyAuthorizer + OAuthAuthorizer (both via the Partner-Authorization header); ResolveAuthorizer (prefers OAuth)
    oauth.go                   OAuthConfig, Token, TokenSource (caching/refresh/code-exchange), AuthCodeURL, UserInfo
    oauthstore.go              TokenStore interface; FileTokenStore (0600, atomic) + MemoryTokenStore; DefaultTokenPath
    types.go                   All data types: Availability, Trip, Segment, SearchResponse, etc.
    search.go                  GET /search — cached multi-program search
    availability.go            GET /availability — single-program bulk fetch
    trips.go                   GET /trips/{id} — flight-level detail from an Availability ID
    live.go                    POST /live — real-time search
    errors.go                  APIError type for non-2xx responses
  roomsaero/                   Typed API client for the rooms.aero Partner API (hotel awards)
    client.go                  HTTP client core: doJSON() applies an Authorizer and parses the {error,message} envelope
    auth.go                    Authorizer interface (structural — the seatsaero Authorizer value is reused)
    query.go                   url.Values helpers for optional string/bool/int/float params
    types.go                   Hotel, BookingLink, Room, Availability, SearchResult, Alert + list envelopes
    search.go                  GET /search — hotels with award availability in an area + date range
    availability.go            GET /availability (bulk rows) + GET /availability/{id} (room-level detail)
    hotels.go                  GET /hotels (directory) + POST /hotels/{id}/refresh
    alerts.go                  GET /alerts — the account's hotel price alerts
    errors.go                  APIError type for non-2xx responses
  mcptools/                    MCP tool layer — wraps the seats.aero and rooms.aero clients as MCP tools
    register.go                Register() + RegisterRooms() wire all tools; jsonResult, apiErrResult, arg helpers
    search.go                  "search_award_availability" tool
    availability.go            "bulk_availability" tool
    trips.go                   "get_trips" tool
    live.go                    "live_search" tool
    rooms_search.go            "search_hotel_availability" tool
    rooms_bulk_availability.go "bulk_hotel_availability" tool
    rooms_availability_details.go "get_hotel_availability_details" tool
    rooms_list_hotels.go       "list_hotels" tool
    rooms_refresh_hotel.go     "refresh_hotel" tool
    rooms_list_alerts.go       "list_hotel_alerts" tool
```

**Flow:** `main.go` → `mcptools.Register(s, client)` / `mcptools.RegisterRooms(s, roomsClient)` → each tool handler calls the matching `seatsaero.Client` or `roomsaero.Client` method → `doJSON()` → the seats.aero or rooms.aero Partner API.

**Shared auth:** `rooms.aero` accepts exactly the same credentials as `seats.aero` — a Pro API key, or a Login with Seats.aero OAuth access token — sent in the `Partner-Authorization` header. `main.go` resolves one `seatsaero.Authorizer` and passes it to both clients; `roomsaero.Authorizer` is a structurally identical interface, so the same value satisfies it. The two APIs have **separate** 1,000 calls/day quotas.

**Auth flow:** `main.go` builds an OAuth `TokenSource` if `SEATS_OAUTH_*` creds + a saved token exist, then `ResolveAuthorizer` picks OAuth over the API key. `doJSON` calls `Authorizer.Apply` per request — `OAuthAuthorizer` lazily refreshes the access token via the `TokenSource`. `auth login` runs the one-time authorization-code consent flow and persists tokens through a `FileTokenStore`.

## Conventions

- **Optional parameters** use `*bool` or `*int` in the seatsaero param structs. The MCP tool layer reads them from `map[string]any` args via `boolPtrArg()` / `intPtrArg()` helpers.
- **Error handling:** The seatsaero layer returns Go errors (including `*APIError` for non-2xx). The mcptools layer converts the Go error to an MCP tool error result via `apiErrResult()` — the second return (Go error) is reserved for transport failures and is always `nil` after an API error.
- **Required MCP params** use `req.RequireString()`. Optional string params use `req.GetString(key, default)`. Optional bool/int use `boolPtrArg()` / `intPtrArg()`.
- **JSON serialization** in MCP results goes through `jsonResult()` which marshals with indentation.
- **Context** is threaded through `doJSON` via `http.NewRequestWithContext`. Don't drop it.
- **The server logs only to stderr.** stdout is the MCP transport (stdio). Use `fmt.Fprintln(os.Stderr, ...)` for logging.

## API key constraints

- Live search (`/live`) is **not available to Pro users** — requires a commercial agreement.
- 1,000 requests/day shared across all uses of a key.
- Pro keys are non-commercial use only.
- rooms.aero uses the **same key/token** but keeps its **own** 1,000 calls/day quota, separate from seats.aero. Only one hotel refresh is allowed per user every 24 hours (shared with the site).

## seats.aero API quick reference

- **Auth:** `Partner-Authorization` header with the API key, **or** the OAuth token as `Partner-Authorization: Bearer <token>`. (The plain `Authorization` header is only used by the OAuth2 `/userinfo` endpoint.)
- **OAuth2 base URL:** `https://seats.aero/oauth2` — `GET /consent` (response_type=code, client_id, redirect_uri, state, scope=openid), `POST /token` (JSON body; `authorization_code` + `refresh_token` grants), `GET /userinfo` (Bearer). Scope must be exactly `openid`. Redirect URIs are pre-registered.
- **Base URL:** `https://seats.aero/partnerapi`
- **Cabin codes:** `Y` = economy, `W` = premium economy, `J` = business, `F` = first.
- **Pagination:** prefer `cursor` over `skip` to avoid duplicates.
- **Live search IDs** are ephemeral — don't pass to `/trips/{id}`.
- **Times** are in local airport timezone, not UTC.

Full API docs: https://developers.seats.aero/reference/getting-started-p

## rooms.aero API quick reference

- **Base URL:** `https://rooms.aero/partnerapi`
- **Auth:** same `Partner-Authorization` header as seats.aero (Pro key, or OAuth token as `Bearer …`). Same credential, separate quota.
- **Endpoints (all GET except refresh):**
  - `GET /search` — hotels with award availability in an area + check-in date range, cheapest first.
  - `GET /availability` — one row per hotel + check-in date + stay length; `source` **or** `hotel_id` (≤50) required.
  - `GET /availability/{id}` — one availability with every room type and a booking link.
  - `GET /hotels` — hotel directory; there is no per-hotel endpoint, pass `id` to fetch one.
  - `POST /hotels/{id}/refresh` — queue a fresh rate check (one per user / 24h; `202` on success).
  - `GET /alerts` — list the account's alerts (read-only; requires a Pro user's key).
- **Sources (program keys):** `hilton`, `hyatt`, `ihg`, `marriott`, `choice`, `wyndham`, `iprefer`.
- **Dates:** `start_date`/`end_date` bound the **check-in date** (not check-out); `nights` is 1–5. Timestamps are RFC 3339 UTC.
- **Money:** points are whole numbers for the stay; cash is an integer in the minor unit (cents) of the named currency (`cash_currency_code` / `cash_cost_currency`); CPP is US cents per point.
- **Result envelope:** `{ "data": [], "count": N, "has_more": bool, "more_url": "…" }`. Paginate with `skip`/`take` (or follow `more_url`); dedupe by `id`.
- **Errors:** `{ "error": "code", "message": "…" }` with a stable code, e.g. `invalid_parameter`, `location_not_found`, `too_many_hotels`, `bad_partner_token`, `user_rate_limit_exceeded`, `refresh_limit_exceeded`.
- **Freshness:** Search/Get Availability only return upcoming check-in dates retrieved within the last 21 days; Refresh Hotel forces a re-check.

Full API docs: https://developers.seats.aero/reference/rooms-getting-started
