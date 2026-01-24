package search

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	Folder string
	Limit  int
	JSON   bool
}

// NewCmdSearch creates the search command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	opts := &searchOptions{}

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search emails",
		Long: `Search emails using Fastmail's search syntax.

Supports text search and filter operators:
  from:alice     - Emails from alice
  to:bob         - Emails to bob
  subject:hello  - Subject contains hello
  has:attachment - Has attachments
  is:unread      - Unread emails only`,
		Example: `  # Search for emails from alice
  fm search "from:alice"

  # Search for meeting-related emails
  fm search "subject:meeting"

  # Search with attachments
  fm search "has:attachment from:bob"

  # Search in a specific folder
  fm search "from:newsletter" --folder inbox

  # Output as JSON
  fm search "from:alice" --json`,
		GroupID: "core",
		Args:    cmdutil.ExactArgs(1, "search query required\n\nUsage: fm search <query>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(f, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Folder, "folder", "", "Restrict search to folder ID or name")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum results (max 500)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runSearch(f *cmdutil.Factory, opts *searchOptions, query string) error {
	client, err := f.JMAPClient()
	if err != nil {
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

	return outputHuman(f, emails, query)
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

func outputHuman(f *cmdutil.Factory, emails []jmap.Email, query string) error {
	out := f.IOStreams.Out

	if len(emails) == 0 {
		fmt.Fprintf(out, "No emails found matching: %s\n", query)
		return nil
	}

	for _, email := range emails {
		unreadMarker := " "
		if email.IsUnread() {
			unreadMarker = "*"
		}
		attachmentMarker := " "
		if email.HasAttachment {
			attachmentMarker = "+"
		}

		from := "(unknown)"
		if len(email.From) > 0 {
			from = formatSender(email.From[0])
		}

		date := formatRelativeDate(email.ReceivedAt)
		subject := email.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		// Truncate for display
		id := truncate(email.ID, 12)
		from = truncate(from, 30)
		subject = truncate(subject, 50)

		fmt.Fprintf(out, "%s%s %-12s  %-12s  %-30s  %s\n",
			unreadMarker, attachmentMarker, id, date, from, subject)
	}

	fmt.Fprintf(out, "\n%d results (* = unread, + = attachment)\n", len(emails))
	return nil
}

func formatSender(addr jmap.EmailAddress) string {
	if addr.Name != "" {
		return addr.Name
	}
	parts := strings.Split(addr.Email, "@")
	return parts[0]
}

func formatRelativeDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 48*time.Hour:
		return "Yesterday"
	case diff < 7*24*time.Hour:
		return t.Weekday().String()[:3]
	case t.Year() == now.Year():
		return t.Format("Jan 2")
	default:
		return t.Format("Jan 2, 2006")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
