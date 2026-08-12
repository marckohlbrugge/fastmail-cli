package attachment

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type downloadOptions struct {
	Output string
	Force  bool
}

// NewCmdDownload creates the attachment download command.
func NewCmdDownload(f *cmdutil.Factory) *cobra.Command {
	opts := &downloadOptions{}

	cmd := &cobra.Command{
		Use:   "download <email-id> <name-or-blob-id>",
		Short: "Download an attachment",
		Long: `Download an attachment from an email.

The attachment can be identified by its filename (as shown by
'fm attachment list' or 'fm email read') or by its blob ID.

By default the file is saved to the current directory under the
attachment's filename. Use --output to choose a different path.`,
		Example: `  # Download by filename
  fm attachment download M1234567890 report.pdf

  # Download by blob ID
  fm attachment download M1234567890 Gb1234567890

  # Download to a specific path
  fm attachment download M1234567890 report.pdf --output ~/Documents/report.pdf`,
		Args: cmdutil.ExactArgs(2, "email ID and attachment name or blob ID required\n\nUsage: fm attachment download <email-id> <name-or-blob-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(f, opts, args[0], args[1])
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output path (default: attachment filename in current directory)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite the output file if it exists")

	return cmd
}

func runDownload(f *cmdutil.Factory, opts *downloadOptions, emailID, nameOrBlobID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	email, err := client.GetEmailByID(emailID)
	if err != nil {
		return err
	}

	att, err := findAttachment(email.Attachments, nameOrBlobID)
	if err != nil {
		return err
	}

	outputPath := opts.Output
	if outputPath == "" {
		// The attachment name comes from the sender, so keep only the
		// base name to prevent writing outside the current directory.
		outputPath = filepath.Base(att.Name)
		if outputPath == "" || outputPath == "." || outputPath == string(filepath.Separator) {
			outputPath = att.BlobID
		}
	}

	if !opts.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite)", outputPath)
		}
	}

	data, err := client.DownloadBlob(att.BlobID, att.Name, att.Type)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Downloaded %s (%s)\n", outputPath, formatSize(int64(len(data))))
	return nil
}

// findAttachment finds an attachment by filename or blob ID.
func findAttachment(attachments []jmap.Attachment, nameOrBlobID string) (jmap.Attachment, error) {
	var byName []jmap.Attachment
	for _, att := range attachments {
		if att.BlobID == nameOrBlobID {
			return att, nil
		}
		if att.Name == nameOrBlobID {
			byName = append(byName, att)
		}
	}

	switch len(byName) {
	case 0:
		return jmap.Attachment{}, fmt.Errorf("no attachment named '%s' found (use 'fm attachment list' to see attachments)", nameOrBlobID)
	case 1:
		return byName[0], nil
	default:
		return jmap.Attachment{}, fmt.Errorf("multiple attachments named '%s'; specify a blob ID instead (see 'fm attachment list')", nameOrBlobID)
	}
}
