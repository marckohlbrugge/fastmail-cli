package folders

import (
	"encoding/json"
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type foldersOptions struct {
	JSON bool
}

// NewCmdFolders creates the folders command.
func NewCmdFolders(f *cmdutil.Factory) *cobra.Command {
	opts := &foldersOptions{}

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List mailboxes",
		Long: `List all mailboxes (folders) in your account.

Displays folder ID, name, role (if any), and unread count.`,
		Example: `  # List all folders
  fm folders

  # Output as JSON
  fm folders --json`,
		GroupID: "core",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFolders(f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runFolders(f *cmdutil.Factory, opts *foldersOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	mailboxes, err := client.GetMailboxes()
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(f, mailboxes)
	}

	return outputHuman(f, mailboxes)
}

func outputJSON(f *cmdutil.Factory, mailboxes []jmap.Mailbox) error {
	encoder := json.NewEncoder(f.IOStreams.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(mailboxes)
}

func outputHuman(f *cmdutil.Factory, mailboxes []jmap.Mailbox) error {
	out := f.IOStreams.Out

	if len(mailboxes) == 0 {
		fmt.Fprintln(out, "No mailboxes found.")
		return nil
	}

	for _, mb := range mailboxes {
		role := ""
		if mb.Role != "" {
			role = fmt.Sprintf(" (%s)", mb.Role)
		}

		unread := ""
		if mb.UnreadEmails > 0 {
			unread = fmt.Sprintf(" [%d unread]", mb.UnreadEmails)
		}

		fmt.Fprintf(out, "%-20s  %s%s%s\n", mb.ID, mb.Name, role, unread)
	}

	return nil
}
