package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdArchive creates the email archive command.
func NewCmdArchive(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <email-id>...",
		Short: "Move emails to archive",
		Long: `Move one or more emails to the Archive folder.

This is a reversible action - emails can be moved back from Archive.`,
		Example: `  # Archive a single email
  fm email archive M1234567890

  # Archive multiple emails
  fm email archive M1234567890 M0987654321`,
		Args: cmdutil.MinimumArgs(1, "at least one email ID required\n\nUsage: fm email archive <email-id>..."),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive(f, args)
		},
	}

	return cmd
}

func runArchive(f *cmdutil.Factory, emailIDs []string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	out := f.IOStreams.Out

	if len(emailIDs) == 1 {
		if err := client.ArchiveEmail(emailIDs[0]); err != nil {
			return err
		}
		fmt.Fprintln(out, "Moved to Archive.")
		return nil
	}

	// Bulk archive
	archived, failed, err := client.ArchiveEmails(emailIDs)
	if err != nil {
		return err
	}

	if len(failed) > 0 {
		fmt.Fprintf(out, "Archived %d emails. Failed: %d\n", archived, len(failed))
		for _, id := range failed {
			fmt.Fprintf(f.IOStreams.ErrOut, "  Failed: %s\n", id)
		}
		return nil
	}

	fmt.Fprintf(out, "Archived %d emails.\n", archived)
	return nil
}
