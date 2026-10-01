package confile

import (
	"fmt"
	"sort"
	"strings"
)

// validate checks the step graph before anything runs, so a bad config fails
// up front instead of partway through execution. It returns every problem found.
func validate(f *File) []string {
	names := sortedStepNames(f.Steps)

	var problems []string

	for _, name := range names {
		problems = append(problems, validateStep(name, f.Steps)...)
	}

	problems = append(problems, findCycles(names, f.Steps)...)

	return problems
}

func validateStep(name string, steps map[string]*Step) []string {
	step := steps[name]
	if step == nil {
		return []string{fmt.Sprintf("step '%s': must define one of cmd, cmds, or steps", name)}
	}

	var problems []string

	var actions []string
	if step.Cmd != "" {
		actions = append(actions, "cmd")
	}

	if len(step.Cmds) > 0 {
		actions = append(actions, "cmds")
	}

	if len(step.Steps) > 0 {
		actions = append(actions, "steps")
	}

	switch {
	case len(actions) == 0:
		problems = append(problems, fmt.Sprintf("step '%s': must define one of cmd, cmds, or steps", name))

	case len(actions) > 1:
		problems = append(problems, fmt.Sprintf("step '%s': only one of cmd, cmds, or steps may be defined, found %s", name, strings.Join(actions, ", ")))
	}

	refLists := []struct {
		field string
		refs  []string
	}{
		{field: "pre", refs: step.Pre},
		{field: "steps", refs: step.Steps},
		{field: "post", refs: step.Post},
	}

	for _, list := range refLists {
		for _, ref := range list.refs {
			_, found := steps[ref]
			if !found {
				problems = append(problems, fmt.Sprintf("step '%s': %s references undefined step '%s'", name, list.field, ref))
			}
		}
	}

	return problems
}

const (
	unvisited = iota
	visiting
	visited
)

// findCycles walks every pre, steps, and post edge depth-first and reports each
// back edge as a cycle. Names are walked in sorted order so output is stable.
func findCycles(names []string, steps map[string]*Step) []string {
	state := map[string]int{}

	var problems []string

	for _, name := range names {
		if state[name] != unvisited {
			continue
		}

		var path []string
		problems = append(problems, walkCycles(name, steps, state, path)...)
	}

	return problems
}

func walkCycles(name string, steps map[string]*Step, state map[string]int, path []string) []string {
	state[name] = visiting
	path = append(path, name)

	var problems []string

	for _, next := range stepEdges(steps[name]) {
		_, found := steps[next]
		if !found {
			continue
		}

		switch state[next] {
		case visiting:
			problems = append(problems, "cycle detected: "+describeCycle(path, next))

		case unvisited:
			problems = append(problems, walkCycles(next, steps, state, path)...)
		}
	}

	state[name] = visited

	return problems
}

func stepEdges(step *Step) []string {
	if step == nil {
		return nil
	}

	var edges []string
	edges = append(edges, step.Pre...)
	edges = append(edges, step.Steps...)
	edges = append(edges, step.Post...)

	return edges
}

// describeCycle renders the portion of path that starts at target, closing the
// loop back to target.
func describeCycle(path []string, target string) string {
	start := 0
	for i, name := range path {
		if name == target {
			start = i
			break
		}
	}

	loop := make([]string, 0, len(path)-start+1)
	loop = append(loop, path[start:]...)
	loop = append(loop, target)

	return strings.Join(loop, " -> ")
}

func sortedStepNames(steps map[string]*Step) []string {
	names := make([]string, 0, len(steps))
	for name := range steps {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
