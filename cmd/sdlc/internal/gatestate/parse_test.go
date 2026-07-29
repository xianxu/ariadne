package gatestate

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

const goodBlock = "prose the judge wrote first\n\n" +
	"```findings\n" +
	"dispose:\n" +
	"  - id: PQ-1\n    disposition: addressed\n    note: seam moved to the filter\n" +
	"findings:\n" +
	"  - id: new\n    severity: Critical\n    title: absorb layer swallows solicited replies\n" +
	"    detail: capability negotiation breaks silently\n" +
	"```\n"

func TestParseFindingsBlock(t *testing.T) {
	rr, ok := ParseFindingsBlock(goodBlock)
	if !ok {
		t.Fatal("valid block should parse")
	}
	if len(rr.Dispositions) != 1 || rr.Dispositions[0].ID != "PQ-1" ||
		rr.Dispositions[0].State != "addressed" || rr.Dispositions[0].Note == "" {
		t.Errorf("dispositions = %+v", rr.Dispositions)
	}
	if len(rr.New) != 1 || rr.New[0].Severity != "Critical" || rr.New[0].ID != "new" ||
		rr.New[0].Detail == "" {
		t.Errorf("findings = %+v", rr.New)
	}
}

// LAST block wins — a judge that shows an example block before its real one must not hand
// us the example (the ParseVerdictBlock precedent).
func TestParseFindingsBlockLastWins(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: example\n```\n" + goodBlock
	rr, ok := ParseFindingsBlock(in)
	if !ok || len(rr.New) != 1 || rr.New[0].Severity != "Critical" {
		t.Errorf("last block should win, got %+v", rr.New)
	}
}

// Fail-closed: an unmodeled severity is a protocol error, not a value to guess.
func TestParseFindingsBlockRejectsUnmodeledSeverity(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n    severity: Catastrophic\n    title: x\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("an unmodeled severity must not parse")
	}
}

func TestParseFindingsBlockRejectsUnmodeledDisposition(t *testing.T) {
	in := "```findings\ndispose:\n  - id: PQ-1\n    disposition: maybe\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("an unmodeled disposition must not parse")
	}
}

// A finding with no title is unusable in the ledger and in the next round's prompt.
func TestParseFindingsBlockRejectsTitlelessFinding(t *testing.T) {
	for _, in := range []string{
		"```findings\nfindings:\n  - id: new\n    severity: Minor\n```\n",
		"```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: \"   \"\n```\n",
	} {
		if _, ok := ParseFindingsBlock(in); ok {
			t.Errorf("a titleless finding must not parse: %q", in)
		}
	}
}

// A disposition with no id cannot be applied to anything.
func TestParseFindingsBlockRejectsIDlessDisposition(t *testing.T) {
	in := "```findings\ndispose:\n  - id: \"\"\n    disposition: addressed\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("an id-less disposition must not parse")
	}
}

func TestParseFindingsBlockAbsent(t *testing.T) {
	if _, ok := ParseFindingsBlock("VERDICT: CLEAN\n\nlooks good"); ok {
		t.Error("no block ⇒ ok=false (caller falls back + warns)")
	}
}

func TestParseFindingsBlockMalformedYAML(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n   severity: [unbalanced\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("unparseable YAML must not yield ok=true")
	}
}

// An EMPTY block is a valid, meaningful statement: "no findings this round".
func TestParseFindingsBlockEmptyIsValid(t *testing.T) {
	rr, ok := ParseFindingsBlock("```findings\nfindings: []\n```\n")
	if !ok {
		t.Fatal("an explicitly empty block must parse")
	}
	if len(rr.New) != 0 || len(rr.Dispositions) != 0 {
		t.Errorf("got %+v", rr)
	}
}

// FuzzParseFindingsBlock is the MECHANICAL guard this issue's own C1 demands of every
// future plan, applied to the plan's riskiest function.
//
// ParseFindingsBlock is a parser over UNBOUNDED LLM output — `title` and `detail` are
// free-form agent text. Enumerated well-formed cases are exactly the pathology #187
// indicts: on pair#127, 30 hand-written cases all fed syntactically valid sequences and
// the close review still found a panic on malformed input.
//
// Invariants: never panic, and never return ok=true carrying a finding whose severity is
// unmodeled or whose title is blank — the two properties every downstream consumer
// (AssignIDs, Decide, Render) assumes without re-checking.
func FuzzParseFindingsBlock(f *testing.F) {
	f.Add(goodBlock)
	f.Add("```findings\nfindings: []\n```\n")
	// Seeds targeting the specific structural hazards, not random noise.
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: |\n      block scalar\n```\n")
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: x\n    detail: \"has\\n---\\nfence\"\n```\n")
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: x\n    detail: \"```findings nested\"\n```\n")
	f.Add("```findings\ndispose:\n  - id: \"\"\n    disposition: addressed\n```\n")
	f.Add("````findings\nfindings: []\n````\n")
	f.Add("```findings\n```\n")

	m := vocab.Finding()
	f.Fuzz(func(t *testing.T, s string) {
		rr, ok := ParseFindingsBlock(s) // must not panic
		if !ok {
			return
		}
		for _, fd := range rr.New {
			if !m.IsSeverity(fd.Severity) {
				t.Fatalf("ok=true with unmodeled severity %q", fd.Severity)
			}
			if strings.TrimSpace(fd.Title) == "" {
				t.Fatalf("ok=true with a blank title: %+v", fd)
			}
		}
		for _, d := range rr.Dispositions {
			if !m.IsDisposition(d.State) {
				t.Fatalf("ok=true with unmodeled disposition %q", d.State)
			}
			if strings.TrimSpace(d.ID) == "" {
				t.Fatalf("ok=true with a blank disposition id: %+v", d)
			}
		}
	})
}

// TestParseFindingsBlockBlockScalarSurvivesHash pins the fidelity hazard found in live use
// on ariadne#187 round 1: in a YAML PLAIN scalar a ` #` begins a comment, so a finding
// whose text contains "## Estimate" or "issue #187" is silently truncated — and the
// truncated text is what the NEXT round is shown as its own prior finding, so the gate's
// memory quietly degrades.
//
// The block instruction now shows title/detail/note as `|` block scalars, which are immune.
// This test pins both halves: block form survives, plain form does not — so if anyone
// "simplifies" the template back to plain scalars, the second assertion documents why not.
func TestParseFindingsBlockBlockScalarSurvivesHash(t *testing.T) {
	blockForm := "```findings\nfindings:\n  - id: new\n    severity: Minor\n" +
		"    title: |\n      drop the \"seed a reconciling ## Estimate block\" option\n" +
		"    detail: |\n      see issue #187 for why\n```\n"
	rr, ok := ParseFindingsBlock(blockForm)
	if !ok {
		t.Fatal("block-scalar form must parse")
	}
	if !strings.Contains(rr.New[0].Title, "## Estimate block") {
		t.Errorf("block scalar truncated the title: %q", rr.New[0].Title)
	}
	if !strings.Contains(rr.New[0].Detail, "#187") {
		t.Errorf("block scalar truncated the detail: %q", rr.New[0].Detail)
	}

	// The hazard itself, documented: plain form loses everything from ` #' onward.
	plainForm := "```findings\nfindings:\n  - id: new\n    severity: Minor\n" +
		"    title: drop the reconciling ## Estimate block option\n```\n"
	plain, ok := ParseFindingsBlock(plainForm)
	if !ok {
		t.Fatal("plain form still parses — it is lossy, not invalid")
	}
	if strings.Contains(plain.New[0].Title, "Estimate") {
		t.Error("plain scalar unexpectedly preserved text after ` #' — re-check the template rationale")
	}
}
