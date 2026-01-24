package completion

import (
	"fmt"
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the completion command.
func NewCmdCompletion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for fm.

The output can be sourced directly or saved to a file for later use.

SUPPORTED SHELLS
  bash, zsh, fish, powershell`,
		Example: `  # Bash - add to ~/.bashrc:
  eval "$(fm completion bash)"

  # Zsh - add to ~/.zshrc:
  eval "$(fm completion zsh)"

  # Or save to a completions directory:
  fm completion zsh > ~/.zsh/completions/_fm

  # Fish:
  fm completion fish > ~/.config/fish/completions/fm.fish`,
		GroupID:           "utility",
		Args:              cmdutil.ExactArgs(1, "shell type required: bash, zsh, fish, or powershell"),
		ValidArgsFunction: completeShellTypes,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			rootCmd := cmd.Root()

			switch shell {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type: %s\n\nSupported: bash, zsh, fish, powershell", shell)
			}
		},
	}
	return cmd
}

func completeShellTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
}
