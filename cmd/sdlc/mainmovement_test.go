package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIssueRecordPaths(t *testing.T) {
	for _, tc := range []struct {
		path string
		want bool
	}{
		{"workshop/issues/000001-a.md", true}, {"workshop/plans/000001-a-plan.md", true},
		{"workshop/history/issues/000001-a.md", true}, {"workshop/history/000001-a.md", true},
		{"workshop/issues/000001-a.side/note with\nnewline.md", true},
		{"ops/team.md", false}, {"workshop/projects/a.md", false}, {"workshop/lessons.md", false},
		{"workshop/issues-old/a.md", false}, {"cmd/sdlc/helptext/merge.md", false}, {"README.md", false},
		{".github/workflows/a.yml", false}, {"workshop/issues/main.go", false}, {"workshop/issues/run.sh", false},
		{"workshop/issues/a.json", false}, {"workshop/issues/../a.md", false}, {"/workshop/issues/a.md", false},
		{"workshop/issues//a.md", false}, {"workshop/issues/a.md/", false}, {"", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			if got := issueRecordOnlyPaths([]string{tc.path}); got != tc.want {
				t.Fatalf("%q: got %v want %v", tc.path, got, tc.want)
			}
		})
	}
	if issueRecordOnlyPaths(nil) || issueRecordOnlyPaths([]string{"workshop/issues/a.md", "main.go"}) {
		t.Fatal("empty or mixed range admitted")
	}
}

func TestMainMovementEveryCommitAndRename(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "-q")
	git(t, root, "config", "user.name", "fixture")
	git(t, root, "config", "user.email", "fixture@example.com")
	commit := func(path, body string) string {
		t.Helper()
		procedureWrite(t, root, path, body)
		git(t, root, "add", ".")
		git(t, root, "commit", "-qm", "fixture")
		return procedureHead(t, root)
	}
	base := commit("main.go", "package fixture\n")
	first := commit("workshop/issues/000001-a.md", "one\n")
	for _, tc := range []struct {
		base, main string
		want       bool
	}{{base, first, true}, {first, first, false}} {
		got, err := mainMovementIssueOnly(landingTransportRunner{}, root, tc.base, tc.main)
		if err != nil || got != tc.want {
			t.Fatalf("issue-only=%v err=%v want %v", got, err, tc.want)
		}
	}
	code := commit("main.go", "package changed\n")
	reverted := commit("main.go", "package fixture\n")
	commit("workshop/issues/000002-b.md", "two\n")
	if got, err := mainMovementIssueOnly(landingTransportRunner{}, root, first, procedureHead(t, root)); err != nil || got {
		t.Fatalf("code/revert was hidden: %v %v", got, err)
	}
	// A rename into the allowed tree must still reveal its non-record source.
	if err := os.Rename(filepath.Join(root, "main.go"), filepath.Join(root, "workshop/issues/copied.md")); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-qm", "rename into records")
	if got, err := mainMovementIssueOnly(landingTransportRunner{}, root, reverted, procedureHead(t, root)); err != nil || got {
		t.Fatalf("cross-root rename admitted: %v %v (prior %s)", got, err, code)
	}
}

type movementCountingGit struct {
	gitRunner
	calls int
	out   string
	err   error
}

func (g *movementCountingGit) GitInDir(_ string, args ...string) ([]byte, error) {
	g.calls++
	if !strings.Contains(strings.Join(args, " "), "--no-renames") || !strings.Contains(strings.Join(args, " "), "-m") {
		return nil, errors.New("history must include rename sources and merge parents")
	}
	return []byte(g.out), g.err
}
func TestMainMovementCountAndUnavailable(t *testing.T) {
	oid := strings.Repeat("a", 40)
	main := strings.Repeat("b", 40)
	for _, n := range []int{8, 16} {
		g := &movementCountingGit{out: strings.Repeat("workshop/issues/a.md\x00", n)}
		if ok, err := mainMovementIssueOnly(g, "fixture", oid, main); err != nil || !ok || g.calls != 1 {
			t.Fatalf("n=%d ok=%v calls=%d err=%v", n, ok, g.calls, err)
		}
	}
	for _, tc := range []struct {
		out string
		err error
	}{{"workshop/issues/a.md", nil}, {"", errors.New("unreadable")}} {
		g := &movementCountingGit{out: tc.out, err: tc.err}
		if ok, err := mainMovementIssueOnly(g, "fixture", oid, main); ok || err == nil {
			t.Fatalf("unreadable/malformed admitted: %v %v", ok, err)
		}
	}
}

func TestIssueRecordReadinessBinding(t *testing.T) {
	pr := landingPR{Repo: "a/b", Number: 1, State: "OPEN", HeadRef: "issue", HeadOID: strings.Repeat("a", 40), BaseRef: "main"}
	main := strings.Repeat("b", 40)
	r := landingReadiness{Repo: pr.Repo, Number: pr.Number, State: "OPEN", HeadRef: pr.HeadRef, HeadOID: pr.HeadOID, BaseRef: pr.BaseRef, BaseTargetOID: main, MergeStateStatus: "CLEAN"}
	if ok, err := readinessAllowsUnintegrated(pr, r, main); err != nil || !ok {
		t.Fatalf("matching CLEAN: %v %v", ok, err)
	}
	for _, status := range []string{"BEHIND", "UNKNOWN", "", "FUTURE"} {
		r.MergeStateStatus = status
		if ok, err := readinessAllowsUnintegrated(pr, r, main); err != nil || ok {
			t.Fatalf("%s: %v %v", status, ok, err)
		}
	}
	r.MergeStateStatus = "CLEAN"
	if ok, err := readinessAllowsUnintegrated(pr, r, strings.Repeat("c", 40)); err != nil || ok {
		t.Fatalf("stale base: %v %v", ok, err)
	}
	r.HeadOID = strings.Repeat("d", 40)
	if ok, err := readinessAllowsUnintegrated(pr, r, main); err == nil || ok {
		t.Fatalf("changed head authorized: %v %v", ok, err)
	}
}
