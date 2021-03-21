package confile

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Step represents details of single step to be executed by the cli.
type Step struct {
	Help string   `yaml:"help,omitempty"`
	Cmd  string   `yaml:"cmd,omitempty"`
	Cmds []string `yaml:"cmds,omitempty"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute() error {
	if len(s.Cmds) > 0 {
		return s.executeCmds()
	}

	return s.executeCmd()
}

func (s *Step) executeCmd() error {
	return executeCmd(s.Cmd)
}

func (s *Step) executeCmds() error {
	for _, c := range s.Cmds {
		err := executeCmd(c)
		if err != nil {
			return fmt.Errorf("cmds: cmd exec: %v", err.Error())
		}
	}

	return nil
}

func executeCmd(cmdString string) error {
	cmd := exec.Command("bash", "-c", cmdString)

	out, _ := cmd.StdoutPipe()
	go func() {
		defer out.Close()
		io.Copy(os.Stdout, out)
	}()

	errout, _ := cmd.StderrPipe()
	go func() {
		defer errout.Close()
		io.Copy(os.Stderr, errout)
	}()

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("execute: %v", err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("execute: %v", err.Error())
	}

	return nil
}
