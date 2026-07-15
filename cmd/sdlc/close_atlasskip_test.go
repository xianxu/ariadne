package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

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
	line := "\x1b[1;36m==>\x1b[0m " + atlasAutoSatisfyLine(2)
	for _, g := range processmanual.GateCatalog {
		if g.AckPat != "" && regexp.MustCompile(g.AckPat).MatchString(line) {
			t.Errorf("auto-satisfy line matches %s/%s AckPat", g.Commands, g.Flag)
		}
		if g.RefusalPat != "" && regexp.MustCompile(g.RefusalPat).MatchString(line) {
			t.Errorf("auto-satisfy line matches %s/%s RefusalPat", g.Commands, g.Flag)
		}
	}
}
