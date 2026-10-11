package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// disputeReady is closeReady plus a committed boundary ledger raising BR-1 (Important)
// and BR-2 (Critical) in round 1. Returns the repo, the issue number and the ledger path.
func disputeReady(t *testing.T, id int) (*trackerRepo, string, string) {
	t.Helper()
	t.Setenv("WF_PLANS_DIR", "")
	r, cardPath, detailPath := closeReady(t, id)
	name := filepath.Base(detailPath)
	l := gatestate.Ledger{Gate: boundaryGateKind.Gate, IssueNum: id, IDPrefix: boundaryGateKind.IDPrefix}
	l = gatestate.Apply(l, gatestate.Round{N: 1, Timestamp: "2026-10-10T00:00:00Z", Agent: "claude", Blocked: true,
		New: []gatestate.Finding{
			{ID: "BR-1", Severity: "Important", Title: "byte path unasserted", Round: 1},
			{ID: "BR-2", Severity: "Critical", Title: "guard reads the index", Round: 1},
		}})
	plansDir := resolvePlansDir("")
	if err := writeBoundaryGateLedger(plansDir, name, l, "ariadne"); err != nil {
		t.Fatal(err)
	}
	path := boundaryGatePath(plansDir, name)
	r.git("add", path)
	r.git("commit", "-qm", fmt.Sprintf("#%d: boundary review round 1", id))
	return r, path, cardPath
}

func disputeArgs(id int, finding string, extra ...string) []string {
	return append([]string{"finding", "dispute", "--issue", fmt.Sprint(id), "--id", finding,
		"--rationale", "the full suite asserts it", "--evidence", "go test ./...: exit = 9, want 0"}, extra...)
}

func readDisputeLedger(t *testing.T, path string, id int) gatestate.Ledger {
	t.Helper()
	l, err := readGateLedger(boundaryGateKind, filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), "-close-gate.md")+".md", id)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestFindingDisputeRecordsAndCommits(t *testing.T) {
	r, path, _ := disputeReady(t, 371)
	before := r.git("rev-parse", "HEAD")
	_, stderr, err := executeSDLCTestCommand(disputeArgs(371, "BR-1")...)
	if err != nil {
		t.Fatalf("dispute: %v\n%s", err, stderr)
	}
	if !strings.Contains(stderr, "BR-1 [Important] byte path unasserted — DISPUTED, not fixed") {
		t.Errorf("the dispute must be loud; stderr:\n%s", stderr)
	}
	l := readDisputeLedger(t, path, 371)
	if len(l.Disputes) != 1 || l.Disputes[0].After != 1 || l.Disputes[0].Evidence != "go test ./...: exit = 9, want 0" || l.Disputes[0].By == "" {
		t.Fatalf("ledger disputes = %+v", l.Disputes)
	}
	if got := strings.Join(findingIDs(gatestate.OpenFindings(l)), ","); got != "BR-2" {
		t.Fatalf("open = %q, want BR-2", got)
	}
	if subj := r.git("log", "-1", "--format=%s"); subj != "#371: dispute BR-1" {
		t.Errorf("commit subject = %q", subj)
	}
	if files := r.git("diff", "--name-only", before, "HEAD"); files != path {
		t.Errorf("the commit must touch only the ledger; touched %q", files)
	}
	if st := r.git("status", "--porcelain", "--", path); st != "" {
		t.Errorf("ledger left dirty: %q", st)
	}

	// An appeal-only finding: recorded, still blocking, and said so.
	_, stderr, err = executeSDLCTestCommand(disputeArgs(371, "BR-2")...)
	if err != nil {
		t.Fatalf("dispute BR-2: %v\n%s", err, stderr)
	}
	if !strings.Contains(stderr, "keeps blocking until the next review round withdraws it") {
		t.Errorf("an appealed Critical must say it still blocks; stderr:\n%s", stderr)
	}
	if got := strings.Join(findingIDs(gatestate.OpenFindings(readDisputeLedger(t, path, 371))), ","); got != "BR-2" {
		t.Fatalf("open = %q, want BR-2 still open", got)
	}
}

func TestFindingDisputeRerunRefuses(t *testing.T) {
	r, _, _ := disputeReady(t, 372)
	if _, stderr, err := executeSDLCTestCommand(disputeArgs(372, "BR-1")...); err != nil {
		t.Fatalf("dispute: %v\n%s", err, stderr)
	}
	head := r.git("rev-parse", "HEAD")
	_, _, err := executeSDLCTestCommand(disputeArgs(372, "BR-1")...)
	if err == nil || !strings.Contains(err.Error(), "BR-1 is already disputed (after round 1)") {
		t.Fatalf("rerun err = %v", err)
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Error("a refused rerun must not commit")
	}
}

// Each refusal names its own reason and leaves the ledger and HEAD untouched.
func TestFindingDisputeRefusals(t *testing.T) {
	r, path, cardPath := disputeReady(t, 373)
	head := r.git("rev-parse", "HEAD")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"blank evidence", []string{"finding", "dispute", "--issue", "373", "--id", "BR-1", "--rationale", "x", "--evidence", "  "}, "--rationale and --evidence are both required"},
		{"missing rationale", []string{"finding", "dispute", "--issue", "373", "--id", "BR-1", "--evidence", "x"}, "--rationale and --evidence are both required"},
		{"unknown prefix", disputeArgs(373, "XX-1"), `"XX-1" is not a gate finding id; known: BR-<n> (boundary-review), PQ-<n> (plan-quality)`},
		{"unissued id", disputeArgs(373, "BR-9"), "no finding BR-9 in the boundary-review ledger"},
		{"plan gate has no such finding", disputeArgs(373, "PQ-1"), "no finding PQ-1 in the plan-quality ledger"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := executeSDLCTestCommand(c.args...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}

	t.Run("another workspace's issue", func(t *testing.T) {
		owner, ok := ownerOf(t, r, cardPath)
		if !ok {
			t.Fatal("fixture: no owner recorded")
		}
		elsewhere := owner
		elsewhere.Worktree = "/elsewhere/ariadne"
		withClaimant(t, elsewhere)
		defer withClaimant(t, owner)
		_, _, err := executeSDLCTestCommand(disputeArgs(373, "BR-1")...)
		if err == nil || !strings.Contains(err.Error(), "owned by "+owner.Operator) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("off the issue branch", func(t *testing.T) {
		branch := r.git("rev-parse", "--abbrev-ref", "HEAD")
		r.git("checkout", "-q", "-b", "elsewhere")
		defer r.git("checkout", "-q", branch)
		_, _, err := executeSDLCTestCommand(disputeArgs(373, "BR-1")...)
		if err == nil || !strings.Contains(err.Error(), "gate ledgers live on its branch") {
			t.Fatalf("err = %v", err)
		}
	})

	if r.git("rev-parse", "HEAD") != head || len(readDisputeLedger(t, path, 373).Disputes) != 0 {
		t.Error("a refusal must write and commit nothing")
	}
}

func findingIDs(fs []gatestate.Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.ID)
	}
	return out
}
