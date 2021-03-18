package confile

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Step represents details of single step to be executed by the cli.
type Step struct {
	Help string   `yaml:"help,omitempty"`
	Cmd  string   `yaml:"cmd,omitempty"`
	Cmds []string `yaml:"cmds,omitempty"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute() (string, error) {
	if len(s.Cmds) > 0 {
		return s.executeCmds()
	}

	return s.executeCmd()
}

func (s *Step) executeCmd() (string, error) {
	return executeCmd(s.Cmd)
}

func (s *Step) executeCmds() (string, error) {
	outs := make([]string, 0, len(s.Cmds))
	for _, c := range s.Cmds {
		out, err := executeCmd(c)
		if err != nil {
			return "", fmt.Errorf("cmds: cmd exec: %v", err.Error())
		}

		if out != "" {
			outs = append(outs, out)
		}
	}
	return strings.Join(outs, "\n"), nil
}

func executeCmd(cmdString string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdString)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cmd exec: %v", err.Error())
	}

	return out.String(), nil
}
