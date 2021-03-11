package config

import (
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
	for target := range conf.Steps {
		help := conf.Steps[target].Help
		if help != "" {
			fmt.Printf("%-"+pad+"v---- %v", target, help)
		}
	}
}

func calcPadding(steps map[string]*confile.Step) string {
	l := 0
	for t := range steps {
		if len(t) > l && len(steps[t].Help) > 0 {
			l = len(t)
		}
	}

	l += 2

	return strconv.Itoa(l)
}
