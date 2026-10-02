package recovery

import (
	"strings"
	"testing"
)

func TestCatalogIsWellFormed(t *testing.T) {
	if err := Validate(Catalog); err != nil {
		t.Fatal(err)
	}
	for verb := range Exempt {
		if _, ok := For(verb); ok {
			t.Errorf("%s is both contracted and exempt", verb)
		}
	}
}

func TestValidateRefusesMalformedContracts(t *testing.T) {
	good := Contract{Verbs: []string{"x"}, Class: ReadOnly, Effects: "e", Evidence: "v", Preconditions: "p", Repeat: "r", LostResponse: "l", Ends: "n",
		Proofs: []Proof{{Claim: "c", Tests: []string{"TestX"}}}}
	for name, spoil := range map[string]func(*Contract){
		"no verb":          func(c *Contract) { c.Verbs = nil },
		"unknown class":    func(c *Contract) { c.Class = "idempotent" },
		"no evidence":      func(c *Contract) { c.Evidence = " " },
		"no lost response": func(c *Contract) { c.LostResponse = "" },
		"no proofs":        func(c *Contract) { c.Proofs = nil },
		"empty claim":      func(c *Contract) { c.Proofs = []Proof{{Tests: []string{"TestX"}}} },
	} {
		c := good
		spoil(&c)
		if err := Validate([]Contract{c}); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if err := Validate([]Contract{good, good}); err == nil || !strings.Contains(err.Error(), "two contracts") {
		t.Errorf("duplicate verb: %v", err)
	}
}

// An unproven claim renders as unknown, never as a guarantee.
func TestSectionMarksUnprovenClaims(t *testing.T) {
	s := Section("pr")
	if !strings.Contains(s, "RECOVERY (#280) — convergent-retry") || !strings.Contains(s, "UNPROVEN (no test; treat as unknown)") ||
		!strings.Contains(s, "Scope:") || !strings.Contains(s, "legacy") || !strings.Contains(s, "unknown") {
		t.Fatalf("section:\n%s", s)
	}
	if s := Section("no-such-verb"); !strings.Contains(s, "no contract recorded") {
		t.Fatalf("missing verb: %s", s)
	}
}
