// Package sidedoc holds the side-doc facts more than one consumer needs
// (#424): the links a document carries to its side docs, its `discussions:`
// index, and its `<name>.side/` tree. The contract is the xx-universal-review
// skill's collaboration-modes.md; the examples checker, tl-lead and sdlc's
// archiving all read it through here. Everything is pure over text and paths.
package sidedoc

import (
	"path"
	"regexp"
	"strings"
)

var linkRE = regexp.MustCompile(`→ \[(?:discussion|answer|notes)\]\(([^)\s]+)\)`)

// SideLinks returns the targets of a document's side-doc links
// (`→ [discussion](…)`, `→ [answer](…)`, `→ [notes](…)`), as written and in
// order, skipping fenced code: a link shown in an example is not a link.
func SideLinks(text string) []string {
	var out []string
	fence := ""
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimLeft(line, " \t")
		if fence == "" && (strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")) {
			fence = t[:3]
			continue
		}
		if fence != "" {
			if strings.HasPrefix(t, fence) {
				fence = ""
			}
			continue
		}
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			out = append(out, m[1])
		}
	}
	return out
}

// linkedSet is the cleaned targets of a document's side links.
func linkedSet(body string) map[string]bool {
	set := map[string]bool{}
	for _, t := range SideLinks(body) {
		set[path.Clean(t)] = true
	}
	return set
}
