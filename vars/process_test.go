package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarsProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vars     map[string]string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "replaces known vars",
			vars:     map[string]string{"ID": "af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42"},
			input:    "curl http://do.the.thing/pie/{{.ID}}",
			contains: []string{"af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42"},
			excludes: []string{"{{.ID}}"},
		},
		{
			name:     "leaves unknown vars untouched",
			vars:     map[string]string{"ID": "af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42"},
			input:    "curl http://do.the.thing/pie/{{.ID}}/{{.Subid}}",
			contains: []string{"af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42", "{{.Subid}}"},
			excludes: []string{"{{.ID}}"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := Vars(tt.vars)
			out := v.Process(tt.input)

			for _, s := range tt.contains {
				assert.Contains(t, out, s)
			}
			for _, s := range tt.excludes {
				assert.NotContains(t, out, s)
			}
		})
	}
}
