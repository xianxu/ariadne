package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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

// TestEveryFlagAppearsInItsHelp (#277 BR-19): a help page with a FLAGS section
// documents every flag its command registers (hidden ones excepted), so a new
// flag cannot ship invisible in --help while the page claims to list them.
func TestEveryFlagAppearsInItsHelp(t *testing.T) {
	catalogued := map[string]bool{} // "<command> --<flag>": rendered from the gate catalog instead
	for _, g := range processmanual.GateCatalog {
		for _, cmd := range g.Commands {
			catalogued[cmd+" --"+g.Flag] = true
		}
	}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		_, flags, listed := strings.Cut(c.Long, "\nFLAGS")
		if !listed {
			return // no hand-written flag list to keep complete
		}
		if end := strings.Index(flags, "\n\n"+"EXAMPLES"); end >= 0 {
			flags = flags[:end]
		}
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden || f.Name == "help" || catalogued[strings.TrimPrefix(c.CommandPath(), "sdlc ")+" --"+f.Name] {
				return
			}
			if !regexp.MustCompile(`--` + regexp.QuoteMeta(f.Name) + `\b`).MatchString(flags) {
				t.Errorf("sdlc %s --help lists FLAGS but omits --%s", c.CommandPath(), f.Name)
			}
		})
	}
	walk(buildRoot())
}
