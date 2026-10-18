package main

import (
	"testing"
)

func TestCompareResourceBytes(t *testing.T) {
	old := ResourceObservation{Kind: "couch-ledgers", Definition: "inventory-v1:a", Bytes: 100, Complete: true}
	for _, tc := range []struct {
		name    string
		current ResourceObservation
		prior   *ResourceObservation
		state   string
		delta   *int64
		alert   bool
	}{
		{"baseline", old, nil, "baseline", nil, false},
		{"growth", ResourceObservation{Kind: old.Kind, Definition: old.Definition, Bytes: 130, Complete: true}, &old, "compared", ptrInt64(30), true},
		{"decline", ResourceObservation{Kind: old.Kind, Definition: old.Definition, Bytes: 80, Complete: true}, &old, "compared", ptrInt64(-20), false},
		{"changed definition", ResourceObservation{Kind: old.Kind, Definition: "inventory-v2:b", Bytes: 130, Complete: true}, &old, "unknown", nil, false},
		{"partial", ResourceObservation{Kind: old.Kind, Definition: old.Definition, Bytes: 130, Complete: false, Reason: "unreadable file"}, &old, "unknown", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compareResourceBytes(tc.current, tc.prior, .25)
			if got.State != tc.state || got.Alert != tc.alert || !sameInt64(got.DeltaBytes, tc.delta) {
				t.Fatalf("comparison = %+v, want state=%s delta=%v alert=%v", got, tc.state, tc.delta, tc.alert)
			}
		})
	}
	zero := ResourceObservation{Kind: old.Kind, Definition: old.Definition, Complete: true}
	got := compareResourceBytes(ResourceObservation{Kind: old.Kind, Definition: old.Definition, Bytes: 1, Complete: true}, &zero, .25)
	if got.State != "new" || got.GrowthShare != nil || got.Alert {
		t.Fatalf("zero baseline must not imply infinite growth: %+v", got)
	}
}

func ptrInt64(n int64) *int64 { return &n }

func sameInt64(a, b *int64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
