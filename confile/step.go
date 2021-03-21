package confile

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

// Step represents details of single step to be executed by the cli.
type Step struct {
	Help  string   `yaml:"help,omitempty"`
	Pre   []string `yaml:"pre,omitempty"`
	Post  []string `yaml:"post,omitempty"`
	Cmd   string   `yaml:"cmd,omitempty"`
	Cmds  []string `yaml:"cmds,omitempty"`
	Steps []string `yaml:"steps,omitempty"`
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute(allSteps map[string]*Step) error {
	if len(s.Pre) > 0 {
		err := s.executeSteps(s.Pre, allSteps)
		if err != nil {
			return fmt.Errorf("step: execute pre: %v", err.Error())
		}
	}

	if len(s.Steps) > 0 {
		err := s.executeSteps(s.Steps, allSteps)
		if err != nil {
			return fmt.Errorf("step: execute steps: %v", err.Error())
		}
	} else if len(s.Cmds) > 0 {
		err := s.executeCmds()
		if err != nil {
			return fmt.Errorf("step: execute cmds: %v", err.Error())
		}
	} else {
		err := s.executeCmd()
		if err != nil {
			return fmt.Errorf("step: execute cmd: %v", err.Error())
		}
	}

	if len(s.Post) > 0 {
		err := s.executeSteps(s.Post, allSteps)
		if err != nil {
			return fmt.Errorf("step: execute post: %v", err.Error())
		}
	}

	return nil
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

func (s *Step) executeSteps(execList []string, allSteps map[string]*Step) error {
	for _, step := range execList {
		err := allSteps[step].Execute(allSteps)
		if err != nil {
			return err
		}
	}

	return nil
}

func executeCmd(cmdString string) error {
	cmd := exec.Command("bash", "-c", cmdString)

	g := errgroup.Group{}

	out, _ := cmd.StdoutPipe()
	g.Go(func() error {
		defer out.Close()
		_, err := io.Copy(os.Stdout, out)
		if err != nil {
			return fmt.Errorf("stdout copy: %v", err.Error())
		}

		return nil
	})

	errout, _ := cmd.StderrPipe()
	g.Go(func() error {
		defer errout.Close()
		_, err := io.Copy(os.Stderr, errout)
		if err != nil {
			return fmt.Errorf("stderr copy: %v", err.Error())
		}

		return nil
	})

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("execute: %v", err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("execute: %v", err.Error())
	}

	return g.Wait()
}
