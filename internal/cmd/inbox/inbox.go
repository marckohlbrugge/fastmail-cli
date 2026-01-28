package inbox

import (
	"encoding/json"
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type inboxOptions struct {
	Limit      int
	JSONFields []string
}

// NewCmdInbox creates the inbox command.
func NewCmdInbox(f *cmdutil.Factory) *cobra.Command {
	opts := &inboxOptions{}

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "List recent inbox emails",
		Long: `List recent emails from your inbox.

By default displays email ID, date, sender, and subject.
Use --json with field names for machine-readable output.`,
		Example: `  # List recent inbox emails
  fm inbox

  # List last 10 emails
  fm inbox --limit 10

  # Output as JSON with specific fields
  fm inbox --json id,subject,from

  # Output all available JSON fields
  fm inbox --json id,threadId,subject,from,to,cc,date,preview,unread,attachment`,
		GroupID: "core",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInbox(f, opts)
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of emails to show (max 50)")
	cmd.Flags().StringSliceVar(&opts.JSONFields, "json", nil, "Output JSON with specified `fields` (id,threadId,subject,from,to,cc,date,preview,unread,attachment)")

	return cmd
}

func runInbox(f *cmdutil.Factory, opts *inboxOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Validate JSON fields if provided
	if opts.JSONFields != nil {
		if err := cmdutil.ValidateFields(opts.JSONFields); err != nil {
			return err
		}
	}

	// Get inbox mailbox
	inbox, err := client.GetMailboxByRole("inbox")
	if err != nil {
		return fmt.Errorf("could not find inbox: %w", err)
	}

	// Fetch recent emails
	emails, err := client.GetRecentEmails(inbox.ID, opts.Limit)
	if err != nil {
		return err
	}

	if opts.JSONFields != nil {
		return outputJSON(f, emails, opts.JSONFields)
	}

	return outputHuman(f, emails, cmdutil.DefaultEmailFields)
}

func outputJSON(f *cmdutil.Factory, emails []jmap.Email, fields []string) error {
	output := make([]map[string]interface{}, len(emails))

	for i, e := range emails {
		row := make(map[string]interface{})
		for _, field := range fields {
			switch field {
			case "id":
				row["id"] = e.ID
			case "threadId":
				row["threadId"] = e.ThreadID
			case "subject":
				row["subject"] = e.Subject
			case "from":
				row["from"] = e.From
			case "to":
				row["to"] = e.To
			case "cc":
				row["cc"] = e.CC
			case "date":
				row["receivedAt"] = e.ReceivedAt
			case "preview":
				row["preview"] = e.Preview
			case "unread":
				row["isUnread"] = e.IsUnread()
			case "attachment":
				row["hasAttachment"] = e.HasAttachment
			}
		}
		output[i] = row
	}

	encoder := json.NewEncoder(f.IOStreams.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputHuman(f *cmdutil.Factory, emails []jmap.Email, fields []string) error {
	out := f.IOStreams.Out

	if len(emails) == 0 {
		fmt.Fprintln(out, "No emails found.")
		return nil
	}

	cmdutil.PrintEmailList(out, emails, fields)

	fmt.Fprintf(out, "\n%d emails\n", len(emails))
	return nil
}
