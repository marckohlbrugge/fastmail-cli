package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

// NewCmdMove creates the email move command.
func NewCmdMove(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <email-id> <folder>",
		Short: "Move an email to a folder",
		Long: `Move an email to a different folder.

The folder can be specified by ID, name, or role (inbox, archive, trash, etc.).`,
		Example: `  # Move by folder ID
  fm email move M1234567890 abc123def456

  # Move by folder name
  fm email move M1234567890 "Work Projects"

  # Move by role
  fm email move M1234567890 inbox`,
		Args: cmdutil.ExactArgs(2, "email ID and folder required\n\nUsage: fm email move <email-id> <folder>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(f, args[0], args[1])
		},
	}

	return cmd
}

func runMove(f *cmdutil.Factory, emailID, folderRef string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Resolve folder
	mailbox, err := resolveMailbox(client, folderRef)
	if err != nil {
		return fmt.Errorf("folder not found: %s", folderRef)
	}

	if err := client.MoveEmail(emailID, mailbox.ID); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Moved to %s.\n", mailbox.Name)
	return nil
}

func resolveMailbox(client *jmap.Client, folderRef string) (*jmap.Mailbox, error) {
	// Try by ID first
	mailbox, err := client.GetMailboxByID(folderRef)
	if err == nil {
		return mailbox, nil
	}

	// Try by name
	mailbox, err = client.GetMailboxByName(folderRef)
	if err == nil {
		return mailbox, nil
	}

	// Try by role
	return client.GetMailboxByRole(folderRef)
}
