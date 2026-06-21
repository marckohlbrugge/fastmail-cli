package attachment

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAttachment creates the attachment parent command.
func NewCmdAttachment(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment <command>",
		Short: "Manage email attachments",
		Long:  "List and download email attachments.",
		Example: `  $ fm attachment list M1234567890
  $ fm attachment download M1234567890 B1234567890
  $ fm attachment download M1234567890 B1234567890 --output report.pdf`,
		GroupID: "email",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdDownload(f))

	return cmd
}
