package draft

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	Yes    bool
	Unsafe bool
}

// NewCmdDraftDelete creates the draft delete command.
func NewCmdDraftDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <draft-id>",
		Short: "Delete a draft email",
		Long: `Delete a draft email.

This action requires confirmation unless --yes is provided.
In non-interactive mode (scripts, AI), this command is blocked unless --unsafe is specified.`,
		Example: `  # Delete with confirmation prompt
  fm draft delete M1234567890

  # Delete without confirmation
  fm draft delete M1234567890 --yes`,
		Args: cmdutil.ExactArgs(1, "draft ID required\n\nUsage: fm draft delete <draft-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDraftDelete(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.Unsafe, "unsafe", false, "Allow in non-interactive mode")

	return cmd
}

func runDraftDelete(f *cmdutil.Factory, opts *deleteOptions, draftID string) error {
	// Check safe mode
	if f.IOStreams.IsSafeMode() && !opts.Unsafe {
		return &cmdutil.SafeModeError{Command: "draft delete"}
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Require confirmation unless --yes
	if !opts.Yes && f.IOStreams.IsInteractive() {
		// Get draft info for confirmation
		draft, err := client.GetEmailByID(draftID)
		if err != nil {
			return err
		}

		subject := draft.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		fmt.Fprintf(f.IOStreams.ErrOut, "Subject: %s\n", subject)
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete this draft? [y/N] ")

		scanner := bufio.NewScanner(f.IOStreams.In)
		response := ""
		if scanner.Scan() {
			response = scanner.Text()
		}

		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			return cmdutil.CancelError
		}
	}

	if err := client.DeleteDraft(draftID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Draft deleted.")
	return nil
}
