package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenSource(t *testing.T) {
	t.Run("captures env token at creation", func(t *testing.T) {
		t.Setenv("FASTMAIL_TOKEN", "env-token-123")

		ts := NewTokenSource()

		assert.Equal(t, "env-token-123", ts.envToken)
	})

	t.Run("empty when env not set", func(t *testing.T) {
		// Ensure env is not set
		t.Setenv("FASTMAIL_TOKEN", "")

		ts := NewTokenSource()

		assert.Empty(t, ts.envToken)
	})
}

func TestGetToken(t *testing.T) {
	t.Run("returns env token when set", func(t *testing.T) {
		t.Setenv("FASTMAIL_TOKEN", "fmu1-env-token")

		ts := NewTokenSource()
		token, err := ts.GetToken()

		require.NoError(t, err)
		assert.Equal(t, "fmu1-env-token", token)
	})

	t.Run("env token takes priority", func(t *testing.T) {
		// Even if keyring has a token, env should win
		t.Setenv("FASTMAIL_TOKEN", "env-priority-token")

		ts := NewTokenSource()
		token, err := ts.GetToken()

		require.NoError(t, err)
		assert.Equal(t, "env-priority-token", token)
	})

	t.Run("returns error when no token available", func(t *testing.T) {
		t.Setenv("FASTMAIL_TOKEN", "")

		ts := NewTokenSource()
		_, err := ts.GetToken()

		// In CI/test environment, keyring likely fails, so we expect an error
		// The error message should guide the user
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
		assert.Contains(t, err.Error(), "fm auth login")
	})
}

func TestIsAuthenticated(t *testing.T) {
	t.Run("true when env token set", func(t *testing.T) {
		t.Setenv("FASTMAIL_TOKEN", "auth-token")

		ts := NewTokenSource()

		assert.True(t, ts.IsAuthenticated())
	})

	t.Run("false when no token", func(t *testing.T) {
		t.Setenv("FASTMAIL_TOKEN", "")

		ts := NewTokenSource()

		// In test environment without keyring, should be false
		assert.False(t, ts.IsAuthenticated())
	})
}

func TestKeyringConstants(t *testing.T) {
	// Verify the keyring constants are set correctly
	assert.Equal(t, "fm-cli", KeyringService)
	assert.Equal(t, "fastmail-token", KeyringUser)

	// Ensure they don't contain spaces or special chars that might cause issues
	assert.False(t, strings.ContainsAny(KeyringService, " \t\n"))
	assert.False(t, strings.ContainsAny(KeyringUser, " \t\n"))
}
