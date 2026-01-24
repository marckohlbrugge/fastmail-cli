package email

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

// NewCmdDelete creates the email delete command.
func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <email-id>",
		Short: "Move an email to trash",
		Long: `Move an email to the Trash folder.

This action requires confirmation unless --yes is provided.
In non-interactive mode (scripts, AI), this command is blocked unless --unsafe is specified.`,
		Example: `  # Delete with confirmation prompt
  fm email delete M1234567890

  # Delete without confirmation
  fm email delete M1234567890 --yes`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email delete <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.Unsafe, "unsafe", false, "Allow in non-interactive mode")

	return cmd
}

func runDelete(f *cmdutil.Factory, opts *deleteOptions, emailID string) error {
	// Check safe mode
	if f.IOStreams.IsSafeMode() && !opts.Unsafe {
		return &cmdutil.SafeModeError{Command: "email delete"}
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Require confirmation unless --yes
	if !opts.Yes && f.IOStreams.IsInteractive() {
		// Get email info for confirmation
		email, err := client.GetEmailByID(emailID)
		if err != nil {
			return err
		}

		subject := email.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		fmt.Fprintf(f.IOStreams.ErrOut, "Subject: %s\n", subject)
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete this email? [y/N] ")

		scanner := bufio.NewScanner(f.IOStreams.In)
		response := ""
		if scanner.Scan() {
			response = scanner.Text()
		}

		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			return cmdutil.CancelError
		}
	}

	if err := client.DeleteEmail(emailID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Moved to Trash.")
	return nil
}
