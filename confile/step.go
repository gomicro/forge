package confile

import (
	"fmt"
	"os"
	"os/exec"
)

// Step represents details of single step to be executed by the cli.
type Step struct {
	Cmd   string            `yaml:"cmd,omitempty"`
	Cmds  []string          `yaml:"cmds,omitempty"`
	Envs  map[string]string `yaml:"envs,omitempty"`
	Help  string            `yaml:"help,omitempty"`
	Post  []string          `yaml:"post,omitempty"`
	Pre   []string          `yaml:"pre,omitempty"`
	Steps []string          `yaml:"steps,omitempty"`

	projectEnvs map[string]string
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute(allSteps map[string]*Step, projectEnvs map[string]string) error {
	s.projectEnvs = projectEnvs

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
	return executeCmd(s.Cmd, s.Envs, s.projectEnvs)
}

func (s *Step) executeCmds() error {
	for _, c := range s.Cmds {
		err := executeCmd(c, s.Envs, s.projectEnvs)
		if err != nil {
			return fmt.Errorf("cmds: cmd exec: %v", err.Error())
		}
	}

	return nil
}

func (s *Step) executeSteps(execList []string, allSteps map[string]*Step) error {
	for _, step := range execList {
		s, ok := allSteps[step]
		if !ok {
			return fmt.Errorf("step does not exist: %v", step)
		}

		err := s.Execute(allSteps, s.projectEnvs)
		if err != nil {
			return err
		}
	}

	return nil
}

func executeCmd(cmdString string, stepEnvs, projectEnvs map[string]string) error {
	cmd := exec.Command("bash", "-c", cmdString)

	cmd.Env = toSlice(stepEnvs)
	cmd.Env = append(cmd.Env, toSlice(projectEnvs)...)
	cmd.Env = append(cmd.Env, os.Environ()...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("execute: %v", err.Error())
	}

	return cmd.Wait()
}

func toSlice(e map[string]string) []string {
	out := []string{}

	for k, v := range e {
		out = append(out, fmt.Sprintf("%v=%v", k, v))
	}

	return out
}
