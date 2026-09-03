package issue

import "strings"

// SectionBody returns the body of a top-level `## <heading>` section — the text
// after `## <heading>` up to the next `## ` heading or end-of-text — and whether
// the section exists.
//
// FENCE-AWARE (#211). A `## ` inside a fenced code block is content, not the
// next section, so an issue that quotes markdown is read in full. That matters
// here more than it looks: this repo's deliverables are frequently markdown
// documents, so specifying one means quoting it, and the quoted headings are
// `##` because the target file's are. The regex this replaced ended a section at
// any `^## ` and silently truncated those issues — which for `## Plan` meant the
// close gates counted zero unchecked items and missed milestones entirely.
//
// The unterminated-fence policy is UnterminatedIsProse: a stray opener must
// never hide the rest of the document, because the sections it would hide are
// the ones the gates read.
func SectionBody(body, heading string) (string, bool) {
	lines := strings.Split(body, "\n")
	start, end, ok := SectionLineBounds(lines, heading, UnterminatedIsProse)
	if !ok {
		return "", false
	}
	return strings.Join(lines[start:end], "\n"), true
}

// SectionLineBounds returns the half-open line range beneath the first `##
// <heading>` that is not inside a fenced code block, ending at the next such
// heading or end-of-input.
//
// The single implementation of "find a section" (#211). An earlier cut of this
// file had its own copy of the loop while `project` kept another — which is the
// duplication this issue exists to remove, reintroduced one package over. The
// policy is a parameter precisely so both callers can share the loop while
// disagreeing about unterminated fences.
func SectionLineBounds(lines []string, heading string, policy UnterminatedPolicy) (start, end int, ok bool) {
	inside := FenceSpans(lines, policy)
	start, end = -1, len(lines)
	for i, line := range lines {
		if inside[i] || !strings.HasPrefix(line, "## ") {
			continue
		}
		if start >= 0 {
			return start, i, true
		}
		if strings.TrimSpace(strings.TrimPrefix(line, "## ")) == heading {
			start = i + 1
		}
	}
	return start, end, start >= 0
}

// PlanSectionBody is SectionBody for the Plan, named because five call sites
// want exactly this and used to reach for a separate regex to get it.
//
// PlanSectionRE is gone (#211): every consumer took FindStringSubmatchIndex only
// to slice body[m[2]:m[3]] and work on the resulting string — none needed the
// byte offsets the old comment here claimed checkPlan required.
func PlanSectionBody(body string) (string, bool) { return SectionBody(body, "Plan") }
