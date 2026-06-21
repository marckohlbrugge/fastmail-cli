package attachment

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type downloadOptions struct {
	Output string
}

// NewCmdDownload creates the attachment download command.
func NewCmdDownload(f *cmdutil.Factory) *cobra.Command {
	opts := &downloadOptions{}

	cmd := &cobra.Command{
		Use:   "download <email-id> <blob-id>",
		Short: "Download an attachment",
		Long: `Download an attachment from an email to the current directory.

The email-id and blob-id can be obtained from 'fm attachment list'.
By default, the file is saved using the attachment's original name.
Use --output to specify a different filename or path.`,
		Example: `  # Download attachment to current directory
  fm attachment download M1234567890 B1234567890

  # Download with a specific filename
  fm attachment download M1234567890 B1234567890 --output report.pdf

  # Download to a specific path
  fm attachment download M1234567890 B1234567890 --output ~/Downloads/report.pdf`,
		Args: cmdutil.ExactArgs(2, "email ID and blob ID required\n\nUsage: fm attachment download <email-id> <blob-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(f, opts, args[0], args[1])
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output filename or path")

	return cmd
}

func runDownload(f *cmdutil.Factory, opts *downloadOptions, emailID, blobID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Get email to find attachment metadata
	email, err := client.GetEmailByID(emailID)
	if err != nil {
		return err
	}

	// Find the attachment by blob ID
	var attachmentName string
	var attachmentType string
	for _, att := range email.Attachments {
		if att.BlobID == blobID {
			attachmentName = att.Name
			attachmentType = att.Type
			break
		}
	}

	if attachmentName == "" && attachmentType == "" {
		return fmt.Errorf("attachment with blob ID '%s' not found on email '%s'", blobID, emailID)
	}

	// Determine output filename
	outputPath := opts.Output
	if outputPath == "" {
		if attachmentName != "" {
			outputPath = attachmentName
		} else {
			// Fallback if no name available
			outputPath = blobID
		}
	}

	// Expand home directory if needed
	if len(outputPath) > 0 && outputPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			outputPath = filepath.Join(home, outputPath[1:])
		}
	}

	// Download the blob
	data, err := client.DownloadBlob(blobID, attachmentName, attachmentType)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Downloaded: %s (%s)\n", outputPath, formatSize(int64(len(data))))
	return nil
}
