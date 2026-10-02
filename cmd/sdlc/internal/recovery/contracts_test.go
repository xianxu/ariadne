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

func TestLookupReadsObservationPaths(t *testing.T) {
	doc := map[string]any{"card": map[string]any{"status": "working"},
		"checkpoints": map[string]any{"reviews": []any{map[string]any{"boundary": "M1", "verdict": "SHIP"}, map[string]any{"boundary": "close", "verdict": "REWORK", "open_blocking": float64(2)}}}}
	for path, want := range map[string]string{"card.status": "working", "checkpoints.reviews[close].verdict": "REWORK", "checkpoints.reviews[close].open_blocking": "2"} {
		if got, ok := Lookup(doc, path); !ok || got != want {
			t.Errorf("%s = %q %v, want %q", path, got, ok, want)
		}
	}
	for _, path := range []string{"card.missing", "checkpoints.reviews[M9].verdict", "card.status.deeper"} {
		if _, ok := Lookup(doc, path); ok {
			t.Errorf("%s resolved", path)
		}
	}
}

func TestPageRenderingCoversEveryContractAndStep(t *testing.T) {
	table, example := Table(), ExampleText()
	for _, c := range Catalog {
		if !strings.Contains(table, c.Verbs[0]) {
			t.Errorf("table lacks %s", c.Verbs[0])
		}
	}
	for _, s := range Example {
		if s.Command != "" && !strings.Contains(example, s.Command) {
			t.Errorf("example lacks %q", s.Command)
		}
	}
	for _, c := range Classes {
		if !strings.Contains(ClassesText(), string(c)) {
			t.Errorf("classes lack %s", c)
		}
	}
}
