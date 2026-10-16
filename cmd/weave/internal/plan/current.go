package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// OutputCurrent observes whether an output action already has its intended
// result. It shares materialization and exact identity matching with managed
// apply; create-if-missing actions retain their ordinary no-op semantics.
func OutputCurrent(fs weavefs.FS, root string, action Action) (bool, error) {
	action = outputAction(action)
	switch a := action.(type) {
	case Seed, SeedOnce:
		var source, target string
		switch a := a.(type) {
		case Seed:
			source, target = a.Src, a.Dst
		case SeedOnce:
			source, target = a.Src, a.Dst
		}
		data, err := fs.ReadFile(source)
		if err != nil {
			return true, nil // Ordinary seed apply skips unreadable sources.
		}
		info, err := fs.Stat(source)
		if err != nil {
			return false, err
		}
		if _, once := action.(SeedOnce); once {
			current, err := fs.Lstat(filepath.Join(root, target))
			if err != nil {
				return false, unlessAbsent(err)
			}
			if !current.Mode().IsRegular() {
				return true, nil // Ordinary adoption preserves non-regular paths.
			}
			contents, err := fs.ReadFile(filepath.Join(root, target))
			return hasWorkflowInclude(string(contents)), err
		}
		mode := info.Mode().Perm()
		action = WriteFile{Path: target, Content: string(data), Mode: &mode}
	case Touch:
		_, err := fs.Lstat(filepath.Join(root, a.Path))
		return err == nil, unlessAbsent(err)
	case Mkdir:
		info, err := fs.Stat(filepath.Join(root, a.Path))
		if err != nil {
			return false, unlessAbsent(err)
		}
		return info.IsDir(), nil
	case DeferredOutput:
		return true, nil
	}
	actions, err := materializeManaged(fs, root, []Action{action})
	if err != nil {
		return false, err
	}
	ids, err := actionIdentities(root, actions, ScopeArtifacts)
	if err != nil {
		return false, err
	}
	if len(ids) != 1 {
		return false, fmt.Errorf("cannot inspect output action %T", action)
	}
	return observe(fs, root, ids[0])
}

func unlessAbsent(err error) error {
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
