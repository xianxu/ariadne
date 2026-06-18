package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// prune.go is weave's lowered-symlink garbage collector (#96): on apply, weave
// removes ORPHANED lowered symlinks it no longer produces — covering a renamed
// or re-prefixed skill (.claude/skills/<old-prefix><name> left behind) AND the
// #95 cutover's now-DEAD symlinks (construct/setup.sh, .../merge-settings.sh,
// .../sync-local-skills.sh → ariadne scripts that were deleted, dangling in
// every derivative).
//
// THIS DELETES FILES — safety is the whole point. A prune may ONLY remove an
// entry that is ALL of (any failure ⇒ KEEP; when uncertain ⇒ KEEP):
//
//  1. a SYMLINK (never a real file or real dir — a repo's own authored content
//     is sacrosanct);
//  2. located in a weave-MANAGED lowered location — a directory weave produced
//     at least one symlink into THIS run (derived from the produced Symlink
//     actions, never hardcoded);
//  3. weave-OWNED — its target (resolved LEXICALLY against the link's dir, so a
//     dangling link still resolves) points into a lowering SOURCE ROOT (a layer
//     root weave's produced symlinks point into), i.e. it looks exactly like
//     something weave lowers;
//  4. NOT in the set weave produced THIS run (the orphan condition).
//
// The pure decision (shouldPrune + the producedSet/managed-location/source-root
// derivations) lives here as string-in/out functions (ARCH-PURE); the actual
// scan + unlink is the IO seam (PruneOrphans), mirroring gitignore.go's
// pure-transform + IO-seam split. The compile lowering (main.run) calls
// PruneOrphans after plan.Apply.

// PruneCandidate is one observed symlink in a managed location, captured by the
// IO scan and handed to the pure shouldPrune. RelPath is the repo-relative path
// of the link (e.g. ".claude/skills/xx-old"); ResolvedTarget is its target made
// absolute by resolving the raw link text LEXICALLY against the link's parent
// dir (filepath.Join(dir, raw) then Clean) — NOT EvalSymlinks, so a DANGLING
// link (its target deleted) still yields the path it WOULD point at. IsSymlink
// records that the IO scan saw a symlink (a real file/dir is never made a
// candidate, but the pure fn re-asserts it as a belt-and-suspenders guard).
type PruneCandidate struct {
	RelPath        string
	ResolvedTarget string
	IsSymlink      bool
}

// ProducedPathSet returns the repo-relative target path of EVERY action weave
// produced this run — the full "weave owns this slot this run" set the orphan
// exclusion (criterion 4) tests against. Broader than ProducedSymlinkSet (which
// is Symlink-only, for the managed-location derivation): a path weave writes as
// a REGULAR file (WriteFile AGENTS.md), seeds, touches, scaffolds, or merges is
// not an orphan and must never be pruned — even while it still occupies the slot
// as the pre-cutover symlink at dry-run-preview time (before Apply rewrites it).
// Without this, `weave compile --dry-run` on an un-woven derivative falsely
// previews `prune AGENTS.md` (the AGENTS.md → ancestor symlink looks orphaned),
// though a real apply never prunes it (Apply converts it to a regular file
// first). Pure.
func ProducedPathSet(actions []Action) map[string]bool {
	set := map[string]bool{}
	for _, a := range actions {
		switch act := a.(type) {
		case Symlink:
			set[filepath.Clean(act.Dst)] = true
		case WriteFile:
			set[filepath.Clean(act.Path)] = true
		case Seed:
			set[filepath.Clean(act.Dst)] = true
		case Touch:
			set[filepath.Clean(act.Path)] = true
		case Mkdir:
			set[filepath.Clean(act.Path)] = true
		case MergeSettings:
			set[filepath.Clean(act.Target)] = true
		}
	}
	return set
}

// ManagedLocations returns the SORTED set of repo-relative directories weave
// produced at least one Symlink into this run (criterion 2: the managed lowered
// locations). A location is managed IFF weave emitted a symlink there — so on a
// self-walk that owns construct/scripts/ as real files, that dir is NOT managed
// and is never scanned. Derived purely from the produced actions, never
// hardcoded. Pure.
func ManagedLocations(actions []Action) []string {
	set := map[string]bool{}
	for _, a := range actions {
		if s, ok := a.(Symlink); ok {
			set[filepath.Dir(filepath.Clean(s.Dst))] = true
		}
	}
	out := make([]string, 0, len(set))
	for d := range set {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

// shouldPrune is the pure safety decision: given one observed symlink candidate,
// the produced-this-run set, and the lowering source roots (absolute layer
// roots weave lowers FROM), report whether it is a weave-owned orphan safe to
// delete. ALL must hold (else KEEP):
//
//   - c.IsSymlink (a real file/dir is never a candidate, but re-assert it);
//   - c.RelPath NOT in producedSet (the orphan condition — a produced symlink is
//     KEPT);
//   - weave-owned: c.ResolvedTarget is within some sourceRoot (the lexical
//     target points into the substrate/ancestor graph weave lowers from — true
//     for a live OR a dangling weave link; FALSE for a repo's own symlink
//     pointing somewhere unrelated).
//
// Pure (no IO).
func shouldPrune(c PruneCandidate, producedSet map[string]bool, sourceRoots []string) bool {
	if !c.IsSymlink {
		return false // criterion 1: never a real file/dir
	}
	if producedSet[filepath.Clean(c.RelPath)] {
		return false // criterion 4: produced this run ⇒ KEEP
	}
	return targetWithinAnyRoot(c.ResolvedTarget, sourceRoots) // criterion 3: weave-owned
}

// targetWithinAnyRoot reports whether the absolute target lies within (equals or
// is under) any of the given absolute roots. It uses filepath.Rel and rejects
// any path that escapes the root (rel == ".." or a "../" prefix), so a sibling
// dir that merely shares a string prefix is NOT counted as within. Pure.
func targetWithinAnyRoot(target string, roots []string) bool {
	target = filepath.Clean(target)
	dotdot := ".." + string(filepath.Separator)
	for _, root := range roots {
		rel, err := filepath.Rel(filepath.Clean(root), target)
		if err != nil {
			continue
		}
		// rel == "." (equals root) or any path NOT escaping upward ⇒ within.
		if rel == ".." || rel == "" {
			continue
		}
		if len(rel) >= len(dotdot) && rel[:len(dotdot)] == dotdot {
			continue // target escapes this root
		}
		return true
	}
	return false
}

// SourceRootsFromPaths returns the SORTED, de-duplicated set of absolute layer
// roots — the lowering source graph the weave-owned check (criterion 3) tests
// target containment against. The production call sites (main.run / the golden
// + dry-run pipelines) hold the resolved []layer.Layer and pass each layer's
// Path here; a weave-lowered symlink's target always resolves under one of
// these roots (the self root on a self-walk, an ancestor/substrate root in a
// derivative), while a repo's own unrelated symlink does not. Pure.
func SourceRootsFromPaths(layerPaths []string) []string {
	set := map[string]bool{}
	for _, p := range layerPaths {
		if p == "" {
			continue
		}
		set[filepath.Clean(p)] = true
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// PrunePlan is the pure plan of what PruneOrphans would do: the SORTED
// repo-relative paths to unlink. Computed from the scanned candidates + the
// produced set + source roots, so the dry-run and the apply share ONE decision
// (ARCH-DRY). Pure.
func PrunePlan(candidates []PruneCandidate, producedSet map[string]bool, sourceRoots []string) []string {
	var out []string
	for _, c := range candidates {
		if shouldPrune(c, producedSet, sourceRoots) {
			out = append(out, c.RelPath)
		}
	}
	sort.Strings(out)
	return out
}

// ScanManagedSymlinks is the read-only IO half of the prune: for each managed
// location it lists the directory, Lstat's each entry, and records ONLY the
// symlinks as PruneCandidates (a real file/dir is skipped — never a candidate,
// so a repo's authored content can't even reach the decision). Each candidate's
// ResolvedTarget is the raw link text resolved LEXICALLY against the link's dir
// (handles dangling links). Read-only; the decision (PrunePlan) and the unlink
// (PruneOrphans) are separate. A managed location that does not exist is
// skipped (no error).
func ScanManagedSymlinks(fs weavefs.FS, repoRoot string, managed []string) ([]PruneCandidate, error) {
	var out []PruneCandidate
	for _, loc := range managed {
		absDir := filepath.Join(repoRoot, loc)
		entries, err := fs.ReadDir(absDir)
		if err != nil {
			continue // location absent (or unreadable) ⇒ nothing to scan there
		}
		for _, e := range entries {
			absEntry := filepath.Join(absDir, e.Name())
			fi, lerr := fs.Lstat(absEntry)
			if lerr != nil {
				continue
			}
			if fi.Mode()&os.ModeSymlink == 0 {
				continue // real file or real dir — never a candidate (criterion 1)
			}
			raw, rerr := fs.Readlink(absEntry)
			if rerr != nil {
				continue // can't read the link ⇒ leave it alone (uncertain ⇒ KEEP)
			}
			resolved := raw
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(absDir, raw)
			}
			out = append(out, PruneCandidate{
				RelPath:        filepath.Join(loc, e.Name()),
				ResolvedTarget: filepath.Clean(resolved),
				IsSymlink:      true,
			})
		}
	}
	return out, nil
}

// PruneOrphans is the prune's IO seam. It takes TWO action sets (Option B #107
// cross-target prune): scanActions decides WHICH locations to scan (the UNION of
// every harness face — so a lean `--target X` compile still SCANS the dirs the
// OTHER faces own), while producedActions is what the CURRENT compile actually
// emitted (the orphan condition tests against this). So a lean codex compile scans
// .claude/skills (claude's face, in the union) but produced nothing there → those
// links are orphans → pruned (BIDIRECTIONAL: a claude compile likewise prunes
// .agents/skills). The Union compile passes scanActions == producedActions, so it
// prunes nothing extra. sourceRoots are the absolute layer roots weave lowers from
// (the weave-owned graph). Idempotent. Returns the SORTED repo-relative paths
// removed; the safety criteria (shouldPrune) gate every removal — no new delete
// logic, no per-target registry (ARCH-DRY).
func PruneOrphans(fs weavefs.FS, repoRoot string, scanActions, producedActions []Action, sourceRoots []string) ([]string, error) {
	candidates, err := ScanManagedSymlinks(fs, repoRoot, ManagedLocations(scanActions))
	if err != nil {
		return nil, err
	}
	toPrune := PrunePlan(candidates, ProducedPathSet(producedActions), sourceRoots)
	for _, rel := range toPrune {
		if err := fs.Remove(filepath.Join(repoRoot, rel)); err != nil {
			return nil, fmt.Errorf("prune orphan symlink %s: %w", rel, err)
		}
	}
	return toPrune, nil
}

// PrunePreview is the read-only twin of PruneOrphans for --dry-run: same
// two-action-set scan + the SAME pure decision, NO unlink. Strictly read-only.
func PrunePreview(fs weavefs.FS, repoRoot string, scanActions, producedActions []Action, sourceRoots []string) ([]string, error) {
	candidates, err := ScanManagedSymlinks(fs, repoRoot, ManagedLocations(scanActions))
	if err != nil {
		return nil, err
	}
	return PrunePlan(candidates, ProducedPathSet(producedActions), sourceRoots), nil
}

// generatedRel is the repo-relative root of the per-repo dynamic-skill
// materialization tree (#115 M3). Every materialized body lives at
// construct/generated/<dir>/SKILL.md; the prune below GCs an entire
// construct/generated/<dir> the current compile no longer produces.
const generatedRel = "construct/generated"

// shouldPruneGenerated is the PURE decision for the generated-class GC: a
// construct/generated/<dir> entry is an orphan IFF <dir> is NOT in this run's
// produced dynamic set (the bare dirs DynamicSkills produced). Removing the marker
// (so a dynamic skill is dropped) leaves its stale materialized dir behind; this
// reclaims it. Scoped to construct/generated by construction — the caller only ever
// hands it children of that dir — so nothing outside is reachable. Pure.
func shouldPruneGenerated(dir string, producedDirs map[string]bool) bool {
	return !producedDirs[dir]
}

// ProducedGeneratedDirs is the set of bare <dir>s the current compile materialized
// under construct/generated (the produced dynamic-skill dirs). Pure helper over the
// dir names so the GC's keep/remove decision is testable without the walk.
func ProducedGeneratedDirs(dirs []string) map[string]bool {
	set := map[string]bool{}
	for _, d := range dirs {
		set[d] = true
	}
	return set
}

// PruneGenerated is the IO seam for the generated-class GC (#115 M3): it lists
// repoRoot/construct/generated and removes each child dir whose name is NOT in the
// produced set (shouldPruneGenerated). It NEVER touches anything outside
// construct/generated/ — it only ever lists that one dir and removes its orphaned
// children (a whole construct/generated/<gone> subtree). An absent
// construct/generated (no dynamic skills, or a fresh clone) is a no-op (not an
// error). Returns the SORTED repo-relative paths removed. RemoveAll deletes the
// subtree (the materialized SKILL.md + its dir). Mirrors PruneOrphans' scan→decide→
// remove split; the decision (shouldPruneGenerated) is pure.
func PruneGenerated(fs weavefs.FS, repoRoot string, producedDirs map[string]bool) ([]string, error) {
	genRoot := filepath.Join(repoRoot, generatedRel)
	entries, err := fs.ReadDir(genRoot)
	if err != nil {
		return nil, nil // absent / unreadable ⇒ nothing to GC
	}
	var removed []string
	for _, e := range entries {
		if !e.IsDir() {
			continue // only <dir>/ children are managed here
		}
		if shouldPruneGenerated(e.Name(), producedDirs) {
			rel := filepath.Join(generatedRel, e.Name())
			if err := fs.RemoveAll(filepath.Join(repoRoot, rel)); err != nil {
				return nil, fmt.Errorf("prune generated %s: %w", rel, err)
			}
			removed = append(removed, rel)
		}
	}
	sort.Strings(removed)
	return removed, nil
}
