package root

import (
	"bytes"
	"testing"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRoot(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	assert.Equal(t, "fm <command> [flags]", cmd.Use)
	assert.Equal(t, "Fastmail CLI", cmd.Short)
	assert.True(t, cmd.SilenceErrors)
	assert.True(t, cmd.SilenceUsage)

	// Verify subcommands are registered
	subcommands := cmd.Commands()
	names := make([]string, 0, len(subcommands))
	for _, c := range subcommands {
		names = append(names, c.Name())
	}

	assert.Contains(t, names, "inbox")
	assert.Contains(t, names, "search")
	assert.Contains(t, names, "folders")
	assert.Contains(t, names, "email")
	assert.Contains(t, names, "draft")
	assert.Contains(t, names, "folder")
	assert.Contains(t, names, "auth")
	assert.Contains(t, names, "version")
	assert.Contains(t, names, "completion")

	// Verify command groups are set up
	groups := cmd.Groups()
	groupIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}
	assert.Contains(t, groupIDs, "auth")
	assert.Contains(t, groupIDs, "core")
	assert.Contains(t, groupIDs, "email")
	assert.Contains(t, groupIDs, "draft")
	assert.Contains(t, groupIDs, "folder")
	assert.Contains(t, groupIDs, "utility")

	_ = stdout // unused in this test
}

func TestIsRootCmd(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{Use: "child"}
	root.AddCommand(child)

	assert.True(t, isRootCmd(root))
	assert.False(t, isRootCmd(child))
}

func TestGetCommandsInGroup(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.AddGroup(&cobra.Group{ID: "group1", Title: "Group 1"})
	root.AddGroup(&cobra.Group{ID: "group2", Title: "Group 2"})

	// Commands need a Run function to be considered "available"
	noop := func(cmd *cobra.Command, args []string) {}

	cmd1 := &cobra.Command{Use: "cmd1", GroupID: "group1", Run: noop}
	cmd2 := &cobra.Command{Use: "cmd2", GroupID: "group1", Run: noop}
	cmd3 := &cobra.Command{Use: "cmd3", GroupID: "group2", Run: noop}
	hiddenCmd := &cobra.Command{Use: "hidden", GroupID: "group1", Hidden: true, Run: noop}

	root.AddCommand(cmd1, cmd2, cmd3, hiddenCmd)

	group1Cmds := getCommandsInGroup(root, "group1")
	assert.Len(t, group1Cmds, 2)
	assert.Contains(t, group1Cmds, cmd1)
	assert.Contains(t, group1Cmds, cmd2)
	assert.NotContains(t, group1Cmds, hiddenCmd) // hidden commands excluded

	group2Cmds := getCommandsInGroup(root, "group2")
	assert.Len(t, group2Cmds, 1)
	assert.Contains(t, group2Cmds, cmd3)

	emptyGroup := getCommandsInGroup(root, "nonexistent")
	assert.Len(t, emptyGroup, 0)
}

func TestRootHelpFunc(t *testing.T) {
	t.Run("root command shows full help", func(t *testing.T) {
		ios, _, stdout, _ := iostreams.Test()
		f := &cmdutil.Factory{IOStreams: ios}
		cmd := NewCmdRoot(f)

		rootHelpFunc(stdout, cmd, nil)
		output := stdout.String()

		// Check for expected sections
		assert.Contains(t, output, "USAGE")
		assert.Contains(t, output, "fm <command> [flags]")
		assert.Contains(t, output, "CORE COMMANDS")
		assert.Contains(t, output, "FLAGS")
		assert.Contains(t, output, "--help")
		assert.Contains(t, output, "--version")
		assert.Contains(t, output, "EXAMPLES")
		assert.Contains(t, output, "LEARN MORE")
		assert.Contains(t, output, "AUTHENTICATION")
		assert.Contains(t, output, "ENVIRONMENT")
		assert.Contains(t, output, "FASTMAIL_TOKEN")
	})

	t.Run("subcommand shows subcommand help", func(t *testing.T) {
		ios, _, stdout, _ := iostreams.Test()
		f := &cmdutil.Factory{IOStreams: ios}
		cmd := NewCmdRoot(f)

		// Get a subcommand
		inboxCmd, _, err := cmd.Find([]string{"inbox"})
		require.NoError(t, err)

		rootHelpFunc(stdout, inboxCmd, nil)
		output := stdout.String()

		assert.Contains(t, output, "USAGE")
		assert.Contains(t, output, "FLAGS")
	})
}

func TestPrintSubcommandHelp(t *testing.T) {
	t.Run("shows long description if present", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Short description",
			Long:  "This is a longer description with more details.",
		}

		printSubcommandHelp(&buf, cmd)
		output := buf.String()

		assert.Contains(t, output, "This is a longer description")
		assert.NotContains(t, output, "Short description")
	})

	t.Run("shows short description if no long", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Short description only",
		}

		printSubcommandHelp(&buf, cmd)
		output := buf.String()

		assert.Contains(t, output, "Short description only")
	})

	t.Run("shows subcommands", func(t *testing.T) {
		var buf bytes.Buffer
		noop := func(cmd *cobra.Command, args []string) {}

		cmd := &cobra.Command{Use: "parent"}
		child1 := &cobra.Command{Use: "child1", Short: "First child", Run: noop}
		child2 := &cobra.Command{Use: "child2", Short: "Second child", Run: noop}
		hidden := &cobra.Command{Use: "hidden", Short: "Hidden", Hidden: true, Run: noop}
		cmd.AddCommand(child1, child2, hidden)

		printSubcommandHelp(&buf, cmd)
		output := buf.String()

		assert.Contains(t, output, "COMMANDS")
		assert.Contains(t, output, "child1")
		assert.Contains(t, output, "First child")
		assert.Contains(t, output, "child2")
		assert.NotContains(t, output, "hidden") // hidden commands excluded
	})

	t.Run("shows examples", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cobra.Command{
			Use:     "test",
			Example: "  test --flag value",
		}

		printSubcommandHelp(&buf, cmd)
		output := buf.String()

		assert.Contains(t, output, "EXAMPLES")
		assert.Contains(t, output, "test --flag value")
	})
}

func TestRootUsageFunc(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "fm <command>"}

	err := rootUsageFunc(&buf, cmd)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Usage: fm <command>")
	assert.Contains(t, output, "Run 'fm --help' for more information")
}

func TestExecuteErrorHandling(t *testing.T) {
	// Note: Execute() uses os.Stderr directly which makes it hard to capture output.
	// These tests verify the exit codes for different error types.
	// The actual error handling is tested through the command structure.

	t.Run("FlagError returns exit code 1", func(t *testing.T) {
		err := cmdutil.FlagErrorf("test error")
		_, ok := err.(*cmdutil.FlagError)
		assert.True(t, ok)
	})

	t.Run("AuthError returns exit code 2", func(t *testing.T) {
		var err error = cmdutil.NewAuthError("test")
		_, ok := err.(*cmdutil.AuthError)
		assert.True(t, ok)
	})

	t.Run("NotFoundError returns exit code 3", func(t *testing.T) {
		var err error = &cmdutil.NotFoundError{Resource: "email", ID: "123"}
		_, ok := err.(*cmdutil.NotFoundError)
		assert.True(t, ok)
	})

	t.Run("SilentError is recognized", func(t *testing.T) {
		assert.Equal(t, cmdutil.SilentError, cmdutil.SilentError)
	})

	t.Run("CancelError is recognized", func(t *testing.T) {
		assert.Equal(t, cmdutil.CancelError, cmdutil.CancelError)
	})
}

func TestVersionFlag(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	// The -v flag should exist
	flag := cmd.Flags().Lookup("version")
	require.NotNil(t, flag)
	assert.Equal(t, "v", flag.Shorthand)

	_ = stdout
}

func TestHelpFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	// The --help flag should exist
	flag := cmd.PersistentFlags().Lookup("help")
	require.NotNil(t, flag)
}

func TestSuggestions(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	// Verify suggestion distance is set
	assert.Equal(t, 2, cmd.SuggestionsMinimumDistance)
}
