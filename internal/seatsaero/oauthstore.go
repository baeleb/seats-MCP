package seatsaero

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TokenStore persists an OAuth token set between runs of the server.
type TokenStore interface {
	// Load returns the stored token, or (nil, nil) if none has been saved.
	Load() (*Token, error)
	// Save persists tok, overwriting any prior token.
	Save(tok *Token) error
}

// DefaultTokenPath returns the path used to persist OAuth tokens when no
// explicit path is configured: <user-config-dir>/seats-mcp/token.json.
func DefaultTokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(dir, "seats-mcp", "token.json"), nil
}

// FileTokenStore persists the token as JSON at Path with 0600 permissions.
type FileTokenStore struct {
	Path string
}

func (s FileTokenStore) Load() (*Token, error) {
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read token file: %w", err)
	}
	if len(b) == 0 {
		return nil, nil
	}
	var tok Token
	if err := json.Unmarshal(b, &tok); err != nil {
		return nil, fmt.Errorf("parse token file %s: %w", s.Path, err)
	}
	return &tok, nil
}

func (s FileTokenStore) Save(tok *Token) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	// Write to a temp file then rename for an atomic, never-partial update.
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return fmt.Errorf("finalize token file: %w", err)
	}
	return nil
}

// Delete removes the token file. A missing file is not an error.
func (s FileTokenStore) Delete() error {
	if err := os.Remove(s.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete token file: %w", err)
	}
	return nil
}

// MemoryTokenStore is an in-memory TokenStore for tests.
type MemoryTokenStore struct {
	mu  sync.Mutex
	tok *Token
}

func (s *MemoryTokenStore) Load() (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tok, nil
}

func (s *MemoryTokenStore) Save(tok *Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tok = tok
	return nil
}
