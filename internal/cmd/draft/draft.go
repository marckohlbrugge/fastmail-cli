package draft

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdDraft creates the draft parent command.
func NewCmdDraft(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "draft <command>",
		Short: "Manage drafts",
		Long:  "Create, edit, and send draft emails.",
		Example: `  $ fm draft new --to bob@example.com --subject "Hello"
  $ fm draft reply M1234567890 --body "Thanks!"
  $ fm draft forward M1234567890 --to alice@example.com
  $ fm draft send M1234567890`,
		GroupID: "draft",
	}

	cmd.AddCommand(NewCmdNew(f))
	cmd.AddCommand(NewCmdEdit(f))
	cmd.AddCommand(NewCmdDraftDelete(f))
	cmd.AddCommand(NewCmdReply(f))
	cmd.AddCommand(NewCmdForward(f))
	cmd.AddCommand(NewCmdSend(f))

	return cmd
}
