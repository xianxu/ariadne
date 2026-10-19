package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
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
	// CloseTrainPR closes the PR and deletes its branch.
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
	t      landingTarget
	f      *mergeFlags
	line   *landingline.Line
	log    *lineLog
	stderr io.Writer
	sched  landingline.Schedule
	id     int64
	self   int64 // the conductor's own ticket: its car is the first

	departed time.Time
	wt       string   // the train's detached worktree
	base     string   // the main the train is built on
	heads    []string // heads[k-1] is prefix k's commit
	pushed   string   // what the train branch points at on the remote
	pr       landingPR
	opened   bool
	local    localRunDeps
	run      localCheckRunner
	suite    localCheck
	baseline *bool // main's baseline verdict, run once on the first failure
	settled  map[int64]bool
	selfEnd  landingline.OutcomeKind
	selfWhy  string
}

// trainGit runs git in the repository's checkout.
func (c *conductor) git(args ...string) (string, error) { return landingGit(c.r, c.t.Root, args...) }

// wgit runs git in the train's worktree.
func (c *conductor) wgit(args ...string) (string, error) { return landingGit(c.r, c.wt, args...) }

// conductTrain runs one train of cars (the conductor's own first) to a
// terminal phase. It returns nil when the conductor's own car landed; any
// other end of its own car is an error naming why. Whatever happens, every
// other car gets an outcome and the train PR does not outlive the train.
func conductTrain(c *conductor, cars []landingline.Car, leftReady int) (err error) {
	c.settled = map[int64]bool{}
	tr := landingline.NewTrain(c.id)
	defer func() {
		// An error mid-train: every car without an outcome goes back to
		// wait (still ready), and the train PR is closed.
		if err != nil && !tr.Phase.Terminal() {
			for _, car := range cars {
				c.settle(car, landingline.OutcomeReleased, "the train stopped: "+firstLine(err.Error()))
			}
			c.closePR()
			tr.Phase = landingline.PhaseAbandoned
		}
		c.removeWorktree()
		c.recordTrain(tr, cars, leftReady)
	}()
	tr, effects, err := landingline.Transition(tr, landingline.Boarded{Cars: cars})
	if err != nil {
		return err
	}
	for _, car := range cars[1:] {
		c.settle(car, landingline.OutcomeBoarded, fmt.Sprintf("train %d, car %d of %d", c.id, indexOf(cars, car.Seq)+1, len(cars)))
	}
	for len(effects) > 0 {
		e := effects[0]
		effects = effects[1:]
		ev, err := c.execute(e, cars)
		if err != nil {
			return err
		}
		if ev == nil {
			continue
		}
		var more []landingline.Effect
		if tr, more, err = landingline.Transition(tr, ev); err != nil {
			return err
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

// execute runs one effect and returns the event its outcome is, if any.
func (c *conductor) execute(e landingline.Effect, cars []landingline.Car) (landingline.TrainEvent, error) {
	switch e := e.(type) {
	case landingline.Build:
		return c.build(e.Cars)
	case landingline.Publish:
		return nil, c.publish(e.Prefix)
	case landingline.RunSuite:
		return c.runSuite(e.Prefix)
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
	main, err := c.t.fetchMain(c.r)
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
	if _, err := c.git("worktree", "add", "--detach", dir, main); err != nil {
		return nil, err
	}
	c.wt = dir
	// The suite resolves its lock helper (Makefile.workflow) from the tree it
	// runs for: the train's, which carries main's, not this checkout's branch.
	c.local.Root = dir
	if c.suite.Name != "" {
		c.run = newLocalCheckRunner(c.local)
	}
	built := landingline.Built{Conflicts: map[int64]string{}}
	for _, car := range cars {
		ref := fmt.Sprintf("refs/sdlc/train/%d/%d", c.id, car.Seq)
		if _, err := c.git("fetch", "--quiet", c.t.Remote, "+refs/heads/"+car.Branch+":"+ref); err != nil {
			built.Conflicts[car.Seq] = "its branch could not be fetched: " + firstLine(err.Error())
			continue
		}
		fetched, err := c.git("rev-parse", "--verify", ref+"^{commit}")
		_, _ = c.git("update-ref", "-d", ref)
		if err != nil || fetched != car.Head {
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
	}
	return built, nil
}

// gatePrefix runs the instance and duplicate-ID gates on one prefix, as a
// single landing runs them on its integrated head; "" is a pass.
func (c *conductor) gatePrefix(head string) string {
	if c.f.NoValidate {
		return ""
	}
	if err := runLandingInstanceGate(c.base, head, nounGates(c.f.IssuesDir), io.Discard, c.stderr, c.r); err != nil {
		return "the instance gate refuses the train with it: " + firstLine(err.Error())
	}
	if err := runLandingDuplicateGate(c.base, head, c.f.IssuesDir, c.f.HistoryDir, c.r); err != nil {
		return "the duplicate-ID gate refuses the train with it: " + firstLine(err.Error())
	}
	return ""
}

// publish points the train branch at prefix k and the train PR at it,
// opening the PR the first time, and waits for GitHub to register the head.
func (c *conductor) publish(k int) error {
	head := c.heads[k-1]
	branch := trainBranch(c.id)
	if _, err := c.git("push", "--quiet", "--force-with-lease=refs/heads/"+branch+":"+c.pushed, c.t.Remote, head+":refs/heads/"+branch); err != nil {
		return fmt.Errorf("push the train branch: %w", err)
	}
	c.pushed = head
	if !c.opened {
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
				return nil
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

// runSuite runs the recorded suite on prefix k. A failure of the full train
// is compared once with main's baseline: main failing the same way is no
// car's fault.
func (c *conductor) runSuite(k int) (landingline.TrainEvent, error) {
	head := c.heads[k-1]
	rec, err := c.run(c.ctx, c.suite, head)
	if err != nil {
		return nil, fmt.Errorf("the train suite on %s could not run: %w", shortSHA(head), err)
	}
	done := landingline.SuiteDone{Prefix: k, Outcome: landingline.SuitePass}
	if rec.passed() {
		return done, nil
	}
	done.Outcome = landingline.SuiteFail
	if c.baseline == nil {
		red := c.mainIsRed(rec)
		c.baseline = &red
	}
	if *c.baseline {
		done.Outcome = landingline.SuiteMainRed
	}
	return done, nil
}

// mainIsRed runs main's baseline, unrecorded, and reports whether it fails
// every test the train failed: then no car introduced the failure.
func (c *conductor) mainIsRed(failed localRunRecord) bool {
	mergedLog, err := os.ReadFile(failed.LogPath)
	if err != nil {
		return false
	}
	deps := c.local
	deps.SkipRecord, deps.PR = true, 0
	baseline, err := runLocalCheck(c.ctx, deps, c.suite, c.base)
	if err != nil || baseline.passed() {
		return false
	}
	baseLog, err := os.ReadFile(baseline.LogPath)
	if err != nil {
		return false
	}
	merged, base := failingTests(string(mergedLog)), failingTests(string(baseLog))
	if len(base) == 0 {
		return true // main is red and unparsed: no car can be blamed
	}
	introduced, _ := classifyFailures(merged, base)
	return len(introduced) == 0
}

// waitCI waits for the train PR's CI on prefix k, the gate a single landing
// waits for; any refusal is the train's CI being red.
func (c *conductor) waitCI(k int) landingline.TrainEvent {
	head := c.heads[k-1]
	ioGit := func(ioCtx context.Context) func(...string) (string, error) {
		rr := withLandingGitContext(ioCtx, c.r)
		return func(args ...string) (string, error) { return landingGit(rr, c.t.Root, args...) }
	}
	if err := waitCIGate(c.ctx, ghClient, c.t.Repo, c.pr.Number, c.base, head, c.git, c.run, c.f.NoCI, c.stderr, ioGit); err != nil {
		return landingline.CIDone{Detail: firstLine(err.Error())}
	}
	return landingline.CIDone{Green: true}
}

// merge lands prefix k: main is fetched first, and a main that moved is
// reintegrated rather than merged past. The merge is pinned to the head and
// confirmed on GitHub; an unconfirmed one is probed, never assumed.
func (c *conductor) merge(k int) (landingline.TrainEvent, error) {
	head := c.heads[k-1]
	main, err := c.t.fetchMain(c.r)
	if err != nil {
		return nil, err
	}
	if main != c.base {
		return landingline.MainMoved{}, nil
	}
	if err := c.gh.LandingMerge(c.ctx, c.t.Repo, c.pr.Number, head); err != nil {
		fmt.Fprintf(c.stderr, "landing train %d: merge outcome uncertain (%v); reading GitHub\n", c.id, err)
		return landingline.MergeUnknown{}, nil
	}
	ev, err := c.probe()
	if err != nil {
		return nil, err
	}
	if p := ev.(landingline.Probed); p.Merged {
		return landingline.Merged{OID: p.OID}, nil
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
			return landingline.Probed{Merged: true, OID: p.MergeOID}, nil
		}
	}
	return landingline.Probed{}, nil
}

// reintegrate merges current main into prefix k's head. The suite carries
// over only when what main brought has no code surface (the publish gate's
// line, as a single landing's docs-only reuse).
func (c *conductor) reintegrate(k int) (landingline.TrainEvent, error) {
	main, err := c.t.fetchMain(c.r)
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
	return landingline.Reintegrated{NeedSuite: publishGateHasCodeSurface(splitNUL(diff))}, nil
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
		return
	}
	err := c.line.Settle(car.Seq, landingline.Outcome{Train: c.id, Kind: kind, Reason: why, PR: c.pr.Number})
	if err != nil && !errors.Is(err, landingline.ErrGone) {
		fmt.Fprintf(c.stderr, "landing train %d: could not tell car #%d it was %s (%v); its sdlc merge resumes from GitHub's state\n", c.id, car.Issue, kind, err)
	}
}

// closePR closes the train PR and deletes its branch, once.
func (c *conductor) closePR() {
	if !c.opened {
		if c.pushed != "" {
			_, _ = c.git("push", "--quiet", c.t.Remote, ":refs/heads/"+trainBranch(c.id))
		}
		return
	}
	c.opened = false
	if err := c.gh.CloseTrainPR(c.ctx, c.t.Repo, c.pr.Number, trainBranch(c.id)); err != nil {
		fmt.Fprintf(c.stderr, "landing train %d: could not close train PR #%d (%v); the next train closes it\n", c.id, c.pr.Number, err)
	}
}

func (c *conductor) removeWorktree() {
	if c.wt == "" {
		return
	}
	if _, err := landingGit(recoveryLandingGit(c.r), c.t.Root, "worktree", "remove", "--force", c.wt); err != nil {
		os.RemoveAll(c.wt)
		_, _ = landingGit(recoveryLandingGit(c.r), c.t.Root, "worktree", "prune")
	}
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
		NextDue: landingline.NextDue(c.departed, c.sched, leftReady > 0 || len(tr.Released) > 0)})
	fmt.Fprintf(c.stderr, "landing train %d: %s; %d car(s), %d split suite(s)\n", c.id, tr.Phase, len(cars), tr.Splits)
}

// reclaimTrains closes train PRs a dead conductor left open: head.lock was
// free when this one was admitted, so no live conductor owns them.
func reclaimTrains(ctx context.Context, gh trainGH, repo string, stderr io.Writer) {
	prs, err := gh.OpenTrainPRs(ctx, repo)
	if err != nil {
		fmt.Fprintf(stderr, "landing train: could not list leftover train PRs (%v)\n", err)
		return
	}
	for _, p := range prs {
		if err := gh.CloseTrainPR(ctx, repo, p.Number, p.HeadRef); err != nil {
			fmt.Fprintf(stderr, "landing train: could not close leftover train PR #%d (%v)\n", p.Number, err)
			continue
		}
		fmt.Fprintf(stderr, "landing train: closed leftover train PR #%d (%s): its conductor is gone\n", p.Number, p.HeadRef)
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
func maybeConductTrain(ctx context.Context, r gitRunner, t landingTarget, f *mergeFlags, turn *landingTurn, branch, head string, pr landingPR, s landingline.Schedule, readiness func(context.Context) landingline.Readiness, stderr io.Writer) (landingPR, bool, error) {
	if s.Off() {
		return pr, false, nil
	}
	now := landingNow()
	tk := turn.handle.Ticket()
	serial := func(why string, leftReady int) (landingPR, bool, error) {
		turn.log.record(landingline.Event{Kind: landingline.EventTrain, Seq: tk.Seq, Train: tk.Seq, Cars: []int64{tk.Seq}, Result: "alone",
			Detail: why, NextDue: landingline.NextDue(now, s, leftReady > 0)})
		return pr, false, nil
	}
	gh, ok := ghClient.(trainGH)
	if !ok {
		return serial("the GitHub adapter has no train support", 0)
	}
	self := readiness(ctx)
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
	b := landingline.Board(landingline.Car{Seq: tk.Seq, Issue: tk.Issue, Branch: branch, Head: head, PR: pr.Number}, self.Alone, waiting, ready, s, now)
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
	reclaimTrains(ctx, gh, t.Repo, stderr)
	git := func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }
	mainRef := t.Remote + "/main"
	decl, err := readLocalChecks(git, mainRef)
	if err != nil {
		return pr, false, err
	}
	c := &conductor{ctx: ctx, r: r, gh: gh, t: t, f: f, line: turn.line, log: turn.log, stderr: stderr, sched: s, id: tk.Seq, self: tk.Seq, departed: now}
	c.local = localRunDeps{GH: ghClient, Repo: t.Repo, Root: t.Root, Git: git, CleanupGit: localCheckCleanupGit(ctx, r, t.Root), Stderr: stderr, FlakeRegistry: gateFlakeRegistry(t.Root)}
	c.run = newLocalCheckRunner(c.local)
	if suite, ok := decl.check(suiteCheckName); ok {
		c.suite = suite
	} else {
		c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return localRunRecord{Exit: 0}, nil }
	}
	fmt.Fprintf(stderr, "landing train %d: %d cars:", c.id, len(cars))
	for _, car := range cars {
		fmt.Fprintf(stderr, " #%d (PR #%d)", car.Issue, car.PR)
	}
	fmt.Fprintln(stderr)
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
