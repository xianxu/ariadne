package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// ciFakeGH is the stateful fake behind the CI gate's GitHub seam (#345): PR
// heads, the CI checks on each head, commit statuses, Actions jobs, and PR
// comments. Its PRChecks folds the statuses on the PR's current head into the
// list, latest state per context winning, as GitHub's rollup does — so a status
// the gate's inline local run posts is visible to the gate's next read.
type ciFakeGH struct {
	stubGH
	mu       sync.Mutex
	heads    map[int]string
	checks   map[string][]prCheck         // sha → CI checks
	statuses map[string]map[string]string // sha → context → state
	jobs     map[int64]jobRun
	comments map[int][]string
	known    map[string]bool // shas GitHub knows (pushed); a status on any other is a 422
	jobErr   error
}

func newCIFakeGH() *ciFakeGH {
	return &ciFakeGH{heads: map[int]string{}, checks: map[string][]prCheck{}, statuses: map[string]map[string]string{},
		jobs: map[int64]jobRun{}, comments: map[int][]string{}, known: map[string]bool{}}
}

// setHead points PR number at sha with the given CI checks.
func (g *ciFakeGH) setHead(number int, sha string, checks ...prCheck) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.heads[number] = sha
	g.known[sha] = true
	g.checks[sha] = checks
}

func (g *ciFakeGH) PRHeadOID(_ context.Context, _ string, number int) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if h, ok := g.heads[number]; ok {
		return h, nil
	}
	return "", fmt.Errorf("ciFakeGH: no PR #%d", number)
}

func (g *ciFakeGH) PRChecks(_ context.Context, _ string, number int) ([]prCheck, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	head := g.heads[number]
	out := append([]prCheck(nil), g.checks[head]...)
	var contexts []string
	for c := range g.statuses[head] {
		contexts = append(contexts, c)
	}
	sort.Strings(contexts)
	for _, c := range contexts {
		bucket := "fail"
		if g.statuses[head][c] == "success" {
			bucket = "pass"
		}
		out = append(out, prCheck{Name: c, Bucket: bucket})
	}
	return out, nil
}

func (g *ciFakeGH) JobRun(_ context.Context, _ string, id int64) (jobRun, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.jobErr != nil {
		return jobRun{}, g.jobErr
	}
	j, ok := g.jobs[id]
	if !ok {
		return jobRun{}, fmt.Errorf("ciFakeGH: HTTP 404 job %d", id)
	}
	return j, nil
}

func (g *ciFakeGH) PostStatus(_ context.Context, _, sha, statusContext, state, _ string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.known[sha] {
		return fmt.Errorf("ciFakeGH: HTTP 422 No commit found for SHA: %s", sha)
	}
	if g.statuses[sha] == nil {
		g.statuses[sha] = map[string]string{}
	}
	g.statuses[sha][statusContext] = state
	return nil
}

func (g *ciFakeGH) CommitStatuses(_ context.Context, _ string, sha string) (map[string]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := map[string]string{}
	for c, st := range g.statuses[sha] {
		out[c] = st
	}
	return out, nil
}

func (g *ciFakeGH) PRComment(_ context.Context, _ string, number int, body string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.comments[number] = append(g.comments[number], body)
	return nil
}

// ciSeamGH is what a flow fake embeds instead of stubGH: the #345 methods go to
// ci when a test sets it, and to stubGH's defaults when not.
type ciSeamGH struct {
	stubGH
	ci *ciFakeGH
}

func (g ciSeamGH) JobRun(ctx context.Context, repo string, id int64) (jobRun, error) {
	if g.ci == nil {
		return g.stubGH.JobRun(ctx, repo, id)
	}
	return g.ci.JobRun(ctx, repo, id)
}

func (g ciSeamGH) PostStatus(ctx context.Context, repo, sha, statusContext, state, desc string) error {
	if g.ci == nil {
		return g.stubGH.PostStatus(ctx, repo, sha, statusContext, state, desc)
	}
	return g.ci.PostStatus(ctx, repo, sha, statusContext, state, desc)
}

func (g ciSeamGH) CommitStatuses(ctx context.Context, repo, sha string) (map[string]string, error) {
	if g.ci == nil {
		return g.stubGH.CommitStatuses(ctx, repo, sha)
	}
	return g.ci.CommitStatuses(ctx, repo, sha)
}

func (g ciSeamGH) PRComment(ctx context.Context, repo string, n int, body string) error {
	if g.ci == nil {
		return g.stubGH.PRComment(ctx, repo, n, body)
	}
	return g.ci.PRComment(ctx, repo, n, body)
}

// fakeJob builds the jobs-API answer: runnerless, or a job that ran n steps.
func fakeJob(conclusion, runner string, steps int) jobRun {
	c := conclusion
	return jobRun{Status: "completed", Conclusion: &c, RunnerName: &runner, Steps: make([]json.RawMessage, steps)}
}

func TestCIFakeGHRollup(t *testing.T) {
	ctx := context.Background()
	g := newCIFakeGH()
	g.setHead(3, "aaa", prCheck{Name: "merge-check", Bucket: "pass"})
	if err := g.PostStatus(ctx, "o/r", "aaa", "sdlc-local/x", "success", "d"); err != nil {
		t.Fatal(err)
	}
	checks, _ := g.PRChecks(ctx, "o/r", 3)
	if len(checks) != 2 || checks[1] != (prCheck{Name: "sdlc-local/x", Bucket: "pass"}) {
		t.Fatalf("%+v", checks)
	}
	_ = g.PostStatus(ctx, "o/r", "aaa", "sdlc-local/x", "failure", "d")
	if checks, _ = g.PRChecks(ctx, "o/r", 3); len(checks) != 2 || checks[1].Bucket != "fail" {
		t.Fatalf("latest state does not win: %+v", checks)
	}
	if err := g.PostStatus(ctx, "o/r", "unpushed", "sdlc-local/x", "success", "d"); err == nil {
		t.Fatal("status on an unknown sha accepted")
	}
	g.setHead(3, "bbb")
	if checks, _ = g.PRChecks(ctx, "o/r", 3); len(checks) != 0 {
		t.Fatalf("statuses leaked to a new head: %+v", checks)
	}
}

// TestLiveJobRunConformance checks the fake's model of the jobs API against
// GitHub: pair run 38453641991 attempt 1 never got a macOS runner, attempt 2 ran.
// Set SDLC_LIVE_GH=1 (needs gh auth with read access to xianxu/pair).
func TestLiveJobRunConformance(t *testing.T) {
	if os.Getenv("SDLC_LIVE_GH") != "1" {
		t.Skip("set SDLC_LIVE_GH=1 to read the live jobs API")
	}
	ctx := context.Background()
	for id, want := range map[int64]bool{115490631095: true, 115497743342: false} {
		j, err := realGH{}.JobRun(ctx, "xianxu/pair", id)
		if err != nil {
			t.Fatal(err)
		}
		if j.runnerless() != want {
			t.Fatalf("job %d: runnerless %v, want %v (%+v)", id, j.runnerless(), want, j)
		}
	}
}

// TestRunCIGateLocalChecks drives the gate against ciFakeGH: a runnerless job is
// satisfied by a recorded or inline local run; anything that ran stays red.
func TestRunCIGateLocalChecks(t *testing.T) {
	ctx := context.Background()
	const label = "couch-zellij-conformance / conformance"
	const link = "https://github.com/o/r/actions/runs/1/job/77"
	decl := `{"version":1,"checks":[{"name":"zellij","command":["make","live"],"substitutes":["` + label + `"]}]}`
	// tree serves the declaration on rev declRev only, and workflows on the head.
	tree := func(declRev, body string) func(args ...string) (string, error) {
		return func(args ...string) (string, error) {
			switch {
			case args[0] == "ls-tree" && args[len(args)-1] == localChecksPath:
				if args[len(args)-2] == declRev {
					return localChecksPath + "\n", nil
				}
				return "", nil
			case args[0] == "show":
				return body, nil
			case args[len(args)-1] == ".github/workflows":
				return ".github/workflows\n", nil
			}
			return "", nil
		}
	}
	job := prCheck{Name: "conformance", Workflow: "couch-zellij-conformance", Bucket: "cancel", Link: link}
	merge := prCheck{Name: "merge-check", Bucket: "pass"}
	setup := func(steps int, runner string) *ciFakeGH {
		g := newCIFakeGH()
		g.setHead(3, "h", merge, job)
		g.jobs[77] = fakeJob("cancelled", runner, steps)
		return g
	}
	// runner records outcome on the fake, as runLocalCheck would.
	runner := func(g *ciFakeGH, calls *[]string, state string, before func()) localCheckRunner {
		return func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			*calls = append(*calls, c.Name+"@"+sha)
			if before != nil {
				before()
			}
			rec := localRunRecord{Name: c.Name, SHA: sha, Exit: 0}
			if state == "failure" {
				rec.Exit = 2
			}
			if state != "" {
				_ = g.PostStatus(ctx, "o/r", sha, localStatusContext(c.Name), state, "d")
			}
			return rec, nil
		}
	}

	t.Run("recorded pass is accepted and announced", func(t *testing.T) {
		g := setup(0, "")
		_ = g.PostStatus(ctx, "o/r", "h", "sdlc-local/zellij", "success", "d")
		var out bytes.Buffer
		var calls []string
		if err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "success", nil), false, &out); err != nil {
			t.Fatal(err)
		}
		if len(calls) != 0 || !strings.Contains(out.String(), "never got a runner") || !strings.Contains(out.String(), label) {
			t.Fatalf("calls %v out %q", calls, out.String())
		}
	})
	t.Run("nothing recorded: runs inline on the pinned head", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		if err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "success", nil), false, io.Discard); err != nil {
			t.Fatal(err)
		}
		if strings.Join(calls, ",") != "zellij@h" {
			t.Fatalf("calls %v", calls)
		}
	})
	t.Run("no runner available names the run", func(t *testing.T) {
		g := setup(0, "")
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), nil, false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "sdlc local-check run --name zellij") {
			t.Fatalf("%v", err)
		}
	})
	t.Run("inline failure refuses", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "failure", nil), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "local check zellij failed") {
			t.Fatalf("%v", err)
		}
	})
	t.Run("head moves during the run", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		move := func() { g.setHead(3, "h2", merge) }
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "success", move), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "another commit") {
			t.Fatalf("%v", err)
		}
	})
	t.Run("a pass that was not recorded refuses", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "", nil), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "needs recorded local checks") || len(calls) != 1 {
			t.Fatalf("calls %v err %v", calls, err)
		}
	})
	t.Run("a job that ran and was cancelled stays red", func(t *testing.T) {
		g := setup(4, "GitHub Actions 7")
		var calls []string
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), runner(g, &calls, "success", nil), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "failing checks") || len(calls) != 0 {
			t.Fatalf("calls %v err %v", calls, err)
		}
	})
	t.Run("a declaration only on the head does not count", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("h", decl), runner(g, &calls, "success", nil), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "failing checks") || len(calls) != 0 {
			t.Fatalf("calls %v err %v", calls, err)
		}
	})
	t.Run("an invalid declaration refuses", func(t *testing.T) {
		g := setup(0, "")
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", `{"version":1,"checks":[{"name":"zellij","cmd":["x"]}]}`), nil, false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), localChecksPath) {
			t.Fatalf("%v", err)
		}
	})
	t.Run("a missing required verdict refuses before any local run", func(t *testing.T) {
		g := setup(0, "")
		var calls []string
		guarded := func(args ...string) (string, error) {
			if args[0] == "ls-tree" && args[len(args)-1] == localGatesWorkflow {
				return localGatesWorkflow + "\n", nil
			}
			return tree("b", decl)(args...)
		}
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", guarded, runner(g, &calls, "success", nil), false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), localGatesVerdict) || len(calls) != 0 {
			t.Fatalf("calls %v err %v", calls, err)
		}
	})
	t.Run("an unreadable job refuses as unread", func(t *testing.T) {
		g := setup(0, "")
		g.jobErr = errors.New("HTTP 502")
		err := runCIGate(ctx, g, "o/r", 3, "b", "h", tree("b", decl), nil, false, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "could not read") {
			t.Fatalf("%v", err)
		}
	})
}

// localChecksMergeRepo is the merge e2e repo with a declaration on main whose
// check runs command, a fake suite lock, and GitHub showing the conformance job
// on the pushed head as cancelled after steps steps on runner.
func localChecksMergeRepo(t *testing.T, command string, steps int, runner string) (*e2eGH, string) {
	t.Helper()
	decl := `{"version":1,"checks":[{"name":"zellij","command":["sh","-c",` + fmt.Sprintf("%q", command) + `],"substitutes":["couch-zellij-conformance / conformance"]}]}`
	tempRepoWith(t, map[string]string{
		localChecksPath:        decl,
		"Makefile.workflow":    "# fake\n",
		"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
	})
	head, err := exec.Command("git", "rev-parse", "origin/feature").Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(string(head))
	ci := newCIFakeGH()
	ci.setHead(42, sha, prCheck{Name: "merge-check", Bucket: "pass"},
		prCheck{Name: "conformance", Workflow: "couch-zellij-conformance", Bucket: "cancel", Link: "https://github.com/test/repo/actions/runs/1/job/77"})
	ci.jobs[77] = fakeJob("cancelled", runner, steps)
	gh := &e2eGH{openPR: "42", ciSeamGH: ciSeamGH{ci: ci}}
	swapMergeDeps(t, gh, nil)
	return gh, sha
}

// Done-when (#345): a merge whose macOS job never got a runner proceeds after a
// passing local run of the declared check, recorded on the pinned head.
func TestRunMerge_RunnerlessJobMergesAfterLocalRun(t *testing.T) {
	gh, sha := localChecksMergeRepo(t, "echo live ok", 0, "")
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	var stderr bytes.Buffer
	if err := runMerge(io.Discard, &stderr, f); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	if gh.prMergeCalls != 1 || gh.mergedHead != sha {
		t.Fatalf("merges %d head %q", gh.prMergeCalls, gh.mergedHead)
	}
	if gh.ci.statuses[sha]["sdlc-local/zellij"] != "success" || !strings.Contains(stderr.String(), "sh -c echo live ok") {
		t.Fatalf("statuses %v\n%s", gh.ci.statuses, stderr.String())
	}
}

// Done-when (#345): a job that ran and failed blocks, and no local run replaces it.
func TestRunMerge_FailedJobStillBlocks(t *testing.T) {
	gh, sha := localChecksMergeRepo(t, "echo live ok", 6, "GitHub Actions 3")
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "failing checks") || !strings.Contains(msg, "conformance") {
		t.Fatalf("died=%v msg=%q", died, msg)
	}
	if gh.prMergeCalls != 0 || len(gh.ci.statuses[sha]) != 0 {
		t.Fatalf("merges %d statuses %v", gh.prMergeCalls, gh.ci.statuses)
	}
}

// A runnerless job whose local substitute fails blocks too.
func TestRunMerge_RunnerlessJobLocalFailureBlocks(t *testing.T) {
	gh, sha := localChecksMergeRepo(t, "echo broken; exit 1", 0, "")
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "local check zellij failed") {
		t.Fatalf("died=%v msg=%q", died, msg)
	}
	if gh.prMergeCalls != 0 || gh.ci.statuses[sha]["sdlc-local/zellij"] != "failure" || len(gh.ci.comments[42]) != 1 {
		t.Fatalf("merges %d statuses %v comments %d", gh.prMergeCalls, gh.ci.statuses, len(gh.ci.comments[42]))
	}
}

// The durable landing path (#345): the declaration on main, the check run in
// the slot's checkout, and the merge pinned to the head the run recorded on.
func TestLandingRunnerlessJobMergesAfterLocalRun(t *testing.T) {
	roots, remote, gh := landingFixture(t, 1)
	root := roots[1]
	// Land the declaration on the server's main, then bring it into the branch.
	seed := filepath.Join(t.TempDir(), "seed")
	git(t, "", "clone", "--quiet", remote, seed)
	git(t, seed, "config", "user.name", "seed")
	git(t, seed, "config", "user.email", "seed@example.com")
	for rel, body := range map[string]string{
		localChecksPath:        `{"version":1,"checks":[{"name":"zellij","command":["sh","-c","echo live ok"],"substitutes":["couch-zellij-conformance / conformance"]}]}`,
		"Makefile.workflow":    "# fake\n",
		"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
	} {
		procedureWrite(t, seed, rel, body)
	}
	git(t, seed, "add", ".")
	git(t, seed, "commit", "-qm", "declare local checks")
	git(t, seed, "push", "--quiet", "origin", "main")
	git(t, root, "fetch", "--quiet", "upstream")
	git(t, root, "merge", "--quiet", "--no-edit", "upstream/main")
	git(t, root, "push", "--quiet", "upstream", landingTestBranch)
	head := procedureHead(t, root)
	gh.pr.HeadOID = head
	gh.ci = newCIFakeGH()
	gh.ci.setHead(1, head, prCheck{Name: "merge-check", Bucket: "pass"},
		prCheck{Name: "conformance", Workflow: "couch-zellij-conformance", Bucket: "cancel", Link: "https://github.com/test/repo/actions/runs/1/job/77"})
	gh.ci.jobs[77] = fakeJob("cancelled", "", 0)
	var stderr bytes.Buffer
	if err := runMerge(io.Discard, &stderr, landingFlags()); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	if gh.merges != 1 || gh.ci.statuses[head]["sdlc-local/zellij"] != "success" {
		t.Fatalf("merges %d statuses %v\n%s", gh.merges, gh.ci.statuses, stderr.String())
	}
}
