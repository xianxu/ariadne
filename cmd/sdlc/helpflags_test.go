package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
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

func TestLocalCheckHelpKeepsIntegrationInLandingTurn(t *testing.T) {
	help := strings.Join(strings.Fields(renderLong("local-check")), " ")
	for _, want := range []string{
		"After the pre-close check passes, close that tested branch without manually integrating each newer main",
		"sdlc local-check run --pre-close",
		"Branch code fixes still need verification",
		"noon freshness gate still applies",
		"it owns a landing turn while integrating current main",
		"Explicit local-check runs do not reserve the landing line",
		"docs-only main integration",
		"Embedded help under cmd/, code, tests and configuration invalidate reuse",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("local-check help omits landing safeguard %q", want)
		}
	}
}

// #561 review: landing help renders the one ramp rule, wrapped to the page
// with every word kept and no line past the width.
func TestLandingHelpRendersTrainRampPolicy(t *testing.T) {
	help := renderLong("landing")
	if strings.Contains(help, "{{TRAIN_RAMP_POLICY}}") {
		t.Fatal("placeholder left unrendered")
	}
	if got, want := strings.Join(strings.Fields(help), " "), strings.Join(strings.Fields(landingline.FormatTrainRampPolicy()), " "); !strings.Contains(got, want) {
		t.Fatalf("landing help lacks the ramp policy:\n%s", help)
	}
	wrapped := wrapHelpText("alpha beta gamma delta epsilon", "  ", 14)
	if strings.Join(strings.Fields(wrapped), " ") != "alpha beta gamma delta epsilon" {
		t.Fatalf("wrapping lost words: %q", wrapped)
	}
	for i, line := range strings.Split(wrapped, "\n") {
		if n := len(line); (i == 0 && n+2 > 14) || (i > 0 && n > 14) {
			t.Fatalf("line %d too long: %q", i, line)
		}
	}
}

// #583 (pair:1's report): close's FIXED-THEN-SHIP prose names the pre-close
// check, as the --full-suite flag help does, never a mandatory full suite.
func TestCloseHelpNamesThePreCloseCheck(t *testing.T) {
	help := strings.Join(strings.Fields(renderLong("close")), " ")
	if !strings.Contains(help, "sdlc local-check run --pre-close") {
		t.Error("close help does not name the pre-close check")
	}
	for _, stale := range []string{"records a full test run", "the suite still runs"} {
		if strings.Contains(help, stale) {
			t.Errorf("close help still says %q", stale)
		}
	}
}
