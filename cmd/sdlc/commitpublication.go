package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// validateCommitSelection accepts or refuses the entire selected change set.
func validateCommitSelection(selected gitx.SelectedCommit, roots []string) error {
	for _, change := range selected.Changes {
		p := change.Path
		asset := false
		for i, root := range roots {
			if i == 0 || i == 2 {
				if i == 2 {
					root = path.Join(root, "issues")
				}
				prefix := path.Join(root, "assets") + "/"
				if strings.HasPrefix(p, prefix) {
					id, _, ok := strings.Cut(strings.TrimPrefix(p, prefix), "/")
					asset = asset || (ok && len(id) == 6 && issueIDPrefix(id+"-x.md") == id)
				}
			}
		}
		eligible := p != "" && path.Clean(p) == p && !strings.HasPrefix(p, "/") && (strings.HasSuffix(p, ".md") || asset)
		inside := false
		for _, root := range roots {
			if strings.HasPrefix(p, strings.TrimSuffix(root, "/")+"/") {
				inside = true
			}
		}
		for _, mode := range []string{change.OldMode, change.NewMode} {
			if mode != "000000" && mode != "100644" {
				eligible = false
			}
		}
		if !eligible || !inside {
			return fmt.Errorf("commit %s contains ineligible workflow path %q (modes %s → %s); split documentation from code, executable files, symlinks and submodules, then select the documentation commit", selected.Source, p, change.OldMode, change.NewMode)
		}
	}
	return nil
}
