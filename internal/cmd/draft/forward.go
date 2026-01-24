package draft

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type forwardOptions struct {
	To       []string
	CC       []string
	Body     string
	BodyFile string
	From     string
}

// NewCmdForward creates the draft forward command.
func NewCmdForward(f *cmdutil.Factory) *cobra.Command {
	opts := &forwardOptions{}

	cmd := &cobra.Command{
		Use:   "forward <email-id>",
		Short: "Create a forward draft",
		Long: `Create a forward draft with the original message.

The forwarded message includes the original headers and body. Any attachments
from the original email are also included.`,
		Example: `  # Forward to someone
  fm draft forward M1234567890 --to bob@example.com

  # Forward with an introduction
  fm draft forward M1234567890 --to bob@example.com --body "FYI, see below"

  # Forward to multiple recipients
  fm draft forward M1234567890 --to alice@example.com --to bob@example.com`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm draft forward <email-id> --to <recipient>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runForward(f, opts, args[0])
		},
	}

	cmd.Flags().StringArrayVar(&opts.To, "to", nil, "Recipient email address (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.CC, "cc", nil, "CC recipient (can be repeated)")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Introduction text before forwarded message")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Read introduction from file")
	cmd.Flags().StringVar(&opts.From, "from", "", "Sender email (default: primary identity)")

	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func runForward(f *cmdutil.Factory, opts *forwardOptions, emailID string) error {
	if len(opts.To) == 0 {
		return cmdutil.FlagErrorf("--to is required")
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

	draftID, err := client.CreateForwardDraft(jmap.ForwardOptions{
		EmailID: emailID,
		To:      opts.To,
		CC:      opts.CC,
		Body:    body,
		From:    opts.From,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Forward draft created: %s\n", draftID)
	return nil
}
