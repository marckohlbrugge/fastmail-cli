package auth

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/auth"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	Token string
}

// NewCmdLogin creates the auth login command.
func NewCmdLogin(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Fastmail",
		Long: `Authenticate with your Fastmail account using an API token.

To create an API token:
  1. Go to Fastmail Settings → Privacy & Security → Integrations
  2. Click "New API Token" (or "API Tokens")
  3. Give it a name and select permissions (Mail Read/Write recommended)
  4. Copy the generated token

The token will be stored securely in your system's credential store.`,
		Example: `  # Interactive login (prompts for token)
  $ fm auth login

  # Login with token from stdin
  $ echo "fmu1-xxx" | fm auth login --with-token

  # Login with token from file
  $ fm auth login --with-token < token.txt`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Token, "with-token", "", "Read token from standard input")

	return cmd
}

func runLogin(f *cmdutil.Factory, opts *loginOptions) error {
	out := f.IOStreams.Out
	errOut := f.IOStreams.ErrOut

	var token string

	if opts.Token != "" || !f.IOStreams.IsStdinTTY() {
		// Read token from stdin
		scanner := bufio.NewScanner(f.IOStreams.In)
		if scanner.Scan() {
			token = strings.TrimSpace(scanner.Text())
		}
		if token == "" {
			return cmdutil.FlagErrorf("token cannot be empty")
		}
	} else {
		// Interactive prompt
		fmt.Fprintln(out, "To create an API token, visit:")
		fmt.Fprintln(out, "  Fastmail Settings → Privacy & Security → Integrations → New API Token")
		fmt.Fprintln(out)
		fmt.Fprint(errOut, "Paste your API token: ")

		scanner := bufio.NewScanner(f.IOStreams.In)
		if scanner.Scan() {
			token = strings.TrimSpace(scanner.Text())
		}
		if token == "" {
			return cmdutil.FlagErrorf("token cannot be empty")
		}
	}

	// Validate token by making a test request
	fmt.Fprintln(errOut, "Validating token...")
	client := jmap.NewClient(token)
	session, err := client.GetSession()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store token in keychain
	if err := auth.SetTokenInKeyring(token); err != nil {
		return fmt.Errorf("failed to store token in keychain: %w", err)
	}

	fmt.Fprintf(out, "✓ Logged in to Fastmail (account: %s)\n", session.AccountID)
	fmt.Fprintln(out, "Token stored in system keychain.")

	return nil
}
