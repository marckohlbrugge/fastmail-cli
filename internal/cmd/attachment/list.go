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
		Long: `List the attachments on an email.

Shows each attachment's name, type, size, and blob ID. Use
'fm attachment download' to download one.`,
		Example: `  # List attachments
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

	out := f.IOStreams.Out

	if opts.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(email.Attachments)
	}

	if len(email.Attachments) == 0 {
		fmt.Fprintln(out, "No attachments.")
		return nil
	}

	for _, att := range email.Attachments {
		name := att.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", name, att.Type, formatSize(att.Size), att.BlobID)
	}

	return nil
}
