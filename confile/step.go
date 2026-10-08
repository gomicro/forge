package confile

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gomicro/forge/vars"

	"github.com/gomicro/scribe"
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
}

// RunOptions carries shared execution settings through the entire step graph.
type RunOptions struct {
	Steps       map[string]*Step
	ProjectEnvs map[string]string
	Vars        *vars.Vars
	Scriber     scribe.Scriber
	SkipPre     bool
	SkipPost    bool
	Verbose     bool
}

// Execute runs the step and its dependencies using the supplied settings.
func (s *Step) Execute(ctx context.Context, name string, options RunOptions) error {
	scrb := options.Scriber

	if len(s.Pre) > 0 && !options.SkipPre {
		scrb.BeginDescribe(name + ": pre")
		err := s.executeSteps(ctx, s.Pre, options)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute pre: %w", err)
		}
	}

	if len(s.Steps) > 0 {
		scrb.BeginDescribe(name)
		err := s.executeSteps(ctx, s.Steps, options)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute steps: %w", err)
		}
	} else if len(s.Cmds) > 0 {
		scrb.BeginDescribe(name)
		err := s.executeCmds(ctx, options)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute cmds: %w", err)
		}
	} else {
		scrb.BeginDescribe(name)
		err := executeCmd(ctx, s.Cmd, s.Envs, options)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute cmd: %w", err)
		}
	}

	if len(s.Post) > 0 && !options.SkipPost {
		scrb.BeginDescribe(name + ": post")
		err := s.executeSteps(ctx, s.Post, options)
		scrb.EndDescribe()
		if err != nil {
			return fmt.Errorf("step: execute post: %w", err)
		}
	}

	return nil
}

func (s *Step) executeCmds(ctx context.Context, options RunOptions) error {
	for _, c := range s.Cmds {
		err := executeCmd(ctx, c, s.Envs, options)
		if err != nil {
			return fmt.Errorf("cmds: cmd exec: %w", err)
		}
	}

	return nil
}

func (s *Step) executeSteps(ctx context.Context, execList []string, options RunOptions) error {
	for _, stepName := range execList {
		step, ok := options.Steps[stepName]
		if !ok {
			return fmt.Errorf("step does not exist: %v", stepName)
		}

		err := step.Execute(ctx, stepName, options)
		if err != nil {
			return fmt.Errorf("executeSteps: executing step %s: %w", stepName, err)
		}
	}

	return nil
}

func executeCmd(ctx context.Context, command string, stepEnvs map[string]string, options RunOptions) error {
	cmdString := options.Vars.Process(command)
	scrb := options.Scriber
	scrb.Print(fmt.Sprintf("$ %s", cmdString))

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdString)

	cmd.Env = toSlice(stepEnvs)
	cmd.Env = append(cmd.Env, toSlice(options.ProjectEnvs)...)

	for i := range cmd.Env {
		cmd.Env[i] = options.Vars.Process(cmd.Env[i])
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
		if options.Verbose {
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
