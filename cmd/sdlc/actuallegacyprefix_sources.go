package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// discoverLegacyPrefixInputs discovers the unowned, pre-cutover interval once.
// Its authority has already been resolved from immutable Git history by the
// caller; this seam adds only exact local transcript evidence.
func discoverLegacyPrefixInputs(ctx context.Context, root, brain string, snapshot gatestate.MeasurementInputs) (*ownedActualInputs, error) {
	roots, err := legacyPrefixDiscoveryRoots(ctx, root, brain)
	if err != nil {
		return nil, err
	}
	return discoverLegacyPrefixSources(ctx, roots, snapshot, transcripts.DiscoverEvidence)
}

type legacyPrefixDiscovery func(context.Context, []string, []transcripts.Harness) ([]transcripts.SessionEvidence, []string, error)

func discoverLegacyPrefixSources(ctx context.Context, roots []string, snapshot gatestate.MeasurementInputs, discover legacyPrefixDiscovery) (*ownedActualInputs, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sessions, warnings, err := discover(ctx, roots, transcripts.DefaultHarnesses())
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return bindLegacyPrefixSources(snapshot, sessions, warnings)
}

// bindLegacyPrefixSources is linear in discovered sessions. Candidate sources
// may exceed retained snapshot limits: the engine first filters contributing
// pins, just as for the modern component. There is no ownership or slot guess.
func bindLegacyPrefixSources(snapshot gatestate.MeasurementInputs, sessions []transcripts.SessionEvidence, warnings []string) (*ownedActualInputs, error) {
	since, err := time.Parse(time.RFC3339Nano, snapshot.Since)
	if err != nil {
		return nil, fmt.Errorf("legacy prefix since: %w", err)
	}
	until, err := time.Parse(time.RFC3339Nano, snapshot.Until)
	if err != nil {
		return nil, fmt.Errorf("legacy prefix until: %w", err)
	}
	if !until.After(since) {
		return nil, fmt.Errorf("legacy prefix until must follow since")
	}
	snapshot.Sources = nil
	out := &ownedActualInputs{Snapshot: snapshot, Prefixes: map[string]activetime.SourcePrefix{}, Windows: map[string][]activetime.ActivityWindow{}, Scope: activetime.Scope{Head: snapshot.CodeHead}, Warnings: append([]string(nil), warnings...)}
	for _, s := range sessions {
		if !s.Started.Before(until) {
			continue
		}
		start := since
		if s.Started.After(start) {
			start = s.Started
		}
		windows := []activetime.ActivityWindow{{Start: start, End: until}}
		out.Files = append(out.Files, s.File)
		out.Sessions = append(out.Sessions, s)
		out.Prefixes[s.File] = activetime.SourcePrefix{Bytes: s.Length, Digest: s.Digest}
		out.Windows[s.File] = windows
		out.Snapshot.Sources = append(out.Snapshot.Sources, ownedMeasurementSource(s, windows))
	}
	if len(out.Files) == 0 {
		return nil, fmt.Errorf("legacy prefix transcript sources are unavailable before cutover")
	}
	return out, nil
}

func legacyPrefixDiscoveryRoots(ctx context.Context, root, brain string) ([]string, error) {
	root, err := workspace.CanonicalPath(root)
	if err != nil {
		return nil, err
	}
	roots := []string{root}
	listing, err := ownedActualGit(ctx, root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	trees, err := gitx.ParseWorktrees([]byte(listing))
	if err != nil {
		return nil, err
	}
	for _, tree := range trees {
		if tree.Bare {
			continue
		}
		primary, err := workspace.CanonicalPath(tree.Path)
		// The primary path is explicit Git evidence. If gone, do not substitute a
		// different linked worker or manufacture a historical path.
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		if primary != root {
			currentCommon, err := legacyPrefixCommonDir(ctx, root)
			if err != nil {
				return nil, err
			}
			primaryCommon, err := legacyPrefixCommonDir(ctx, primary)
			if err != nil {
				return nil, err
			}
			if primaryCommon != currentCommon {
				return nil, fmt.Errorf("legacy prefix primary worktree now belongs to another repository")
			}
			roots = append(roots, primary)
		}
		break
	}
	if brain != "" {
		canonical, err := workspace.CanonicalPath(brain)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			roots = append(roots, canonical)
		}
	}
	return uniqueStrings(roots), nil
}

func legacyPrefixCommonDir(ctx context.Context, root string) (string, error) {
	common, err := ownedActualGit(ctx, root, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	return workspace.CanonicalPath(common)
}
