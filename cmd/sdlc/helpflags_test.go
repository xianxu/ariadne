package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// TestGateFlagListsInHelpAreComplete: cobra's generated Flags section lists
// every registered flag, and TestGateCatalogMatchesRegisteredFlags pins those to
// processmanual.GateCatalog. What can drift is PROSE that restates a command's
// gate flags — a table or a FLAGS list written by hand. So: a help page that
// names any of its command's gate flags must name all of them; a partial list
// is exactly the restatement that went stale (#231 M2 review BR-27, when
// --no-done-when-fresh was missing from close's gate table).
func TestGateFlagListsInHelpAreComplete(t *testing.T) {
	flags := map[string][]string{} // help page -> gate flags of its commands
	for _, g := range processmanual.GateCatalog {
		for _, cmd := range g.Commands {
			page, _, _ := strings.Cut(cmd, " ") // a subcommand's help lives on its parent's page
			flags[page] = append(flags[page], "--"+g.Flag)
		}
	}
	for page, fs := range flags {
		help := renderLong(page)
		var named, missing []string
		for _, f := range fs {
			if strings.Contains(help, f) {
				named = append(named, f)
			} else {
				missing = append(missing, f)
			}
		}
		if len(named) > 0 && len(missing) > 0 {
			sort.Strings(missing)
			t.Errorf("`sdlc %s --help` lists some gate flags but not %v — a partial restatement of GateCatalog", page, missing)
		}
	}
}
