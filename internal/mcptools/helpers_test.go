package mcptools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func callTool(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	msg, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := s.HandleMessage(context.Background(), msg)

	switch resp := raw.(type) {
	case mcp.JSONRPCResponse:
		b, err := json.Marshal(resp.Result)
		if err != nil {
			t.Fatalf("marshal result: %v", err)
		}
		var toolResult mcp.CallToolResult
		if err := json.Unmarshal(b, &toolResult); err != nil {
			t.Fatalf("unmarshal tool result: %v", err)
		}
		return &toolResult
	case mcp.JSONRPCError:
		b, _ := json.Marshal(resp.Error)
		t.Fatalf("RPC error: %s", string(b))
		return nil
	default:
		b, _ := json.Marshal(raw)
		t.Fatalf("unexpected response type: %s", string(b))
		return nil
	}
}

func checkToolSuccess(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Fatal("expected success result, got error")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
}

func checkToolError(t *testing.T, result *mcp.CallToolResult, substr string) {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Fatal("expected error result, got success")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	found := false
	for i := 0; i <= len(tc.Text)-len(substr); i++ {
		if tc.Text[i:i+len(substr)] == substr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error text %q does not contain %q", tc.Text, substr)
	}
}
