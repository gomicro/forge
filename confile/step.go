package confile

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Step represents details of single step to be executed by the cli.
type Step struct {
	Help string `yaml:"help,omitempty"`
	Cmd  string `yaml:"cmd"`
	//Cmds []string `yaml:"cmds"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute() (string, error) {
	cmd := exec.Command("bash", "-c", s.Cmd)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cmd exec: %v", err.Error())
	}

	return out.String(), nil
}
