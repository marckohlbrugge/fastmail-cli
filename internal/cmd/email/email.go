package email

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdEmail creates the email parent command.
func NewCmdEmail(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email <command>",
		Short: "Manage emails",
		Long:  "Read, archive, move, and delete emails.",
		Example: `  $ fm email read M1234567890
  $ fm email thread M1234567890
  $ fm email archive M1234567890
  $ fm email move M1234567890 inbox`,
		GroupID: "email",
	}

	cmd.AddCommand(NewCmdRead(f))
	cmd.AddCommand(NewCmdThread(f))
	cmd.AddCommand(NewCmdArchive(f))
	cmd.AddCommand(NewCmdMove(f))
	cmd.AddCommand(NewCmdDelete(f))

	return cmd
}
