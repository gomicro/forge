package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the forge.yaml configuration",
	Long:  `Commands for initializing and formatting the forge.yaml configuration file.`,
}
