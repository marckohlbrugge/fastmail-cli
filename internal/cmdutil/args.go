package cmdutil

import (
	"github.com/spf13/cobra"
)

// MinimumArgs returns a PositionalArgs that requires at least n arguments
// with a custom error message.
func MinimumArgs(n int, msg string) cobra.PositionalArgs {
	if msg == "" {
		return cobra.MinimumNArgs(n)
	}

	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return FlagErrorf("%s", msg)
		}
		return nil
	}
}

// ExactArgs returns a PositionalArgs that requires exactly n arguments
// with a custom error message.
func ExactArgs(n int, msg string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			return FlagErrorf("too many arguments")
		}

		if len(args) < n {
			return FlagErrorf("%s", msg)
		}

		return nil
	}
}

// NoArgsQuoteReminder returns an error if any arguments are provided,
// with a hint about quoting values with spaces.
func NoArgsQuoteReminder(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return nil
	}

	errMsg := "unknown argument"
	if len(args) > 1 {
		errMsg = "unknown arguments"
	}

	return FlagErrorf("%s; please quote values that contain spaces", errMsg)
}

// RangeArgs returns a PositionalArgs that requires between min and max arguments.
func RangeArgs(min, max int, msg string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min {
			return FlagErrorf("%s", msg)
		}
		if len(args) > max {
			return FlagErrorf("too many arguments")
		}
		return nil
	}
}
