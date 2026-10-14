package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestParseLocalChecksWeight(t *testing.T) {
	d, err := parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"a","command":["x"]},{"name":"b","command":["y"],"weight":2}]}`))
	if err != nil || d.Checks[0].Weight != 1 || d.Checks[1].Weight != 2 {
		t.Fatalf("%+v %v", d, err)
	}
	for _, w := range []string{"-1", "9"} {
		if _, err := parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"a","command":["x"],"weight":` + w + `}]}`)); err == nil || !strings.Contains(err.Error(), "weight") {
			t.Errorf("weight %s: %v", w, err)
		}
	}
}

func TestEffectiveLocalChecks(t *testing.T) {
	none := localCheckDecl{}
	got := effectiveLocalChecks(none, true, 2)
	c, ok := got.check(suiteCheckName)
	if !ok || !c.Required || c.Weight != 2 || strings.Join(c.Command, " ") != "make test" {
		t.Fatalf("default suite: %+v", got)
	}
	if got := effectiveLocalChecks(none, false, 2); len(got.Checks) != 0 {
		t.Fatalf("no make test, no suite: %+v", got)
	}
	declared := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "suite", Command: []string{"go", "test", "./..."}, Required: true, Weight: 1}}}
	if got := effectiveLocalChecks(declared, true, 2); len(got.Checks) != 1 || got.Checks[0].Command[0] != "go" {
		t.Fatalf("a declared suite must win: %+v", got)
	}
}

func TestParseMakeTestFacts(t *testing.T) {
	has, w, err := parseMakeTestFacts("# GNU Make\nfoo: bar\ntest: build\n__SDLC_SUITE_WEIGHT=2\n")
	if err != nil || !has || w != 2 {
		t.Fatalf("%v %d %v", has, w, err)
	}
	if has, w, err := parseMakeTestFacts("build:\n__SDLC_SUITE_WEIGHT=1\n"); err != nil || has || w != 1 {
		t.Fatalf("no test: %v %d %v", has, w, err)
	}
	if _, _, err := parseMakeTestFacts("test:\n"); err == nil {
		t.Fatal("missing weight line accepted")
	}
}

// R1 PQ-1: the default suite's facts come from main, so a branch that drops
// its test target or lowers its weight still owes main's suite; and the probe
// never takes the machine lock (test is never its goal).
func TestDefaultSuiteFactsComeFromMain(t *testing.T) {
	dir := testfix.Repo(t, testfix.InitialCommit())
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
	marker := filepath.Join(t.TempDir(), "lock-taken")
	mainMakefile := "WF_TEST_WEIGHT ?= 2\nifneq ($(filter test,$(MAKECMDGOALS)),)\n$(shell touch " + marker + ")\nendif\ntest:\n\t@true\n"
	writeRepoFile(t, dir, "Makefile", mainMakefile)
	testfix.Git(t, dir, "add", "-A")
	testfix.Git(t, dir, "commit", "-qm", "suite on main")
	testfix.Git(t, dir, "branch", "-M", "main")
	testfix.Git(t, dir, "switch", "-qc", "feature")
	writeRepoFile(t, dir, "Makefile", "WF_TEST_WEIGHT ?= 1\nbuild:\n\t@true\n")
	testfix.Git(t, dir, "commit", "-qam", "branch drops the suite")

	has, weight, err := makeTestFacts(git, "main")
	if err != nil || !has || weight != 2 {
		t.Fatalf("main's facts: %v %d %v", has, weight, err)
	}
	if has, _, _ := makeTestFacts(git, "feature"); has {
		t.Fatal("the branch's own Makefile reported a test target it deleted")
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the probe made test a goal (it would take the machine lock)")
	}
	if out, _ := git("worktree", "list"); strings.Count(out, "\n") != 1 {
		t.Fatalf("probe worktree left behind:\n%s", out)
	}
}

// Live: this repository's Makefile declares make test at weight 2 (#346).
func TestProbeMakeTestFactsAriadne(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "Makefile.local")); err != nil {
		t.Skip("not in the ariadne checkout")
	}
	has, weight, err := probeMakeTestFacts(root)
	if err != nil || !has || weight != 2 {
		t.Fatalf("ariadne: %v %d %v", has, weight, err)
	}
}

// #332: a declared weight reaches the suite lock as --slots.
func TestAcquireSuiteSlotWeight(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "acquired")
	argv := filepath.Join(t.TempDir(), "argv")
	writeRepoFile(t, d.Root, "scripts/test-lock.py", fakeSuiteLockHelper("open("+pyQuote(argv)+", 'w').write(' '.join(sys.argv[1:]))\nprint('acquired')\n"))
	if _, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 30, Weight: 2}, sha); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(argv)
	if !strings.Contains(string(got), "--slots 2") {
		t.Fatalf("lock argv %q", got)
	}
}

func pyQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `\'`) + "'" }

func TestDecideCIGateSuiteAncestor(t *testing.T) {
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "suite", Command: []string{"make", "test"}, Required: true}}}
	merge := prCheck{Name: "merge-check", Bucket: "pass"}
	if v := decideCIGate(ciInputs{Checks: []prCheck{merge}, Local: decl}); v.Reason != ciLocalNeeded {
		t.Fatalf("unrecorded suite: %d", v.Reason)
	}
	if v := decideCIGate(ciInputs{Checks: []prCheck{merge}, Local: decl, RecordedAncestors: map[string]bool{"suite": true}}); v.Reason != ciGreen {
		t.Fatalf("suite recorded on a docs-only ancestor: %d", v.Reason)
	}
}

// #332 with the hub's (b): the suite run before close stays valid through
// close's own docs-only commits; a code commit after it needs a new run.
func TestRunCIGate_SuiteRecordedBeforeClose(t *testing.T) {
	ctx := context.Background()
	dir := testfix.Repo(t, testfix.InitialCommit())
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	commit := func(path, body string) string {
		writeRepoFile(t, dir, path, body)
		testfix.Git(t, dir, "add", "-A")
		testfix.Git(t, dir, "commit", "-qm", path)
		out, _ := git("rev-parse", "HEAD")
		return out
	}
	writeRepoFile(t, dir, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["true"],"required":true}]}`)
	base := commit("README.md", "x\n")
	testfix.Git(t, dir, "branch", "base-ref")
	tested := commit("pkg/code.go", "package pkg\n")
	closed := commit("workshop/issues/000001-x.md", "---\nid: 000001\n---\n")
	_ = base

	g := newCIFakeGH()
	g.setHead(3, closed, prCheck{Name: "merge-check", Bucket: "pass"})
	g.known[tested] = true
	_ = g.PostStatus(ctx, "o/r", tested, "sdlc-local/suite", "success", "d")
	ran := 0
	runner := func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
		ran++
		_ = g.PostStatus(ctx, "o/r", sha, localStatusContext(c.Name), "success", "d")
		return localRunRecord{Name: c.Name, SHA: sha}, nil
	}
	if err := runCIGate(ctx, g, "o/r", 3, "base-ref", closed, git, runner, false, io.Discard); err != nil || ran != 0 {
		t.Fatalf("docs-only commits after the suite: err %v, reran %d", err, ran)
	}

	code := commit("pkg/more.go", "package pkg\n")
	g.setHead(3, code, prCheck{Name: "merge-check", Bucket: "pass"})
	var out bytes.Buffer
	if err := runCIGate(ctx, g, "o/r", 3, "base-ref", code, git, runner, false, &out); err != nil || ran != 1 {
		t.Fatalf("a code commit after the suite: err %v, ran %d\n%s", err, ran, out.String())
	}
}

// #332 (hub b): with no declaration, `sdlc local-check run --name suite` runs
// main's make test and records sdlc-local/suite: THE close suite.
func TestLocalCheckCmd_DefaultSuite(t *testing.T) {
	dir := tempRepoWith(t, map[string]string{
		"Makefile":             "test:\n\t@echo suite-ran\n",
		"Makefile.workflow":    "# fake\n",
		"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
	})
	ci := newCIFakeGH()
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	var stderr bytes.Buffer
	if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{Names: []string{"suite"}, Rev: "HEAD"}); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	if ci.statuses[head]["sdlc-local/suite"] != "success" || !strings.Contains(stderr.String(), "suite-ran") || !strings.Contains(stderr.String(), "make test") {
		t.Fatalf("statuses %v\n%s", ci.statuses, stderr.String())
	}
}

func TestParseCommitStatuses(t *testing.T) {
	// GitHub lists statuses newest first: the first state per context is current.
	got, err := parseCommitStatuses([]byte(`[
		{"context":"sdlc-local/suite","state":"success"},
		{"context":"ci/x","state":"pending"},
		{"context":"sdlc-local/suite","state":"failure"},
		{"context":"","state":"success"}]`))
	if err != nil || len(got) != 2 || got["sdlc-local/suite"] != "success" || got["ci/x"] != "pending" {
		t.Fatalf("%v %v", got, err)
	}
	if _, err := parseCommitStatuses([]byte(`{}`)); err == nil {
		t.Fatal("a non-list body decoded")
	}
}

// The ancestor walk's limits (M1 BR-2): bounded by the merge base, stops at the
// first code difference (including code renamed to docs, BR-1), stops once
// every name is found, caps its length, and treats a root commit as nothing to
// inherit.
func TestLocalRecordsOnAncestors(t *testing.T) {
	ctx := context.Background()
	dir := testfix.Repo(t, testfix.InitialCommit())
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	commit := func(path, body string) string {
		writeRepoFile(t, dir, path, body)
		testfix.Git(t, dir, "add", "-A")
		testfix.Git(t, dir, "commit", "-qm", path)
		out, _ := git("rev-parse", "HEAD")
		return out
	}
	g := newCIFakeGH()
	record := func(sha string) {
		g.known[sha] = true
		_ = g.PostStatus(ctx, "o/r", sha, "sdlc-local/suite", "success", "d")
	}
	calls := 0
	counting := &countingStatusGH{ciFakeGH: g, calls: &calls}

	// A record before the merge base never counts: it is main's, not the branch's.
	beforeBase := commit("pkg/a.go", "package pkg\n")
	record(beforeBase)
	testfix.Git(t, dir, "branch", "base-ref")
	head := commit("workshop/notes.md", "x\n")
	if found, err := localRecordsOnAncestors(ctx, counting, "o/r", git, head, "base-ref", []string{"suite"}); err != nil || found["suite"] {
		t.Fatalf("record before the merge base counted: %v %v", found, err)
	}

	// Docs-only distance: counted, and the walk stops at the first find.
	tested := commit("pkg/b.go", "package pkg\n")
	record(tested)
	commit("workshop/one.md", "1\n")
	head = commit("workshop/two.md", "2\n")
	calls = 0
	found, err := localRecordsOnAncestors(ctx, counting, "o/r", git, head, "base-ref", []string{"suite"})
	if err != nil || !found["suite"] || calls != 2 {
		t.Fatalf("docs-only distance: %v %v, %d status calls (want 2: one.md's parent… then tested)", found, err, calls)
	}

	// BR-1: code renamed to a docs name is a code difference.
	testfix.Git(t, dir, "mv", "pkg/b.go", "pkg/b.md")
	testfix.Git(t, dir, "commit", "-qm", "rename code to docs")
	renamed, _ := git("rev-parse", "HEAD")
	if found, err := localRecordsOnAncestors(ctx, counting, "o/r", git, renamed, "base-ref", []string{"suite"}); err != nil || found["suite"] {
		t.Fatalf("a rename of code to .md kept the record: %v %v", found, err)
	}

	// The cap: more docs-only commits than the limit hide an older record.
	for i := 0; i < localAncestorLimit+2; i++ {
		commit("workshop/many.md", strings.Repeat("x", i+1)+"\n")
	}
	far, _ := git("rev-parse", "HEAD")
	record(renamed) // the rename commit itself, now beyond the cap
	calls = 0
	if found, err := localRecordsOnAncestors(ctx, counting, "o/r", git, far, "base-ref", []string{"suite"}); err != nil || found["suite"] || calls > localAncestorLimit {
		t.Fatalf("cap: %v %v, %d calls", found, err, calls)
	}

	// A root commit has no ancestor.
	root := testfix.Repo(t)
	rgit := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	writeRepoFile(t, root, "a.txt", "a\n")
	testfix.Git(t, root, "add", "-A")
	testfix.Git(t, root, "commit", "-qm", "root")
	only, _ := rgit("rev-parse", "HEAD")
	if found, err := localRecordsOnAncestors(ctx, counting, "o/r", rgit, only, only, []string{"suite"}); err != nil || found["suite"] {
		t.Fatalf("root commit: %v %v", found, err)
	}
}

// countingStatusGH counts CommitStatuses calls.
type countingStatusGH struct {
	*ciFakeGH
	calls *int
}

func (g *countingStatusGH) CommitStatuses(ctx context.Context, repo, sha string) (map[string]string, error) {
	*g.calls++
	return g.ciFakeGH.CommitStatuses(ctx, repo, sha)
}

func TestParseMakeTestFactsSkipsNotATarget(t *testing.T) {
	db := "# Not a target:\ntest:\n#  Implicit rule search has not been done.\n__SDLC_SUITE_WEIGHT=1\n"
	if has, _, err := parseMakeTestFacts(db); err != nil || has {
		t.Fatalf("a prerequisite-only test entry read as a rule: %v %v", has, err)
	}
}
