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
	Long:  `Rewrite forge.yaml in the current directory in canonical format.`,
	RunE:  fmtFunc,
}

func fmtFunc(cmd *cobra.Command, args []string) error {
	conf, err := confile.ParseFromFile()
	if err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	err = conf.Fmt()
	if err != nil {
		return fmt.Errorf("formatting config file: %w", err)
	}

	return nil
}
