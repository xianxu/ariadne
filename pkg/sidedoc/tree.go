package sidedoc

import (
	"path"
	"strings"
)

// SideDir is the folder holding a document's side docs: `<name>.side/`
// beside `<name>.md` (collaboration-modes.md §4), in the path's own form.
func SideDir(doc string) string {
	return strings.TrimSuffix(doc, ".md") + ".side"
}

// Owner returns the synthetic document path owning a canonical side root or
// any member beneath it. The first side root owns nested side trees. It uses
// only the slash path, so the document need not exist. Noncanonical .md.side
// names have no owner; callers retain their existing naming refusals.
func Owner(p string) (string, bool) {
	segs := strings.Split(path.Clean(p), "/")
	for i, s := range segs {
		if !strings.HasSuffix(s, ".side") || s == ".side" {
			continue
		}
		if strings.HasSuffix(s, ".md.side") {
			return "", false
		}
		return strings.TrimSuffix(strings.Join(segs[:i+1], "/"), ".side") + ".md", true
	}
	return "", false
}

// MoveMember maps a file inside fromDoc's side folder to its place when the
// document moves to toDoc (collaboration-modes.md §5: the whole tree moves
// with its document). ok is false for a path outside the folder.
func MoveMember(member, fromDoc, toDoc string) (string, bool) {
	return MoveUnder(member, SideDir(fromDoc), SideDir(toDoc))
}

// Folder returns the side folder (`<name>.side`) a slash path lies in: the
// path up to its first segment ending in `.side`, before its last segment.
func Folder(p string) (string, bool) {
	segs := strings.Split(path.Clean(p), "/")
	for i, s := range segs[:len(segs)-1] {
		if strings.HasSuffix(s, ".side") && s != ".side" {
			return strings.Join(segs[:i+1], "/"), true
		}
	}
	return "", false
}

// InSideDir reports whether a slash path lies inside some document's side
// folder.
func InSideDir(p string) bool {
	_, ok := Folder(p)
	return ok
}
