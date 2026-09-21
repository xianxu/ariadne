// propagatebase.go — `sdlc propagate-base`: re-weave every recursive DEPENDENT of
// the owner repo (the cwd), foundation-first. The downstream counterpart to
// substrateChain (owner→ancestors): this is owner→recursive-dependents. #106.
//
// Per dependent, in topological order: a clean-working-tree precheck
// (workingTreeDirty) → `weave compile` (prepare and re-weave) → `weave
// verify-complete` (the gate) → commit the consumption. A dependent with
// pre-existing uncommitted work (e.g. a concurrent agent session in a sibling repo)
// is SKIPPED untouched — never `git add -A`'d — and the run exits non-zero (#109).
// Push is a separate, optional concern (not here). The propagation substance is
// uniform across every dependent (a leaf or a gcrypt brain re-weave the same way);
// the only wrinkle is the RUNNER's sandbox — run out-of-sandbox so a brain's
// protected .claude/.git are writable. Emits ONE status table (operator-attention-shaped).
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/pkg/weaveownership"
)

// propDep is one recursive dependent: its repo root + its substrate chain (the
// canonical ancestor roots, from substrateChain).
type propDep struct {
	root  string
	chain []string
}

// canonRoot canonicalizes a path for stable identity comparison (EvalSymlinks when
// it exists, else Abs+Clean), matching substrateChain's keying.
func canonRoot(p string) string {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real
	}
	if abs, err := filepath.Abs(p); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(p)
}

// recursiveDependents finds the present sibling repos whose substrate chain
// transitively includes ownerRoot — the repos that consume ownerRoot's base layer.
// A sibling qualifies iff it is a Git repo whose declared substrate chain
// contains ownerRoot. Generated workflow files need not exist yet.
// IO: scans the parent directory and reads dependency declarations.
func recursiveDependents(ownerRoot string) []propDep {
	ownerKey := canonRoot(ownerRoot)
	parent := filepath.Dir(ownerRoot)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}
	var deps []propDep
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		root := filepath.Join(parent, e.Name())
		if canonRoot(root) == ownerKey {
			continue // skip self
		}
		if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
			continue // not a git repo (e.g. a setup.sh-era scratch dir) — can't commit
		}
		chain := substrateChain(root)
		for _, anc := range chain {
			if canonRoot(anc) == ownerKey {
				deps = append(deps, propDep{root: root, chain: chain})
				break
			}
		}
	}
	return deps
}

// orderDependentsFoundationFirst returns the dependents in topological order: a
// dependent that is itself in ANOTHER dependent's chain comes first (so nous, on
// which the brains depend, propagates before the brains). Pure: rank d = how many
// OTHER dependents appear in d's chain; sort ascending (ties → root path, stable).
// Valid because if X is in Y's chain, X's chain ⊊ Y's chain ⇒ rank(X) < rank(Y).
func orderDependentsFoundationFirst(deps []propDep) []propDep {
	depSet := map[string]bool{}
	for _, d := range deps {
		depSet[canonRoot(d.root)] = true
	}
	rank := func(d propDep) int {
		n := 0
		for _, anc := range d.chain {
			if depSet[canonRoot(anc)] {
				n++
			}
		}
		return n
	}
	out := append([]propDep(nil), deps...)
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := rank(out[i]), rank(out[j])
		if ri != rj {
			return ri < rj
		}
		return out[i].root < out[j].root
	})
	return out
}

// propResult is one dependent's outcome (for the status table).
type propResult struct {
	repo   string
	status string // re-wove+verified+committed | re-wove+verified (no change) | SKIPPED: dirty working tree | FAILED: <why>
}

// runPropagateBase orchestrates the propagation. ownerRoot is the owner repo (cwd).
// ref is the commit-message reference (e.g. "ariadne#107"). dryRun reports the plan
// without mutating. Returns an error if any dependent FAILED.
func runPropagateBase(ownerRoot, ref string, dryRun bool, out io.Writer) error {
	deps := orderDependentsFoundationFirst(recursiveDependents(ownerRoot))
	if len(deps) == 0 {
		fmt.Fprintln(out, "propagate-base: no recursive dependents found")
		return nil
	}
	fmt.Fprintf(out, "propagate-base: %d dependent(s), foundation-first:\n", len(deps))
	for i, d := range deps {
		fmt.Fprintf(out, "  %d. %s\n", i+1, filepath.Base(d.root))
	}
	if dryRun {
		fmt.Fprintln(out, "(dry-run: would `weave compile` + verify-complete + commit each, in order)")
		return nil
	}

	// Resolve the installed gateway only when the first clean dependent needs it.
	// Empty, preview-only, and all-dirty runs need no executable.
	var weaveBin string
	compile := func(root string) error {
		if weaveBin == "" {
			var err error
			weaveBin, err = exec.LookPath("weave")
			if err != nil {
				return fmt.Errorf("find installed weave: %w", err)
			}
			weaveBin, err = filepath.Abs(weaveBin)
			if err != nil {
				return err
			}
		}
		return run(out, root, weaveBin, "compile")
	}
	var results []propResult
	failed := false
	skipped := 0
	for _, d := range deps {
		name := filepath.Base(d.root)
		res := propResult{repo: name}
		dirty, derr := workingTreeDirty(d.root)
		switch {
		case derr != nil:
			res.status = "FAILED: " + derr.Error()
		// A dirty dependent has pre-existing in-flight work (e.g. a concurrent
		// session). NEVER `git add -A` it — skip untouched + report. The operator
		// commits/stashes that work and re-runs (the re-weave is idempotent).
		case dirty:
			res.status = "SKIPPED: dirty working tree (pre-existing uncommitted work — commit/stash + re-run)"
		default:
			if err := compile(d.root); err != nil {
				res.status = "FAILED: weave compile: " + err.Error()
				break
			}
			if err := run(out, d.root, weaveBin, "verify-complete"); err != nil {
				res.status = "FAILED: verify-complete: " + err.Error()
				break
			}
			changed, cerr := commitConsumption(d.root, ref)
			switch {
			case cerr != nil:
				res.status = "FAILED: commit — " + cerr.Error()
			case changed:
				res.status = "re-wove + verified + committed"
			default:
				res.status = "re-wove + verified (no change)"
			}
		}
		switch {
		case strings.HasPrefix(res.status, "FAILED"):
			failed = true
		case strings.HasPrefix(res.status, "SKIPPED"):
			skipped++
		}
		results = append(results, res)
	}

	fmt.Fprintln(out, "\n── propagate-base summary ──")
	for _, r := range results {
		fmt.Fprintf(out, "  %-16s %s\n", r.repo, r.status)
	}
	if failed {
		return fmt.Errorf("propagate-base: one or more dependents FAILED — see summary")
	}
	// A skipped dependent is left STALE (didn't get the new base), so an incomplete
	// propagation must not read as success — exit non-zero, distinct from FAILED.
	if skipped > 0 {
		return fmt.Errorf("propagate-base: %d dependent(s) SKIPPED (dirty working tree) — commit/stash their work and re-run", skipped)
	}
	return nil
}

// run executes a command in dir, streaming output to w; returns its error.
func run(w io.Writer, dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

// gitStatusPorcelain returns `git status --porcelain` (trimmed) for repoRoot — the
// shared porcelain read behind BOTH the clean-tree precheck (workingTreeDirty) and
// commitConsumption's nothing-to-commit check (ARCH-DRY). Gitignored paths are
// excluded by default, so weave's generated output (CLAUDE.md, .claude/skills, …)
// never reads as a change.
//
// --untracked-files=all is PINNED (not left to config): a `status.showUntrackedFiles=no`
// gitconfig would otherwise hide untracked files, blinding the precheck to exactly the
// untracked concurrent-session file it exists to catch (#109 review). Matches the
// sibling untracked-detection-critical path in push.go.
func gitStatusPorcelain(repoRoot string) (string, error) {
	out, err := exec.Command("git", "-C", repoRoot, "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// workingTreeDirty reports whether repoRoot has pre-existing uncommitted/untracked
// work (tracked modifications, staged changes, or untracked non-ignored files).
// propagate-base checks this BEFORE re-weaving so it never sweeps a dependent's
// unrelated in-flight work into a consumption commit via `git add -A` — the
// concurrent-session hazard: a sibling repo being edited by ANOTHER agent. Because
// the woven output is gitignored, a previously-propagated CLEAN dependent reads as
// not-dirty, so a clean-before gate is exact (any post-re-weave delta is then the
// re-weave's OWN output).
func workingTreeDirty(repoRoot string) (bool, error) {
	st, err := gitStatusPorcelain(repoRoot)
	if err != nil {
		return false, err
	}
	return st != "", nil
}

// commitConsumption stages all changes in repoRoot and commits them with the
// standard consumption message. Returns (changed, err): changed=false when the
// re-weave produced no tracked diff (idempotent re-run).
//
// PRECONDITION: the caller (runPropagateBase) verified the tree was CLEAN before
// re-weaving (workingTreeDirty), so every change `git add -A` stages here is the
// re-weave's OWN output — never a concurrent session's unrelated in-flight work.
func commitConsumption(repoRoot, ref string) (changed bool, err error) {
	// A failed Git command may already have changed the index or even committed.
	// Preserve Git's actual state; never infer rollback from an error. The public
	// propagation precheck will reject dirty retries until the operator resolves
	// them, while a completed commit naturally becomes a no-op on retry.
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; index changes were not reset: inspect git status and git diff --cached, resolve or commit retained changes, then retry propagation", err)
		}
	}()
	// Ignored does not mean generated: only current identities recorded by weave
	// authorize untracking. Local negations remain authoritative because Git
	// supplies the ignored set. Read and validate everything before index edits.
	owned, err := weaveownership.MatchingPaths(weaveownership.OSReader{}, repoRoot)
	if err != nil {
		return false, fmt.Errorf("read weave ownership: %w", err)
	}
	ignored, err := exec.Command("git", "-C", repoRoot, "ls-files", "-z", "-i", "-c", "--exclude-standard").Output()
	if err != nil {
		return false, fmt.Errorf("list tracked ignored files: %w", err)
	}
	ignoredSet := map[string]bool{}
	for _, path := range strings.Split(string(ignored), "\x00") {
		if path != "" {
			ignoredSet[path] = true
		}
	}
	for _, path := range owned {
		if !ignoredSet[path] {
			continue
		}
		if err := exec.Command("git", "--literal-pathspecs", "-C", repoRoot, "rm", "--cached", "-q", "--", path).Run(); err != nil {
			return false, fmt.Errorf("untrack owned generated file %s: %w", path, err)
		}
	}
	st, err := gitStatusPorcelain(repoRoot)
	if err != nil {
		return false, err
	}
	if st == "" {
		return false, nil // nothing to commit
	}
	if err := exec.Command("git", "-C", repoRoot, "add", "-A").Run(); err != nil {
		return false, fmt.Errorf("git add: %w", err)
	}
	msg := fmt.Sprintf("%s: consume base-layer change (propagate-base)", ref)
	if err := exec.Command("git", "-C", repoRoot, "commit", "-q", "-m", msg).Run(); err != nil {
		return false, fmt.Errorf("git commit: %w", err)
	}
	return true, nil
}

func newPropagateBaseCmd() *cobra.Command {
	var ref string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "propagate-base",
		Short: "Re-weave every recursive dependent of this repo (foundation-first)",
		Long: "Propagate THIS repo's base-layer change to all recursive dependents:\n" +
			"discover the dependents (siblings whose substrate chain includes this\n" +
			"repo), order them foundation-first, then per dependent `weave compile` +\n" +
			"verify-complete + commit. A dependent with a DIRTY working tree (pre-existing\n" +
			"uncommitted work — e.g. a concurrent session) is SKIPPED untouched and the\n" +
			"run exits non-zero; commit/stash there and re-run. Run from the OWNER repo,\n" +
			"out-of-sandbox so a gcrypt brain's protected paths are writable. Push is a\n" +
			"separate step.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return runPropagateBase(canonRoot(cwd), ref, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&ref, "ref", "", "commit-message reference for the consumption (e.g. ariadne#107)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the dependents + order; mutate nothing")
	return cmd
}
