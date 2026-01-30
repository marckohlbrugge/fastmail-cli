package identity

import (
	"encoding/json"
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type listOptions struct {
	JSON bool
}

// NewCmdList creates the identity list command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all identities",
		Long: `List all sender identities (email addresses) you can send from.

The primary identity (non-deletable) is marked with an asterisk.`,
		Example: `  # List all identities
  fm identity list

  # Output as JSON
  fm identity list --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output in JSON format")

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	identities, err := client.GetIdentities()
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(f, identities)
	}

	return outputHuman(f, identities)
}

func outputJSON(f *cmdutil.Factory, identities []jmap.Identity) error {
	encoder := json.NewEncoder(f.IOStreams.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(identities)
}

func outputHuman(f *cmdutil.Factory, identities []jmap.Identity) error {
	out := f.IOStreams.Out

	if len(identities) == 0 {
		fmt.Fprintln(out, "No identities found.")
		return nil
	}

	for _, id := range identities {
		primary := ""
		if !id.MayDelete {
			primary = " *"
		}

		name := id.Name
		if name == "" {
			name = "(no name)"
		}

		fmt.Fprintf(out, "%-30s  %s%s\n", id.Email, name, primary)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "* = primary identity")

	return nil
}
