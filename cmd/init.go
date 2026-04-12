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
	Short: "Initialize a forge config file",
	Long:  `Collects the basic information to initialize a forge config file.`,
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
