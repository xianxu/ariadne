//go:build unix

package testfix

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/testgate"
)

// A hung build is bounded even when a descendant holds the output pipe: the
// whole group dies at the limit and the wait returns (#438 review).
func TestRunBoundedKillsTheGroupAtItsLimit(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	// The child outlives every ceiling, so only the bound can end the wait.
	cmd := exec.Command("sh", "-c", `sleep 600 & echo $! > "$0"; wait`, pidfile)
	const limit = 300 * time.Millisecond
	var err error
	testgate.Returns(t, "the bounded build", func() { _, err = runBounded(cmd, limit) })
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("hung command: %v", err)
	}
	child := AwaitPIDFile(pidfile, HangCeiling, nil)
	if child == 0 {
		t.Fatal("the fixture never reported its child")
	}
	if err := AwaitExit(child, HangCeiling); err != nil {
		t.Errorf("a descendant survived the bound: %v", err)
	}
}

func TestRunBoundedReportsAFinishedCommandAsItIs(t *testing.T) {
	out, err := runBounded(exec.Command("sh", "-c", "echo built"), time.Minute)
	if err != nil || strings.TrimSpace(string(out)) != "built" {
		t.Fatalf("finished command: %q %v", out, err)
	}
}
