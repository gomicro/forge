package cmd

import (
	"github.com/spf13/cobra"

	"github.com/gomicro/forge/confile"
	"github.com/gomicro/forge/fmt"
)

func init() {
	RootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a forge config file",
	Long:  `Collects the basic information to initialize a forge config file.`,
	Run:   initFunc,
}

func initFunc(cmd *cobra.Command, args []string) {
	f := &confile.File{
		Project: &confile.Project{
			Name: "sample-forge-project",
		},
		Steps: map[string]*confile.Step{
			"build": {
				Help: "build the project",
				Cmd:  "echo \"run the build\"",
			},
		},
	}

	err := f.Fmt()
	if err != nil {
		fmt.Printf("Error Initializing File: %v", err.Error())
	}
}
