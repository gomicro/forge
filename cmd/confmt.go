package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gomicro/forge/confile"
)

func init() {
	RootCmd.AddCommand(confmtCmd)
}

var confmtCmd = &cobra.Command{
	Use:   "confmt",
	Short: "Format the forge config file",
	Long:  `Format and adjust the forge file for consistency.`,
	RunE:  confmtFunc,
}

func confmtFunc(cmd *cobra.Command, args []string) error {
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
