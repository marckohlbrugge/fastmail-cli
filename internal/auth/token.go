package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// TokenCacheTTL is how long a cached token is valid
	TokenCacheTTL = 1 * time.Hour

	// OnePasswordPath is the 1Password CLI path for the Fastmail token
	OnePasswordPath = "op://Services/Fastmail/credential"
)

// TokenSource provides authentication tokens for the JMAP API.
type TokenSource struct {
	envToken   string
	cacheFile  string
	cacheTTL   time.Duration
	opPath     string
}

// NewTokenSource creates a new TokenSource.
func NewTokenSource() *TokenSource {
	cacheDir := os.TempDir()
	return &TokenSource{
		envToken:  os.Getenv("FASTMAIL_TOKEN"),
		cacheFile: filepath.Join(cacheDir, ".fm-token-cache"),
		cacheTTL:  TokenCacheTTL,
		opPath:    OnePasswordPath,
	}
}

// GetToken retrieves the API token from environment, cache, or 1Password.
// Priority: FASTMAIL_TOKEN env var > cached token > 1Password
func (ts *TokenSource) GetToken() (string, error) {
	// 1. Environment variable takes precedence
	if ts.envToken != "" {
		return ts.envToken, nil
	}

	// 2. Check cache
	if token := ts.getCachedToken(); token != "" {
		return token, nil
	}

	// 3. Fetch from 1Password
	token, err := ts.fetchFromOnePassword()
	if err != nil {
		return "", fmt.Errorf("no FastMail API token found.\n\n" +
			"Set FASTMAIL_TOKEN environment variable, or store in 1Password at:\n" +
			"  op://Services/Fastmail/credential")
	}

	// Cache the token for future use
	ts.cacheToken(token)

	return token, nil
}

// getCachedToken returns the cached token if it exists and hasn't expired.
func (ts *TokenSource) getCachedToken() string {
	info, err := os.Stat(ts.cacheFile)
	if err != nil {
		return ""
	}

	// Check if cache has expired
	if time.Since(info.ModTime()) > ts.cacheTTL {
		return ""
	}

	data, err := os.ReadFile(ts.cacheFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// cacheToken stores the token in the cache file.
func (ts *TokenSource) cacheToken(token string) {
	// Write with restricted permissions (owner read/write only)
	_ = os.WriteFile(ts.cacheFile, []byte(token), 0600)
}

// fetchFromOnePassword retrieves the token from 1Password CLI.
func (ts *TokenSource) fetchFromOnePassword() (string, error) {
	cmd := exec.Command("op", "read", ts.opPath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("empty token from 1Password")
	}

	return token, nil
}

// ClearCache removes the cached token.
func (ts *TokenSource) ClearCache() error {
	return os.Remove(ts.cacheFile)
}
