// propagatebase.go — `sdlc propagate-base`: re-weave every recursive DEPENDENT of
// the owner repo (the cwd), foundation-first. The downstream counterpart to
// substrateChain (owner→ancestors): this is owner→recursive-dependents. #106.
//
// Per dependent, in topological order: `make weave` (re-weave via build-in-owner) →
// `weave verify-complete` (the gate) → commit the consumption. Push is a separate,
// optional concern (not here). The propagation substance is uniform across every
// dependent (a leaf or a gcrypt brain re-weave the same way); the only wrinkle is
// the RUNNER's sandbox — run out-of-sandbox so a brain's protected .claude/.git are
// writable. Emits ONE status table (operator-attention-shaped).
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
// A sibling qualifies iff it is a git repo carrying a Makefile.workflow (the
// universal "uses the ariadne base layer" signal, per Makefile.local:refresh-recursive)
// AND ownerRoot is in its substrateChain. IO (scans the parent dir + reads deps).
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
		if _, err := os.Stat(filepath.Join(root, "Makefile.workflow")); err != nil {
			continue // not an ariadne-base-layer repo
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
	status string // re-wove+verified+committed | clean (no change) | FAILED: <why>
}

// runPropagateBase orchestrates the propagation. ownerRoot is the owner repo (cwd).
// ref is the commit-message reference (e.g. "ariadne#107"). dryRun reports the plan
// without mutating. Returns an error if any dependent FAILED.
func runPropagateBase(ownerRoot, ref string, dryRun bool, out io.Writer) error {
	weaveBin := filepath.Join(ownerRoot, "bin", "weave")
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
		fmt.Fprintln(out, "(dry-run: would `make weave` + verify-complete + commit each, in order)")
		return nil
	}

	var results []propResult
	failed := false
	for _, d := range deps {
		name := filepath.Base(d.root)
		res := propResult{repo: name}
		switch {
		case run(out, d.root, "make", "weave") != nil:
			res.status = "FAILED: make weave"
		case run(io.Discard, d.root, weaveBin, "verify-complete") != nil:
			res.status = "FAILED: verify-complete (under-production)"
		default:
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
		if strings.HasPrefix(res.status, "FAILED") {
			failed = true
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

// commitConsumption stages all changes in repoRoot and commits them with the
// standard consumption message. Returns (changed, err): changed=false when the
// re-weave produced no tracked diff (idempotent re-run).
func commitConsumption(repoRoot, ref string) (bool, error) {
	// Untrack any file the re-weave just made gitignored — i.e. a file that USED to
	// be tracked but is now a weave-generated artifact the EnsureGitignore covered
	// (e.g. a CLAUDE.md that was a tracked @AGENTS.md bridge and is now generated
	// prose). Without this, `git add -A` would RE-TRACK the generated content (the
	// inert-gitignore trap). `ls-files -i -c` lists tracked-but-ignored files.
	if ignored, err := exec.Command("git", "-C", repoRoot, "ls-files", "-i", "-c", "--exclude-standard").Output(); err == nil {
		for _, f := range strings.Fields(strings.TrimSpace(string(ignored))) {
			_ = exec.Command("git", "-C", repoRoot, "rm", "--cached", "-q", f).Run()
		}
	}
	st, err := exec.Command("git", "-C", repoRoot, "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(st))) == 0 {
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
			"repo), order them foundation-first, then per dependent `make weave` +\n" +
			"verify-complete + commit. Run from the OWNER repo, out-of-sandbox so a\n" +
			"gcrypt brain's protected paths are writable. Push is a separate step.",
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
