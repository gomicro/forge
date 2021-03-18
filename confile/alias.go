package confile

import (
	"strings"
)

// Alias represents an alias of one or many steps to be executed by the cli.
type Alias struct {
	Help  string   `yaml:"help,omitempty"`
	Steps []string `yaml:"steps"`
}

// Execute runs the steps that are specified for the alias. It returns the
// output of the steps executed and the first, if any, errors it encounters.
func (a *Alias) Execute(steps map[string]*Step) (string, error) {
	var outs []string
	var err error

	for _, s := range a.Steps {
		var out string
		out, err = steps[s].Execute()
		if err != nil {
			break
		}

		outs = append(outs, out)
	}

	return strings.Join(outs, "\n"), nil
}
