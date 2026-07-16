package seatsaero

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileTokenStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "token.json")
	store := FileTokenStore{Path: path}

	want := &Token{
		AccessToken:  "seats:ota:abc",
		RefreshToken: "seats:otr:def",
		TokenType:    "Bearer",
		Scope:        "openid",
		Expiry:       time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if !got.Expiry.Equal(want.Expiry) {
		t.Errorf("expiry got %v, want %v", got.Expiry, want.Expiry)
	}
}

func TestFileTokenStorePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token.json")
	store := FileTokenStore{Path: path}
	if err := store.Save(&Token{AccessToken: "x"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("perm = %o, want 600", perm)
	}
}

func TestFileTokenStoreLoadMissing(t *testing.T) {
	store := FileTokenStore{Path: filepath.Join(t.TempDir(), "absent.json")}
	tok, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if tok != nil {
		t.Errorf("expected nil token, got %+v", tok)
	}
}

func TestFileTokenStoreLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	store := FileTokenStore{Path: path}
	tok, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if tok != nil {
		t.Errorf("expected nil token for empty file, got %+v", tok)
	}
}

func TestFileTokenStoreLoadCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := FileTokenStore{Path: path}
	if _, err := store.Load(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestFileTokenStoreDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token.json")
	store := FileTokenStore{Path: path}
	if err := store.Save(&Token{AccessToken: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still exists after delete")
	}
	// Deleting an absent file is not an error.
	if err := store.Delete(); err != nil {
		t.Errorf("delete absent: %v", err)
	}
}

func TestMemoryTokenStore(t *testing.T) {
	var store MemoryTokenStore
	tok, err := store.Load()
	if err != nil || tok != nil {
		t.Fatalf("empty load = %+v, %v", tok, err)
	}
	if err := store.Save(&Token{AccessToken: "y"}); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Load()
	if got.AccessToken != "y" {
		t.Errorf("got %+v", got)
	}
}

func TestDefaultTokenPath(t *testing.T) {
	p, err := DefaultTokenPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}
	if filepath.Base(p) != "token.json" {
		t.Errorf("base = %q, want token.json", filepath.Base(p))
	}
	if filepath.Base(filepath.Dir(p)) != "seats-mcp" {
		t.Errorf("dir = %q, want seats-mcp", filepath.Base(filepath.Dir(p)))
	}
}
