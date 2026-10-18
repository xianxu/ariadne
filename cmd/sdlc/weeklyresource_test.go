package main

import (
	"strings"
	"testing"
	"time"
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

func TestRenderWeeklyResourceSectionComparisons(t *testing.T) {
	observed := time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC)
	prior := previousResourceResult{State: "complete", Section: &WeeklyResourceSection{
		ObservedAt:     observed.Add(-7 * 24 * time.Hour),
		Families:       []ResourceComparison{{ResourceObservation: ResourceObservation{Kind: "couch-ledgers", Definition: "same", Bytes: 100, Complete: true}}},
		LargestLedgers: []ResourceFileComparison{{ResourceFile: ResourceFile{Path: "ledger-a", Bytes: 40}}},
		Suites:         []WeeklyResourceSuite{{Repo: "ariadne", Wall: &SuiteQuantiles{P50: 10, P90: 20, Samples: 2}}},
	}}
	current := ResourceCollection{
		Families:       []ResourceObservation{{Kind: "couch-ledgers", Definition: "same", Bytes: 130, Complete: true}},
		LargestLedgers: []ResourceFile{{Path: "ledger-a", Bytes: 50}, {Path: "ledger-b", Bytes: 70}},
	}
	repos := []WeeklyRepoStats{{Repo: "ariadne", Suite: RepoStats{Wall: &SuiteQuantiles{P50: 12, P90: 26, Samples: 3}}}}
	got := renderWeeklyResourceSection(current, repos, prior, observed, .25)
	if len(got.Families) != 1 || !got.Families[0].Alert || !sameInt64(got.Families[0].DeltaBytes, ptrInt64(30)) {
		t.Fatalf("family trend = %+v", got.Families)
	}
	if len(got.LargestLedgers) != 2 || got.LargestLedgers[0].Path != "ledger-b" || got.LargestLedgers[0].State != "unknown" || !sameInt64(got.LargestLedgers[1].DeltaBytes, ptrInt64(10)) {
		t.Fatalf("ledger trends = %+v", got.LargestLedgers)
	}
	if len(got.Suites) != 1 || got.Suites[0].State != "compared" || got.Suites[0].DeltaP50 == nil || *got.Suites[0].DeltaP50 != 2 || got.Suites[0].DeltaP90 == nil || *got.Suites[0].DeltaP90 != 6 {
		t.Fatalf("suite trend = %+v", got.Suites)
	}
	prior.State = "unknown"
	prior.Reason = "malformed prior"
	got = renderWeeklyResourceSection(current, repos, prior, observed, .25)
	if got.Families[0].State != "unknown" || got.Suites[0].State != "unknown" {
		t.Fatalf("malformed prior cannot be a baseline: %+v", got)
	}
}

func TestResourcePointPriorRequiresWeeklyObservationGap(t *testing.T) {
	observed := time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		gap   time.Duration
		state string
	}{
		{6 * 24 * time.Hour, "complete"},
		{7 * 24 * time.Hour, "complete"},
		{8 * 24 * time.Hour, "complete"},
		{24 * time.Hour, "unknown"},
		{9 * 24 * time.Hour, "unknown"},
	} {
		prior := previousResourceResult{State: "complete", Section: &WeeklyResourceSection{ObservedAt: observed.Add(-tc.gap)}}
		got := resourcePointPrior(prior, observed)
		if got.State != tc.state {
			t.Fatalf("gap %s: %+v, want %s", tc.gap, got, tc.state)
		}
		if tc.state == "unknown" && !strings.Contains(got.Reason, tc.gap.String()) {
			t.Fatalf("gap %s missing from reason %q", tc.gap, got.Reason)
		}
	}
}

func TestRenderWeeklyResourceSectionSuppressesPointDeltaButKeepsSuiteDelta(t *testing.T) {
	observed := time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC)
	prior := previousResourceResult{State: "complete", Section: &WeeklyResourceSection{
		ObservedAt:     observed.Add(-24 * time.Hour),
		Families:       []ResourceComparison{{ResourceObservation: ResourceObservation{Kind: "cache", Definition: "same", Bytes: 100, Complete: true}}},
		LargestLedgers: []ResourceFileComparison{{ResourceFile: ResourceFile{Path: "ledger", Bytes: 100}}},
		Suites:         []WeeklyResourceSuite{{Repo: "ariadne", Wall: &SuiteQuantiles{P50: 10, P90: 20}}},
	}}
	current := ResourceCollection{Families: []ResourceObservation{{Kind: "cache", Definition: "same", Bytes: 130, Complete: true}}, LargestLedgers: []ResourceFile{{Path: "ledger", Bytes: 130}}}
	repos := []WeeklyRepoStats{{Repo: "ariadne", Suite: RepoStats{Wall: &SuiteQuantiles{P50: 12, P90: 23}}}}
	got := renderWeeklyResourceSection(current, repos, prior, observed, .25)
	if !got.ObservedAt.Equal(observed) || got.Families[0].State != "unknown" || got.Families[0].DeltaBytes != nil || got.LargestLedgers[0].State != "unknown" || got.Suites[0].State != "compared" || got.Suites[0].DeltaP50 == nil {
		t.Fatalf("point/suite timing = %+v", got)
	}
}
