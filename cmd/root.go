package cmd

import (
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

	RootCmd.PersistentFlags().Bool("solo", false, "run a step solo, without its pre or post steps")
	err = viper.BindPFlag("solo", RootCmd.PersistentFlags().Lookup("solo"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}

	RootCmd.PersistentFlags().Bool("no-pre", false, "skip running pre steps")
	err = viper.BindPFlag("no-pre", RootCmd.PersistentFlags().Lookup("no-pre"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}

	RootCmd.PersistentFlags().Bool("no-post", false, "skip running post steps")
	err = viper.BindPFlag("no-post", RootCmd.PersistentFlags().Lookup("no-post"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}
}

func initEnvs() {
}

// RootCmd represents the base command without any params on it.
var RootCmd = &cobra.Command{
	Use:               "forge step [step]...",
	Short:             "A CLI for building projects",
	Long:              `Forge is a CLI tool for executing, in a consistent manner, scripts and commands for building and maintaining projects.`,
	Args:              cobra.MinimumNArgs(1),
	Run:               rootFunc,
	ValidArgsFunction: validArgsFunc,
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
		os.Exit(1)
	}

	for _, a := range args {
		_, found := conf.Steps[a]
		if !found {
			fmt.Printf("target not found: %v", a)
			os.Exit(1)
		}
	}

	for _, s := range args {
		err := conf.Steps[s].Execute(conf.Steps, conf.Envs, conf.Vars)
		if err != nil {
			fmt.Printf("failed executing step %v: %v", s, err.Error())
			os.Exit(1)
		}
	}
}

func validArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return argStrings(), cobra.ShellCompDirectiveNoFileComp
}

func argStrings() []string {
	conf, err := confile.ParseFromFile()
	if err != nil {
		return nil
	}

	sorted := make([]string, 0, len(conf.Steps))

	for s := range conf.Steps {
		sorted = append(sorted, s)
	}

	sort.Strings(sorted)

	args := make([]string, 0, len(sorted))

	for _, t := range sorted {
		var help string

		help = conf.Steps[t].Help
		if help == "" {
			help = conf.Steps[t].Cmd
		}

		args = append(args, t+"\t"+help)
	}

	return args
}
