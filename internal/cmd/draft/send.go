package draft

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type sendOptions struct {
	Yes    bool
	Unsafe bool
}

// NewCmdSend creates the draft send command.
func NewCmdSend(f *cmdutil.Factory) *cobra.Command {
	opts := &sendOptions{}

	cmd := &cobra.Command{
		Use:   "send <draft-id>",
		Short: "Send a draft email",
		Long: `Send a draft email.

This is a critical action and requires confirmation. The email will be
sent immediately and moved to your Sent folder.

In non-interactive mode (scripts, AI), this command is blocked unless
--unsafe is specified. This prevents accidental sending by automated tools.`,
		Example: `  # Send with confirmation prompt
  fm draft send M1234567890

  # Send without confirmation
  fm draft send M1234567890 --yes

  # Send in script/AI mode (requires explicit unsafe flag)
  fm draft send M1234567890 --unsafe --yes`,
		Args: cmdutil.ExactArgs(1, "draft ID required\n\nUsage: fm draft send <draft-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.Unsafe, "unsafe", false, "Allow in non-interactive mode")

	return cmd
}

func runSend(f *cmdutil.Factory, opts *sendOptions, draftID string) error {
	// Check safe mode - sending is critical
	if f.IOStreams.IsSafeMode() && !opts.Unsafe {
		return &cmdutil.SafeModeError{Command: "draft send"}
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Get draft info for confirmation
	draft, err := client.GetEmailByID(draftID)
	if err != nil {
		return err
	}

	// Validate it's a draft
	if !draft.IsDraft() {
		return fmt.Errorf("email %s is not a draft", draftID)
	}

	// Require confirmation unless --yes
	if !opts.Yes {
		showSendConfirmation(f, draft)

		if f.IOStreams.IsInteractive() {
			fmt.Fprintf(f.IOStreams.ErrOut, "Send this email? [y/N] ")

			scanner := bufio.NewScanner(f.IOStreams.In)
			response := ""
			if scanner.Scan() {
				response = scanner.Text()
			}

			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
				return cmdutil.CancelError
			}
		} else {
			// Non-interactive but --unsafe was provided, still need --yes
			return cmdutil.FlagErrorf("non-interactive mode requires --yes flag")
		}
	}

	if err := client.SendEmail(draftID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Email sent successfully.")
	return nil
}

func showSendConfirmation(f *cmdutil.Factory, draft *jmap.Email) {
	out := f.IOStreams.ErrOut

	to := jmap.FormatAddresses(draft.To)
	subject := draft.Subject
	if subject == "" {
		subject = "(no subject)"
	}

	fmt.Fprintf(out, "To:      %s\n", to)
	if len(draft.CC) > 0 {
		fmt.Fprintf(out, "Cc:      %s\n", jmap.FormatAddresses(draft.CC))
	}
	fmt.Fprintf(out, "Subject: %s\n", subject)
	fmt.Fprintln(out)
}
