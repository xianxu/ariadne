package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

func (g *ciFakeGH) PRComment(_ context.Context, _ string, number int, body string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.comments[number] = append(g.comments[number], body)
	return nil
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
