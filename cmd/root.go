package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gomicro/forge/cmd/config"
	"github.com/gomicro/forge/confile"
	"github.com/gomicro/forge/fmt"
)

func init() {
	cobra.OnInitialize(initEnvs)

	RootCmd.PersistentFlags().Bool("verbose", false, "show more verbose output")
	err := viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}

	RootCmd.AddCommand(config.ConfigCmd)
}

func initEnvs() {
}

// RootCmd represents the base command without any params on it.
var RootCmd = &cobra.Command{
	Use:   "forge step [step]...",
	Short: "A CLI for building projects",
	Long:  `Forge is a CLI tool for executing, in a consistent manner, scripts and commands for building and maintaining projects.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   rootFunc,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("Failed to execute: %v", err.Error())
		os.Exit(1)
	}
}

func rootFunc(cmd *cobra.Command, args []string) {
	conf, err := confile.ParseFromFile()
	if err != nil {
		fmt.Printf("Failed: %v", err.Error())
	}

	for _, target := range args {
		step, found := conf.Steps[target]
		if !found {
			fmt.Printf("step '%v' not found", target)
			os.Exit(1)
		}

		out, err := step.Execute()
		if err != nil {
			fmt.Printf("failed executing %v: %v", target, err.Error())
			os.Exit(1)
		}

		if out != "" {
			fmt.Printf("%v", out)
		}
	}
}
