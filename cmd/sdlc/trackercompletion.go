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
	"path/filepath"
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
		if present, err := env.gitTest("rev-parse", "-q", "--verify", b.EvidenceCommit+"^{commit}"); err != nil {
			return nil, err
		} else if !present {
			continue
		}
		if inHead, err := env.gitTest("merge-base", "--is-ancestor", b.EvidenceCommit, head); err != nil {
			return nil, err
		} else if !inHead {
			continue
		}
		if base != "" {
			if onBase, err := env.gitTest("merge-base", "--is-ancestor", b.EvidenceCommit, base); err != nil {
				return nil, err
			} else if onBase {
				continue
			}
		}
		owned = append(owned, ownedCompletion{ID: rec.ID, Card: *rec.Card, Binding: b, DetailPath: rec.DetailPath})
	}
	return owned, nil
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
		return nil
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
