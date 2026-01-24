package folder

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Parent string
}

// NewCmdCreate creates the folder create command.
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new folder",
		Long: `Create a new folder (mailbox).

You can optionally specify a parent folder to create a nested folder.`,
		Example: `  # Create a top-level folder
  fm folder create "Work Projects"

  # Create a nested folder
  fm folder create "Q1 Reports" --parent abc123`,
		Args: cmdutil.ExactArgs(1, "folder name required\n\nUsage: fm folder create <name>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(f, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Parent, "parent", "", "Parent folder ID for nested folder")

	return cmd
}

func runCreate(f *cmdutil.Factory, opts *createOptions, name string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	folderID, err := client.CreateMailbox(name, opts.Parent)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Folder created: %s\n", folderID)
	return nil
}
