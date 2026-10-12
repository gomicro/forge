package confile

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var topLevelOrder = map[string]int{
	"project": 0,
	"envs":    1,
	"steps":   2,
}

var stepFieldOrder = map[string]int{
	"help":  0,
	"envs":  1,
	"pre":   2,
	"steps": 3,
	"cmds":  4,
	"cmd":   5,
	"post":  6,
}

// Format rewrites the YAML config at path into the canonical format.
func Format(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	_, err = parse(content)
	if err != nil {
		return fmt.Errorf("validating config: %w", err)
	}

	formatted, err := formatConfig(content)
	if err != nil {
		return fmt.Errorf("formatting config file: %w", err)
	}

	err = os.WriteFile(path, formatted, 0644)
	if err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func formatConfig(content []byte) ([]byte, error) {
	var doc yaml.Node
	err := yaml.Unmarshal(content, &doc)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config must be a YAML mapping")
	}

	root := doc.Content[0]
	sortMapping(root, topLevelOrder)

	steps := mappingValue(root, "steps")
	if steps != nil && steps.Kind == yaml.MappingNode {
		sortMapping(steps, stepFieldOrder)
		for i := 0; i+1 < len(steps.Content); i += 2 {
			step := steps.Content[i+1]
			if step.Kind == yaml.MappingNode {
				sortMapping(step, stepFieldOrder)
			}
		}
	}

	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)

	err = encoder.Encode(&doc)
	if err != nil {
		return nil, fmt.Errorf("encoding config: %w", err)
	}

	err = encoder.Close()
	if err != nil {
		return nil, fmt.Errorf("closing config encoder: %w", err)
	}

	return addConfigSpacing(output.Bytes(), steps), nil
}

func sortMapping(node *yaml.Node, preferredOrder map[string]int) {
	type pair struct {
		key   *yaml.Node
		value *yaml.Node
	}

	pairs := make([]pair, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		pairs = append(pairs, pair{key: node.Content[i], value: node.Content[i+1]})
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		left, leftFound := preferredOrder[pairs[i].key.Value]
		right, rightFound := preferredOrder[pairs[j].key.Value]
		if leftFound || rightFound {
			if !leftFound {
				return false
			}
			if !rightFound {
				return true
			}
			return left < right
		}

		return pairs[i].key.Value < pairs[j].key.Value
	})

	node.Content = node.Content[:0]
	for _, item := range pairs {
		node.Content = append(node.Content, item.key, item.value)
	}
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

func addConfigSpacing(content []byte, steps *yaml.Node) []byte {
	lines := strings.SplitAfter(string(content), "\n")
	stepNames := map[string]struct{}{}
	if steps != nil && steps.Kind == yaml.MappingNode {
		for i := 0; i < len(steps.Content); i += 2 {
			stepNames[steps.Content[i].Value] = struct{}{}
		}
	}

	breaks := map[int]struct{}{}
	seenTopLevel := false
	seenSteps := false
	previousStep := false
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\r\n")
		indent := len(trimmed) - len(strings.TrimLeft(trimmed, " "))

		if indent == 0 && isMappingKey(trimmed) {
			key := mappingKey(trimmed)
			if seenTopLevel {
				breaks[commentGroupStart(lines, i, indent)] = struct{}{}
			}
			seenTopLevel = true
			seenSteps = key == "steps"
			previousStep = false
		} else if seenSteps && indent == 2 && isMappingKey(trimmed) {
			if _, found := stepNames[mappingKey(trimmed)]; found {
				if previousStep {
					breaks[commentGroupStart(lines, i, indent)] = struct{}{}
				}
				previousStep = true
			}
		}
	}

	var output strings.Builder
	for i, line := range lines {
		if _, found := breaks[i]; found {
			output.WriteByte('\n')
		}

		output.WriteString(line)
	}

	return []byte(output.String())
}

func commentGroupStart(lines []string, keyLine, indent int) int {
	start := keyLine
	for start > 0 {
		previous := strings.TrimRight(lines[start-1], "\r\n")
		previousIndent := len(previous) - len(strings.TrimLeft(previous, " "))
		if previousIndent != indent || !strings.HasPrefix(strings.TrimSpace(previous), "#") {
			break
		}

		start--
	}

	return start
}

func isMappingKey(line string) bool {
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return false
	}

	return strings.Contains(line, ":")
}

func mappingKey(line string) string {
	key, _, found := strings.Cut(strings.TrimSpace(line), ":")
	if !found {
		return ""
	}

	return strings.Trim(key, `"'`)
}
