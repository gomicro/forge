package confile

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	file = "./forge.yaml"
)

// File represents the build options for a project
type File struct {
	Project *Project         `yaml:"project"`
	Steps   map[string]*Step `yaml:"steps"`
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

	return &conf, nil
}

// Fmt marshals the config file struct into yaml and overwrites the original
// config file in the current directory. It returns any errors it encounters.
func (f *File) Fmt() error {
	b, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("fmt: marshal: %v", err.Error())
	}

	err = ioutil.WriteFile(file, b, 644)
	if err != nil {
		return fmt.Errorf("fmt: write file: %v", err.Error())
	}

	return nil
}
