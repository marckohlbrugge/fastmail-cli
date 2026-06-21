package vacation

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdStatus creates the vacation status command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show vacation auto-responder status",
		Long:  "Display the current vacation auto-responder settings.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(f)
		},
	}

	return cmd
}

func runStatus(f *cmdutil.Factory) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	vacation, err := client.GetVacationResponse()
	if err != nil {
		return err
	}

	out := f.IOStreams.Out

	if vacation.IsEnabled {
		fmt.Fprintln(out, "Status: ON")
	} else {
		fmt.Fprintln(out, "Status: OFF")
	}

	if vacation.Subject != nil && *vacation.Subject != "" {
		fmt.Fprintf(out, "Subject: %s\n", *vacation.Subject)
	}
	if vacation.TextBody != nil && *vacation.TextBody != "" {
		fmt.Fprintf(out, "Message:\n%s\n", *vacation.TextBody)
	}
	if vacation.FromDate != nil && *vacation.FromDate != "" {
		fmt.Fprintf(out, "From: %s\n", *vacation.FromDate)
	}
	if vacation.ToDate != nil && *vacation.ToDate != "" {
		fmt.Fprintf(out, "To: %s\n", *vacation.ToDate)
	}

	return nil
}
