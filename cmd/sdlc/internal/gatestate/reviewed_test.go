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
