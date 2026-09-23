package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// A review safety failure cannot be waived as a content-quality finding.
type planningReviewSafetyError struct{ err error }

func (e *planningReviewSafetyError) Error() string { return e.err.Error() }
func (e *planningReviewSafetyError) Unwrap() error { return e.err }
func planningReviewUnsafe(err error) error         { return &planningReviewSafetyError{err: err} }

// planningReviewTransaction owns the command's lock and its immutable review
// inputs. Only the plan ledger is refreshed after the command writes a round.
type planningReviewTransaction struct {
	cmd        *cobra.Command
	release    func() error
	snapshot   preparedReview
	supplied   []reviewArtifact
	ledgerPath string
}

func runChangeCodeCommand(cmd *cobra.Command, f *changeCodeFlags) error {
	tx := &planningReviewTransaction{cmd: cmd}
	if err := tx.lock(); err != nil {
		return err
	}
	unregister := registerDieCleanup(func() { _ = tx.unlock() })
	defer unregister()
	defer tx.unlock()
	f.review = tx
	defer func() { f.review = nil }()
	return runChangeCode(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), f)
}

func (t *planningReviewTransaction) context() context.Context {
	if t.cmd.Context() != nil {
		return t.cmd.Context()
	}
	return context.Background()
}

func (t *planningReviewTransaction) lock() error {
	if err := t.context().Err(); err != nil {
		return planningReviewUnsafe(fmt.Errorf("review interrupted: %w", err))
	}
	release, err := repoLockAcquireForCommand(t.cmd)
	if err != nil {
		return planningReviewUnsafe(fmt.Errorf("review lock acquisition failed: %w", err))
	}
	t.release = release
	return nil
}

func (t *planningReviewTransaction) unlock() error {
	if t.release == nil {
		return nil
	}
	release := t.release
	t.release = nil
	return release()
}

func (t *planningReviewTransaction) prepare(supplied []reviewArtifact, ledgerPath string) error {
	t.supplied = supplied
	t.ledgerPath = ledgerPath
	s, err := capturePreparedReview("", supplied, ledgerPath)
	if err != nil {
		return planningReviewUnsafe(err)
	}
	t.snapshot = s
	return t.validate()
}

func (t *planningReviewTransaction) validate() error {
	if err := t.context().Err(); err != nil {
		return planningReviewUnsafe(fmt.Errorf("review interrupted: %w", err))
	}
	if err := t.snapshot.validateExact(); err != nil {
		return planningReviewUnsafe(fmt.Errorf("planning review stale: %w; rerun change-code", err))
	}
	if err := t.context().Err(); err != nil {
		return planningReviewUnsafe(fmt.Errorf("review interrupted: %w", err))
	}
	return nil
}

func (t *planningReviewTransaction) refreshOwnLedger() error {
	s, err := capturePreparedReview(t.snapshot.head, t.supplied, t.ledgerPath)
	if err != nil {
		return planningReviewUnsafe(fmt.Errorf("planning review stale: %w", err))
	}
	if err := compareReviewIdentity(t.snapshot, s); err != nil {
		return planningReviewUnsafe(fmt.Errorf("planning review stale: %w", err))
	}
	t.snapshot = s
	return t.validate()
}

func dispatchPlanningReview(f *changeCodeFlags, opts judge.DispatchOptions) (string, error) {
	if f.review == nil {
		return judge.Dispatch(context.Background(), opts)
	}
	t := f.review
	if err := t.validate(); err != nil {
		return "", err
	}
	if err := t.unlock(); err != nil {
		return "", planningReviewUnsafe(fmt.Errorf("release review lock: %w", err))
	}
	output, dispatchErr := judge.Dispatch(t.context(), opts)
	if err := t.lock(); err != nil {
		return "", err
	}
	// Validate before any verdict, fallback round, cache, or subsequent gate.
	if err := t.validate(); err != nil {
		return "", err
	}
	if dispatchErr != nil {
		return "", planningReviewUnsafe(fmt.Errorf("review dispatch failed: %w", dispatchErr))
	}
	return output, nil
}
