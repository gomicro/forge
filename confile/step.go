package confile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

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
	Stdin       io.Reader
	Stderr      io.Writer
	SkipPre     bool
	SkipPost    bool
	Recursive   bool
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
	childOptions := options
	if !options.Recursive {
		childOptions.SkipPre = false
		childOptions.SkipPost = false
	}

	for _, stepName := range execList {
		step, ok := options.Steps[stepName]
		if !ok {
			return fmt.Errorf("step does not exist: %v", stepName)
		}

		err := step.Execute(ctx, stepName, childOptions)
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
	cmd.Stdin = options.Stdin
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}

	configEnvs := toSlice(options.ProjectEnvs)
	configEnvs = append(configEnvs, toSlice(stepEnvs)...)

	for i := range configEnvs {
		configEnvs[i] = options.Vars.Process(configEnvs[i])
	}

	// exec keeps the last value for duplicate keys.
	cmd.Env = append(os.Environ(), configEnvs...)

	var outputMu sync.Mutex
	stdout := &lineWriter{
		writeLine: func(line []byte) error {
			outputMu.Lock()
			defer outputMu.Unlock()
			scrb.PrintLines(bytes.NewBuffer(line))
			return nil
		},
	}
	if !options.Verbose {
		stdout = nil
	}

	stderrWriter := options.Stderr
	if stderrWriter == nil {
		stderrWriter = os.Stderr
	}
	stderrHeaderWritten := false
	stderr := &lineWriter{
		writeLine: func(line []byte) error {
			outputMu.Lock()
			defer outputMu.Unlock()

			if options.Verbose {
				scrb.BeginDescribe("\033[1;31mstderr\033[0m")
				scrb.PrintLines(bytes.NewBuffer(line))
				scrb.EndDescribe()
				return nil
			}

			if !stderrHeaderWritten {
				_, err := fmt.Fprintln(stderrWriter, "\033[1;31mstderr\033[0m")
				if err != nil {
					return fmt.Errorf("writing stderr header: %w", err)
				}
				stderrHeaderWritten = true
			}

			_, err := stderrWriter.Write(line)
			if err != nil {
				return fmt.Errorf("writing command stderr: %w", err)
			}

			return nil
		},
	}
	if options.Verbose {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = io.Discard
	}
	cmd.Stderr = stderr

	err := cmd.Run()
	if stdout != nil {
		err = errors.Join(err, stdout.Flush())
	}
	err = errors.Join(err, stderr.Flush())
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}

type lineWriter struct {
	mu        sync.Mutex
	buffer    []byte
	writeLine func([]byte) error
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	written := len(p)
	w.buffer = append(w.buffer, p...)
	for {
		i := bytes.IndexByte(w.buffer, '\n')
		if i < 0 {
			return written, nil
		}

		err := w.writeLine(w.buffer[:i+1])
		if err != nil {
			return 0, err
		}

		w.buffer = w.buffer[i+1:]
	}
}

func (w *lineWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buffer) == 0 {
		return nil
	}

	err := w.writeLine(w.buffer)
	if err != nil {
		return err
	}

	w.buffer = nil

	return nil
}

func toSlice(e map[string]string) []string {
	out := []string{}

	for k, v := range e {
		out = append(out, fmt.Sprintf("%v=%v", k, v))
	}

	return out
}
