package email

import (
	"fmt"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type snoozeOptions struct {
	Duration string
	Until    string
}

// NewCmdSnooze creates the email snooze command.
func NewCmdSnooze(f *cmdutil.Factory) *cobra.Command {
	opts := &snoozeOptions{}

	cmd := &cobra.Command{
		Use:   "snooze <email-id>",
		Short: "Snooze an email",
		Long: `Snooze an email until a specified time. The email will reappear in your inbox at that time.

Specify either a duration (--for) or an absolute time (--until).`,
		Example: `  # Snooze for 2 hours
  fm email snooze M1234567890 --for 2h

  # Snooze until tomorrow morning
  fm email snooze M1234567890 --until "2024-12-25T09:00:00"

  # Snooze for 1 day
  fm email snooze M1234567890 --for 24h`,
		Args: cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email snooze <email-id> [--for duration | --until time]"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnooze(f, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Duration, "for", "", "Snooze duration (e.g., 2h, 30m, 1d)")
	cmd.Flags().StringVar(&opts.Until, "until", "", "Snooze until time (RFC3339 format)")

	return cmd
}

// NewCmdUnsnooze creates the email unsnooze command.
func NewCmdUnsnooze(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsnooze <email-id>",
		Short: "Remove snooze from an email",
		Long:  "Remove snooze from an email and move it back to your inbox.",
		Args:  cmdutil.ExactArgs(1, "email ID required\n\nUsage: fm email unsnooze <email-id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnsnooze(f, args[0])
		},
	}

	return cmd
}

func runSnooze(f *cmdutil.Factory, opts *snoozeOptions, emailID string) error {
	if opts.Duration == "" && opts.Until == "" {
		return fmt.Errorf("specify either --for or --until")
	}
	if opts.Duration != "" && opts.Until != "" {
		return fmt.Errorf("specify only one of --for or --until")
	}

	var until string
	if opts.Duration != "" {
		// Parse duration and calculate until time
		d, err := parseDuration(opts.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		until = time.Now().Add(d).UTC().Format(time.RFC3339)
	} else {
		// Validate until time
		t, err := time.Parse(time.RFC3339, opts.Until)
		if err != nil {
			// Try parsing without timezone
			t, err = time.Parse("2006-01-02T15:04:05", opts.Until)
			if err != nil {
				return fmt.Errorf("invalid time format, use RFC3339 (e.g., 2024-12-25T09:00:00Z)")
			}
		}
		until = t.UTC().Format(time.RFC3339)
	}

	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.SnoozeEmail(emailID, until); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Snoozed until %s\n", until)
	return nil
}

func runUnsnooze(f *cmdutil.Factory, emailID string) error {
	client, err := f.JMAPClient()
	if err != nil {
		return err
	}

	if err := client.UnsnoozeEmail(emailID); err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Unsnoozed and moved to inbox.")
	return nil
}

// parseDuration parses a duration string with support for days (d).
func parseDuration(s string) (time.Duration, error) {
	// Handle days
	if len(s) > 0 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
