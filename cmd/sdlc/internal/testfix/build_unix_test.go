//go:build unix

package testfix

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A hung build is bounded even when a descendant holds the output pipe: the
// whole group dies at the limit and the wait returns (#438 review).
func TestRunBoundedKillsTheGroupAtItsLimit(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	cmd := exec.Command("sh", "-c", `sleep 60 & echo $! > "$0"; wait`, pidfile)
	start := time.Now()
	_, err := runBounded(cmd, 300*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("hung command: %v", err)
	}
	if took := time.Since(start); took > HangCeiling {
		t.Fatalf("the bound did not hold: %s", took)
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
