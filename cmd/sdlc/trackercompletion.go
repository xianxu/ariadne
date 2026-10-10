// trackercompletion.go — landing for tracker-era issues (#252). A publish owns
// the issues whose card completion binding names this repository and whose
// evidence commit it carries; the reviewed state is checked against that
// evidence commit; after the landing is confirmed the card goes done for the
// same close generation; an interrupted publish is settled by re-derivation.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// publishIssue is one issue a publish completes, with the anchor its reviewed
// state is checked against (the close commit).
type publishIssue struct{ Path, Anchor string }

// ownedCompletion is a codecomplete card a publish owns through its binding.
type ownedCompletion struct {
	ID         string
	Card       tracker.Record
	Binding    issue.Completion
	DetailPath string // absolute; "" when this checkout has no details for it
}

// ownedCompletions selects codecomplete cards bound to this repository whose
// evidence commit head contains and base (when given) does not. withDone also
// selects cards already done for such a close — an archive that runs after the
// done write must still find what it owns. A missing evidence object is simply
// not this branch's; any other Git failure refuses.
func ownedCompletions(env *trackerEnv, rs tracker.Records, head, base string, withDone bool) ([]ownedCompletion, error) {
	var owned []ownedCompletion
	for _, rec := range rs.All() {
		if rec.CardErr != nil && !rec.Duplicate {
			// It may hold a completion this landing owns (#288): refuse, as
			// for a malformed completion binding, rather than skip it.
			return nil, fmt.Errorf("card #%s: %w", rec.ID, rec.CardErr)
		}
		if rec.Duplicate || rec.Card == nil {
			continue
		}
		if status := rec.Status(); status != "codecomplete" && !(withDone && status == "done") {
			continue
		}
		b, ok, err := issue.CardCompletion(rec.Card.Raw)
		if err != nil {
			return nil, fmt.Errorf("card #%s: %w", rec.ID, err)
		}
		if !ok || b.Repository != env.target.Repository {
			continue
		}
		inHead, err := evidenceInHead(env, b.EvidenceCommit, head)
		if err != nil {
			return nil, err
		}
		if inHead && base != "" {
			if onBase, err := env.gitTest("merge-base", "--is-ancestor", b.EvidenceCommit, base); err != nil {
				return nil, err
			} else if onBase {
				continue
			}
		}
		if !inHead {
			// #304 D8: a rebase or a merge rewrites the evidence commit but never its
			// message, so the close is still this head's when a commit in range carries
			// the card's CURRENT binding token. A reopen-and-close mints a new token, so
			// an older close's commits never claim a newer one (#283/#301).
			carried, err := closeTokenCarried(env, b.Token, rec.Card.Path, head, base)
			if err != nil {
				return nil, err
			}
			if !carried {
				continue
			}
		}
		owned = append(owned, ownedCompletion{ID: rec.ID, Card: *rec.Card, Binding: b, DetailPath: rec.DetailPath})
	}
	return owned, nil
}

// evidenceInHead reports whether the evidence commit is present and an ancestor of
// head. A missing object is simply not in head.
func evidenceInHead(env *trackerEnv, evidence, head string) (bool, error) {
	if present, err := env.gitTest("rev-parse", "-q", "--verify", evidence+"^{commit}"); err != nil || !present {
		return false, err
	}
	return env.gitTest("merge-base", "--is-ancestor", evidence, head)
}

// closeTokenCarried reports whether a commit carrying `Close-Token: <token>` is in
// head beyond base — merge-base(base, head)..head when a base is given (a publish, a
// PR landing). Settling against main has no base, so the scan is bounded by time: from
// a day before the tracker commit that introduced the token (the card has no close
// date, and `updated` moves on later setter edits).
func closeTokenCarried(env *trackerEnv, token, cardPath, head, base string) (bool, error) {
	if token == "" {
		return false, nil
	}
	var rangeArgs []string
	if base != "" {
		mb, err := env.git("merge-base", base, head)
		if err != nil {
			return false, err
		}
		rangeArgs = []string{mb + ".." + head}
	} else {
		since, err := closeTokenSince(env, token, cardPath)
		if err != nil || since == "" {
			return false, err
		}
		rangeArgs = []string{head, "--since=" + since}
	}
	shas, err := gitx.CommitsWithCloseToken(env.root, rangeArgs, token)
	return len(shas) > 0, err
}

// closeTokenSince is a day before the tracker commit that introduced token on the card,
// or "" when no such commit is reachable (nothing to scan from).
func closeTokenSince(env *trackerEnv, token, cardPath string) (string, error) {
	// --reverse lists oldest first (a -1 limit would apply BEFORE the reversal and
	// return the newest), so the first line is the commit that introduced it.
	when, err := env.git("log", "--reverse", "-S"+token, "--format=%cI", env.repo.TrackingRef(), "--", cardPath)
	if err != nil || when == "" {
		return "", err
	}
	t, err := time.Parse(time.RFC3339, strings.Fields(when)[0])
	if err != nil {
		return "", fmt.Errorf("read the close time of %s: %w", token, err)
	}
	return t.Add(-24 * time.Hour).Format(time.RFC3339), nil
}

// doneCard is the pure landing change: codecomplete → done for the same close
// generation, recording where it landed. A card already done for this close is
// ErrNoChange (a finished retry); another generation or status refuses.
func doneCard(current []byte, closeToken, landed, today string) ([]byte, error) {
	fm, _, err := issue.Parse(string(current))
	if err != nil {
		return nil, err
	}
	b, ok, err := issue.CardCompletion(current)
	if err != nil {
		return nil, err
	}
	if !ok || b.Token != closeToken {
		return nil, errors.New("the card was closed again (another generation) — this landing's close is stale; re-run close")
	}
	switch status, _ := issue.GetField(fm, "status"); status {
	case "done":
		if b.LandedCommit != "" {
			return nil, tracker.ErrNoChange
		}
	case "codecomplete":
	default:
		return nil, fmt.Errorf("the card is %s (reopened after its close) — refusing to complete it", status)
	}
	b.LandedCommit = landed
	next, err := issue.SetCardCompletion(current, b)
	if err == nil {
		next, err = issue.SetCardField(next, "updated", today)
	}
	if err == nil {
		next, err = issue.SetCardField(next, "status", "done")
	}
	return next, err
}

// completeOnCard publishes done for one owned completion.
func completeOnCard(env *trackerEnv, oc ownedCompletion, landed string) error {
	today := time.Now().Format("2006-01-02")
	err := env.repo.ChangeCard(oc.ID, oc.Card.Path, "done", operationToken("done"), func(current []byte) ([]byte, error) {
		return doneCard(current, oc.Binding.Token, landed, today)
	})
	invalidateIssueRecords(env.ctx) // the archive that follows must see done
	if errors.Is(err, tracker.ErrNoChange) {
		err = nil
	}
	if err == nil {
		// #304 D4: the issue is done; its reviewed-head pins end. Best effort —
		// a pin is retention, and the settle sweep catches a miss.
		_ = unpinReviewed(env.root, oc.ID)
	}
	return err
}

// settleLandedCompletions finishes interrupted publishes by re-derivation: a
// codecomplete card bound to this repository whose evidence commit is already
// on fresh main landed, so it goes done. Idempotent; used by every publishing
// verb before it selects its own issues, and by recovery reconcile.
func settleLandedCompletions(ctx context.Context, env *trackerEnv, issuesDir string) ([]string, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil || !rs.Tracker {
		return nil, err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return nil, err
	}
	landed, err := ownedCompletions(env, rs, view.Ref(), "", false)
	if err != nil {
		return nil, err
	}
	var settled []string
	for _, oc := range landed {
		if err := completeOnCard(env, oc, view.Ref()); err != nil {
			return settled, fmt.Errorf("complete #%s (its close is on main): %w", issue.CLIRef(oc.ID), err)
		}
		settled = append(settled, oc.ID)
	}
	// #304 D4: also end the pins of issues landed and settled elsewhere, or archived by
	// hand — the per-site unpins only run inside sdlc verbs here. (Recovery reconcile
	// runs this settle, so it sweeps too.)
	_ = sweepTrackedPins(env, rs)
	return settled, nil
}

// completeLandingPR completes, in the tracked repository at root, the closes a
// confirmed PR landing owns: their evidence is in the PR (HeadOID --not
// BaseOID), and the landed commit is the PR's integrated merge — whatever the
// strategy, since a squash or rebase leaves the evidence commit off main. It
// also settles earlier landings that did keep their evidence (idempotent).
func completeLandingPR(ctx context.Context, root, issuesDir string, pr landingPR) error {
	dir := filepath.Join(root, filepath.FromSlash(issuesDir))
	rs, err := loadIssueRecords(ctx, dir, tracker.Fresh)
	if err != nil || !rs.Tracker {
		return err
	}
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return err
	}
	owned, err := ownedCompletions(env, rs, pr.HeadOID, pr.BaseOID, false)
	if err != nil {
		return err
	}
	for _, oc := range owned {
		if err := completeOnCard(env, oc, pr.MergeOID); err != nil {
			return fmt.Errorf("complete #%s landed by PR #%d: %w", issue.CLIRef(oc.ID), pr.Number, err)
		}
	}
	_, err = settleLandedCompletions(ctx, env, dir)
	return err
}

// branchOwnedCompletions selects, in a tracked repository, the closes the
// current branch carries beyond main (nil env without a tracker).
func branchOwnedCompletions(ctx context.Context, issuesDir string) ([]ownedCompletion, *trackerEnv, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil || !rs.Tracker {
		return nil, nil, err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return nil, nil, err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return nil, nil, err
	}
	owned, err := ownedCompletions(env, rs, "HEAD", view.Ref(), false)
	return owned, env, err
}
