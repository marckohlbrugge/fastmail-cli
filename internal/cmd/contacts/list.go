package contacts

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Limit int
}

// NewCmdList creates the contacts list command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Long:  "List contacts in your address book.",
		Example: `  # List first 50 contacts
  fm contacts list

  # List first 100 contacts
  fm contacts list --limit 100`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f, opts)
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of contacts to return")

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	contacts, err := client.GetContacts(opts.Limit)
	if err != nil {
		return err
	}

	return outputContacts(f, contacts)
}

func outputContacts(f *cmdutil.Factory, contacts []jmap.ContactCard) error {
	out := f.IOStreams.Out

	if len(contacts) == 0 {
		fmt.Fprintln(out, "No contacts found.")
		return nil
	}

	for _, c := range contacts {
		name := getContactName(&c)
		email := getPrimaryEmail(&c)
		fmt.Fprintf(out, "%-20s  %-30s  %s\n", c.ID, name, email)
	}

	return nil
}

// getContactName returns the full name from a ContactCard.
func getContactName(c *jmap.ContactCard) string {
	if c.Name == nil {
		return ""
	}
	return c.Name.Full
}

// getPrimaryEmail returns the first/preferred email from a ContactCard.
func getPrimaryEmail(c *jmap.ContactCard) string {
	if len(c.Emails) == 0 {
		return ""
	}
	// Return first email found (maps don't have order, but typically there's only one)
	for _, email := range c.Emails {
		return email.Address
	}
	return ""
}
