package config

import (
	"fmt"

	"github.com/gomicro/forge/confile"
	"github.com/spf13/cobra"
)

func init() {
	ConfigCmd.AddCommand(fmtCmd)
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format the forge.yaml config file",
	Long:  `Rewrite forge.yaml in the current directory in canonical format, preserving comments.`,
	RunE:  fmtFunc,
}

func fmtFunc(cmd *cobra.Command, args []string) error {
	err := confile.Format(confile.DefaultPath)
	if err != nil {
		return fmt.Errorf("formatting config file: %w", err)
	}

	return nil
}
