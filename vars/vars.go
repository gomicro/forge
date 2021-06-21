package vars

import (
	"regexp"
	"strings"
)

var (
	match = regexp.MustCompile(`\{\{\ ?\.[a-zA-Z][a-zA-Z0-9]*\ ?\}\}`)
)

// Vars represents a set of known variables to replace in a given command
type Vars map[string]string

// Process takes a template string and replaces all instances of known variabless
// stored in the Vars set with the known values. Any variables not in the Vals
// set will be ignored.
func (v *Vars) Process(template string) string {
	out := template
	vals := match.FindAllString(template, -1)

	for i := range vals {
		key := strings.TrimRight(strings.TrimLeft(vals[i], "{. "), "}")
		val, found := map[string]string(*v)[key]
		if !found {
			continue
		}

		out = strings.ReplaceAll(out, vals[i], val)
	}

	return out
}

// Set takes a key value pair and writes it to the Vals set.
func (v *Vars) Set(key, value string) {
	map[string]string(*v)[key] = value
}
