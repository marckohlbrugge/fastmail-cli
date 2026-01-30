package identity

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdIdentity creates the identity command group.
func NewCmdIdentity(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "identity <command>",
		Short:   "Manage identities",
		Long:    "View sender identities (email addresses you can send from).",
		GroupID: "identity",
		Example: `  $ fm identity list`,
	}

	cmd.AddCommand(NewCmdList(f))

	return cmd
}
