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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
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
	// lostReply: the merge happens, but its reply is lost (an error).
	lostReply bool
	// lagMerged: a merged train PR still lists OPEN for this many reads.
	lagMerged int
	// trainChecks are the CI checks every train head reports (green: none).
	trainChecks []prCheck
	created     []string
	closed      []int
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
	g.ciFakeGH.setHead(number, head, g.checksFor(branch)...)
	return pr
}

func (g *trainFakeGH) checksFor(branch string) []prCheck {
	if strings.HasPrefix(branch, trainBranchPrefix) {
		return g.trainChecks
	}
	return nil
}

func (g *trainFakeGH) LandingPRs(_ context.Context, _ string, branch string) ([]landingPR, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	pr, ok := g.prs[branch]
	if !ok {
		return nil, nil
	}
	if pr.State == "MERGED" && g.lagMerged > 0 && strings.HasPrefix(branch, trainBranchPrefix) {
		g.lagMerged--
		lagging := *pr
		lagging.State, lagging.MergeOID = "OPEN", ""
		return []landingPR{lagging}, nil
	}
	if pr.State == "OPEN" { // GitHub follows pushes to an open PR's branch
		if head, err := trainGitOut(g.remote, "rev-parse", "refs/heads/"+branch); err == nil && head != pr.HeadOID {
			pr.HeadOID = head
			g.ciFakeGH.setHead(pr.Number, head, g.checksFor(branch)...)
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
	if g.lostReply {
		g.lostReply = false
		return errors.New("trainFakeGH: connection reset after the merge")
	}
	return nil
}

// PostStatus accepts a status on any commit the server has, as GitHub does
// for a pushed commit: a bisect suite records on a prefix that is no PR's
// head, only an ancestor of the train branch.
func (g *trainFakeGH) PostStatus(ctx context.Context, repo, sha, name, state, desc string) error {
	if _, err := trainGitOut(g.remote, "cat-file", "-e", sha+"^{commit}"); err == nil {
		g.ciFakeGH.mu.Lock()
		g.ciFakeGH.known[sha] = true
		g.ciFakeGH.mu.Unlock()
	}
	return g.ciFakeGH.PostStatus(ctx, repo, sha, name, state, desc)
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
	return nil // GitHub keeps a closed PR's branch; the conductor deletes it
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

// trainSetup is a repository with trains on: the first slot's car (PR #1,
// a real sdlc merge), and a synthetic waiting car per extra branch, plus a
// blocker holding the line until the test releases it.
type trainSetup struct {
	first, server string
	gh            *trainFakeGH
	line          *landingline.Line
	blocker       *landingline.Handle
	heads         map[string]string // branch → car head
	rootKey       any
}

type trainRootKey struct{}

func newTrainSetup(t *testing.T, variant string) *trainSetup {
	t.Helper()
	first, server, gh1 := integrationFixture(t)
	procedureWrite(t, server, landingConfigPath, `{"train_interval_minutes": 60}`+"\n")
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "trains on")
	git(t, server, "push", "-q", "origin", "main")
	git(t, first, "fetch", "-q", "upstream")
	gh := newTrainFakeGH(t, gh1.remote, gh1.ci, variant)
	s := &trainSetup{first: first, server: server, gh: gh, heads: map[string]string{}, rootKey: trainRootKey{}}
	s.addCar(t, landingTestBranch, 1)
	ghClient = gh
	oldResolve := resolveLandingWorkspace
	resolveLandingWorkspace = func(r workspace.GitReader, dir, address string) (workspace.Identity, error) {
		if bounded, ok := r.(boundedLandingGitRunner); ok && dir == "." {
			if root, ok := bounded.ctx.Value(trainRootKey{}).(string); ok {
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
	if s.line, err = landingline.Open(dir, landingline.Options{Poll: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	// The blocker holds the line until every car has vouched, so boarding is
	// decided by the test, not by who started first.
	if s.blocker, err = s.line.Enqueue(landingline.Ticket{Issue: 99, Branch: "000099-blocker"}); err != nil {
		t.Fatal(err)
	}
	if err := s.line.Wait(context.Background(), s.blocker); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.blocker.Release)
	return s
}

// addCar registers PR number at branch's pushed head with its branch suite
// recorded there (the readiness every car vouches with).
func (s *trainSetup) addCar(t *testing.T, branch string, number int) {
	t.Helper()
	pr := s.gh.addPR(number, branch)
	if err := s.gh.ciFakeGH.PostStatus(context.Background(), "test/repo", pr.HeadOID, localStatusContext(suiteCheckName), "success", "branch suite"); err != nil {
		t.Fatal(err)
	}
	s.heads[branch] = pr.HeadOID
}

// pushCar makes a closed, pushed issue branch in its own clone of the server.
func (s *trainSetup) pushCar(t *testing.T, branch, file string, issue, number int) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), file)
	git(t, "", "clone", "-q", s.gh.remote, dir)
	git(t, dir, "config", "user.name", file)
	git(t, dir, "config", "user.email", file+"@example.com")
	git(t, dir, "remote", "rename", "origin", "upstream")
	trainCar(t, dir, branch, file, issue)
	s.addCar(t, branch, number)
	return dir
}

// merge starts a real sdlc merge in root.
func (s *trainSetup) merge(root string) <-chan error {
	done := make(chan error, 1)
	go func() {
		f := landingFlags()
		ctx := context.WithValue(context.Background(), trainRootKey{}, root)
		f.Context = ctx
		// The real repo transaction lock, as production's merge takes it:
		// cars run their cleanup concurrently once a train lands.
		cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
		cmd.SetContext(ctx)
		f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
		done <- runMerge(io.Discard, trainStderr, f)
	}()
	return done
}

// queuedCar enqueues a synthetic car that vouches for branch but never polls
// for admission (it can't take the line; a train can still board it).
func (s *trainSetup) queuedCar(t *testing.T, branch string, issue, number int) *landingline.Handle {
	t.Helper()
	h, err := s.line.Enqueue(landingline.Ticket{Issue: issue, Branch: branch})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.Release)
	if err := s.line.Vouch(h, landingline.Readiness{Branch: branch, Head: s.heads[branch], PR: number, Ready: true}); err != nil {
		t.Fatal(err)
	}
	return h
}

// waitingCar is a queuedCar that also waits, as a car's sdlc merge does.
func (s *trainSetup) waitingCar(t *testing.T, branch string, issue, number int) <-chan error {
	t.Helper()
	h := s.queuedCar(t, branch, issue, number)
	done := make(chan error, 1)
	go func() {
		err := s.line.Wait(context.Background(), h)
		var settled *landingline.Settled
		if errors.As(err, &settled) {
			h.Release() // a settled car leaves the line, as its merge does
		}
		done <- err
	}()
	return done
}

func wantSettled(t *testing.T, done <-chan error, kind landingline.OutcomeKind) landingline.Outcome {
	t.Helper()
	select {
	case err := <-done:
		var settled *landingline.Settled
		if !errors.As(err, &settled) || settled.Outcome.Kind != kind {
			t.Fatalf("car: %v, want %s", err, kind)
		}
		return settled.Outcome
	case <-time.After(2 * time.Minute):
		t.Fatal("car never settled")
	}
	return landingline.Outcome{}
}

// trainEvents are the logged trains with result.
func trainEvents(t *testing.T, line *landingline.Line, result string) []landingline.Event {
	var out []landingline.Event
	for _, e := range lineEvents(t, line)[landingline.EventTrain] {
		if e.Result == result {
			out = append(out, e)
		}
	}
	return out
}

func landTrainOfThree(t *testing.T, variant string) {
	s := newTrainSetup(t, variant)
	second := filepath.Join(filepath.Dir(filepath.Dir(s.first)), "sample-slot2", "sample")
	git(t, second, "fetch", "-q", "upstream")
	git(t, second, "reset", "-q", "--hard", "upstream/main")
	trainCar(t, second, "000002-second", "second", 2)
	s.addCar(t, "000002-second", 2)
	s.pushCar(t, "000003-third", "third", 3, 3)
	carHeads := []string{s.heads[landingTestBranch], s.heads["000002-second"], s.heads["000003-third"]}
	gh, line := s.gh, s.line

	one := s.merge(s.first)
	waitQueued(t, line, 1)
	two := s.merge(second)
	waitQueued(t, line, 2)
	three := s.waitingCar(t, "000003-third", 3, 3)
	waitVouched(t, line, 3)
	s.blocker.Release()

	for name, done := range map[string]<-chan error{"first": one, "second": two} {
		if err := finished(t, done); err != nil {
			t.Fatalf("%s car: %v", name, err)
		}
	}
	wantSettled(t, three, landingline.OutcomeLanded)

	if gh.merges != 1 || len(gh.created) != 1 {
		t.Fatalf("merges %d, train PRs %v: want one train merged once", gh.merges, gh.created)
	}
	train := gh.pr(gh.created[0])
	trainSuites := slices.DeleteFunc(gh.suiteHeads(), func(h string) bool { return slices.Contains(carHeads, h) })
	if len(trainSuites) != 1 || trainSuites[0] != train.HeadOID {
		t.Fatalf("integration suites %v, want exactly the train head %s", trainSuites, train.HeadOID)
	}
	main := git(t, gh.remote, "rev-parse", "refs/heads/main")
	for i, h := range carHeads {
		if _, err := trainGitOut(gh.remote, "merge-base", "--is-ancestor", h, main); err != nil {
			t.Fatalf("car %d's head %s is not on main", i+1, h)
		}
	}
	for _, b := range []string{landingTestBranch, "000002-second", "000003-third"} {
		if p := gh.pr(b); p.State != "MERGED" {
			t.Fatalf("%s: %s", b, p.State)
		}
	}
	if ev := trainEvents(t, line, string(landingline.PhaseLanded)); len(ev) != 1 || len(ev[0].Cars) != 3 || ev[0].NextDue.IsZero() {
		t.Fatalf("train events %+v", lineEvents(t, line)[landingline.EventTrain])
	}
	landed := 0
	for _, e := range lineEvents(t, line)[landingline.EventCar] {
		if e.Result == string(landingline.OutcomeLanded) {
			landed++
		}
	}
	if landed != 3 {
		t.Fatalf("car events %+v", lineEvents(t, line)[landingline.EventCar])
	}
	if train.State != "MERGED" {
		t.Fatalf("train PR %s", train.State)
	}
	wantNoTrainWorktree(t, s.first)
	wantNoTrainBranch(t, gh.remote)
}

func wantNoTrainBranch(t *testing.T, remote string) {
	t.Helper()
	if refs := git(t, remote, "for-each-ref", "--format=%(refname)", "refs/heads/"+trainBranchPrefix); refs != "" {
		t.Fatalf("train branches outlived their trains: %s", refs)
	}
}

func wantNoTrainWorktree(t *testing.T, root string) {
	t.Helper()
	if out := git(t, root, "worktree", "list", "--porcelain"); strings.Contains(out, "sdlc-train-") {
		t.Fatalf("a train worktree outlived its train:\n%s", out)
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

// With trains on and no other car ready, the head lands alone exactly as
// before #449, and logs the departure so the schedule paces.
func TestLandingTrainOfOneIsTodaysLanding(t *testing.T) {
	s := newTrainSetup(t, "train")
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	waitVouched(t, s.line, 1)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	if len(s.gh.created) != 0 || s.gh.merges != 1 || s.gh.pr(landingTestBranch).State != "MERGED" {
		t.Fatalf("train PRs %v merges %d", s.gh.created, s.gh.merges)
	}
	ev := trainEvents(t, s.line, "alone")
	if len(ev) != 1 || ev[0].NextDue.IsZero() || !strings.Contains(ev[0].Detail, "no other car") {
		t.Fatalf("train events %+v", lineEvents(t, s.line)[landingline.EventTrain])
	}
}

// A dead conductor's train PR is closed by the next train; a merge whose
// reply is lost is read back from GitHub and lands, never merged twice.
func TestLandingTrainReclaimsAnOrphanAndSurvivesALostMergeReply(t *testing.T) {
	s := newTrainSetup(t, "null")
	git(t, s.server, "fetch", "-q", "origin")
	git(t, s.server, "push", "-q", "origin", "origin/main:refs/heads/landing-train/7")
	// And a landed train's branch whose conductor died before deleting it.
	git(t, s.server, "push", "-q", "origin", "origin/main:refs/heads/landing-train/6")
	if err := s.gh.CreateTrainPR(context.Background(), "test/repo", "landing-train/7", "", ""); err != nil {
		t.Fatal(err)
	}
	orphan := s.gh.pr("landing-train/7").Number
	s.gh.lostReply = true
	s.pushCar(t, "000002-second", "second", 2, 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	wantSettled(t, two, landingline.OutcomeLanded)
	if !slices.Contains(s.gh.closed, orphan) || s.gh.pr("landing-train/7").State != "CLOSED" {
		t.Fatalf("the orphan train PR #%d was not closed: %v", orphan, s.gh.closed)
	}
	if s.gh.merges != 1 {
		t.Fatalf("merged %d times after a lost reply", s.gh.merges)
	}
	for _, b := range []string{landingTestBranch, "000002-second"} {
		if p := s.gh.pr(b); p.State != "MERGED" {
			t.Fatalf("%s: %s", b, p.State)
		}
	}
	wantNoTrainWorktree(t, s.first)
	wantNoTrainBranch(t, s.gh.remote)
}

// Main moving with docs only while the train's suite runs: the train merges
// main in again and keeps its suite (one suite, as a single landing would).
func TestLandingTrainReintegratesADocsOnlyMain(t *testing.T) {
	s := newTrainSetup(t, "train")
	started, release := blockLandingSuite(t, s.server)
	s.pushCar(t, "000002-second", "second", 2, 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	deadline := time.Now().Add(time.Minute)
	for {
		if b, err := os.ReadFile(started); err == nil && len(strings.TrimSpace(string(b))) == 40 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the train suite never started")
		}
		time.Sleep(5 * time.Millisecond)
	}
	prose := integrationAdvance(t, s.server, "workshop/parley/during-train.md", "published while the train's suite ran\n")
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	wantSettled(t, two, landingline.OutcomeLanded)
	main := git(t, s.gh.remote, "rev-parse", "refs/heads/main")
	if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", prose, main); err != nil {
		t.Fatal("landed without the docs main brought")
	}
	cars := []string{s.heads[landingTestBranch], s.heads["000002-second"]}
	runs := slices.DeleteFunc(s.gh.suiteHeads(), func(h string) bool { return slices.Contains(cars, h) })
	if len(runs) != 1 {
		t.Fatalf("suites on %v: a docs-only main move must not rerun the suite", runs)
	}
	if s.gh.merges != 1 {
		t.Fatalf("merges %d", s.gh.merges)
	}
}

// trainStderr is where the train tests' merges report; os.Stderr while debugging.
var trainStderr io.Writer = io.Discard

// A car whose PR is merged while it waits, with no outcome ever coming (its
// conductor died after the merge), finishes as a resume instead of hanging
// (BR-16): its own reading of GitHub ends the wait.
func TestLandingCarMergedWithoutAWordFinishes(t *testing.T) {
	s := newTrainSetup(t, "null")
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	waitVouched(t, s.line, 1)
	// The conductor's merge, with no settle after it: the blocker still holds
	// the line, so the car is never admitted.
	if err := s.gh.LandingMerge(context.Background(), "test/repo", 1, s.heads[landingTestBranch]); err != nil {
		t.Fatal(err)
	}
	if err := finished(t, one); err != nil {
		t.Fatalf("the merged car did not finish: %v", err)
	}
	if out := git(t, s.first, "branch", "--list", landingTestBranch); out != "" {
		t.Fatalf("the car's cleanup did not run: branch %q remains", out)
	}
}

// After an accepted train merge whose listing lags, the conductor sees its
// own head on main and lands, never reintegrating past its own merge (BR-17).
func TestLandingTrainLaggingListingIsItsOwnMerge(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.gh.lagMerged = 3
	s.pushCar(t, "000002-second", "second", 2, 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	wantSettled(t, two, landingline.OutcomeLanded)
	if s.gh.merges != 1 || len(trainEvents(t, s.line, string(landingline.PhaseLanded))) != 1 {
		t.Fatalf("merges %d trains %+v", s.gh.merges, lineEvents(t, s.line)[landingline.EventTrain])
	}
	wantNoTrainBranch(t, s.gh.remote)
}

// declareSuite replaces main's suite with command (sh -c), as an operator's
// commit to main would.
func (s *trainSetup) declareSuite(t *testing.T, command string) {
	t.Helper()
	procedureWrite(t, s.server, localChecksPath, `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c",`+strconv.Quote(command)+`]}]}`)
	git(t, s.server, "add", localChecksPath)
	git(t, s.server, "commit", "-qm", "suite for this test")
	git(t, s.server, "push", "-q", "origin", "main")
}

// An unexplained failure splits the train: the culprit is held, the clean
// prefix lands, and the cars behind the culprit ride the next train at once.
func TestLandingTrainSplitHoldsTheCulpritAndLandsThePrefix(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.declareSuite(t, "test ! -e second.go") // car 2 breaks the suite
	s.pushCar(t, "000002-second", "second", 2, 2)
	s.pushCar(t, "000003-third", "third", 3, 3)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	three := s.waitingCar(t, "000003-third", 3, 3)
	waitVouched(t, s.line, 3)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatalf("the clean prefix's car: %v", err)
	}
	held := wantSettled(t, two, landingline.OutcomeHeld)
	if !strings.Contains(held.Reason, "car 2 (#2)") {
		t.Fatalf("held for %q", held.Reason)
	}
	// Car 3 was released, still waiting; this test's line has no train rule,
	// so it is admitted next, as the immediate next train would admit it.
	if err := <-three; err != nil {
		t.Fatalf("the released car was not admitted next: %v", err)
	}
	if p := s.gh.pr("000002-second"); p.State != "OPEN" {
		t.Fatalf("the culprit's PR is %s", p.State)
	}
	if p := s.gh.pr("000003-third"); p.State != "OPEN" {
		t.Fatalf("a car behind the culprit is %s", p.State)
	}
	if p := s.gh.pr(landingTestBranch); p.State != "MERGED" {
		t.Fatalf("the clean prefix is %s", p.State)
	}
	ev := trainEvents(t, s.line, string(landingline.PhaseLanded))
	if len(ev) != 1 || ev[0].Splits != 2 || ev[0].Culprit == 0 {
		t.Fatalf("train %+v", lineEvents(t, s.line)[landingline.EventTrain])
	}
	if departed := ev[0].NextDue; departed.After(time.Now()) {
		t.Fatalf("a split left ready cars behind, yet the next train waits until %v", departed)
	}
	wantNoTrainBranch(t, s.gh.remote)
}

// A car whose merge conflicts with the train is held; the rest land.
func TestLandingTrainHoldsAConflictingCar(t *testing.T) {
	s := newTrainSetup(t, "head")
	dir := s.pushCar(t, "000002-second", "second", 2, 2)
	// Both cars write shipped.txt differently: car 2 cannot merge onto car 1.
	procedureWrite(t, dir, "shipped.txt", "a different feature\n")
	git(t, dir, "add", "shipped.txt")
	git(t, dir, "commit", "-qm", "conflicting change")
	git(t, dir, "push", "-q", "upstream", "000002-second")
	s.addCar(t, "000002-second", 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	held := wantSettled(t, two, landingline.OutcomeHeld)
	if !strings.Contains(held.Reason, "conflicts") {
		t.Fatalf("held for %q", held.Reason)
	}
	if s.gh.pr(landingTestBranch).State != "MERGED" || s.gh.pr("000002-second").State != "OPEN" {
		t.Fatal("the wrong car landed")
	}
}

// Red CI on the train PR refuses the whole train (no CI bisect).
func TestLandingTrainRedCIRefusesTheTrain(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.gh.trainChecks = []prCheck{{Name: "merge-check", Workflow: "merge-check", Bucket: "fail"}}
	s.pushCar(t, "000002-second", "second", 2, 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	err := finished(t, one)
	if err == nil || !strings.Contains(err.Error(), "CI is red") {
		t.Fatalf("the conductor's merge: %v", err)
	}
	refused := wantSettled(t, two, landingline.OutcomeRefused)
	if !strings.Contains(refused.Reason, "CI is red") {
		t.Fatalf("refused for %q", refused.Reason)
	}
	if s.gh.merges != 0 || s.gh.pr(s.gh.created[0]).State != "CLOSED" {
		t.Fatalf("merges %d; train PR %s", s.gh.merges, s.gh.pr(s.gh.created[0]).State)
	}
	if ev := trainEvents(t, s.line, string(landingline.PhaseRefused)); len(ev) != 1 {
		t.Fatalf("train %+v", lineEvents(t, s.line)[landingline.EventTrain])
	}
	wantNoTrainBranch(t, s.gh.remote) // closed, then deleted leased
}

// A fleet-blocking grant departs at once, though no train is due, and lands
// alone, outside any train.
func TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.blocker.Release() // this test admits by the train rule alone
	if err := s.line.Record(landingline.Event{Kind: landingline.EventTrain, Train: 1, Seq: 1, Result: "landed", NextDue: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := s.line.AddGrant(landingline.Grant{Issue: 1, Grantor: landingline.Grantor{Address: "operator shell"}, Reason: "blocks the fleet", Expires: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	s.pushCar(t, "000002-second", "second", 2, 2)
	s.queuedCar(t, "000002-second", 2, 2) // ready, ahead in line, but no train is due
	one := s.merge(s.first)
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	if len(s.gh.created) != 0 || s.gh.pr(landingTestBranch).State != "MERGED" || s.gh.pr("000002-second").State != "OPEN" {
		t.Fatalf("train PRs %v; the grant did not land alone", s.gh.created)
	}
	if ev := trainEvents(t, s.line, "alone"); len(ev) != 1 || !strings.Contains(ev[0].Detail, "fleet-blocking") {
		t.Fatalf("train %+v", lineEvents(t, s.line)[landingline.EventTrain])
	}
	if cars := lineEvents(t, s.line)[landingline.EventCar]; len(cars) != 0 {
		t.Fatalf("a grant boarded cars: %+v", cars)
	}
}

// A known flake whose rerun passed (#394) is the train's pass: it never
// splits. Any other failure is a split, unless main fails the same way.
func TestTrainSuiteOutcome(t *testing.T) {
	c := &conductor{heads: []string{strings.Repeat("a", 40)}}
	for name, tc := range map[string]struct {
		rec  localRunRecord
		want landingline.SuiteOutcome
	}{
		"pass":                    {localRunRecord{Exit: 0}, landingline.SuitePass},
		"known flake, rerun pass": {localRunRecord{Exit: 2, Outcome: flakes.KnownFlakePassed}, landingline.SuitePass},
		"known flake, accepted":   {localRunRecord{Exit: 2, Outcome: flakes.KnownFlakeAccept}, landingline.SuitePass},
	} {
		c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return tc.rec, nil }
		ev, err := c.runSuite(1)
		if err != nil || ev.(landingline.SuiteDone).Outcome != tc.want {
			t.Errorf("%s: %+v %v", name, ev, err)
		}
	}
	// An unexplained failure, with main's baseline already known green.
	green := false
	c.baseline = &green
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) {
		return localRunRecord{Exit: 2, Outcome: flakes.Unknown}, nil
	}
	if ev, err := c.runSuite(1); err != nil || ev.(landingline.SuiteDone).Outcome != landingline.SuiteFail {
		t.Fatalf("unexplained: %+v %v", ev, err)
	}
	red := true
	c.baseline = &red
	if ev, _ := c.runSuite(1); ev.(landingline.SuiteDone).Outcome != landingline.SuiteMainRed {
		t.Fatalf("main red: %+v", ev)
	}
}
