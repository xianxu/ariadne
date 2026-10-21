package landingline

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestClassifySplit(t *testing.T) {
	positive := SplitFailure{Test: "example/touched/TestBug", Impact: ImpactAffected}
	registered := SplitFailure{Test: "example/flake/TestFlaky", RegisteredIssue: "42", Impact: ImpactAffected}
	untouched := SplitFailure{Test: "example/other/TestOther", Impact: ImpactUntouched}
	unknown := SplitFailure{Test: "example/unknown/TestMystery", Impact: ImpactUnknown}
	cases := []struct {
		name       string
		failures   []SplitFailure
		unverified string
		kind       AttributionKind
		contains   []string
		absent     []string
	}{
		{name: "empty cannot attribute", kind: AttributionUnverified},
		{name: "affected exact test", failures: []SplitFailure{positive}, kind: AttributionAttributable, contains: []string{positive.Test}},
		{name: "global error vetoes affected evidence", failures: []SplitFailure{positive}, unverified: "metadata unavailable", kind: AttributionUnverified, contains: []string{"metadata unavailable"}},
		{name: "registration explains affected test", failures: []SplitFailure{registered}, kind: AttributionUnattributed, contains: []string{"registered flake " + registered.Test + " (issue #42)"}},
		{name: "untouched explains exact test", failures: []SplitFailure{untouched}, kind: AttributionUnattributed, contains: []string{"untouched package " + untouched.Test}},
		{name: "unknown refuses attribution", failures: []SplitFailure{unknown}, kind: AttributionUnverified, contains: []string{unknown.Test}},
		{name: "positive survives unrelated uncertainty", failures: []SplitFailure{unknown, registered, untouched, positive}, kind: AttributionAttributable, contains: []string{positive.Test}, absent: []string{unknown.Test, registered.Test, untouched.Test}},
		{name: "unresolved fact prevents all explained verdict", failures: []SplitFailure{registered, untouched, unknown}, kind: AttributionUnverified, contains: []string{unknown.Test}},
		{name: "mixed explanations retain both categories", failures: []SplitFailure{registered, untouched}, kind: AttributionUnattributed, contains: []string{"registered flake " + registered.Test, "untouched package " + untouched.Test}},
		{name: "empty test is invalid", failures: []SplitFailure{{Impact: ImpactAffected}}, kind: AttributionUnverified},
		{name: "invalid name vetoes positive", failures: []SplitFailure{positive, {RegisteredIssue: "42", Impact: ImpactUntouched}}, kind: AttributionUnverified},
		{name: "invalid impact vetoes positive", failures: []SplitFailure{positive, {Test: "bad/TestBad", Impact: Impact(99)}}, kind: AttributionUnverified},
		{name: "registered invalid impact is still malformed", failures: []SplitFailure{{Test: "bad/TestBad", RegisteredIssue: "42", Impact: Impact(-1)}}, kind: AttributionUnverified},
		{name: "name grammar belongs to caller", failures: []SplitFailure{{Test: "nonempty", Impact: ImpactAffected}}, kind: AttributionAttributable, contains: []string{"nonempty"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifySplit(tc.failures, tc.unverified)
			if got.Kind != tc.kind || got.Reason == "" {
				t.Fatalf("ClassifySplit = %+v, want %s with reason", got, tc.kind)
			}
			for _, fragment := range tc.contains {
				if !strings.Contains(got.Reason, fragment) {
					t.Errorf("reason %q missing %q", got.Reason, fragment)
				}
			}
			for _, fragment := range tc.absent {
				if strings.Contains(got.Reason, fragment) {
					t.Errorf("reason %q names nonsupporting evidence %q", got.Reason, fragment)
				}
			}
		})
	}
}

func TestClassifySplitRegistrationOverridesEveryValidImpact(t *testing.T) {
	for _, impact := range []Impact{ImpactUnknown, ImpactUntouched, ImpactAffected} {
		got := ClassifySplit([]SplitFailure{{Test: "pkg/TestFlaky", RegisteredIssue: "7", Impact: impact}}, "")
		if got.Kind != AttributionUnattributed || got.Reason != "unattributed: registered flake pkg/TestFlaky (issue #7)" {
			t.Errorf("impact %d: %+v", impact, got)
		}
	}
}

func TestClassifySplitDeterministicWithoutMutation(t *testing.T) {
	cases := [][]SplitFailure{
		{{Test: "z/TestZ", Impact: ImpactAffected}, {Test: "a/TestA", Impact: ImpactAffected}},
		{{Test: "z/TestZ", Impact: ImpactUnknown}, {Test: "a/TestA", Impact: ImpactUnknown}},
		{{Test: "z/TestZ", RegisteredIssue: "2"}, {Test: "a/TestA", RegisteredIssue: "9"}, {Test: "a/TestA", RegisteredIssue: "1"}, {Test: "y/TestY", Impact: ImpactUntouched}, {Test: "b/TestB", Impact: ImpactUntouched}},
		{{Test: "z/TestZ", Impact: Impact(99)}, {Test: "a/TestA", Impact: Impact(-1)}, {Impact: ImpactAffected}},
	}
	for _, failures := range cases {
		original := slices.Clone(failures)
		want := ClassifySplit(failures, "")
		if !reflect.DeepEqual(failures, original) {
			t.Fatal("classifier mutated input")
		}
		for i := range failures {
			rotated := append(slices.Clone(failures[i:]), failures[:i]...)
			slices.Reverse(rotated)
			before := slices.Clone(rotated)
			if got := ClassifySplit(rotated, ""); got != want {
				t.Errorf("order changed result: %+v, want %+v", got, want)
			}
			if !reflect.DeepEqual(rotated, before) {
				t.Fatal("classifier mutated reordered input")
			}
		}
	}
	got := ClassifySplit(cases[0], "")
	if got.Reason != "attributable: affected package a/TestA" {
		t.Errorf("positive reason = %q", got.Reason)
	}
}
