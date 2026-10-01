package confile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		steps   string
		wantErr string
	}{
		{
			name: "valid graph with hooks and groups",
			steps: `
  build: {cmd: go build}
  lint: {cmds: [go vet, golangci-lint run]}
  test: {pre: [build], cmd: go test, post: [lint]}
  all: {steps: [build, test]}
`,
		},
		{
			name: "empty step",
			steps: `
  build: {help: does nothing}
`,
			wantErr: "invalid config:\n  - step 'build': must define one of cmd, cmds, or steps",
		},
		{
			name: "null step",
			steps: `
  build:
`,
			wantErr: "invalid config:\n  - step 'build': must define one of cmd, cmds, or steps",
		},
		{
			name: "cmd and cmds together",
			steps: `
  build: {cmd: a, cmds: [b]}
`,
			wantErr: "invalid config:\n  - step 'build': only one of cmd, cmds, or steps may be defined, found cmd, cmds",
		},
		{
			name: "steps alongside cmd",
			steps: `
  build: {cmd: a, steps: [other]}
  other: {cmd: b}
`,
			wantErr: "invalid config:\n  - step 'build': only one of cmd, cmds, or steps may be defined, found cmd, steps",
		},
		{
			name: "missing references across steps pre and post",
			steps: `
  build: {pre: [gen], steps: [compile], post: [ship]}
`,
			wantErr: "invalid config:\n" +
				"  - step 'build': pre references undefined step 'gen'\n" +
				"  - step 'build': steps references undefined step 'compile'\n" +
				"  - step 'build': post references undefined step 'ship'",
		},
		{
			name: "self reference",
			steps: `
  build: {steps: [build]}
`,
			wantErr: "invalid config:\n  - cycle detected: build -> build",
		},
		{
			name: "indirect cycle through steps",
			steps: `
  a: {steps: [b]}
  b: {steps: [c]}
  c: {steps: [a]}
`,
			wantErr: "invalid config:\n  - cycle detected: a -> b -> c -> a",
		},
		{
			name: "cycle through pre and post hooks",
			steps: `
  build: {pre: [gen], cmd: go build}
  gen: {cmd: go generate, post: [build]}
`,
			wantErr: "invalid config:\n  - cycle detected: build -> gen -> build",
		},
		{
			name: "reports every problem at once",
			steps: `
  a: {steps: [a]}
  b: {cmd: x, cmds: [y]}
  c: {pre: [missing], cmd: z}
`,
			wantErr: "invalid config:\n" +
				"  - step 'b': only one of cmd, cmds, or steps may be defined, found cmd, cmds\n" +
				"  - step 'c': pre references undefined step 'missing'\n" +
				"  - cycle detected: a -> a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conf, err := parse([]byte("project:\n  name: demo\nsteps:" + tt.steps))

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, conf)
				return
			}

			assert.NoError(t, err)
		})
	}
}
