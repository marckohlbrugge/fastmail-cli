package email

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type threadOptions struct {
	JSON bool
}

// NewCmdThread creates the email thread command.
func NewCmdThread(f *cmdutil.Factory) *cobra.Command {
	opts := &threadOptions{}

	cmd := &cobra.Command{
		Use:   "thread <email-id>",
		Short: "View all emails in a conversation",
		Long: `Display all emails in a conversation thread.

You can pass either an email ID or thread ID. If an email ID is provided,
the thread containing that email will be displayed.`,
		Example: `  # View a thread by email ID
  fm email thread M1234567890

  # Output as JSON
  fm email thread M1234567890 --json`,
		Args: cmdutil.ExactArgs(1, "email or thread ID required\n\nUsage: fm email thread <id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThread(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runThread(f *cmdutil.Factory, opts *threadOptions, id string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	emails, err := client.GetThread(id)
	if err != nil {
		return err
	}

	if len(emails) == 0 {
		return fmt.Errorf("thread not found")
	}

	if opts.JSON {
		encoder := json.NewEncoder(f.IOStreams.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(emails)
	}

	return printThread(f, emails)
}

func printThread(f *cmdutil.Factory, emails []jmap.Email) error {
	out := f.IOStreams.Out

	fmt.Fprintf(out, "Thread with %d emails:\n\n", len(emails))

	for i, email := range emails {
		from := "(unknown)"
		if len(email.From) > 0 {
			from = jmap.FormatAddresses(email.From)
		}

		date := email.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM")
		subject := email.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		fmt.Fprintf(out, "[%d] %s\n", i+1, from)
		fmt.Fprintf(out, "    %s\n", date)
		fmt.Fprintf(out, "    %s\n", subject)

		// Show preview
		preview := getPreview(&email)
		if preview != "" {
			// Truncate and clean up preview
			preview = strings.ReplaceAll(preview, "\n", " ")
			preview = strings.TrimSpace(preview)
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			fmt.Fprintf(out, "    %s\n", preview)
		}

		fmt.Fprintln(out)
	}

	return nil
}

func getPreview(email *jmap.Email) string {
	if email.Preview != "" {
		return email.Preview
	}

	// Try to get from body
	if email.BodyValues != nil {
		for _, part := range email.TextBody {
			if bv, ok := email.BodyValues[part.PartID]; ok {
				return strings.TrimSpace(bv.Value[:min(200, len(bv.Value))])
			}
		}
	}

	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
