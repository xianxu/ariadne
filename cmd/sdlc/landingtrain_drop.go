package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// Only a positively completed drop cleanup permits this internal retry. The
// outer attempt has released its line and queue operation before it is retried.
type trainReformReleaseKey struct{}

type trainReformError struct {
	train int64
	turn  *landingTurn
}

func (e *trainReformError) Error() string {
	return fmt.Sprintf("train %d dropped a car; survivors re-form with a fresh landing ticket", e.train)
}

func (c *conductor) beginTrainControl(cars []landingline.Car) error {
	_, waiting, err := c.line.Status()
	if err != nil {
		return err
	}
	roots := map[int64]string{}
	for _, t := range waiting {
		roots[t.Seq] = t.QueueRoot
	}
	roots[c.self] = c.t.Root
	members := make([]landingline.TrainMember, 0, len(cars))
	for _, car := range cars {
		members = append(members, landingline.TrainMember{Seq: car.Seq, Issue: car.Issue, Branch: car.Branch, Head: car.Head, PR: car.PR, OwnerRoot: roots[car.Seq]})
	}
	return c.line.BeginTrain(c.handle, landingline.TrainControl{Version: landingline.TrainControlVersion, Train: c.id, Phase: landingline.PhaseForming, Members: members,
		Publication: landingline.TrainPublication{Repository: c.t.Repo, Remote: c.t.Remote, Branch: trainBranch(c.id), RecoveryBranch: cars[0].Branch}})
}

// The effect runs synchronously. Cancellation joins its subprocesses before
// execute returns; stopping this watcher also restores a live cleanup context.
func (c *conductor) watchTrainDrops() func() {
	parent, runner := c.ctx, c.r
	ctx, cancel := context.WithCancelCause(parent)
	c.ctx = ctx
	c.r = withLandingGitContext(ctx, runner)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(landingLinePoll)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			control, err := c.line.ReadTrainControl()
			if err != nil {
				cancel(fmt.Errorf("train control unavailable: %w", err))
				return
			}
			if control.Train != c.id {
				cancel(landingline.ErrTrainControlStale)
				return
			}
			if len(control.CurrentDrops()) > 0 {
				cancel(landingline.ErrTrainDropPending)
				return
			}
		}
	}()
	var once sync.Once
	return func() { once.Do(func() { cancel(context.Canceled); <-done; c.ctx = parent; c.r = runner }) }
}

func (c *conductor) trainDropsPending() (bool, error) {
	control, err := c.line.ReadTrainControl()
	if err != nil {
		return false, err
	}
	if control.Train != c.id {
		return false, landingline.ErrTrainControlStale
	}
	return len(control.CurrentDrops()) > 0, nil
}

func (c *conductor) applyTrainDrops(tr *landingline.Train, cars []landingline.Car, stop func()) error {
	stop() // effect and watcher are joined before any abort effects or head release
	next, effects, control, err := c.line.ApplyTrainDrops(c.handle, *tr)
	if err != nil {
		return err
	}
	*tr = next
	for _, effect := range effects {
		if _, closing := effect.(landingline.ClosePR); closing {
			if err := closeDroppedTrain(c.ctx, c.gh, c.git, c.locked, control.Cleanup); err != nil {
				return dropCleanupError(control, err)
			}
			if err := c.line.FinishTrainControl(c.handle, c.id, control.Cleanup); err != nil {
				return dropCleanupError(control, err)
			}
		} else if _, err := c.execute(effect, cars); err != nil {
			return err
		}
	}
	if c.selfEnd == landingline.OutcomeHeld {
		return fmt.Errorf("landing train %d: this merge's PR was held: %s", c.id, c.selfWhy)
	}
	return &trainReformError{train: c.id}
}

func dropCleanupError(control *landingline.TrainControl, cause error) error {
	return fmt.Errorf("train %d: drop accepted; cleanup pending: %w; Recovery: sdlc merge --branch %s --yes", control.Train, cause, control.Cleanup.Publication.RecoveryBranch)
}

// The caller holds head.lock, but never line.lock, during these external reads.
// A missing listing after publication is uncertainty, not closure evidence.
func closeDroppedTrain(ctx context.Context, gh trainGH, git func(...string) (string, error), lock func(func() error) error, cleanup landingline.TrainCleanup) error {
	p := cleanup.Publication
	if !cleanup.Pending {
		return nil
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(p.Branch, trainBranchPrefix), 10, 64)
	if err != nil || id <= 0 || trainBranch(id) != p.Branch {
		return errors.New("cleanup target is not a train branch; original obligation retained")
	}
	// The lease pins content; it does not pin where a mutable remote alias
	// points. Re-read effective fetch/push destinations under the same repo
	// lock as deletion, as normal landing cleanup revalidates its target.
	checkedLock := func(run func() error) error {
		return lock(func() error {
			for _, args := range [][]string{{"remote", "get-url", "--all", p.Remote}, {"remote", "get-url", "--push", "--all", p.Remote}} {
				url, err := git(args...)
				if err != nil {
					return err
				}
				repo, err := landingRepoURL(strings.TrimSpace(url))
				if err != nil || !strings.EqualFold(repo, p.Repository) {
					return errors.New("train cleanup destination changed; original obligation retained")
				}
			}
			return run()
		})
	}
	if err := checkedLock(func() error { return nil }); err != nil {
		return err
	}
	if p.Started {
		prs, err := gh.LandingPRs(ctx, p.Repository, p.Branch)
		if err != nil {
			return err
		}
		var selected *landingPR
		for i := range prs {
			pr := &prs[i]
			if pr.HeadRef == p.Branch && (p.PR == 0 || pr.Number == p.PR) {
				if selected != nil {
					return errors.New("ambiguous old train PR identity")
				}
				selected = pr
			}
		}
		if selected == nil {
			return errors.New("published train PR not positively observed; closure unconfirmed")
		}
		if selected.State == "OPEN" {
			// A lost close reply is resolved by the following positive state read.
			_ = gh.CloseTrainPR(ctx, p.Repository, selected.Number, p.Branch)
			prs, err = gh.LandingPRs(ctx, p.Repository, p.Branch)
			if err != nil {
				return err
			}
			closed := false
			for _, pr := range prs {
				if pr.Number == selected.Number && pr.HeadRef == p.Branch && pr.State == "CLOSED" {
					closed = true
				}
			}
			if !closed {
				return errors.New("old train PR closure unconfirmed")
			}
		} else if selected.State != "CLOSED" {
			return fmt.Errorf("old train PR #%d is %s; cannot confirm abandoned", selected.Number, selected.State)
		}
	}
	tip, err := remoteRefTip(git, p.Remote, "refs/heads/"+p.Branch)
	if err != nil {
		return err
	}
	if tip != "" {
		if tip != p.ExpectedTip && tip != p.PreviousTip {
			return errors.New("old train ref moved to an unrecognized head; retained")
		}
		if err := deleteTrainRef(git, checkedLock, p.Remote, p.Branch, tip); err != nil {
			return err
		}
	}
	if err := checkedLock(func() error {
		var readErr error
		tip, readErr = remoteRefTip(git, p.Remote, "refs/heads/"+p.Branch)
		return readErr
	}); err != nil {
		return err
	}
	if tip != "" {
		return errors.New("old train ref deletion unconfirmed")
	}
	return nil
}

// Every admitted attempt, before selecting an unmerged PR or the serial/train
// path, satisfies the predecessor's durable obligation. Its generation cannot
// be replaced until positive external proof is committed under line.lock.
func reconcileDroppedTrain(ctx context.Context, turn *landingTurn, t landingTarget, r gitRunner, gh landingGH, f *mergeFlags) error {
	control, err := turn.line.ReadTrainControl()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !control.Cleanup.Pending {
		return nil
	}
	if control.Cleanup.Publication.Remote != t.Remote || !strings.EqualFold(control.Cleanup.Publication.Repository, t.Repo) {
		return dropCleanupError(control, errors.New("train cleanup destination differs from this landing target; original obligation retained"))
	}
	trainGH, ok := gh.(trainGH)
	if !ok {
		return dropCleanupError(control, errors.New("GitHub adapter lacks train cleanup support"))
	}
	git := func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }
	if err := closeDroppedTrain(ctx, trainGH, git, f.withLock, control.Cleanup); err != nil {
		return dropCleanupError(control, err)
	}
	if err := turn.line.FinishTrainControl(turn.handle, control.Train, control.Cleanup); err != nil {
		return dropCleanupError(control, err)
	}
	return nil
}

func printTrainControl(out io.Writer, line *landingline.Line) {
	control, head, err := line.TrainControlStatus()
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		fmt.Fprintf(out, "train control unavailable: %v\n", err)
		return
	}
	live := head != nil && !head.Unreadable && head.Seq == control.Train
	if live || control.Cleanup.Pending || len(control.Drops) > 0 {
		fmt.Fprintf(out, "train %d: %s", control.Train, control.Phase)
		if control.MergeStarted {
			fmt.Fprint(out, " (merge operation started)")
		}
		if !live {
			fmt.Fprint(out, " (conductor no longer holds the line)")
		}
		fmt.Fprintln(out)
	}
	for _, drop := range control.Drops {
		fmt.Fprintf(out, "  #%d held by drop from train %d: %s; explicit retry: sdlc merge --branch %s --yes\n", drop.Member.Issue, drop.Train, drop.Reason, drop.Member.Branch)
	}
	if control.Cleanup.Pending {
		fmt.Fprintf(out, "  drop accepted; cleanup/recomposition pending; Recovery: sdlc merge --branch %s --yes\n", control.Cleanup.Publication.RecoveryBranch)
	}
}

// retryTrainReformation owns the inherited head across attempts. The attempt
// seam lets the resource-lifetime paths be exercised against actual line locks.
func retryTrainReformation(ctx context.Context, out io.Writer, attempt func(context.Context) error) error {
	var retained *landingTurn
	defer func() {
		if retained != nil {
			retained.leave(errors.New("survivor re-formation interrupted before fresh ticket vouch"))
		}
	}()
	for {
		attemptCtx := ctx
		if retained != nil {
			attemptCtx = context.WithValue(ctx, trainReformReleaseKey{}, func() {
				retained.leave(errors.New("train dropped a car; fresh survivor ticket vouched"))
				retained = nil
			})
		}
		err := attempt(attemptCtx)
		if retained != nil {
			retained.leave(err)
			retained = nil
		}
		var reform *trainReformError
		if !errors.As(err, &reform) {
			return err
		}
		retained = reform.turn
		if err := ctx.Err(); err != nil {
			return err
		}
		fmt.Fprintln(out, reform.Error())
	}
}
