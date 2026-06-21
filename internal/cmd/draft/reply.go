package draft

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type replyOptions struct {
	Body     string
	BodyFile string
	All      bool
	Attach   []string
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
	cmd.Flags().StringSliceVar(&opts.Attach, "attach", nil, "Files to attach (can be repeated)")

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

	// Upload attachments
	var attachments []jmap.Attachment
	for _, path := range opts.Attach {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", path, err)
		}
		filename := filepath.Base(path)
		resp, err := client.UploadBlob(filename, "", data)
		if err != nil {
			return fmt.Errorf("cannot upload %s: %w", path, err)
		}
		attachments = append(attachments, jmap.Attachment{
			BlobID: resp.BlobID,
			Type:   resp.Type,
			Name:   filename,
			Size:   resp.Size,
		})
	}

	draftID, err := client.CreateReplyDraft(emailID, body, opts.All, attachments)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Reply draft created: %s\n", draftID)
	return nil
}
