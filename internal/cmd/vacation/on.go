package vacation

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type onOptions struct {
	Subject  string
	Body     string
	FromDate string
	ToDate   string
}

// NewCmdOn creates the vacation on command.
func NewCmdOn(f *cmdutil.Factory) *cobra.Command {
	opts := &onOptions{}

	cmd := &cobra.Command{
		Use:   "on",
		Short: "Enable vacation auto-responder",
		Long:  "Enable the vacation auto-responder with an optional message.",
		Example: `  # Enable with a message
  fm vacation on --subject "Out of office" --body "I'm away until Monday."

  # Enable with date range
  fm vacation on --subject "Vacation" --body "Away" --from 2024-12-20 --to 2024-12-31`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOn(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Subject, "subject", "", "Auto-reply subject line")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Auto-reply message body")
	cmd.Flags().StringVar(&opts.FromDate, "from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.ToDate, "to", "", "End date (YYYY-MM-DD)")

	return cmd
}

func runOn(f *cmdutil.Factory, opts *onOptions) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	var subject, body, fromDate, toDate *string
	if opts.Subject != "" {
		subject = &opts.Subject
	}
	if opts.Body != "" {
		body = &opts.Body
	}
	if opts.FromDate != "" {
		// JMAP expects UTC datetime, convert date to start of day
		from := opts.FromDate + "T00:00:00Z"
		fromDate = &from
	}
	if opts.ToDate != "" {
		// JMAP expects UTC datetime, convert date to end of day
		to := opts.ToDate + "T23:59:59Z"
		toDate = &to
	}

	if err := client.SetVacationResponse(true, subject, body, fromDate, toDate); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Vacation auto-responder enabled.")
	return nil
}
