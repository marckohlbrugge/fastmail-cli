package draft

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type newOptions struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Body     string
	BodyFile string
	From     string
}

// NewCmdNew creates the draft new command.
func NewCmdNew(f *cmdutil.Factory) *cobra.Command {
	opts := &newOptions{}

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new draft email",
		Long: `Create a new draft email.

The draft will be saved to your Drafts folder. You can then edit it
in Fastmail or send it with 'fm draft send'.`,
		Example: `  # Create a simple draft
  fm draft new --to bob@example.com --subject "Hello" --body "Hi Bob!"

  # Create with multiple recipients
  fm draft new --to alice@example.com --to bob@example.com --subject "Team"

  # Create with body from file
  fm draft new --to bob@example.com --subject "Report" --body-file report.txt

  # Create with CC
  fm draft new --to bob@example.com --cc manager@example.com --subject "Update"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(f, opts)
		},
	}

	cmd.Flags().StringArrayVar(&opts.To, "to", nil, "Recipient email address (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.CC, "cc", nil, "CC recipient (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.BCC, "bcc", nil, "BCC recipient (can be repeated)")
	cmd.Flags().StringVar(&opts.Subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Email body text")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Read body from file")
	cmd.Flags().StringVar(&opts.From, "from", "", "Sender email (default: primary identity)")

	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")

	return cmd
}

func runNew(f *cmdutil.Factory, opts *newOptions) error {
	if len(opts.To) == 0 {
		return cmdutil.FlagErrorf("--to is required")
	}
	if opts.Subject == "" {
		return cmdutil.FlagErrorf("--subject is required")
	}

	// Get body content
	body := opts.Body
	if opts.BodyFile != "" {
		content, err := os.ReadFile(opts.BodyFile)
		if err != nil {
			return fmt.Errorf("failed to read body file: %w", err)
		}
		body = string(content)
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	draftID, err := client.SaveDraft(jmap.DraftEmail{
		To:       opts.To,
		CC:       opts.CC,
		BCC:      opts.BCC,
		Subject:  opts.Subject,
		TextBody: body,
		From:     opts.From,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Draft created: %s\n", draftID)
	return nil
}
