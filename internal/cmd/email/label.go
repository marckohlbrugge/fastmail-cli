package email

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdLabel creates the email label command.
func NewCmdLabel(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label <email-id> <add|remove|list> [label]",
		Short: "Manage labels (keywords) on an email",
		Long: `Add, remove, or list labels (keywords) on an email.

Labels in JMAP are called "keywords". Standard keywords start with $
(like $seen, $flagged, $draft). Custom labels can be any string.`,
		Example: `  # List labels on an email
  fm email label M1234567890 list

  # Add a custom label
  fm email label M1234567890 add important

  # Remove a label
  fm email label M1234567890 remove important`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabel(f, args)
		},
	}

	return cmd
}

func runLabel(f *cmdutil.Factory, args []string) error {
	emailID := args[0]
	action := strings.ToLower(args[1])

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	switch action {
	case "list":
		keywords, err := client.GetKeywords(emailID)
		if err != nil {
			return err
		}
		if len(keywords) == 0 {
			fmt.Fprintln(f.IOStreams.Out, "No labels.")
			return nil
		}
		sort.Strings(keywords)
		for _, k := range keywords {
			fmt.Fprintln(f.IOStreams.Out, k)
		}
		return nil

	case "add":
		if len(args) < 3 {
			return fmt.Errorf("label name required for add")
		}
		label := args[2]
		if err := client.SetKeyword(emailID, label, true); err != nil {
			return err
		}
		fmt.Fprintf(f.IOStreams.Out, "Added label: %s\n", label)
		return nil

	case "remove":
		if len(args) < 3 {
			return fmt.Errorf("label name required for remove")
		}
		label := args[2]
		if err := client.SetKeyword(emailID, label, false); err != nil {
			return err
		}
		fmt.Fprintf(f.IOStreams.Out, "Removed label: %s\n", label)
		return nil

	default:
		return fmt.Errorf("unknown action %q: use list, add, or remove", action)
	}
}
