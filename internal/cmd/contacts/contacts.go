package contacts

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdContacts creates the contacts parent command.
func NewCmdContacts(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts <command>",
		Short: "Manage contacts",
		Long:  "List, search, create, and delete contacts.",
		Example: `  $ fm contacts list
  $ fm contacts show abc123
  $ fm contacts search "John"
  $ fm contacts create --name "John Doe" --email john@example.com`,
		GroupID: "contacts",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdShow(f))
	cmd.AddCommand(NewCmdSearch(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdDelete(f))

	return cmd
}
