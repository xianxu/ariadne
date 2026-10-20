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
// merged reports that the car's PR is already MERGED at its exact head (a
// train landed it, and its outcome never arrived: a conductor that died after
// the merge, a lost settle): the wait is over, and the merge resumes.
func carReadiness(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch string, g carGates, now time.Time) (ready landingline.Readiness, merged bool) {
	no := func(format string, a ...any) (landingline.Readiness, bool) {
		return landingline.Readiness{Branch: branch, Reason: fmt.Sprintf(format, a...), At: now}, false
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
	if slices.ContainsFunc(prs, func(p landingPR) bool { return p.State == "MERGED" && p.HeadOID == head }) {
		r, _ := no("its PR is merged at %s", shortSHA(head))
		return r, true
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
	return landingline.Readiness{Branch: branch, Head: head, PR: pr.Number, Ready: true, Alone: alone, Reason: reason, At: now}, false
}

// errMergedWhileWaiting ends a car's wait when its own PR turned up merged:
// the merge then finishes as a resume, like a car a train settled landed.
var errMergedWhileWaiting = errors.New("landing train: this PR was merged while it waited")

// productionCarGates are the gates sdlc merge itself runs after admission,
// run here read-only: the CI gate is given no local-check runner, so a
// missing suite record is "not ready", not a suite run. A gate the merge was
// told to skip (--no-judge, --no-ci) is skipped here too, so a car vouches
// exactly what its own landing would check.
func productionCarGates(r gitRunner, ci ghCaller, t landingTarget, f *mergeFlags) carGates {
	git := func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }
	return carGates{
		publish: func(ctx context.Context, pr landingPR) error {
			if f.NoJudge {
				return nil
			}
			return runLandingPublishGate(ctx, pr, f.IssuesDir, io.Discard)
		},
		ci: func(ctx context.Context, pr landingPR) error {
			if f.NoCI {
				return nil
			}
			base := t.Remote + "/main"
			facts, err := carCIGateFacts(ctx, git, base, pr.HeadOID)
			if err != nil {
				return err
			}
			return runPreparedCIGate(ctx, ci, t.Repo, pr.Number, base, pr.HeadOID, git, nil, false, io.Discard, facts)
		},
	}
}

// carCIGateFacts are the CI gate's facts for a car's vouch. Where main declares
// the affected suite (#473 D3), a recorded affected run stands in for suite:
// the train runs the full suite on its prefix before any car lands. A docs-only
// affected record is how a docs-only branch boards (pair#723).
func carCIGateFacts(ctx context.Context, git func(...string) (string, error), base, head string) (*ciGateFacts, error) {
	facts, err := prepareCIGateFactsContext(ctx, git, base, head)
	if err != nil {
		return nil, err
	}
	// An affected.json that cannot be read or trusted degrades to requiring
	// the suite, never to a stopped train (#473 M2): the car's refusal names
	// suite, and the affected run itself reports the file's error.
	if _, ok, err := readAffectedConfig(git, base); err == nil && ok {
		facts.standIn = map[string]string{suiteCheckName: affectedCheckName}
	}
	return facts, nil
}

// vouchWhileWaiting re-vouches every carVouchEvery until ctx ends, which the
// caller does as soon as Wait returns: the goroutine never outlives the wait.
// It vouches only while trains are on, so a merge that waits with them off
// spends one config read a minute, and one turned on meanwhile starts
// vouching. A failed write is reported once; the car then doesn't board.
//
// A car that finds its own PR merged calls merged and stops: no outcome may
// ever come for it (its conductor died after the merge), and only its own
// observation of GitHub can end the wait.
func vouchWhileWaiting(ctx context.Context, line *landingline.Line, h *landingline.Handle, sched func() landingline.Schedule, readiness func(context.Context) (landingline.Readiness, bool), merged func(), stderr io.Writer) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		warned := false
		last := ""
		for {
			if !sched().Off() {
				r, isMerged := readiness(ctx)
				if ctx.Err() != nil {
					return
				}
				if isMerged {
					fmt.Fprintln(stderr, "landing train: this PR is already merged; finishing as a resume")
					merged()
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

// trainAdmission is the line's admission rule (#449). The train setting is
// read once per admission attempt, outside the line mutex (the source caches
// it); the departure ledger comes from the line, read under the mutex after
// head.lock. Trains off: every ticket is admissible, the serial line. A
// granted merge departs at once and alone, as before #449, and runs its own
// gates after admission. Any other merge departs only when the train is due
// and its vouch holds for its own ticket: a merge whose gates don't hold
// never holds the line.
func trainAdmission(sched func() landingline.Schedule, now func() time.Time) func() landingline.AdmissionRule {
	return func() landingline.AdmissionRule {
		s := sched()
		if s.Off() {
			return func([]landingline.Event, landingline.Ticket, *landingline.Readiness, bool, landingline.Waiters) (bool, string) {
				return true, ""
			}
		}
		return func(events []landingline.Event, t landingline.Ticket, r *landingline.Readiness, granted bool, w landingline.Waiters) (bool, string) {
			if granted {
				return true, ""
			}
			at := now()
			if due, why := landingline.DepartureDue(at, s, events, w); !due {
				return false, why
			}
			if r == nil {
				return false, "train due; not vouched ready yet"
			}
			if why := r.For(t, at); why != "" {
				return false, "train due; " + why
			}
			return true, ""
		}
	}
}
