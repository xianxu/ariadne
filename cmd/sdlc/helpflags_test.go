package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// TestGateFlagListsAreRenderedFromTheCatalog: a command's gate-flag set is
// OWNED by processmanual.GateCatalog, so no help page may keep its own copy.
// Line-granular (#231 M2 review BR-27 — a page-granular check passed a stale
// list because the flag appeared elsewhere on the page): no raw helptext line
// may LIST a gate flag — as a list item (the line starts with it) or a table row
// (it ends the line after a column gap). Mid-sentence mentions and command
// examples are fine. And every page with gates renders the full table.
func TestGateFlagListsAreRenderedFromTheCatalog(t *testing.T) {
	pages := map[string][]string{}
	for _, g := range processmanual.GateCatalog {
		for _, cmd := range g.Commands {
			page, _, _ := strings.Cut(cmd, " ")
			pages[page] = append(pages[page], "--"+g.Flag)
		}
	}
	for page, flags := range pages {
		raw := helptext.MustGet(page)
		for i, line := range strings.Split(raw, "\n") {
			for _, f := range flags {
				item := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(f) + `(\s|,|$)`)
				row := regexp.MustCompile(`\S\s{2,}` + regexp.QuoteMeta(f) + `\s*$`)
				if item.MatchString(line) || row.MatchString(line) {
					t.Errorf("helptext/%s.md:%d hand-lists %s — render the set with {{GATE_FLAGS}}:\n  %s", page, i+1, f, line)
				}
			}
		}
		if !strings.Contains(raw, "{{GATE_FLAGS}}") {
			continue // the page lists no gate flags at all
		}
		rendered := renderLong(page)
		for _, f := range flags {
			if !strings.Contains(rendered, f) {
				t.Errorf("sdlc %s --help renders no %s", page, f)
			}
		}
	}
}
