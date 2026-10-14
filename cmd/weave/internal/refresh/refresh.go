package refresh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

func discoverRefresh(ctx context.Context, root string, c acquire.Client) (acquire.Result, error) {
	r, e := c.Restore(ctx, root, true)
	if e == nil && len(r.Missing) > 0 {
		e = fmt.Errorf("missing existing dependency checkouts")
	}
	if e != nil {
		return r, fmt.Errorf("read-only dependency discovery failed; reconcile declarations or run ordinary setup first: %w", e)
	}
	return r, nil
}

// declaredOrigins checks each sourced substrate's effective origin. A row
// resolves through layergraph.Resolver, as discovery's Restore did (#230):
// from a linked worktree, an absent row is the primary checkout's sibling.
func declaredOrigins(snapshots []Snapshot, anchors layergraph.Anchors) error {
	resolver := layergraph.NewResolver(layergraph.OSFS{})
	if anchors != nil {
		resolver.Anchors = anchors
	}
	byPath := map[string]Snapshot{}
	for _, s := range snapshots {
		byPath[s.path] = s
	}
	for _, s := range snapshots {
		for _, row := range s.rows {
			if row.Kind != "substrate" || row.Source == "" {
				continue
			}
			res, e := resolver.Substrate(s.path, row.Path)
			if e != nil {
				return e
			}
			p, e := canonical(res.Path)
			if e != nil {
				return e
			}
			want, e := acquire.ResolveSource(row.Source, s.path)
			if e != nil {
				return fmt.Errorf("invalid declared source in %s", s.path)
			}
			got, ok := byPath[p]
			if !ok || got.origin != want.Identity {
				return fmt.Errorf("%s: effective origin does not match its declared source", p)
			}
		}
	}
	return nil
}

// Run refreshes current branches, then compiles once. It never acquires missing
// checkouts or rolls back Git updates. The caller owns any setup lease.
func Run(ctx context.Context, root string, c acquire.Client, rebase bool, out io.Writer, compile func() error) (retErr error) {
	return RunWithOptions(ctx, root, c, Options{Rebase: rebase}, out, compile)
}

// RunWithOptions preserves the complete observation graph, even when the root
// is excluded from the allowed update set. The caller owns the setup lease.
func RunWithOptions(ctx context.Context, root string, c acquire.Client, options Options, out io.Writer, compile func() error) (retErr error) {
	if options.SubstrateOnly && options.Rebase {
		return fmt.Errorf("--substrate and --rebase cannot be combined; run weave refresh --substrate to fast-forward substrates only")
	}
	rebase := options.Rebase
	if out == nil {
		out = io.Discard
	}
	if c.Git == nil {
		c.Git = acquire.ExecGit{Raw: true, Timeout: 2 * time.Minute, MaxOutputBytes: 4 << 20}
	}
	c.MaxLayers = 128
	c.MaxDeclarationBytes = declarationLimit
	root, e := canonical(root)
	if e != nil {
		return e
	}
	phase := inspecting
	step := func(e event) error {
		next, err := advance(phase, e)
		if err == nil {
			phase = next
		}
		return err
	}
	confirmedCount := 0
	defer func() {
		if retErr != nil {
			_ = step(failed)
			fmt.Fprintf(out, "refresh stopped; %d repository updates confirmed; completed Git updates remain\n", confirmedCount)
		}
	}()
	initial, e := discoverRefresh(ctx, root, c)
	if e != nil {
		return e
	}
	var snapshots []Snapshot
	var blockers []error
	for _, p := range initial.Layers {
		s, e := observe(ctx, c.Git, p)
		if e != nil {
			blockers = append(blockers, fmt.Errorf("%s: %w", p, e))
		} else {
			snapshots = append(snapshots, s)
		}
	}
	if len(blockers) > 0 {
		return errors.Join(blockers...)
	}
	if e := declaredOrigins(snapshots, c.Anchors); e != nil {
		return e
	}
	for i := range snapshots {
		s := &snapshots[i]
		if options.SubstrateOnly && s.path == root {
			s.target = s.head
			continue
		}
		// Fetch errors can contain credential-bearing transport URLs. Keep diagnostics
		// to the checkout and operation; never echo subprocess stderr here.
		if _, e := c.Git.Run(ctx, s.path, "fetch", "--no-recurse-submodules", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); e != nil {
			blockers = append(blockers, fmt.Errorf("%s: fetch origin/main failed; check origin access and retry", s.path))
			continue
		}
		s.target, e = oid(ctx, c.Git, s.path, "refs/remotes/origin/main")
		if e != nil {
			blockers = append(blockers, fmt.Errorf("%s: %w", s.path, e))
			continue
		}
		isAncestor, e := ancestor(ctx, c.Git, s.path, s.head, s.target)
		if e != nil {
			blockers = append(blockers, fmt.Errorf("%s: ancestry probe failed: %w", s.path, e))
			continue
		}
		if !eligibility(s.head == s.target, isAncestor, rebase) {
			if options.SubstrateOnly {
				blockers = append(blockers, fmt.Errorf("%s: substrate branch %s at %s is ahead or divergent; ask the operator to reconcile this checkout with origin/main, then retry weave refresh --substrate", s.path, s.branch, s.head))
			} else {
				blockers = append(blockers, fmt.Errorf("%s: local branch is ahead or divergent; reconcile it or explicitly use --rebase", s.path))
			}
		}
		rows, e := targetDeclarations(ctx, c.Git, *s)
		if e != nil {
			blockers = append(blockers, fmt.Errorf("%s: %w", s.path, e))
		} else if !reflect.DeepEqual(rows, s.rows) {
			blockers = append(blockers, fmt.Errorf("%s: target dependency declarations changed; reconcile that change separately", s.path))
		}
		if e := ignoredCollisions(ctx, c.Git, *s, rebase); e != nil {
			blockers = append(blockers, fmt.Errorf("%s: %w", s.path, e))
		}
	}
	if len(blockers) > 0 {
		return errors.Join(blockers...)
	}
	prepared := Prepared{snapshots: snapshots}
	if e := step(checked); e != nil {
		return e
	}
	rediscovered, e := discoverRefresh(ctx, root, c)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(initial.Layers, rediscovered.Layers) || !reflect.DeepEqual(initial.Mounts, rediscovered.Mounts) {
		return fmt.Errorf("dependency graph changed before application; retry")
	}
	for _, s := range prepared.snapshots {
		if e := revalidate(ctx, c.Git, s); e != nil {
			return fmt.Errorf("%s: %w", s.path, e)
		}
	}
	if e := step(validated); e != nil {
		return e
	}
	final := make([]Snapshot, 0, len(snapshots))
	for _, s := range prepared.snapshots {
		if e := ctx.Err(); e != nil {
			return e
		}
		if e := revalidate(ctx, c.Git, s); e != nil {
			return fmt.Errorf("%s: %w", s.path, e)
		}
		if options.SubstrateOnly && s.path == root {
			fmt.Fprintf(out, "%s: consumer branch preserved at %s\n", s.path, s.head)
			final = append(final, s)
			continue
		}
		if e := ignoredCollisions(ctx, c.Git, s, rebase); e != nil {
			return fmt.Errorf("%s: %w", s.path, e)
		}
		fmt.Fprintf(out, "%s: %s -> %s\n", s.path, s.head, s.target)
		if s.head != s.target {
			args := []string{"-c", "submodule.recurse=false", "-c", "merge.autoStash=false", "merge", "--ff-only", "--no-edit", "--no-overwrite-ignore", s.target}
			if rebase {
				args = []string{"-c", "submodule.recurse=false", "-c", "rebase.autoStash=false", "-c", "rebase.updateRefs=false", "-c", "rebase.autoSquash=false", "rebase", "--no-fork-point", "--no-autostash", s.target}
			}
			if _, e := c.Git.Run(ctx, s.path, args...); e != nil {
				fmt.Fprintf(out, "%s: update outcome uncertain; inspect Git status and resolve or abort any conflict explicitly before retry\n", s.path)
				return fmt.Errorf("%s: Git update failed: %w", s.path, e)
			}
		}
		now, e := observe(ctx, c.Git, s.path)
		if e != nil {
			return fmt.Errorf("%s: update not confirmed: %w", s.path, e)
		}
		expected := s
		expected.head = now.head
		ok, e := ancestor(ctx, c.Git, s.path, s.target, now.head)
		if e != nil || !ok || !sameStart(expected, now) || (!rebase && now.head != s.target) {
			return fmt.Errorf("%s: update outcome could not be confirmed", s.path)
		}
		if s.head != s.target {
			confirmedCount++
			fmt.Fprintf(out, "%s: confirmed %s\n", s.path, now.head)
		}
		final = append(final, now)
		if e := step(confirmed); e != nil {
			return e
		}
	}
	for _, s := range final {
		if e := revalidate(ctx, c.Git, s); e != nil {
			return fmt.Errorf("%s: %w", s.path, e)
		}
	}
	current, e := discoverRefresh(ctx, root, c)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(initial.Layers, current.Layers) || !reflect.DeepEqual(initial.Mounts, current.Mounts) {
		return fmt.Errorf("dependency graph changed before compile; reconcile separately")
	}
	// Discovery performs IO; verify its complete subject set once more afterward.
	for _, s := range final {
		if e := revalidate(ctx, c.Git, s); e != nil {
			return fmt.Errorf("%s: %w", s.path, e)
		}
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := step(finished); e != nil {
		return e
	}
	if options.SubstrateOnly && len(snapshots) == 1 {
		fmt.Fprintln(out, "no substrate Git checkouts to refresh; compiling the current source branch (integrate main separately to update source-owned helpers)")
	}
	if compile == nil {
		return fmt.Errorf("refresh requires a compile callback")
	}
	if e := compile(); e != nil {
		return fmt.Errorf("compile failed; Git updates remain: %w", e)
	}
	return step(compiled)
}
