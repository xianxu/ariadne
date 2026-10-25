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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// affectedDeps are runAffectedCheck's seams (#473); production wires git,
// affected.LoadGoGraph in a worktree of the head, and the local-check runner.
type affectedDeps struct {
	changed func(sha string) ([]string, error)
	graph   func(ctx context.Context, sha string) (affected.Graph, error)
	// mapping is the declared ownership read at main's pinned rev (#598), never the head's.
	mapping func(ctx context.Context) (affected.Mapping, error)
	run     func(ctx context.Context, c localCheck, sha string) (localRunRecord, error)
	post    func(ctx context.Context, rec localRunRecord) error
}

// affectedExecutionMode decides what a full selection does (#561); selection
// itself stays affected.Plan's. Ordinary falls back to the suite; boarding,
// a train car's preparation, refuses instead, because the train runs the full
// suite on its prefix and a per-car suite is the cost trains exist to avoid.
type affectedExecutionMode int

const (
	affectedOrdinary affectedExecutionMode = iota
	affectedBoarding
)

// affectedModeFor validates --boarding (#561): it qualifies a run of affected
// alone, never suite, a mix of names or --all, where it would be ignored.
func affectedModeFor(names []string, all, boarding bool) (affectedExecutionMode, error) {
	if !boarding {
		return affectedOrdinary, nil
	}
	if all || len(names) != 1 || names[0] != affectedCheckName {
		return affectedOrdinary, fmt.Errorf("local-check run: --boarding qualifies --name %s alone (got --name %s, --all %v)", affectedCheckName, strings.Join(names, ","), all)
	}
	return affectedBoarding, nil
}

// runAffectedCheck is the affected suite (#473): it plans the head's change
// against base and runs the repo's affected command on the selection. A full
// selection runs and records suite instead (D5), never an affected record that
// claims to be full, unless mode is boarding, which refuses (#561); a
// docs-only change records affected without running anything.
func runAffectedCheck(ctx context.Context, d localRunDeps, ad affectedDeps, cfg affected.Config, suite localCheck, sha string, mode affectedExecutionMode) (localRunRecord, error) {
	if cfg.GraphPrepare != nil {
		// The generator precedes runUnderSlot's signal owner. Keep its
		// cancellation alive through graph cleanup, fallback and publication.
		var stop func()
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}
	sel, err := planAffectedSelection(ctx, d, ad, cfg, suite, sha)
	if err != nil {
		return localRunRecord{}, err
	}
	switch {
	case sel.Full && mode == affectedBoarding:
		return localRunRecord{}, boardingRefusal(sha, sel.Reason, suite.Name)
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

// planAffectedSelection plans the head's change against base (#473): what a
// fresh affected run selects, shared by the run and the saved resume (#586).
// A caller whose config declares graph_prepare owns the signal context.
func planAffectedSelection(ctx context.Context, d localRunDeps, ad affectedDeps, cfg affected.Config, suite localCheck, sha string) (affected.Selection, error) {
	paths, err := ad.changed(sha)
	if err != nil {
		return affected.Selection{}, fmt.Errorf("local check %s: changed paths of %s: %w", affectedCheckName, abbrevSHA(sha), err)
	}
	if len(paths) == 0 {
		// No change is not a docs-only change: recording a pass here would be
		// evidence nothing ran for (#473 M1 BR-2).
		return affected.Selection{}, fmt.Errorf("local check %s: %s has no changes against its merge-base with main; there is nothing to vouch for, run --name %s", affectedCheckName, abbrevSHA(sha), suite.Name)
	}
	// Graph preparation can execute even for a docs-only selection, before
	// the selected check reaches the ordinary runner.
	if err := preflightSuite(ctx, d, localCheck{Name: affectedCheckName}, sha); err != nil {
		return affected.Selection{}, err
	}
	// The graph is loaded even for a docs-only diff: a package may embed or
	// test against a doc, and only the graph knows (close review).
	g, graphErr := ad.graph(ctx, sha)
	if err := ctx.Err(); err != nil {
		return affected.Selection{}, fmt.Errorf("local check %s: graph interrupted: %w", affectedCheckName, err)
	}
	if graphErr != nil {
		return affected.Selection{Full: true, Reason: "package graph unavailable: " + graphErr.Error()}, nil
	}
	m, err := ad.mapping(ctx)
	if cerr := ctx.Err(); cerr != nil {
		// A cancelled read is an interruption, never evidence that the declarations are unusable.
		return affected.Selection{}, fmt.Errorf("local check %s: declared ownership interrupted: %w", affectedCheckName, cerr)
	}
	if err != nil {
		return affected.Selection{Full: true, Reason: "declared ownership unavailable: " + err.Error()}, nil
	}
	return affected.Plan(paths, g, cfg, m, affectedIsDoc), nil
}

// affectedIsDoc is the planner's docs rule: a path with no code surface.
func affectedIsDoc(p string) bool { return !publishGateHasCodeSurface([]string{p}) }

// loadAffectedMapping reads the sidecars and traceability files cfg declares from rev's tree in dir's repository (#598), one listing and one
// batch per convention, and compiles the Mapping. rev is the pinned main the config came from, so the head can
// neither declare nor undeclare its own selection; any error is the caller's full fallback.
func loadAffectedMapping(ctx context.Context, dir, rev string, cfg affected.Config) (affected.Mapping, error) {
	var src affected.Sources
	for _, sc := range cfg.Sidecars {
		got, err := gitx.TreeFilesSelected(ctx, dir, rev, strings.TrimSuffix(sc.Prefix, "/"), func(f gitx.TreeFile) bool {
			return strings.HasSuffix(f.Path, sc.Suffix)
		})
		if err != nil {
			return affected.Mapping{}, fmt.Errorf("sidecars under %s on %s: %w", sc.Prefix, abbrevSHA(rev), err)
		}
		for _, f := range got {
			src.Sidecars = append(src.Sidecars, affected.SidecarFile{Path: f.Path, Content: f.Content})
		}
	}
	// One listing and one batch for every traceability file (#616); Compile names a declared file main lacks.
	got, err := gitx.TreeFilesAt(ctx, dir, rev, cfg.Traceability)
	if err != nil {
		return affected.Mapping{}, fmt.Errorf("traceability files on %s: %w", abbrevSHA(rev), err)
	}
	for _, f := range got {
		src.Traceability = append(src.Traceability, affected.TraceabilityFile{Path: f.Path, Content: f.Content})
	}
	return affected.Compile(cfg, src)
}

// boardingRefusal is a full selection in boarding mode (#561): nothing ran
// and nothing was posted. It names the reason, where to repair the selection,
// and the full-proof alternative a car may board on instead.
func boardingRefusal(sha, reason, suite string) error {
	return fmt.Errorf("local check %s on %s: boarding refused: full selection (%s); boarding never runs the full suite.\n"+
		"  → repair the selection: classify the path in %s on main (or fix its graph_prepare / the package graph), push, and rerun boarding; or\n"+
		"  → prove the full suite instead: %s, then sdlc merge --yes (a passing suite also boards).\n"+
		"  %s",
		affectedCheckName, abbrevSHA(sha), reason, affectedConfigPath, localCheckRunCommand(suite, sha), landingline.SerialLandingPrerequisite)
}

// affectedLocalCheck is the selection as a local check: the repo's command
// under the fixed env contract (#473), sorted and space-separated, passed as
// argv through env(1) and never through a shell.
func affectedLocalCheck(cfg affected.Config, sel affected.Selection, suite localCheck) localCheck {
	c := affectedCheckDecl(cfg, suite)
	c.Command = affectedEnvCommand(cfg.Command, sel.GoPkgs, sel.Groups)
	c.Scope = affectedScope(sel)
	c.Rerun = &affectedRerun{Command: cfg.Command, GoPkgs: sel.GoPkgs, Groups: sel.Groups}
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
		mapping: func(ctx context.Context) (affected.Mapping, error) {
			return loadAffectedMapping(ctx, d.Root, base, cfg)
		},
		run: func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			return runLocalCheck(ctx, d, c, sha)
		},
		post: func(ctx context.Context, rec localRunRecord) error { return postLocalCheckRecord(ctx, d, rec) },
	}
}
