package folder

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdFolder creates the folder parent command.
func NewCmdFolder(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folder <command>",
		Short: "Manage folders",
		Long:  "Create, rename, and delete folders (mailboxes).",
		Example: `  $ fm folder list
  $ fm folder create "Work Projects"
  $ fm folder rename abc123 "New Name"`,
		GroupID: "folder",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdRename(f))

	return cmd
}
