package folder

import (
	"encoding/json"
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type listOptions struct {
	JSON bool
}

// NewCmdList creates the folder list command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all folders",
		Long: `List all mailboxes (folders) in your account.

This is an alias for 'fm folders'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	mailboxes, err := client.GetMailboxes()
	if err != nil {
		return err
	}

	if opts.JSON {
		encoder := json.NewEncoder(f.IOStreams.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(mailboxes)
	}

	return outputHuman(f, mailboxes)
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
