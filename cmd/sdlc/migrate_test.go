package main

import (
	"strings"
	"testing"
)

// TestRewriteRefs pins the #179 rewrite semantics: three rules (bare →
// source-qualified; dest-qualified → bare; everything else passes through),
// fence/span awareness, and the parseRef candidate filter. src repo "src",
// dst repo "dst".
func TestRewriteRefs(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		want        string
		wantRewrite int // len(rewrites)
		wantSkip    int // len(skipped)
	}{
		{"bare qualifies", "see #12 for detail\n", "see src#12 for detail\n", 1, 0},
		{"dest-qualified normalizes", "tracked in dst#5.\n", "tracked in #5.\n", 1, 0},
		{"self-qualified passes through", "src#12 stays\n", "src#12 stays\n", 0, 0},
		{"third-repo passes through", "see third#9\n", "see third#9\n", 0, 0},
		{"gh ref reported not rewritten", "inbox gh#4 item\n", "inbox gh#4 item\n", 0, 1},
		{"qualified gh reported not rewritten", "see ariadne gh#4\n", "see ariadne gh#4\n", 0, 1},
		{"fenced block untouched", "pre #12\n```\nref #99\n```\npost\n", "pre src#12\n```\nref #99\n```\npost\n", 1, 0},
		{"unterminated fence untouched", "pre\n```\nbroken #99", "pre\n```\nbroken #99", 0, 0},
		{"single-ref span rewritten", "fixed in `#12` today\n", "fixed in `src#12` today\n", 1, 0},
		{"single-ref span with milestone rewritten", "see `dst#5 M2` there\n", "see `#5 M2` there\n", 1, 0},
		{"multi-token span skipped + reported", "run `git log --grep \"^#15\"` now\n", "run `git log --grep \"^#15\"` now\n", 0, 1},
		{"mixed-ref span skipped + reported", "pair `nous#41 #11` case\n", "pair `nous#41 #11` case\n", 0, 1},
		{"milestone form in prose", "closed #15 M4 already\n", "closed src#15 M4 already\n", 1, 0},
		{"heading no match", "## Log\n", "## Log\n", 0, 0},
		{"six-digit id", "see #000175\n", "see src#000175\n", 1, 0},
		{"seven digits no match", "hex #1234567 blob\n", "hex #1234567 blob\n", 0, 0},
		{"id zero skipped + reported", "weird #0 token\n", "weird #0 token\n", 0, 1},
		{"hex-alike is scanned and rewritten", "color #123456 pops\n", "color src#123456 pops\n", 1, 0},
		{"punctuation forms", "(#12) and #12, done\n", "(src#12) and src#12, done\n", 2, 0},
		{"multiple refs one line", "#1 then dst#5 then third#9\n", "src#1 then #5 then third#9\n", 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, rewrites, skipped := rewriteRefs(tc.in, "src", "dst")
			if out != tc.want {
				t.Errorf("out:\n got %q\nwant %q", out, tc.want)
			}
			if len(rewrites) != tc.wantRewrite {
				t.Errorf("rewrites = %d (%v), want %d", len(rewrites), rewrites, tc.wantRewrite)
			}
			if len(skipped) != tc.wantSkip {
				t.Errorf("skipped = %d (%v), want %d", len(skipped), skipped, tc.wantSkip)
			}
		})
	}
}

// TestRewriteRefs_LineNumbers pins that the rewrite report carries the
// 1-indexed line of each rewrite (the operator-review surface).
func TestRewriteRefs_LineNumbers(t *testing.T) {
	in := "line one\nsee #12 here\n```\n#99\n```\nand dst#5 last\n"
	_, rewrites, _ := rewriteRefs(in, "src", "dst")
	if len(rewrites) != 2 {
		t.Fatalf("want 2 rewrites, got %v", rewrites)
	}
	if rewrites[0].Line != 2 || rewrites[0].Old != "#12" || rewrites[0].New != "src#12" {
		t.Errorf("first rewrite = %+v, want line 2 #12→src#12", rewrites[0])
	}
	if rewrites[1].Line != 6 || rewrites[1].Old != "dst#5" || rewrites[1].New != "#5" {
		t.Errorf("second rewrite = %+v, want line 6 dst#5→#5", rewrites[1])
	}
}

// TestRefScan_GrammarRoundTrip: every NEW form a rewrite produces must parse
// under parseRef — true by construction (candidates are parseRef-filtered),
// and this test pins the construction.
func TestRefScan_GrammarRoundTrip(t *testing.T) {
	in := "#12 dst#5 src#12 third#9 `#7` #15 M4 #000175 (#3)\n"
	_, rewrites, _ := rewriteRefs(in, "src", "dst")
	if len(rewrites) == 0 {
		t.Fatal("fixture produced no rewrites")
	}
	for _, r := range rewrites {
		if _, err := parseRef(strings.TrimSpace(r.New)); err != nil {
			t.Errorf("rewritten form %q does not parse: %v", r.New, err)
		}
	}
}
