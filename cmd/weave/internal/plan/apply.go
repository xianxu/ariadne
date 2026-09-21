package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/settingsx"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Apply executes a []Action against fs, idempotently. ApplyManaged wraps it
// with generated-output ownership and retirement; planning remains pure.
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
//   - MergeSettings → settings merge: read ordered sources + optional sibling
//     settings.local.json, run the pure settingsx.MergeChain, write the target.
//   - EnsureGitignore → replace the delimited generated block, preserving
//     authored rules and their precedence. ApplyManaged derives its entries
//     from the scoped inventory rather than this display action.
//
// The retired `tool` verb (#95 M5) has no Action and no IO here: Go-tool
// ownership is location-based (construct/dev-aliases.sh scans sibling cmd/X dirs)
// and deps come from `weave link` / construct/deps, so weave never edits go.mod.
func Apply(fs weavefs.FS, repoRoot string, actions []Action) error {
	if err := weavefs.ReclaimPublications(fs, repoRoot); err != nil {
		return err
	}
	for _, a := range actions {
		var err error
		switch act := a.(type) {
		case Symlink:
			err = applySymlink(fs, filepath.Join(repoRoot, act.Dst), act.Src)
		case Mkdir:
			err = applyMkdir(fs, filepath.Join(repoRoot, act.Path))
		case Seed:
			err = applySeed(fs, repoRoot, act.Src, filepath.Join(repoRoot, act.Dst))
		case SeedOnce:
			err = applySeedOnce(fs, repoRoot, act.Src, filepath.Join(repoRoot, act.Dst))
		case Touch:
			err = applyTouch(fs, repoRoot, filepath.Join(repoRoot, act.Path))
		case WriteFile:
			err = applyWriteFile(fs, repoRoot, filepath.Join(repoRoot, act.Path), act.Content, act.Mode)
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

const workflowInclude = "-include Makefile.workflow"

// applySeedOnce creates a repo-owned file only when absent. A regular existing
// Makefile is adopted by prepending the shared workflow include; symlinks and
// other non-regular paths are intentionally left untouched.
func applySeedOnce(fs weavefs.FS, root, src, dst string) error {
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil
	}
	source, err := fs.Stat(src)
	if err != nil {
		return err
	}
	current, err := fs.Lstat(dst)
	if os.IsNotExist(err) {
		mode := source.Mode().Perm()
		if err := weavefs.Publish(fs, root, dst, data, &mode); err != nil {
			return fmt.Errorf("apply seed-once: write %s: %w", dst, err)
		}
		return nil
	}
	if err != nil {
		return err
	}
	if !current.Mode().IsRegular() {
		return nil
	}
	contents, err := fs.ReadFile(dst)
	if err != nil {
		return err
	}
	if hasWorkflowInclude(string(contents)) {
		return nil
	}
	updated := []byte(workflowInclude + "\n" + string(contents))
	mode := current.Mode().Perm()
	if err := weavefs.Publish(fs, root, dst, updated, &mode); err != nil {
		return fmt.Errorf("apply seed-once: adopt %s: %w", dst, err)
	}
	return nil
}

func hasWorkflowInclude(contents string) bool {
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == workflowInclude || trimmed == "include Makefile.workflow" {
			return true
		}
	}
	return false
}

// SeedOnceInstruction reports why a non-regular existing destination could
// not be adopted automatically. The compile remains successful and preserves
// the user's path.
func SeedOnceInstruction(fs weavefs.FS, root string, action SeedOnce) (string, error) {
	info, err := fs.Lstat(filepath.Join(root, action.Dst))
	if os.IsNotExist(err) || (err == nil && info.Mode().IsRegular()) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("weave: %s is not a regular file; add %q as its first line to enable shared workflow targets", action.Dst, workflowInclude), nil
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
func mergedSettings(fs weavefs.FS, repoRoot string, act MergeSettings) ([]byte, error) {
	if len(act.Sources) == 0 {
		return nil, fmt.Errorf("apply merge: %s: no sources", act.Target)
	}
	sources := make([][]byte, 0, len(act.Sources)+1)
	for _, sourcePath := range act.Sources {
		data, err := fs.ReadFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("apply merge: read source %s: %w", sourcePath, err)
		}
		sources = append(sources, data)
	}

	targetPath := filepath.Join(repoRoot, act.Target)
	localPath := filepath.Join(filepath.Dir(targetPath), "settings.local.json")
	if data, lerr := fs.ReadFile(localPath); lerr == nil {
		sources = append(sources, data)
	} else if !os.IsNotExist(lerr) {
		return nil, fmt.Errorf("apply merge: read local %s: %w", localPath, lerr)
	}

	merged, err := settingsx.MergeChain(sources)
	if err != nil {
		return nil, fmt.Errorf("apply merge: %s: %w", targetPath, err)
	}
	return merged, nil
}

func applyMergeSettings(fs weavefs.FS, repoRoot string, act MergeSettings) error {
	merged, err := mergedSettings(fs, repoRoot, act)
	if err != nil {
		return err
	}
	return applyWriteFile(fs, repoRoot, filepath.Join(repoRoot, act.Target), string(merged), nil)
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
func applyTouch(fs weavefs.FS, root, path string) error {
	if _, err := fs.Lstat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return weavefs.Publish(fs, root, path, []byte{}, nil)
}

// applySeed materializes an authored entrypoint with source permissions. Source
// read failure retains the established non-fatal skip. Publication replaces any
// destination symlink atomically, never following it into an ancestor.
func applySeed(fs weavefs.FS, root, src, dst string) error {
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil
	}
	source, err := fs.Stat(src)
	if err != nil {
		return err
	}
	mode := source.Mode().Perm()
	current, err := fs.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil && current.Mode().IsRegular() && current.Mode().Perm() == mode {
		previous, err := fs.ReadFile(dst)
		if err != nil {
			return err
		}
		if string(previous) == string(data) {
			return nil
		}
	}
	if err := weavefs.Publish(fs, root, dst, data, &mode); err != nil {
		return fmt.Errorf("apply seed: write %s: %w", dst, err)
	}
	return nil
}

// applyWriteFile atomically replaces a composed or generated output. Explicit
// generated permissions travel with the bytes; nil retains ordinary-file modes.
func applyWriteFile(fs weavefs.FS, root, path, content string, mode *os.FileMode) error {
	if err := weavefs.Publish(fs, root, path, []byte(content), mode); err != nil {
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
