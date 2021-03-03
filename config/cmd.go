package config

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

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

	return out.String(), nil
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

// MarshalYAML meets the marshaller interface for the yaml library being used.
func (c Cmd) MarshalYAML() (interface{}, error) {
	cmd := make([]string, 0, len(c.Args)+1)
	cmd = append(cmd, c.Command)
	cmd = append(cmd, c.Args...)

	return strings.Join(cmd, " "), nil
}
