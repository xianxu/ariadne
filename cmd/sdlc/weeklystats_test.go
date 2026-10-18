package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestNewWeeklyWindowHalfOpenUTC(t *testing.T) {
	end := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	w, err := newWeeklyWindow(end, end.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if !w.Start.Equal(end.Add(-7*24*time.Hour)) || !w.End.Equal(end) {
		t.Fatalf("window = %+v", w)
	}
	for _, tc := range []struct {
		at   time.Time
		want bool
	}{
		{w.Start.Add(-time.Nanosecond), false},
		{w.Start, true},
		{w.End.Add(-time.Nanosecond), true},
		{w.End, false},
	} {
		if got := w.Contains(tc.at); got != tc.want {
			t.Errorf("Contains(%s) = %t, want %t", tc.at, got, tc.want)
		}
	}
	if _, err := newWeeklyWindow(end.Add(time.Nanosecond), end); err == nil {
		t.Error("future end accepted")
	}
	if _, err := newWeeklyWindow(end.In(time.FixedZone("offset", 3600)), end); err == nil {
		t.Error("non-UTC end accepted")
	}
}

func TestAggregateWeeklyRepoCoverageAndStableOrder(t *testing.T) {
	end := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	w, _ := newWeeklyWindow(end, end)
	wall10, wall30, queue2 := 10.0, 30.0, 2.0
	rows := []SuiteAttempt{
		{ID: "after", Repo: "ariadne", Check: suiteCheckName, FinishedAt: w.End, WallSeconds: &wall30},
		{ID: "b", Repo: "ariadne", Check: suiteCheckName, SHA: "head", StartedAt: w.Start.Add(time.Hour), FinishedAt: w.Start.Add(2 * time.Hour), Exit: 1, Failures: []string{"TestB", "TestA"}, WallSeconds: &wall30, QueueSeconds: &queue2},
		{ID: "other", Repo: "pair", Check: suiteCheckName, FinishedAt: w.Start.Add(time.Hour), WallSeconds: &wall30},
		{ID: "a", Repo: "ariadne", Check: suiteCheckName, SHA: "head", StartedAt: w.Start.Add(3 * time.Hour), FinishedAt: w.Start.Add(4 * time.Hour), WallSeconds: &wall10, Exit: 0},
		{ID: "legacy", Repo: "ariadne", Check: suiteCheckName, FinishedAt: w.Start, Legacy: true},
		{ID: "pre", Repo: "ariadne", Check: suiteCheckName, FinishedAt: w.Start.Add(-time.Nanosecond), WallSeconds: &wall30},
	}
	got := aggregateWeeklyRepo("ariadne", rows, w)
	if got.Suite.Runs != 3 || got.Suite.FullyRecorded != 2 || got.Suite.Wall.Samples != 2 || got.Suite.Wall.P90 != 30 || got.Suite.Queue.Samples != 1 {
		t.Fatalf("suite coverage = %+v", got.Suite)
	}
	if len(got.FailedTests) != 2 || got.FailedTests[0].Name != "TestA" || got.FailedTests[0].Kind != "confirmed" || got.FailedTests[1].Name != "TestB" {
		t.Fatalf("failed tests = %+v", got.FailedTests)
	}
	first, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	second, err := json.Marshal(aggregateWeeklyRepo("ariadne", rows, w))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("unstable JSON: %s != %s", first, second)
	}
}

func TestAggregateWeeklyRepoUnknownFailure(t *testing.T) {
	end := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	w, _ := newWeeklyWindow(end, end)
	rows := []SuiteAttempt{{ID: "failed", Repo: "ariadne", Check: suiteCheckName, SHA: "head", StartedAt: w.Start, FinishedAt: w.Start.Add(time.Second), Exit: 1, Failures: []string{"TestUnknown"}}}
	got := aggregateWeeklyRepo("ariadne", rows, w)
	if len(got.FailedTests) != 1 || got.FailedTests[0].Kind != "unknown" {
		t.Fatalf("unclassified failure disappeared: %+v", got.FailedTests)
	}
}

func TestAggregateWeeklyRepoNamesSlowPackagesWithoutConfusingShards(t *testing.T) {
	end := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	w, _ := newWeeklyWindow(end, end)
	rows := []SuiteAttempt{{ID: "one", Repo: "ariadne", Check: suiteCheckName, FinishedAt: w.Start.Add(time.Hour), Units: []SuiteDuration{
		{Kind: "shard", Name: "shard0", Seconds: 900},
		{Kind: "package", Name: "slow/pkg", Seconds: 120},
		{Kind: "package", Name: "fast/pkg", Seconds: 2},
	}}}
	got := aggregateWeeklyRepo("ariadne", rows, w)
	if got.PackageDurationCoverage != 1 || len(got.SlowPackages) != 2 || got.SlowPackages[0].Name != "slow/pkg" || got.SlowPackages[0].Kind != "package" {
		t.Fatalf("slow packages = %+v (coverage %d)", got.SlowPackages, got.PackageDurationCoverage)
	}
}

func TestWeeklySnapshotVersionedNullableRatio(t *testing.T) {
	s := WeeklySnapshot{SchemaVersion: 1, Repos: []WeeklyRepoStats{{Repo: "ariadne"}}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["schema_version"] != float64(1) {
		t.Fatalf("missing schema version: %s", b)
	}
	repos := decoded["repos"].([]any)
	if repos[0].(map[string]any)["attempts_per_landed_issue"] != nil {
		t.Fatalf("unknown ratio encoded as number: %s", b)
	}
}
