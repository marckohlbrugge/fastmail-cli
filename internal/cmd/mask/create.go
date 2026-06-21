package mask

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Domain      string
	Description string
}

// NewCmdCreate creates the mask create command.
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new masked email address",
		Long:  "Create a new masked email address with an optional domain and description.",
		Example: `  # Create a basic masked email
  fm mask create

  # Create for a specific domain
  fm mask create --domain example.com

  # Create with description
  fm mask create --domain shopping.com --desc "Online store account"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Domain, "domain", "", "Associate with this domain")
	cmd.Flags().StringVar(&opts.Description, "desc", "", "Description for the masked email")

	return cmd
}

func runCreate(f *cmdutil.Factory, opts *createOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	masked, err := client.CreateMaskedEmail(opts.Domain, opts.Description)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Created: %s\n", masked.Email)
	return nil
}
