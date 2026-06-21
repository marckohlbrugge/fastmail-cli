package mask

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type listOptions struct {
	JSON  bool
	State string
}

// NewCmdList creates the mask list command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List masked email addresses",
		Long:  "List all masked email addresses with their status.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.State, "state", "", "Filter by state (enabled, disabled, pending)")

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	masks, err := client.GetMaskedEmails()
	if err != nil {
		return err
	}

	// Filter by state if specified
	if opts.State != "" {
		var filtered []jmap.MaskedEmail
		for _, m := range masks {
			if strings.EqualFold(m.State, opts.State) {
				filtered = append(filtered, m)
			}
		}
		masks = filtered
	}

	if opts.JSON {
		return json.NewEncoder(f.IOStreams.Out).Encode(masks)
	}

	if len(masks) == 0 {
		fmt.Fprintln(f.IOStreams.Out, "No masked emails found.")
		return nil
	}

	w := tabwriter.NewWriter(f.IOStreams.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tSTATE\tDOMAIN\tDESCRIPTION")
	for _, m := range masks {
		desc := m.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Email, m.State, m.ForDomain, desc)
	}
	w.Flush()

	return nil
}
