package auth

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	// KeyringService is the service name used in the system keychain
	KeyringService = "fm-cli"
	// KeyringUser is the user/account name in the keychain
	KeyringUser = "fastmail-token"
)

// TokenSource provides authentication tokens for the JMAP API.
type TokenSource struct {
	envToken string
}

// NewTokenSource creates a new TokenSource.
func NewTokenSource() *TokenSource {
	return &TokenSource{
		envToken: os.Getenv("FASTMAIL_TOKEN"),
	}
}

// GetToken retrieves the API token from environment or system keychain.
// Priority: FASTMAIL_TOKEN env var > system keychain
func (ts *TokenSource) GetToken() (string, error) {
	// 1. Environment variable takes precedence
	if ts.envToken != "" {
		return ts.envToken, nil
	}

	// 2. Try system keychain
	token, err := GetTokenFromKeyring()
	if err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("not authenticated.\n\n" +
		"Run 'fm auth login' to authenticate, or set FASTMAIL_TOKEN environment variable.")
}

// GetTokenFromKeyring retrieves the token from the system keychain.
func GetTokenFromKeyring() (string, error) {
	return keyring.Get(KeyringService, KeyringUser)
}

// SetTokenInKeyring stores the token in the system keychain.
func SetTokenInKeyring(token string) error {
	return keyring.Set(KeyringService, KeyringUser, token)
}

// DeleteTokenFromKeyring removes the token from the system keychain.
func DeleteTokenFromKeyring() error {
	return keyring.Delete(KeyringService, KeyringUser)
}

// IsAuthenticated returns true if a token is available.
func (ts *TokenSource) IsAuthenticated() bool {
	token, _ := ts.GetToken()
	return token != ""
}
