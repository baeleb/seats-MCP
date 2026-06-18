# CLAUDE.md

Guidance for coding agents working with this repository.

## Build

```sh
go build -o bin/seats-mcp .
```

Requires Go 1.26+.

## Test

There is no test suite yet. When adding tests:

```sh
go test ./...
```

Use the standard `testing` package. Mock HTTP calls with `httptest.NewServer`. Tests live alongside source as `*_test.go`.

## Architecture

```
main.go                        Entry point — wires up auth → seats.aero client → MCP server (default), or dispatches `auth` subcommand
auth.go                        `auth login|status|logout` CLI: OAuth2 consent flow (loopback callback / paste fallback), env→config, token store
internal/
  seatsaero/                   Typed API client for seats.aero Partner API
    client.go                  HTTP client core: doJSON() applies an Authorizer, handles errors, JSON encoding
    auth.go                    Authorizer interface; APIKeyAuthorizer (Partner-Authorization) + OAuthAuthorizer (Bearer); ResolveAuthorizer (prefers OAuth)
    oauth.go                   OAuthConfig, Token, TokenSource (caching/refresh/code-exchange), AuthCodeURL, UserInfo
    oauthstore.go              TokenStore interface; FileTokenStore (0600, atomic) + MemoryTokenStore; DefaultTokenPath
    types.go                   All data types: Availability, Trip, Segment, SearchResponse, etc.
    search.go                  GET /search — cached multi-program search
    availability.go            GET /availability — single-program bulk fetch
    trips.go                   GET /trips/{id} — flight-level detail from an Availability ID
    live.go                    POST /live — real-time search
    errors.go                  APIError type for non-2xx responses
  mcptools/                    MCP tool layer — wraps seats.aero client methods as MCP tools
    register.go                Register() wires all tools; jsonResult, apiErrResult, arg helpers
    search.go                  "search_award_availability" tool
    availability.go            "bulk_availability" tool
    trips.go                   "get_trips" tool
    live.go                    "live_search" tool
```

**Flow:** `main.go` → `mcptools.Register(s, client)` → each tool handler calls `seatsaero.Client` methods → `doJSON()` → seats.aero Partner API.

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

## seats.aero API quick reference

- **Auth:** `Partner-Authorization` header with the API key, **or** OAuth2 `Authorization: Bearer <token>`.
- **OAuth2 base URL:** `https://seats.aero/oauth2` — `GET /consent` (response_type=code, client_id, redirect_uri, state, scope=openid), `POST /token` (JSON body; `authorization_code` + `refresh_token` grants), `GET /userinfo` (Bearer). Scope must be exactly `openid`. Redirect URIs are pre-registered.
- **Base URL:** `https://seats.aero/partnerapi`
- **Cabin codes:** `Y` = economy, `W` = premium economy, `J` = business, `F` = first.
- **Pagination:** prefer `cursor` over `skip` to avoid duplicates.
- **Live search IDs** are ephemeral — don't pass to `/trips/{id}`.
- **Times** are in local airport timezone, not UTC.

Full API docs: https://developers.seats.aero/reference/getting-started-p
