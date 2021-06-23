package confile

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
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

// ParseFromFile reads an Forge config file from the from the curent directory.
// A File with the populated values is returned and any errors encountered while
// trying to read the file.
func ParseFromFile() (*File, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %v", err.Error())
	}

	var conf File
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config file: %v", err.Error())
	}

	for name, step := range conf.Steps {
		if len(step.Steps) > 0 {
			infloop := false
			for _, s := range step.Steps {
				if strings.EqualFold(s, name) {
					infloop = true
				}
			}

			if infloop {
				return nil, fmt.Errorf("infinite loop detected: step '%v'", name)
			}
		}
	}

	vars := &vars.Vars{}

	shaBytes, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to get sha: %v", err.Error())
	}

	branchBytes, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to get branch: %v", err.Error())
	}

	vars.Set("Project", conf.Project.Name)
	vars.Set("Os", runtime.GOOS)
	vars.Set("Sha", string(shaBytes))
	vars.Set("ShortSha", string(shaBytes)[:7])
	vars.Set("Branch", string(branchBytes))

	conf.Vars = vars

	return &conf, nil
}

// Fmt marshals the config file struct into yaml and overwrites the original
// config file in the current directory. It returns any errors it encounters.
func (f *File) Fmt() error {
	b, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("fmt: marshal: %v", err.Error())
	}

	err = ioutil.WriteFile(file, b, 0644)
	if err != nil {
		return fmt.Errorf("fmt: write file: %v", err.Error())
	}

	return nil
}

// Exists checks whether or not the preferred config file exists or not. It
// returns true if the file exists, and false if the file doesn't exist.
func Exists() bool {
	_, err := os.Stat(file)
	return !errors.Is(err, fs.ErrNotExist)
}
