package auth

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/auth"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

// NewCmdStatus creates the auth status command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Display authentication status",
		Long:  `Display the current authentication status and token source.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(f)
		},
	}

	return cmd
}

func runStatus(f *cmdutil.Factory) error {
	out := f.IOStreams.Out

	// Check environment variable first
	envToken := os.Getenv("FASTMAIL_TOKEN")
	if envToken != "" {
		fmt.Fprintln(out, "api.fastmail.com")
		fmt.Fprintln(out, "  ✓ Authenticated via FASTMAIL_TOKEN environment variable")
		fmt.Fprintf(out, "  - Token: %s...%s\n", envToken[:4], envToken[len(envToken)-4:])

		// Validate token
		client := jmap.NewClient(envToken)
		session, err := client.GetSession()
		if err != nil {
			fmt.Fprintf(out, "  ✗ Token validation failed: %v\n", err)
			return cmdutil.SilentError
		}
		fmt.Fprintf(out, "  - Account ID: %s\n", session.AccountID)
		return nil
	}

	// Check keychain
	token, err := auth.GetTokenFromKeyring()
	if err != nil || token == "" {
		fmt.Fprintln(out, "api.fastmail.com")
		fmt.Fprintln(out, "  ✗ Not authenticated")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Run 'fm auth login' to authenticate.")
		return cmdutil.SilentError
	}

	fmt.Fprintln(out, "api.fastmail.com")
	fmt.Fprintln(out, "  ✓ Authenticated via system keychain")
	fmt.Fprintf(out, "  - Token: %s...%s\n", token[:4], token[len(token)-4:])

	// Validate token
	client := jmap.NewClient(token)
	session, err := client.GetSession()
	if err != nil {
		fmt.Fprintf(out, "  ✗ Token validation failed: %v\n", err)
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Run 'fm auth login' to re-authenticate.")
		return cmdutil.SilentError
	}
	fmt.Fprintf(out, "  - Account ID: %s\n", session.AccountID)

	return nil
}
