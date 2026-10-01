package confile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gomicro/forge/vars"
	"gopkg.in/yaml.v3"
)

// DefaultPath is the config file forge reads, relative to the working directory.
const DefaultPath = "forge.yaml"

// File represents the build options for a project
type File struct {
	Project *Project          `yaml:"project"`
	Envs    map[string]string `yaml:"envs,omitempty"`
	Steps   map[string]*Step  `yaml:"steps"`
	Vars    *vars.Vars        `yaml:"-"`
}

// Parse reads and validates the config file at path. It does not consult git,
// so Vars is nil until ResolveVars is called.
func Parse(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	conf, err := parse(b)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return conf, nil
}

// ResolveVars populates Vars with the built-in template variables, reading git
// metadata from the repository containing dir.
func (f *File) ResolveVars(dir string) error {
	sha, err := gitOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("resolveVars: getting sha: %w", err)
	}

	branch, err := gitOutput(dir, "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("resolveVars: getting branch: %w", err)
	}

	f.Vars = builtinVars(f.Project.Name, sha, branch, dir)

	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		// %v rather than %w: git's exit status must not be mistaken for a step's.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("running git %v: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(exitErr.Stderr)))
		}

		return "", fmt.Errorf("running git %v: %v", strings.Join(args, " "), err)
	}

	return string(out), nil
}

func parse(b []byte) (*File, error) {
	var conf File
	err := yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling: %w", err)
	}

	if conf.Project == nil {
		return nil, errors.New("missing required project block")
	}

	for name, step := range conf.Steps {
		for _, s := range step.Steps {
			if strings.EqualFold(s, name) {
				return nil, fmt.Errorf("infinite loop detected: step '%v'", name)
			}
		}
	}

	return &conf, nil
}

// builtinVars returns the template variables every step can reference. Raw git
// output is accepted as-is and trimmed here.
func builtinVars(project, sha, branch, dir string) *vars.Vars {
	sha = strings.TrimSpace(sha)
	branch = strings.TrimSpace(branch)

	shortSha := sha
	if len(shortSha) > 7 {
		shortSha = shortSha[:7]
	}

	v := &vars.Vars{}
	v.Set("Branch", branch)
	v.Set("Dir", dir)
	v.Set("Os", runtime.GOOS)
	v.Set("Project", project)
	v.Set("Sha", sha)
	v.Set("ShortSha", shortSha)

	return v
}

// Fmt marshals the config into yaml and writes it to path, replacing any
// existing file.
func (f *File) Fmt(path string) error {
	b, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("fmt: marshaling: %w", err)
	}

	err = os.WriteFile(path, b, 0644)
	if err != nil {
		return fmt.Errorf("fmt: writing file: %w", err)
	}

	return nil
}

// Exists reports whether a config file is present at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}
