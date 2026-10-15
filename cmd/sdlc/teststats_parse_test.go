package main

import (
	"reflect"
	"testing"
)

func TestSuiteLogFactsAriadne(t *testing.T) {
	log := `packages exit 0     81.2s  0 failed
shard0    exit 1    102.4s  1 failed
shard1    exit 0     98.0s  0 failed
shard2 exit nope 23.0s 0 failed
--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestFailure
--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestFailure
wall 123.5s  (2 shards over 10 cmd/sdlc tests, 20 other packages)
`
	got := parseSuiteLogFacts(log)
	wantUnits := []SuiteDuration{
		{Kind: "package-group", Name: "packages", Seconds: 81.2},
		{Kind: "shard", Name: "shard0", Seconds: 102.4},
		{Kind: "shard", Name: "shard1", Seconds: 98},
	}
	if !reflect.DeepEqual(got.Units, wantUnits) {
		t.Fatalf("units = %#v, want %#v", got.Units, wantUnits)
	}
	if !reflect.DeepEqual(got.Failures, failingTests(log)) || len(got.Failures) != 1 {
		t.Fatalf("failures = %q, baseline extractor = %q", got.Failures, failingTests(log))
	}
	if got.WallSeconds == nil || *got.WallSeconds != 123.5 {
		t.Fatalf("wall = %v, want 123.5", got.WallSeconds)
	}
}

func TestSuiteLogFactsPairAndMalformedLines(t *testing.T) {
	log := "ok  \tgithub.com/xianxu/pair/cmd/internal/wrapcmd\t57.120s\n" +
		"ok  \tgithub.com/xianxu/pair/cmd/internal/titlepoller\t0.176s\n" +
		"ok package-not-a-module 3.2s\n" +
		"ok  \tgithub.com/xianxu/pair/bad\tNaNs\n" +
		"? github.com/xianxu/pair/no-tests [no test files]\n" +
		"--- FAIL: TestPair (0.00s)\nFAIL github.com/xianxu/pair/cmd/internal/wrapcmd 57.120s\n"
	got := parseSuiteLogFacts(log)
	want := []SuiteDuration{
		{Kind: "package", Name: "github.com/xianxu/pair/cmd/internal/wrapcmd", Seconds: 57.12},
		{Kind: "package", Name: "github.com/xianxu/pair/cmd/internal/titlepoller", Seconds: 0.176},
	}
	if !reflect.DeepEqual(got.Units, want) {
		t.Fatalf("units = %#v, want %#v", got.Units, want)
	}
	if !reflect.DeepEqual(got.Failures, []string{"github.com/xianxu/pair/cmd/internal/wrapcmd.TestPair"}) {
		t.Fatalf("failures = %q", got.Failures)
	}
	if got.WallSeconds != nil {
		t.Fatalf("pair's historical log has no wall timing: %v", *got.WallSeconds)
	}
}
