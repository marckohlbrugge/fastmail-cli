package contacts

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Name  string
	Email string
	Phone string
	Org   string
}

// NewCmdCreate creates the contacts create command.
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contact",
		Long:  "Create a new contact with name, email, phone, and organization.",
		Example: `  # Create a contact with name only
  fm contacts create --name "John Doe"

  # Create a contact with all fields
  fm contacts create --name "John Doe" --email john@example.com --phone "+1-555-1234" --org "Acme Inc"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Contact name (required)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Email address")
	cmd.Flags().StringVar(&opts.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization name")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func runCreate(f *cmdutil.Factory, opts *createOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	contact := &jmap.ContactCard{
		Name: &jmap.ContactName{
			Full: opts.Name,
		},
	}

	if opts.Email != "" {
		contact.Emails = map[string]jmap.ContactEmail{
			"email1": {Address: opts.Email, Label: "personal"},
		}
	}

	if opts.Phone != "" {
		contact.Phones = map[string]jmap.ContactPhone{
			"phone1": {Number: opts.Phone, Label: "mobile"},
		}
	}

	if opts.Org != "" {
		contact.Orgs = map[string]jmap.ContactOrg{
			"org1": {Name: opts.Org},
		}
	}

	contactID, err := client.CreateContact(contact)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Created contact: %s\n", contactID)
	return nil
}
