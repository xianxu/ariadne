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

// PlanItemsBody is the Plan section with fenced blocks removed — the body every
// consumer that COUNTS plan items must read.
//
// Since #211 M1 fenced content survives into the Plan body, so a `- [ ]` or
// `- [x] Mx` inside a quoted example is now visible to the item regexes. Filtering
// per-consumer is how `sdlc state` and `sdlc close` ended up disagreeing about
// the same Plan (M2 review BR-4): the filter reached CountPlanItems and nothing
// else. One extraction point, one answer.
//
// Every reader that COUNTS items uses this. The milestone tick is a writer and
// needs offsets into the real body, so it takes SectionByteBounds + FenceSpans
// directly (close.go) rather than this — but it applies the same two filters.
// An earlier version of this comment claimed the tick "rewrites a line it
// already matched", which was false: it ran ReplaceAll over the whole document.
func PlanItemsBody(body string) (string, bool) {
	section, ok := PlanSectionBody(body)
	if !ok {
		return "", false
	}
	return StripFenced(section), true
}

// SectionByteBounds is SectionLineBounds in byte offsets into body — the form a
// caller needs when it splices rather than reads (close's log insert).
//
// Returns the half-open [start, end) of the section BODY, excluding the heading
// line itself. Splitting and rejoining loses nothing here because both use "\n".
func SectionByteBounds(body, heading string, policy UnterminatedPolicy) (start, end int, ok bool) {
	lines := strings.Split(body, "\n")
	first, last, found := SectionLineBounds(lines, heading, policy)
	if !found {
		return 0, 0, false
	}
	off := func(line int) int {
		n := 0
		for i := 0; i < line && i < len(lines); i++ {
			n += len(lines[i]) + 1 // +1 for the newline split consumed
		}
		return n
	}
	start = off(first)
	if start > len(body) {
		start = len(body)
	}
	// off(last) points at the START of the next heading line, which is one byte
	// past the separator strings.Join does NOT emit after the final body line —
	// so drop it, keeping body[start:end] byte-identical to SectionBody's result.
	end = off(last) - 1
	if last >= len(lines) || end > len(body) {
		end = len(body)
	}
	if end < start {
		end = start
	}
	return start, end, true
}

// SectionHeadingByteOffset returns the byte offset of the `## <heading>` line
// itself (not its body), for callers that rewrite from the header down.
func SectionHeadingByteOffset(body, heading string, policy UnterminatedPolicy) (int, bool) {
	start, _, ok := SectionByteBounds(body, heading, policy)
	if !ok {
		return 0, false
	}
	// The body starts one line after the heading; walk back over that line.
	head := body[:start]
	if i := strings.LastIndex(strings.TrimRight(head, "\n"), "\n"); i >= 0 {
		return i + 1, true
	}
	return 0, true
}
