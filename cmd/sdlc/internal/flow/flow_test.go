package flow

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestFlowRoundTrip: Format then Parse is the identity, including for a hash
// that an unquoted YAML reader would retype — all digits read as an int (#231
// PQ-9) — so Format must quote it. The `NNeNNNNN` case stays as a shape worth
// round-tripping, though both readers were measured to read it as a string.
func TestFlowRoundTrip(t *testing.T) {
	cases := []Flow{
		{kind: Full, provenance: Inferred},
		{kind: Quick, provenance: Operator},
		{kind: Quick, provenance: Inferred, spec: "1a2b3c4d", done: "5e6f7a8b"},
		{kind: Quick, provenance: Inferred, spec: "12345678", done: "00000000"},
		{kind: Quick, provenance: Inferred, spec: "12e45678", done: "9e999999"},
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

// TestFromFrontmatter covers the three states a reader meets: no record (reads
// as full, not recorded), a valid record, and a malformed one (an error).
func TestFromFrontmatter(t *testing.T) {
	f, recorded, err := FromFrontmatter("id: 000001\nstatus: working")
	if err != nil || recorded || f.kind != Full {
		t.Errorf("absent: got %+v recorded=%v err=%v, want full, not recorded, nil", f, recorded, err)
	}
	f, recorded, err = FromFrontmatter("id: 000001\nflow: {kind: quick, provenance: operator}\nstatus: working")
	if err != nil || !recorded || f.kind != Quick || f.provenance != Operator {
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
			if fl.kind != Full {
				t.Fatalf("error path returned kind %q, want full (the stricter flow), from %q", fl.kind, fm)
			}
			return
		}
		if fl.kind != Full && fl.kind != Quick {
			t.Fatalf("kind %q escaped the closed set from %q", fl.kind, fm)
		}
		if fl.provenance != "" && fl.provenance != Inferred && fl.provenance != Operator {
			t.Fatalf("provenance %q escaped the closed set from %q", fl.provenance, fm)
		}
		for _, h := range []string{fl.spec, fl.done} {
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
	qi := &Flow{kind: Quick, provenance: Inferred}
	qo := &Flow{kind: Quick, provenance: Operator}
	fi := &Flow{kind: Full, provenance: Inferred}
	fo := &Flow{kind: Full, provenance: Operator}
	cases := []struct {
		name string
		in   DecideInput
		want Flow
		rule Rule
	}{
		{"absent, no plan, no Mx → quick", DecideInput{}, Flow{Quick, Inferred, "", ""}, ruleNoPlan},
		{"absent, durable plan → full", DecideInput{HasPlan: true}, Flow{Full, Inferred, "", ""}, rulePlan},
		{"absent, Mx rows → full", DecideInput{HasMilestones: true}, Flow{Full, Inferred, "", ""}, ruleMilestones},
		{"absent, pin full", DecideInput{Pin: "full"}, Flow{Full, Operator, "", ""}, rulePinned},
		{"absent, pin quick", DecideInput{Pin: "quick"}, Flow{Quick, Operator, "", ""}, rulePinned},
		{"quick/inferred, plan appeared → full", DecideInput{Recorded: qi, HasPlan: true}, Flow{Full, Inferred, "", ""}, rulePlan},
		{"quick/inferred, unchanged → quick", DecideInput{Recorded: qi}, Flow{Quick, Inferred, "", ""}, ruleNoPlan},
		{"quick/inferred, Mx appeared → full", DecideInput{Recorded: qi, HasMilestones: true}, Flow{Full, Inferred, "", ""}, ruleMilestones},
		{"quick/operator stays even with a plan", DecideInput{Recorded: qo, HasPlan: true}, Flow{Quick, Operator, "", ""}, ruleOperatorStands},
		{"quick/operator, Mx appeared → full", DecideInput{Recorded: qo, HasMilestones: true}, Flow{Full, Inferred, "", ""}, ruleMilestones},
		{"full/inferred never downgrades", DecideInput{Recorded: fi}, Flow{Full, Inferred, "", ""}, ruleNoDowngrade},
		{"full/operator stays", DecideInput{Recorded: fo}, Flow{Full, Operator, "", ""}, ruleOperatorStands},
		{"full/inferred, pin quick", DecideInput{Recorded: fi, Pin: "quick"}, Flow{Quick, Operator, "", ""}, rulePinned},
		{"quick/operator, pin full", DecideInput{Recorded: qo, Pin: "full"}, Flow{Full, Operator, "", ""}, rulePinned},
		{"full/operator, Mx → stays full/operator", DecideInput{Recorded: fo, HasMilestones: true}, Flow{Full, Operator, "", ""}, ruleOperatorStands},
	}
	for _, c := range cases {
		got, rule, err := Decide(c.in)
		if err != nil {
			t.Errorf("%s: unexpected error %v", c.name, err)
			continue
		}
		if got != c.want || rule != c.rule {
			t.Errorf("%s: got %+v by %q, want %+v by %q", c.name, got, rule, c.want, c.rule)
		}
	}
	for _, bad := range []DecideInput{
		{Pin: "quick", HasMilestones: true},
		{Recorded: fi, Pin: "quick", HasMilestones: true},
		{Pin: "fast"},
	} {
		if got, _, err := Decide(bad); err == nil {
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

// corpusRow is one line of the shared accept/reject corpus both readers of
// the record assert (construct/vocabulary/testdata/flow_records.txt).
type corpusRow struct {
	accept bool
	reason Reason // the refusing branch, for a reject row
	value  string
}

func flowRecordCorpus(t *testing.T) []corpusRow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "construct", "vocabulary", "testdata", "flow_records.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []corpusRow
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		verdict, value, ok := strings.Cut(line, "\t")
		switch {
		case ok && verdict == "accept":
			rows = append(rows, corpusRow{accept: true, value: value})
		case ok && strings.HasPrefix(verdict, "reject:"):
			rows = append(rows, corpusRow{reason: Reason(strings.TrimPrefix(verdict, "reject:")), value: value})
		default:
			t.Fatalf("corpus line %q is not accept<TAB>value or reject:<reason><TAB>value", line)
		}
	}
	return rows
}

// TestFlowRecordCorpus: the Go codec accepts exactly the corpus's accepts and
// rejects each reject through the branch the row names; and every reject
// branch has a row, so none exists unpinned. cmd/vocabulary asserts the same
// corpus against the cue model, so the two readers agree value for value
// (#231 BR-12, BR-17).
func TestFlowRecordCorpus(t *testing.T) {
	covered := map[Reason]bool{}
	for _, row := range flowRecordCorpus(t) {
		f, err := Parse(row.value)
		if row.accept {
			if err != nil {
				t.Errorf("Parse(%q) rejected a value the corpus accepts: %v", row.value, err)
			}
			continue
		}
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("Parse(%q) = %+v, %v; the corpus rejects it for %s", row.value, f, err, row.reason)
			continue
		}
		if pe.Reason != row.reason {
			t.Errorf("Parse(%q) refused for %s, the corpus says %s", row.value, pe.Reason, row.reason)
		}
		covered[pe.Reason] = true
	}
	for _, r := range Reasons {
		if !covered[r] {
			t.Errorf("reject branch %s has no corpus row — add one to flow_records.txt", r)
		}
	}
}

// TestRecorded: the decision input from frontmatter — nil when absent, the
// record when valid, and full/inferred plus the error when malformed.
func TestRecorded(t *testing.T) {
	if r, err := Recorded("id: 000001"); r != nil || err != nil {
		t.Errorf("absent: got %+v, %v; want nil, nil", r, err)
	}
	if r, err := Recorded("flow: {kind: quick, provenance: operator}"); err != nil || r == nil || *r != (Flow{kind: Quick, provenance: Operator}) {
		t.Errorf("valid: got %+v, %v", r, err)
	}
	if r, err := Recorded("flow: {kind: quikc}"); err == nil || r == nil || *r != (Flow{kind: Full, provenance: Inferred}) {
		t.Errorf("malformed: got %+v, %v; want full/inferred with an error", r, err)
	}
}
