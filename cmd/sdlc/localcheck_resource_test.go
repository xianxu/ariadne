package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLocalCheckRunnerResource(t *testing.T) {
	d, sha, gh, output := localRunRepo(t, "must never acquire")
	// Missing helper is deliberate: runner admission must neither discover nor
	// invoke it. Inherited suite authority must not reach the command.
	if err := os.Remove(filepath.Join(d.Root, "scripts/test-lock.py")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_TEST_LOCK_HELD_BY", "foreign-slot")
	c := localCheck{Name: "vm", RunScope: "landing", Resource: "runner", Command: []string{"sh", "-c", `test -z "$WF_TEST_LOCK_HELD_BY" || exit 99; echo runner-output; exit 37`}, TimeoutMinutes: 1}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || rec.Exit != 37 {
		t.Fatalf("%+v %v\n%s", rec, err, output)
	}
	if gh.statuses[sha][localStatusContext("vm")] != "failure" {
		t.Fatal("failure not recorded")
	}
	raw, err := os.ReadFile(rec.LogPath)
	if err != nil || !strings.Contains(string(raw), "runner-output") {
		t.Fatalf("retained output %q %v", raw, err)
	}
}
