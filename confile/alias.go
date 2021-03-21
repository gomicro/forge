package confile

import (
	"fmt"
)

// Alias represents an alias of one or many steps to be executed by the cli.
type Alias struct {
	Help  string   `yaml:"help,omitempty"`
	Steps []string `yaml:"steps"`
}

// Execute runs the steps that are specified for the alias. It returns the
// output of the steps executed and the first, if any, errors it encounters.
func (a *Alias) Execute(steps map[string]*Step) error {
	for _, s := range a.Steps {
		err := steps[s].Execute()
		if err != nil {
			return fmt.Errorf("alias: execute: step: %v: %v", s, err.Error())
		}
	}

	return nil
}
