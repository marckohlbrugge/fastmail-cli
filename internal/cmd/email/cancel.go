package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdCancel creates the email cancel command.
func NewCmdCancel(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <submission-id>",
		Short: "Cancel a pending email (undo send)",
		Long: `Cancel a pending email submission before it's sent.

This only works for emails sent with --hold that haven't been
released yet. The submission ID is shown when using --hold.`,
		Example: `  # Cancel a pending send
  fm email cancel ES1234567890`,
		Args: cmdutil.ExactArgs(1, "submission ID required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(f, args[0])
		},
	}

	return cmd
}

func runCancel(f *cmdutil.Factory, submissionID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.CancelSubmission(submissionID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Send canceled.")
	return nil
}
