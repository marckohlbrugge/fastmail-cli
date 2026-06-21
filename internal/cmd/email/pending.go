package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdPending creates the email pending command.
func NewCmdPending(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending",
		Short: "List pending email submissions",
		Long: `List emails that are queued but not yet sent.

These are emails sent with --hold that can still be canceled.`,
		Example: `  fm email pending`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPending(f)
		},
	}

	return cmd
}

func runPending(f *cmdutil.Factory) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	submissions, err := client.GetPendingSubmissions()
	if err != nil {
		return err
	}

	if len(submissions) == 0 {
		fmt.Fprintln(f.IOStreams.Out, "No pending submissions.")
		return nil
	}

	for _, s := range submissions {
		sendAt := "unknown"
		if s.SendAt != nil {
			sendAt = s.SendAt.Local().Format("15:04:05")
		}
		fmt.Fprintf(f.IOStreams.Out, "%s  email:%s  sends:%s\n", s.ID, s.EmailID, sendAt)
	}

	return nil
}
