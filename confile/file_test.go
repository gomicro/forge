package confile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()

	path := filepath.Join(dir, DefaultPath)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("writing config: %v", err)
	}

	return path
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	base := []string{"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}
	cmd := exec.Command("git", append(base, args...)...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}

	return string(out)
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	t.Run("parses without a git repository", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, t.TempDir(), "project:\n  name: demo\nsteps:\n  build:\n    cmd: echo hi\n")

		conf, err := Parse(path)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "demo", conf.Project.Name)
		assert.Equal(t, "echo hi", conf.Steps["build"].Cmd)
		assert.Nil(t, conf.Vars)
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		conf, err := Parse(filepath.Join(t.TempDir(), DefaultPath))

		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, conf)
	})
}

func TestResolveVars(t *testing.T) {
	t.Parallel()

	t.Run("populates vars from the repository", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		runGit(t, dir, "init", "-q", "-b", "trunk")
		runGit(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
		sha := runGit(t, dir, "rev-parse", "HEAD")

		conf := &File{Project: &Project{Name: "demo"}}
		err := conf.ResolveVars(dir)

		assert.NoError(t, err)
		assert.Equal(t, builtinVars("demo", sha, "trunk", dir), conf.Vars)
	})

	t.Run("errors outside a repository", func(t *testing.T) {
		t.Parallel()

		conf := &File{Project: &Project{Name: "demo"}}
		err := conf.ResolveVars(t.TempDir())

		assert.ErrorContains(t, err, "resolveVars: getting sha")
		assert.ErrorContains(t, err, "not a git repository")

		var exitErr *exec.ExitError
		assert.False(t, errors.As(err, &exitErr), "git exit status must not propagate as a step failure")
		assert.Nil(t, conf.Vars)
	})
}

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
			wantErr: "invalid config:\n  - cycle detected: build -> build",
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
