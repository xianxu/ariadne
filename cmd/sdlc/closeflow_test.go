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

// overLineLimit is a window that leaves the shell by its added lines alone,
// spread over two files so neither file is over the limit by itself.
func overLineLimit() map[string]string {
	return map[string]string{"cmd/a.go": goLines(60), "cmd/b.go": goLines(60)}
}

// TestCloseQuickCrossingShellUpgrades: 120 added lines leave the shell — the
// close records full/inferred with the measured reason and runs the full review.
func TestCloseQuickCrossingShellUpgrades(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", overLineLimit())
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
	if !strings.Contains(text, "flow upgraded quick → full") || !strings.Contains(text, "120 added lines") {
		t.Errorf("the Log does not carry the measured reason:\n%s", text)
	}
}

// TestCloseSpreadButTinyAtTheLimit pins the shell's edge at close, on a diff
// spread across many code files (pair#283's shape): exactly MaxAddedLines stays
// quick however many files carry them, one line more upgrades. The file count
// is not a limit (operator, 2026-09-18).
func TestCloseSpreadButTinyAtTheLimit(t *testing.T) {
	const nFiles = 7
	spread := func(total int) map[string]string {
		code := map[string]string{}
		for i := 0; i < nFiles; i++ {
			n := total / nFiles
			if i == 0 {
				n += total % nFiles
			}
			code[fmt.Sprintf("cmd/f%d.go", i)] = goLines(n)
		}
		return code
	}
	for _, c := range []struct {
		name  string
		total int
		want  flow.Kind
	}{
		{"exactly at the limit", flow.MaxAddedLines, flow.Quick},
		{"one line past it", flow.MaxAddedLines + 1, flow.Full},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := quickCloseRepo(t, 231, "", spread(c.total))
			_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
			if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
				t.Fatal(err)
			}
			f, text := issueFlowAfterClose(t, dir)
			if f.Kind() != c.want || strings.Contains(*prompt, smallDiffMarker) != (c.want == flow.Quick) {
				t.Errorf("%d lines over %d code files: flow %+v, want %s with its recipe:\n%s", c.total, nFiles, f, c.want, text)
			}
		})
	}
}

// TestCloseDesignGrewPastTheLimit: the design limit holds at close too. A quick
// issue whose durable plan grew during the work is measured as it stands at
// close — exactly at the limit stays quick, one line past upgrades — with the
// same Crossings change-code used at entry.
func TestCloseDesignGrewPastTheLimit(t *testing.T) {
	// quickIssueTmpl's design is 2 lines: one of Spec, one of Plan.
	for _, c := range []struct {
		name      string
		planLines int
		want      flow.Kind
	}{
		{"exactly at the limit", flow.MaxDesignLines - 2, flow.Quick},
		{"one line past it", flow.MaxDesignLines - 1, flow.Full},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := quickCloseRepo(t, 231, "", map[string]string{"cmd/a.go": goLines(3)})
			plan := filepath.Join("workshop/plans", "000231-x-plan.md")
			os.MkdirAll(filepath.Dir(plan), 0o755)
			if err := os.WriteFile(plan, []byte(strings.Repeat("step\n", c.planLines)), 0o644); err != nil {
				t.Fatal(err)
			}
			testfix.Git(t, ".", "add", ".")
			testfix.Git(t, ".", "commit", "-q", "-m", "#231: plan")
			_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
			if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
				t.Fatal(err)
			}
			f, text := issueFlowAfterClose(t, dir)
			if f.Kind() != c.want || strings.Contains(*prompt, smallDiffMarker) != (c.want == flow.Quick) {
				t.Errorf("a %d-line plan: flow %+v, want %s with its recipe:\n%s", c.planLines, f, c.want, text)
			}
			if c.want == flow.Full && !strings.Contains(text, "a design of") {
				t.Errorf("upgrade without the design reason in the Log:\n%s", text)
			}
		})
	}
}

// TestCloseOperatorQuickStillUpgrades: the shell is hard — an operator pin to
// quick does not exempt a diff that leaves it.
func TestCloseOperatorQuickStillUpgrades(t *testing.T) {
	dir := quickCloseRepo(t, 231, "quick", map[string]string{"cmd/a.go": goLines(flow.MaxAddedLines + 1)})
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
	dir := quickCloseRepo(t, 231, "", overLineLimit())
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
	quick := closeFlowOutcome{flow: mustFlow(t, "{kind: quick, provenance: inferred}"), size: flow.Size{AddedLines: 3}}
	up := closeFlowOutcome{flow: flow.Upgrade(quick.flow), crossings: flow.Size{AddedLines: flow.MaxAddedLines + 1}.Crossings()}
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

// TestCloseQuickReworkShrinkThenReclose is #231 BR-18's reproduction: a window
// over the line limit gets the full review; it returns REWORK; the fix DELETES a
// file, so the net diff is back inside the shell. The next round must still be
// the full review — the earlier full round is recorded in the boundary ledger,
// and a quick issue that already needed the full review does not get easier to
// close by shrinking.
func TestCloseQuickReworkShrinkThenReclose(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", overLineLimit())
	_, first := stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nno\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err == nil {
		t.Fatal("a REWORK verdict should not finalize")
	}
	if strings.Contains(*first, smallDiffMarker) {
		t.Fatal("precondition: the first round must be the full review")
	}
	os.Remove("cmd/b.go")
	testfixGit(t, "add", "-A")
	testfixGit(t, "commit", "-q", "-m", "#231: drop b")

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
// Docs and a test with such names, each over the line limit by itself, beside
// one small code file must stay inside the shell.
func TestCloseNonASCIIPathsClassifyAsThemselves(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{
		"cmd/a.go": goLines(5), "docs/café.md": goLines(150), "docs/naïve.md": goLines(150), "tests/überprüfung_spec.lua": goLines(300)})
	_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Quick || !strings.Contains(*prompt, smallDiffMarker) {
		t.Errorf("non-ASCII docs/tests counted as code: flow %+v\n%s", f, text)
	}
}

// TestCloseQuotedDocAndTestNamesDoNotCount pins numstat's -z (#231 BR-32).
// Without it git quotes a name holding a double quote, the quoted path no longer
// reads as a doc or a test, and its lines count as code. Here 300 such lines
// sit beside a 5-line code file, so a lost -z upgrades a change that is quick.
// (The earlier fixture, a quoted CODE file, could not fail once the shell
// stopped intersecting names: a quoted code path still defaults to code.)
func TestCloseQuotedDocAndTestNamesDoNotCount(t *testing.T) {
	dir := quickCloseRepo(t, 231, "", map[string]string{
		"cmd/a.go":            goLines(5),
		"docs/a\"b.md":        goLines(150),
		"tests/x\"y_spec.lua": goLines(150),
	})
	_, prompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nok\n")
	if err := runCloseWithReview(io.Discard, io.Discard, quickFlags(dir, 231)); err != nil {
		t.Fatal(err)
	}
	if f, text := issueFlowAfterClose(t, dir); f.Kind() != flow.Quick || !strings.Contains(*prompt, smallDiffMarker) {
		t.Errorf("quoted doc and test names counted as code: flow %+v\n%s", f, text)
	}
}
