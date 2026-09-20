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
//   - SeedOnce → the OWNERSHIP sibling of Seed (#239): write the slot at most
//     ONCE, then hand it to the repo permanently. Anything that is not a symlink
//     occupying the slot is presence (no-op, src never read); a symlink — live or
//     dangling — is weave's own prior lowering and is materialized.
//   - WriteFile → AGENTS.md/touch: ensure parents, then write.
//   - MergeSettings → settings merge: read ordered sources + optional sibling
//     settings.local.json, run the pure settingsx.MergeChain, write the target.
//   - EnsureGitignore → the generated-runtime ignore mechanism (gitignore.go):
//     read the repo's .gitignore, replace weave's DELIMITED REGION wholesale
//     (#239 M2 — so a retired entry loses its line), preserve everything outside
//     it, write back only on change. Fails closed on an unreadable file or an
//     unparseable region. weave OWNS this because weave generates those
//     artifacts; emitted once per compile so every derivative gets a clean
//     `git status` with no per-repo hand-edit.
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
		case SeedOnce:
			err = applySeedOnce(fs, act.Src, filepath.Join(repoRoot, act.Dst))
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

// applyMergeSettings is the IO half of the settings cascade: read ordered
// sources and the optional sibling local (settings.local.json, alongside
// act.Target), run the pure settingsx.MergeChain, and write the result to
// act.Target. The local file's path is derived the same way the bash did —
// LOCAL_FILE="$TARGET_DIR/settings.local.json", i.e. the settings.local.json
// sibling of the target — so an arbitrary Target dir resolves its local
// correctly. A missing source is an error; a missing local takes the
// source-only path (sources with meta stripped at the end). All IO lives here
// (ARCH-PURE); the merge itself is pure.
func applyMergeSettings(fs weavefs.FS, repoRoot string, act MergeSettings) error {
	if len(act.Sources) == 0 {
		return fmt.Errorf("apply merge: %s: no sources", act.Target)
	}
	sources := make([][]byte, 0, len(act.Sources)+1)
	for _, sourcePath := range act.Sources {
		data, err := fs.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("apply merge: read source %s: %w", sourcePath, err)
		}
		sources = append(sources, data)
	}

	targetPath := filepath.Join(repoRoot, act.Target)
	localPath := filepath.Join(filepath.Dir(targetPath), "settings.local.json")
	if data, lerr := fs.ReadFile(localPath); lerr == nil {
		sources = append(sources, data)
	}

	merged, err := settingsx.MergeChain(sources)
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
// A destination symlink is removed before comparing bytes or syncing mode,
// including matching and dangling links. Source read failure leaves it intact.
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
	if err := removeDestinationSymlink(fs, dst); err != nil {
		return err
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
	return syncExecBit(fs, src, dst, "seed")
}

// syncExecBit replicates the load-bearing half of `cp -p` for both seed verbs:
// OBSERVE the source's mode and, if any exec bit is set, mirror its full perm
// onto dst (weavefs.FS.WriteFile writes a fixed 0o644, so a seeded bootstrap.sh
// would otherwise land non-executable, and a non-peer bootstrap invokes it
// directly). Non-exec source → the 0o644 default stands.
//
// Shared by applySeed and applySeedOnce: the two differ on WHEN to write, never
// on how to carry the mode, so the mode logic lives once (ARCH-DRY). verb names
// the caller for the error message.
func syncExecBit(fs weavefs.FS, src, dst, verb string) error {
	fi, err := fs.Stat(src)
	if err != nil || fi.Mode().Perm()&0o111 == 0 {
		return nil
	}
	if err := fs.Chmod(dst, fi.Mode().Perm()); err != nil {
		return fmt.Errorf("apply %s: chmod %s: %w", verb, dst, err)
	}
	return nil
}

// SlotState is what occupies a seed-once destination — a TOTAL classification
// of the destination fact, not a predicate over part of it.
//
// It exists because the boolean it replaced (`SeedOnceSlotIsRepoOwned(mode)
// bool`) could only express repo-owned-vs-symlink: handed a zero FileMode for an
// ABSENT slot it answered "repo-owned", the exact opposite of the truth, and
// every caller had to reconstruct the missing cases for itself. Three callers
// re-encoded it three different ways in one milestone (#239 M1 BR-12). A sum
// type that the classifier alone produces makes a partial reconstruction
// unrepresentable (ARCH-ORDER, ARCH-DRY).
type SlotState int

const (
	// SlotUnknown — the destination could not be classified (an Lstat error
	// that is not "not exist"). Callers must FAIL CLOSED: never write.
	SlotUnknown SlotState = iota
	// SlotAbsent — nothing occupies the slot; seed-once writes the template.
	SlotAbsent
	// SlotRepoOwned — a regular file or directory. The REPO owns it: seed-once
	// no-ops forever, whatever it contains.
	SlotRepoOwned
	// SlotWeaveSymlink — a symlink, live or dangling. This is weave's OWN prior
	// `symlink Makefile` lowering, not repo content, so seed-once removes it and
	// materializes the template (the #225 convergence).
	SlotWeaveSymlink
)

func (s SlotState) String() string {
	switch s {
	case SlotAbsent:
		return "absent"
	case SlotRepoOwned:
		return "repo-owned"
	case SlotWeaveSymlink:
		return "weave-symlink"
	default:
		return "unknown"
	}
}

// ClassifySlot is the SINGLE source of truth for "what is in this seed-once
// destination". It takes the raw observation — exactly what Lstat returns — so
// no caller can hand it a partial one.
//
// Exported because two callers must agree and previously did not: applySeedOnce
// (which decides what to DO) and golden.classifyAction (which predicts what
// weave WOULD do). The classifier once called a symlinked slot "present" and
// reported MATCH on exactly the fleet state this verb converges — nous and metis
// both carry Makefile -> ../ariadne/Makefile today. A drift harness that predicts
// the opposite of the seam is worse than no harness (#239 M1 BR-1).
func ClassifySlot(fi os.FileInfo, err error) SlotState {
	switch {
	case os.IsNotExist(err):
		return SlotAbsent
	case err != nil:
		return SlotUnknown // fail closed: an unreadable slot is never written
	case fi.Mode()&os.ModeSymlink != 0:
		return SlotWeaveSymlink
	default:
		return SlotRepoOwned
	}
}

// applySeedOnce is the WRITE-ONCE half of the seed pair (#239). Where applySeed
// converges on upstream every compile, this one writes the slot at most once and
// then hands it to the repo permanently:
//
//   - ANYTHING THAT IS NOT A SYMLINK in the slot (a regular file, a directory)
//     → no-op, with NO read of src and no comparison. The repo owns it, whatever
//     it now contains. This is the whole point: a repo adopting ariadne keeps its
//     own root Makefile, and a repo that later edits its root Makefile keeps that
//     edit across every subsequent weave. (applySeed would reach WriteFile on a
//     directory and error; no-op is the safer behavior here, and is deliberate.)
//   - A SYMLINK in the slot → NOT presence, whether live or DANGLING (Lstat
//     reports ModeSymlink either way). It is weave's own pre-#239 `symlink
//     Makefile` lowering; removing it and materializing a real file is the #225
//     convergence that nous and metis still carry. removeDestinationSymlink also
//     guarantees we never write THROUGH it into the ancestor's own Makefile.
//   - Absent → write src's bytes, preserving its executable bits exactly as
//     applySeed does.
//
// A missing src is a non-fatal skip, matching applySeed: weave can't read the
// template, so it leaves the slot alone rather than erroring the walk.
//
// NOTE the ordering: the presence check runs BEFORE the src read, so a
// repo-owned file is never even compared against upstream.
func applySeedOnce(fs weavefs.FS, src, dst string) error {
	switch state := ClassifySlot(fs.Lstat(dst)); state {
	case SlotRepoOwned:
		return nil // sacrosanct — the repo owns it; src is never even read
	case SlotUnknown:
		// An Lstat error we cannot interpret. Refuse rather than proceed toward
		// a write: the destination may be a symlink into the ANCESTOR, and a
		// wrong guess overwrites ariadne's own Makefile through it.
		return fmt.Errorf("apply seed-once: cannot classify %s", dst)
	case SlotAbsent, SlotWeaveSymlink:
		// Fall through to the write.
	}
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil // template missing/unreadable → non-fatal skip (applySeed's contract)
	}
	if err := removeDestinationSymlink(fs, dst); err != nil {
		return err
	}
	if err := ensureParent(fs, dst); err != nil {
		return err
	}
	if err := fs.WriteFile(dst, data); err != nil {
		return fmt.Errorf("apply seed-once: write %s: %w", dst, err)
	}
	return syncExecBit(fs, src, dst, "seed-once")
}

// applyWriteFile ensures parents then writes content (the composed AGENTS.md).
// Overwrites unconditionally — the planner decides content; convergence-on-drift
// is implicit (same content → same bytes).
//
// If a SYMLINK occupies the slot it is removed FIRST, so we write a fresh regular
// file here and never follow the link to clobber its target. This is the #95
// cutover hazard: until a derivative's first weave, its AGENTS.md is a symlink
// into the ancestor (nous/AGENTS.md → ../ariadne/AGENTS.md), and fs.WriteFile
// (os.WriteFile) follows a symlink — so a naive write would overwrite ariadne's
// source constitution THROUGH the link. Mirrors applySymlink's [[ -L ]] → rm
// guard. A regular file in the slot is fine to truncate-overwrite (WriteFile's
// own O_TRUNC); only a symlink must be unlinked first.
func applyWriteFile(fs weavefs.FS, path, content string) error {
	if err := ensureParent(fs, path); err != nil {
		return err
	}
	if err := removeDestinationSymlink(fs, path); err != nil {
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

// removeDestinationSymlink makes fixed output slots safe for regular-file
// materialization. Unknown destination state fails closed; absence is safe.
// Callers read seed source bytes before invoking this destructive step.
func removeDestinationSymlink(fs weavefs.FS, path string) error {
	fi, err := fs.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("materialize: inspect %s: %w", path, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		if err := fs.Remove(path); err != nil {
			return fmt.Errorf("materialize: remove stale symlink %s: %w", path, err)
		}
	}
	return nil
}
