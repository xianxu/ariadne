package main

import (
	"reflect"
	"testing"
	"time"
)

func TestSuiteStatsAggregationCoverageQuantilesAndFlakes(t *testing.T) {
	t0 := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	f := func(x float64) *float64 { return &x }
	shaA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	shaB := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	rows := []SuiteAttempt{
		{Repo: "ariadne", SHA: shaA, Check: "suite", StartedAt: t0.Add(2 * time.Hour), Exit: 0, WallSeconds: f(20), QueueSeconds: f(4), Units: []SuiteDuration{{"shard", "shard1", 10}}},
		{Repo: "ariadne", SHA: shaA, Check: "suite", StartedAt: t0, Exit: 1, WallSeconds: f(10), QueueSeconds: f(0), Failures: []string{"pkg.TestOne"}, Units: []SuiteDuration{{"shard", "shard1", 8}}},
		{Repo: "ariadne", SHA: shaB, Check: "suite", StartedAt: t0.Add(time.Hour), Exit: 1, MainBaseline: true, WallSeconds: f(30), Failures: []string{"pkg.TestBaseline"}, Units: []SuiteDuration{{"package", "pkg", 15}}},
		{Repo: "ariadne", SHA: shaB, Check: "suite", StartedAt: t0.Add(3 * time.Hour), Exit: 0, WallSeconds: f(40), QueueSeconds: f(8), Units: []SuiteDuration{{"shard", "shard1", 12}}},
		{Repo: "ariadne", SHA: shaA, Check: "suite", StartedAt: t0.Add(-48 * time.Hour), Exit: 1, WallSeconds: f(99), Failures: []string{"pkg.TestOld"}},
		{Repo: "pair", SHA: shaA, Check: "suite", StartedAt: t0.Add(time.Hour), Exit: 0, Legacy: true, Units: []SuiteDuration{{"package", "pair/pkg", 6}}},
	}
	got := aggregateSuiteAttempts(rows, t0.Add(-time.Hour), "ariadne")
	if len(got) != 1 {
		t.Fatalf("rows = %#v, want one ariadne row", got)
	}
	s := got[0]
	if s.Runs != 4 || s.FullyRecorded != 4 || s.WallCoverage != 4 || s.QueueCoverage != 3 || s.FlakeDenominator != 4 || s.ConfirmedFlakeAttempts != 1 {
		t.Fatalf("counts = %#v", s)
	}
	if s.Wall == nil || *s.Wall != (SuiteQuantiles{P50: 20, P90: 40, Samples: 4}) {
		t.Fatalf("wall = %#v", s.Wall)
	}
	if s.Queue == nil || *s.Queue != (SuiteQuantiles{P50: 4, P90: 8, Samples: 3}) {
		t.Fatalf("queue = %#v", s.Queue)
	}
	if len(s.Flakes) != 2 || s.Flakes[0].Kind != "suspected-main-baseline" || s.Flakes[1].Kind != "confirmed" {
		t.Fatalf("flakes = %#v", s.Flakes)
	}
	if len(s.Slow) != 2 || s.Slow[0] != (SuiteSlowUnit{Kind: "package", Name: "pkg", P90: 15, Samples: 1}) || s.Slow[1] != (SuiteSlowUnit{Kind: "shard", Name: "shard1", P90: 12, Samples: 3}) {
		t.Fatalf("slow = %#v", s.Slow)
	}
	// Input order is intentionally not chronological: classification uses time,
	// not the order returned by a directory walk.
	reordered := append([]SuiteAttempt(nil), rows...)
	for i, j := 0, len(reordered)-1; i < j; i, j = i+1, j-1 {
		reordered[i], reordered[j] = reordered[j], reordered[i]
	}
	if other := aggregateSuiteAttempts(reordered, t0.Add(-time.Hour), "ariadne"); !reflect.DeepEqual(other, got) {
		t.Fatalf("order-sensitive aggregation: %#v vs %#v", got, other)
	}
}

func TestSuiteStatsLegacyCoverageAndRepoSort(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	rows := []SuiteAttempt{
		{Repo: "pair", Check: "suite", StartedAt: now, Legacy: true, Exit: 1, Failures: []string{"pkg.TestLegacy"}},
		{Repo: "ariadne", Check: "suite", StartedAt: now, Exit: 0},
	}
	got := aggregateSuiteAttempts(rows, now.Add(-time.Hour), "")
	if len(got) != 2 || got[0].Repo != "ariadne" || got[1].Repo != "pair" {
		t.Fatalf("repo order = %#v", got)
	}
	if got[1].FullyRecorded != 0 || got[1].FlakeDenominator != 0 || got[1].ConfirmedFlakeAttempts != 0 || got[1].Wall != nil || got[1].Queue != nil {
		t.Fatalf("legacy attempt counted as complete evidence: %#v", got[1])
	}
}

func TestSuiteStatsFlakeNeedsPassAfterFailureCompleted(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	rows := []SuiteAttempt{
		{Repo: "ariadne", Check: "suite", SHA: sha, StartedAt: now, FinishedAt: now.Add(3 * time.Hour), Exit: 1, Failures: []string{"pkg.TestA"}},
		{Repo: "ariadne", Check: "suite", SHA: sha, StartedAt: now.Add(time.Hour), FinishedAt: now.Add(2 * time.Hour), Exit: 0},
	}
	got := aggregateSuiteAttempts(rows, now.Add(-time.Hour), "ariadne")
	if len(got) != 1 || got[0].ConfirmedFlakeAttempts != 0 {
		t.Fatalf("pass ended before failure but counted as retry: %+v", got)
	}
}
