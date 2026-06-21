package vacation

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdOff creates the vacation off command.
func NewCmdOff(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "off",
		Short: "Disable vacation auto-responder",
		Long:  "Disable the vacation auto-responder.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOff(f)
		},
	}

	return cmd
}

func runOff(f *cmdutil.Factory) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.SetVacationResponse(false, nil, nil, nil, nil); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Vacation auto-responder disabled.")
	return nil
}
