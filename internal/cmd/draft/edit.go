package draft

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type editOptions struct {
	To       []string
	CC       []string
	Subject  string
	Body     string
	BodyFile string
	From     string
}

// NewCmdEdit creates the draft edit command.
func NewCmdEdit(f *cmdutil.Factory) *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <draft-id>",
		Short: "Edit a draft email",
		Long: `Edit an existing draft email.

Only the specified fields will be updated. The original values are preserved
for fields not specified.`,
		Example: `  # Update subject
  fm draft edit M1234567890 --subject "New subject"

  # Update body
  fm draft edit M1234567890 --body "Updated content"

  # Update recipients
  fm draft edit M1234567890 --to new@example.com`,
		Args: cmdutil.ExactArgs(1, "draft ID required\n\nUsage: fm draft edit <draft-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(f, opts, args[0])
		},
	}

	cmd.Flags().StringArrayVar(&opts.To, "to", nil, "Replace recipient(s)")
	cmd.Flags().StringArrayVar(&opts.CC, "cc", nil, "Replace CC recipient(s)")
	cmd.Flags().StringVar(&opts.Subject, "subject", "", "Replace subject")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Replace body")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Replace body from file")
	cmd.Flags().StringVar(&opts.From, "from", "", "Replace sender")

	return cmd
}

func runEdit(f *cmdutil.Factory, opts *editOptions, draftID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Fetch existing draft
	existing, err := client.GetEmailByID(draftID)
	if err != nil {
		return err
	}

	// Build updated draft
	to := extractEmails(existing.To)
	if len(opts.To) > 0 {
		to = opts.To
	}

	cc := extractEmails(existing.CC)
	if len(opts.CC) > 0 {
		cc = opts.CC
	}

	subject := existing.Subject
	if opts.Subject != "" {
		subject = opts.Subject
	}

	from := ""
	if len(existing.From) > 0 {
		from = existing.From[0].Email
	}
	if opts.From != "" {
		from = opts.From
	}

	// Get body
	body := getBodyFromEmail(existing)
	if opts.BodyFile != "" {
		content, err := os.ReadFile(opts.BodyFile)
		if err != nil {
			return fmt.Errorf("failed to read body file: %w", err)
		}
		body = string(content)
	} else if opts.Body != "" {
		body = opts.Body
	}

	// Create new draft
	newDraftID, err := client.SaveDraft(jmap.DraftEmail{
		To:       to,
		CC:       cc,
		Subject:  subject,
		TextBody: body,
		From:     from,
	})
	if err != nil {
		return err
	}

	// Delete old draft
	if err := client.DeleteDraft(draftID); err != nil {
		// Log but don't fail - new draft was created
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: could not delete old draft: %v\n", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Draft updated: %s\n", newDraftID)
	return nil
}

func extractEmails(addrs []jmap.EmailAddress) []string {
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.Email
	}
	return result
}

func getBodyFromEmail(email *jmap.Email) string {
	if email.BodyValues == nil {
		return ""
	}

	for _, part := range email.TextBody {
		if bv, ok := email.BodyValues[part.PartID]; ok {
			return bv.Value
		}
	}

	return ""
}
