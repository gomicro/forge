package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultShell = "zsh"
)

var (
	shell string
)

func init() {
	RootCmd.AddCommand(CompletionCmd)

	CompletionCmd.Flags().StringVar(&shell, "shell", defaultShell, "desired shell to generate completions for")
}

// CompletionCmd represents the command for generating completion files for the
// forge cli.
var CompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion scripts",
	Long:  `Generate completion scripts for bash, zsh, or powershell. Use --shell to specify the target shell (default: zsh).`,
	RunE:  completionFunc,
}

func completionFunc(cmd *cobra.Command, args []string) error {
	return generateCompletion(shell, os.Stdout)
}

func generateCompletion(targetShell string, out io.Writer) error {
	var err error
	switch strings.ToLower(targetShell) {
	case "bash":
		err = RootCmd.GenBashCompletion(out)
	case "ps", "powershell", "power_shell":
		err = RootCmd.GenPowerShellCompletion(out)
	case "zsh":
		err = RootCmd.GenZshCompletion(out)
	default:
		return fmt.Errorf("unsupported shell: %q", targetShell)
	}

	if err != nil {
		return fmt.Errorf("generating completion output: %w", err)
	}

	return nil
}
