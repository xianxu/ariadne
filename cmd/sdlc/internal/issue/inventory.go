package issue

import (
	"path"
	"strings"
)

// InventoryRole distinguishes live details from either archive layout.
type InventoryRole string

const (
	InventoryActive  InventoryRole = "active"
	InventoryArchive InventoryRole = "archive"
)

// IssueInventoryPath classifies canonical repository-relative direct children.
// roots[0] is active; later roots are archives.
func IssueInventoryPath(p string, roots []string) (id string, role InventoryRole, ok bool) {
	canonical := func(p string) bool {
		return p != "" && p != "." && !path.IsAbs(p) && path.Clean(p) == p && p != ".." && !strings.HasPrefix(p, "../") && !strings.ContainsAny(p, "\\\x00")
	}
	if !canonical(p) {
		return "", "", false
	}
	id, _, ok = ParseFilename(path.Base(p))
	if !ok {
		return "", "", false
	}
	for i, root := range roots {
		if canonical(root) && path.Dir(p) == root {
			if i == 0 {
				return id, InventoryActive, true
			}
			return id, InventoryArchive, true
		}
	}
	return "", "", false
}
