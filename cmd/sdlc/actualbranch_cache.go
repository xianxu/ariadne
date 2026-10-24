package main

import (
	"context"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type ownedBranchRequestKey struct{}
type ownedBranchHistoryResult struct {
	events []tracker.OwnershipEvent
	err    error
}

// One sequential measurement request freezes each source once. In particular,
// standalone multi-issue allocation must not observe different checkout histories
// between its partitions. Nothing is cached across commands or tracker pins.
type ownedBranchRequest struct {
	histories   map[[2]string]ownedBranchHistoryResult
	captures    map[string]ownedBranchCapture
	readHistory func(*tracker.Repository, string, string) ([]tracker.OwnershipEvent, error)
	capture     func(string, time.Time, time.Time) ownedBranchCapture
}

func ownedBranchRequestContext(ctx context.Context) context.Context {
	if ctx.Value(ownedBranchRequestKey{}) != nil {
		return ctx
	}
	r := &ownedBranchRequest{histories: map[[2]string]ownedBranchHistoryResult{}, captures: map[string]ownedBranchCapture{}, readHistory: func(repo *tracker.Repository, pin, path string) ([]tracker.OwnershipEvent, error) {
		return repo.EngagementHistory(pin, path)
	}}
	r.capture = func(root string, since, until time.Time) ownedBranchCapture {
		return captureOwnedBranchEvidence(ctx, root, since, until)
	}
	return context.WithValue(ctx, ownedBranchRequestKey{}, r)
}

func loadOwnedBranchHistory(ctx context.Context, repo *tracker.Repository, pin, path string) ([]tracker.OwnershipEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r, ok := ctx.Value(ownedBranchRequestKey{}).(*ownedBranchRequest)
	if !ok {
		return repo.EngagementHistory(pin, path)
	}
	key := [2]string{pin, path}
	if saved, ok := r.histories[key]; ok {
		return saved.events, saved.err
	}
	events, err := r.readHistory(repo, pin, path)
	r.histories[key] = ownedBranchHistoryResult{events, err}
	return events, err
}

func captureOwnedBranchCached(ctx context.Context, root string, since, until time.Time) ownedBranchCapture {
	r, ok := ctx.Value(ownedBranchRequestKey{}).(*ownedBranchRequest)
	if !ok {
		return captureOwnedBranchEvidence(ctx, root, since, until)
	}
	if saved, ok := r.captures[root]; ok {
		if saved.Unavailable == "" {
			var err error
			saved.Intervals, err = activetime.ParseHeadHistory(saved.raw, saved.Branch, saved.Head, since, until)
			if err != nil {
				saved.Unavailable = err.Error()
				saved.Intervals = nil
			}
		}
		return saved
	}
	saved := r.capture(root, since, until)
	r.captures[root] = saved
	return saved
}
