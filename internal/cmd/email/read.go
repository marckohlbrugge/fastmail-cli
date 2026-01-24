package email

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type readOptions struct {
	JSON bool
}

// NewCmdRead creates the email read command.
func NewCmdRead(f *cmdutil.Factory) *cobra.Command {
	opts := &readOptions{}

	cmd := &cobra.Command{
		Use:   "read <email-id>",
		Short: "Display the full content of an email",
		Long: `Display the full content of an email including headers, body, and attachments.

The email-id can be obtained from 'fm inbox' or 'fm search' output.`,
		Example: `  # Read an email
  fm email read M1234567890

  # Output as JSON
  fm email read M1234567890 --json`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email read <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRead(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runRead(f *cmdutil.Factory, opts *readOptions, emailID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	email, err := client.GetEmailByID(emailID)
	if err != nil {
		return err
	}

	if opts.JSON {
		encoder := json.NewEncoder(f.IOStreams.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(email)
	}

	return printEmail(f, email)
}

func printEmail(f *cmdutil.Factory, email *jmap.Email) error {
	out := f.IOStreams.Out
	sep := strings.Repeat("─", 72)

	fmt.Fprintln(out, sep)
	fmt.Fprintf(out, "ID:      %s\n", email.ID)
	fmt.Fprintf(out, "Thread:  %s\n", email.ThreadID)
	fmt.Fprintf(out, "From:    %s\n", jmap.FormatAddresses(email.From))
	fmt.Fprintf(out, "To:      %s\n", jmap.FormatAddresses(email.To))
	if len(email.CC) > 0 {
		fmt.Fprintf(out, "Cc:      %s\n", jmap.FormatAddresses(email.CC))
	}
	fmt.Fprintf(out, "Date:    %s\n", email.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM"))

	subject := email.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	fmt.Fprintf(out, "Subject: %s\n", subject)
	fmt.Fprintln(out, sep)

	// Get body content
	body := getBodyText(email)
	if body == "" {
		body = "(no body)"
	}
	fmt.Fprintln(out, body)

	// Show attachments
	if len(email.Attachments) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, sep)
		fmt.Fprintln(out, "Attachments:")
		for _, att := range email.Attachments {
			name := att.Name
			if name == "" {
				name = att.PartID
			}
			fmt.Fprintf(out, "  - %s (%s, %d bytes)\n", name, att.Type, att.Size)
		}
	}

	return nil
}

func getBodyText(email *jmap.Email) string {
	if email.BodyValues == nil {
		return ""
	}

	// Try text body first
	for _, part := range email.TextBody {
		if bv, ok := email.BodyValues[part.PartID]; ok && bv.Value != "" {
			// If text body is substantial, use it
			if len(strings.TrimSpace(bv.Value)) > 100 {
				return bv.Value
			}
		}
	}

	// Fall back to HTML body
	for _, part := range email.HTMLBody {
		if bv, ok := email.BodyValues[part.PartID]; ok && bv.Value != "" {
			return htmlToText(bv.Value)
		}
	}

	// Use short text body if that's all we have
	for _, part := range email.TextBody {
		if bv, ok := email.BodyValues[part.PartID]; ok && bv.Value != "" {
			return bv.Value
		}
	}

	return ""
}

func htmlToText(html string) string {
	// Add newlines for block elements
	replacements := []struct {
		pattern string
		replace string
	}{
		{`<br\s*/?>`, "\n"},
		{`</p>`, "\n\n"},
		{`</div>`, "\n"},
		{`</tr>`, "\n"},
		{`</li>`, "\n"},
		{`<hr\s*/?>`, "\n───\n"},
	}

	text := html

	// Apply replacements
	for _, r := range replacements {
		re := regexp.MustCompile("(?i)" + r.pattern)
		text = re.ReplaceAllString(text, r.replace)
	}

	// Remove style and script content
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	text = styleRe.ReplaceAllString(text, "")
	text = scriptRe.ReplaceAllString(text, "")

	// Remove remaining tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text = tagRe.ReplaceAllString(text, "")

	// Decode common HTML entities
	entities := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  `"`,
		"&#39;":   "'",
		"&rsquo;": "'",
		"&lsquo;": "'",
		"&rdquo;": `"`,
		"&ldquo;": `"`,
		"&ndash;": "–",
		"&mdash;": "—",
	}
	for entity, char := range entities {
		text = strings.ReplaceAll(text, entity, char)
	}

	// Clean up whitespace
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n `).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(` \n`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
