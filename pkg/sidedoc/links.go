// Package sidedoc holds the side-doc facts more than one consumer needs
// (#424): the links a document carries to its side docs, its `discussions:`
// index, and its `<name>.side/` tree. The contract is the xx-universal-review
// skill's collaboration-modes.md; the examples checker, tl-lead and sdlc's
// archiving all read it through here. Everything is pure over text and paths.
package sidedoc

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	linkRE = regexp.MustCompile(`→ \[(?:discussion|answer|notes|proposal)\]\(([^)\s]+)\)`)
	// Pair's proposal converter escapes Markdown brackets in a bare 🤖{…}
	// marker. The pending link still owns its side doc and index entry until
	// the operator accepts or rejects it. Match only this marker shape, not
	// arbitrary escaped Markdown in prose.
	pendingLinkRE = regexp.MustCompile(`🤖\{[^}\n]*?→ \\\[(?:discussion|answer|notes|proposal)\\\]\(([^)\s]+)\)[^}\n]*\}`)
)

// SideLinks returns the targets of a document's side-doc links
// (`→ [discussion](…)`, `→ [answer](…)`, `→ [notes](…)`,
// `→ [proposal](…)`), including Pair's escaped pending-link marker, as written
// and in order. Fenced examples are not links.
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
		type found struct {
			at     int
			target string
		}
		var links []found
		for _, m := range linkRE.FindAllStringSubmatchIndex(line, -1) {
			links = append(links, found{m[0], line[m[2]:m[3]]})
		}
		for _, m := range pendingLinkRE.FindAllStringSubmatchIndex(line, -1) {
			links = append(links, found{m[0], line[m[2]:m[3]]})
		}
		sort.Slice(links, func(i, j int) bool { return links[i].at < links[j].at })
		for _, link := range links {
			out = append(out, link.target)
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
