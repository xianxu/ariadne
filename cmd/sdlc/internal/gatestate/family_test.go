package gatestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// #194 M3 D3 mechanism 2: normalization catches casing and punctuation drift, and
// NOTHING else. Stated as a test so the limit is visible rather than assumed.
func TestNormalizeFamily(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"block-opener-rule", "block-opener-rule"},
		{"Block Opener Rule", "block-opener-rule"},
		{"block_opener_rule", "block-opener-rule"},
		{"  BLOCK--OPENER  ", "block-opener"},
		{"oracle blind direction!", "oracle-blind-direction"},
		{"", ""},
	} {
		if got := NormalizeFamily(tc.in); got != tc.want {
			t.Errorf("NormalizeFamily(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFamilyCounts_NormalizesAndSpansBoundaries(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: "M1", New: []Finding{
			{ID: "BR-1", Severity: "Critical", Title: "paren case", Family: "block-opener-rule"},
		}},
		{N: 2, Boundary: "M1", New: []Finding{
			{ID: "BR-2", Severity: "Important", Title: "bracket case", Family: "Block Opener Rule"},
		}},
		// A family recurring at a DIFFERENT boundary is the signal one issue-wide ledger
		// exists to preserve — tools#1's rule spanned M1 rounds and the close review.
		{N: 3, Boundary: "", New: []Finding{
			{ID: "BR-3", Severity: "Important", Title: "plain prose case", Family: "block_opener_rule"},
		}},
		{N: 4, Boundary: "M1", New: []Finding{
			{ID: "BR-4", Severity: "Minor", Title: "unrelated", Family: "doc-drift"},
		}},
	}}
	counts := FamilyCounts(l)
	if counts["block-opener-rule"] != 3 {
		t.Errorf("three spellings of one family must collapse to 3, got %d (%v)", counts["block-opener-rule"], counts)
	}
	if counts["doc-drift"] != 1 {
		t.Errorf("doc-drift = %d, want 1", counts["doc-drift"])
	}
}

// The residual risk D3 accepted, pinned by its own test name: a genuine SYNONYM is not
// caught by normalization. Mechanism 1 — rendering the in-play vocabulary into the
// prompt with a reuse instruction — is what makes this unlikely, not impossible.
func TestFamilyCounts_TrueSynonymsAreNotMerged_AcceptedResidualRisk(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, New: []Finding{{ID: "BR-1", Severity: "Critical", Title: "a", Family: "block-opener-rule"}}},
		{N: 2, New: []Finding{{ID: "BR-2", Severity: "Critical", Title: "b", Family: "block-opener"}}},
	}}
	counts := FamilyCounts(l)
	if counts["block-opener-rule"] != 1 || counts["block-opener"] != 1 {
		t.Fatalf("documenting the limit: synonyms stay separate, got %v", counts)
	}
}

// #194 M3 D3 mechanism 1: the reviewer must SEE which families are already in play, or
// it cannot reuse a slug — and escalation never fires.
func TestRenderPriorFindings_ShowsFamilyVocabularyAndEscalates(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, New: []Finding{{ID: "BR-1", Severity: "Critical", Title: "paren case", Family: "block-opener-rule"}}},
		{N: 2, Dispositions: []Disposition{{ID: "BR-1", State: "addressed", Round: 2}}},
	}}
	got := RenderPriorFindings(l)
	if !strings.Contains(got, "block-opener-rule") {
		t.Error("the in-play family vocabulary must be rendered so the reviewer can reuse a slug")
	}
	if !strings.Contains(strings.ToLower(got), "reuse") {
		t.Error("the vocabulary must carry a reuse instruction, not just a list")
	}
	// One prior disposed finding in the family ⇒ the next is the 2nd ⇒ escalate.
	for _, want := range []string{"2nd finding in family", "state the rule"} {
		if !strings.Contains(got, want) {
			t.Errorf("a repeat family must escalate from fix-this-instance to state-the-rule; missing %q:\n%s", want, got)
		}
	}
}

func TestConvergenceLine(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, New: []Finding{
			{ID: "BR-1", Severity: "Critical", Title: "a", Family: "block-opener-rule"},
			{ID: "BR-2", Severity: "Important", Title: "b", Family: "oracle-blind"},
		}},
		{N: 2,
			Dispositions: []Disposition{{ID: "BR-1", State: "addressed", Round: 2}, {ID: "BR-2", State: "addressed", Round: 2}},
			New:          []Finding{{ID: "BR-3", Severity: "Minor", Title: "c", Family: "block-opener-rule"}},
		},
	}}
	// Pin the EXACT shape (#194 M3 review): a Contains check passes whether or not the
	// segments the helptext advertises are present, so it cannot keep the docs honest.
	// It is also terminal output — no markdown emphasis.
	if got, want := ConvergenceLine(l, 2), "round 2 — 1 new finding, 1 repeat family, 2 disposed. Not converging: fix rules, not instances."; got != want {
		t.Errorf("convergence line shape drifted:\n got: %q\nwant: %q", got, want)
	}

	// No repeat families ⇒ converging.
	clean := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, New: []Finding{{ID: "BR-1", Severity: "Minor", Title: "a", Family: "one"}}},
		{N: 2, Dispositions: []Disposition{{ID: "BR-1", State: "addressed", Round: 2}},
			New: []Finding{{ID: "BR-2", Severity: "Minor", Title: "b", Family: "two"}}},
	}}
	if got, want := ConvergenceLine(clean, 2), "round 2 — 1 new finding, 0 repeat families, 1 disposed. Converging."; got != want {
		t.Errorf("converging shape drifted:\n got: %q\nwant: %q", got, want)
	}
}

// #194 M3 acceptance test, and the only one that measures whether the feature would have
// WORKED. The fixture is a real four-round history (tools#1); the question is whether a
// correct implementation names `block-opener-rule` as a repeat at ROUND 2 — where a human
// reviewer would have said "you are patching cases" — rather than at round 3, which is
// when it was actually noticed.
func TestFamilyEscalation_AgainstRealFourRoundHistory(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "tools-1-m1-families.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var full Ledger
	if err := yaml.Unmarshal(raw, &full); err != nil {
		t.Fatalf("fixture does not parse as a ledger: %v", err)
	}
	if len(full.Rounds) != 4 {
		t.Fatalf("fixture should carry four rounds, got %d", len(full.Rounds))
	}

	// What round 2's reviewer would have been shown, i.e. the ledger after round 1.
	afterRound1 := full
	afterRound1.Rounds = full.Rounds[:1]
	round2Prompt := RenderPriorFindings(afterRound1)

	if !strings.Contains(round2Prompt, "block-opener-rule") {
		t.Fatal("round 2 must see round 1's family in the in-play vocabulary")
	}
	if !strings.Contains(round2Prompt, "2nd finding in family") {
		t.Errorf("round 2 must escalate to 'state the rule' — that is one round earlier "+
			"than tools#1 managed, and the whole claim of this milestone:\n%s", round2Prompt)
	}
	if !strings.Contains(round2Prompt, "state the rule") {
		t.Error("the escalation must change the recommendation, not merely count")
	}

	// The convergence signal at each round, read off the real shape of the history.
	if got := ConvergenceLine(full, 2); !strings.Contains(got, "Not converging") {
		t.Errorf("round 2 repeated block-opener-rule twice — it was NOT converging: %q", got)
	}
	if got := ConvergenceLine(full, 4); !strings.Contains(got, "Not converging") {
		t.Errorf("round 4 repeated BOTH families — still not converging: %q", got)
	}

	// Across the whole history the two families are visible with their real weights.
	counts := FamilyCounts(full)
	if counts["block-opener-rule"] != 4 {
		t.Errorf("block-opener-rule recurred four times across the real history, got %d", counts["block-opener-rule"])
	}
	if counts["oracle-blind-direction"] != 2 {
		t.Errorf("oracle-blind-direction recurred twice, got %d", counts["oracle-blind-direction"])
	}
}

// #194 M3 review BR-22: ConvergenceLine must count only STRICTLY EARLIER rounds as
// prior. With `!= round`, replaying an older round reports repeats against families that
// did not exist yet — the signal reads "not converging" for work that was converging.
func TestConvergenceLine_LaterRoundsAreNotPriorFamilies(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, New: []Finding{{ID: "BR-1", Severity: "Minor", Title: "a", Family: "alpha"}}},
		{N: 2, New: []Finding{{ID: "BR-2", Severity: "Minor", Title: "b", Family: "beta"}}},
		{N: 3, New: []Finding{{ID: "BR-3", Severity: "Minor", Title: "c", Family: "beta"}}},
	}}
	// Round 2 is the FIRST beta — round 3's beta is later and must not count as prior.
	got := ConvergenceLine(l, 2)
	if strings.Contains(got, "Not converging") {
		t.Errorf("round 2 introduced beta; a LATER round's beta must not make it a repeat: %q", got)
	}
	if !strings.Contains(got, "0 repeat families") {
		t.Errorf("round 2 should report zero repeat families: %q", got)
	}
	// Round 3 genuinely repeats beta.
	if got := ConvergenceLine(l, 3); !strings.Contains(got, "1 repeat family") {
		t.Errorf("round 3 repeats beta: %q", got)
	}
}

// #194 close review BR-29: family must survive the durable round-trip, and the emitted
// fence must actually name the key — otherwise the judge is never asked for it and every
// count stays zero while the code looks correct.
func TestFamily_SurvivesRoundTripAndIsNamedInTheFence(t *testing.T) {
	l := Ledger{Gate: "boundary-review", IssueNum: 194, IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: "M1", New: []Finding{
			{ID: "BR-1", Severity: "Critical", Title: "t", Family: "block-opener-rule", Round: 1},
		}},
	}}
	got, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil {
		t.Fatalf("ParseSidecar: %v", err)
	}
	if fam := got.Rounds[0].New[0].Family; fam != "block-opener-rule" {
		t.Errorf("family lost in round-trip: %q", fam)
	}
	// The human projection shows it too (BR-28) — a recurrence invisible to a reader is
	// half a feature.
	if rendered := Render(l, "ariadne"); !strings.Contains(rendered, "`block-opener-rule`") {
		t.Errorf("the prose projection must show the family:\n%s", rendered)
	}
}

// #194 close review BR-27/BR-39: the displayed round number must skip no-cap rounds, so a
// seed round the binary wrote does not make the first real review read as "round 2". The
// first version of this fix shipped unpinned.
func TestConvergenceLine_RoundNumberSkipsNoCapRounds(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: BoundaryAll, NoCap: true, New: []Finding{
			{ID: "BR-1", Severity: "Minor", Title: "seeded", Family: "carried"},
		}},
		{N: 2, Boundary: "M1", New: []Finding{{ID: "BR-2", Severity: "Minor", Title: "first real"}}},
	}}
	if got := ConvergenceLine(l, 2); !strings.HasPrefix(got, "round 1 — ") {
		t.Errorf("the first REVIEWED round must display as round 1, not 2: %q", got)
	}
}
