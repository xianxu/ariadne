package golden

import (
	"strings"
	"testing"
)

func TestRenderLedger(t *testing.T) {
	divs := []Divergence{
		{Class: Match, Verb: "symlink", Path: "Makefile", Detail: "link -> ../ariadne/Makefile"},
		{Class: Match, Verb: "mkdir", Path: ".claude/skills", Detail: "dir exists"},
		{Class: Expected, Verb: "seed", Path: "bootstrap.sh", Detail: "weave defers seed"},
		{Class: Unexpected, Verb: "symlink", Path: "X", Detail: "live link -> Y"},
	}
	out := Render("/ws/nous", divs)

	// Header names the repo.
	if !strings.Contains(out, "/ws/nous") {
		t.Fatalf("ledger missing repo path:\n%s", out)
	}
	// A summary line with the per-class counts.
	for _, want := range []string{"MATCH 2", "EXPECTED 1", "UNEXPECTED 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ledger missing summary %q:\n%s", want, out)
		}
	}
	// Each UNEXPECTED line is shown (it's the failure evidence): class, verb,
	// path, and detail all present on one line.
	if !lineWith(out, "UNEXPECTED", "symlink", "X", "live link -> Y") {
		t.Fatalf("ledger missing UNEXPECTED line:\n%s", out)
	}
}

// lineWith reports whether some line of out contains all the given substrings.
func lineWith(out string, parts ...string) bool {
	for _, line := range strings.Split(out, "\n") {
		all := true
		for _, p := range parts {
			if !strings.Contains(line, p) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

func TestRenderCleanSummary(t *testing.T) {
	divs := []Divergence{{Class: Match, Verb: "symlink", Path: "a"}, {Class: Expected, Verb: "seed", Path: "b"}}
	out := Render("/ws/x", divs)
	if !strings.Contains(out, "UNEXPECTED 0") {
		t.Fatalf("clean ledger should report UNEXPECTED 0:\n%s", out)
	}
}
