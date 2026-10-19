package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// trainFakeGH is the stateful GitHub fake for trains (#449): several PRs on
// one real bare server, train PRs opened and closed, and a merge that marks
// every open PR whose head became reachable from main merged, as GitHub
// documents. What mergeCommit such a PR reports is undocumented, so it is
// the fake's parameter: the train's merge commit, null, or the car's head.
type trainFakeGH struct {
	*ciFakeGH
	t       *testing.T
	remote  string
	variant string // "train", "null" or "head"
	mu      sync.Mutex
	prs     map[string]*landingPR // by head branch
	next    int
	merges  int
	created []string
	closed  []int
}

func newTrainFakeGH(t *testing.T, remote string, ci *ciFakeGH, variant string) *trainFakeGH {
	return &trainFakeGH{ciFakeGH: ci, t: t, remote: remote, variant: variant, prs: map[string]*landingPR{}, next: 100}
}

// addPR registers an open PR at branch's pushed head.
func (g *trainFakeGH) addPR(number int, branch string) *landingPR {
	g.mu.Lock()
	defer g.mu.Unlock()
	head := git(g.t, g.remote, "rev-parse", "refs/heads/"+branch)
	pr := &landingPR{Number: number, State: "OPEN", Repo: "test/repo", HeadRef: branch, HeadOID: head, BaseRef: "main", BaseOID: git(g.t, g.remote, "rev-parse", "refs/heads/main")}
	g.prs[branch] = pr
	g.ciFakeGH.setHead(number, head)
	return pr
}

func (g *trainFakeGH) LandingPRs(_ context.Context, _ string, branch string) ([]landingPR, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	pr, ok := g.prs[branch]
	if !ok {
		return nil, nil
	}
	if pr.State == "OPEN" { // GitHub follows pushes to an open PR's branch
		if head, err := trainGitOut(g.remote, "rev-parse", "refs/heads/"+branch); err == nil && head != pr.HeadOID {
			pr.HeadOID = head
			g.ciFakeGH.setHead(pr.Number, head)
		}
	}
	return []landingPR{*pr}, nil
}

func (g *trainFakeGH) LandingMerge(_ context.Context, repo string, number int, head string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	var pr *landingPR
	for _, p := range g.prs {
		if p.Number == number {
			pr = p
		}
	}
	if pr == nil || pr.State != "OPEN" || pr.HeadOID != head {
		return errors.New("trainFakeGH: merge of an unknown, closed or moved PR")
	}
	g.merges++
	server := filepath.Join(g.t.TempDir(), "server")
	git(g.t, "", "clone", "--quiet", g.remote, server)
	git(g.t, server, "config", "user.name", "server")
	git(g.t, server, "config", "user.email", "server@example.com")
	git(g.t, server, "merge", "--no-ff", "-m", fmt.Sprintf("Merge pull request #%d", number), head)
	git(g.t, server, "push", "--quiet", "origin", "HEAD:main")
	merge := procedureHead(g.t, server)
	pr.State, pr.MergeOID = "MERGED", merge
	for _, p := range g.prs {
		if p.State != "OPEN" {
			continue
		}
		if _, err := trainGitOut(server, "merge-base", "--is-ancestor", p.HeadOID, merge); err != nil {
			continue
		}
		p.State = "MERGED"
		switch g.variant {
		case "train":
			p.MergeOID = merge
		case "head":
			p.MergeOID = p.HeadOID
		default:
			p.MergeOID = ""
		}
	}
	return nil
}

func (g *trainFakeGH) OpenTrainPRs(_ context.Context, _ string) ([]landingPR, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []landingPR
	for branch, p := range g.prs {
		if strings.HasPrefix(branch, trainBranchPrefix) && p.State == "OPEN" {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (g *trainFakeGH) CreateTrainPR(_ context.Context, _ string, branch, _, _ string) error {
	g.mu.Lock()
	n := g.next
	g.next++
	g.created = append(g.created, branch)
	g.mu.Unlock()
	g.addPR(n, branch)
	return nil
}

func (g *trainFakeGH) CloseTrainPR(_ context.Context, _ string, number int, branch string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if p, ok := g.prs[branch]; ok && p.Number == number {
		p.State = "CLOSED"
	}
	g.closed = append(g.closed, number)
	_, _ = trainGitOut(g.remote, "update-ref", "-d", "refs/heads/"+branch)
	return nil
}

func (g *trainFakeGH) pr(branch string) landingPR {
	g.mu.Lock()
	defer g.mu.Unlock()
	return *g.prs[branch]
}

// suiteRuns lists the commits a suite status was posted on, in order.
func (g *trainFakeGH) suiteHeads() []string {
	g.ciFakeGH.mu.Lock()
	defer g.ciFakeGH.mu.Unlock()
	var out []string
	for sha, st := range g.ciFakeGH.statuses {
		if _, ok := st[localStatusContext(suiteCheckName)]; ok {
			out = append(out, sha)
		}
	}
	sort.Strings(out)
	return out
}

func trainGitOut(dir string, args ...string) (string, error) {
	out, err := execGitRunner{}.GitInDir(dir, args...)
	return strings.TrimSpace(string(out)), err
}

var _ trainGH = (*trainFakeGH)(nil)

// trainCar is one issue branch, closed and pushed, with its branch suite
// recorded on its head (the readiness every car vouches with).
func trainCar(t *testing.T, root, branch, file string, issue int) {
	t.Helper()
	git(t, root, "switch", "--quiet", "--no-track", "-c", branch)
	procedureWrite(t, root, fmt.Sprintf("workshop/issues/%06d-%s.md", issue, file), fmt.Sprintf("---\nid: %06d\nstatus: codecomplete\n---\n# %s\n\n## Log\n\nClosed and reviewed.\n", issue, file))
	procedureWrite(t, root, file+".go", "package "+file+"\n")
	git(t, root, "add", ".")
	git(t, root, "commit", "-qm", "close "+file)
	git(t, root, "push", "-q", "upstream", branch)
}

// TestLandingTrainLandsEveryCarOnOneSuite is #449's done-when: with three
// ready PRs, one train runs one integration suite and lands all three. Two
// cars are real sdlc merge invocations in their own slots; the third is a
// waiting ticket the test vouches for, as another slot's merge would. Every
// mergeCommit variant GitHub might report for a car runs end to end.
func TestLandingTrainLandsEveryCarOnOneSuite(t *testing.T) {
	for _, variant := range []string{"train", "null", "head"} {
		t.Run(variant, func(t *testing.T) { landTrainOfThree(t, variant) })
	}
}

func landTrainOfThree(t *testing.T, variant string) {
	first, server, gh1 := integrationFixture(t)
	procedureWrite(t, server, landingConfigPath, `{"train_interval_minutes": 60}`+"\n")
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "trains on")
	git(t, server, "push", "-q", "origin", "main")
	for _, root := range []string{first} {
		git(t, root, "fetch", "-q", "upstream")
	}
	second := filepath.Join(filepath.Dir(filepath.Dir(first)), "sample-slot2", "sample")
	git(t, second, "fetch", "-q", "upstream")
	git(t, second, "reset", "-q", "--hard", "upstream/main")
	trainCar(t, second, "000002-second", "second", 2)
	third := filepath.Join(t.TempDir(), "third")
	git(t, "", "clone", "-q", gh1.remote, third)
	git(t, third, "config", "user.name", "third")
	git(t, third, "config", "user.email", "third@example.com")
	git(t, third, "remote", "rename", "origin", "upstream")
	trainCar(t, third, "000003-third", "third", 3)

	gh := newTrainFakeGH(t, gh1.remote, gh1.ci, variant)
	gh.addPR(1, landingTestBranch)
	gh.addPR(2, "000002-second")
	gh.addPR(3, "000003-third")
	// The branch suites, recorded on each car's head before it lands.
	for _, b := range []string{landingTestBranch, "000002-second", "000003-third"} {
		if err := gh.ciFakeGH.PostStatus(context.Background(), "test/repo", gh.pr(b).HeadOID, localStatusContext(suiteCheckName), "success", "branch suite"); err != nil {
			t.Fatal(err)
		}
	}
	carHeads := []string{gh.pr(landingTestBranch).HeadOID, gh.pr("000002-second").HeadOID, gh.pr("000003-third").HeadOID}
	ghClient = gh

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
	oldPoll, oldEvery, oldCarPoll := landingLinePoll, carVouchEvery, trainCarPoll
	landingLinePoll, carVouchEvery, trainCarPoll = time.Millisecond, 5*time.Millisecond, 5*time.Millisecond
	t.Cleanup(func() {
		resolveLandingWorkspace = oldResolve
		landingLinePoll, carVouchEvery, trainCarPoll = oldPoll, oldEvery, oldCarPoll
	})
	dir, err := landingLineDir(mergeRunner, first)
	if err != nil {
		t.Fatal(err)
	}
	line, err := landingline.Open(dir, landingline.Options{Poll: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	// A blocker holds the line until every car has vouched, so the train's
	// boarding is decided by the test, not by who started first.
	blocker, err := line.Enqueue(landingline.Ticket{Issue: 99, Branch: "000099-blocker"})
	if err != nil {
		t.Fatal(err)
	}
	if err := line.Wait(context.Background(), blocker); err != nil {
		t.Fatal(err)
	}
	start := func(root string) <-chan error {
		done := make(chan error, 1)
		go func() {
			f := landingFlags()
			f.Context = context.WithValue(context.Background(), invocationRootKey{}, root)
			done <- runMerge(io.Discard, io.Discard, f)
		}()
		return done
	}
	one := start(first)
	waitQueued(t, line, 1)
	two := start(second)
	waitQueued(t, line, 2)
	car3, err := line.Enqueue(landingline.Ticket{Issue: 3, Branch: "000003-third"})
	if err != nil {
		t.Fatal(err)
	}
	defer car3.Release()
	if err := line.Vouch(car3, landingline.Readiness{Branch: "000003-third", Head: carHeads[2], PR: 3, Ready: true}); err != nil {
		t.Fatal(err)
	}
	waitVouched(t, line, 3)
	third3 := make(chan error, 1)
	go func() { third3 <- line.Wait(context.Background(), car3) }()
	blocker.Release()

	for name, done := range map[string]<-chan error{"first": one, "second": two} {
		if err := finished(t, done); err != nil {
			t.Fatalf("%s car: %v", name, err)
		}
	}
	var settled *landingline.Settled
	if err := <-third3; !errors.As(err, &settled) || settled.Outcome.Kind != landingline.OutcomeLanded {
		t.Fatalf("third car: %v", err)
	}

	if gh.merges != 1 || len(gh.created) != 1 {
		t.Fatalf("merges %d, train PRs %v: want one train merged once", gh.merges, gh.created)
	}
	train := gh.pr(gh.created[0])
	suites := gh.suiteHeads()
	trainSuites := slices.DeleteFunc(slices.Clone(suites), func(s string) bool { return slices.Contains(carHeads, s) })
	if len(trainSuites) != 1 || trainSuites[0] != train.HeadOID {
		t.Fatalf("integration suites %v, want exactly the train head %s", trainSuites, train.HeadOID)
	}
	main := git(t, gh1.remote, "rev-parse", "refs/heads/main")
	for i, h := range carHeads {
		if _, err := trainGitOut(gh1.remote, "merge-base", "--is-ancestor", h, main); err != nil {
			t.Fatalf("car %d's head %s is not on main", i+1, h)
		}
	}
	for _, b := range []string{landingTestBranch, "000002-second", "000003-third"} {
		if p := gh.pr(b); p.State != "MERGED" {
			t.Fatalf("%s: %s", b, p.State)
		}
	}
	events := lineEvents(t, line)
	var trainEvents []landingline.Event
	for _, e := range events[landingline.EventTrain] {
		if e.Result == string(landingline.PhaseLanded) {
			trainEvents = append(trainEvents, e)
		}
	}
	if len(trainEvents) != 1 || len(trainEvents[0].Cars) != 3 || trainEvents[0].NextDue.IsZero() {
		t.Fatalf("train events %+v", events[landingline.EventTrain])
	}
	landed := 0
	for _, e := range events[landingline.EventCar] {
		if e.Result == string(landingline.OutcomeLanded) {
			landed++
		}
	}
	if landed != 3 {
		t.Fatalf("car events %+v", events[landingline.EventCar])
	}
	if p := gh.pr(gh.created[0]); p.State != "MERGED" {
		t.Fatalf("train PR %s", p.State)
	}
	if left, _ := filepath.Glob(filepath.Join(os.TempDir(), "sdlc-train-*")); len(left) > 0 {
		for _, l := range left {
			if _, err := os.Stat(filepath.Join(l, ".git")); err == nil {
				t.Errorf("a train worktree was left behind: %s", l)
			}
		}
	}
}

func waitVouched(t *testing.T, line *landingline.Line, n int) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for {
		ready, err := line.Readiness()
		count := 0
		for _, r := range ready {
			if r.Ready {
				count++
			}
		}
		if err == nil && count >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("only %d of %d cars vouched ready: %+v %v", count, n, ready, err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
