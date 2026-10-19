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
	// Pair escapes Markdown brackets inside its bare 🤖{…} proposal. Inspect
	// each such marker for every pending link; one marker can contain several.
	// Escaped Markdown elsewhere in prose is not a link.
	pendingMarkerRE = regexp.MustCompile(`🤖\{[^}\n]*\}`)
	escapedLinkRE   = regexp.MustCompile(`→ \\\[(?:discussion|answer|notes|proposal)\\\]\(([^)\s]+)\)`)
)

// SideLinks returns the targets of a document's side-doc links
// (`→ [discussion](…)`, `→ [answer](…)`, `→ [notes](…)`,
// `→ [proposal](…)`), including Pair's escaped pending-link marker, as written
// and in order. Fenced examples are not links.
func SideLinks(text string) []string {
	var out []string
	forUnfenced(strings.Split(text, "\n"), func(_ int, line string) {
		type found struct {
			at     int
			target string
		}
		var links []found
		for _, m := range linkRE.FindAllStringSubmatchIndex(line, -1) {
			links = append(links, found{m[0], line[m[2]:m[3]]})
		}
		for _, marker := range pendingMarkerRE.FindAllStringIndex(line, -1) {
			body := line[marker[0]:marker[1]]
			for _, m := range escapedLinkRE.FindAllStringSubmatchIndex(body, -1) {
				links = append(links, found{marker[0] + m[0], body[m[2]:m[3]]})
			}
		}
		sort.Slice(links, func(i, j int) bool { return links[i].at < links[j].at })
		for _, link := range links {
			out = append(out, link.target)
		}
	})
	return out
}

// forUnfenced calls f with each line outside fenced code (``` or ~~~), and
// its index; the fence lines themselves are code too.
func forUnfenced(lines []string, f func(i int, line string)) {
	fence := ""
	for i, line := range lines {
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
		f(i, line)
	}
}

// linkedSet is the cleaned targets of a document's side links.
func linkedSet(body string) map[string]bool {
	set := map[string]bool{}
	for _, t := range SideLinks(body) {
		set[path.Clean(t)] = true
	}
	return set
}
