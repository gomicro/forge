package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/gomicro/forge/confile"
)

func init() {
	RootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a forge.yaml in the current directory",
	Long:  `Scaffold a forge.yaml using the current directory name as the project name, with a sample build step included.`,
	RunE:  initFunc,
}

func initFunc(cmd *cobra.Command, args []string) error {
	if confile.Exists() {
		return errors.New("config file already exists")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working dir: %w", err)
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
		return fmt.Errorf("initializing file: %w", err)
	}

	return nil
}
