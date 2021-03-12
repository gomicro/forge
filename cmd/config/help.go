package config

import (
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/gomicro/forge/confile"
	"github.com/gomicro/forge/fmt"
)

func init() {
	ConfigCmd.AddCommand(confHelpCmd)
}

var confHelpCmd = &cobra.Command{
	Use:   "help",
	Short: "Display help info of the config file",
	Long:  `Display the defined help info from the config file.`,
	Run:   confHelpFunc,
}

func confHelpFunc(cmd *cobra.Command, args []string) {
	conf, err := confile.ParseFromFile()
	if err != nil {
		fmt.Printf("Failed: %v", err.Error())
	}

	pad := calcPadding(conf.Steps)

	targets := make([]string, 0, len(conf.Steps))
	for t := range conf.Steps {
		targets = append(targets, t)
	}

	sort.Strings(targets)

	for _, t := range targets {
		help := conf.Steps[t].Help
		if help == "" {
			help = conf.Steps[t].Cmd
		}

		fmt.Printf("%-"+pad+"v---- %v", t, help)
	}
}

func calcPadding(steps map[string]*confile.Step) string {
	l := 0
	for t := range steps {
		if len(t) > l {
			l = len(t)
		}
	}

	l += 2

	return strconv.Itoa(l)
}
