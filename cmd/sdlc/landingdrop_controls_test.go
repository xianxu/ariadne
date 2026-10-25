package main

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// These are stateful transport failures: a lost reply commits the close, a
// refused close leaves OPEN, and an unavailable listing proves neither state.
type dropCleanupGH struct {
	*trainFakeGH
	branch         string
	hide           bool
	lostCloseReply bool
	refuseClose    bool
	hideClosure    bool
}

func (g *dropCleanupGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	if branch == g.branch && g.hide {
		return nil, nil
	}
	prs, err := g.trainFakeGH.LandingPRs(ctx, repo, branch)
	if branch == g.branch && g.hideClosure {
		for i := range prs {
			if prs[i].State == "CLOSED" {
				prs[i].State = "OPEN"
			}
		}
	}
	return prs, err
}

func (g *dropCleanupGH) CloseTrainPR(ctx context.Context, repo string, number int, branch string) error {
	if branch == g.branch && g.refuseClose {
		return errors.New("close refused; PR remains open")
	}
	if err := g.trainFakeGH.CloseTrainPR(ctx, repo, number, branch); err != nil {
		return err
	}
	if branch == g.branch && g.lostCloseReply {
		return errors.New("connection reset after closing PR")
	}
	return nil
}

func TestLandingDropCleanupRequiresPositiveExternalProof(t *testing.T) {
	root, server, original := integrationFixture(t)
	gh := newTrainFakeGH(t, original.remote, original.ci, "train")
	head := procedureHead(t, root)
	other := procedureHead(t, server)
	for i, tc := range []struct {
		name, state, wantState, wantError                                         string
		hide, lostReply, refuse, hideClosure, moved, absent, notStarted, ordinary bool
	}{
		{name: "lost close reply", state: "OPEN", wantState: "CLOSED", lostReply: true},
		{name: "already closed", state: "CLOSED", wantState: "CLOSED"},
		{name: "closed and ref already absent", state: "CLOSED", wantState: "CLOSED", absent: true},
		{name: "missing publication listing", state: "OPEN", wantState: "OPEN", hide: true, wantError: "not positively observed"},
		{name: "close refused", state: "OPEN", wantState: "OPEN", refuse: true, wantError: "closure unconfirmed"},
		{name: "close state not yet visible", state: "OPEN", wantState: "CLOSED", hideClosure: true, wantError: "closure unconfirmed"},
		{name: "merged is not abandoned", state: "MERGED", wantState: "MERGED", wantError: "cannot confirm abandoned"},
		{name: "unknown ref tip retained", state: "CLOSED", wantState: "CLOSED", moved: true, wantError: "unrecognized head"},
		{name: "pushed before PR creation started", notStarted: true, hide: true},
		{name: "ordinary branch cannot authorize cleanup", state: "OPEN", wantState: "OPEN", ordinary: true, wantError: "not a train branch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			branch := trainBranch(int64(100 + i))
			if tc.ordinary {
				branch = "ordinary-feature"
			}
			git(t, root, "push", "-q", "upstream", head+":refs/heads/"+branch)
			if !tc.notStarted {
				gh.addPR(100+i, branch)
				gh.mu.Lock()
				gh.prs[branch].State = tc.state
				gh.mu.Unlock()
			}
			if tc.moved {
				git(t, gh.remote, "update-ref", "refs/heads/"+branch, other)
			}
			if tc.absent {
				git(t, gh.remote, "update-ref", "-d", "refs/heads/"+branch)
			}
			adapter := &dropCleanupGH{trainFakeGH: gh, branch: branch, hide: tc.hide, lostCloseReply: tc.lostReply, refuseClose: tc.refuse, hideClosure: tc.hideClosure}
			cleanup := landingline.TrainCleanup{Pending: true, Publication: landingline.TrainPublication{
				Started: !tc.notStarted, Repository: "test/repo", Remote: "upstream", Branch: branch,
				ExpectedTip: head, PR: 100 + i, RecoveryBranch: landingTestBranch,
			}}
			if tc.notStarted {
				cleanup.Publication.PR = 0
			}
			gitEffect := func(args ...string) (string, error) {
				// Observe actual state at the destructive effect, rather than only
				// checking a final state that could hide deletion-before-closure.
				if len(args) > 0 && args[0] == "push" && !tc.notStarted && gh.pr(branch).State != "CLOSED" {
					t.Error("train ref deletion attempted before PR was CLOSED")
				}
				return landingGit(landingTransportRunner{}, root, args...)
			}
			cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
			cmd.SetContext(context.Background())
			err := closeDroppedTrain(context.Background(), adapter, gitEffect, func(run func() error) error {
				return withRequiredRepoTransactionLock(cmd, run)
			}, cleanup)
			if tc.wantError == "" && err != nil || tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Fatalf("cleanup: %v, want error containing %q", err, tc.wantError)
			}
			if !tc.notStarted && gh.pr(branch).State != tc.wantState {
				t.Fatalf("PR state %s, want %s", gh.pr(branch).State, tc.wantState)
			}
			tip, err := remoteRefTip(func(args ...string) (string, error) { return landingGit(landingTransportRunner{}, root, args...) }, "upstream", "refs/heads/"+branch)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantError == "" && tip != "" || tc.wantError != "" && tip == "" {
				t.Fatalf("remote tip %q after cleanup, error expected %v", tip, tc.wantError != "")
			}
			if tc.moved && tip != other {
				t.Fatal("cleanup changed an unrecognized train ref")
			}
		})
	}
}

func TestLandingDropDeadConductorBlocksSerialUntilCleanupConfirmed(t *testing.T) {
	root, server, original := integrationFixture(t)
	gh := newTrainFakeGH(t, original.remote, original.ci, "train")
	s := &trainSetup{first: root, server: server, gh: gh, heads: map[string]string{}, boardOn: suiteCheckName}
	s.addCar(t, landingTestBranch, 1)
	const droppedBranch = "000002-dropped"
	droppedRoot := s.pushCar(t, droppedBranch, "dropped", 2, 2)
	line, err := openRepoLine(root)
	if err != nil {
		t.Fatal(err)
	}
	head, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch, QueueRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(head.Release)
	if err := line.Wait(context.Background(), head); err != nil {
		t.Fatal(err)
	}
	dropped, err := line.Enqueue(landingline.Ticket{Issue: 2, Branch: droppedBranch, QueueRoot: droppedRoot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dropped.Release)
	branch := trainBranch(head.Ticket().Seq)
	git(t, root, "push", "-q", "upstream", s.heads[landingTestBranch]+":refs/heads/"+branch)
	oldPR := gh.addPR(100, branch)
	member := landingline.TrainMember{Seq: dropped.Ticket().Seq, Issue: 2, Branch: droppedBranch, Head: s.heads[droppedBranch], PR: 2, OwnerRoot: droppedRoot}
	control := landingline.TrainControl{Version: landingline.TrainControlVersion, Train: head.Ticket().Seq, Phase: landingline.PhaseChecking,
		Members:     []landingline.TrainMember{{Seq: head.Ticket().Seq, Issue: 1, Branch: landingTestBranch, Head: s.heads[landingTestBranch], PR: 1, OwnerRoot: root}, member},
		Publication: landingline.TrainPublication{Started: true, Repository: "test/repo", Remote: "upstream", Branch: branch, ExpectedTip: oldPR.HeadOID, PR: oldPR.Number, RecoveryBranch: landingTestBranch}}
	if err := line.BeginTrain(head, control); err != nil {
		t.Fatal(err)
	}
	if _, err := line.RequestTrainDrop(control.Train, member, "must not board after crash"); err != nil {
		t.Fatal(err)
	}
	head.Release() // A process death releases this same lifetime lock.
	adapter := &dropCleanupGH{trainFakeGH: gh, branch: branch, hide: true}
	ghClient = adapter
	before := git(t, gh.remote, "rev-parse", "refs/heads/main")
	merge := func() error {
		f := landingFlags()
		cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
		cmd.SetContext(context.Background())
		f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
		return runMerge(io.Discard, io.Discard, f)
	}
	// The schedule has always been serial: reconciliation must precede even
	// the no-train path, rather than relying on a future conductor to notice.
	if err := merge(); err == nil || !strings.Contains(err.Error(), "drop accepted; cleanup pending") {
		t.Fatalf("serial successor with unknown publication: %v", err)
	}
	if got := git(t, gh.remote, "rev-parse", "refs/heads/main"); got != before || gh.pr(landingTestBranch).State != "OPEN" {
		t.Fatal("serial successor landed before predecessor cleanup was proved")
	}
	if tip := git(t, gh.remote, "rev-parse", "refs/heads/"+branch); tip != oldPR.HeadOID {
		t.Fatal("unknown publication ref changed")
	}
	stored, err := line.ReadTrainControl()
	if err != nil || !stored.Cleanup.Pending || len(stored.Drops) != 1 {
		t.Fatalf("lost cleanup/exclusion obligation: %+v %v", stored, err)
	}
	var held *landingline.Settled
	if err := line.Wait(context.Background(), dropped); !errors.As(err, &held) || held.Outcome.Kind != landingline.OutcomeHeld {
		t.Fatalf("accepted dropped ticket was not held after conductor death: %v", err)
	}
	adapter.hide = false
	gh.mu.Lock()
	gh.prs[branch].State = "CLOSED" // External positive proof becomes available.
	gh.mu.Unlock()
	if err := merge(); err != nil {
		t.Fatalf("serial successor after positive closure: %v", err)
	}
	stored, err = line.ReadTrainControl()
	if err != nil || stored.Cleanup.Pending {
		t.Fatalf("confirmed cleanup not durably retired: %+v %v", stored, err)
	}
	wantNoTrainBranch(t, gh.remote)
	main := git(t, gh.remote, "rev-parse", "refs/heads/main")
	if _, err := trainGitOut(gh.remote, "merge-base", "--is-ancestor", s.heads[landingTestBranch], main); err != nil {
		t.Fatal("serial successor did not land")
	}
	if _, err := trainGitOut(gh.remote, "merge-base", "--is-ancestor", s.heads[droppedBranch], main); err == nil || gh.pr(droppedBranch).State != "OPEN" {
		t.Fatal("dropped car landed with serial successor")
	}
	d := defaultLandingDeps()
	d.couchTree = func() (*int, bool) { return nil, false }
	d.dropIdentity = func(dir string) (workspace.Identity, error) {
		if dir == "." {
			dir = droppedRoot
		}
		id, err := workspace.Resolve(execGitRunner{}, dir, "")
		// This fixture uses independent local-file clones; production resolves
		// their common GitHub remote to the same repository identity.
		id.RepoIdentity = "github.com/test/repo"
		return id, err
	}
	out, err := runLanding(t, d, "drop", "--repo", root, "--issue", "2", "--reason", "retry receipt")
	if err != nil || !strings.Contains(out, "cleanup completed") || strings.Contains(out, "cleanup/recomposition pending") {
		t.Fatalf("completed drop receipt: %v\n%s", err, out)
	}
}

// The same commit and branch may exist in two forks. A content lease cannot
// establish which repository owns the predecessor's cleanup obligation.
func TestLandingDropRecoveryRefusesRetargetedRemote(t *testing.T) {
	root, _, original := integrationFixture(t)
	gh := newTrainFakeGH(t, original.remote, original.ci, "train")
	line, err := openRepoLine(root)
	if err != nil {
		t.Fatal(err)
	}
	old, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch, QueueRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(old.Release)
	if err := line.Wait(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	branch := trainBranch(old.Ticket().Seq)
	head := procedureHead(t, root)
	git(t, root, "push", "-q", "upstream", head+":refs/heads/"+branch)
	pr := gh.addPR(100, branch)
	member := landingline.TrainMember{Seq: old.Ticket().Seq, Issue: 1, Branch: landingTestBranch, Head: head, PR: 1, OwnerRoot: root}
	control := landingline.TrainControl{Version: landingline.TrainControlVersion, Train: old.Ticket().Seq, Phase: landingline.PhaseChecking,
		Members: []landingline.TrainMember{member}, Publication: landingline.TrainPublication{
			Started: true, Repository: "test/repo", Remote: "upstream", Branch: branch, ExpectedTip: head, PR: pr.Number, RecoveryBranch: landingTestBranch,
		}}
	if err := line.BeginTrain(old, control); err != nil {
		t.Fatal(err)
	}
	if _, err := line.RequestTrainDrop(control.Train, member, "repair before explicit retry"); err != nil {
		t.Fatal(err)
	}
	old.Release()
	// Clone after publication: the foreign destination deliberately contains
	// the exact expected SHA, so force-with-lease alone would permit deletion.
	foreign := filepath.Join(t.TempDir(), "foreign.git")
	git(t, "", "clone", "--bare", "-q", original.remote, foreign)
	const foreignURL = "https://github.com/foreign/repo.git"
	git(t, root, "config", "remote.upstream.url", foreignURL)
	git(t, root, "config", "url."+foreign+".insteadOf", foreignURL)
	target, err := resolveLandingTarget(mergeRunner)
	if err != nil || target == nil || target.Repo != "foreign/repo" {
		t.Fatalf("retargeted production identity: %+v %v", target, err)
	}
	next, err := line.Enqueue(landingline.Ticket{Issue: 3, Branch: "000003-successor", QueueRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(next.Release)
	if err := line.Wait(context.Background(), next); err != nil {
		t.Fatal(err)
	}
	f := landingFlags()
	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
	err = reconcileDroppedTrain(context.Background(), &landingTurn{line: line, handle: next}, *target, mergeRunner, gh, f)
	if err == nil || !strings.Contains(err.Error(), "cleanup pending") {
		t.Errorf("retargeted recovery must remain pending: %v", err)
	}
	if state := gh.pr(branch).State; state != "OPEN" {
		t.Errorf("closed predecessor PR before confirming cleanup destination: %s", state)
	}
	for name, remote := range map[string]string{"original": original.remote, "foreign": foreign} {
		tip, err := trainGitOut(remote, "rev-parse", "--verify", "refs/heads/"+branch)
		if err != nil || tip != head {
			t.Errorf("%s train ref changed during refused recovery: %q %v", name, tip, err)
		}
	}
	stored, err := line.ReadTrainControl()
	if err != nil || !stored.Cleanup.Pending {
		t.Errorf("retargeted recovery cleared original obligation: %+v %v", stored, err)
	}
}

type dropBlockedMergeGH struct {
	*trainFakeGH
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (g *dropBlockedMergeGH) LandingMerge(ctx context.Context, repo string, number int, head string) error {
	// Block outside the fake's mutex so the drop command can read authority,
	// control and PR state while the irreversible operation is in flight.
	g.once.Do(func() { close(g.entered) })
	select {
	case <-g.release:
	case <-ctx.Done():
		return ctx.Err()
	}
	return g.trainFakeGH.LandingMerge(ctx, repo, number, head)
}

func TestLandingDropRefusesWhileMergeAPIIsBlocked(t *testing.T) {
	s := newTrainSetup(t, "train")
	const secondBranch = "000002-second"
	s.pushCar(t, secondBranch, "second", 2, 2)
	adapter := &dropBlockedMergeGH{trainFakeGH: s.gh, entered: make(chan struct{}), release: make(chan struct{})}
	ghClient = adapter
	ctx, cancel := context.WithCancel(context.Background())
	var release sync.Once
	var pending chan error
	defer func() {
		cancel()
		release.Do(func() { close(adapter.release) })
		s.blocker.Release()
		if pending != nil {
			_ = finished(t, pending)
		}
	}()
	pending = s.mergeContext(ctx, s.first)
	waitQueued(t, s.line, 1, []chan error{pending})
	other := s.queuedCar(t, secondBranch, 2, 2)
	waitVouched(t, s.line, 2, []chan error{pending})
	s.blocker.Release()
	awaitDropCondition(t, "blocked train merge API", func() bool {
		select {
		case <-adapter.entered:
			return true
		default:
			return false
		}
	})
	before, err := s.line.ReadTrainControl()
	if err != nil || !before.MergeStarted || before.Phase != landingline.PhaseMerging {
		t.Fatalf("API began before durable merge fence: %+v %v", before, err)
	}
	d := defaultLandingDeps()
	d.couchTree = func() (*int, bool) { return nil, false }
	d.dropIdentity = func(dir string) (workspace.Identity, error) {
		if dir == "." {
			dir = s.first
		}
		return workspace.Resolve(execGitRunner{}, dir, "")
	}
	out, err := runLanding(t, d, "drop", "--repo", s.first, "--issue", "1", "--reason", "too late")
	if !errors.Is(err, landingline.ErrTrainMergeStarted) || strings.Contains(out, "dropped from train") {
		t.Fatalf("late drop accepted or wrong refusal: %v\n%s", err, out)
	}
	after, err := s.line.ReadTrainControl()
	if err != nil || !after.MergeStarted || len(after.Drops) != 0 || after.Train != before.Train {
		t.Fatalf("late refusal changed train/exclusions: %+v %v", after, err)
	}
	release.Do(func() { close(adapter.release) })
	done := pending
	err = finished(t, done)
	pending = nil
	if err != nil {
		t.Fatalf("merge after late-drop refusal: %v", err)
	}
	var landed *landingline.Settled
	if err := s.line.Wait(context.Background(), other); !errors.As(err, &landed) || landed.Outcome.Kind != landingline.OutcomeLanded {
		t.Fatalf("second car outcome after refused drop: %v", err)
	}
	main := git(t, s.gh.remote, "rev-parse", "refs/heads/main")
	for _, branch := range []string{landingTestBranch, secondBranch} {
		if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", s.heads[branch], main); err != nil {
			t.Fatalf("%s absent from main after late refusal", branch)
		}
	}
	t.Logf("train %d refused late drop while API blocked; both original heads landed", before.Train)
}
