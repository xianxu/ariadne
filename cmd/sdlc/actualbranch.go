package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// ownedBranchCapture is a bounded observation, never a replacement for saved
// measurement windows. A failed capture is unknown and can support only the
// explicitly labelled sole-reservation fallback.
type ownedBranchCapture struct {
	Root, Path, Head, Branch, Digest string
	Bytes                            int64
	Intervals                        []activetime.BranchInterval
	Unavailable                      string
}

func captureOwnedBranchEvidence(ctx context.Context, root string, since, until time.Time) (capture ownedBranchCapture) {
	capture.Root = root
	capture.Unavailable = "HEAD history unavailable"
	git := func(args ...string) (string, error) {
		return ownedActualGit(ctx, root, append([]string{"--no-optional-locks"}, args...)...)
	}
	observe := func() (string, string, string, error) {
		head, err := git("rev-parse", "--verify", "HEAD^{commit}")
		if err != nil {
			return "", "", "", err
		}
		branch, err := git("rev-parse", "--symbolic-full-name", "HEAD")
		if err != nil {
			return "", "", "", err
		}
		path, err := git("rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
		return head, strings.TrimPrefix(branch, "refs/heads/"), path, err
	}
	head, branch, path, err := observe()
	if err != nil {
		return
	}
	capture.Path, capture.Head, capture.Branch = path, head, branch
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return
	}
	if before.Size() > activetime.HeadHistoryByteLimit {
		capture.Unavailable = "HEAD history exceeds capture byte limit"
		return
	}
	if ctx.Err() != nil {
		return
	}
	raw, err := io.ReadAll(io.LimitReader(f, activetime.HeadHistoryByteLimit+1))
	if err != nil || ctx.Err() != nil {
		return
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || int64(len(raw)) != before.Size() || !before.ModTime().Equal(after.ModTime()) {
		capture.Unavailable = "HEAD history changed during capture"
		return
	}
	nextHead, nextBranch, nextPath, err := observe()
	if err != nil || head != nextHead || branch != nextBranch || path != nextPath {
		capture.Unavailable = "HEAD identity changed during capture"
		return
	}
	capture.Bytes = int64(len(raw))
	capture.Digest = fmt.Sprintf("%x", sha256.Sum256(raw))
	capture.Intervals, err = activetime.ParseHeadHistory(raw, branch, head, since, until)
	if err != nil {
		capture.Unavailable = err.Error()
		return
	}
	capture.Unavailable = ""
	return
}

func branchOwnerKey(owner issue.Claimant) string {
	// A relocation changes its physical root, not the owning slot. An unnumbered
	// workspace has no slot identity and therefore uses its verified worktree.
	slot := owner.Workspace
	if slot == "" {
		slot = filepath.Clean(owner.Worktree)
	}
	return owner.Machine + "\x00" + owner.Repository + "\x00" + slot
}

// resolveSlotClaimWindows reads each immutable card history once. Historical
// owner changes matter even when today's owner is a different slot.
func resolveSlotClaimWindows(ctx context.Context, repo *tracker.Repository, inventory tracker.Snapshot, candidate string, until time.Time) (map[string][]activetime.BranchClaim, map[string]string, error) {
	if bad := inventory.Unreadable(); len(bad) != 0 {
		return nil, nil, fmt.Errorf("branch attribution cannot establish sole ownership: unreadable card %s", bad[0].Path)
	}
	claims, roots := map[string][]activetime.BranchClaim{}, map[string]string{}
	for _, card := range inventory.Records() {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		events, err := repo.EngagementHistory(inventory.Ref(), card.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("branch claim history %s: %w", card.ID, err)
		}
		// Validate owner chronology with the existing authority; keep its
		// retained-terminal semantics unchanged for other consumers.
		if _, err := tracker.DeriveEngagements(events, until); err != nil {
			return nil, nil, err
		}
		id, err := strconv.Atoi(card.ID)
		if err != nil {
			return nil, nil, err
		}
		firstClaim := true
		for i, event := range events {
			if event.Owner == nil {
				continue
			}
			start, end := event.At, until
			if firstClaim && event.Operation == tracker.OpClaim && !event.Started.IsZero() && !event.Started.Before(events[0].At) {
				start = event.Started
			}
			firstClaim = false
			if i+1 < len(events) && events[i+1].At.Before(end) {
				end = events[i+1].At
			}
			if !start.Before(end) {
				continue
			}
			key := branchOwnerKey(*event.Owner)
			if strconv.Itoa(id) == candidate {
				root := filepath.Clean(event.Owner.Worktree)
				if prior := roots[root]; prior != "" && prior != key {
					return nil, nil, fmt.Errorf("ambiguous historical owner for %s", root)
				}
				roots[root] = key
			}
			model := vocab.Issue()
			if !model.CanHoldOwner(event.Status) && !model.IsTerminal(event.Status) {
				return nil, nil, fmt.Errorf("branch claim %s has unknown status %q", card.ID, event.Status)
			}
			claims[key] = append(claims[key], activetime.BranchClaim{Start: start, End: end, Issue: strconv.Itoa(id), Terminal: model.IsTerminal(event.Status)})
		}
	}
	return claims, roots, nil
}

// selectOwnedBranchWindows runs only after source lineage admission. Spawned
// sessions inherit their root author's worktree; a child cwd (e.g. brain) is
// never mistaken for the checkout whose HEAD selects this issue.
func selectOwnedBranchWindows(ctx context.Context, repo *tracker.Repository, inventory tracker.Snapshot, id string, sessions []transcripts.SessionEvidence, admitted map[string][]activetime.ActivityWindow, since, until time.Time) (map[string][]activetime.ActivityWindow, []ownedBranchCapture, []string, error) {
	claims, roots, err := resolveSlotClaimWindows(ctx, repo, inventory, id, until)
	if err != nil {
		return nil, nil, nil, err
	}
	byID := map[string]transcripts.SessionEvidence{}
	for _, s := range sessions {
		byID[s.Harness+":"+s.ID] = s
	}
	captures := map[string]ownedBranchCapture{}
	out := map[string][]activetime.ActivityWindow{}
	var warnings []string
	for _, s := range sessions {
		if len(admitted[s.File]) == 0 {
			continue
		}
		ancestor := s
		seen := map[string]bool{}
		for ancestor.Spawned {
			key := ancestor.Harness + ":" + ancestor.ID
			if seen[key] {
				return nil, nil, nil, fmt.Errorf("branch source lineage cycle")
			}
			seen[key] = true
			parent, ok := byID[ancestor.Harness+":"+ancestor.ParentID]
			if !ok {
				return nil, nil, nil, fmt.Errorf("branch source %s lacks admitted root lineage", s.ID)
			}
			ancestor = parent
		}
		root := filepath.Clean(ancestor.CWD)
		owner, ok := roots[root]
		if !ok {
			return nil, nil, nil, fmt.Errorf("branch source %s has no pinned worktree owner", s.ID)
		}
		capture, ok := captures[root]
		if !ok {
			capture = captureOwnedBranchEvidence(ctx, root, since, until)
			captures[root] = capture
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		selected, err := activetime.SelectBranchActivity(id, admitted[s.File], capture.Intervals, claims[owner])
		if err != nil {
			return nil, nil, nil, fmt.Errorf("branch attribution for %s (%s): %w; %s; close recovery: --no-actual with the evidence in --verified", s.ID, root, err, capture.Unavailable)
		}
		out[s.File] = selected.Windows
		for _, note := range selected.Notes {
			if note.Kind == "branch" {
				continue
			}
			warnings = append(warnings, fmt.Sprintf("branch attribution %s: %s %s during %s → %s", s.ID, note.Kind, note.Branch, note.Start.Format(time.RFC3339Nano), note.End.Format(time.RFC3339Nano)))
		}
	}
	var ordered []ownedBranchCapture
	for _, capture := range captures {
		ordered = append(ordered, capture)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Root < ordered[j].Root })
	return out, ordered, warnings, nil
}
