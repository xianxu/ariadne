package main

import (
	"context"
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
	impacts := make(map[string]landingline.Impact, len(packages))
	for _, pkg := range packages {
		impacts[pkg] = landingline.ImpactUnknown
	}
	if err := ctx.Err(); err != nil {
		return impacts, err
	}
	if !landingOIDValid(loSHA) || !landingOIDValid(hiSHA) {
		return impacts, fmt.Errorf("split impact requires exact endpoint commit IDs")
	}
	r := boundedLandingGit(ctx, execGitRunner{})
	git := func(args ...string) (string, error) { return landingGit(r, root, args...) }
	out, err := git(gitx.ChangedPathArgs(loSHA, hiSHA)...)
	if err != nil {
		return impacts, fmt.Errorf("split changed paths: %w", err)
	}
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return impacts, fmt.Errorf("split endpoints have no changed paths")
	}
	// Retire both private checkouts together under one independent five-second
	// budget. Start that clock when cleanup begins, after graph acquisition.
	var cleanupCtx context.Context
	var cleanupCancel context.CancelFunc
	defer func() {
		if cleanupCancel != nil {
			cleanupCancel()
		}
	}()
	cleanupGit := func(args ...string) (string, error) {
		if cleanupCtx == nil {
			cleanupCtx, cleanupCancel = context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		}
		return landingGit(boundedLandingGit(cleanupCtx, execGitRunner{}), root, args...)
	}
	var graphs [2]affected.Graph
	var cfgs [2]affected.Config
	var mappings [2]affected.Mapping
	var selections [2]affected.Selection
	for i, sha := range []string{loSHA, hiSHA} {
		cfg, ok, err := readAffectedConfig(git, sha)
		if err != nil {
			return impacts, err
		}
		if !ok {
			return impacts, fmt.Errorf("split impact: %s missing on %s", affectedConfigPath, abbrevSHA(sha))
		}
		wt, cleanup, err := detachedWorktreeWithCleanup(git, cleanupGit, sha)
		if err != nil {
			return impacts, err
		}
		defer cleanup()
		if err := prepareAffectedGraph(ctx, localRunDeps{Stderr: io.Discard}, wt, cfg.GraphPrepare); err != nil {
			return impacts, err
		}
		graphs[i], err = affected.LoadGoGraph(ctx, wt)
		if err != nil {
			return impacts, fmt.Errorf("split package graph at %s: %w", abbrevSHA(sha), err)
		}
		// Each endpoint's own declarations, read from its exact commit, as its graph is (#598).
		mappings[i], err = loadAffectedMapping(ctx, root, sha, cfg)
		if err != nil {
			return impacts, fmt.Errorf("split declared ownership at %s: %w", abbrevSHA(sha), err)
		}
		cfgs[i] = cfg
		selections[i] = affected.Plan(paths, graphs[i], cfg, mappings[i], affectedIsDoc)
	}
	if err := ctx.Err(); err != nil {
		return impacts, err
	}
	if graphs[0].Module == "" || graphs[0].Module != graphs[1].Module {
		return impacts, fmt.Errorf("split endpoint module identities do not match")
	}
	// A broad/unknown selection never proves innocence. Still retain explicit
	// ownership/dependency evidence: one unrelated unknown path cannot erase a
	// known connection between the candidate and another exact failing package.
	uncertain := false
	affectedPackages := map[string]bool{}
	for i, sel := range selections {
		pkgs := sel.GoPkgs
		if sel.Full || len(sel.Groups) != 0 {
			uncertain = true
			changed := map[string]bool{}
			for _, p := range paths {
				for _, pkg := range affected.Owned(p, graphs[i], cfgs[i], mappings[i]) {
					changed[pkg] = true
				}
			}
			pkgs = graphs[i].Dependents(changed)
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
	return impacts, nil
}
