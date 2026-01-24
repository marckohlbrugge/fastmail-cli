package auth

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/auth"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdLogout creates the auth logout command.
func NewCmdLogout(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove authentication",
		Long: `Remove the stored authentication token from your system keychain.

Note: This does not revoke the token on Fastmail's side. To fully revoke
access, delete the API token in Fastmail Settings → Privacy & Security → Integrations.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(f)
		},
	}

	return cmd
}

func runLogout(f *cmdutil.Factory) error {
	out := f.IOStreams.Out

	err := auth.DeleteTokenFromKeyring()
	if err != nil {
		// Check if it's just "not found" which is fine
		fmt.Fprintln(out, "Not logged in.")
		return nil
	}

	fmt.Fprintln(out, "Logged out of Fastmail.")
	fmt.Fprintln(out, "Token removed from system keychain.")

	return nil
}
