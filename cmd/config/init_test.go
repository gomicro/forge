package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gomicro/forge/confile"
	"github.com/stretchr/testify/assert"
)

func TestInitializeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr string
	}{
		{name: "writes canonical config without git"},
		{name: "does not overwrite existing config", wantErr: "config file already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(t.TempDir(), "forge-init-sample")
			err := os.Mkdir(dir, 0755)
			if !assert.NoError(t, err) {
				return
			}

			configPath := filepath.Join(dir, confile.DefaultPath)
			if tt.wantErr != "" {
				err = os.WriteFile(configPath, []byte("existing config\n"), 0644)
				if !assert.NoError(t, err) {
					return
				}
			}

			err = initializeConfig(configPath, dir)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			content, err := os.ReadFile(configPath)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, `project:
  name: forge-init-sample

steps:
  build:
    help: build the project
    cmd: echo "run the build"
`, string(content))

			parsed, err := confile.Parse(configPath)
			if assert.NoError(t, err) {
				assert.Equal(t, "forge-init-sample", parsed.Project.Name)
			}
		})
	}
}
