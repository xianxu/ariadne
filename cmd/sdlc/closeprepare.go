package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// closePreparationReadSet lives for one invocation, never in a saved proof.
// The initial artifact bytes fence preparation as well as reviewer execution.
type closePreparationReadSet struct {
	initial  closeReviewSnapshot
	issues   string
	issue    string
	tracked  bool
	cardPath string
	cardOID  string
	adopted  *adoptionReadSet
	ctx      context.Context
	resting  string
	target   gitx.PublicationTarget
}

func captureCloseCard(ctx context.Context, issues, id string) (bool, string, string, error) {
	records, err := loadIssueRecords(ctx, issues, tracker.Fresh)
	if err != nil {
		return false, "", "", err
	}
	return closeCardAuthority(records, id)
}

func closeCardAuthority(records tracker.Records, id string) (bool, string, string, error) {
	if !records.Tracker {
		return false, "", "", nil
	}
	record, ok, err := records.Require(id)
	if err != nil {
		return false, "", "", err
	}
	if !ok || record.Card == nil {
		return false, "", "", fmt.Errorf("close card #%s is unavailable", id)
	}
	return true, record.Card.Path, record.Card.BlobOID, nil
}

func (s *closePreparationReadSet) validate() error {
	if err := commandContext(s.ctx).Err(); err != nil {
		return err
	}
	if err := s.initial.prepared.validateIdentityAndArtifacts(); err != nil {
		return err
	}
	if s.tracked {
		target, err := gitx.ResolvePublicationTarget(commandContext(s.ctx), s.initial.prepared.root, s.resting)
		if err != nil {
			return err
		}
		if target != s.target {
			return fmt.Errorf("close publication target changed during preparation or review")
		}
	}
	// A fresh scope is essential: the preparation's command cache deliberately
	// retains one generation and cannot establish current write authority.
	records, err := observeClosePreparationCards(s.ctx, s.initial.prepared.root, s.issues)
	if err != nil {
		return err
	}
	tracked, path, oid, err := closeCardAuthority(records, s.issue)
	if err != nil {
		return err
	}
	if tracked != s.tracked || path != s.cardPath || oid != s.cardOID {
		return fmt.Errorf("close card #%s changed during preparation or review", s.issue)
	}
	return compareAdoptionReadSet(s.adopted, records.All())
}

// prepareCloseTransaction confines the lock to capture and admission. In
// particular, computeClose's transcript/history scans never monopolize it.
func prepareCloseTransaction(cmd *cobra.Command, stderr io.Writer, f *closeFlags) (closeResult, closeReviewSnapshot, error) {
	var guard closePreparationReadSet
	guard.ctx, guard.issues, guard.issue = f.Context, f.IssuesDir, fmt.Sprintf("%06d", f.Issue)
	err := withRequiredRepoTransactionLock(cmd, func() error {
		path, err := issueFilePath(f.IssuesDir, f.Issue)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		guard.initial, err = captureCloseReviewSnapshot(closeResult{issuePath: path, issueText: string(data)}, "", f.Milestone, f.plansDir())
		if err != nil {
			return err
		}
		guard.tracked, guard.cardPath, guard.cardOID, err = captureCloseCard(commandContext(f.Context), f.IssuesDir, guard.issue)
		if err == nil && guard.tracked {
			env, e := openTracker(commandContext(f.Context))
			if e != nil {
				return e
			}
			guard.resting, guard.target = env.resting, env.target
		}
		return err
	})
	if err != nil {
		return closeResult{}, closeReviewSnapshot{}, err
	}
	r := computeClose(stderr, f)
	if f.effects != nil {
		base := boundaryWindowBase(commandContext(f.Context), r.issueStr, f.Milestone, r.issuePath, f.plansDir())
		if report, err := churnForWindow(base); err != nil {
			cwarn(stderr, fmt.Sprintf("churn not measured (%v)", err))
		} else {
			f.effects.churn = report
		}
	}
	// An unmeasured close (#603) adopts nothing, so it holds no adoption
	// authority to revalidate; its read-set may be unavailable by cause.
	if !f.unmeasured() {
		guard.adopted = f.Adoptions
	}
	snapshot, err := captureCloseReviewSnapshot(r, guard.initial.reviewed, f.Milestone, f.plansDir())
	if err != nil {
		return r, snapshot, err
	}
	snapshot.preparation = &guard
	err = withRequiredRepoTransactionLock(cmd, func() error {
		if err := guard.initial.prepared.validateExact(); err != nil {
			return fmt.Errorf("close preparation stale: %w; rerun %s", err, closeVerb(f.Milestone))
		}
		_, err := snapshot.validate()
		return err
	})
	if err != nil {
		return r, snapshot, fmt.Errorf("close preparation admission/reacquire: %w; rerun %s", err, closeVerb(f.Milestone))
	}
	return r, snapshot, nil
}
