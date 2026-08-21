package main

import (
	"strings"
	"testing"
)

func TestClassifyReviewAnchor(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    reviewAnchorDelta
		want anchorOutcome
	}{
		{"no anchor recorded degrades to unchanged", reviewAnchorDelta{Reviewed: "", Current: "bbb"}, anchorUnchanged},
		{"head still at the reviewed commit", reviewAnchorDelta{Reviewed: "aaa", Current: "aaa", Descendant: true}, anchorUnchanged},
		{"doc-only delta finalizes", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "#69: lessons"}},
			Paths:   []string{"workshop/lessons.md", "atlas/index.md"},
		}, anchorDocsOnly},
		{"code delta refuses", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "#69: fix"}},
			Paths:   []string{"cmd/sdlc/close.go"},
		}, anchorCodeDelta},
		// publishGateHasCodeSurface tightens hasCodePath for exactly this: helptext is
		// //go:embed'ed into the binary, so a post-review edit to it IS shipped behavior.
		{"embedded helptext counts as code", reviewAnchorDelta{
			Reviewed: "aaa", Current: "bbb", Descendant: true,
			Commits: []deltaCommit{{SHA: "bbb", Subject: "#69: helptext"}},
			Paths:   []string{"cmd/sdlc/helptext/close.md"},
		}, anchorCodeDelta},
		// Without the ancestry check this would read as doc-only: `git diff A B` between
		// unrelated commits returns paths perfectly happily.
		{"rebased away refuses as diverged", reviewAnchorDelta{
			Reviewed: "aaa", Current: "ccc", Descendant: false,
			Paths: []string{"workshop/lessons.md"},
		}, anchorDiverged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyReviewAnchor(tc.d); got != tc.want {
				t.Fatalf("classifyReviewAnchor = %v, want %v", got, tc.want)
			}
		})
	}
}

// #172: gatesig classifies transcripts by substring, so a PASS line that echoes
// refusal vocabulary corrupts friction attribution — the same constraint
// formatPublishGateDocsOnly documents for the publish gate's docs-only line.
func TestFormatAnchorDocsOnly_SharesNoRefusalVocabulary(t *testing.T) {
	pass := formatAnchorDocsOnly(reviewAnchorDelta{
		Reviewed: "aaaaaaaaaaaa", Current: "bbbbbbbbbbbb", Descendant: true,
		Commits: []deltaCommit{{SHA: "bbbbbbbbbbbb", Subject: "docs"}},
	})
	for _, forbidden := range []string{"NOT finalized", "unreviewed", "re-run", "stale", "landed after"} {
		if strings.Contains(pass, forbidden) {
			t.Errorf("pass line must not contain refusal word %q:\n%s", forbidden, pass)
		}
	}
}

func TestFormatAnchorRefusal_NamesEveryCommit(t *testing.T) {
	d := reviewAnchorDelta{
		Reviewed: "aaaaaaaaaaaa", Current: "cccccccccccc", Descendant: true,
		Commits: []deltaCommit{
			{SHA: "cccccccccccc", Subject: "second fix"},
			{SHA: "bbbbbbbbbbbb", Subject: "first fix"},
		},
		Paths: []string{"cmd/sdlc/close.go", "workshop/lessons.md"},
	}
	msg := formatAnchorRefusal(d, anchorCodeDelta, "sdlc close")
	for _, want := range []string{"cccccccc", "second fix", "bbbbbbbb", "first fix", "cmd/sdlc/close.go", "sdlc close"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal must name %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "workshop/lessons.md") {
		t.Errorf("the code-surface list must not include doc paths:\n%s", msg)
	}
	if strings.Contains(msg, "HEAD changed from") {
		t.Errorf("refusal must report commits, not bare HEAD identity:\n%s", msg)
	}
}

func TestFormatAnchorRefusal_DivergedNamesBothEnds(t *testing.T) {
	msg := formatAnchorRefusal(reviewAnchorDelta{
		Reviewed: "aaaaaaaaaaaa", Current: "cccccccccccc", Descendant: false,
	}, anchorDiverged, "sdlc milestone-close")
	for _, want := range []string{"aaaaaaaa", "cccccccc", "sdlc milestone-close"} {
		if !strings.Contains(msg, want) {
			t.Errorf("diverged refusal must name %q:\n%s", want, msg)
		}
	}
}
