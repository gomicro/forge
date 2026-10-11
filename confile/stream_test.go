package confile

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gomicro/forge/vars"
	"github.com/gomicro/scribe"
	"github.com/stretchr/testify/assert"
)

type notifyingWriter struct {
	mu       sync.Mutex
	output   bytes.Buffer
	notified chan struct{}
	notifyOn string
}

func (w *notifyingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.output.Write(p)
	if bytes.Contains(p, []byte(w.notifyOn)) {
		select {
		case w.notified <- struct{}{}:
		default:
		}
	}

	return n, err
}

func (w *notifyingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.output.String()
}

func testScriberForStream(t *testing.T, output *bytes.Buffer) scribe.Scriber {
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

func TestExecuteCmdStreamsVerboseOutput(t *testing.T) {
	t.Parallel()

	writer := &notifyingWriter{notified: make(chan struct{}, 1), notifyOn: "ready"}
	scrb, err := scribe.NewScribe(writer, &scribe.Theme{
		Describe: func(s string) string { return s },
		Print:    func(s string) string { return s },
		Error:    func(err error) string { return err.Error() },
	})
	if !assert.NoError(t, err) {
		return
	}

	v := vars.Vars{}
	done := make(chan error, 1)
	go func() {
		done <- executeCmd(context.Background(), "printf 'ready\\n'; sleep 2; printf tail", nil, RunOptions{
			Vars: &v, Scriber: scrb, Verbose: true, Stdin: strings.NewReader(""),
		})
	}()

	select {
	case <-writer.notified:
		assert.Contains(t, writer.String(), "ready")

	case err := <-done:
		t.Fatalf("command completed before streaming check: %v", err)

	case <-time.After(500 * time.Millisecond):
		t.Fatal("verbose stdout was not streamed while the command was running")
	}

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.Contains(t, writer.String(), "tail")

	case <-time.After(3 * time.Second):
		t.Fatal("command did not finish")
	}
}

func TestExecuteCmdQuietOutputAndStdin(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	v := vars.Vars{}
	err := executeCmd(context.Background(), "printf 'hidden\\n'; printf 'warning\\n' >&2; read input; test \"$input\" = 'interactive input'", nil, RunOptions{
		Vars: &v, Scriber: testScriberForStream(t, &bytes.Buffer{}), Stderr: &stderr,
		Stdin: strings.NewReader("interactive input\n"),
	})

	assert.NoError(t, err)
	assert.Contains(t, stderr.String(), "warning")
	assert.Contains(t, stderr.String(), "stderr")
	assert.NotContains(t, stderr.String(), "hidden")
}

func TestExecuteCmdStreamsQuietStderr(t *testing.T) {
	t.Parallel()

	stderr := &notifyingWriter{notified: make(chan struct{}, 1), notifyOn: "warning"}
	v := vars.Vars{}
	scrb := testScriberForStream(t, &bytes.Buffer{})
	done := make(chan error, 1)
	go func() {
		done <- executeCmd(context.Background(), "printf 'warning\\n' >&2; sleep 2", nil, RunOptions{
			Vars: &v, Scriber: scrb, Stderr: stderr,
		})
	}()

	select {
	case <-stderr.notified:
		assert.Contains(t, stderr.String(), "warning")

	case err := <-done:
		t.Fatalf("command completed before streaming check: %v", err)

	case <-time.After(500 * time.Millisecond):
		t.Fatal("quiet stderr was not streamed while the command was running")
	}

	select {
	case err := <-done:
		assert.NoError(t, err)

	case <-time.After(3 * time.Second):
		t.Fatal("command did not finish")
	}
}
