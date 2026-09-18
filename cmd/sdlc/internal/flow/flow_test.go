package flow

import (
	"strconv"
	"strings"
	"testing"
)

// TestFlowRoundTrip: Format then Parse is the identity, including for hashes
// that an unquoted YAML reader would retype — all digits read as an int and an
// `NNeNNNNN` shape reads as a float (#231 PQ-9). Format must quote them.
func TestFlowRoundTrip(t *testing.T) {
	cases := []Flow{
		{Kind: Full, Provenance: Inferred},
		{Kind: Quick, Provenance: Operator},
		{Kind: Quick, Provenance: Inferred, Spec: "1a2b3c4d", Done: "5e6f7a8b"},
		{Kind: Quick, Provenance: Inferred, Spec: "12345678", Done: "00000000"},
		{Kind: Quick, Provenance: Inferred, Spec: "12e45678", Done: "9e999999"},
	}
	for _, want := range cases {
		text := Format(want)
		if strings.Contains(text, "\n") {
			t.Errorf("Format(%+v) = %q spans lines; the frontmatter helpers are line-based", want, text)
		}
		got, err := Parse(text)
		if err != nil {
			t.Errorf("Parse(Format(%+v)) = %q: %v", want, text, err)
			continue
		}
		if got != want {
			t.Errorf("round trip: got %+v, want %+v (via %q)", got, want, text)
		}
	}
}

// TestParseRejects: every malformed shape is an error, never a silently
// defaulted value — callers resolve an error to full, the stricter flow.
func TestParseRejects(t *testing.T) {
	for _, bad := range []string{
		"",
		"quick",
		"[quick, inferred]",
		"{kind: quikc, provenance: inferred}",
		"{kind: quick, provenance: guessed}",
		"{kind: quick}",
		"{provenance: inferred}",
		"{kind: quick, provenance: inferred, spec: 12345678}", // unquoted: an int to cue
		"{kind: quick, provenance: inferred, spec: \"1A2B3C4D\"}",
		"{kind: quick, provenance: inferred, spec: \"abc\"}",
		"{kind: quick, provenance: inferred, color: blue}",
		"{kind: quick, kind: full, provenance: inferred}",
		"{kind: {nested: map}, provenance: inferred}",
	} {
		if f, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) = %+v, want an error", bad, f)
		}
	}
}

// TestFromFrontmatter covers the three states a reader meets: no record (reads
// as full, not recorded), a valid record, and a malformed one (an error).
func TestFromFrontmatter(t *testing.T) {
	f, recorded, err := FromFrontmatter("id: 000001\nstatus: working")
	if err != nil || recorded || f.Kind != Full {
		t.Errorf("absent: got %+v recorded=%v err=%v, want full, not recorded, nil", f, recorded, err)
	}
	f, recorded, err = FromFrontmatter("id: 000001\nflow: {kind: quick, provenance: operator}\nstatus: working")
	if err != nil || !recorded || f.Kind != Quick || f.Provenance != Operator {
		t.Errorf("valid: got %+v recorded=%v err=%v", f, recorded, err)
	}
	if _, recorded, err = FromFrontmatter("flow: {kind: quikc, provenance: inferred}"); err == nil || !recorded {
		t.Errorf("malformed: recorded=%v err=%v, want recorded with an error", recorded, err)
	}
	if _, _, err = FromFrontmatter("flow:"); err == nil {
		t.Error("empty value: want an error")
	}
}

// FuzzFromFrontmatter: hand-edited frontmatter is untrusted input. Whatever the
// bytes, reading it never panics, and a result is either an error or a Flow
// whose fields are all in their closed sets — never a half-valid record.
func FuzzFromFrontmatter(f *testing.F) {
	for _, seed := range []string{
		"flow: {kind: quick, provenance: inferred, spec: \"12345678\", done: \"5e6f7a8b\"}",
		"flow: {kind: quick, provenance: inferred, spec: 12345678}",
		"flow: {kind: quick, kind: full}",
		"flow: {kind: {a: b}}",
		"flow: {kind: quick, provenance: inferred",
		"flow: [a, b]",
		"flow: \"{kind: quick}\"",
		"flow: {kind: full, provenance: operator, done: \"deadbeef\"}",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, fm string) {
		fl, _, err := FromFrontmatter(fm)
		if err != nil {
			return
		}
		if fl.Kind != Full && fl.Kind != Quick {
			t.Fatalf("kind %q escaped the closed set from %q", fl.Kind, fm)
		}
		if fl.Provenance != "" && fl.Provenance != Inferred && fl.Provenance != Operator {
			t.Fatalf("provenance %q escaped the closed set from %q", fl.Provenance, fm)
		}
		for _, h := range []string{fl.Spec, fl.Done} {
			if h != "" && !hashRE.MatchString(h) {
				t.Fatalf("hash %q is not 8 lowercase hex, from %q", h, fm)
			}
		}
	})
}

// TestDecide: one case per cell of the plan's ARCH-ORDER table, plus the pin
// refusals. The invariant: only an operator pin produces quick from full, and
// nothing produces quick while Mx rows exist.
func TestDecide(t *testing.T) {
	qi := &Flow{Kind: Quick, Provenance: Inferred}
	qo := &Flow{Kind: Quick, Provenance: Operator}
	fi := &Flow{Kind: Full, Provenance: Inferred}
	fo := &Flow{Kind: Full, Provenance: Operator}
	cases := []struct {
		name string
		in   DecideInput
		want Flow
	}{
		{"absent, no plan, no Mx → quick", DecideInput{}, Flow{Quick, Inferred, "", ""}},
		{"absent, durable plan → full", DecideInput{HasPlan: true}, Flow{Full, Inferred, "", ""}},
		{"absent, Mx rows → full", DecideInput{HasMilestones: true}, Flow{Full, Inferred, "", ""}},
		{"absent, pin full", DecideInput{Pin: "full"}, Flow{Full, Operator, "", ""}},
		{"absent, pin quick", DecideInput{Pin: "quick"}, Flow{Quick, Operator, "", ""}},
		{"quick/inferred, plan appeared → full", DecideInput{Recorded: qi, HasPlan: true}, Flow{Full, Inferred, "", ""}},
		{"quick/inferred, unchanged → quick", DecideInput{Recorded: qi}, Flow{Quick, Inferred, "", ""}},
		{"quick/inferred, Mx appeared → full", DecideInput{Recorded: qi, HasMilestones: true}, Flow{Full, Inferred, "", ""}},
		{"quick/operator stays even with a plan", DecideInput{Recorded: qo, HasPlan: true}, Flow{Quick, Operator, "", ""}},
		{"quick/operator, Mx appeared → full", DecideInput{Recorded: qo, HasMilestones: true}, Flow{Full, Inferred, "", ""}},
		{"full/inferred never downgrades", DecideInput{Recorded: fi}, Flow{Full, Inferred, "", ""}},
		{"full/operator stays", DecideInput{Recorded: fo}, Flow{Full, Operator, "", ""}},
		{"full/inferred, pin quick", DecideInput{Recorded: fi, Pin: "quick"}, Flow{Quick, Operator, "", ""}},
		{"quick/operator, pin full", DecideInput{Recorded: qo, Pin: "full"}, Flow{Full, Operator, "", ""}},
		{"full/operator, Mx → stays full/operator", DecideInput{Recorded: fo, HasMilestones: true}, Flow{Full, Operator, "", ""}},
	}
	for _, c := range cases {
		got, err := Decide(c.in)
		if err != nil {
			t.Errorf("%s: unexpected error %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
	for _, bad := range []DecideInput{
		{Pin: "quick", HasMilestones: true},
		{Recorded: fi, Pin: "quick", HasMilestones: true},
		{Pin: "fast"},
	} {
		if got, err := Decide(bad); err == nil {
			t.Errorf("Decide(%+v) = %+v, want an error", bad, got)
		}
	}
}

// TestShellSummaryReadsTheConstants: the one sentence every surface prints is
// derived from the limits, so changing a limit cannot leave the prose behind.
func TestShellSummaryReadsTheConstants(t *testing.T) {
	s := ShellSummary()
	for _, want := range []string{strconv.Itoa(MaxCodeFiles) + " code files", strconv.Itoa(MaxChangedLines) + " added lines", "tests and docs", "shared surface", "milestones"} {
		if !strings.Contains(s, want) {
			t.Errorf("ShellSummary() = %q, missing %q", s, want)
		}
	}
}
