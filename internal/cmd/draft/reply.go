package draft

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type replyOptions struct {
	Body     string
	BodyFile string
	All      bool
}

// NewCmdReply creates the draft reply command.
func NewCmdReply(f *cmdutil.Factory) *cobra.Command {
	opts := &replyOptions{}

	cmd := &cobra.Command{
		Use:   "reply <email-id>",
		Short: "Create a reply draft",
		Long: `Create a draft reply to an email.

Automatically sets the recipient, subject (with Re: prefix), and threading
headers for proper conversation grouping.`,
		Example: `  # Reply with body text
  fm draft reply M1234567890 --body "Thanks for your email!"

  # Reply with body from file
  fm draft reply M1234567890 --body-file response.txt

  # Reply-all to include all recipients
  fm draft reply M1234567890 --all --body "Thanks everyone!"`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm draft reply <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReply(f, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Body, "body", "", "Reply body text")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Read body from file")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Reply to all recipients")

	return cmd
}

func runReply(f *cmdutil.Factory, opts *replyOptions, emailID string) error {
	// Get body content
	body := opts.Body
	if opts.BodyFile != "" {
		content, err := os.ReadFile(opts.BodyFile)
		if err != nil {
			return fmt.Errorf("failed to read body file: %w", err)
		}
		body = string(content)
	}

	if body == "" {
		return cmdutil.FlagErrorf("--body or --body-file required")
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	draftID, err := client.CreateReplyDraft(emailID, body, opts.All)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Reply draft created: %s\n", draftID)
	return nil
}
