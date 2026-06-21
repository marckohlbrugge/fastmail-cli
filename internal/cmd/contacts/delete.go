package contacts

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

// NewCmdDelete creates the contacts delete command.
func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <contact-id>",
		Short: "Delete a contact",
		Long: `Delete a contact from your address book.

This action requires confirmation unless --yes is provided.
In non-interactive mode (scripts, AI), this command is blocked unless --unsafe is specified.`,
		Example: `  # Delete with confirmation prompt
  fm contacts delete abc123

  # Delete without confirmation
  fm contacts delete abc123 --yes`,
		Args: cmdutil.ExactArgs(1, "contact ID required\n\nUsage: fm contacts delete <contact-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.Unsafe, "unsafe", false, "Allow in non-interactive mode")

	return cmd
}

func runDelete(f *cmdutil.Factory, opts *deleteOptions, contactID string) error {
	// Check safe mode
	if f.IOStreams.IsSafeMode() && !opts.Unsafe {
		return &cmdutil.SafeModeError{Command: "contacts delete"}
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// Require confirmation unless --yes
	if !opts.Yes && f.IOStreams.IsInteractive() {
		// Get contact info for confirmation
		contact, err := client.GetContact(contactID)
		if err != nil {
			return err
		}

		name := "(no name)"
		if contact.Name != nil && contact.Name.Full != "" {
			name = contact.Name.Full
		}

		fmt.Fprintf(f.IOStreams.ErrOut, "Contact: %s\n", name)
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete this contact? [y/N] ")

		scanner := bufio.NewScanner(f.IOStreams.In)
		response := ""
		if scanner.Scan() {
			response = scanner.Text()
		}

		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			return cmdutil.CancelError
		}
	}

	if err := client.DeleteContact(contactID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Contact deleted.")
	return nil
}
