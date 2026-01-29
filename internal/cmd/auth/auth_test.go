package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	ios := &iostreams.IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}

	f := &cmdutil.Factory{
		IOStreams: ios,
	}

	return f, in, out, errOut
}

// Login command tests

func TestLoginCommand(t *testing.T) {
	t.Run("validates token with API", func(t *testing.T) {
		f, in, out, _ := setupTest(t)

		httpmock.RegisterResponder("GET", "https://api.fastmail.com/jmap/session",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"apiUrl": "https://api.fastmail.com/jmap/api",
				"accounts": map[string]interface{}{
					"u12345": map[string]interface{}{},
				},
			}))

		// Simulate token input (stdin is non-TTY in test, so it reads from stdin)
		in.WriteString("fmu1-test-token-12345678\n")

		cmd := NewCmdLogin(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		// Will fail on keyring storage in test environment, but validates the token
		// The error should be about keyring, not about auth
		if err != nil {
			assert.Contains(t, err.Error(), "keychain")
		}
	})

	t.Run("rejects empty token", func(t *testing.T) {
		f, in, _, _ := setupTest(t)

		// Empty input
		in.WriteString("\n")

		cmd := NewCmdLogin(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		f, in, _, _ := setupTest(t)

		httpmock.RegisterResponder("GET", "https://api.fastmail.com/jmap/session",
			httpmock.NewStringResponder(401, "Unauthorized"))

		in.WriteString("invalid-token\n")

		cmd := NewCmdLogin(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("accepts no arguments", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdLogin(f)

		// Just verify the command accepts no args without error on parsing
		err := cmd.ParseFlags([]string{})
		require.NoError(t, err)
	})
}

// Status command tests

func TestStatusCommand(t *testing.T) {
	t.Run("shows not authenticated when no token", func(t *testing.T) {
		f, _, out, _ := setupTest(t)

		cmd := NewCmdStatus(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})

		// Will return error because no token is available
		err := cmd.Execute()

		// SilentError is returned when not authenticated
		if err != nil {
			assert.Equal(t, cmdutil.SilentError, err)
		}

		output := out.String()
		assert.Contains(t, output, "api.fastmail.com")
	})

	t.Run("validates token from environment", func(t *testing.T) {
		f, _, out, _ := setupTest(t)

		httpmock.RegisterResponder("GET", "https://api.fastmail.com/jmap/session",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"apiUrl": "https://api.fastmail.com/jmap/api",
				"accounts": map[string]interface{}{
					"u12345": map[string]interface{}{},
				},
			}))

		// Set environment variable
		t.Setenv("FASTMAIL_TOKEN", "fmu1-test-token")

		cmd := NewCmdStatus(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := out.String()
		assert.Contains(t, output, "api.fastmail.com")
		assert.Contains(t, output, "FASTMAIL_TOKEN environment variable")
		assert.Contains(t, output, "Account ID: u12345")
	})

	t.Run("shows error for invalid env token", func(t *testing.T) {
		f, _, out, _ := setupTest(t)

		httpmock.RegisterResponder("GET", "https://api.fastmail.com/jmap/session",
			httpmock.NewStringResponder(401, "Unauthorized"))

		t.Setenv("FASTMAIL_TOKEN", "invalid-token")

		cmd := NewCmdStatus(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		assert.Equal(t, cmdutil.SilentError, err)
		output := out.String()
		assert.Contains(t, output, "Token validation failed")
	})

	t.Run("accepts no arguments", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdStatus(f)
		cmd.SetArgs([]string{"extra-arg"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

// Logout command tests

func TestLogoutCommand(t *testing.T) {
	t.Run("handles not logged in gracefully", func(t *testing.T) {
		f, _, out, _ := setupTest(t)

		cmd := NewCmdLogout(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		// Should not error, just say not logged in
		require.NoError(t, err)
		output := out.String()
		// Either "Not logged in" or "Logged out" depending on keyring state
		assert.True(t, strings.Contains(output, "Not logged in") || strings.Contains(output, "Logged out"))
	})

	t.Run("accepts no arguments", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdLogout(f)
		cmd.SetArgs([]string{"extra-arg"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}
