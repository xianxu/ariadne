package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const quickIssueTmpl = "---\nid: %s\nstatus: working\nestimate_hours:\n---\n\n# x\n\n## Spec\n\nThing.\n\n" +
	"## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n"

// quickCloseRepo builds a temp repo whose issue carries a flow record written
// the way change-code writes it (decideChangeCodeFlow, so the contract hashes
// are real), then commits code under the issue. Returns the issues dir.
func quickCloseRepo(t *testing.T, issueNum int, pin string, code map[string]string) string {
	t.Helper()
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	git := func(args ...string) { t.Helper(); testfix.Git(t, dir, args...) }
	write := func(files map[string]string) {
		for p, text := range files {
			os.MkdirAll(filepath.Dir(p), 0o755)
			if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	padded := fmt.Sprintf("%06d", issueNum)
	d, err := decideChangeCodeFlow(fmt.Sprintf(quickIssueTmpl, padded), "", pin)
	if err != nil {
		t.Fatal(err)
	}
	if d.flow.Kind() != flow.Quick {
		t.Fatalf("fixture flow %+v, want quick", d.flow)
	}
	write(map[string]string{filepath.Join("workshop/issues", padded+"-x.md"): d.content})
	git("add", ".")
	git("commit", "-q", "-m", fmt.Sprintf("#%d: issue-sync: spec/plan at change-code", issueNum))
	if len(code) > 0 {
		write(code)
		git("add", ".")
		git("commit", "-q", "-m", fmt.Sprintf("#%d: implement", issueNum))
	}
	return "workshop/issues"
}

func goLines(n int) string { return strings.Repeat("var _ = 1\n", n) }

func quickFlags(issuesDir string, issueNum int) *closeFlags {
	return &closeFlags{Issue: issueNum, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
}

func issueFlowAfterClose(t *testing.T, issuesDir string) (flow.Flow, string) {
	t.Helper()
	text := readQuick(t, issuesDir)
	fm, _, err := issue.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	f, _, err := flow.FromFrontmatter(fm)
	if err != nil {
		t.Fatal(err)
	}
	return f, text
}

const smallDiffMarker = "## Small-diff focus"

// TestCloseQuickWithinShellSelectsSmallDiff: a quick issue whose diff stays in
// the shell is reviewed with the small-diff recipe, and its record stays quick.
func TestCloseQuickWithinShellSelectsSmallDiff(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(20), "cmd/a_test.go": goLines(300)})
	calls, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if *calls != 1 || !strings.Contains(*prompt, smallDiffMarker) {
		t.Fatalf("calls=%d; want one dispatch of the small-diff recipe", *calls)
	}
	f, text := issueFlowAfterClose(t, dir)
	if f.Kind() != flow.Quick || strings.Contains(text, "flow upgraded") {
		t.Errorf("flow after close = %+v, want quick with no upgrade:\n%s", f, text)
	}
}

// TestCloseQuickCrossingShellUpgrades: three code files leave the shell — the
// close records full/inferred with the measured reason and runs the full review.
func TestCloseQuickCrossingShellUpgrades(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5), "cmd/b.go": goLines(5), "cmd/c.go": goLines(5)})
	calls, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if *calls != 1 || strings.Contains(*prompt, smallDiffMarker) || !strings.Contains(*prompt, "each of the 8 entries") {
		t.Fatalf("want one dispatch of the FULL recipe (calls=%d)", *calls)
	}
	f, text := issueFlowAfterClose(t, dir)
	if f.Kind() != flow.Full || f.Provenance() != flow.Inferred {
		t.Errorf("flow after close = %+v, want full/inferred", f)
	}
	if !strings.Contains(text, "flow upgraded quick → full") || !strings.Contains(text, "3 code files") {
		t.Errorf("the Log does not carry the measured reason:\n%s", text)
	}
}

// TestCloseOperatorQuickStillUpgrades: the shell is hard — an operator pin to
// quick does not exempt a diff that leaves it.
func TestCloseOperatorQuickStillUpgrades(t *testing.T) {
	dir := quickCloseRepo(t, 231, "quick", map[string]string{"cmd/a.go": goLines(flow.MaxChangedLines + 1)})
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Full || !strings.Contains(text, "added lines") {
		t.Errorf("operator quick over the line limit: flow %+v; want full with the line reason:\n%s", f, text)
	}
}

// TestCloseQuickReworkThenReclose: a REWORK writes nothing (#139) — the record
// stays quick on disk — and the re-close upgrades it (from the same window here;
// TestCloseQuickReworkShrinkThenReclose covers a fix that shrinks it).
func TestCloseQuickReworkThenReclose(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5), "cmd/b.go": goLines(5), "cmd/c.go": goLines(5)})
	before := readQuick(t, dir)
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nno\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err == nil {
		t.Fatal("a REWORK verdict should not finalize")
	}
	if after := readQuick(t, dir); after != before {
		t.Fatalf("REWORK wrote the issue:\n%s", after)
	}
	_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, _ := issueFlowAfterClose(t, dir); f.Kind() != flow.Full || strings.Contains(*prompt, smallDiffMarker) {
		t.Errorf("re-close after REWORK: flow %+v — the upgrade must be re-derived", f)
	}
}

// TestCloseQuickEmptyDoneWhenRefuses: on the quick flow Done-when is the only
// oracle, so close refuses without a bullet.
func TestCloseQuickEmptyDoneWhenRefuses(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5)})
	path := filepath.Join(dir, "000231-x.md")
	text := readQuick(t, dir)
	os.WriteFile(path, []byte(strings.Replace(text, "- it works", "-", 1)), 0o644)
	calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n")
	msg, died := expectDie(t, func() { runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)) })
	if !died || !strings.Contains(msg, "Done when") || *calls != 0 {
		t.Errorf("died=%v calls=%d msg=%q; want a refusal naming Done when, before any review", died, *calls, msg)
	}
}

// TestCloseQuickStaleDoneWhenRefuses: the contract moved (a Spec reframe) but
// Done-when did not — refused, with the skip flag named; the skip lets it through.
func TestCloseQuickStaleDoneWhenRefuses(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5)})
	path := filepath.Join(dir, "000231-x.md")
	os.WriteFile(path, []byte(strings.Replace(readQuick(t, dir), "Thing.", "A reframed thing.", 1)), 0o644)
	calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n")
	msg, died := expectDie(t, func() { runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)) })
	if !died || !strings.Contains(msg, "--no-done-when-fresh") || *calls != 0 {
		t.Fatalf("died=%v calls=%d msg=%q; want a refusal naming --no-done-when-fresh", died, *calls, msg)
	}
	f := quickFlags(dir, 231)
	f.NoDoneWhenFresh = true
	if err := runCloseWithReview(io.Discard, io.Discard, f); err != nil || *calls != 1 {
		t.Errorf("with --no-done-when-fresh: err=%v calls=%d, want the review to run", err, *calls)
	}
}

// TestCloseNoFlowUnchanged: an issue with no record (every issue before #231)
// closes exactly as before — the full recipe, nothing flow-related written.
func TestCloseNoFlowUnchanged(t *testing.T) {
	dir := closeRepo(t, 69)
	_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, closeFlagsFor(dir)); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(*prompt, smallDiffMarker) {
		t.Error("an issue with no flow record got the small-diff recipe")
	}
	if text := readIssue(t, dir); strings.Contains(text, "flow:") || strings.Contains(text, "flow upgraded") {
		t.Errorf("close wrote flow state onto an issue that had none:\n%s", text)
	}
}

// TestMilestoneCloseUpgradesQuick: an Mx row on a quick issue crosses the shell
// (a single boundary is part of what quick means) — milestone-close upgrades it.
func TestMilestoneCloseUpgradesQuick(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5)})
	path := filepath.Join(dir, "000231-x.md")
	os.WriteFile(path, []byte(strings.Replace(readQuick(t, dir), "- [x] do it", "- [x] do it\n- [ ] M1 — first boundary", 1)), 0o644)
	f := &milestoneCloseFlags{Issue: 231, Milestone: "M1", Actual: "1", Verified: "ok", NoAtlas: true, NoJudge: true,
		IssuesDir: dir, BrainDir: "../nonexistent-brain"}
	if err := runMilestoneClose(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	if fl, text := issueFlowAfterClose(t, dir); fl.Kind() != flow.Full || !strings.Contains(text, "milestones") {
		t.Errorf("milestone-close on a quick issue: flow %+v, want full with the Mx reason:\n%s", fl, text)
	}
}

// TestCloseFlowLinesNoGatesigCollision: the flow lines close prints must not
// read as gate bypasses or refusals to the friction instrument (#172).
func TestCloseFlowLinesNoGatesigCollision(t *testing.T) {
	quick := closeFlowOutcome{flow: mustFlow(t, "{kind: quick, provenance: inferred}"), size: flow.Size{CodeFiles: []string{"a.go"}, AddedLines: 3}}
	up := closeFlowOutcome{flow: flow.Upgrade(quick.flow), crossings: []string{"3 code files changed (limit 2): a, b, c"}}
	for _, o := range []closeFlowOutcome{quick, up} {
		assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+closeFlowLine(o))
	}
}

// readQuick reads the quick fixtures' issue (#231); readIssue is pinned to #69.
func readQuick(t *testing.T, issuesDir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(issuesDir, "000231-x.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestCloseQuickReworkShrinkThenReclose is #231 BR-18's reproduction: three code
// files get the full review; it returns REWORK; the fix DELETES a file, so the
// net diff is back inside the shell. The next round must still be the full
// review — the earlier full round is recorded in the boundary ledger, and a
// quick issue that already needed the full review does not get easier to close
// by shrinking.
func TestCloseQuickReworkShrinkThenReclose(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5), "cmd/b.go": goLines(5), "cmd/c.go": goLines(5)})
	_, first := stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nno\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err == nil {
		t.Fatal("a REWORK verdict should not finalize")
	}
	if strings.Contains(*first, smallDiffMarker) {
		t.Fatal("precondition: the first round must be the full review")
	}
	os.Remove("cmd/c.go")
	testfixGit(t, "add", "-A")
	testfixGit(t, "commit", "-q", "-m", "#231: drop c")

	_, second := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(*second, smallDiffMarker) {
		t.Error("after a full-review REWORK, the shrunk re-close got the small-diff recipe")
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Full || !strings.Contains(text, "earlier round") {
		t.Errorf("flow %+v; want full with the earlier-round reason:\n%s", f, text)
	}
}

// TestFullRoundIn: which recorded rounds count as the full review.
func TestFullRoundIn(t *testing.T) {
	for _, c := range []struct {
		name   string
		rounds []gatestate.Round
		want   bool
	}{
		{"no rounds", nil, false},
		{"a small-diff round", []gatestate.Round{{Recipe: string(judge.SmallDiffReview)}}, false},
		{"a full round", []gatestate.Round{{Recipe: string(judge.MilestoneReview)}}, true},
		{"an unstamped round predates #231: full", []gatestate.Round{{}}, true},
		{"a round that never ran", []gatestate.Round{{NoCap: true}}, false},
	} {
		if got := fullRoundIn(c.rounds); got != c.want {
			t.Errorf("%s: fullRoundIn = %v, want %v", c.name, got, c.want)
		}
	}
}

func testfixGit(t *testing.T, args ...string) {
	t.Helper()
	wd, _ := os.Getwd()
	testfix.Git(t, wd, args...)
}

// TestCloseQuickSmallDiffReworkStaysQuick: the other half of the stamp — a
// REWORK from the SMALL-DIFF review must not read as a full round, or every
// quick issue that needed one fix would be upgraded for nothing.
func TestCloseQuickSmallDiffReworkStaysQuick(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(5)})
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nno\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err == nil {
		t.Fatal("a REWORK verdict should not finalize")
	}
	_, second := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(*second, smallDiffMarker) {
		t.Error("a small-diff REWORK sent the re-close to the full review")
	}
	if f, _ := issueFlowAfterClose(t, dir); f.Kind() != flow.Quick {
		t.Errorf("flow %+v, want quick", f)
	}
}

// TestCloseNonASCIIPathsClassifyAsThemselves: git quotes non-ASCII paths unless
// told not to, and a quoted "docs/caf\303\251.md" is not a doc to the classifier.
// Two such docs beside one code file must stay inside the shell.
func TestCloseNonASCIIPathsClassifyAsThemselves(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{
		"cmd/a.go": goLines(5), "docs/café.md": "x\n", "docs/naïve.md": "y\n", "tests/überprüfung_spec.lua": goLines(300)})
	_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Quick || !strings.Contains(*prompt, smallDiffMarker) {
		t.Errorf("non-ASCII docs/tests counted as code: flow %+v\n%s", f, text)
	}
}

// TestCloseQuotedNameCodeFileLinesCount: a code file whose name git quotes
// without -z (a double quote in it) must still have its lines counted — over
// the limit, it upgrades.
func TestCloseQuotedNameCodeFileLinesCount(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a\"b.go": goLines(flow.MaxChangedLines + 50)})
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Full || !strings.Contains(text, "added lines") {
		t.Errorf("a quoted-name code file over the line limit stayed %+v:\n%s", f, text)
	}
}
