package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const (
	instanceGateGood = "---\nid: 2\nstatus: open\n---\n## Spec\nspec\n## Done when\n- it works\n## Plan\n- [ ] do it\n"
	// A started issue owes Plan content; an open one does not (#308).
	instanceGateNoPlan       = "---\nid: 1\nstatus: working\n---\n## Spec\nspec\n## Done when\n- it works\n"
	instanceGateOpenSkeleton = "---\nid: 2\nstatus: open\n---\n## Spec\n\n## Done when\n\n-\n\n## Plan\n\n- [ ]\n"
)

// instanceGateRepo cuts a branch that adds branchIssue, then moves main on: main
// archives a NONCONFORMING issue that existed at the branch point (#322's shape:
// #189's landing refused on #317, archived on main after the branch was cut). It
// returns main's tip and the branch head, with the branch checked out.
func instanceGateRepo(t *testing.T, branchIssue string) (mainOID, headOID string) {
	t.Helper()
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	g := func(args ...string) string { t.Helper(); return strings.TrimSpace(testfix.Git(t, dir, args...)) }
	write := func(p, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	trunk := g("branch", "--show-current")
	write("workshop/issues/000001-old.md", instanceGateNoPlan)
	g("add", ".")
	g("commit", "-q", "-m", "old issue predates the gate")

	g("switch", "-q", "-c", "000002-new")
	write("workshop/issues/000002-new.md", branchIssue)
	g("add", ".")
	g("commit", "-q", "-m", "#2: details")
	headOID = g("rev-parse", "HEAD")

	g("switch", "-q", trunk)
	if err := os.MkdirAll("workshop/history/issues", 0o755); err != nil {
		t.Fatal(err)
	}
	g("mv", "workshop/issues/000001-old.md", "workshop/history/issues/000001-old.md")
	g("commit", "-q", "-m", "archive completed issues")
	mainOID = g("rev-parse", "HEAD")
	g("switch", "-q", "000002-new")
	return mainOID, headOID
}

func stubConformingFrontmatter(t *testing.T) {
	t.Helper()
	prev := validateFrontmatterFn
	validateFrontmatterFn = func(_, _ string) (string, bool, error) { return "", true, nil }
	t.Cleanup(func() { validateFrontmatterFn = prev })
}

// #322: main archiving a nonconforming issue after the branch point must not make
// the branch's landing validate that file. The pre-fix window (main tip..worktree)
// saw it as ADDED and refused on a file the branch never touched.
func TestLandingInstanceGate_IgnoresMainsChangesSinceBranchPoint(t *testing.T) {
	stubConformingFrontmatter(t)
	mainOID, headOID := instanceGateRepo(t, instanceGateGood)
	var out, errw bytes.Buffer
	if err := runLandingInstanceGate(mainOID, headOID, nounGates("workshop/issues"), &out, &errw, execGitRunner{}); err != nil {
		t.Fatalf("landing refused on main's own change: %v\n%s", err, errw.String())
	}
	if strings.Contains(out.String()+errw.String(), "000001-old.md") {
		t.Errorf("gate inspected main's archived issue:\n%s%s", out.String(), errw.String())
	}
	if !strings.Contains(errw.String(), "1 changed instance file(s) conform") {
		t.Errorf("gate did not validate exactly the branch's one issue:\n%s", errw.String())
	}
}

// The gate keeps its teeth: the branch's own nonconforming issue still refuses.
func TestLandingInstanceGate_RefusesBranchsOwnNonconformingIssue(t *testing.T) {
	stubConformingFrontmatter(t)
	mainOID, headOID := instanceGateRepo(t, instanceGateNoPlan)
	var out, errw bytes.Buffer
	err := runLandingInstanceGate(mainOID, headOID, nounGates("workshop/issues"), &out, &errw, execGitRunner{})
	if err == nil {
		t.Fatal("branch's own nonconforming issue landed")
	}
	if listed := out.String(); !strings.Contains(listed, "000002-new.md") || strings.Contains(listed, "000001-old.md") {
		t.Errorf("refusal is not about the branch's own issue alone:\n%s", listed)
	}
}

// #308: issue new publishes skeletons to main with placeholder Plan and
// Done-when; a branch carrying one (as filed, not yet started) must land.
func TestLandingInstanceGate_AcceptsAnOpenSkeleton(t *testing.T) {
	stubConformingFrontmatter(t)
	mainOID, headOID := instanceGateRepo(t, instanceGateOpenSkeleton)
	var out, errw bytes.Buffer
	if err := runLandingInstanceGate(mainOID, headOID, nounGates("workshop/issues"), &out, &errw, execGitRunner{}); err != nil {
		t.Fatalf("an open skeleton refused the landing: %v\n%s%s", err, out.String(), errw.String())
	}
}

func TestLandingInstanceGate_FailsClosedWithoutMergeBase(t *testing.T) {
	stubConformingFrontmatter(t)
	_, headOID := instanceGateRepo(t, instanceGateGood)
	var out, errw bytes.Buffer
	err := runLandingInstanceGate("0000000000000000000000000000000000000000", headOID, nounGates("workshop/issues"), &out, &errw, execGitRunner{})
	if err == nil || !strings.Contains(err.Error(), "merge-base") {
		t.Fatalf("unresolvable main must fail closed on the merge-base, got %v", err)
	}
}

// The production call site: runMerge must hand the instance gate the branch's own
// window — merge-base..PR head — after integrating a main that moved. The
// direct tests above also cover the helper on branches not yet integrated.
func TestLandingPassesBranchWindowToInstanceGate(t *testing.T) {
	_, remote, gh := landingFixture(t, 1)
	gh.followPush = true
	server := filepath.Join(t.TempDir(), "advance")
	git(t, "", "clone", "--quiet", remote, server)
	git(t, server, "config", "user.name", "server")
	git(t, server, "config", "user.email", "server@example.com")
	git(t, server, "commit", "--allow-empty", "-qm", "main moves on after the branch point")
	git(t, server, "push", "-q", "origin", "HEAD:main")
	gh.pr.BaseOID = procedureHead(t, server) // GitHub reports the PR base as main's tip
	integratedMain := gh.pr.BaseOID

	var gotBase, gotHead string
	prev := validateChangedInstancesFn
	validateChangedInstancesFn = func(base, head string, _ []nounGate, _, _ io.Writer) error {
		gotBase, gotHead = base, head
		return nil
	}
	t.Cleanup(func() { validateChangedInstancesFn = prev })

	f := landingFlags()
	f.NoValidate = false
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	if gotBase != integratedMain || gotHead != gh.pr.HeadOID {
		t.Fatalf("instance gate window = %s..%q, want integrated main %s..PR head %s",
			gotBase, gotHead, integratedMain, gh.pr.HeadOID)
	}
}

// #486: a landing train records only the gate's error, by its first line, so
// that line must name each nonconforming file and the rule it breaks, the
// validator's multi-line output included.
func TestLandingInstanceGate_ErrorNamesEachFileOnOneLine(t *testing.T) {
	prev := validateFrontmatterFn
	validateFrontmatterFn = func(_, _ string) (string, bool, error) {
		return "status: \"bogus\" is not one of open, working\n  at line 3", false, nil
	}
	t.Cleanup(func() { validateFrontmatterFn = prev })
	mainOID, headOID := instanceGateRepo(t, instanceGateNoPlan)
	var out, errw bytes.Buffer
	err := runLandingInstanceGate(mainOID, headOID, nounGates("workshop/issues"), &out, &errw, execGitRunner{})
	if err == nil {
		t.Fatal("nonconforming issue landed")
	}
	line := firstLine(err.Error())
	for _, want := range []string{"000002-new.md (frontmatter):", `status: "bogus" is not one of open, working at line 3`, "000002-new.md (section):"} {
		if !strings.Contains(line, want) {
			t.Errorf("first line lacks %q:\n%s", want, line)
		}
	}
}
