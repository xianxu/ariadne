// Package scripttest runs the repository's shell tests from `make test` (#362):
// the Go suite is what `make test` runs, so a shell test needs a Go entry point
// to count.
package scripttest

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot (#324)
	os.Exit(m.Run())
}

// runnerDeadline bounds the whole runner (it takes about 13s) with the shared
// test hang ceiling, so a hanging shell test fails here instead of holding the
// suite until go test's timeout.
const runnerDeadline = testgate.HangCeiling

// TestConstructScripts runs every construct/scripts/test/*.test.sh through the
// one runner the CI merge check also runs, so both apply the same exclusions.
func TestConstructScripts(t *testing.T) {
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
			runConstructScripts(t)
		})
	}
}

func runConstructScripts(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), runnerDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "construct-scripts.test.sh")
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
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if ctx.Err() != nil {
		t.Fatalf("construct/scripts/test did not finish within %v", runnerDeadline)
	}
	if err != nil {
		t.Fatalf("construct/scripts/test: %v (failing tests are named above)", err)
	}
}
