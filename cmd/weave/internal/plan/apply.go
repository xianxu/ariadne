package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Apply executes a []Action against fs, idempotently. It is the ONLY mutating
// code in weave (ARCH-PURE: the planner computes Actions; this seam runs them).
// repoRoot is the consuming repo's absolute root; every Action's repo-relative
// path (WriteFile.Path, Mkdir.Path, Symlink.Dst) is resolved against it here —
// the planner deliberately leaves them relative (pure string joins) so this
// IO seam owns the abs-path resolution. Symlink.Src is already absolute (the
// walk supplies each layer's absolute Path).
//
// Behaviors are ported from setup.sh (ARCH-DRY); the part-3 golden-diff checks
// parity:
//   - Symlink → create_symlink: a RELATIVE link target computed from the
//     destination's dir (so the repo can move), replacing an existing symlink
//     (rm + relink) or a regular file/dir (rm -rf) occupying the slot, and a
//     no-op when the link already points where it should.
//   - Mkdir → create_scaffold: mkdir -p, no-op when the dir already exists.
//   - WriteFile → seed/AGENTS.md/touch: ensure parents, then write.
func Apply(fs weavefs.FS, repoRoot string, actions []Action) error {
	for _, a := range actions {
		var err error
		switch act := a.(type) {
		case Symlink:
			err = applySymlink(fs, filepath.Join(repoRoot, act.Dst), act.Src)
		case Mkdir:
			err = applyMkdir(fs, filepath.Join(repoRoot, act.Path))
		case Touch:
			err = applyTouch(fs, filepath.Join(repoRoot, act.Path))
		case WriteFile:
			err = applyWriteFile(fs, filepath.Join(repoRoot, act.Path), act.Content)
		case MergeSettings, ToolDep:
			// Deferred lowerings emit no concrete file-op yet (M4 settings, the
			// exec seam). Skip — never reached today, but keeps Apply total.
		default:
			err = fmt.Errorf("apply: unknown action type %T", a)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// applySymlink ports create_symlink. src is the absolute upstream path; dst the
// absolute destination in the target repo. The link target is RELATIVE
// (rel_path(src, dirname(dst)) = filepath.Rel(dir(dst), src)) so the repo
// survives a move, matching setup.sh.
func applySymlink(fs weavefs.FS, dst, src string) error {
	if err := ensureParent(fs, dst); err != nil {
		return err
	}
	rel, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		return fmt.Errorf("apply symlink: relpath %s from %s: %w", src, filepath.Dir(dst), err)
	}

	// Idempotency: inspect what currently occupies the slot.
	if fi, lerr := fs.Lstat(dst); lerr == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			// Existing symlink: no-op if already correct, else replace.
			if existing, rerr := fs.Readlink(dst); rerr == nil && existing == rel {
				return nil // already correct ([[ "$existing" == "$rel" ]] → return 0)
			}
			if err := fs.Remove(dst); err != nil {
				return fmt.Errorf("apply symlink: remove stale link %s: %w", dst, err)
			}
		} else {
			// Regular file/dir in the slot: rm -rf, then relink.
			if err := fs.RemoveAll(dst); err != nil {
				return fmt.Errorf("apply symlink: rm -rf %s: %w", dst, err)
			}
		}
	}

	if err := fs.Symlink(rel, dst); err != nil {
		return fmt.Errorf("apply symlink: link %s -> %s: %w", dst, rel, err)
	}
	return nil
}

// applyMkdir ports create_scaffold: mkdir -p, idempotent (MkdirAll is a no-op
// when the dir exists). setup.sh also drops a .gitkeep; that is a part-3
// golden-diff detail — left as a TODO so this seam stays minimal for M2.
// TODO(part-3): create_scaffold also `touch "$dir/.gitkeep"` for git-tracking
// an otherwise-empty dir; add when the golden-diff requires parity.
func applyMkdir(fs weavefs.FS, dir string) error {
	if err := fs.MkdirAll(dir); err != nil {
		return fmt.Errorf("apply mkdir: %s: %w", dir, err)
	}
	return nil
}

// applyTouch ports setup.sh's `touch` case (line 347): ensure parents, then
// create an EMPTY file ONLY if it does not already exist. Crucially does NOT
// overwrite an existing file — a Touch target (e.g. workshop/lessons.md)
// accumulates content over time and must survive a re-weave. Idempotent.
func applyTouch(fs weavefs.FS, path string) error {
	if err := ensureParent(fs, path); err != nil {
		return err
	}
	if _, err := fs.Lstat(path); err == nil {
		return nil // already exists (with any content) — no-op, never clobber
	}
	if err := fs.WriteFile(path, []byte{}); err != nil {
		return fmt.Errorf("apply touch: %s: %w", path, err)
	}
	return nil
}

// applyWriteFile ensures parents then writes content (the composed AGENTS.md or
// a seed). Overwrites unconditionally — the planner decides content;
// convergence-on-drift is implicit (same content → same bytes).
func applyWriteFile(fs weavefs.FS, path, content string) error {
	if err := ensureParent(fs, path); err != nil {
		return err
	}
	if err := fs.WriteFile(path, []byte(content)); err != nil {
		return fmt.Errorf("apply writefile: %s: %w", path, err)
	}
	return nil
}

// ensureParent ports setup.sh's ensure_parent: mkdir -p the dir holding path.
func ensureParent(fs weavefs.FS, path string) error {
	if err := fs.MkdirAll(filepath.Dir(path)); err != nil {
		return fmt.Errorf("ensure parent of %s: %w", path, err)
	}
	return nil
}
