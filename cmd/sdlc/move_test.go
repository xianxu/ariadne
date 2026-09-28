package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// moveFixture: the procedure fixture with slot 1 on issue branch
// 000001-procedure, one commit ahead (a tracked file "feature").
func moveFixture(t *testing.T) (roots []string, selected string) {
	t.Helper()
	roots, _ = procedureFixture(t)
	testfix.Git(t, roots[1], "switch", "-q", "-c", "000001-procedure")
	procedureWrite(t, roots[1], "feature", "selected branch\n")
	testfix.Git(t, roots[1], "add", "feature")
	testfix.Git(t, roots[1], "commit", "-qm", "feature change")
	return roots, procedureHead(t, roots[1])
}

// moveSnapshot captures what a refused move must leave unchanged.
func moveSnapshot(t *testing.T, roots ...string) string {
	t.Helper()
	var b strings.Builder
	for _, r := range roots {
		b.WriteString(testfix.Capture(t, r, "status", "--porcelain=v1", "--untracked-files=all"))
		b.WriteString(testfix.Capture(t, r, "symbolic-ref", "HEAD"))
	}
	b.WriteString(testfix.Capture(t, roots[0], "show-ref"))
	return b.String()
}

func runMoveTest(t *testing.T, dir, address string, dryRun bool) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := runMove(dir, address, dryRun, &out, &out)
	return out.String(), err
}

func assertMoved(t *testing.T, from, to, fromRest, fromRestHead, selected string) {
	t.Helper()
	if got := strings.TrimSpace(testfix.Capture(t, to, "branch", "--show-current")); got != "000001-procedure" {
		t.Fatalf("destination branch = %q", got)
	}
	if got := procedureHead(t, to); got != selected {
		t.Fatalf("destination HEAD = %s, want %s", got, selected)
	}
	if got := strings.TrimSpace(testfix.Capture(t, from, "branch", "--show-current")); got != fromRest {
		t.Fatalf("source branch = %q, want %s", got, fromRest)
	}
	if got := procedureHead(t, from); got != fromRestHead {
		t.Fatalf("source resting HEAD moved: %s", got)
	}
}

func TestMoveToPrimary(t *testing.T) {
	roots, selected := moveFixture(t)
	rest := strings.TrimSpace(testfix.Capture(t, roots[1], "rev-parse", "main-slot1"))
	procedureWrite(t, roots[0], "operator-scratch", "keep me\n")
	out, err := runMoveTest(t, roots[1], "", false)
	if err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	assertMoved(t, roots[1], roots[0], "main-slot1", rest, selected)
	if data, err := os.ReadFile(filepath.Join(roots[0], "operator-scratch")); err != nil || string(data) != "keep me\n" {
		t.Fatalf("scratch changed: %q %v", data, err)
	}
}

func TestMoveToSlot(t *testing.T) {
	roots, selected := moveFixture(t)
	rest := strings.TrimSpace(testfix.Capture(t, roots[1], "rev-parse", "main-slot1"))
	if out, err := runMoveTest(t, roots[1], ":2", false); err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	assertMoved(t, roots[1], roots[2], "main-slot1", rest, selected)
}

func TestMoveReportsParked(t *testing.T) {
	roots, selected := moveFixture(t)
	rest := strings.TrimSpace(testfix.Capture(t, roots[1], "rev-parse", "main-slot1"))
	testfix.Git(t, roots[0], "commit", "--allow-empty", "-qm", "local main work")
	parked := procedureHead(t, roots[0])
	out, err := runMoveTest(t, roots[1], ":0", false)
	if err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	assertMoved(t, roots[1], roots[0], "main-slot1", rest, selected)
	if !strings.Contains(out, "local main work") {
		t.Fatalf("parked commit not reported:\n%s", out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, roots[0], "rev-parse", "main")); got != parked {
		t.Fatalf("parked main moved: %s", got)
	}
}

func TestMoveDryRun(t *testing.T) {
	roots, _ := moveFixture(t)
	before := moveSnapshot(t, roots[0], roots[1])
	out, err := runMoveTest(t, roots[1], ":0", true)
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	if !strings.Contains(out, "000001-procedure") {
		t.Fatalf("dry run did not describe the move:\n%s", out)
	}
	if moveSnapshot(t, roots[0], roots[1]) != before {
		t.Fatal("dry run changed a slot")
	}
}

func TestMoveRefusals(t *testing.T) {
	for _, tc := range []struct {
		name, address, want string
		setup               func(t *testing.T, roots []string)
	}{
		{"missing slot", ":5", ":5", nil},
		{"destination tracked change", ":0", ":0 has uncommitted changes", func(t *testing.T, r []string) {
			procedureWrite(t, r[0], "README", "pending\n")
		}},
		{"destination not resting", ":0", "not its resting branch", func(t *testing.T, r []string) {
			testfix.Git(t, r[0], "switch", "-q", "-c", "other")
		}},
		{"destination collision", ":0", "would overwrite: feature", func(t *testing.T, r []string) {
			procedureWrite(t, r[0], "feature", "local\n")
		}},
		{"source tracked change", ":0", ":1 has uncommitted changes", func(t *testing.T, r []string) {
			procedureWrite(t, r[1], "README", "pending\n")
		}},
		{"source untracked", ":0", ":1 has untracked files", func(t *testing.T, r []string) {
			procedureWrite(t, r[1], "tmp", "x\n")
		}},
		{"source resting", ":0", "no issue branch to move", func(t *testing.T, r []string) {
			testfix.Git(t, r[1], "switch", "-q", "main-slot1")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			roots, _ := moveFixture(t)
			if tc.setup != nil {
				tc.setup(t, roots)
			}
			before := moveSnapshot(t, roots[0], roots[1], roots[2])
			out, err := runMoveTest(t, roots[1], tc.address, false)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q\n%s", err, tc.want, out)
			}
			if moveSnapshot(t, roots[0], roots[1], roots[2]) != before {
				t.Fatal("refused move changed a slot")
			}
		})
	}
}

// Git's --no-overwrite-ignore is the final guard: an ignored file the
// preflight does not see makes the second switch fail after the first ran.
func TestMoveSecondSwitchFails(t *testing.T) {
	roots, _ := moveFixture(t)
	procedureWrite(t, roots[1], "build/output", "incoming tracked output\n")
	testfix.Git(t, roots[1], "add", "-f", "build/output")
	testfix.Git(t, roots[1], "commit", "-qm", "track build output")
	selected := procedureHead(t, roots[1])
	procedureWrite(t, roots[0], "build/output", "local build\n")
	out, err := runMoveTest(t, roots[1], ":0", false)
	if err == nil || !strings.Contains(err.Error(), "switch 000001-procedure") {
		t.Fatalf("err = %v, want the retry command\n%s", err, out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, roots[0], "rev-parse", "000001-procedure")); got != selected {
		t.Fatalf("branch ref moved: %s", got)
	}
	for root, want := range map[string]string{roots[0]: "main", roots[1]: "main-slot1"} {
		if got := strings.TrimSpace(testfix.Capture(t, root, "branch", "--show-current")); got != want {
			t.Fatalf("%s on %q, want %s", root, got, want)
		}
	}
	if data, err := os.ReadFile(filepath.Join(roots[0], "build/output")); err != nil || string(data) != "local build\n" {
		t.Fatalf("ignored output overwritten: %q %v", data, err)
	}
}
