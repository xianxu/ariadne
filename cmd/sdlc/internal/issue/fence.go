// fence.go — the ONE fenced-code-block scanner for cmd/sdlc (#211).
//
// Before this file the tree had three, and they disagreed:
//
//	issue.fencedCodeRE   (?s)```.*?```      backticks only, no width rule
//	issue.SplitFences    strings.Index      backticks only, indent-blind
//	project.scanMarkdownLines                backticks + tildes, width-matched
//
// …while SectionBody and PlanSectionRE used none of them and simply ended a
// section at any `^## `, including one quoted inside a fence. That silently
// truncated sections, and because two close gates count things whose ABSENCE
// means pass (unchecked plan items, milestones lacking review evidence), the
// truncation could pass an issue that should have been refused.
//
// This is project.scanMarkdownLines — the only CommonMark-correct one — moved
// DOWN into `issue` (project imports issue, so it cannot go the other way) and
// given the one thing it lacked: an explicit unterminated-fence policy.
package issue

import "strings"

// UnterminatedPolicy decides what an opening fence with no closer means.
//
// This is a real fork, not a detail, and each consumer picks deliberately:
//
//	SectionBody, plan extraction  UnterminatedIsProse   a swallowed `## Plan`
//	                                                    disarms the close gates
//	stripCodeFences (word count)  UnterminatedIsProse   pre-existing, deliberate
//	SplitFences (#179 migrate)    UnterminatedIsFenced  a rewriter must not edit
//	                                                    inside a maybe-code tail
//	project section scan          UnterminatedIsFenced  pre-existing behavior
//
// Getting this wrong on SectionBody is worse than the bug it fixes: instead of
// one truncated section, every heading after the stray fence disappears. The
// issue that introduced this file demonstrated it on itself — a line of prose
// beginning with four backticks hid its own `## Plan`, `## Done when` and
// `## Log`.
type UnterminatedPolicy int

const (
	// UnterminatedIsProse: a fence that never closes was probably not a fence.
	// Fail toward seeing MORE of the document.
	UnterminatedIsProse UnterminatedPolicy = iota
	// UnterminatedIsFenced: a fence that never closes swallows the rest.
	// Fail toward touching LESS of the document.
	UnterminatedIsFenced
)

// FenceSpans returns, for each line, whether it lies inside a fenced code block
// (the fence delimiter lines themselves count as inside). Pure.
//
// Two passes, because the unterminated policy cannot be decided until the end of
// input is reached: the first pass records where fences open and close, the
// second applies the policy to any run left open.
func FenceSpans(lines []string, policy UnterminatedPolicy) []bool {
	inside := make([]bool, len(lines))
	var marker byte
	var width, openedAt int
	for i, line := range lines {
		m, w, rest, ok := fenceMarker(line)
		if !ok {
			if marker != 0 {
				inside[i] = true
			}
			continue
		}
		if marker == 0 {
			marker, width, openedAt = m, w, i
			inside[i] = true
			continue
		}
		// CommonMark: a closer matches the opener's character, is at least as
		// long, and carries no info string. Anything else is content — which is
		// how a ``` line inside a ```` block stays content.
		if m == marker && w >= width && strings.TrimSpace(rest) == "" {
			marker, width = 0, 0
		}
		inside[i] = true
	}
	if marker != 0 && policy == UnterminatedIsProse {
		// Unwind the unclosed run: it was never a fence.
		for i := openedAt; i < len(lines); i++ {
			inside[i] = false
		}
	}
	return inside
}

// ScanMarkdownLines calls visit for every line NOT inside a fenced code block.
// The line-oriented counterpart to FenceSpans, kept because most callers only
// want the prose lines and never the spans.
func ScanMarkdownLines(lines []string, policy UnterminatedPolicy, visit func(int, string)) {
	inside := FenceSpans(lines, policy)
	for i, line := range lines {
		if !inside[i] {
			visit(i, line)
		}
	}
}

// fenceMarker recognizes CommonMark-style backtick or tilde fences indented by
// at most three spaces, returning the delimiter char, its run length, whatever
// follows it (the info string on an opener), and whether the line is a fence at
// all. The caller owns opener/closer state.
func fenceMarker(line string) (byte, int, string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || len(trimmed) < 3 {
		return 0, 0, "", false
	}
	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return 0, 0, "", false
	}
	width := 0
	for width < len(trimmed) && trimmed[width] == marker {
		width++
	}
	return marker, width, trimmed[width:], width >= 3
}
