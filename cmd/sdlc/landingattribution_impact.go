package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// splitImpact proves package impact only for the exact adjacent prefix delta.
// Both endpoint graphs contribute reverse dependencies: deleting an import or
// moving a package cannot erase the old endpoint's affected packages. The caller
// owns the overall assessment deadline. Each failed proof remains Unknown.
func splitImpact(ctx context.Context, root, loSHA, hiSHA string, packages []string) (map[string]landingline.Impact, error) {
	impacts, _, err := splitImpactPaths(ctx, root, loSHA, hiSHA, packages)
	return impacts, err
}

// splitImpactPaths is splitImpact plus the delta's changed paths, the evidence a missed mapping records (#616).
func splitImpactPaths(ctx context.Context, root, loSHA, hiSHA string, packages []string) (impacts map[string]landingline.Impact, paths []string, err error) {
	impacts = make(map[string]landingline.Impact, len(packages))
	for _, pkg := range packages {
		impacts[pkg] = landingline.ImpactUnknown
	}
	if err := ctx.Err(); err != nil {
		return impacts, paths, err
	}
	if !landingOIDValid(loSHA) || !landingOIDValid(hiSHA) {
		return impacts, paths, fmt.Errorf("split impact requires exact endpoint commit IDs")
	}
	r := boundedLandingGit(ctx, execGitRunner{})
	git := func(args ...string) (string, error) { return landingGit(r, root, args...) }
	if paths, err = changedPathsBetween(git, loSHA, hiSHA); err != nil {
		return impacts, paths, fmt.Errorf("split changed paths: %w", err)
	}
	if len(paths) == 0 {
		return impacts, paths, fmt.Errorf("split endpoints have no changed paths")
	}
	cleanupGit, retire := retiringCleanupGit(ctx, root)
	defer retire()
	var at [2]affectedAt
	for i, sha := range []string{loSHA, hiSHA} {
		var cleanup func()
		var err error
		// Each endpoint's own declarations, read from its exact commit, as its graph is (#598).
		at[i], cleanup, err = affectedSelectionAt(ctx, root, git, cleanupGit, sha, sha, paths)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return impacts, paths, fmt.Errorf("split %w", err)
		}
	}
	if err := ctx.Err(); err != nil {
		return impacts, paths, err
	}
	graphs := [2]affected.Graph{at[0].graph, at[1].graph}
	if graphs[0].Module == "" || graphs[0].Module != graphs[1].Module {
		return impacts, paths, fmt.Errorf("split endpoint module identities do not match")
	}
	// A broad/unknown selection never proves innocence. Still retain explicit
	// ownership/dependency evidence: one unrelated unknown path cannot erase a
	// known connection between the candidate and another exact failing package.
	uncertain := false
	affectedPackages := map[string]bool{}
	for _, a := range at {
		sel := a.sel
		pkgs := sel.GoPkgs
		if sel.Full || len(sel.Groups) != 0 {
			uncertain = true
			changed := map[string]bool{}
			for _, p := range paths {
				for _, pkg := range affected.Owned(p, a.graph, a.cfg, a.mapping) {
					changed[pkg] = true
				}
			}
			pkgs = a.graph.Dependents(changed)
		}
		for _, pkg := range pkgs {
			affectedPackages[pkg] = true
		}
	}
	for pkg := range impacts {
		if affectedPackages[pkg] {
			impacts[pkg] = landingline.ImpactAffected
			continue
		}
		_, inLo := graphs[0].Dir[pkg]
		_, inHi := graphs[1].Dir[pkg]
		if !uncertain && inLo && inHi {
			impacts[pkg] = landingline.ImpactUntouched
		}
	}
	return impacts, paths, nil
}

// affectedAt is one commit range's affected selection and what produced it.
type affectedAt struct {
	sel     affected.Selection
	graph   affected.Graph
	cfg     affected.Config
	mapping affected.Mapping
}

// errAffectedUndeclared: the declaring commit has no affected.json, so there is no selection to speak of.
var errAffectedUndeclared = errors.New(affectedConfigPath + " missing")

// affectedSelectionAt plans paths as an affected run would (#616): declRev's config and declared ownership over
// graphRev's package graph, built in a private detached checkout. splitImpact plans each endpoint from itself; a
// merge's missed-mapping check plans the head's delta from main's declarations. The returned cleanup retires the
// checkout and is non-nil once one exists, even on error.
func affectedSelectionAt(ctx context.Context, root string, git, cleanupGit func(args ...string) (string, error), declRev, graphRev string, paths []string) (affectedAt, func(), error) {
	cfg, ok, err := readAffectedConfig(git, declRev)
	if err != nil {
		return affectedAt{}, nil, err
	}
	if !ok {
		return affectedAt{}, nil, fmt.Errorf("%w on %s", errAffectedUndeclared, abbrevSHA(declRev))
	}
	wt, cleanup, err := detachedWorktreeWithCleanup(git, cleanupGit, graphRev)
	if err != nil {
		return affectedAt{}, nil, err
	}
	if err := prepareAffectedGraph(ctx, localRunDeps{Stderr: io.Discard}, wt, cfg.GraphPrepare); err != nil {
		return affectedAt{}, cleanup, err
	}
	g, err := affected.LoadGoGraph(ctx, wt)
	if err != nil {
		return affectedAt{}, cleanup, fmt.Errorf("package graph at %s: %w", abbrevSHA(graphRev), err)
	}
	m, err := loadAffectedMapping(ctx, root, declRev, cfg)
	if err != nil {
		return affectedAt{}, cleanup, fmt.Errorf("declared ownership at %s: %w", abbrevSHA(declRev), err)
	}
	return affectedAt{sel: affected.Plan(paths, g, cfg, m, affectedIsDoc), graph: g, cfg: cfg, mapping: m}, cleanup, nil
}

// changedPathsBetween lists the paths that differ from lo to hi.
func changedPathsBetween(git func(args ...string) (string, error), lo, hi string) ([]string, error) {
	out, err := git(gitx.ChangedPathArgs(lo, hi)...)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

// retiringCleanupGit retires private checkouts under one independent five-second budget, whose clock starts when
// cleanup begins (after graph acquisition), so a spent caller deadline never strands a checkout. done releases it.
func retiringCleanupGit(ctx context.Context, root string) (git func(args ...string) (string, error), done func()) {
	var cleanupCtx context.Context
	var cleanupCancel context.CancelFunc
	git = func(args ...string) (string, error) {
		if cleanupCtx == nil {
			cleanupCtx, cleanupCancel = context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		}
		return landingGit(boundedLandingGit(cleanupCtx, execGitRunner{}), root, args...)
	}
	return git, func() {
		if cleanupCancel != nil {
			cleanupCancel()
		}
	}
}
