package confile

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sha    string
		branch string
		want   map[string]string
	}{
		{
			name:   "trims trailing newlines from git output",
			sha:    "0123456789abcdef0123456789abcdef01234567\n",
			branch: "main\n",
			want: map[string]string{
				"Branch":   "main",
				"Dir":      "/work/demo",
				"Os":       runtime.GOOS,
				"Project":  "demo",
				"Sha":      "0123456789abcdef0123456789abcdef01234567",
				"ShortSha": "0123456",
			},
		},
		{
			name:   "short sha shorter than seven characters",
			sha:    "abc\n",
			branch: "",
			want: map[string]string{
				"Branch":   "",
				"Dir":      "/work/demo",
				"Os":       runtime.GOOS,
				"Project":  "demo",
				"Sha":      "abc",
				"ShortSha": "abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := builtinVars("demo", tt.sha, tt.branch, "/work/demo")

			assert.Equal(t, tt.want, map[string]string(*v))
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:  "valid config",
			input: "project:\n  name: demo\nsteps:\n  build:\n    cmd: echo hi\n",
		},
		{
			name:    "missing project block",
			input:   "steps:\n  build:\n    cmd: echo hi\n",
			wantErr: "missing required project block",
		},
		{
			name:    "step references itself",
			input:   "project:\n  name: demo\nsteps:\n  build:\n    steps: [build]\n",
			wantErr: "infinite loop detected: step 'build'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conf, err := parse([]byte(tt.input))

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, conf)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, "demo", conf.Project.Name)
		})
	}
}
