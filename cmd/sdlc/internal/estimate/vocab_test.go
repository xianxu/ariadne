package estimate

import "testing"

func TestKnownPrimitive(t *testing.T) {
	if !KnownPrimitive("smaller-go-module") {
		t.Error("smaller-go-module should be a known primitive")
	}
	if KnownPrimitive("made-up-thing") {
		t.Error("made-up-thing should not be a known primitive")
	}
}

func TestKnownModel(t *testing.T) {
	if !KnownModel("estimate-logic-v2") {
		t.Error("estimate-logic-v2 should be a known model")
	}
	if KnownModel("vibes") {
		t.Error("vibes should not be a known model")
	}
}

func TestCurrentModel(t *testing.T) {
	if CurrentModel() != "estimate-logic-v3.1" {
		t.Fatalf("CurrentModel() = %q, want estimate-logic-v3.1", CurrentModel())
	}
	if !KnownModel(CurrentModel()) {
		t.Fatalf("CurrentModel() = %q is not recognized", CurrentModel())
	}
}

func TestPrimitivesSorted(t *testing.T) {
	got := Primitives()
	if len(got) != len(primitives) {
		t.Fatalf("Primitives() len = %d, want %d", len(got), len(primitives))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("Primitives() not sorted at %d: %q > %q", i, got[i-1], got[i])
		}
	}
}
