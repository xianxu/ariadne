package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// affectedDeps are runAffectedCheck's seams (#473); production wires git,
// affected.LoadGoGraph in a worktree of the head, and the local-check runner.
type affectedDeps struct {
	changed func(sha string) ([]string, error)
	graph   func(ctx context.Context, sha string) (affected.Graph, error)
	run     func(ctx context.Context, c localCheck, sha string) (localRunRecord, error)
	post    func(ctx context.Context, rec localRunRecord) error
}

// runAffectedCheck is the affected suite (#473): it plans the head's change
// against base and runs the repo's affected command on the selection. A full
// selection runs and records suite instead (D5), never an affected record that
// claims to be full; a docs-only change records affected without running
// anything.
func runAffectedCheck(ctx context.Context, d localRunDeps, ad affectedDeps, cfg affected.Config, suite localCheck, sha string) (localRunRecord, error) {
	paths, err := ad.changed(sha)
	if err != nil {
		return localRunRecord{}, fmt.Errorf("local check %s: changed paths of %s: %w", affectedCheckName, abbrevSHA(sha), err)
	}
	if len(paths) == 0 {
		// No change is not a docs-only change: recording a pass here would be
		// evidence nothing ran for (#473 M1 BR-2).
		return localRunRecord{}, fmt.Errorf("local check %s: %s has no changes against its merge-base with main; there is nothing to vouch for, run --name %s", affectedCheckName, abbrevSHA(sha), suite.Name)
	}
	// Graph preparation can execute even for a docs-only selection, before
	// the selected check reaches the ordinary runner.
	if err := preflightSuite(ctx, d, localCheck{Name: affectedCheckName}, sha); err != nil {
		return localRunRecord{}, err
	}
	if cfg.GraphPrepare != nil {
		// The generator precedes runUnderSlot's signal owner. Keep its
		// cancellation alive through graph cleanup, fallback and publication.
		var stop func()
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}
	// The graph is loaded even for a docs-only diff: a package may embed or
	// test against a doc, and only the graph knows (close review).
	g, graphErr := ad.graph(ctx, sha)
	if err := ctx.Err(); err != nil {
		return localRunRecord{}, fmt.Errorf("local check %s: graph interrupted: %w", affectedCheckName, err)
	}
	var sel affected.Selection
	if graphErr != nil {
		sel = affected.Selection{Full: true, Reason: "package graph unavailable: " + graphErr.Error()}
	} else {
		sel = affected.Plan(paths, g, cfg, func(p string) bool { return !publishGateHasCodeSurface([]string{p}) })
	}
	switch {
	case sel.Full:
		fmt.Fprintf(d.Stderr, "==> local check %s on %s: full selection (%s); running %s instead\n", affectedCheckName, abbrevSHA(sha), sel.Reason, suite.Name)
		return ad.run(ctx, suite, sha)
	case sel.Empty():
		rec := localRunRecord{Name: affectedCheckName, SHA: sha, Host: hostName(), Scope: "nothing affected (docs-only)"}
		fmt.Fprintf(d.Stderr, "==> local check %s on %s: %s; recording without a run\n", affectedCheckName, abbrevSHA(sha), rec.Scope)
		return rec, ad.post(ctx, rec)
	}
	return ad.run(ctx, affectedLocalCheck(cfg, sel, suite), sha)
}

// affectedLocalCheck is the selection as a local check: the repo's command
// under the fixed env contract (#473), sorted and space-separated, passed as
// argv through env(1) and never through a shell.
func affectedLocalCheck(cfg affected.Config, sel affected.Selection, suite localCheck) localCheck {
	c := affectedCheckDecl(cfg, suite)
	c.Command = append([]string{"env",
		"SDLC_AFFECTED_GO_PKGS=" + strings.Join(sel.GoPkgs, " "),
		"SDLC_AFFECTED_GROUPS=" + strings.Join(sel.Groups, " "),
	}, cfg.Command...)
	c.Scope = affectedScope(sel)
	return c
}

// affectedCheckDecl is the affected suite as a declared check, the one place
// its name resolves to a localCheck: affected.json's command and limits, the
// suite's known-flake policy (D4).
func affectedCheckDecl(cfg affected.Config, suite localCheck) localCheck {
	return localCheck{Name: affectedCheckName, Command: cfg.Command, TimeoutMinutes: cfg.TimeoutMinutes, Weight: cfg.Weight,
		KnownFlakes: suite.KnownFlakes, DeclarationVersion: suite.DeclarationVersion, FlakeRunners: suite.FlakeRunners}
}

// affectedScope summarizes a selection for the commit status.
func affectedScope(sel affected.Selection) string {
	var parts []string
	switch n := len(sel.GoPkgs); n {
	case 0:
	case 1:
		parts = append(parts, "1 package")
	default:
		parts = append(parts, fmt.Sprintf("%d packages", n))
	}
	if len(sel.Groups) > 0 {
		parts = append(parts, "groups "+strings.Join(sel.Groups, " "))
	}
	return strings.Join(parts, ", ")
}

// productionAffectedDeps diffs the head against its merge-base with base and
// loads the package graph from a detached worktree of the head, released
// before the run makes its own.
func productionAffectedDeps(d localRunDeps, base string, cfg affected.Config) affectedDeps {
	cleanupGit := d.CleanupGit
	if cleanupGit == nil {
		cleanupGit = d.Git
	}
	return affectedDeps{
		changed: func(sha string) ([]string, error) {
			mb, err := d.Git("merge-base", base, sha)
			if err != nil {
				return nil, fmt.Errorf("merge-base %s %s: %w", base, abbrevSHA(sha), err)
			}
			out, err := d.Git(gitx.ChangedPathArgs(strings.TrimSpace(mb), sha)...)
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
		},
		graph: func(ctx context.Context, sha string) (affected.Graph, error) {
			wt, cleanup, err := detachedWorktreeWithCleanup(d.Git, cleanupGit, sha)
			if err != nil {
				return affected.Graph{}, err
			}
			defer cleanup()
			if err := prepareAffectedGraph(ctx, d, wt, cfg.GraphPrepare); err != nil {
				return affected.Graph{}, err
			}
			return affected.LoadGoGraph(ctx, wt)
		},
		run: func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			return runLocalCheck(ctx, d, c, sha)
		},
		post: func(ctx context.Context, rec localRunRecord) error { return postLocalCheckRecord(ctx, d, rec) },
	}
}
