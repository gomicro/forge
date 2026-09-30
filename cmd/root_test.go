package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportError(t *testing.T) {
	t.Parallel()

	exitErr := exec.Command("bash", "-c", "exit 3").Run()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantOut  string
	}{
		{
			name:     "plain error",
			err:      errors.New("target not found: nope"),
			wantCode: 1,
			wantOut:  "Error: target not found: nope\n",
		},
		{
			name:     "wrapped exit error",
			err:      fmt.Errorf("executing step build: %w", exitErr),
			wantCode: 3,
			wantOut:  "Error: executing step build: exit status 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			code := reportError(tt.err, &out)

			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, tt.wantOut, out.String())
		})
	}
}
