package config

import (
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config subcommand.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the forge.yaml configuration",
	Long:  `Commands for initializing and formatting the forge.yaml configuration file.`,
}
