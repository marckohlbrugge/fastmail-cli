package folder

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdRename creates the folder rename command.
func NewCmdRename(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <folder-id> <new-name>",
		Short: "Rename a folder",
		Long: `Rename an existing folder.

Use 'fm folders' or 'fm folder list' to find the folder ID.`,
		Example: `  # Rename a folder
  fm folder rename abc123 "New Name"`,
		Args: cmdutil.ExactArgs(2, "folder ID and new name required\n\nUsage: fm folder rename <folder-id> <new-name>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRename(f, args[0], args[1])
		},
	}

	return cmd
}

func runRename(f *cmdutil.Factory, folderID, newName string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.RenameMailbox(folderID, newName); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Folder renamed.")
	return nil
}
