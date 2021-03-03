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

// ParseFromFile reads an Duty config file from the file specified in the
// environment or from the default file location if no environment is specified.
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

type Project struct {
	Name string `yaml:"name"`
}

type Step struct {
	Cmd Cmd `yaml:"cmd"`
	//Cmds []string `yaml:"cmds"`
}

func (s *Step) Execute() (string, error) {
	return s.Cmd.Execute()
}

type Cmd struct {
	Command string
	Args    []string
}

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
