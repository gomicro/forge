package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gomicro/forge/confile"
	"github.com/spf13/cobra"
)

func init() {
	ConfigCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a forge.yaml in the current directory",
	Long:  `Scaffold a forge.yaml using the current directory name as the project name, with a sample build step included.`,
	RunE:  initFunc,
}

func initFunc(cmd *cobra.Command, args []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working dir: %w", err)
	}

	err = initializeConfig(confile.DefaultPath, currentDir)
	if err != nil {
		return fmt.Errorf("initializing config file: %w", err)
	}

	return nil
}

func initializeConfig(configPath, currentDir string) error {
	if confile.Exists(configPath) {
		return errors.New("config file already exists")
	}

	projName := filepath.Base(currentDir)
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

	err := f.Fmt(configPath)
	if err != nil {
		return fmt.Errorf("formatting initialized file: %w", err)
	}

	return nil
}
