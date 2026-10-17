package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Each PR retains the existing real-Git landing fake. Only the GitHub seam is
// multiplexed: both invocations see the same server and thread-safe CI state.
type concurrentLandingGH struct {
	*ciFakeGH
	mu     sync.Mutex
	prs    map[string]*landingFakeGH
	suites map[int][]string
}

func (g *concurrentLandingGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.prs[branch].LandingPRs(ctx, repo, branch)
}

func (g *concurrentLandingGH) LandingMerge(ctx context.Context, repo string, number int, head string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, pr := range g.prs {
		if pr.pr.Number == number {
			return pr.LandingMerge(ctx, repo, number, head)
		}
	}
	return fmt.Errorf("unknown PR #%d", number)
}

func (g *concurrentLandingGH) PostStatus(ctx context.Context, repo, sha, name, state, desc string) error {
	if err := g.ciFakeGH.PostStatus(ctx, repo, sha, name, state, desc); err != nil {
		return err
	}
	if name == localStatusContext(suiteCheckName) {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.ciFakeGH.mu.Lock()
		defer g.ciFakeGH.mu.Unlock()
		for number, head := range g.ciFakeGH.heads {
			if head == sha {
				g.suites[number] = append(g.suites[number], sha)
			}
		}
	}
	return nil
}

// The real suite subprocess announces its exact checkout head, then waits for
// release. The readiness marker, not elapsed time, establishes the ordering.
func blockLandingSuite(t *testing.T, server string) (started, release string) {
	t.Helper()
	dir := t.TempDir()
	started, release = filepath.Join(dir, "started"), filepath.Join(dir, "release")
	command := fmt.Sprintf("git rev-parse HEAD > %s; i=0; while [ ! -e %s ] && [ \"$i\" -lt 2400 ]; do sleep 0.05; i=$((i+1)); done; test -e %s", quoteHelperShell(started), quoteHelperShell(release), quoteHelperShell(release))
	procedureWrite(t, server, localChecksPath, `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c",`+strconv.Quote(command)+`]}]}`)
	git(t, server, "add", localChecksPath)
	git(t, server, "commit", "-qm", "suite readiness protocol")
	git(t, server, "push", "-q", "origin", "main")
	return
}

func landingSuiteReady(t *testing.T, path string, worker *blockedMergeWorker) string {
	t.Helper()
	deadline := time.After(30 * time.Second)
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if b, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(b))) == 40 {
			return strings.TrimSpace(string(b))
		}
		select {
		case <-tick.C:
		case <-worker.done:
			t.Fatalf("landing ended before announcing its suite: %v", worker.err)
		case <-deadline:
			t.Fatal("suite did not announce its checkout head")
		}
	}
}

func finishLandingWorker(t *testing.T, w *blockedMergeWorker) {
	t.Helper()
	select {
	case <-w.done:
		if w.err != nil {
			t.Fatal(w.err)
		}
	case <-time.After(time.Minute):
		t.Fatal("landing did not finish after the suite was released")
	}
}

// These are two real durable landing invocations in separate slot worktrees,
// not an artificial ticket standing in for the first landing. The existing
// workspace seam supplies each invocation's cwd without changing process cwd.
func TestLandingLineTwoLandingsIntegrateInTheirOwnTurns(t *testing.T) {
	first, server, gh1 := integrationFixture(t)
	started, release := blockLandingSuite(t, server)
	initialMain := procedureHead(t, server)
	second := filepath.Join(filepath.Dir(filepath.Dir(first)), "sample-slot2", "sample")
	const branch2 = "000002-second"
	git(t, second, "switch", "--no-track", "-c", branch2)
	procedureWrite(t, second, "workshop/issues/000002-second.md", "---\nid: 000002\nstatus: codecomplete\n---\n# Second\n\n## Log\n\nClosed and reviewed.\n")
	procedureWrite(t, second, "second.go", "package second\n")
	git(t, second, "add", ".")
	git(t, second, "commit", "-qm", "close second issue")
	git(t, second, "push", "-q", "upstream", branch2)
	secondBefore := procedureHead(t, second)
	gh2 := &landingFakeGH{t: t, remote: gh1.remote, followPush: true, ciSeamGH: ciSeamGH{ci: gh1.ci}, pr: landingPR{
		Number: 2, State: "OPEN", Repo: gh1.pr.Repo, HeadRef: branch2, HeadOID: secondBefore, BaseRef: "main", BaseOID: gh1.pr.BaseOID,
	}}
	gh1.ci.setHead(2, secondBefore)
	gh := &concurrentLandingGH{ciFakeGH: gh1.ci, prs: map[string]*landingFakeGH{landingTestBranch: gh1, branch2: gh2}, suites: map[int][]string{}}
	ghClient = gh

	// Separate CLI processes have separate cwd values. Supply that one input
	// through the existing seam, retaining real workspace identity resolution.
	type invocationRootKey struct{}
	oldResolve := resolveLandingWorkspace
	resolveLandingWorkspace = func(r workspace.GitReader, dir, address string) (workspace.Identity, error) {
		if bounded, ok := r.(boundedLandingGitRunner); ok && dir == "." {
			if root, ok := bounded.ctx.Value(invocationRootKey{}).(string); ok {
				dir = root
			}
		}
		return oldResolve(r, dir, address)
	}
	t.Cleanup(func() { resolveLandingWorkspace = oldResolve })
	dir, err := landingLineDir(mergeRunner, first)
	if err != nil {
		t.Fatal(err)
	}
	line, err := landingline.Open(dir, landingline.Options{Poll: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	oldPoll := landingLinePoll
	landingLinePoll = time.Millisecond
	t.Cleanup(func() { landingLinePoll = oldPoll })
	start := func(root string) *blockedMergeWorker {
		return startBlockedMergeWorker(t, release, func(ctx context.Context) error {
			f := landingFlags()
			f.Context = context.WithValue(ctx, invocationRootKey{}, root)
			return runMerge(io.Discard, io.Discard, f)
		})
	}
	one := start(first)
	testedFirst := landingSuiteReady(t, started, one)
	two := start(second)
	waitQueued(t, line, 1)

	// The first suite is still executing. The second branch and server head
	// are untouched, and the durable log contains no second admission.
	gh.mu.Lock()
	merges, suites := gh1.merges+gh2.merges, len(gh.suites[1])+len(gh.suites[2])
	gh.mu.Unlock()
	if merges != 0 || suites != 0 || procedureHead(t, second) != secondBefore || git(t, gh1.remote, "rev-parse", "refs/heads/"+branch2) != secondBefore {
		t.Fatalf("second landing acted before admission: merges=%d completed suites=%d", merges, suites)
	}
	if b, err := os.ReadFile(started); err != nil || strings.TrimSpace(string(b)) != testedFirst {
		t.Fatalf("another suite started before admission: marker=%q err=%v", b, err)
	}
	for _, event := range lineEvents(t, line)[landingline.EventAdmit] {
		if event.Branch == branch2 {
			t.Fatal("second landing admitted while the first suite was active")
		}
	}
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	finishLandingWorker(t, one)
	finishLandingWorker(t, two)
	if gh1.merges != 1 || gh2.merges != 1 || len(gh.suites[1]) != 1 || len(gh.suites[2]) != 1 || gh.suites[1][0] != testedFirst {
		t.Fatalf("merges %d/%d; suite heads %v", gh1.merges, gh2.merges, gh.suites)
	}
	for _, tc := range []struct{ root, ancestor, head string }{
		{first, initialMain, gh.suites[1][0]},
		{second, gh1.pr.MergeOID, gh.suites[2][0]},
	} {
		if ok, err := integrationAncestor(mergeRunner, tc.root, tc.ancestor, tc.head); err != nil || !ok {
			t.Fatalf("suite %s omitted the main it must contain (%s): %v", tc.head, tc.ancestor, err)
		}
	}
	if gh1.pr.HeadOID != gh.suites[1][0] || gh2.pr.HeadOID != gh.suites[2][0] {
		t.Fatal("landing merged a head other than its own tested integration")
	}
	events := lineEvents(t, line)
	if len(events[landingline.EventAdmit]) != 2 || len(events[landingline.EventLanded]) != 2 {
		t.Fatalf("expected two completed turns: %v", events)
	}
	if events[landingline.EventAdmit][0].Branch != landingTestBranch || events[landingline.EventAdmit][1].Branch != branch2 {
		t.Fatalf("wrong admission order: %v", events)
	}
}

func TestLandingReusesSuiteWhenProsePublishedDuringRun(t *testing.T) {
	root, server, gh := integrationFixture(t)
	started, release := blockLandingSuite(t, server)
	var tested []string
	gh.afterStatus = func(sha string) { tested = append(tested, sha) }
	worker := startBlockedMergeWorker(t, release, func(ctx context.Context) error {
		f := landingFlags()
		f.Context = ctx
		return runMerge(io.Discard, io.Discard, f)
	})
	testedHead := landingSuiteReady(t, started, worker)
	// Use the same publication/CAS transaction as trunk publish, against the
	// shared bare origin, while the suite subprocess is waiting for release.
	tf, err := gitx.NewTrunkFileContext(context.Background(), server, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	beforePublication := git(t, gh.remote, "rev-parse", "refs/heads/main")
	const prose = "workshop/parley/during-suite.md"
	if err := tf.UpdateMany("trunk: publish "+prose, func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{prose: []byte("published during the landing suite\n")}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	published := git(t, gh.remote, "rev-parse", "refs/heads/main")
	if published == beforePublication || git(t, server, "rev-parse", published+"^") != beforePublication {
		t.Fatal("prose publication did not advance main by one commit during the suite")
	}
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	finishLandingWorker(t, worker)
	if gh.merges != 1 || len(tested) != 1 || tested[0] != testedHead || gh.pr.HeadOID == testedHead {
		t.Fatalf("expected one suite followed by docs integration: merges=%d suites=%v merged=%s", gh.merges, tested, gh.pr.HeadOID)
	}
	if ok, err := integrationAncestor(mergeRunner, root, published, gh.pr.HeadOID); err != nil || !ok {
		t.Fatalf("landed without the prose publication: %v %v", ok, err)
	}
	paths := strings.Fields(git(t, root, "diff", "--name-only", testedHead, gh.pr.HeadOID))
	if len(paths) != 1 || paths[0] != prose {
		t.Fatalf("reused suite across an unexpected cumulative delta: %v", paths)
	}
	gitCall := func(args ...string) (string, error) { return landingGit(mergeRunner, root, args...) }
	found, err := localRecordsOnAncestors(context.Background(), gh, gh.pr.Repo, gitCall, gh.pr.HeadOID, published, []string{suiteCheckName})
	if err != nil || !found[suiteCheckName] {
		t.Fatalf("the authoritative integrated head did not inherit its suite record: %v %v", found, err)
	}
}

// Ancestor reuse checks the cumulative tree delta, not the latest commit's
// prose suffix: code, gate configuration and embedded help still invalidate it.
func TestLocalRecordsOnAncestorsRejectsNonProseDrift(t *testing.T) {
	for _, path := range []string{"pkg/feature.go", localChecksPath, "cmd/sdlc/helptext/merge.md"} {
		t.Run(path, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			base := procedureHead(t, root)
			commit := func(path, body string) string {
				procedureWrite(t, root, path, body)
				git(t, root, "add", ".")
				git(t, root, "commit", "-qm", path)
				return procedureHead(t, root)
			}
			tested := commit("feature.txt", "reviewed feature\n")
			gh := newCIFakeGH()
			gh.setHead(1, tested)
			if err := gh.PostStatus(context.Background(), "test/repo", tested, localStatusContext(suiteCheckName), "success", "passed"); err != nil {
				t.Fatal(err)
			}
			commit(path, "changed\n")
			head := commit("workshop/parley/after-drift.md", "prose after a code/config/help change\n")
			gitCall := func(args ...string) (string, error) { return landingGit(execGitRunner{}, root, args...) }
			found, err := localRecordsOnAncestors(context.Background(), gh, "test/repo", gitCall, head, base, []string{suiteCheckName})
			if err != nil || found[suiteCheckName] {
				t.Fatalf("suite reused across cumulative %s drift: %v %v", path, found, err)
			}
		})
	}
}
