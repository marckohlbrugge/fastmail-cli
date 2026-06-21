package mask

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdMask creates the mask parent command.
func NewCmdMask(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mask <command>",
		Short: "Manage masked email addresses",
		Long:  "Create, list, and manage Fastmail masked email addresses.",
		Example: `  $ fm mask list
  $ fm mask create --domain example.com --desc "Shopping site"
  $ fm mask disable abc123
  $ fm mask enable abc123`,
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdEnable(f))
	cmd.AddCommand(NewCmdDisable(f))
	cmd.AddCommand(NewCmdDelete(f))

	return cmd
}
