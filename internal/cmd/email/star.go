package email

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdStar creates the email star command.
func NewCmdStar(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "star <email-id>",
		Short: "Star (flag) an email",
		Long:  `Star an email by setting the $flagged keyword.`,
		Example: `  fm email star M1234567890`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email star <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStar(f, args[0], true)
		},
	}

	return cmd
}

// NewCmdUnstar creates the email unstar command.
func NewCmdUnstar(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstar <email-id>",
		Short: "Unstar (unflag) an email",
		Long:  `Unstar an email by removing the $flagged keyword.`,
		Example: `  fm email unstar M1234567890`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email unstar <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStar(f, args[0], false)
		},
	}

	return cmd
}

func runStar(f *cmdutil.Factory, emailID string, star bool) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.SetFlagged(emailID, star); err != nil {
		return err
	}

	if star {
		fmt.Fprintln(f.IOStreams.Out, "Starred.")
	} else {
		fmt.Fprintln(f.IOStreams.Out, "Unstarred.")
	}
	return nil
}
