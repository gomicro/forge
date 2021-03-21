package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gomicro/forge/confile"
	ffmt "github.com/gomicro/forge/fmt"
)

func init() {
	cobra.OnInitialize(initEnvs)

	RootCmd.PersistentFlags().Bool("verbose", false, "show more verbose output")
	err := viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		ffmt.Printf("Error setting up: %v\n", err.Error())
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
		ffmt.Printf("Failed to execute: %v", err.Error())
		os.Exit(1)
	}
}

func rootFunc(cmd *cobra.Command, args []string) {
	conf, err := confile.ParseFromFile()
	if err != nil {
		ffmt.Printf("Failed: %v", err.Error())
	}

	aliases, steps, err := parseTargets(conf, args)
	if err != nil {
		ffmt.Printf("target not found: %v", err.Error())
		os.Exit(1)
	}

	for _, a := range aliases {
		err := conf.Aliases[a].Execute(conf.Steps)
		if err != nil {
			ffmt.Printf("failed executing alias %v: %v", a, err.Error())
			os.Exit(1)
		}
	}

	for _, s := range steps {
		err := conf.Steps[s].Execute()
		if err != nil {
			ffmt.Printf("failed executing step %v: %v", s, err.Error())
			os.Exit(1)
		}
	}
}

func parseTargets(conf *confile.File, args []string) (aliases []string, steps []string, err error) {
	for _, a := range args {
		_, found := conf.Aliases[a]
		if !found {
			_, found := conf.Steps[a]
			if !found {
				return nil, nil, fmt.Errorf("%v", a)
			}

			steps = append(steps, a)
			continue
		}

		aliases = append(aliases, a)
	}

	return
}

func validArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return argStrings(), cobra.ShellCompDirectiveNoFileComp
}

func argStrings() []string {
	conf, err := confile.ParseFromFile()
	if err != nil {
		return nil
	}

	sorted := make([]string, 0, len(conf.Steps)+len(conf.Aliases))
	lookup := make(map[string]string, cap(sorted))

	for a := range conf.Aliases {
		sorted = append(sorted, a)
		lookup[a] = "alias"
	}

	for s := range conf.Steps {
		sorted = append(sorted, s)
		lookup[s] = "step"
	}

	sort.Strings(sorted)

	args := make([]string, 0, len(sorted))

	for _, t := range sorted {
		var help string

		switch lookup[t] {
		case "alias":
			help = conf.Aliases[t].Help
			if help == "" {
				help = fmt.Sprintf("steps: %v", conf.Aliases[t].Steps)
			}
		case "step":
			help = conf.Steps[t].Help
			if help == "" {
				help = conf.Steps[t].Cmd
			}
		}

		args = append(args, t+"\t"+help)
	}

	return args
}
