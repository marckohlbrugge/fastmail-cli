package inbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type inboxOptions struct {
	Limit  int
	JSON   bool
	Fields string
}

// NewCmdInbox creates the inbox command.
func NewCmdInbox(f *cmdutil.Factory) *cobra.Command {
	opts := &inboxOptions{}

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "List recent inbox emails",
		Long: `List recent emails from your inbox.

By default displays email ID, date, sender, and subject.
Use --fields to customize which columns are shown.`,
		Example: `  # List recent inbox emails
  fm inbox

  # List last 10 emails
  fm inbox --limit 10

  # Show custom fields
  fm inbox --fields "id,date,from,to,subject"

  # Output as JSON (for scripting)
  fm inbox --json`,
		GroupID: "core",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInbox(f, opts)
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of emails to show (max 50)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&opts.Fields, "fields", "", "Comma-separated list of fields to display (id,threadId,subject,from,to,cc,date,preview,unread,attachment)")

	return cmd
}

func runInbox(f *cmdutil.Factory, opts *inboxOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Parse and validate fields
	fields := cmdutil.ParseFields(opts.Fields)
	if err := cmdutil.ValidateFields(fields); err != nil {
		return err
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

	if opts.JSON {
		return outputJSON(f, emails)
	}

	return outputHuman(f, emails, fields)
}

func outputJSON(f *cmdutil.Factory, emails []jmap.Email) error {
	type jsonEmail struct {
		ID            string              `json:"id"`
		ThreadID      string              `json:"threadId"`
		Subject       string              `json:"subject"`
		From          []jmap.EmailAddress `json:"from"`
		To            []jmap.EmailAddress `json:"to"`
		CC            []jmap.EmailAddress `json:"cc,omitempty"`
		ReceivedAt    time.Time           `json:"receivedAt"`
		IsUnread      bool                `json:"isUnread"`
		HasAttachment bool                `json:"hasAttachment"`
		Preview       string              `json:"preview"`
	}

	output := make([]jsonEmail, len(emails))
	for i, e := range emails {
		output[i] = jsonEmail{
			ID:            e.ID,
			ThreadID:      e.ThreadID,
			Subject:       e.Subject,
			From:          e.From,
			To:            e.To,
			CC:            e.CC,
			ReceivedAt:    e.ReceivedAt,
			IsUnread:      e.IsUnread(),
			HasAttachment: e.HasAttachment,
			Preview:       e.Preview,
		}
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
