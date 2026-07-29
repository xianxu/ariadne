package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// The plan-gate ledger lands beside the boundary-review sidecars, under a DISTINCT suffix
// so a verdict consumer globbing *-review.md (verdict.cue's discovery glob) never sees a
// gate ledger it cannot validate.
func TestPlanGatePath(t *testing.T) {
	got := planGatePath("workshop/plans", "000187-tune-change-code-gate.md")
	want := filepath.Join("workshop/plans", "000187-tune-change-code-gate-plan-gate.md")
	if got != want {
		t.Errorf("planGatePath = %q, want %q", got, want)
	}
	if strings.HasSuffix(got, "-review.md") {
		t.Error("plan-gate ledger must not match verdict.cue's *-review.md discovery glob")
	}
}

// sidecarPathFor's extraction must not have changed the boundary-review paths.
func TestSidecarPathUnchangedByExtraction(t *testing.T) {
	cases := []struct{ milestone, want string }{
		{"", "000187-x-close-review.md"},
		{"M1", "000187-x-m1-review.md"},
		{"M2", "000187-x-m2-review.md"},
	}
	for _, c := range cases {
		got := sidecarPath("plans", "000187-x.md", c.milestone)
		if got != filepath.Join("plans", c.want) {
			t.Errorf("sidecarPath(%q) = %q, want plans/%s", c.milestone, got, c.want)
		}
	}
}

// A missing ledger is the normal round-1 state: an empty ledger, not an error.
func TestReadPlanGateLedgerAbsent(t *testing.T) {
	l, err := readPlanGateLedger(t.TempDir(), "000187-x.md", 187)
	if err != nil {
		t.Fatalf("absent ledger should not error: %v", err)
	}
	if len(l.Rounds) != 0 || l.Gate != "plan-quality" || l.IDPrefix != "PQ" || l.IssueNum != 187 {
		t.Errorf("empty ledger = %+v", l)
	}
}

func TestWriteThenReadPlanGateLedger(t *testing.T) {
	dir := t.TempDir()
	l, err := readPlanGateLedger(dir, "000187-x.md", 187)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	l = gatestate.Apply(l, gatestate.AssignIDs(l, gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Critical", Title: "seam in wrong layer"}},
	}, 1, "2026-07-29T10:00:00Z", "claude"))
	l.ContentHash = gatestate.ContentHash("issue", "plan")

	if err := writePlanGateLedger(dir, "000187-x.md", l, "ariadne"); err != nil {
		t.Fatalf("write: %v", err)
	}
	back, err := readPlanGateLedger(dir, "000187-x.md", 187)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(back.Rounds) != 1 || back.Rounds[0].New[0].ID != "PQ-1" {
		t.Fatalf("round-tripped ledger = %+v", back)
	}
	if back.ContentHash != l.ContentHash {
		t.Errorf("ContentHash lost: %q want %q", back.ContentHash, l.ContentHash)
	}
	// The file must be human-readable too, not just machine-readable.
	raw, _ := os.ReadFile(planGatePath(dir, "000187-x.md"))
	if !strings.Contains(string(raw), "seam in wrong layer") {
		t.Error("rendered ledger should carry the finding in its prose")
	}
}

// A corrupt ledger must NOT be silently reset to empty — that would erase the memory the
// whole feature exists to keep, and silently re-open every disposed finding.
func TestReadPlanGateLedgerCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := planGatePath(dir, "000187-x.md")
	if err := os.WriteFile(path, []byte("---\n:::not: [valid: yaml\n---\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readPlanGateLedger(dir, "000187-x.md", 187); err == nil {
		t.Error("a corrupt ledger must error, not silently start from an empty ledger")
	}
}

// A ledger with no frontmatter at all is equally corrupt.
func TestReadPlanGateLedgerNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(planGatePath(dir, "000187-x.md"), []byte("# just prose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readPlanGateLedger(dir, "000187-x.md", 187); err == nil {
		t.Error("a ledger without frontmatter must error")
	}
}

// Identity fields are owned by the binary: a hand-edited header must not be able to
// redirect the gate's ID prefix or issue number.
func TestReadPlanGateLedgerRepairsIdentity(t *testing.T) {
	dir := t.TempDir()
	tampered := gatestate.Ledger{Gate: "bogus", IssueNum: 999, IDPrefix: "XX",
		Rounds: []gatestate.Round{{N: 1, Timestamp: "t", Agent: "a"}}}
	if err := writePlanGateLedger(dir, "000187-x.md", tampered, "ariadne"); err != nil {
		t.Fatal(err)
	}
	l, err := readPlanGateLedger(dir, "000187-x.md", 187)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if l.Gate != "plan-quality" || l.IssueNum != 187 || l.IDPrefix != "PQ" {
		t.Errorf("identity not repaired: %+v", l)
	}
	if len(l.Rounds) != 1 {
		t.Errorf("repair must preserve the rounds, got %d", len(l.Rounds))
	}
}
