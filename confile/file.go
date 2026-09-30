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

const (
	file = "./forge.yaml"
)

// File represents the build options for a project
type File struct {
	Project *Project          `yaml:"project"`
	Envs    map[string]string `yaml:"envs,omitempty"`
	Steps   map[string]*Step  `yaml:"steps"`
	Vars    *vars.Vars        `yaml:"-"`
}

// ParseFromFile reads an Forge config file from the from the current directory.
// A File with the populated values is returned and any errors encountered while
// trying to read the file.
func ParseFromFile() (*File, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("parseFromFile: reading config file: %w", err)
	}

	conf, err := parse(b)
	if err != nil {
		return nil, fmt.Errorf("parseFromFile: %w", err)
	}

	shaBytes, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("parseFromFile: getting sha: %w", err)
	}

	branchBytes, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return nil, fmt.Errorf("parseFromFile: getting branch: %w", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("parseFromFile: getting working directory: %w", err)
	}

	conf.Vars = builtinVars(conf.Project.Name, string(shaBytes), string(branchBytes), currentDir)

	return conf, nil
}

func parse(b []byte) (*File, error) {
	var conf File
	err := yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config file: %w", err)
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

// Fmt marshals the config file struct into yaml and overwrites the original
// config file in the current directory. It returns any errors it encounters.
func (f *File) Fmt() error {
	b, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("fmt: marshaling: %w", err)
	}

	err = os.WriteFile(file, b, 0644)
	if err != nil {
		return fmt.Errorf("fmt: writing file: %w", err)
	}

	return nil
}

// Exists checks whether or not the preferred config file exists or not. It
// returns true if the file exists, and false if the file doesn't exist.
func Exists() bool {
	_, err := os.Stat(file)
	return !errors.Is(err, fs.ErrNotExist)
}
