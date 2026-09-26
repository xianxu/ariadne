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
	"io"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
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
// evidence commit head contains and base (when given) does not. A missing
// evidence object is simply not this branch's; any other Git failure refuses.
func ownedCompletions(env *trackerEnv, rs tracker.Records, head, base string) ([]ownedCompletion, error) {
	var owned []ownedCompletion
	for _, rec := range rs.All() {
		if rec.Duplicate || rec.Card == nil || rec.Status() != "codecomplete" {
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
	landed, err := ownedCompletions(env, rs, view.Ref(), "")
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

// refreshDoneMirrors brings the details of terminal cards up to the card before
// they are archived, so a history file mirrors its final state. A mirror that
// cannot refresh (a hand edit) is archived as it is and reported.
func refreshDoneMirrors(ctx context.Context, stderr io.Writer, env *trackerEnv, issuesDir string) error {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil || !rs.Tracker {
		return err
	}
	for _, rec := range rs.All() {
		if rec.Duplicate || rec.Card == nil || rec.DetailPath == "" || !vocab.Issue().IsTerminal(rec.Status()) {
			continue
		}
		if !issue.HasMirror([]byte(issue.Compose(rec.DetailFM, rec.DetailBody))) {
			continue
		}
		if warn := refreshLocalMirrorAt(env, rec.DetailPath); warn != "" {
			cwarn(stderr, warn)
		}
	}
	return nil
}
