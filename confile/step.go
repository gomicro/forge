package confile

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/gomicro/forge/vars"
	"github.com/gomicro/scribe"
	"github.com/spf13/viper"
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
	vars        *vars.Vars
}

// Execute runs the command that is specified for the step. It returns the output
// of the command and any errors it encounters.
func (s *Step) Execute(name string, allSteps map[string]*Step, projectEnvs map[string]string, vars *vars.Vars, scrb scribe.Scriber) error {
	skipPre := viper.GetBool("solo") || viper.GetBool("no-pre")
	skipPost := viper.GetBool("solo") || viper.GetBool("no-post")

	s.projectEnvs = projectEnvs
	s.vars = vars

	if len(s.Pre) > 0 && !skipPre {
		scrb.BeginDescribe(name + ": pre")
		err := s.executeSteps(s.Pre, allSteps, scrb)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute pre: %w", err)
		}
	}

	if len(s.Steps) > 0 {
		scrb.BeginDescribe(name)
		err := s.executeSteps(s.Steps, allSteps, scrb)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute steps: %w", err)
		}
	} else if len(s.Cmds) > 0 {
		scrb.BeginDescribe(name)
		err := s.executeCmds(scrb)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute cmds: %w", err)
		}
	} else {
		scrb.BeginDescribe(name)
		err := s.executeCmd(scrb)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute cmd: %w", err)
		}
	}

	if len(s.Post) > 0 && !skipPost {
		scrb.BeginDescribe(name + ": post")
		err := s.executeSteps(s.Post, allSteps, scrb)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute post: %w", err)
		}
	}

	return nil
}

func (s *Step) executeCmd(scrb scribe.Scriber) error {
	cmdString := s.vars.Process(s.Cmd)
	return executeCmd(cmdString, s.Envs, s.projectEnvs, s.vars, scrb)
}

func (s *Step) executeCmds(scrb scribe.Scriber) error {
	for _, c := range s.Cmds {
		cmdString := s.vars.Process(c)

		err := executeCmd(cmdString, s.Envs, s.projectEnvs, s.vars, scrb)
		if err != nil {
			return fmt.Errorf("cmds: cmd exec: %w", err)
		}
	}

	return nil
}

func (s *Step) executeSteps(execList []string, allSteps map[string]*Step, scrb scribe.Scriber) error {
	for _, stepName := range execList {
		step, ok := allSteps[stepName]
		if !ok {
			return fmt.Errorf("step does not exist: %v", stepName)
		}

		err := step.Execute(stepName, allSteps, step.projectEnvs, step.vars, scrb)
		if err != nil {
			return err
		}
	}

	return nil
}

func executeCmd(cmdString string, stepEnvs, projectEnvs map[string]string, vars *vars.Vars, scrb scribe.Scriber) error {
	scrb.Print(fmt.Sprintf("$ %s", cmdString))

	cmd := exec.Command("bash", "-c", cmdString)

	cmd.Env = toSlice(stepEnvs)
	cmd.Env = append(cmd.Env, toSlice(projectEnvs)...)

	for i := range cmd.Env {
		cmd.Env[i] = vars.Process(cmd.Env[i])
	}

	cmd.Env = append(cmd.Env, os.Environ()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	waitErr := cmd.Wait()

	if stdout.Len() > 0 {
		scrb.PrintLines(&stdout)
	}

	if stderr.Len() > 0 {
		if viper.GetBool("verbose") {
			scrb.BeginDescribe("\033[1;31mstderr\033[0m")
			scrb.PrintLines(&stderr)
			scrb.EndDescribe()
		} else {
			fmt.Fprintf(os.Stderr, "\033[1;31mstderr\033[0m\n%s", stderr.String())
		}
	}

	if waitErr != nil {
		return fmt.Errorf("execute: %w", waitErr)
	}

	return nil
}

func toSlice(e map[string]string) []string {
	out := []string{}

	for k, v := range e {
		out = append(out, fmt.Sprintf("%v=%v", k, v))
	}

	return out
}
