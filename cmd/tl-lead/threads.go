package main

import (
	"errors"
	"strings"

	"github.com/xianxu/ariadne/pkg/markdoc"
)

// The `## Threads` section is a list of `### ` subsections in focus order
// (`### Project: <slug>`, `### Train: <name>`, `### Miscellaneous`), with an
// optional preamble before the first. A thread entry is one such subsection.

const miscThread = "### Miscellaneous"

// threadBlocks splits a Threads body into its preamble and its `### `
// subsections (each block keeps its heading line and body), skipping `### `
// lines inside a fence.
func threadBlocks(body string) (preamble string, blocks []string) {
	var pre []string
	lines := strings.Split(strings.Trim(body, "\n"), "\n")
	// A `### ` line inside a fenced block is content, not a heading.
	fenced := markdoc.FenceSpans(lines, markdoc.UnterminatedIsProse)
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "### ") && !fenced[i]:
			blocks = append(blocks, line)
		case len(blocks) == 0:
			pre = append(pre, line)
		default:
			blocks[len(blocks)-1] += "\n" + line
		}
	}
	for i := range blocks {
		blocks[i] = strings.TrimRight(blocks[i], "\n")
	}
	return strings.Trim(strings.Join(pre, "\n"), "\n"), blocks
}

func heading(block string) string {
	h, _, _ := strings.Cut(block, "\n")
	return strings.TrimSpace(h)
}

// upsertThread puts entry (a `### ` subsection) into the Threads body: it
// replaces the subsection with the same heading in place, or else goes in
// before `### Miscellaneous` (which stays last), or else at the end. Every
// existing subsection whose heading starts with one of the drop prefixes is
// removed first.
func upsertThread(body, entry string, drop []string) (string, error) {
	entry = strings.Trim(entry, "\n")
	if !strings.HasPrefix(entry, "### ") {
		return "", errors.New("a Threads entry is a `### ` subsection: its first line starts with '### '")
	}
	pre, blocks := threadBlocks(body)
	var kept []string
	for _, b := range blocks {
		dropped := false
		for _, p := range drop {
			if strings.HasPrefix(heading(b), p) {
				dropped = true
				break
			}
		}
		if !dropped {
			kept = append(kept, b)
		}
	}
	placed := false
	for i, b := range kept {
		if heading(b) == heading(entry) {
			kept[i], placed = entry, true
			break
		}
	}
	if !placed {
		at := len(kept)
		for i, b := range kept {
			if heading(b) == miscThread && heading(entry) != miscThread {
				at = i
				break
			}
		}
		kept = append(kept[:at], append([]string{entry}, kept[at:]...)...)
	}
	parts := kept
	if pre != "" {
		parts = append([]string{pre}, kept...)
	}
	return strings.Join(parts, "\n\n"), nil
}
