package flakes

import (
	"strings"
	"testing"
	"time"
)

func typedEvidencePair(name string) (Attempt, Attempt) {
	id, _ := ParseIdentity(name)
	failFacts := &UnitFacts{Seen: true, Failures: []string{name}}
	passFacts := &UnitFacts{Seen: true, Units: map[string]UnitResult{
		id.Unit(): {Passed: true, Complete: true, Cases: map[string]string{id.Case(): "pass"}},
	}}
	failed := Attempt{ID: "a", SHA: "head", Failures: []string{name}, FinishedAt: time.Unix(1, 0), TypedFacts: failFacts, RunnerContractSHA256: strings.Repeat("a", 64)}
	passed := Attempt{ID: "b", SHA: "head", Passed: true, FinishedAt: time.Unix(2, 0), TypedFacts: passFacts, RunnerContractSHA256: failed.RunnerContractSHA256}
	return failed, passed
}

func TestTypedEvidenceRequiresExactRawProof(t *testing.T) {
	for _, name := range []string{"script:tests/check.sh", "lua:tests/check_spec.lua#suite case/opaque #label"} {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				mutate func(*Attempt, *Attempt)
				want   bool
			}{
				{"exact proof", func(*Attempt, *Attempt) {}, true},
				{"missing failure facts", func(a, b *Attempt) { a.TypedFacts = nil }, false},
				{"wrong failed identity", func(a, b *Attempt) { a.TypedFacts.Failures = []string{"script:tests/other.sh"} }, false},
				{"unattributed failure", func(a, b *Attempt) { a.TypedFacts.Unattributed = []string{"incomplete"} }, false},
				{"missing pass facts", func(a, b *Attempt) { b.TypedFacts = nil }, false},
				{"unattributed pass", func(a, b *Attempt) { b.TypedFacts.Unattributed = []string{"incomplete"} }, false},
				{"absent unit", func(a, b *Attempt) { b.TypedFacts.Units = nil }, false},
				{"incomplete unit", func(a, b *Attempt) {
					for k, u := range b.TypedFacts.Units {
						u.Complete = false
						b.TypedFacts.Units[k] = u
					}
				}, false},
				{"unit did not pass", func(a, b *Attempt) {
					for k, u := range b.TypedFacts.Units {
						u.Passed = false
						b.TypedFacts.Units[k] = u
					}
				}, false},
				{"missing failed contract", func(a, b *Attempt) { a.RunnerContractSHA256 = "" }, false},
				{"missing passed contract", func(a, b *Attempt) { b.RunnerContractSHA256 = "" }, false},
				{"different contract", func(a, b *Attempt) { b.RunnerContractSHA256 = strings.Repeat("b", 64) }, false},
				{"different head", func(a, b *Attempt) { b.SHA = "other" }, false},
				{"subset pass", func(a, b *Attempt) { b.Subset = true }, false},
				{"baseline pass", func(a, b *Attempt) { b.Baseline = true }, false},
				{"legacy pass", func(a, b *Attempt) { b.Legacy = true }, false},
				{"rejected pass", func(a, b *Attempt) { b.Passed = false }, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a, b := typedEvidencePair(name)
					tc.mutate(&a, &b)
					if f, p, ok := SameHeadFailPass([]Attempt{b, a}, name); ok != tc.want || ok && (f != a.ID || p != b.ID) {
						t.Fatalf("proof = %q, %q, %v; want %v", f, p, ok, tc.want)
					}
					if _, ok := Confirmed([]Attempt{a, b}, 0); ok != tc.want {
						t.Fatalf("whole confirmation = %v; want %v", ok, tc.want)
					}
				})
			}
		})
	}
}

func TestTypedEvidenceLuaNeedsExactCaseAndWholeFile(t *testing.T) {
	name := "lua:tests/check_spec.lua#suite selected"
	for _, result := range []string{"", "skip", "fail"} {
		a, b := typedEvidencePair(name)
		b.TypedFacts.Units["lua:tests/check_spec.lua"].Cases["suite selected"] = result
		b.TypedFacts.Units["lua:tests/check_spec.lua"].Cases["suite sibling"] = "pass"
		if _, _, ok := SameHeadFailPass([]Attempt{a, b}, name); ok {
			t.Fatalf("selected case %q confirmed by sibling", result)
		}
	}
}

func TestTypedEvidenceWholeConfirmationRequiresAllTypedFailures(t *testing.T) {
	one, two := "script:tests/one.sh", "lua:tests/two_spec.lua#suite selected"
	a, b := typedEvidencePair(one)
	a.Failures = append(a.Failures, two, tA)
	a.TypedFacts.Failures = append(a.TypedFacts.Failures, two)
	if _, ok := Confirmed([]Attempt{a, b}, 0); ok {
		t.Fatal("partial typed pass confirmed whole attempt")
	}
	if _, _, ok := SameHeadFailPass([]Attempt{a, b}, one); !ok {
		t.Fatal("exact typed proof lost because another failure lacks proof")
	}
	if _, _, ok := SameHeadFailPass([]Attempt{a, b}, tA); !ok {
		t.Fatal("legacy Go confirmation acquired typed proof requirements")
	}
	b.TypedFacts.Units["lua:tests/two_spec.lua"] = UnitResult{Passed: true, Complete: true, Cases: map[string]string{"suite selected": "pass"}}
	if p, ok := Confirmed([]Attempt{a, b}, 0); !ok || p != b.ID {
		t.Fatalf("complete proof = %q, %v", p, ok)
	}
}

func TestTypedEvidenceSkipsEarlierUnprovenPass(t *testing.T) {
	name := "script:tests/check.sh"
	a, b := typedEvidencePair(name)
	earlier := b
	earlier.ID, earlier.TypedFacts = "ab", nil
	earlier.FinishedAt = a.FinishedAt.Add(time.Millisecond)
	if f, p, ok := SameHeadFailPass([]Attempt{earlier, b, a}, name); !ok || f != a.ID || p != b.ID {
		t.Fatalf("proof = %q,%q,%v", f, p, ok)
	}
}

func TestTypedEvidencePassingRecordCannotHideAnotherRawFailure(t *testing.T) {
	name := "script:tests/check.sh"
	failed, passed := typedEvidencePair(name)
	passed.TypedFacts.Failures = []string{"script:tests/other.sh"}
	if _, _, ok := SameHeadFailPass([]Attempt{failed, passed}, name); ok {
		t.Fatal("raw failing suite confirmed typed flake through exit-only pass")
	}
	if _, ok := Confirmed([]Attempt{failed, passed}, 0); ok {
		t.Fatal("raw failing suite confirmed whole attempt")
	}
}
