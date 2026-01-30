package identities

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/identity"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdIdentities creates the identities command (alias for identity list).
func NewCmdIdentities(f *cmdutil.Factory) *cobra.Command {
	cmd := identity.NewCmdList(f)

	// Override to make it a top-level command
	cmd.Use = "identities"
	cmd.Short = "List sender identities"
	cmd.GroupID = "core"
	cmd.Example = `  # List all identities
  fm identities

  # Output as JSON
  fm identities --json`

	return cmd
}
