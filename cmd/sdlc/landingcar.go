package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// A car on the landing train (#449): a waiting `sdlc merge` vouches, while it
// waits, that its own gates hold on its exact head, so a conductor can board
// it without re-running another slot's gates. The gates are today's: the
// reviewed anchor unchanged (publish gate), a recorded branch suite and green
// CI on the PR head (the CI gate, without running anything).

// carVouchEvery is how often a waiting car re-checks its gates; tests shorten it.
var carVouchEvery = time.Minute

// localGatesLabel is the operator's verification of a local-gates change. A
// PR carrying it rides alone: on a train PR the guard would see its paths
// without the label and fail the whole train.
const localGatesLabel = "local-gates-verified"

// carGates are the checks a car vouches with; tests replace them.
type carGates struct {
	publish func(ctx context.Context, pr landingPR) error
	ci      func(ctx context.Context, pr landingPR) error
}

// carReadiness reads the car's head, PR and gates now. Any read that fails is
// "not ready" with the reason, never ready: an unread gate is not a pass.
func carReadiness(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch string, g carGates, now time.Time) landingline.Readiness {
	no := func(format string, a ...any) landingline.Readiness {
		return landingline.Readiness{Branch: branch, Reason: fmt.Sprintf(format, a...), At: now}
	}
	head, err := landingGit(r, t.Root, "rev-parse", "--verify", "refs/heads/"+branch)
	if err != nil || !landingOIDValid(head) {
		return no("branch %s has no local head", branch)
	}
	tip, err := remoteRefTip(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, "refs/heads/"+branch)
	if err != nil {
		return no("could not read %s's pushed tip: %v", branch, err)
	}
	if tip != head {
		return no("%s is not pushed at its head (local %s, %s %s)", branch, shortSHA(head), t.Remote, shortSHA(tip))
	}
	prs, err := gh.LandingPRs(ctx, t.Repo, branch)
	if err != nil {
		return no("could not list PRs: %v", err)
	}
	i := slices.IndexFunc(prs, func(p landingPR) bool { return p.State == "OPEN" && p.HeadOID == head })
	if i < 0 {
		return no("no open PR at %s", shortSHA(head))
	}
	pr := prs[i]
	if err := g.publish(ctx, pr); err != nil {
		return no("publish gate: %s", firstLine(err.Error()))
	}
	if err := g.ci(ctx, pr); err != nil {
		return no("CI gate: %s", firstLine(err.Error()))
	}
	alone := pr.LocalGatesVerified
	reason := ""
	if alone {
		reason = "touches local-gates paths (" + localGatesLabel + ")"
	}
	return landingline.Readiness{Branch: branch, Head: head, PR: pr.Number, Ready: true, Alone: alone, Reason: reason, At: now}
}

// productionCarGates are the gates sdlc merge itself runs after admission,
// run here read-only: the CI gate is given no local-check runner, so a
// missing suite record is "not ready", not a suite run.
func productionCarGates(r gitRunner, t landingTarget, issuesDir string) carGates {
	git := func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }
	return carGates{
		publish: func(ctx context.Context, pr landingPR) error {
			return runLandingPublishGate(ctx, pr, issuesDir, io.Discard)
		},
		ci: func(ctx context.Context, pr landingPR) error {
			return runPreparedCIGate(ctx, ghClient, t.Repo, pr.Number, t.Remote+"/main", pr.HeadOID, git, nil, false, io.Discard, nil)
		},
	}
}

// vouchWhileWaiting re-vouches every carVouchEvery until ctx ends, which the
// caller does as soon as Wait returns: the goroutine never outlives the wait.
// It vouches only while trains are on, so a merge that waits with them off
// spends one config read a minute, and one turned on meanwhile starts
// vouching. A failed write is reported once; the car then doesn't board.
func vouchWhileWaiting(ctx context.Context, line *landingline.Line, h *landingline.Handle, sched func() landingline.Schedule, readiness func(context.Context) landingline.Readiness, stderr io.Writer) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		warned := false
		last := ""
		for {
			if !sched().Off() {
				r := readiness(ctx)
				if ctx.Err() != nil {
					return
				}
				err := line.Vouch(h, r)
				if errors.Is(err, landingline.ErrNotWaiting) {
					return // admitted or released meanwhile: nothing to vouch for
				}
				if err != nil && !warned && ctx.Err() == nil {
					warned = true
					fmt.Fprintf(stderr, "landing train: could not record readiness (%v); this merge boards no train until it can\n", err)
				}
				if state := readinessLine(r); state != last {
					last = state
					fmt.Fprintln(stderr, "landing train: "+state)
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(carVouchEvery):
			}
		}
	}()
	return done
}

func readinessLine(r landingline.Readiness) string {
	switch {
	case !r.Ready:
		return "not ready: " + r.Reason
	case r.Alone:
		return fmt.Sprintf("ready at %s (PR #%d); rides alone: %s", shortSHA(r.Head), r.PR, r.Reason)
	default:
		return fmt.Sprintf("ready at %s (PR #%d)", shortSHA(r.Head), r.PR)
	}
}

// trainAdmission is the line's admission rule (#449), built once per
// admission attempt so the schedule and the events log are read once, not
// once per ticket. Trains off: every ticket is admissible, the serial line. A
// granted merge departs at once and alone, as before #449, and runs its own
// gates after admission. Any other merge departs only when the train is due
// and its vouch holds for its own ticket: a merge whose gates don't hold
// never holds the line.
func trainAdmission(sched func() landingline.Schedule, events func() []landingline.Event, now func() time.Time) func() func(landingline.Ticket, *landingline.Readiness, bool) (bool, string) {
	return func() func(landingline.Ticket, *landingline.Readiness, bool) (bool, string) {
		s := sched()
		if s.Off() {
			return func(landingline.Ticket, *landingline.Readiness, bool) (bool, string) { return true, "" }
		}
		at := now()
		due, notDue := landingline.DepartureDue(at, s, events())
		return func(t landingline.Ticket, r *landingline.Readiness, granted bool) (bool, string) {
			switch {
			case granted:
				return true, ""
			case !due:
				return false, notDue
			case r == nil:
				return false, "train due; not vouched ready yet"
			}
			if why := r.For(t, at); why != "" {
				return false, "train due; " + why
			}
			return true, ""
		}
	}
}
