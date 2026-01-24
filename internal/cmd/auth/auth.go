package auth

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAuth creates the auth parent command.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Authenticate fm with Fastmail",
		Long: `Authenticate fm with your Fastmail account.

The token is stored securely in your system's credential store
(macOS Keychain, Windows Credential Manager, or Linux Secret Service).

Alternatively, set the FASTMAIL_TOKEN environment variable.`,
		Example: `  $ fm auth login
  $ fm auth status
  $ fm auth logout`,
		GroupID: "auth",
	}

	cmd.AddCommand(NewCmdLogin(f))
	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdLogout(f))

	return cmd
}
