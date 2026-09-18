package flow

import (
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
			if fl.Kind != Full {
				t.Fatalf("error path returned kind %q, want full (the stricter flow), from %q", fl.Kind, fm)
			}
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
		rule Rule
	}{
		{"absent, no plan, no Mx → quick", DecideInput{}, Flow{Quick, Inferred, "", ""}, RuleNoPlan},
		{"absent, durable plan → full", DecideInput{HasPlan: true}, Flow{Full, Inferred, "", ""}, RulePlan},
		{"absent, Mx rows → full", DecideInput{HasMilestones: true}, Flow{Full, Inferred, "", ""}, RuleMilestones},
		{"absent, pin full", DecideInput{Pin: "full"}, Flow{Full, Operator, "", ""}, RulePinned},
		{"absent, pin quick", DecideInput{Pin: "quick"}, Flow{Quick, Operator, "", ""}, RulePinned},
		{"quick/inferred, plan appeared → full", DecideInput{Recorded: qi, HasPlan: true}, Flow{Full, Inferred, "", ""}, RulePlan},
		{"quick/inferred, unchanged → quick", DecideInput{Recorded: qi}, Flow{Quick, Inferred, "", ""}, RuleNoPlan},
		{"quick/inferred, Mx appeared → full", DecideInput{Recorded: qi, HasMilestones: true}, Flow{Full, Inferred, "", ""}, RuleMilestones},
		{"quick/operator stays even with a plan", DecideInput{Recorded: qo, HasPlan: true}, Flow{Quick, Operator, "", ""}, RuleOperatorStands},
		{"quick/operator, Mx appeared → full", DecideInput{Recorded: qo, HasMilestones: true}, Flow{Full, Inferred, "", ""}, RuleMilestones},
		{"full/inferred never downgrades", DecideInput{Recorded: fi}, Flow{Full, Inferred, "", ""}, RuleNoDowngrade},
		{"full/operator stays", DecideInput{Recorded: fo}, Flow{Full, Operator, "", ""}, RuleOperatorStands},
		{"full/inferred, pin quick", DecideInput{Recorded: fi, Pin: "quick"}, Flow{Quick, Operator, "", ""}, RulePinned},
		{"quick/operator, pin full", DecideInput{Recorded: qo, Pin: "full"}, Flow{Full, Operator, "", ""}, RulePinned},
		{"full/operator, Mx → stays full/operator", DecideInput{Recorded: fo, HasMilestones: true}, Flow{Full, Operator, "", ""}, RuleOperatorStands},
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

// flowRecordCorpus reads the shared accept/reject corpus both readers of the
// record assert (construct/vocabulary/testdata/flow_records.txt).
func flowRecordCorpus(t *testing.T) (accept, reject []string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "construct", "vocabulary", "testdata", "flow_records.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		verdict, value, ok := strings.Cut(line, "\t")
		switch {
		case ok && verdict == "accept":
			accept = append(accept, value)
		case ok && verdict == "reject":
			reject = append(reject, value)
		default:
			t.Fatalf("corpus line %q is not <accept|reject><TAB><value>", line)
		}
	}
	return accept, reject
}

// TestFlowRecordCorpus: the Go codec accepts exactly the corpus's accepts and
// rejects exactly its rejects. cmd/vocabulary asserts the same corpus against
// the cue model, so the two readers agree value for value (#231 BR-12).
func TestFlowRecordCorpus(t *testing.T) {
	accept, reject := flowRecordCorpus(t)
	if len(accept) == 0 || len(reject) == 0 {
		t.Fatal("corpus has no accepts or no rejects")
	}
	for _, v := range accept {
		if _, err := Parse(v); err != nil {
			t.Errorf("Parse(%q) rejected a value the corpus accepts: %v", v, err)
		}
	}
	for _, v := range reject {
		if f, err := Parse(v); err == nil {
			t.Errorf("Parse(%q) = %+v, but the corpus rejects it", v, f)
		}
	}
}

// TestRecorded: the decision input from frontmatter — nil when absent, the
// record when valid, and full/inferred plus the error when malformed.
func TestRecorded(t *testing.T) {
	if r, err := Recorded("id: 000001"); r != nil || err != nil {
		t.Errorf("absent: got %+v, %v; want nil, nil", r, err)
	}
	if r, err := Recorded("flow: {kind: quick, provenance: operator}"); err != nil || r == nil || *r != (Flow{Kind: Quick, Provenance: Operator}) {
		t.Errorf("valid: got %+v, %v", r, err)
	}
	if r, err := Recorded("flow: {kind: quikc}"); err == nil || r == nil || *r != (Flow{Kind: Full, Provenance: Inferred}) {
		t.Errorf("malformed: got %+v, %v; want full/inferred with an error", r, err)
	}
}
