package mask

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdEnable creates the mask enable command.
func NewCmdEnable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <email-or-id>",
		Short: "Enable a masked email address",
		Long:  "Enable a disabled masked email address so it receives mail again.",
		Args:  cmdutil.ExactArgs(1, "masked email address or ID required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetState(f, args[0], "enabled")
		},
	}

	return cmd
}

// NewCmdDisable creates the mask disable command.
func NewCmdDisable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <email-or-id>",
		Short: "Disable a masked email address",
		Long:  "Disable a masked email address. Mail to this address will be rejected.",
		Args:  cmdutil.ExactArgs(1, "masked email address or ID required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetState(f, args[0], "disabled")
		},
	}

	return cmd
}

// NewCmdDelete creates the mask delete command.
func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <email-or-id>",
		Short: "Delete a masked email address",
		Long:  "Permanently delete a masked email address.",
		Args:  cmdutil.ExactArgs(1, "masked email address or ID required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetState(f, args[0], "deleted")
		},
	}

	return cmd
}

func runSetState(f *cmdutil.Factory, emailOrID, state string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	// If it looks like an email address, find the ID
	id := emailOrID
	if len(emailOrID) > 10 && emailOrID[len(emailOrID)-3:] != ".fm" {
		// Might be an ID already, but let's check if it's an email
		masks, err := client.GetMaskedEmails()
		if err != nil {
			return err
		}
		for _, m := range masks {
			if m.Email == emailOrID {
				id = m.ID
				break
			}
		}
	}

	if err := client.SetMaskedEmailState(id, state); err != nil {
		return err
	}

	switch state {
	case "enabled":
		fmt.Fprintln(f.IOStreams.Out, "Enabled.")
	case "disabled":
		fmt.Fprintln(f.IOStreams.Out, "Disabled.")
	case "deleted":
		fmt.Fprintln(f.IOStreams.Out, "Deleted.")
	}
	return nil
}
