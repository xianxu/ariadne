package flakes

import "testing"

func TestCouldNotRunNeverExplainsCode(t *testing.T) {
	neutral := Attempt{ID: "b2", SHA: "head", Kind: Kind("could-not-run")}
	failed := Attempt{ID: "a1", SHA: "head", Kind: Unattributable}
	if neutral.Kind.Lands() {
		t.Fatal("could-not-run lands")
	}
	if Explainable(neutral.Kind) {
		t.Fatal("neutral execution evidence became a registrable code failure")
	}
	if got := Unexplained([]Attempt{neutral}, "head", Registry{}); len(got) != 0 {
		t.Fatalf("neutral run blocks a real rerun: %+v", got)
	}
	if got := Unexplained([]Attempt{failed, neutral}, "head", Registry{}); len(got) != 1 || got[0].ID != "a1" {
		t.Fatalf("neutral run explained code failure: %+v", got)
	}
	if _, ok := Confirmed([]Attempt{failed, neutral}, 0); ok {
		t.Fatal("neutral run confirmed intermittency")
	}
	if _, ok := Confirmed([]Attempt{neutral, {ID: "c3", SHA: "head", Passed: true}}, 0); ok {
		t.Fatal("neutral run became a failure confirmed by a pass")
	}
}
