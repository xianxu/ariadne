package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The landing train's conductor (#449). The admitted head conducts when it
// boards other cars: it builds one train branch (main, then each car's exact
// head), opens one train PR, runs one recorded suite and CI on it, and merges
// it once; GitHub then marks each car's own PR merged, because its head is
// reachable from main. The state machine is landingline.Transition; this file
// is the IO shell that executes its effects and feeds back their outcomes.

// trainBranchPrefix names the conductor's ephemeral branches.
const trainBranchPrefix = "landing-train/"

func trainBranch(id int64) string { return trainBranchPrefix + strconv.FormatInt(id, 10) }

// trainGH is GitHub as a conductor needs it, beyond a single landing.
type trainGH interface {
	landingGH
	// OpenTrainPRs lists the open PRs whose head is a landing-train branch.
	OpenTrainPRs(ctx context.Context, repo string) ([]landingPR, error)
	CreateTrainPR(ctx context.Context, repo, branch, title, body string) error
	// CloseTrainPR closes the PR; the caller deletes its branch, leased.
	CloseTrainPR(ctx context.Context, repo string, number int, branch string) error
}

// trainCarPoll is how often a landed car looks for GitHub to show its PR
// merged; trainCarWait bounds that wait. Tests shorten both.
var (
	trainCarPoll = 2 * time.Second
	trainCarWait = 2 * time.Minute
)

// conductor is one train's IO state. Only conductTrain mutates it.
type conductor struct {
	ctx    context.Context
	r      gitRunner
	gh     trainGH
	ci     ghCaller // the CI gate's and the suite's GitHub seam
	t      landingTarget
	f      *mergeFlags
	line   *landingline.Line
	log    *lineLog
	stderr io.Writer
	sched  landingline.Schedule
	id     int64
	self   int64               // the conductor's own ticket: its car is the first
	handle *landingline.Handle // lifetime capability for shared train control

	departed   time.Time
	wt         string   // the train's detached worktree
	base       string   // the main the train is built on
	heads      []string // heads[k-1] is prefix k's commit
	pushed     string   // what the train branch points at on the remote
	pr         landingPR
	opened     bool
	local      localRunDeps
	run        localCheckRunner
	suite      localCheck
	suiteDecl  *localCheckDecl // immutable main policy; derive again for each tested prefix
	bootstrap  bool            // the selected suite exists only on the subject
	landing    landingCheckOwner
	carHeads   []string // admitted originals, aligned with kept prefixes
	carIssues  []int    // same successfully composed prefix order; no rejected car inherits priority
	suiteReady map[int]bool
	evidence   map[int]localRunRecord // exact prefix attempts, including diagnostic main at zero
	baseline   *bool                  // main's baseline verdict, run once on the first failure
	settled    map[int64]bool
	selfEnd    landingline.OutcomeKind
	selfWhy    string
	// Measurement (#561): the tracker and the proofs the cars boarded on.
	meter  landingline.MeasurementTracker
	proofs map[int64]landingline.BoardingProof
}

// trainGit runs git in the repository's checkout.
func (c *conductor) git(args ...string) (string, error) { return landingGit(c.r, c.t.Root, args...) }

// wgit runs git in the train's worktree.
func (c *conductor) wgit(args ...string) (string, error) { return landingGit(c.r, c.wt, args...) }

// locked runs fn under the repo transaction lock, as merge's own ref writes
// do: a fetch, push or worktree change races every other sdlc in the
// repository's slots on the shared refs otherwise (#370).
func (c *conductor) locked(fn func() error) error { return c.f.withLock(fn) }

// fetchMain fetches main under the lock.
func (c *conductor) fetchMain() (main string, err error) {
	err = c.locked(func() error {
		main, err = c.t.fetchMain(c.r)
		return err
	})
	return main, err
}

// conductTrain runs one train of cars (the conductor's own first) to a
// terminal phase. It returns nil when the conductor's own car landed; any
// other end of its own car is an error naming why. Whatever happens, every
// other car gets an outcome and the train PR does not outlive the train.
func conductTrain(c *conductor, cars []landingline.Car, leftReady int) (err error) {
	c.settled = map[int64]bool{}
	tr := landingline.NewTrain(c.id)
	if err = c.beginTrainControl(cars); err != nil {
		return err
	}
	stopWatch := c.watchTrainDrops()
	defer stopWatch()
	defer func() {
		stopWatch()
		// An error mid-train is the machine's Aborted: every car still
		// riding goes back to wait, ready, and the train PR is closed. The
		// abort's effects are side effects only; their own failures are
		// reported, never retried here.
		if err != nil && !tr.Phase.Terminal() {
			err = c.abortTrain(&tr, cars, stopWatch, err)
		}
		if tr.Phase == landingline.PhaseLanded {
			c.deleteTrainBranch()
		}
		c.removeWorktree()
		c.measure(landingline.MeasureFinal{At: landingNow(), Phase: string(tr.Phase)})
		c.recordTrain(tr, cars, leftReady)
	}()
	tr, effects, err := landingline.Transition(tr, landingline.Boarded{Cars: cars})
	if err != nil {
		return err
	}
	if _, err = c.line.UpdateTrain(c.handle, tr); err != nil {
		return err
	}
	c.measureDeparture(cars)
	for _, car := range cars[1:] {
		if err := c.line.BindWait(c.handle, car); err != nil {
			fmt.Fprintf(c.stderr, "landing line: retained train wait evidence incomplete: binding ticket %d to train %d failed (%v); train continues\n", car.Seq, c.id, err)
		}
		c.settle(car, landingline.OutcomeBoarded, fmt.Sprintf("train %d, car %d of %d", c.id, indexOf(cars, car.Seq)+1, len(cars)))
	}
	for len(effects) > 0 {
		pending, controlErr := c.trainDropsPending()
		if controlErr != nil {
			return controlErr
		}
		if pending {
			return c.applyTrainDrops(&tr, cars, stopWatch)
		}
		e := effects[0]
		effects = effects[1:]
		ev, effectErr := c.execute(e, cars, suiteRunKind(tr))
		// The synchronous effect has joined before its completion competes with a
		// drop. Never feed a canceled, obsolete completion back into Transition.
		pending, controlErr = c.trainDropsPending()
		if controlErr != nil {
			return controlErr
		}
		if pending {
			return c.applyTrainDrops(&tr, cars, stopWatch)
		}
		if effectErr != nil {
			return effectErr
		}
		if ev == nil {
			continue
		}
		next, more, transitionErr := landingline.Transition(tr, ev)
		if transitionErr != nil {
			return transitionErr
		}
		before := tr
		if controlErr = c.commitTrainCompletion(&tr, next, cars, stopWatch); controlErr != nil {
			return controlErr
		}
		c.measureTransition(before, tr, ev)
		if _, built := ev.(landingline.Built); built {
			printTrainMembership(c.stderr, c.id, "composed", tr.Cars)
		}
		effects = append(effects, more...)
	}
	switch c.selfEnd {
	case landingline.OutcomeLanded:
		return nil
	case "":
		return fmt.Errorf("landing train %d ended %s without this merge's car", c.id, tr.Phase)
	}
	return fmt.Errorf("landing train %d: this merge's PR was %s: %s", c.id, c.selfEnd, c.selfWhy)
}

// execute runs one effect and returns the event its outcome is, if any. kind
// is what a RunSuite effect's run measures as (#561); "" outside the main
// loop, where no suite runs.
func (c *conductor) execute(e landingline.Effect, cars []landingline.Car, kind landingline.ExecutionKind) (landingline.TrainEvent, error) {
	switch e := e.(type) {
	case landingline.Build:
		return c.build(e.Cars)
	case landingline.Publish:
		// Composition's Hold effects precede Publish. Tell every conflicting
		// car its result before stopping; deferred Aborted cleanup releases the
		// survivors to a conductor whose own car can ride their suite.
		if c.selfEnd == landingline.OutcomeHeld {
			return nil, fmt.Errorf("landing train %d: this merge's PR was held: %s", c.id, c.selfWhy)
		}
		return nil, c.publish(e.Prefix)
	case landingline.RunSuite:
		return c.runSuite(e.Prefix, kind)
	case landingline.AssessSplit:
		done := c.assessSplit(e)
		if c.ctx.Err() != nil {
			return nil, c.ctx.Err()
		}
		return done, nil
	case landingline.RunLandingChecks:
		err := c.landing.run(c.ctx, c.git, c.base, c.heads[e.Prefix-1], c.carHeads[:e.Prefix], c.suiteReady[e.Prefix], c.run, func(ctx context.Context, check localCheck, head string) (localRunRecord, error) {
			return runLocalCheck(ctx, c.local, check, head)
		})
		if c.ctx.Err() != nil {
			return nil, c.ctx.Err()
		}
		done := landingline.LandingChecksDone{Prefix: e.Prefix, Passed: err == nil}
		if err != nil {
			done.Detail = err.Error()
		}
		return done, nil
	case landingline.WaitCI:
		return c.waitCI(e.Prefix), nil
	case landingline.Merge:
		return c.merge(e.Prefix)
	case landingline.Probe:
		return c.probe()
	case landingline.Reintegrate:
		return c.reintegrate(e.Prefix)
	case landingline.Hold:
		c.settle(carBySeq(cars, e.Seq), landingline.OutcomeHeld, e.Reason)
	case landingline.Release:
		for _, s := range e.Seqs {
			c.settle(carBySeq(cars, s), landingline.OutcomeReleased, "rides the next train")
		}
	case landingline.Refuse:
		for _, s := range e.Seqs {
			c.settle(carBySeq(cars, s), landingline.OutcomeRefused, e.Reason)
		}
	case landingline.LandCars:
		for _, s := range e.Seqs {
			c.settle(carBySeq(cars, s), landingline.OutcomeLanded, fmt.Sprintf("landed with train PR #%d", c.pr.Number))
		}
	case landingline.ClosePR:
		c.closePR()
	default:
		return nil, fmt.Errorf("landing train: unknown effect %T", e)
	}
	return nil, nil
}

// build merges main, then each car's exact head, in a detached worktree. A
// car whose merge conflicts, or whose prefix fails the instance or
// duplicate-ID gate, is a conflict: the merge is undone and the next car is
// tried on the train so far.
func (c *conductor) build(cars []landingline.Car) (landingline.TrainEvent, error) {
	main, err := c.fetchMain()
	if err != nil {
		return nil, err
	}
	c.base = main
	dir, err := os.MkdirTemp("", "sdlc-train-*")
	if err != nil {
		return nil, err
	}
	if err := os.Remove(dir); err != nil { // worktree add creates it
		return nil, err
	}
	if err := c.locked(func() error { _, err := c.git("worktree", "add", "--detach", dir, main); return err }); err != nil {
		return nil, err
	}
	c.wt = dir
	// The suite is the one main declares, read from the exact main the train
	// is built on (as a single landing reads its base). Its lock helper
	// (Makefile.workflow) resolves from the train's tree, falling back to the
	// checkout the train was built from, since the layered repos' workflow is
	// an untracked link to the sibling substrate (#486).
	decl, err := readLocalChecks(c.git, main)
	if err != nil {
		return nil, err
	}
	c.local.Root = dir
	c.local.PriorityIssues = c.priorityIssues
	c.local.WorkflowAnchor = c.t.Root // the train's tree sits apart from its substrate sibling (#486)
	// Keep execution native even when main has no declaration: each exact
	// prefix may introduce the bootstrap target. A later prefix's target or
	// weight must never authorize an earlier prefix.
	c.suiteDecl = &decl
	c.run = newLocalCheckRunner(c.local)
	c.suiteReady = map[int]bool{}
	built := landingline.Built{Conflicts: map[int64]string{}}
	for _, car := range cars {
		ref := fmt.Sprintf("refs/sdlc/train/%d/%d", c.id, car.Seq)
		var fetched string
		ferr := c.locked(func() error {
			if _, err := c.git("fetch", "--quiet", c.t.Remote, "+refs/heads/"+car.Branch+":"+ref); err != nil {
				return err
			}
			var err error
			fetched, err = c.git("rev-parse", "--verify", ref+"^{commit}")
			_, _ = c.git("update-ref", "-d", ref)
			return err
		})
		if ferr != nil {
			built.Conflicts[car.Seq] = "its branch could not be fetched: " + firstLine(ferr.Error())
			continue
		}
		if fetched != car.Head {
			built.Conflicts[car.Seq] = fmt.Sprintf("its branch moved off the vouched head %s", shortSHA(car.Head))
			continue
		}
		msg := fmt.Sprintf("Landing train %d: #%d (PR #%d)\n\nLanding-Train: %d\nLanding-PR: %d", c.id, car.Issue, car.PR, c.id, car.PR)
		if _, err := c.wgit("-c", "merge.autoStash=false", "-c", "rerere.enabled=false", "merge", "--no-ff", "--no-edit", "-m", agentid.StampCurrent(msg, c.ctx), car.Head); err != nil {
			_, _ = landingGit(recoveryLandingGit(c.r), c.wt, "merge", "--abort")
			built.Conflicts[car.Seq] = "its merge conflicts with main or with the cars before it on the train; integrate main, re-close and run sdlc merge"
			continue
		}
		head, err := c.wgit("rev-parse", "HEAD")
		if err != nil {
			return nil, err
		}
		if why := c.gatePrefix(head); why != "" {
			if _, err := c.wgit("reset", "--hard", "HEAD^"); err != nil {
				return nil, err
			}
			built.Conflicts[car.Seq] = why
			continue
		}
		built.Kept = append(built.Kept, car.Seq)
		c.heads = append(c.heads, head)
		c.carHeads = append(c.carHeads, car.Head)
		c.carIssues = append(c.carIssues, car.Issue)
	}
	return built, nil
}

// gatePrefix runs the instance and duplicate-ID gates on one prefix, as a
// single landing runs them on its integrated head; "" is a pass.
func (c *conductor) gatePrefix(head string) string {
	return c.gatePrefixAt(c.base, head)
}

func (c *conductor) gatePrefixAt(base, head string) string {
	if c.f.NoValidate {
		return ""
	}
	// Each prefix is gated in the train tree that holds it, never the
	// conductor's checkout, which lacks later cars' files (#490).
	if err := runLandingInstanceGate(c.wt, base, head, nounGates(c.f.IssuesDir), io.Discard, c.stderr, c.r); err != nil {
		return "the instance gate refuses the train with it: " + firstLine(err.Error())
	}
	if err := runLandingDuplicateGate(base, head, c.f.IssuesDir, c.f.HistoryDir, c.r); err != nil {
		return "the duplicate-ID gate refuses the train with it: " + firstLine(err.Error())
	}
	return ""
}

// publish points the train branch at prefix k and the train PR at it,
// opening the PR the first time, and waits for GitHub to register the head.
func (c *conductor) publish(k int) error {
	head := c.heads[k-1]
	branch := trainBranch(c.id)
	if _, err := c.line.UpdateTrainPublication(c.handle, landingline.TrainPublication{
		Started: c.opened, Repository: c.t.Repo, Remote: c.t.Remote, Branch: branch,
		ExpectedTip: head, PreviousTip: c.pushed, PR: c.pr.Number, RecoveryBranch: c.handle.Ticket().Branch,
	}); err != nil {
		return err
	}
	if err := c.locked(func() error {
		_, err := c.git("push", "--quiet", "--force-with-lease=refs/heads/"+branch+":"+c.pushed, c.t.Remote, head+":refs/heads/"+branch)
		return err
	}); err != nil {
		return fmt.Errorf("push the train branch: %w", err)
	}
	c.pushed = head
	if !c.opened {
		// The push identity is already durable. Fence PR creation separately:
		// a drop during the push can prove no PR create was ever attempted.
		if _, err := c.line.UpdateTrainPublication(c.handle, landingline.TrainPublication{
			Started: true, Repository: c.t.Repo, Remote: c.t.Remote, Branch: branch,
			ExpectedTip: head, PreviousTip: head, RecoveryBranch: c.handle.Ticket().Branch,
		}); err != nil {
			return err
		}
		title, body := c.prText()
		if err := c.gh.CreateTrainPR(c.ctx, c.t.Repo, branch, title, body); err != nil {
			return fmt.Errorf("open the train PR: %w", err)
		}
		c.opened = true
	}
	ctx, cancel := context.WithTimeout(c.ctx, integrationSyncTimeout)
	defer cancel()
	for {
		prs, err := c.gh.LandingPRs(ctx, c.t.Repo, branch)
		if err != nil {
			return err
		}
		for _, p := range prs {
			if p.State == "OPEN" && p.HeadOID == head {
				c.pr = p
				c.local.PR = p.Number
				_, err := c.line.UpdateTrainPublication(c.handle, landingline.TrainPublication{
					Started: true, Repository: c.t.Repo, Remote: c.t.Remote, Branch: branch,
					ExpectedTip: head, PreviousTip: c.pushed, PR: p.Number, RecoveryBranch: c.handle.Ticket().Branch,
				})
				return err
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("GitHub has not registered the train head %s: %w", shortSHA(head), ctx.Err())
		case <-time.After(integrationPollInterval):
		}
	}
}

func (c *conductor) prText() (string, string) {
	title := fmt.Sprintf("Landing train %d", c.id)
	body := fmt.Sprintf("sdlc merge landing train %d (#449): main %s, then each car's exact head. Merging this PR lands every car; GitHub marks each car's PR merged.\n", c.id, shortSHA(c.base))
	return title, body
}

// priorityIssues binds priority only to the currently composed, tested prefix.
// Recomposition replaces heads; split runs cannot borrow a later car's grant.
func (c *conductor) priorityIssues(sha string) []int {
	for i, head := range c.heads {
		if head != "" && head == sha && i < len(c.carIssues) {
			return append([]int(nil), c.carIssues[:i+1]...)
		}
	}
	return nil
}

// deriveSuite selects policy for one exact composed subject. Explicit main
// declarations may intentionally omit a suite; proven undeclared absence may
// not stand in for a passing bootstrap run.
func (c *conductor) deriveSuite(head string) error {
	decl, bootstrap, err := withSubjectDefaultSuiteContext(c.ctx, c.git, c.base, head, *c.suiteDecl)
	if err != nil {
		return err
	}
	suite, ok := decl.check(suiteCheckName)
	if !ok && c.suiteDecl.Version == 0 {
		return fmt.Errorf("unverified: train prefix %s has no bootstrap make test target; add the target, push and retry sdlc merge", shortSHA(head))
	}
	c.suite, c.bootstrap = suite, bootstrap
	if bootstrap {
		fmt.Fprintf(c.stderr, "bootstrap suite: make test from train prefix %s (main has no declaration or test target)\n", shortSHA(head))
	}
	return nil
}

// runSuite runs the recorded suite on prefix k. A failure of the full train
// is compared once with main's baseline: main failing the same way is no
// car's fault.
func (c *conductor) runSuite(k int, kind landingline.ExecutionKind) (landingline.TrainEvent, error) {
	head := c.heads[k-1]
	if head == "" {
		return nil, fmt.Errorf("unverified: prefix %d needs recomposition after main's code advance", k)
	}
	if c.suiteDecl != nil {
		if err := c.deriveSuite(head); err != nil {
			return nil, err
		}
	}
	var rec localRunRecord
	var err error
	if c.suiteDecl != nil && len(c.suite.Command) == 0 {
		rec = localRunRecord{Exit: 0} // an explicit main policy intentionally has no suite
	} else {
		rec, err = c.run(c.ctx, c.suite, head)
		c.measureExecution(kind, rec, err)
	}
	if err != nil {
		return nil, fmt.Errorf("the train suite on %s could not run: %w", shortSHA(head), err)
	}
	if c.evidence == nil {
		c.evidence = map[int]localRunRecord{}
	}
	c.evidence[k] = rec
	done := landingline.SuiteDone{Prefix: k, Outcome: landingline.SuitePass}
	if rec.passed() {
		if c.suiteReady == nil {
			c.suiteReady = map[int]bool{}
		}
		c.suiteReady[k] = len(c.suite.Command) > 0
		return done, nil
	}
	done.Outcome = landingline.SuiteFail
	if c.bootstrap {
		// Main has no target to run. Leave its baseline evidence absent so
		// split attribution reports unverified rather than inventing main-red.
		return done, nil
	}
	if c.baseline == nil {
		red, err := c.mainIsRed(rec)
		if err != nil {
			return nil, err
		}
		c.baseline = &red
	}
	if *c.baseline {
		done.Outcome = landingline.SuiteMainRed
	}
	return done, nil
}

// mainIsRed runs main's baseline, unrecorded, and reports whether it fails
// every test the train failed: then no car introduced the failure.
func (c *conductor) mainIsRed(failed localRunRecord) (bool, error) {
	mergedLog, err := os.ReadFile(failed.LogPath)
	if err != nil {
		return false, fmt.Errorf("train baseline comparison could not read suite log: %w", err)
	}
	deps := c.local
	deps.SkipRecord, deps.PR = true, 0
	baseline, err := runLocalCheck(c.ctx, deps, c.suite, c.base)
	c.measureExecution(landingline.ExecBaseline, baseline, err)
	if err != nil {
		return false, fmt.Errorf("train baseline on main could not run: %w", err)
	}
	c.evidence[0] = baseline
	if baseline.passed() {
		return false, nil
	}
	baseLog, err := os.ReadFile(baseline.LogPath)
	if err != nil {
		return false, fmt.Errorf("train baseline log unreadable: %w", err)
	}
	merged, base := failingTests(string(mergedLog)), failingTests(string(baseLog))
	if len(base) == 0 {
		return true, nil // main is red and unparsed: no car can be blamed
	}
	introduced, _ := classifyFailures(merged, base)
	return len(introduced) == 0, nil
}

// waitCI waits for the train PR's CI on prefix k, the gate a single landing
// waits for; any refusal is the train's CI being red.
func (c *conductor) waitCI(k int) landingline.TrainEvent {
	head := c.heads[k-1]
	ioGit := func(ioCtx context.Context) func(...string) (string, error) {
		rr := withLandingGitContext(ioCtx, c.r)
		return func(args ...string) (string, error) { return landingGit(rr, c.t.Root, args...) }
	}
	if err := waitCIGate(c.ctx, c.ci, c.t.Repo, c.pr.Number, c.base, head, c.git, c.run, c.f.NoCI, c.stderr, ioGit); err != nil {
		return landingline.CIDone{Detail: firstLine(err.Error())}
	}
	return landingline.CIDone{Green: true}
}

// merge lands prefix k: moved main reintegrates unless the shared issue-record
// proof permits keeping the exact checked head. The merge is pinned to the head and
// confirmed on GitHub; an unconfirmed one is probed, never assumed.
func (c *conductor) merge(k int) (landingline.TrainEvent, error) {
	head := c.heads[k-1]
	// Read outside the transaction; only its exact live base can authorize
	// retaining this checked head after the final fetch below.
	readiness, readinessErr := landingReadiness{}, fmt.Errorf("adapter has no readiness evidence")
	if source, ok := c.gh.(landingReadinessGH); ok {
		readiness, readinessErr = source.LandingReadiness(c.ctx, c.t.Repo, c.pr.Number)
	}
	if err := c.ctx.Err(); err != nil {
		return nil, err
	}
	moved, uncertain, landed := false, false, false
	// One transaction, as a single landing's: main re-read and the pinned
	// merge request under the lock, so main can't move between them.
	if err := c.locked(func() error {
		main, err := c.t.fetchMain(c.r)
		if err != nil {
			return err
		}
		if main != c.base {
			// Main moved, or the train itself landed while GitHub's listing
			// lagged: the train's head on main is this train merged, never a
			// move to reintegrate past.
			if _, err := c.git("merge-base", "--is-ancestor", head, main); err == nil {
				landed = true
				return nil
			}
			only, proofErr := mainMovementIssueOnly(c.r, c.t.Root, c.base, main)
			if proofErr != nil {
				fmt.Fprintf(c.stderr, "issue-record optimization unavailable (%v); reintegrating train normally\n", proofErr)
			}
			if only && readinessErr != nil {
				fmt.Fprintf(c.stderr, "issue-record readiness unavailable (%v); reintegrating train normally\n", readinessErr)
			}
			ready := false
			if only && proofErr == nil && readinessErr == nil {
				ready, err = readinessAllowsUnintegrated(c.pr, readiness, main)
				if err != nil {
					return err
				}
				ready = ready && c.landing.allowsMainMove(c.git, c.base, main, c.carHeads[:k])
			}
			if !ready {
				moved = true
				return nil
			}
			if why := c.gatePrefixAt(main, head); why != "" {
				return errors.New(why)
			}
			fmt.Fprintf(c.stderr, "landing train %d: issue records advanced main; retaining checked head %s and CI\n", c.id, shortSHA(head))
		}
		if err := c.gh.LandingMerge(c.ctx, c.t.Repo, c.pr.Number, head); err != nil {
			fmt.Fprintf(c.stderr, "landing train %d: merge outcome uncertain (%v); reading GitHub\n", c.id, err)
			uncertain = true
		}
		return nil
	}); err != nil {
		return nil, err
	}
	switch {
	case landed:
		return landingline.Merged{}, nil
	case moved:
		return landingline.MainMoved{}, nil
	case uncertain:
		return landingline.MergeUnknown{}, nil
	}
	ev, err := c.probe()
	if err != nil {
		return nil, err
	}
	if p := ev.(landingline.Probed); p.Merged {
		return landingline.Merged{}, nil
	}
	return landingline.MergeUnknown{}, nil
}

// probe reads the train PR's state.
func (c *conductor) probe() (landingline.TrainEvent, error) {
	prs, err := c.gh.LandingPRs(c.ctx, c.t.Repo, trainBranch(c.id))
	if err != nil {
		return nil, err
	}
	for _, p := range prs {
		if p.Number == c.pr.Number && p.State == "MERGED" {
			// Each car proves its own landing through landedCommit; the train
			// PR's raw mergeCommit is not evidence for any of them.
			return landingline.Probed{Merged: true}, nil
		}
	}
	return landingline.Probed{}, nil
}

// reintegrate merges current main into prefix k's head. The suite carries
// over only when what main brought has no code surface (the publish gate's
// line, as a single landing's docs-only reuse).
func (c *conductor) reintegrate(k int) (landingline.TrainEvent, error) {
	main, err := c.fetchMain()
	if err != nil {
		return nil, err
	}
	old := c.heads[k-1]
	if _, err := c.wgit("checkout", "--quiet", "--detach", old); err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("Merge %s into landing train %d (sdlc merge, #449)\n\n%s: %s\nLanding-Train: %d", main, c.id, integrationTrailer, main, c.id)
	if _, err := c.wgit("-c", "merge.autoStash=false", "-c", "rerere.enabled=false", "merge", "--no-ff", "--no-edit", "-m", agentid.StampCurrent(msg, c.ctx), main); err != nil {
		_, _ = landingGit(recoveryLandingGit(c.r), c.wt, "merge", "--abort")
		return nil, fmt.Errorf("main %s conflicts with the train; the cars go back to wait", shortSHA(main))
	}
	head, err := c.wgit("rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	diff, err := c.git(gitx.ChangedPathArgs(old, head)...)
	if err != nil {
		return nil, err
	}
	c.heads[k-1], c.base = head, main
	needSuite := publishGateHasCodeSurface(splitNUL(diff))
	if needSuite {
		// Smaller prefix commits still contain the old base. The reducer
		// refuses a later split instead of treating them as current evidence.
		clear(c.heads[:k-1])
		c.evidence = nil
		c.baseline = nil
		c.suiteReady = map[int]bool{}
	}
	decl, err := readLocalChecks(c.git, main)
	if err != nil {
		return nil, err
	}
	c.suiteDecl = &decl
	c.run = newLocalCheckRunner(c.local)
	if err := c.deriveSuite(head); err != nil {
		return nil, err
	}
	return landingline.Reintegrated{NeedSuite: needSuite}, nil
}

// settle tells one car what happened, and logs it. The conductor's own car
// is remembered instead: it has no waiter to tell.
func (c *conductor) settle(car landingline.Car, kind landingline.OutcomeKind, why string) {
	if car.Seq == 0 || c.settled[car.Seq] && kind != landingline.OutcomeBoarded {
		return
	}
	if kind != landingline.OutcomeBoarded {
		c.settled[car.Seq] = true
		c.log.record(landingline.Event{Kind: landingline.EventCar, Seq: car.Seq, Issue: car.Issue, Branch: car.Branch, PR: car.PR, Train: c.id, Result: string(kind), Detail: why})
	}
	if car.Seq == c.self {
		if kind != landingline.OutcomeBoarded {
			c.selfEnd, c.selfWhy = kind, why
		}
		if kind == landingline.OutcomeHeld {
			reportTrainHeld(c.stderr, c.id, car.Issue, car.Branch, why)
		} else if kind == landingline.OutcomeRefused {
			reportTrainUnattributed(c.stderr, c.id, car.Issue, car.Branch, why)
		}
		return
	}
	err := c.line.Settle(car.Seq, landingline.Outcome{Train: c.id, Kind: kind, Reason: why, PR: c.pr.Number})
	if err != nil && !errors.Is(err, landingline.ErrGone) {
		fmt.Fprintf(c.stderr, "landing train %d: could not tell car #%d it was %s (%v); its sdlc merge sees its PR's state on GitHub instead\n", c.id, car.Issue, kind, err)
	}
}

// reportTrainHeld tells the owning runner why its car stopped and how to
// recover; an immediate retry without fixing and re-closing isn't sufficient.
func reportTrainHeld(stderr io.Writer, train int64, issue int, branch, reason string) {
	fmt.Fprintf(stderr, "landing train %d: #%d held: %s\n", train, issue, reason)
	fmt.Fprintf(stderr, "Recovery: fix the reported problem, integrate main, re-close, then run sdlc merge --branch %s --yes\n", branch)
}

// reportTrainUnattributed stops explicitly; it never suggests that a car caused
// the failure or that an immediate automatic retry can create passing evidence.
func reportTrainUnattributed(stderr io.Writer, train int64, issue int, branch, reason string) {
	if !strings.HasPrefix(reason, "unattributed:") && !strings.HasPrefix(reason, "unverified:") {
		return
	}
	fmt.Fprintf(stderr, "landing train %d: #%d stopped: %s\n", train, issue, reason)
	fmt.Fprintf(stderr, "Recovery: inspect the named evidence and obtain valid verification, then run sdlc merge --branch %s --yes\n", branch)
}

func printTrainMembership(stderr io.Writer, train int64, label string, cars []landingline.Car) {
	fmt.Fprintf(stderr, "landing train %d %s: %d car(s):", train, label, len(cars))
	for _, car := range cars {
		fmt.Fprintf(stderr, " #%d (PR #%d)", car.Issue, car.PR)
	}
	fmt.Fprintln(stderr)
}

// closePR closes the train PR and deletes its branch, once.
func (c *conductor) closePR() {
	if !c.opened {
		c.deleteTrainBranch()
		return
	}
	c.opened = false
	if err := c.gh.CloseTrainPR(c.ctx, c.t.Repo, c.pr.Number, trainBranch(c.id)); err != nil {
		fmt.Fprintf(c.stderr, "landing train %d: could not close train PR #%d (%v); the next train closes it\n", c.id, c.pr.Number, err)
		return
	}
	c.deleteTrainBranch()
}

// deleteTrainBranch removes the train's branch from the remote once nothing
// reads it again (landed, or closed without a PR).
func (c *conductor) deleteTrainBranch() {
	if c.pushed == "" {
		return
	}
	branch := trainBranch(c.id)
	if err := deleteTrainRef(c.git, c.locked, c.t.Remote, branch, c.pushed); err != nil {
		fmt.Fprintf(c.stderr, "landing train %d: could not delete %s (%v); the next train removes it\n", c.id, branch, err)
	}
}

// deleteTrainRef is every train-branch delete: under the repo lock, and
// leased on the tip its caller pushed or observed, so it never removes a
// commit the deleter did not see.
func deleteTrainRef(git func(args ...string) (string, error), lock func(func() error) error, remote, branch, tip string) error {
	if !strings.HasPrefix(branch, trainBranchPrefix) || !landingOIDValid(tip) {
		return fmt.Errorf("refusing to delete %q at %q: not a train branch at a known tip", branch, tip)
	}
	return lock(func() error {
		_, err := git("push", "--quiet", "--force-with-lease=refs/heads/"+branch+":"+tip, remote, ":refs/heads/"+branch)
		return err
	})
}

func (c *conductor) removeWorktree() {
	if c.wt == "" {
		return
	}
	_ = c.locked(func() error {
		if _, err := landingGit(recoveryLandingGit(c.r), c.t.Root, "worktree", "remove", "--force", c.wt); err != nil {
			os.RemoveAll(c.wt)
			_, _ = landingGit(recoveryLandingGit(c.r), c.t.Root, "worktree", "prune")
		}
		return nil
	})
	c.wt = ""
}

// recordTrain logs the train's outcome; its next_due is the schedule.
func (c *conductor) recordTrain(tr landingline.Train, cars []landingline.Car, leftReady int) {
	seqs := make([]int64, len(cars))
	for i, car := range cars {
		seqs[i] = car.Seq
	}
	c.log.record(landingline.Event{Kind: landingline.EventTrain, Seq: c.id, Train: c.id, Cars: seqs, PR: c.pr.Number,
		Result: string(tr.Phase), Splits: tr.Splits, Culprit: tr.Culprit,
		NextDue: landingline.NextDue(c.departed, c.sched, leftReady > 0 || len(tr.Released) > 0), Departed: c.departed})
	fmt.Fprintf(c.stderr, "landing train %d: %s; %d car(s), %d split suite(s)\n", c.id, tr.Phase, len(tr.Cars), tr.Splits)
}

// reclaimTrains closes train PRs a dead conductor left open: head.lock was
// free when this one was admitted, so no live conductor owns them. Train
// branches with no open PR (a landed train whose conductor died before
// deleting it) are deleted too, leased on the tip read.
func reclaimTrains(ctx context.Context, gh trainGH, git func(args ...string) (string, error), remote, repo string, lock func(func() error) error, stderr io.Writer) {
	prs, err := gh.OpenTrainPRs(ctx, repo)
	if err != nil {
		fmt.Fprintf(stderr, "landing train: could not list leftover train PRs (%v)\n", err)
		return
	}
	open := map[string]bool{}
	for _, p := range prs {
		open[p.HeadRef] = true
		if err := gh.CloseTrainPR(ctx, repo, p.Number, p.HeadRef); err != nil {
			fmt.Fprintf(stderr, "landing train: could not close leftover train PR #%d (%v)\n", p.Number, err)
			continue
		}
		fmt.Fprintf(stderr, "landing train: closed leftover train PR #%d (%s): its conductor is gone\n", p.Number, p.HeadRef)
		_ = deleteTrainRef(git, lock, remote, p.HeadRef, p.HeadOID)
	}
	out, err := git("ls-remote", remote, "refs/heads/"+trainBranchPrefix+"*")
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		branch := strings.TrimPrefix(f[1], "refs/heads/")
		if open[branch] || !strings.HasPrefix(branch, trainBranchPrefix) {
			continue
		}
		if err := deleteTrainRef(git, lock, remote, branch, f[0]); err == nil {
			fmt.Fprintf(stderr, "landing train: deleted leftover branch %s\n", branch)
		}
	}
}

func indexOf(cars []landingline.Car, seq int64) int {
	for i, c := range cars {
		if c.Seq == seq {
			return i
		}
	}
	return -1
}

func carBySeq(cars []landingline.Car, seq int64) landingline.Car {
	if i := indexOf(cars, seq); i >= 0 {
		return cars[i]
	}
	return landingline.Car{}
}

// maybeConductTrain decides, as the admitted head, whether this landing is a
// train: it boards every waiter whose vouch holds and whose remote tip and
// open PR still sit at the vouched head. With no other car, it records the
// departure and returns to today's single landing. Otherwise it conducts,
// and on success returns its own PR as GitHub shows it merged.
func maybeConductTrain(ctx context.Context, r gitRunner, lgh landingGH, ci ghCaller, t landingTarget, f *mergeFlags, turn *landingTurn, branch, head string, pr landingPR, s landingline.Schedule, readiness func(context.Context) (landingline.Readiness, bool, error), stderr io.Writer) (landingPR, bool, error) {
	if s.Off() {
		return pr, false, nil
	}
	now := landingNow()
	tk := turn.handle.Ticket()
	serial := func(why string, leftReady int) (landingPR, bool, error) {
		turn.log.record(landingline.Event{Kind: landingline.EventTrain, Seq: tk.Seq, Train: tk.Seq, Cars: []int64{tk.Seq}, Result: "alone",
			// Not a departure (#475): it never pushes the next train back.
			Detail: why, NextDue: landingline.AloneNextDue(now, leftReady > 0), Departed: now})
		return pr, false, nil
	}
	gh, ok := lgh.(trainGH)
	if !ok {
		return serial("the GitHub adapter has no train support", 0)
	}
	grants, err := turn.line.GrantsNow()
	if err != nil {
		grants = nil // unreadable: no priority, as the line's admission reads it
	}
	if _, granted := grants[tk.Issue]; granted {
		return serial("fleet-blocking: lands alone, outside the train", 0)
	}
	self, _, _ := readiness(ctx) // already admitted: fresh landing gates own any effects
	if !self.Ready || self.Head != head || self.PR != pr.Number {
		return serial("this merge's own gates do not hold: "+self.Reason, 0)
	}
	_, waiting, err := turn.line.Status()
	if err != nil {
		return pr, false, err
	}
	ready, err := turn.line.Readiness()
	if err != nil {
		return pr, false, err
	}
	// A waiter can refresh while our own gates run. Evaluate the snapshot
	// against a clock sampled after reading it, so a new vouch isn't future.
	now = landingNow()
	b := landingline.Board(landingline.Car{Seq: tk.Seq, Issue: tk.Issue, Branch: branch, Head: head, PR: pr.Number}, self.Alone, waiting, ready, grants, s, now)
	cars := b.Cars[:1]
	for _, car := range b.Cars[1:] {
		if why := carStillAtHead(ctx, r, gh, t, car); why != "" {
			fmt.Fprintf(stderr, "landing train: #%d does not board: %s\n", car.Issue, why)
			b.LeftReady++
			continue
		}
		cars = append(cars, car)
	}
	if len(cars) == 1 {
		return serial("no other car is ready", b.LeftReady)
	}
	git := func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }
	reclaimTrains(ctx, gh, git, t.Remote, t.Repo, f.withLock, stderr)
	c := &conductor{ctx: ctx, r: r, gh: gh, ci: ci, t: t, f: f, line: turn.line, log: turn.log, stderr: stderr, sched: s, id: tk.Seq, self: tk.Seq, handle: turn.handle, departed: now,
		proofs: boardedProofs(tk.Seq, self, ready)}
	c.local = localRunDeps{sysvObserver: nativeSysVObserver(), GH: ci, Repo: t.Repo, Root: t.Root, Git: git, CleanupGit: localCheckCleanupGit(ctx, r, t.Root), Stderr: stderr, FlakeRegistry: gateFlakeRegistry(t.Root)}
	printTrainMembership(stderr, c.id, "candidates", cars)
	if err := conductTrain(c, cars, b.LeftReady); err != nil {
		return pr, false, err
	}
	prs, err := gh.LandingPRs(ctx, t.Repo, branch)
	if err != nil {
		return pr, false, err
	}
	prs, err = awaitMergedByTrain(ctx, gh, t, branch, prs, stderr)
	if err != nil {
		return pr, false, err
	}
	for _, p := range prs {
		if p.Number == pr.Number && p.State == "MERGED" {
			return p, true, nil
		}
	}
	return pr, false, fmt.Errorf("landing train %d merged, but PR #%d is not shown merged; re-run sdlc merge --branch %s --yes to finish", c.id, pr.Number, branch)
}

// carStillAtHead re-reads a boarding car's remote tip and open PR: a vouch is
// a claim from another process, and only its exact head may board.
func carStillAtHead(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, car landingline.Car) string {
	tip, err := remoteRefTip(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, "refs/heads/"+car.Branch)
	if err != nil {
		return "its pushed tip could not be read"
	}
	if tip != car.Head {
		return fmt.Sprintf("its branch moved to %s since it vouched %s", shortSHA(tip), shortSHA(car.Head))
	}
	prs, err := gh.LandingPRs(ctx, t.Repo, car.Branch)
	if err != nil {
		return "its PRs could not be listed"
	}
	for _, p := range prs {
		if p.Number == car.PR && p.State == "OPEN" && p.HeadOID == car.Head {
			return ""
		}
	}
	return fmt.Sprintf("PR #%d is not open at %s", car.PR, shortSHA(car.Head))
}

// awaitMergedByTrain re-lists the branch's PRs until one shows MERGED, up to
// trainCarWait: GitHub marks a car merged by reachability shortly after the
// train's merge, and a lagging listing is not an open PR.
func awaitMergedByTrain(ctx context.Context, gh landingGH, t landingTarget, branch string, prs []landingPR, stderr io.Writer) ([]landingPR, error) {
	deadline := time.Now().Add(trainCarWait)
	for {
		for _, p := range prs {
			if p.State == "MERGED" {
				return prs, nil
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("the train landed this branch, but GitHub does not show its PR merged after %v; re-run sdlc merge --branch %s --yes", trainCarWait, branch)
		}
		fmt.Fprintf(stderr, "landing train: waiting for GitHub to show %s's PR merged\n", branch)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(trainCarPoll):
		}
		var err error
		if prs, err = gh.LandingPRs(ctx, t.Repo, branch); err != nil {
			return nil, err
		}
	}
}
