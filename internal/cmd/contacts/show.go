package contacts

import (
	"fmt"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

// NewCmdShow creates the contacts show command.
func NewCmdShow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <contact-id>",
		Short: "Show contact details",
		Long:  "Display full details of a contact including all emails, phones, organization, and notes.",
		Example: `  fm contacts show abc123`,
		Args:  cmdutil.ExactArgs(1, "contact ID required\n\nUsage: fm contacts show <contact-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(f, args[0])
		},
	}

	return cmd
}

func runShow(f *cmdutil.Factory, contactID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	contact, err := client.GetContact(contactID)
	if err != nil {
		return err
	}

	return printContact(f, contact)
}

func printContact(f *cmdutil.Factory, c *jmap.ContactCard) error {
	out := f.IOStreams.Out
	sep := strings.Repeat("-", 50)

	fmt.Fprintln(out, sep)
	fmt.Fprintf(out, "ID:           %s\n", c.ID)

	name := ""
	if c.Name != nil {
		name = c.Name.Full
	}
	fmt.Fprintf(out, "Name:         %s\n", name)

	if len(c.Emails) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Emails:")
		for key, email := range c.Emails {
			label := email.Label
			if label == "" {
				label = key
			}
			fmt.Fprintf(out, "  %-10s  %s\n", label, email.Address)
		}
	}

	if len(c.Phones) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Phones:")
		for key, phone := range c.Phones {
			label := phone.Label
			if label == "" {
				label = key
			}
			fmt.Fprintf(out, "  %-10s  %s\n", label, phone.Number)
		}
	}

	if len(c.Orgs) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Organizations:")
		for _, org := range c.Orgs {
			fmt.Fprintf(out, "  %s\n", org.Name)
		}
	}

	if len(c.Notes) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Notes:")
		for _, note := range c.Notes {
			fmt.Fprintln(out, note.Note)
		}
	}

	fmt.Fprintln(out, sep)
	return nil
}
