package root

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/auth"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/completion"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/draft"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/email"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/folder"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/folders"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/inbox"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/search"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/version"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "dev"

// NewCmdRoot creates the root command for the CLI.
func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fm <command> [flags]",
		Short: "Fastmail CLI",
		Long:  "Work seamlessly with Fastmail from the command line.",
		Example: `  $ fm inbox
  $ fm email read M1234567890
  $ fm search "from:alice"
  $ fm draft new --to bob@example.com --subject "Hello"`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Enable suggestions for typos
	cmd.SuggestionsMinimumDistance = 2

	// Set up custom help and usage functions
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		rootHelpFunc(f.IOStreams.Out, c, args)
	})
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		return rootUsageFunc(f.IOStreams.ErrOut, c)
	})

	// Global flags
	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	cmd.Flags().BoolP("version", "v", false, "Show fm version")

	// Add command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "auth",
		Title: "Authentication",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "email",
		Title: "Email commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "draft",
		Title: "Draft commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "folder",
		Title: "Folder commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "utility",
		Title: "Utility commands",
	})

	// Auth commands
	cmd.AddCommand(auth.NewCmdAuth(f))

	// Core commands (top-level)
	cmd.AddCommand(inbox.NewCmdInbox(f))
	cmd.AddCommand(search.NewCmdSearch(f))
	cmd.AddCommand(folders.NewCmdFolders(f))

	// Email subcommands
	cmd.AddCommand(email.NewCmdEmail(f))

	// Draft subcommands
	cmd.AddCommand(draft.NewCmdDraft(f))

	// Folder subcommands
	cmd.AddCommand(folder.NewCmdFolder(f))

	// Utility commands
	cmd.AddCommand(version.NewCmdVersion(f, Version))
	cmd.AddCommand(completion.NewCmdCompletion(f))

	return cmd
}

// rootHelpFunc provides custom help output similar to gh CLI
func rootHelpFunc(w io.Writer, cmd *cobra.Command, args []string) {
	if isRootCmd(cmd) {
		printRootHelp(w, cmd)
		return
	}

	// Default help for subcommands - print usage and flags
	printSubcommandHelp(w, cmd)
}

func printSubcommandHelp(w io.Writer, cmd *cobra.Command) {
	if cmd.Long != "" {
		fmt.Fprintln(w, cmd.Long)
		fmt.Fprintln(w)
	} else if cmd.Short != "" {
		fmt.Fprintln(w, cmd.Short)
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "USAGE\n  %s\n\n", cmd.UseLine())

	// Show subcommands if any
	subcommands := cmd.Commands()
	if len(subcommands) > 0 {
		fmt.Fprintln(w, "COMMANDS")
		for _, c := range subcommands {
			if c.IsAvailableCommand() {
				fmt.Fprintf(w, "  %-16s %s\n", c.Name(), c.Short)
			}
		}
		fmt.Fprintln(w)
	}

	// Show flags
	flags := cmd.Flags()
	if flags.HasAvailableFlags() {
		fmt.Fprintln(w, "FLAGS")
		fmt.Fprintln(w, flags.FlagUsages())
	}

	// Show examples
	if cmd.Example != "" {
		fmt.Fprintln(w, "EXAMPLES")
		fmt.Fprintln(w, cmd.Example)
	}
}

func isRootCmd(cmd *cobra.Command) bool {
	return cmd.Parent() == nil
}

func printRootHelp(w io.Writer, cmd *cobra.Command) {
	fmt.Fprintf(w, "%s\n\n", cmd.Long)

	fmt.Fprintf(w, "USAGE\n  %s\n\n", cmd.Use)

	// Print command groups
	groups := cmd.Groups()
	for _, group := range groups {
		cmds := getCommandsInGroup(cmd, group.ID)
		if len(cmds) == 0 {
			continue
		}

		fmt.Fprintf(w, "%s\n", strings.ToUpper(group.Title))
		for _, c := range cmds {
			fmt.Fprintf(w, "  %-16s %s\n", c.Name(), c.Short)
		}
		fmt.Fprintln(w)
	}

	// Print flags
	fmt.Fprintln(w, "FLAGS")
	fmt.Fprintln(w, "  -h, --help      Show help for command")
	fmt.Fprintln(w, "  -v, --version   Show fm version")
	fmt.Fprintln(w)

	// Print examples
	if cmd.Example != "" {
		fmt.Fprintln(w, "EXAMPLES")
		fmt.Fprintln(w, cmd.Example)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "LEARN MORE")
	fmt.Fprintln(w, "  Use 'fm <command> --help' for more information about a command.")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "AUTHENTICATION")
	fmt.Fprintln(w, "  Run 'fm auth login' to authenticate with your Fastmail account.")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "ENVIRONMENT")
	fmt.Fprintln(w, "  FASTMAIL_TOKEN  API token (overrides stored credentials)")
	fmt.Fprintln(w, "  FM_UNSAFE=1     Allow destructive operations in non-interactive mode")
	fmt.Fprintln(w, "  NO_COLOR        Disable color output")
}

func getCommandsInGroup(cmd *cobra.Command, groupID string) []*cobra.Command {
	var cmds []*cobra.Command
	for _, c := range cmd.Commands() {
		if c.GroupID == groupID && c.IsAvailableCommand() {
			cmds = append(cmds, c)
		}
	}
	return cmds
}

func rootUsageFunc(w io.Writer, cmd *cobra.Command) error {
	fmt.Fprintf(w, "Usage: %s\n", cmd.UseLine())
	fmt.Fprintf(w, "\nRun '%s --help' for more information.\n", cmd.CommandPath())
	return nil
}

// Execute runs the root command
func Execute() int {
	f := cmdutil.NewFactory()
	rootCmd := NewCmdRoot(f)

	if err := rootCmd.Execute(); err != nil {
		// Handle different error types
		switch e := err.(type) {
		case *cmdutil.FlagError:
			fmt.Fprintf(os.Stderr, "Error: %s\n", e.Error())
			return 1
		case *cmdutil.SafeModeError:
			fmt.Fprintf(os.Stderr, "Error: %s\n", e.Error())
			return 1
		case *cmdutil.AuthError:
			fmt.Fprintf(os.Stderr, "Authentication error: %s\n", e.Error())
			return 2
		case *cmdutil.NotFoundError:
			fmt.Fprintf(os.Stderr, "Error: %s\n", e.Error())
			return 3
		default:
			if err == cmdutil.SilentError {
				return 1
			}
			if err == cmdutil.CancelError {
				return 0
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return 1
		}
	}
	return 0
}
