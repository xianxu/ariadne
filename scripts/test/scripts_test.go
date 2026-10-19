// Package scripttest runs the repository's shell tests from `make test` (#362):
// the Go suite is what `make test` runs, so a shell test needs a Go entry point
// to count.
package scripttest

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot (#324)
	os.Exit(m.Run())
}

// runnerDeadline bounds each command, not cumulative progress across a batch
// of independent scripts (#448). Keep the shared hang ceiling unchanged.
const runnerDeadline = testgate.HangCeiling

// TestConstructScripts gets selection from the CI runner, so both apply the
// same exclusions, then gives each script its own hang budget.
func TestConstructScripts(t *testing.T) {
	testConstructScripts(t, executeConstructScript)
}

// The command seam lets regressions observe context lifetimes without waiting
// for the hang ceiling or depending on machine load.
type constructCommand func(context.Context, string, ...string) ([]byte, error)

func testConstructScripts(t *testing.T, execute constructCommand) {
	t.Helper()
	for _, env := range []struct {
		name, noColor, forceColor, term, ci string
	}{
		{"plain", "", "", "dumb", ""},
		{"no_color_ci", "1", "1", "dumb", "1"},
		{"forced_color_terminal", "", "1", "xterm-256color", ""},
	} {
		t.Run(env.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", env.noColor)
			t.Setenv("CLICOLOR_FORCE", env.forceColor)
			t.Setenv("TERM", env.term)
			t.Setenv("CI", env.ci)
			runConstructScripts(t, execute)
		})
	}
}

func runConstructScripts(t *testing.T, execute constructCommand) {
	t.Helper()
	out := runConstructScript(t, execute, "construct-scripts.test.sh", "--list")
	if len(out) > 0 && out[len(out)-1] != 0 {
		t.Fatalf("construct script selection is not NUL terminated: %q", out)
	}
	for _, path := range bytes.Split(bytes.TrimSuffix(out, []byte{0}), []byte{0}) {
		if len(path) == 0 {
			continue // all scripts may be explicitly excluded
		}
		script := string(path)
		t.Run(filepath.Base(script), func(t *testing.T) {
			runConstructScript(t, execute, script)
		})
	}
}

func runConstructScript(t *testing.T, execute constructCommand, script string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), runnerDeadline)
	defer cancel()
	out, err := execute(ctx, script, args...)
	t.Logf("%s", out)
	if ctx.Err() != nil {
		t.Fatalf("%s did not finish within %v", script, runnerDeadline)
	}
	if err != nil {
		t.Fatalf("%s: %v", script, err)
	}
	return out
}

func executeConstructScript(ctx context.Context, script string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "bash", append([]string{script}, args...)...)
	// Its own process group, killed whole on the deadline: a test's background
	// children must not outlive it holding the output pipe.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			return err
		}
		return nil
	}
	cmd.WaitDelay = 5 * time.Second
	return cmd.CombinedOutput()
}
