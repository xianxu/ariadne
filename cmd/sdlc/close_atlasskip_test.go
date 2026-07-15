package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// assertNoGatesigCollision matches an info line against every GateCatalog
// pattern the way the #172 classifier does — ANSI-STRIPPED (friction.go strips
// before matching; a pattern spanning the "==> " prefix would otherwise slip
// the test while firing live). Shared by the #177 and #178 info-line guards.
func assertNoGatesigCollision(t *testing.T, renderedLine string) {
	t.Helper()
	stripped := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(renderedLine, "")
	for _, g := range processmanual.GateCatalog {
		if g.AckPat != "" && regexp.MustCompile(g.AckPat).MatchString(stripped) {
			t.Errorf("line matches %s/%s AckPat: %q", g.Commands, g.Flag, stripped)
		}
		if g.RefusalPat != "" && regexp.MustCompile(g.RefusalPat).MatchString(stripped) {
			t.Errorf("line matches %s/%s RefusalPat: %q", g.Commands, g.Flag, stripped)
		}
	}
}

// #177: the single docs classifier — *.md anywhere, or anything under
// workshop/, atlas/, docs/. Everything else (Makefile, .gitignore,
// extensionless files) conservatively counts as code: build files ARE
// architectural surface, so they keep the refusal.
func TestHasCodePath(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"go file is code", []string{"cmd/sdlc/close.go"}, true},
		{"README.md is docs", []string{"README.md"}, false},
		{"workshop is docs", []string{"workshop/issues/000177-x.md"}, false},
		{"docs dir is docs", []string{"docs/vision/x.md"}, false},
		{"atlas is docs (single classifier, even though the gate peels atlas/ first)", []string{"atlas/workflow/x.md"}, false},
		{"Makefile is code (build surface)", []string{"Makefile"}, true},
		{"extensionless is code (conservative)", []string{"LICENSE"}, true},
		{"empty window has no code", nil, false},
		{"mixed: one code file flips it", []string{"workshop/x.md", "pkg/vocab/v.go"}, true},
	}
	for _, tc := range cases {
		if got := hasCodePath(tc.paths); got != tc.want {
			t.Errorf("%s: hasCodePath(%v) = %v, want %v", tc.name, tc.paths, got, tc.want)
		}
	}
}

func TestAtlasAutoSatisfyLine(t *testing.T) {
	line := atlasAutoSatisfyLine(3)
	for _, want := range []string{"atlas gate", "no code surface", "auto-satisfied", "3"} {
		if !strings.Contains(line, want) {
			t.Errorf("auto-satisfy line missing %q: %q", want, line)
		}
	}
}

// #172-instrument guard (same shape as #178's): the auto-satisfy info line must
// classify as NEITHER a bypass ACK nor a refusal under any gate signature —
// else the friction report would count routine auto-satisfactions as gate events.
func TestAtlasAutoSatisfyLineNoGatesigCollision(t *testing.T) {
	// as rendered: cinfo prefixes "==> " with ANSI + the reset marker
	assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+atlasAutoSatisfyLine(2))
}
