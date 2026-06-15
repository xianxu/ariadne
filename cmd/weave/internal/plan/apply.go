package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/settingsx"
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
//   - Seed → create_seed: a content-tracking real-file copy — create the target
//     (copy the upstream source bytes) when missing, refresh it when its content
//     drifted from the source, a no-op when already identical, and a non-fatal
//     skip when the source is absent. Distinct from WriteFile (whose content the
//     planner holds): a Seed's content is read from Src here in the IO seam.
//   - WriteFile → AGENTS.md/touch: ensure parents, then write.
//   - MergeSettings → merge-settings.sh: read base + optional sibling
//     settings.local.json, run the pure mergeSettings, write the merged target.
//   - EnsureGitignore → the generated-runtime ignore mechanism (gitignore.go):
//     read the repo's .gitignore, append the missing fixed entries (idempotent
//     whole-line append, never duplicating), write back only on change. weave
//     OWNS this because weave generates those artifacts; emitted once per compile
//     so every derivative gets a clean `git status` with no per-repo hand-edit.
//
// The retired `tool` verb (#95 M5) has no Action and no IO here: Go-tool
// ownership is location-based (construct/dev-aliases.sh scans sibling cmd/X dirs)
// and deps come from `weave link` / construct/deps, so weave never edits go.mod.
func Apply(fs weavefs.FS, repoRoot string, actions []Action) error {
	for _, a := range actions {
		var err error
		switch act := a.(type) {
		case Symlink:
			err = applySymlink(fs, filepath.Join(repoRoot, act.Dst), act.Src)
		case Mkdir:
			err = applyMkdir(fs, filepath.Join(repoRoot, act.Path))
		case Seed:
			err = applySeed(fs, act.Src, filepath.Join(repoRoot, act.Dst))
		case Touch:
			err = applyTouch(fs, filepath.Join(repoRoot, act.Path))
		case WriteFile:
			err = applyWriteFile(fs, filepath.Join(repoRoot, act.Path), act.Content)
		case MergeSettings:
			err = applyMergeSettings(fs, repoRoot, act)
		case EnsureGitignore:
			err = applyEnsureGitignore(fs, filepath.Join(repoRoot, ".gitignore"), act.Entries)
		default:
			err = fmt.Errorf("apply: unknown action type %T", a)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// applyMergeSettings is the IO half of the settings cascade, ported from
// merge-settings.sh: read the base (act.Source) and the optional sibling local
// (settings.local.json, alongside act.Target), run the pure settingsx.Merge, and
// write the result to act.Target. The local file's path is derived the same way
// the bash does — LOCAL_FILE="$TARGET_DIR/settings.local.json", i.e. the
// settings.local.json sibling of the target — so an arbitrary Target dir resolves
// its local correctly. A missing base is an error (the bash's `[[ ! -f BASE ]]`
// exit 1); a missing local takes the local-absent path (base with meta stripped).
// All IO lives here (ARCH-PURE); the merge itself is the pure settingsx.Merge.
func applyMergeSettings(fs weavefs.FS, repoRoot string, act MergeSettings) error {
	basePath := filepath.Join(repoRoot, act.Source)
	base, err := fs.ReadFile(basePath)
	if err != nil {
		return fmt.Errorf("apply merge: read base %s: %w", basePath, err)
	}

	targetPath := filepath.Join(repoRoot, act.Target)
	localPath := filepath.Join(filepath.Dir(targetPath), "settings.local.json")
	var local []byte
	if data, lerr := fs.ReadFile(localPath); lerr == nil {
		local = data // present ⇒ deep-merge; absent ⇒ nil ⇒ base-with-meta-stripped
	}

	merged, err := settingsx.Merge(base, local)
	if err != nil {
		return fmt.Errorf("apply merge: %s: %w", targetPath, err)
	}
	if err := ensureParent(fs, targetPath); err != nil {
		return err
	}
	if err := fs.WriteFile(targetPath, merged); err != nil {
		return fmt.Errorf("apply merge: write %s: %w", targetPath, err)
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

// applySeed ports create_seed: a content-tracking real-file copy of the
// upstream src into dst (dst already absolute). src is the absolute upstream
// path. Behaviors, verbatim from setup.sh:
//
//   - Missing src → non-fatal skip (the bash `[[ ! -f "$src" ]]` warn + return
//     0). weave can't read the source, so it leaves the target intact and does
//     NOT error the walk. A read failure (absent or unreadable) takes this path.
//   - Existing dst with identical content → silent no-op (the `cmp -s` guard),
//     so a re-weave produces no churn.
//   - Otherwise (dst absent, or drifted from src) → ensure parents, then write
//     src's bytes to dst (created on first run, refreshed when it drifted). This
//     is the convergence #45 added: a derivative stranded on a stale entrypoint
//     catches up to upstream.
//
// NOTE on mode: setup.sh uses `cp -p` to preserve the source's mode (an
// executable source lands executable). weavefs.FS.WriteFile writes a fixed
// 0o644; applySeed then replicates the load-bearing part of `cp -p` by
// OBSERVING the source's mode (fs.Stat) and chmod-ing the seeded file to match
// its executable bits — so a seeded bootstrap.sh stays `./bootstrap.sh`-runnable
// (a non-peer bootstrap invokes it directly, where the bit IS load-bearing). The
// mode is read from disk in this IO seam, never carried in the pure Action
// (ARCH-PURE). Non-exec source → the WriteFile 0o644 default stands.
//
// We sync the executable bits even on a content-identical dst (a file seeded by
// an older mode-blind weave is +x-less; a re-weave should converge its mode too,
// like create_seed's `cp -p` would). The cmp -s content no-op still skips the
// rewrite; only the chmod (cheap, idempotent) runs unconditionally below.
func applySeed(fs weavefs.FS, src, dst string) error {
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil // source missing/unreadable → warn-equivalent non-fatal skip
	}
	// Content already current → idempotent no-op on the bytes (cmp -s), but still
	// fall through to the mode sync below so a stale-mode dst converges.
	contentCurrent := false
	if existing, rerr := fs.ReadFile(dst); rerr == nil && string(existing) == string(data) {
		contentCurrent = true
	}
	if !contentCurrent {
		if err := ensureParent(fs, dst); err != nil {
			return err
		}
		if err := fs.WriteFile(dst, data); err != nil {
			return fmt.Errorf("apply seed: write %s: %w", dst, err)
		}
	}
	// Preserve the source's executable bit (the `cp -p` mode-preservation).
	// Observe the source mode in the IO seam; if any exec bit is set, mirror the
	// source's full perm onto dst, else leave the 0o644 WriteFile default.
	if fi, serr := fs.Stat(src); serr == nil && fi.Mode().Perm()&0o111 != 0 {
		if err := fs.Chmod(dst, fi.Mode().Perm()); err != nil {
			return fmt.Errorf("apply seed: chmod %s: %w", dst, err)
		}
	}
	return nil
}

// applyWriteFile ensures parents then writes content (the composed AGENTS.md).
// Overwrites unconditionally — the planner decides content; convergence-on-drift
// is implicit (same content → same bytes).
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
