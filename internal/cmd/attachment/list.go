package attachment

import (
	"encoding/json"
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type listOptions struct {
	JSON bool
}

// NewCmdList creates the attachment list command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list <email-id>",
		Short: "List attachments on an email",
		Long: `List all attachments on an email with their name, type, size, and blob ID.

The email-id can be obtained from 'fm inbox' or 'fm search' output.
The blob ID can be used with 'fm attachment download' to download the file.`,
		Example: `  # List attachments on an email
  fm attachment list M1234567890

  # Output as JSON
  fm attachment list M1234567890 --json`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm attachment list <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions, emailID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	email, err := client.GetEmailByID(emailID)
	if err != nil {
		return err
	}

	if len(email.Attachments) == 0 {
		fmt.Fprintln(f.IOStreams.Out, "No attachments.")
		return nil
	}

	if opts.JSON {
		encoder := json.NewEncoder(f.IOStreams.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(email.Attachments)
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Attachments (%d):\n", len(email.Attachments))
	for _, att := range email.Attachments {
		name := att.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Fprintf(out, "  %s\n", name)
		fmt.Fprintf(out, "    Type:   %s\n", att.Type)
		fmt.Fprintf(out, "    Size:   %s\n", formatSize(att.Size))
		fmt.Fprintf(out, "    BlobID: %s\n", att.BlobID)
	}

	return nil
}

// formatSize formats bytes as a human-readable size.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
