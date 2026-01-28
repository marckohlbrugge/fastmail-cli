package search

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	Folder string
	Limit  int
	JSON   bool
	Fields string
}


// NewCmdSearch creates the search command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	opts := &searchOptions{}

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search emails",
		Long: `Search emails using Fastmail's search syntax.

Query is optional when using --folder to list all emails in a folder.

Supports text search and filter operators:
  from:alice     - Emails from alice
  to:bob         - Emails to bob
  subject:hello  - Subject contains hello
  has:attachment - Has attachments
  is:unread      - Unread emails only
  is:read        - Read emails only
  is:flagged     - Flagged/starred emails
  is:draft       - Draft emails
  before:DATE    - Emails before date (YYYY-MM-DD)
  after:DATE     - Emails after date (YYYY-MM-DD)

Boolean operators (case-insensitive):
  OR             - Match either term
  AND            - Match both terms (also implicit between terms)
  NOT            - Exclude matching emails
  ()             - Group expressions`,
		Example: `  # List all drafts
  fm search --folder drafts

  # List all emails in a folder
  fm search --folder inbox --limit 100

  # Search for emails from alice
  fm search "from:alice"

  # Search for meeting-related emails
  fm search "subject:meeting"

  # Search with attachments
  fm search "has:attachment from:bob"

  # Boolean OR: match either term
  fm search "hiring OR discount"

  # Boolean AND with NOT
  fm search "from:newsletter AND NOT is:unread"

  # Grouped expressions
  fm search "(from:alice OR from:bob) AND subject:meeting"

  # Search within a specific folder
  fm search "from:newsletter" --folder inbox

  # Output as JSON
  fm search "from:alice" --json`,
		GroupID: "core",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			return runSearch(f, opts, query)
		},
	}

	cmd.Flags().StringVar(&opts.Folder, "folder", "", "Restrict search to folder ID or name")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum results (max 500)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&opts.Fields, "fields", "", "Comma-separated list of fields to display (id,threadId,subject,from,to,cc,date,preview,unread,attachment)")

	return cmd
}

func runSearch(f *cmdutil.Factory, opts *searchOptions, query string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Parse and validate fields
	fields := cmdutil.ParseFields(opts.Fields)
	if err := cmdutil.ValidateFields(fields); err != nil {
		return err
	}

	filters := jmap.SearchFilters{
		Query: query,
		Limit: opts.Limit,
	}

	// Resolve folder if specified
	if opts.Folder != "" {
		mailbox, err := resolveMailbox(client, opts.Folder)
		if err != nil {
			return err
		}
		filters.MailboxID = mailbox.ID
	}

	emails, err := client.Search(filters)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(f, emails)
	}

	return outputHuman(f, emails, query, fields)
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

func outputHuman(f *cmdutil.Factory, emails []jmap.Email, query string, fields []string) error {
	out := f.IOStreams.Out

	if len(emails) == 0 {
		if query == "" {
			fmt.Fprintln(out, "No emails found")
		} else {
			fmt.Fprintf(out, "No emails found matching: %s\n", query)
		}
		return nil
	}

	cmdutil.PrintEmailList(out, emails, fields)

	fmt.Fprintf(out, "\n%d results\n", len(emails))
	return nil
}

