package contacts

import (
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	Limit int
}

// NewCmdSearch creates the contacts search command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	opts := &searchOptions{}

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search contacts",
		Long:  "Search contacts by name, email, or other fields.",
		Example: `  # Search for contacts named John
  fm contacts search "John"

  # Search with custom limit
  fm contacts search "example.com" --limit 20`,
		Args: cmdutil.ExactArgs(1, "search query required\n\nUsage: fm contacts search <query>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(f, opts, args[0])
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of contacts to return")

	return cmd
}

func runSearch(f *cmdutil.Factory, opts *searchOptions, query string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	contacts, err := client.SearchContacts(query, opts.Limit)
	if err != nil {
		return err
	}

	return outputContacts(f, contacts)
}
