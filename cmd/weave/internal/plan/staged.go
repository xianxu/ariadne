package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// stagedOutput requires proven ownership before replacing an occupied slot,
// including a cold compile with no previous inventory. ApplyManaged unwraps it
// only after recording this policy; ordinary composition retains migration rules.
type stagedOutput struct{ Action }

func outputAction(action Action) Action {
	if staged, ok := action.(stagedOutput); ok {
		return staged.Action
	}
	return action
}

func generationStageDestination(root string) string {
	return filepath.Join(root, filepath.Dir(InventoryPath), "generation")
}

// ReclaimGenerationStages recovers interrupted producers even when no dynamic
// skills remain selected. It never executes during read-only compilation.
func ReclaimGenerationStages(fs weavefs.FS, root string) error {
	if err := safeParents(fs, root, InventoryPath); err != nil {
		return err
	}
	dir := filepath.Join(root, filepath.Dir(InventoryPath))
	if err := fs.MkdirAll(dir); err != nil {
		return err
	}
	return staging.Reclaim(generationStageDestination(root))
}

// NewGenerationStage establishes durable ownership before any generator runs.
// The caller keeps the returned directory until collection/planning completes
// and removes it on success and ordinary failure; retries reclaim dead owners.
func NewGenerationStage(fs weavefs.FS, root string) (string, error) {
	if err := ReclaimGenerationStages(fs, root); err != nil {
		return "", err
	}
	return staging.New(generationStageDestination(root))
}
func RemoveGenerationStage(stage string) error { return staging.Remove(stage) }

// StagedActions reads a private producer output tree into final-path actions.
// It never writes live outputs. A marker must have produced a regular nonempty
// SKILL.md; every other regular file/link is included, including side outputs.
// Links within the output tree are relocated to its final tree. Relative links
// escaping that tree are rejected rather than retaining temporary dependencies.
func StagedActions(fs weavefs.FS, root, stageDir, outputRel string) ([]Action, error) {
	if !safeRelative(outputRel) || !strings.HasPrefix(outputRel, walk.GeneratedRel+string(filepath.Separator)) || reservedOutput(outputRel) {
		return nil, fmt.Errorf("invalid generated output directory %q", outputRel)
	}
	if !filepath.IsAbs(stageDir) {
		return nil, fmt.Errorf("staged output directory must be absolute: %s", stageDir)
	}
	info, err := fs.Lstat(stageDir)
	if err != nil {
		return nil, fmt.Errorf("staged output %s: %w", stageDir, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("staged output is not a directory: %s", stageDir)
	}
	skillPath := filepath.Join(stageDir, "SKILL.md")
	info, err = fs.Lstat(skillPath)
	if err != nil {
		return nil, fmt.Errorf("generator must produce staged SKILL.md: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, fmt.Errorf("generator must produce regular nonempty staged SKILL.md: %s", skillPath)
	}
	var actions []Action
	var visit func(string) error
	visit = func(rel string) error {
		entries, err := fs.ReadDir(filepath.Join(stageDir, rel))
		if err != nil {
			return err
		}
		for _, entry := range entries {
			path := filepath.Join(rel, entry.Name())
			source := filepath.Join(stageDir, path)
			dest := filepath.Join(outputRel, path)
			info, err := fs.Lstat(source)
			if err != nil {
				return err
			}
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				target, err := fs.Readlink(source)
				if err != nil {
					return err
				}
				resolved := target
				if !filepath.IsAbs(resolved) {
					resolved = filepath.Join(filepath.Dir(source), resolved)
				}
				inside, err := filepath.Rel(stageDir, resolved)
				if err != nil {
					return err
				}
				if safeRelative(inside) || inside == "." {
					resolved = filepath.Join(root, outputRel, inside)
				} else if !filepath.IsAbs(target) || targetWithinAnyRoot(resolved, []string{filepath.Join(root, filepath.Dir(InventoryPath))}) {
					return fmt.Errorf("generated link %s escapes staging output", path)
				}
				actions = append(actions, stagedOutput{Symlink{Src: resolved, Dst: dest}})
			case info.IsDir():
				if err := visit(path); err != nil {
					return err
				}
			case info.Mode().IsRegular():
				content, err := fs.ReadFile(source)
				if err != nil {
					return err
				}
				mode := info.Mode().Perm()
				actions = append(actions, stagedOutput{WriteFile{Path: dest, Content: string(content), Mode: &mode}})
			default:
				return fmt.Errorf("unsupported staged output type: %s", source)
			}
		}
		return nil
	}
	if err := visit(""); err != nil {
		return nil, err
	}
	return actions, nil
}
