package vacation

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdVacation creates the vacation parent command.
func NewCmdVacation(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vacation <command>",
		Short: "Manage vacation auto-responder",
		Long:  "Get, enable, or disable the vacation auto-responder.",
		Example: `  $ fm vacation status
  $ fm vacation on --subject "Out of office" --body "I'm away until Monday."
  $ fm vacation off`,
	}

	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdOn(f))
	cmd.AddCommand(NewCmdOff(f))

	return cmd
}
