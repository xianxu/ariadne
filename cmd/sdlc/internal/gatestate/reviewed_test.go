package gatestate

import (
	"strings"
	"testing"
)

func TestLatestReviewed(t *testing.T) {
	a := strings.Repeat("a", 40)
	b := strings.Repeat("b", 40)
	seed := strings.Repeat("c", 40)
	l := Ledger{Gate: "boundary-review", IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: BoundaryAll, Reviewed: seed}, // impossible in production; must be ignored
		{N: 2, Boundary: "M1", Reviewed: a},
		{N: 3, Boundary: "M1"}, // a REWORK round carries no reviewed head
		{N: 4, Boundary: "M2", Reviewed: b},
	}}
	cases := []struct {
		name, exclude, sha, boundary string
		ok                           bool
	}{
		{"newest outside the current boundary", "M3", b, "M2", true},
		{"skips the current boundary", "M2", a, "M1", true},
		{"whole-issue close sees milestones", "", b, "M2", true},
		{"nothing before the first milestone", "M1", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ll := l
			if c.exclude == "M1" {
				ll.Rounds = l.Rounds[:3] // only the seed and M1 rounds exist yet
			}
			sha, boundary, ok := LatestReviewed(ll, c.exclude)
			if sha != c.sha || boundary != c.boundary || ok != c.ok {
				t.Fatalf("LatestReviewed(%q) = (%q, %q, %v), want (%q, %q, %v)", c.exclude, sha, boundary, ok, c.sha, c.boundary, c.ok)
			}
		})
	}
	if _, _, ok := LatestReviewed(Ledger{}, "M1"); ok {
		t.Fatal("an empty ledger has no reviewed head")
	}
}

func TestLatestFinalizedAt(t *testing.T) {
	a := strings.Repeat("a", 40)
	b := strings.Repeat("b", 40)
	m := strings.Repeat("d", 40)
	l := Ledger{Gate: "boundary-review", IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: BoundaryAll, Reviewed: m}, // never a whole-issue anchor
		{N: 2, Boundary: "", Reviewed: a},
		{N: 3, Boundary: "M1", Reviewed: m}, // a milestone's head is not the close's
		{N: 4, Boundary: ""},                // REWORK: no reviewed head
	}}
	if sha, ok := LatestFinalizedAt(l, ""); !ok || sha != a {
		t.Fatalf("LatestFinalizedAt = (%q, %v), want the round-2 close head", sha, ok)
	}
	l.Rounds = append(l.Rounds, Round{N: 5, Boundary: "", Reviewed: b})
	if sha, _ := LatestFinalizedAt(l, ""); sha != b {
		t.Fatalf("LatestFinalizedAt = %q, want the newest finalized close head", sha)
	}
	if _, ok := LatestFinalizedAt(Ledger{Rounds: []Round{{Boundary: ""}, {Boundary: "M1", Reviewed: m}}}, ""); ok {
		t.Fatal("no finalized whole-issue round: no anchor")
	}
}

// A round without a reviewed head must render byte-identically to its pre-#304 form,
// so no historical ledger changes on rewrite.
func TestRoundReviewedOmittedWhenEmpty(t *testing.T) {
	l := Ledger{Gate: "boundary-review", IDPrefix: "BR", IssueNum: 304, Rounds: []Round{{N: 1, Boundary: "M1", Agent: "claude", Timestamp: "t"}}}
	if out := Render(l, "repo"); strings.Contains(out, "reviewed") {
		t.Fatalf("empty Reviewed rendered a key:\n%s", out)
	}
	l.Rounds[0].Reviewed = strings.Repeat("d", 40)
	out := Render(l, "repo")
	if !strings.Contains(out, "reviewed: "+strings.Repeat("d", 40)) {
		t.Fatalf("Reviewed not rendered:\n%s", out)
	}
	back, err := ParseSidecar(out)
	if err != nil || back.Rounds[0].Reviewed != l.Rounds[0].Reviewed {
		t.Fatalf("round-trip lost Reviewed: %+v %v", back.Rounds, err)
	}
}
