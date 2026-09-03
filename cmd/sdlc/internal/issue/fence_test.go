package issue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFenceMarker walks the boundaries of the CommonMark rules the scanner
// implements: delimiter char, run length, indent, and the info string that
// distinguishes an opener from a closer.
func TestFenceMarker(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		wantOK     bool
		wantChar   byte
		wantWidth  int
		wantRest   string
	}{
		{"three backticks", "```", true, '`', 3, ""},
		{"with info string", "```markdown", true, '`', 3, "markdown"},
		{"tildes", "~~~", true, '~', 3, ""},
		{"four backticks", "````", true, '`', 4, ""},
		{"three-space indent is still a fence", "   ```", true, '`', 3, ""},
		{"four-space indent is NOT", "    ```", false, 0, 0, ""},
		{"two backticks is not a fence", "``", false, 0, 0, ""},
		{"inline code is not a fence", "text ``` more", false, 0, 0, ""},
		{"other char", "---", false, 0, 0, ""},
		{"empty", "", false, 0, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, w, rest, ok := fenceMarker(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if c != tc.wantChar || w != tc.wantWidth || rest != tc.wantRest {
				t.Errorf("= (%q, %d, %q), want (%q, %d, %q)", c, w, rest, tc.wantChar, tc.wantWidth, tc.wantRest)
			}
		})
	}
}

// TestSectionBody_FenceForms is the regression this issue exists for. Each case
// quotes a `##` heading the way a real issue does — specifying a markdown
// artifact means showing it — and the section must come back whole.
func TestSectionBody_FenceForms(t *testing.T) {
	body := func(inner string) string {
		return "# t\n\n## Spec\n\n" + inner + "\n## Plan\n\n- [ ] a\n\n## Log\n\nx\n"
	}
	for _, tc := range []struct {
		name, inner string
		wantTail    string // must survive to the end of ## Spec
	}{
		{"plain fence", "before\n```markdown\n## Quoted\n```\nafter\n", "after"},
		{"tilde fence", "before\n~~~markdown\n## Quoted\n~~~\nafter\n", "after"},
		{"wide fence holding a narrow line", "before\n````markdown\n```\n## Quoted\n```\n````\nafter\n", "after"},
		{"indented fence", "before\n  ```markdown\n  ## Quoted\n  ```\nafter\n", "after"},
		{"closed fence then a REAL heading", "before\n```\ncode\n```\nafter\n", "after"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sec, ok := SectionBody(body(tc.inner), "Spec")
			if !ok {
				t.Fatal("## Spec not found")
			}
			if !strings.Contains(sec, tc.wantTail) {
				t.Errorf("Spec truncated before %q:\n%s", tc.wantTail, sec)
			}
			// And the following sections must still be reachable — the failure
			// mode an unterminated fence would otherwise cause.
			if _, ok := SectionBody(body(tc.inner), "Plan"); !ok {
				t.Error("## Plan became unreachable")
			}
		})
	}
}

// TestSectionBody_UnterminatedFenceIsProse documents the price of the policy,
// so nobody "fixes" it later. Under UnterminatedIsProse a `##` after a stray
// opener IS read as a real heading, so the text following it lands in that
// accidental section rather than the one that opened the fence.
//
// That is the deliberate trade: over-segmenting a malformed document is
// recoverable and visible, while under-segmenting hides `## Plan` from the close
// gates. The prior behaviour lost the section anyway AND hid everything after
// it, so this is strictly better on both counts.
func TestSectionBody_UnterminatedFenceIsProse(t *testing.T) {
	body := "# t\n\n## Spec\n\nbefore\n```markdown\n## Quoted\nafter\n\n## Plan\n\n- [ ] a\n"
	spec, ok := SectionBody(body, "Spec")
	if !ok {
		t.Fatal("## Spec not found")
	}
	if strings.Contains(spec, "after") {
		t.Error("expected the stray-fence heading to end ## Spec — the policy accepts over-segmentation")
	}
	if _, ok := SectionBody(body, "Quoted"); !ok {
		t.Error("the quoted heading should be reachable as a real section under this policy")
	}
	if _, ok := SectionBody(body, "Plan"); !ok {
		t.Fatal("## Plan must remain reachable — that is the whole point of the policy")
	}
}

// TestSectionBody_UnterminatedFenceKeepsLaterSections pins the policy choice
// directly. A stray opener in prose must not hide `## Plan` — those are the
// sections the close gates read, so failing the other way turns one truncated
// section into a disarmed gate.
func TestSectionBody_UnterminatedFenceKeepsLaterSections(t *testing.T) {
	body := "# t\n\n## Spec\n\n```markdown\nnever closed\n\n## Plan\n\n- [ ] open item\n\n## Log\n\nx\n"
	plan, ok := SectionBody(body, "Plan")
	if !ok {
		t.Fatal("## Plan hidden by an unterminated fence — the close gates would see nothing")
	}
	if !strings.Contains(plan, "- [ ] open item") {
		t.Errorf("plan body missing its item:\n%s", plan)
	}
	// The opposite policy is what `project` and the migrate rewriter want.
	lines := strings.Split(body, "\n")
	if !FenceSpans(lines, UnterminatedIsFenced)[len(lines)-2] {
		t.Error("UnterminatedIsFenced should treat the unclosed tail as fenced")
	}
	if FenceSpans(lines, UnterminatedIsProse)[len(lines)-2] {
		t.Error("UnterminatedIsProse should treat the unclosed tail as prose")
	}
}

// TestFenceSpans_ReproducesInput is the property every consumer relies on:
// classifying lines must not lose or invent any. SplitFences' byte-exact
// reassembly guarantee rests on this.
func TestFenceSpans_ReproducesInput(t *testing.T) {
	for _, policy := range []UnterminatedPolicy{UnterminatedIsProse, UnterminatedIsFenced} {
		for _, f := range workshopMarkdown(t) {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(b), "\n")
			if got := len(FenceSpans(lines, policy)); got != len(lines) {
				t.Fatalf("%s: FenceSpans returned %d flags for %d lines", f, got, len(lines))
			}
		}
	}
}

// TestSectionBody_CorpusLosesNoRealSection is the corpus-seeded property test.
// The hand-written form table above is the cases I thought of; this is the cases
// the tree actually contains — and the first draft of this fix was caught by
// exactly such a case (an unterminated fence in prose) that no table had.
//
// Invariant: every heading a naive line scan finds OUTSIDE a fence must still be
// reachable through SectionBody. Headings inside fences are quoted examples and
// are SUPPOSED to disappear; those are the bug being fixed.
func TestSectionBody_CorpusLosesNoRealSection(t *testing.T) {
	files := workshopMarkdown(t)
	if len(files) == 0 {
		t.Skip("no workshop/ corpus here — a downstream repo consuming this package")
	}
	if len(files) < 100 {
		t.Fatalf("only %d workshop markdown files found — the corpus seed is mis-wired", len(files))
	}
	checked := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		body := string(b)
		lines := strings.Split(body, "\n")
		inside := FenceSpans(lines, UnterminatedIsProse)
		for i, line := range lines {
			if inside[i] || !strings.HasPrefix(line, "## ") {
				continue
			}
			h := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if h == "" {
				continue
			}
			if _, ok := SectionBody(body, h); !ok {
				t.Errorf("%s: real section %q is unreachable through SectionBody", f, h)
			}
			checked++
		}
	}
	t.Logf("verified %d real sections across %d workshop markdown files", checked, len(files))
}

func workshopMarkdown(t *testing.T) []string {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..", "workshop")
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // absent corpus is handled by the Skip in the caller
		}
		if !info.IsDir() && strings.HasSuffix(p, ".md") {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestSectionByteBounds_MatchesSectionBody pins the two extraction forms against
// each other: the byte-offset form exists only so a splicing caller can avoid
// re-implementing the search, so it must agree with the reading form exactly.
func TestSectionByteBounds_MatchesSectionBody(t *testing.T) {
	bodies := []string{
		"# t\n\n## Log\n\n- a\n\n## Plan\n\n- [ ] x\n",
		"# t\n\n## Problem\n\n```markdown\n## Log\nquoted\n```\n\n## Log\n\nreal\n",
		"# t\n\n## Log\n\ntrailing section, no following heading\n",
		"# t\n\n## Spec\n\n```\nunterminated\n\n## Log\n\nreached under prose policy\n",
	}
	for _, body := range bodies {
		want, okWant := SectionBody(body, "Log")
		start, end, okGot := SectionByteBounds(body, "Log", UnterminatedIsProse)
		if okGot != okWant {
			t.Fatalf("presence disagrees (byte %v vs body %v) for:\n%s", okGot, okWant, body)
		}
		if !okWant {
			continue
		}
		if got := strings.TrimSuffix(body[start:end], "\n"); got != strings.TrimSuffix(want, "\n") {
			t.Errorf("byte bounds %q != SectionBody %q\nfor:\n%s", got, want, body)
		}
	}
}

// TestUnterminatedPolicies_DisagreeOnPurpose pins the fork itself. These two
// consumers reach OPPOSITE conclusions about the same input, and that is the
// design — a test exists so a later reader "unifying" them has to argue with it.
//
//	stripCodeFences  a word-count gate: an unclosed fence is prose, so the gate
//	                 still sees the tail and can refuse on real content.
//	SplitFences      a file rewriter (#179 migrate): an unclosed tail stays
//	                 fenced, because editing inside maybe-code is unrecoverable.
func TestUnterminatedPolicies_DisagreeOnPurpose(t *testing.T) {
	const in = "prose here\n```\nnever closed but still words\n"

	if got := stripCodeFences(in); !strings.Contains(got, "never closed but still words") {
		t.Errorf("stripCodeFences must treat an unclosed tail as PROSE so the word count sees it; got %q", got)
	}
	var fencedTail bool
	for _, seg := range SplitFences(in) {
		if seg.Fenced && strings.Contains(seg.Text, "never closed") {
			fencedTail = true
		}
	}
	if !fencedTail {
		t.Error("SplitFences must treat an unclosed tail as FENCED so migrate won't rewrite inside it")
	}
}

// TestStripFenced_HidesQuotedPlanItems covers the counter filter: a `- [ ]` in a
// quoted example is not open work. Fails safe either way, but an issue quoting a
// plan format is precisely the kind this repo writes.
func TestStripFenced_HidesQuotedPlanItems(t *testing.T) {
	in := "- [x] real\n```markdown\n- [ ] quoted example\n```\n- [ ] also real\n"
	got := StripFenced(in)
	if strings.Contains(got, "quoted example") {
		t.Errorf("fenced plan item survived the strip:\n%s", got)
	}
	for _, want := range []string{"- [x] real", "- [ ] also real"} {
		if !strings.Contains(got, want) {
			t.Errorf("real item %q was stripped:\n%s", want, got)
		}
	}
}
