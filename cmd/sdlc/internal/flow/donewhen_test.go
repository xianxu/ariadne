package flow

import (
	"strings"
	"testing"
)

const baseBody = `# T

## Problem

p

## Spec

The contract.

## Done when

- it works

## Plan

- [ ] do it

## Log

### 2026-09-17

started
`

// TestContractHashes: spec covers Spec + Revisions, done covers Done when, and
// edits elsewhere (Log, Plan ticks) change neither. A heading quoted inside a
// fence is content, not a section.
func TestContractHashes(t *testing.T) {
	spec0, done0 := ContractHashes(baseBody)
	if !hashRE.MatchString(spec0) || !hashRE.MatchString(done0) {
		t.Fatalf("hashes %q/%q are not 8 lowercase hex", spec0, done0)
	}
	edits := []struct {
		name                 string
		body                 string
		specMoves, doneMoves bool
	}{
		{"log entry", strings.Replace(baseBody, "started\n", "started\n\nmore notes\n", 1), false, false},
		{"plan tick", strings.Replace(baseBody, "- [ ] do it", "- [x] do it", 1), false, false},
		{"spec reframe", strings.Replace(baseBody, "The contract.", "A new contract.", 1), true, false},
		{"revisions appended", baseBody + "\n## Revisions\n\n### 2026-09-18 — reframe\n\nReason: x.\n", true, false},
		{"done when restated", strings.Replace(baseBody, "- it works", "- it works in both modes", 1), false, true},
		{"trailing whitespace only", strings.Replace(baseBody, "The contract.", "The contract.  \n", 1), false, false},
		{"fenced heading in the log", strings.Replace(baseBody, "started\n", "started\n\n```md\n## Spec\n\nquoted\n```\n", 1), false, false},
	}
	for _, e := range edits {
		spec, done := ContractHashes(e.body)
		if (spec != spec0) != e.specMoves {
			t.Errorf("%s: spec hash moved=%v, want %v", e.name, spec != spec0, e.specMoves)
		}
		if (done != done0) != e.doneMoves {
			t.Errorf("%s: done hash moved=%v, want %v", e.name, done != done0, e.doneMoves)
		}
	}
}

// TestDoneWhenPresent: the quick review's only oracle is a Done-when bullet; a
// populated related: frontmatter (which satisfies the structural gate) does not.
func TestDoneWhenPresent(t *testing.T) {
	if err := DoneWhenPresent(baseBody); err != nil {
		t.Errorf("bullet present: %v", err)
	}
	for name, body := range map[string]string{
		"empty seed": strings.Replace(baseBody, "- it works", "-", 1),
		"no section": strings.Replace(baseBody, "## Done when\n\n- it works\n", "", 1),
		"prose only": strings.Replace(baseBody, "- it works", "it works", 1),
	} {
		if err := DoneWhenPresent(body); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// TestDoneWhenFresh: a contract that moved without its acceptance criteria is
// refused — including a reframe recorded only under ## Revisions (#231 PQ-2).
func TestDoneWhenFresh(t *testing.T) {
	rec := WithContract(Flow{kind: Quick, provenance: Inferred}, baseBody)
	cases := []struct {
		name    string
		body    string
		refuses bool
	}{
		{"unchanged", baseBody, false},
		{"log-only edit", baseBody + "\nmore\n", false},
		{"spec changed, done unchanged", strings.Replace(baseBody, "The contract.", "Reframed.", 1), true},
		{"revisions-only reframe", baseBody + "\n## Revisions\n\n### x\n\nReason: y.\n", true},
		{"spec and done both changed", strings.Replace(strings.Replace(baseBody, "The contract.", "Reframed.", 1), "- it works", "- it works, reframed", 1), false},
		{"done changed alone", strings.Replace(baseBody, "- it works", "- it works better", 1), false},
	}
	for _, c := range cases {
		err := DoneWhenFresh(rec, c.body)
		if (err != nil) != c.refuses {
			t.Errorf("%s: err=%v, want refuses=%v", c.name, err, c.refuses)
		}
	}
	err := DoneWhenFresh(Flow{kind: Quick, provenance: Inferred}, baseBody)
	if err == nil || !strings.Contains(err.Error(), "sdlc change-code") || !strings.Contains(err.Error(), "--no-done-when-fresh") {
		t.Errorf("no anchor: err=%v, want a refusal naming both fixes", err)
	}
}
