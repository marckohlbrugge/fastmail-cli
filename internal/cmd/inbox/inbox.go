package inbox

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type inboxOptions struct {
	Limit int
	JSON  bool
}

// NewCmdInbox creates the inbox command.
func NewCmdInbox(f *cmdutil.Factory) *cobra.Command {
	opts := &inboxOptions{}

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "List recent inbox emails",
		Long: `List recent emails from your inbox.

Displays email ID, date, sender, and subject. Unread emails are marked with *.
Emails with attachments are marked with +.`,
		Example: `  # List recent inbox emails
  fm inbox

  # List last 10 emails
  fm inbox --limit 10

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

	return cmd
}

func runInbox(f *cmdutil.Factory, opts *inboxOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
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

	return outputHuman(f, emails)
}

func outputJSON(f *cmdutil.Factory, emails []jmap.Email) error {
	type jsonEmail struct {
		ID            string             `json:"id"`
		ThreadID      string             `json:"threadId"`
		Subject       string             `json:"subject"`
		From          []jmap.EmailAddress `json:"from"`
		ReceivedAt    time.Time          `json:"receivedAt"`
		IsUnread      bool               `json:"isUnread"`
		HasAttachment bool               `json:"hasAttachment"`
		Preview       string             `json:"preview"`
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

func outputHuman(f *cmdutil.Factory, emails []jmap.Email) error {
	out := f.IOStreams.Out

	if len(emails) == 0 {
		fmt.Fprintln(out, "No emails found.")
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

	fmt.Fprintf(out, "\n%d emails (* = unread, + = attachment)\n", len(emails))
	return nil
}

func formatSender(addr jmap.EmailAddress) string {
	if addr.Name != "" {
		return addr.Name
	}
	// Just show the part before @ for brevity
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
