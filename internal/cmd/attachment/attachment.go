package attachment

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAttachment creates the attachment parent command.
func NewCmdAttachment(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment <command>",
		Short: "Manage email attachments",
		Long:  "List and download email attachments.",
		Example: `  $ fm attachment list M1234567890
  $ fm attachment download M1234567890 report.pdf`,
		GroupID: "email",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdDownload(f))

	return cmd
}

// formatSize formats a byte count as a human-readable string.
func formatSize(size int64) string {
	const (
		kb = 1 << 10
		mb = 1 << 20
		gb = 1 << 30
	)

	switch {
	case size >= gb:
		return fmt.Sprintf("%.1f GB", float64(size)/gb)
	case size >= mb:
		return fmt.Sprintf("%.1f MB", float64(size)/mb)
	case size >= kb:
		return fmt.Sprintf("%.1f KB", float64(size)/kb)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
