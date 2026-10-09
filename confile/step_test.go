package confile

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/gomicro/forge/vars"

	"github.com/gomicro/scribe"
	"github.com/stretchr/testify/assert"
)

func testScriber(t *testing.T, output *bytes.Buffer) scribe.Scriber {
	t.Helper()

	scrb, err := scribe.NewScribe(output, &scribe.Theme{
		Describe: func(s string) string { return s },
		Print:    func(s string) string { return s },
		Error:    func(err error) string { return err.Error() },
	})
	if err != nil {
		t.Fatalf("creating test output: %v", err)
	}

	return scrb
}

func TestStepExecuteChildContext(t *testing.T) {
	t.Parallel()

	steps := map[string]*Step{
		"root":  {Pre: []string{"pre"}, Steps: []string{"group"}, Post: []string{"post"}},
		"pre":   {Cmd: `printf 'pre:%s:%s\n' '{{.Project}}' "$FORGE_TEST_PROJECT_VALUE"`},
		"group": {Steps: []string{"child"}},
		"child": {Cmds: []string{
			`printf 'child:%s:%s\n' '{{.Project}}' "$FORGE_TEST_PROJECT_VALUE"`,
			`printf 'again:%s\n' '{{.Project}}'`,
		}},
		"post": {Cmd: `printf 'post:%s:%s\n' '{{.Project}}' "$FORGE_TEST_PROJECT_VALUE"`},
	}
	v := vars.Vars{"Project": "demo"}
	envs := map[string]string{"FORGE_TEST_PROJECT_VALUE": "{{.Project}}-env"}
	var output bytes.Buffer
	scrb := testScriber(t, &output)

	var err error
	assert.NotPanics(t, func() {
		err = steps["root"].Execute(context.Background(), "root", RunOptions{
			Steps: steps, ProjectEnvs: envs, Vars: &v, Scriber: scrb,
		})
	})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "pre:demo:demo-env")
	assert.Contains(t, output.String(), "child:demo:demo-env")
	assert.Contains(t, output.String(), "again:demo")
	assert.Contains(t, output.String(), "post:demo:demo-env")
}

func TestStepExecuteOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skipPre  bool
		skipPost bool
		want     []string
	}{
		{name: "all hooks", want: []string{"PRE", "NESTED_PRE", "BODY", "NESTED_POST", "POST"}},
		{name: "skip pre at every level", skipPre: true, want: []string{"BODY", "NESTED_POST", "POST"}},
		{name: "skip post at every level", skipPost: true, want: []string{"PRE", "NESTED_PRE", "BODY"}},
		{name: "skip both at every level", skipPre: true, skipPost: true, want: []string{"BODY"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			steps := map[string]*Step{
				"root":        {Pre: []string{"pre"}, Steps: []string{"child"}, Post: []string{"post"}},
				"child":       {Pre: []string{"nested-pre"}, Cmd: "printf 'BODY\\n'", Post: []string{"nested-post"}},
				"pre":         {Cmd: "printf 'PRE\\n'"},
				"post":        {Cmd: "printf 'POST\\n'"},
				"nested-pre":  {Cmd: "printf 'NESTED_PRE\\n'"},
				"nested-post": {Cmd: "printf 'NESTED_POST\\n'"},
			}
			v := vars.Vars{}
			var output bytes.Buffer
			options := RunOptions{
				Steps: steps, Vars: &v, Scriber: testScriber(t, &output),
				SkipPre: tt.skipPre, SkipPost: tt.skipPost,
			}

			err := steps["root"].Execute(context.Background(), "root", options)

			assert.NoError(t, err)
			lines := regexp.MustCompile(`(?m)^\s*(PRE|NESTED_PRE|BODY|NESTED_POST|POST)$`)
			matches := lines.FindAllStringSubmatch(output.String(), -1)
			var got []string
			for _, match := range matches {
				got = append(got, match[1])
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStepExecuteFailure(t *testing.T) {
	t.Parallel()

	steps := map[string]*Step{
		"root":  {Steps: []string{"fail", "later"}, Post: []string{"later"}},
		"fail":  {Cmd: "printf 'failure-output\\n' >&2; exit 3"},
		"later": {Cmd: "printf 'should-not-run\\n'"},
	}

	v := vars.Vars{}
	var output bytes.Buffer
	options := RunOptions{
		Steps: steps, Vars: &v, Scriber: testScriber(t, &output), Verbose: true,
	}

	err := steps["root"].Execute(context.Background(), "root", options)

	var exitErr *exec.ExitError
	if assert.True(t, errors.As(err, &exitErr)) {
		assert.Equal(t, 3, exitErr.ExitCode())
	}

	assert.ErrorContains(t, err, "executing step fail")
	assert.Contains(t, output.String(), "failure-output")
	assert.NotContains(t, output.String(), "should-not-run")
}

func TestStepExecuteEnvPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectEnvs map[string]string
		stepEnvs    map[string]string
		wantHome    string
	}{
		{name: "inherits shell", wantHome: os.Getenv("HOME")},
		{
			name:        "project overrides shell",
			projectEnvs: map[string]string{"HOME": "project-{{.Project}}"},
			wantHome:    "project-demo",
		},
		{
			name:     "step overrides shell without project envs",
			stepEnvs: map[string]string{"HOME": "step-{{.Project}}"},
			wantHome: "step-demo",
		},
		{
			name:        "step overrides project and shell",
			projectEnvs: map[string]string{"HOME": "project"},
			stepEnvs:    map[string]string{"HOME": "step-{{.Project}}"},
			wantHome:    "step-demo",
		},
		{
			name:        "empty step value overrides project and shell",
			projectEnvs: map[string]string{"HOME": "project"},
			stepEnvs:    map[string]string{"HOME": ""},
			wantHome:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			step := &Step{Cmd: `printf 'HOME=[%s]\n' "$HOME"`, Envs: tt.stepEnvs}
			v := vars.Vars{"Project": "demo"}
			var output bytes.Buffer
			options := RunOptions{
				ProjectEnvs: tt.projectEnvs, Vars: &v, Scriber: testScriber(t, &output),
			}

			err := step.Execute(context.Background(), "env", options)

			assert.NoError(t, err)
			assert.Contains(t, output.String(), "HOME=["+tt.wantHome+"]")
		})
	}
}
