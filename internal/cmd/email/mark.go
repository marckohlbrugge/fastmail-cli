package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdMark creates the email mark command.
func NewCmdMark(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark <email-id> <read|unread>",
		Short: "Mark an email as read or unread",
		Long:  `Mark an email as read or unread by setting the $seen keyword.`,
		Example: `  # Mark as read
  fm email mark M1234567890 read

  # Mark as unread
  fm email mark M1234567890 unread`,
		Args: cmdutil.ExactArgs(2, "email ID and state (read/unread) required\n\nUsage: fm email mark <email-id> <read|unread>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMark(f, args[0], args[1])
		},
	}

	return cmd
}

func runMark(f *cmdutil.Factory, emailID, state string) error {
	var read bool
	switch state {
	case "read":
		read = true
	case "unread":
		read = false
	default:
		return fmt.Errorf("invalid state %q: must be 'read' or 'unread'", state)
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.MarkRead(emailID, read); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Marked as %s.\n", state)
	return nil
}
