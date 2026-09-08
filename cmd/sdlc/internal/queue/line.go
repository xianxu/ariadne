// Package queue is the pure core of the advisory work queue (ariadne#209): the
// line format, the document round-trip, and the edit intents.
//
// Nothing here touches git, the filesystem, or the clock. The trunk-backed
// storage lives behind gitx.TrunkFile, which receives Intent.Apply as its
// transform — so the merge policy is decided here, in code that unit-tests
// without IO, and the CAS retry is decided there (ARCH-PURE).
package queue

import (
	"fmt"
	"strings"
)

// Kind distinguishes the two entry types the datatype defines.
type Kind int

const (
	// KindIssue is a next action — something that can be started.
	KindIssue Kind = iota
	// KindProject is a declared intent to work in an area. Coarser, and it
	// should be replaced by an issue line once it becomes the actual next
	// thing. A project line that never becomes actionable is permanently
	// present, carries no ordering information, and trains the reader to stop
	// reading the file.
	KindProject
)

// projectPrefix marks a project line.
//
// Issue and project refs are syntactically identical in this fleet (both are
// `repo#id`), so the kind cannot be inferred from the ref — it has to be
// declared. Making it visible in the text is also what lets a reader spot the
// failure mode above at a glance, which an implicit encoding would hide.
const projectPrefix = "project:"

// sep separates the ref from the why-now note. An em-dash, matching the
// datatype's examples; WhyNow may not contain one (see ValidateWhyNow), so the
// split is unambiguous.
const sep = " — "

// Line is one queue entry: a ref, a few words of why-now, and an optional
// project tag for grouping.
//
// A line that does not parse keeps its source in raw and renders back verbatim.
// The file is hand-editable, and silently dropping an operator's line is the
// worst failure this feature can have — so an unrecognized line is preserved,
// never discarded.
type Line struct {
	Ref    string
	WhyNow string
	Tag    string
	Kind   Kind

	// parsed is EXPLICIT, not inferred from raw being empty. Inferring it made a
	// blank source line — whose raw is also "" — look like a parsed entry, and
	// render as "-  — ". Two states in one field that cannot tell them apart is
	// the same defect M1 shipped three times; here the fuzz target caught it on
	// the second seed.
	parsed bool
	raw    string // the verbatim source line, when !parsed
	cr     bool   // the source line ended CRLF; String restores it
}

// Parsed reports whether this line was understood.
func (l Line) Parsed() bool { return l.parsed }

// ParseLine parses one queue line. An unrecognized line is returned with raw set
// and ok false; it still renders verbatim.
func ParseLine(s string) (Line, bool) {
	unparsed := Line{raw: s}
	trimmed, cr := strings.CutSuffix(s, "\r")
	body, isItem := strings.CutPrefix(trimmed, "- ")
	if !isItem {
		return unparsed, false
	}
	ref, why, found := strings.Cut(body, sep)
	if !found {
		return unparsed, false
	}
	ref, why = strings.TrimSpace(ref), strings.TrimSpace(why)
	if ref == "" || why == "" {
		return unparsed, false
	}

	l := Line{Kind: KindIssue, parsed: true, cr: cr}
	if r, ok := strings.CutPrefix(ref, projectPrefix); ok {
		l.Kind, ref = KindProject, strings.TrimSpace(r)
		if ref == "" {
			return unparsed, false
		}
	}
	l.Ref = ref

	// A trailing [tag] is the project grouping. Only a tag that closes at the
	// very end counts, so a bracket inside the prose does not become one.
	if strings.HasSuffix(why, "]") {
		if i := strings.LastIndex(why, "["); i > 0 {
			if tag := strings.TrimSpace(why[i+1 : len(why)-1]); tag != "" {
				l.Tag = tag
				why = strings.TrimSpace(why[:i])
			}
		}
	}
	if why == "" {
		return unparsed, false
	}
	l.WhyNow = why

	// SELF-CHECK: a parse is only accepted if it renders back byte-identically.
	//
	// The parse trims whitespace in five places, and every one of them is lossy:
	// "-  a#1 — why" and "- a#1 — why  " both parse to the same fields and would
	// render normalized, silently rewriting the operator's line. Fixing each trim
	// is possible and was the first instinct; this is better, because it makes the
	// round-trip invariant hold BY CONSTRUCTION rather than by five separate
	// arguments that a sixth trim could quietly break. Anything that does not
	// round-trip is preserved verbatim instead — which is the behavior the file
	// needs anyway.
	if l.String() != s {
		return unparsed, false
	}
	return l, true
}

// String renders the line. Round-trips: String(ParseLine(x)) == x for any x.
func (l Line) String() string {
	if !l.parsed {
		return l.raw
	}
	ref := l.Ref
	if l.Kind == KindProject {
		ref = projectPrefix + ref
	}
	s := "- " + ref + sep + l.WhyNow
	if l.Tag != "" {
		s += " [" + l.Tag + "]"
	}
	if l.cr {
		s += "\r"
	}
	return s
}

// ValidateWhyNow rejects text that would break the format, BEFORE any git call.
//
// These are rejections, not escapes. The format's whole value is that a human
// reads it as a list, and escaping would cost exactly that — a why-now needing
// an escape is a why-now that wants to be an issue body.
func ValidateWhyNow(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("why-now must not be empty — an order you cannot evaluate is worse than no order")
	}
	for _, r := range s {
		if r == '\n' || r == '\r' {
			return fmt.Errorf("why-now must not contain a newline: it would render as a second queue entry")
		}
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("why-now must not contain control characters (found %q)", r)
		}
	}
	if strings.Contains(s, sep) {
		return fmt.Errorf("why-now must not contain %q — it is the ref separator and would re-split the line", sep)
	}
	if strings.Contains(s, "[") || strings.Contains(s, "]") {
		return fmt.Errorf("why-now must not contain brackets — a trailing [...] is the project tag")
	}
	return nil
}

// ValidateTag rejects a tag that would break the format.
//
// Tag was the one user-supplied field with no validator, which is the whole
// lesson here: Ref and WhyNow were guarded and Tag was not, because the guard was
// written per-field as each was added rather than from an enumeration of the
// fields the format interpolates. A newline in --tag published a forged entry and
// exited 0.
func ValidateTag(s string) error {
	if s == "" {
		return nil // absent is fine; the tag is optional
	}
	if strings.TrimSpace(s) != s || s == "" {
		return fmt.Errorf("tag must not have surrounding whitespace")
	}
	for _, r := range s {
		if r == '\n' || r == '\r' || r < 0x20 || r == 0x7f {
			return fmt.Errorf("tag must not contain newlines or control characters")
		}
	}
	if strings.ContainsAny(s, "[]") || strings.Contains(s, sep) {
		return fmt.Errorf("tag %q contains a character the line format reserves", s)
	}
	return nil
}

// ValidateRef rejects a ref that would break the format.
func ValidateRef(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("ref must not be empty")
	}
	if s != strings.TrimSpace(s) {
		return fmt.Errorf("ref must not have surrounding whitespace")
	}
	if strings.ContainsAny(s, " \t\n\r[]") || strings.Contains(s, sep) {
		return fmt.Errorf("ref %q contains a character the line format reserves", s)
	}
	return nil
}
