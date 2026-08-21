package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// A missing sidecar is the normal round-1 state — an empty ledger carrying the gate's
// identity, not an error.
func TestReadBoundaryGateLedger_MissingIsEmptyNotError(t *testing.T) {
	dir := t.TempDir()
	l, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err != nil {
		t.Fatalf("missing sidecar must not error: %v", err)
	}
	if len(l.Rounds) != 0 || l.IDPrefix != "BR" || l.Gate != "boundary-review" || l.IssueNum != 194 {
		t.Fatalf("want an empty BR ledger for #194, got %+v", l)
	}
}

// The behavior that makes the ledger trustworthy: a sidecar that EXISTS but does not
// parse is an ERROR. Silently resetting would erase every disposition and re-open
// findings the operator already addressed — the exact forgetting this gate prevents,
// and worse than the status quo because it would look like it worked.
func TestReadBoundaryGateLedger_CorruptIsErrorNotSilentReset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000194-x-close-gate.md")
	if err := os.WriteFile(path, []byte("---\nthis: [is not: valid yaml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err == nil {
		t.Fatal("a corrupt ledger must error, never silently reset to empty")
	}
	if !strings.Contains(err.Error(), "do NOT let the gate silently forget") {
		t.Errorf("the error must say why it refuses, got: %v", err)
	}
}

// The ledger is durable state the NEXT invocation reads, so the round-trip through disk
// must preserve ids, dispositions and boundaries.
func TestBoundaryGateLedger_RoundTripsThroughDisk(t *testing.T) {
	dir := t.TempDir()
	l := gatestate.Ledger{Gate: "boundary-review", IssueNum: 194, IDPrefix: "BR", Rounds: []gatestate.Round{
		{N: 1, Boundary: "M1", New: []gatestate.Finding{{ID: "BR-1", Severity: "Important", Title: "first", Round: 1}}},
		{N: 2, Boundary: "M1", Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed", Round: 2}}},
	}}
	if err := writeBoundaryGateLedger(dir, "000194-x.md", l, "ariadne"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Rounds) != 2 {
		t.Fatalf("got %d rounds, want 2", len(got.Rounds))
	}
	if open := gatestate.OpenFindings(got); len(open) != 0 {
		t.Errorf("BR-1 was disposed addressed; it must not read as open: %+v", open)
	}
	if got.Rounds[0].Boundary != "M1" {
		t.Errorf("boundary lost in round-trip: %q", got.Rounds[0].Boundary)
	}
}

// The prose sidecar and the gate ledger are DIFFERENT artifacts for the same boundary and
// must not collide — verdict.cue's `*-review.md` glob asserts "carries a boundary
// verdict", which a findings ledger does not.
func TestBoundaryGateLedger_PathDoesNotCollideWithProseSidecar(t *testing.T) {
	ledger := boundaryGatePath("workshop/plans", "000194-x.md")
	for _, milestone := range []string{"", "M1"} {
		prose := sidecarPath("workshop/plans", "000194-x.md", milestone)
		if ledger == prose {
			t.Fatalf("ledger path collides with the prose sidecar: %s", ledger)
		}
	}
	if strings.HasSuffix(ledger, "-review.md") {
		t.Errorf("the ledger must stay out of verdict.cue's *-review.md family: %s", ledger)
	}
}

// #194 M2 end-to-end: a boundary review that emits the findings fence persists a ledger,
// and the NEXT review at the same boundary is shown those findings to dispose of.
func TestBoundaryReview_PersistsLedgerAndFeedsTheNextRound(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{
		Label: "#69 M1", Base: "abc1234", BaseLong: "abc1234", Head: "def5678",
		IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir,
	}
	round := gatestate.RoundReport{New: []gatestate.Finding{
		{ID: "new", Severity: "Important", Title: "the oracle cannot see what it certifies"},
	}}
	var stderr strings.Builder
	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &round}, "2026-08-20T18:00:00-07:00")

	if !d.Block || len(d.OpenBlocking) != 1 {
		t.Fatalf("an undisposed Important must block the boundary: %+v", d)
	}
	if d.OpenBlocking[0].ID != "BR-1" {
		t.Errorf("the binary assigns the stable id, got %q", d.OpenBlocking[0].ID)
	}
	// The next round at this boundary is shown it.
	prior := boundaryPriorFindings(&stderr, p)
	if !strings.Contains(prior, "BR-1") || !strings.Contains(prior, "oracle cannot see") {
		t.Errorf("the next round must be shown BR-1 to dispose of:\n%s", prior)
	}
	// A different boundary is not — the cap and the open set scope per boundary (D1).
	other := p
	other.Milestone = "M2"
	if got := boundaryPriorFindings(&stderr, other); strings.Contains(got, "BR-1") {
		t.Errorf("M1's finding must not block M2:\n%s", got)
	}
}

// Disposing the finding clears the boundary — the convergence mechanic.
func TestBoundaryReview_DisposingAFindingClearsTheGate(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{
		Label: "#69 M1", IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir,
	}
	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Critical", Title: "boom"}},
	}}, "2026-08-20T18:00:00-07:00")

	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed", Note: "fixed"}},
	}}, "2026-08-20T18:30:00-07:00")
	if d.Block {
		t.Errorf("disposing the only blocking finding must clear the gate: %+v", d)
	}
	if d.Rounds != 2 {
		t.Errorf("both rounds must be recorded, got %d", d.Rounds)
	}
}

// #194 D2: the plan gate's still-open findings are SEEDED into this ledger on its first
// round, under this gate's id namespace — replacing code-review.md's instruction to read
// the plan-gate file directly, which would have put PQ-* and BR-* ids in one fence.
func TestBoundaryReview_SeedsDeferredPlanGateFindings(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))

	planLedger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 69, IDPrefix: "PQ", Rounds: []gatestate.Round{
		{N: 1, New: []gatestate.Finding{
			{ID: "PQ-1", Severity: "Minor", Title: "deferred to the boundary", Round: 1},
			{ID: "PQ-2", Severity: "Important", Title: "already handled", Round: 1},
		}},
		{N: 2, Dispositions: []gatestate.Disposition{{ID: "PQ-2", State: "addressed", Round: 2}}},
	}}
	if err := writePlanGateLedger(plansDir, issueFile, planLedger, "ariadne"); err != nil {
		t.Fatal(err)
	}

	p := boundaryReviewParams{IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir}
	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{}}, "2026-08-20T18:00:00-07:00")

	l, err := readBoundaryGateLedger(plansDir, issueFile, 69)
	if err != nil {
		t.Fatal(err)
	}
	open := gatestate.OpenFindings(l)
	if len(open) != 1 {
		t.Fatalf("only the still-open plan-gate finding seeds, got %+v", open)
	}
	if open[0].Severity != "Minor" || !strings.Contains(open[0].Title, "deferred to the boundary") {
		t.Errorf("seeded finding lost its identity: %+v", open[0])
	}
	if !strings.Contains(open[0].Detail, "PQ-1") {
		t.Errorf("the seeded finding must record its plan-gate origin: %q", open[0].Detail)
	}
	// D5: seeded findings are visible at EVERY boundary, since they were deferred to
	// "the boundary review" generically, not to whichever milestone ran first.
	for _, boundary := range []string{"M1", "M2", ""} {
		q := p
		q.Milestone = boundary
		if got := boundaryPriorFindings(&stderr, q); !strings.Contains(got, "deferred to the boundary") {
			t.Errorf("seeded finding must be visible at boundary %q:\n%s", boundary, got)
		}
	}
}

func mustIssuePath(t *testing.T, issuesDir string, n int) string {
	t.Helper()
	p, err := issueFilePath(issuesDir, n)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// Past the round cap, gatestate demotes non-Critical findings. At the PLAN gate that is
// safe because the boundary review picks them up; at the BOUNDARY gate there is no later
// gate, so the demotion must at least be announced rather than absorbed silently.
func TestBoundaryReview_DemotionPastCapIsAnnounced(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir}
	t.Setenv("WF_BOUNDARY_ROUND_CAP", "1")

	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Important", Title: "carried past the cap"}},
	}}, "2026-08-20T18:00:00-07:00")
	// Round 2 is past a cap of 1.
	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{}}, "2026-08-20T18:30:00-07:00")

	if len(d.Demoted) != 1 {
		t.Fatalf("the Important finding should be demoted past the cap, got %+v", d)
	}
	if d.Block {
		t.Error("a demoted finding must not block")
	}
	if !strings.Contains(stderr.String(), "no later gate picks it up") {
		t.Errorf("the demotion must be announced — it ships having blocked nothing:\n%s", stderr.String())
	}
}

// ARCH-MOCK, raised by M1's boundary review: `judge.Run` is a STATELESS override, which
// modelled the dependency adequately while a boundary review was a single call. M2 makes
// the reviewer stateful — ledger in, dispositions out — so a canned-output fake stops
// modelling it. This fake carries round state: it reads the prior-findings block out of
// the prompt it is given and responds the way a converging reviewer would.
type fakeReviewer struct {
	round      int
	sawPrior   []string // the prior-findings block observed on each round
	raiseOnce  string   // finding title raised on round 1
	disposeIDs []string // ids disposed on round 2+
}

func (f *fakeReviewer) install(t *testing.T) {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		f.round++
		prompt := strings.Join(args, "\n")
		f.sawPrior = append(f.sawPrior, prompt)
		if f.round == 1 {
			return []byte("```verdict\nverdict: REWORK\nconfidence: high\n```\n\n" +
				"```findings\nfindings:\n  - id: new\n    severity: Critical\n    title: |\n      " +
				f.raiseOnce + "\n```\n"), nil
		}
		var b strings.Builder
		b.WriteString("```verdict\nverdict: SHIP\nconfidence: high\n```\n\n```findings\ndispose:\n")
		for _, id := range f.disposeIDs {
			b.WriteString("  - id: " + id + "\n    disposition: addressed\n    note: |\n      fixed\n")
		}
		b.WriteString("```\n")
		return []byte(b.String()), nil
	}
}

// The convergence loop end to end: round 1 raises and blocks, round 2 is SHOWN what
// round 1 said, disposes it, and the gate clears. This is the behavior #195 asked for
// and the reason the two issues merged.
func TestBoundaryReview_ConvergesAcrossRounds(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{
		Label: "#69 M1", IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir,
	}
	// Real refs: dispatch collects a real diff, so a fabricated base makes it bail
	// before the reviewer is ever invoked.
	base := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	if err := os.WriteFile("m1.go", []byte("package main // work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, "", "add", "m1.go")
	git(t, "", "commit", "-q", "-m", "#69 M1: the work under review")
	p.BaseLong, p.Base = base, shortSHA(base)
	p.Head = strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))

	fake := &fakeReviewer{raiseOnce: "the measurement cannot fail in the direction that matters"}
	fake.install(t)

	var out, errb bytes.Buffer
	r1 := dispatchBoundaryReview(&out, &errb, p)
	d1 := persistBoundaryRound(&errb, p, r1, "2026-08-20T18:00:00-07:00")
	if !d1.Block {
		t.Fatalf("round 1 raised a Critical; the gate must block: %+v", d1)
	}

	fake.disposeIDs = []string{d1.OpenBlocking[0].ID}
	// Mirror production: the caller reads the prior-findings block UNDER THE LOCK and
	// carries it on the params. dispatchBoundaryReview does not look it up itself —
	// see C1 in M2's boundary review for what happened when it did.
	p.PriorFindings = boundaryPriorFindings(&errb, p)
	r2 := dispatchBoundaryReview(&out, &errb, p)
	d2 := persistBoundaryRound(&errb, p, r2, "2026-08-20T18:30:00-07:00")
	if d2.Block {
		t.Fatalf("round 2 disposed the finding; the gate must clear: %+v", d2)
	}

	// The load-bearing assertion: round 2's PROMPT carried round 1's finding. Without
	// it the reviewer is memoryless and would renumber a fresh C1 instead of disposing.
	if len(fake.sawPrior) < 2 {
		t.Fatal("expected two dispatches")
	}
	if strings.Contains(fake.sawPrior[0], "the measurement cannot fail") {
		t.Error("round 1 should have had no prior findings to show")
	}
	if !strings.Contains(fake.sawPrior[1], "the measurement cannot fail") {
		t.Error("round 2's prompt must carry round 1's finding — that is the whole mechanism")
	}
}

// #194 M2 boundary review C1: the convergence test above dispatches directly, which is
// NOT how production reaches the reviewer. reviewThenFinalizeLocked blanks PlansDir
// before dispatch (to defer the sidecar repo write), so a dispatch-time prior-findings
// lookup keyed on PlansDir returned "" in 100% of live reviews — the ledger blocked on
// findings the reviewer was never shown and therefore could not dispose.
//
// This test drives the real command so the wiring itself is pinned.
func TestCloseCommand_LiveReviewSeesPriorFindings(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))

	// A prior round left an open finding at the whole-issue boundary.
	seeded := gatestate.Ledger{Gate: "boundary-review", IssueNum: 69, IDPrefix: "BR", Rounds: []gatestate.Round{
		{N: 1, Boundary: "", New: []gatestate.Finding{
			{ID: "BR-1", Severity: "Critical", Title: "PRIOR_FINDING_MARKER", Round: 1},
		}},
	}}
	if err := writeBoundaryGateLedger(plansDir, issueFile, seeded, "ariadne"); err != nil {
		t.Fatal(err)
	}

	var prompt string
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		prompt = strings.Join(args, "\n")
		return []byte("```verdict\nverdict: SHIP\nconfidence: high\n```\n\n" +
			"```findings\ndispose:\n  - id: BR-1\n    disposition: addressed\n    note: |\n      fixed\n```\n"), nil
	}

	_, _, _ = executeSDLCTestCommand("close", "--issue", "69", "--actual", "1",
		"--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir,
		"--plans-dir", plansDir, "--brain-dir", "../nonexistent-brain")

	if !strings.Contains(prompt, "PRIOR_FINDING_MARKER") {
		t.Fatal("the LIVE review path must show the reviewer its prior findings — " +
			"without this the ledger blocks on findings that cannot be disposed")
	}
	if !strings.Contains(prompt, "BR-1") {
		t.Error("the prompt must carry the stable id so the reviewer can dispose it by name")
	}
}

// #194 M2 review BR-11: persistBoundaryRound's new unconditional operator lines are the
// same class as the info lines M1's I5 put under this guard — gatesig classifies
// transcripts by substring, so a line colliding with a GateCatalog Ack/Refusal pattern
// corrupts friction attribution (#172).
func TestBoundaryGateOperatorLines_NoGatesigCollision(t *testing.T) {
	for _, line := range []string{
		"boundary gate: carried 3 deferred plan-gate finding(s) into this issue's ledger (#194)",
		"boundary gate ledger: workshop/plans/000194-x-close-gate.md",
		"boundary gate: [BR-2] Important demoted past the round cap and will NOT block — no later gate picks it up: some title",
		"boundary gate ledger unusable — refusing to finalize rather than close without it: parse error",
	} {
		assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+line)
	}
}

// #194 M2 review BR-14: the gate-ledger refusal, --no-ledger, and blockOnLedgerFailure
// shipped with no test at any level — the refusal is the whole point of the ledger, and
// blockOnLedgerFailure is the fail-closed answer to the worst failure mode a gate has.
func TestGateLedgerRefusal_BlocksAPassingVerdictAndIsBypassable(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))
	seeded := gatestate.Ledger{Gate: "boundary-review", IssueNum: 69, IDPrefix: "BR", Rounds: []gatestate.Round{
		{N: 1, Boundary: "", New: []gatestate.Finding{
			{ID: "BR-1", Severity: "Critical", Title: "undisposed and blocking", Round: 1},
		}},
	}}
	if err := writeBoundaryGateLedger(plansDir, issueFile, seeded, "ariadne"); err != nil {
		t.Fatal(err)
	}

	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	// A SHIP verdict that does NOT dispose the open finding — the reviewer contradicting
	// itself, which is the surprising case operators hit.
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		return []byte("```verdict\nverdict: SHIP\nconfidence: high\n```\n\n```findings\n```\n"), nil
	}

	args := []string{"close", "--issue", "69", "--actual", "1", "--verified", "tests pass",
		"--no-atlas", "--issues-dir", issuesDir, "--plans-dir", plansDir, "--brain-dir", "../nonexistent-brain"}
	_, stderr, err := executeSDLCTestCommand(args...)
	if err == nil {
		t.Fatal("a SHIP verdict with an undisposed Critical must NOT finalize — verdict AND ledger must both clear")
	}
	if !strings.Contains(stderr, "BR-1") {
		t.Errorf("the refusal must name the blocking finding:\n%s", stderr)
	}
	if strings.Contains(readIssue(t, issuesDir), "status: codecomplete") {
		t.Error("the close finalized despite the ledger refusal")
	}

	// --no-ledger waives exactly this refusal (AGENTS.md §5's per-gate bypass).
	_, stderr2, err2 := executeSDLCTestCommand(append(args, "--no-ledger")...)
	if err2 != nil {
		t.Fatalf("--no-ledger must waive the ledger refusal, got: %v\n%s", err2, stderr2)
	}
	if !strings.Contains(stderr2, "--no-ledger") {
		t.Errorf("the bypass must log an explicit acknowledgment:\n%s", stderr2)
	}
}

// blockOnLedgerFailure: an unusable ledger must FAIL CLOSED. Returning a zero Decision
// would mean Block:false — findings dropped, corrupt file unwritten, close finalizing —
// the exact anti-behavior readGateLedger refuses at a finer grain.
func TestBlockOnLedgerFailure_FailsClosed(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))
	if err := os.WriteFile(filepath.Join(plansDir, strings.TrimSuffix(issueFile, ".md")+"-close-gate.md"),
		[]byte("---\nbroken: [yaml: here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	d := persistBoundaryRound(&stderr, boundaryReviewParams{
		IssuesDir: issuesDir, IssueNum: 69, Milestone: "", PlansDir: plansDir,
	}, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{}}, "2026-08-20T22:00:00-07:00")

	if !d.Block {
		t.Fatal("an unusable ledger must block, never silently pass the gate")
	}
	if !strings.Contains(stderr.String(), "refusing to finalize rather than close without it") {
		t.Errorf("the refusal must say why:\n%s", stderr.String())
	}
}

// #194 M3: the convergence line reaches the operator, and says NOT converging when a
// family repeats — the signal that was missing when tools#1 ran four rounds blind.
func TestBoundaryReview_EmitsConvergenceLine(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir}

	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Critical", Title: "paren case", Family: "block-opener-rule"}},
	}}, "2026-08-20T22:00:00-07:00")

	stderr.Reset()
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed"}},
		// Same family, different spelling — normalization must still see the repeat.
		New: []gatestate.Finding{{ID: "new", Severity: "Important", Title: "bracket case", Family: "Block Opener Rule"}},
	}}, "2026-08-20T22:30:00-07:00")

	out := stderr.String()
	if !strings.Contains(out, "round 2") || !strings.Contains(out, "1 repeat family") {
		t.Errorf("the convergence line must reach the operator and count the repeat:\n%s", out)
	}
	if !strings.Contains(out, "Not converging") {
		t.Errorf("a repeat family means not converging:\n%s", out)
	}
}

// #194 M3 Task 3.6: M4 was considered and REJECTED — the boundary review keeps reading
// the whole branch. This test is what keeps it rejected: a whole-issue close resolves its
// window to merge-base(main, HEAD), and no later change may quietly narrow it.
func TestBoundaryWindowBase_WholeIssueStaysAtMergeBase(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 194)
	runGit("checkout", "-q", "-b", "000194-work")
	commitTouchingIssue(t, runGit, issuePath, "w1", "#194 M1: first", "")
	// A finalized boundary exists — under M4 this would have become the new base.
	commitTouchingIssue(t, runGit, issuePath, "w2", "#194 M1: close",
		"Done.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..def5678")
	commitTouchingIssue(t, runGit, issuePath, "w3", "#194 M2: more work", "")

	mergeBase := strings.TrimSpace(captureGit(t, "merge-base", "main", "HEAD"))
	if got := boundaryWindowBase("194", "", issuePath); got != mergeBase {
		t.Fatalf("a whole-issue close must review the WHOLE branch: base = %q, want merge-base %q\n"+
			"(M4 — round-scoping a re-review — was rejected precisely so no reviewer ever "+
			"sees less than the integrated result)", got, mergeBase)
	}
}

// #194 M3 review BR-20: family counts must come from the WHOLE ledger, not the
// boundary-filtered view. Otherwise a family can never be seen recurring across
// milestones — which is precisely the tools#1 evidence this feature was built from — and
// at the whole-issue close (scope "") every milestone round drops out entirely.
// Tested through boundaryPriorFindings, the path production actually uses.
func TestBoundaryPriorFindings_FamiliesSpanMilestones(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))
	if err := writeBoundaryGateLedger(plansDir, issueFile, gatestate.Ledger{
		Gate: "boundary-review", IssueNum: 69, IDPrefix: "BR", Rounds: []gatestate.Round{
			{N: 1, Boundary: "M1", New: []gatestate.Finding{
				{ID: "BR-1", Severity: "Critical", Title: "paren case", Family: "block-opener-rule", Round: 1},
			}},
			{N: 2, Boundary: "M1", Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed", Round: 2}}},
		},
	}, "ariadne"); err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder

	// A LATER milestone must still see the family, and be escalated on it.
	m2 := boundaryPriorFindings(&stderr, boundaryReviewParams{
		IssuesDir: issuesDir, IssueNum: 69, Milestone: "M2", PlansDir: plansDir})
	if !strings.Contains(m2, "block-opener-rule") {
		t.Errorf("M2 must see M1's family — cross-milestone recurrence is the whole point:\n%s", m2)
	}
	if !strings.Contains(m2, "state the rule") {
		t.Errorf("M2 must be escalated on the repeat family:\n%s", m2)
	}

	// And so must the whole-issue close, whose scope is "" — the case that dropped
	// EVERY milestone round before this fix.
	closeBlock := boundaryPriorFindings(&stderr, boundaryReviewParams{
		IssuesDir: issuesDir, IssueNum: 69, Milestone: "", PlansDir: plansDir})
	if !strings.Contains(closeBlock, "block-opener-rule") {
		t.Errorf("the whole-issue close must see families raised at milestones:\n%s", closeBlock)
	}

	// The scoping of what must be DISPOSED is unchanged: M2 does not inherit M1's
	// open findings.
	if strings.Contains(m2, "BR-1") && strings.Contains(m2, "MUST dispose") {
		if idx := strings.Index(m2, "OPEN FINDINGS"); idx >= 0 && strings.Contains(m2[idx:idx+400], "BR-1") {
			t.Error("M2 must not inherit M1's open findings — only the family vocabulary crosses")
		}
	}
}
