package cmd

import (
	"os"
	"path"

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
	if confile.Exists() {
		fmt.Printf("config file already exists")
		os.Exit(1)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working dir: %v", err.Error())
		os.Exit(1)
	}

	projName := path.Base(currentDir)

	f := &confile.File{
		Project: &confile.Project{
			Name: projName,
		},
		Steps: map[string]*confile.Step{
			"build": {
				Help: "build the project",
				Cmd:  "echo \"run the build\"",
			},
		},
	}

	err = f.Fmt()
	if err != nil {
		fmt.Printf("Error Initializing File: %v", err.Error())
	}
}
