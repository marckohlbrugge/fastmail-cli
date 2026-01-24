package version

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdVersion creates the version command.
func NewCmdVersion(f *cmdutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Show fm version",
		GroupID: "utility",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.IOStreams.Out, "fm version %s\n", version)
			return nil
		},
	}
	return cmd
}
