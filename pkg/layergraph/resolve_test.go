package layergraph

import (
	"reflect"
	"testing"
)

// Resolve returns layers in foundation-first topological order (a node's
// dependencies appear before the node), deduped — the order setup.sh's
// discover_ancestors produces. Ported behavior, ARCH-DRY.

func TestResolveFoundationFirst(t *testing.T) {
	// brain → {ariadne, nous}; nous → ariadne. Foundation-first: ariadne, nous, brain.
	deps := map[string][]string{
		"ariadne": {},
		"nous":    {"ariadne"},
		"brain":   {"ariadne", "nous"},
	}
	got, err := Resolve("brain", deps)
	if err != nil {
		t.Fatalf("Resolve: unexpected error: %v", err)
	}
	want := []string{"ariadne", "nous", "brain"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Resolve = %v, want %v", got, want)
	}
}

func TestResolveDiamondDedupes(t *testing.T) {
	// D → {B, C}; B → A; C → A. A is reached via two paths but must be
	// applied once, before B/C, and D last (foundation-first).
	deps := map[string][]string{
		"A": {},
		"B": {"A"},
		"C": {"A"},
		"D": {"B", "C"},
	}
	got, err := Resolve("D", deps)
	if err != nil {
		t.Fatalf("Resolve: unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("Resolve = %v, want 4 unique nodes (A applied once)", got)
	}
	pos := map[string]int{}
	for i, n := range got {
		if _, dup := pos[n]; dup {
			t.Fatalf("Resolve = %v: %s emitted twice", got, n)
		}
		pos[n] = i
	}
	if pos["A"] != 0 || pos["D"] != 3 {
		t.Fatalf("Resolve = %v, want A first and D last", got)
	}
	if pos["B"] < pos["A"] || pos["C"] < pos["A"] || pos["D"] < pos["B"] || pos["D"] < pos["C"] {
		t.Fatalf("Resolve = %v violates foundation-first ordering", got)
	}
}

func TestResolveCycleErrors(t *testing.T) {
	deps := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	if _, err := Resolve("a", deps); err == nil {
		t.Fatalf("Resolve: expected a cycle error, got nil")
	}
}
