package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

// File represents the build options for a project
type File struct {
	Project *Project         `yaml:"project"`
	Steps   map[string]*Step `yaml:"steps"`
}

// ParseFromFile reads an Forge config file from the from the curent directory.
// A File with the populated values is returned and any errors encountered while
// trying to read the file.
func ParseFromFile() (*File, error) {
	b, err := ioutil.ReadFile("./forge.yaml")
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %v", err.Error())
	}

	var conf File
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config file: %v", err.Error())
	}

	return &conf, nil
}

// Project represents metadata about the project being built that can be
// included in the config file.
type Project struct {
	Name string `yaml:"name"`
}

// Step represents details of single step to be executed by the cli.
type Step struct {
	Cmd Cmd `yaml:"cmd"`
	//Cmds []string `yaml:"cmds"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute() (string, error) {
	return s.Cmd.Execute()
}

// Cmd represents a custom unmarshallable entity that is broken down into an
// executable command line entry.
type Cmd struct {
	Command string
	Args    []string
}

// Execute runs the command defined. It returns the output of the command and
// any errors it encounters.
func (c *Cmd) Execute() (string, error) {
	cmd := exec.Command(c.Command, c.Args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cmd exec: %v", err.Error())
	}

	return string(out.Bytes()), nil
}

// UnmarshalYAML meets the unmarshaller interface for the yaml library being
// used. It parses and splits the command string into an executable shell
// command.
func (c *Cmd) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmdStr string

	err := unmarshal(&cmdStr)
	if err != nil {
		return err
	}

	parts := strings.Split(cmdStr, " ")
	if len(parts) > 0 {
		c.Command = parts[0]

		if len(parts) > 1 {
			c.Args = parts[1:]
		}
	}

	return nil
}
