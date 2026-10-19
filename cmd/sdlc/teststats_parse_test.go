package main

import (
	"reflect"
	"strings"
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
		{Kind: "package", Name: "github.com/xianxu/pair/cmd/internal/wrapcmd", Seconds: 57.12},
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

func TestSuiteLogFactsGoTestJSON(t *testing.T) {
	log := `{"Time":"2026-10-14T12:00:00Z","Action":"output","Package":"example.com/a","Test":"TestBad","Output":"--- FAIL: TestBad (0.01s)\n"}
{"Time":"2026-10-14T12:00:00Z","Action":"fail","Package":"example.com/a","Test":"TestBad","Elapsed":0.01}
{"Time":"2026-10-14T12:00:00Z","Action":"fail","Package":"example.com/a","Test":"TestBad/sub","Elapsed":0.01}
{"Time":"2026-10-14T12:00:01Z","Action":"fail","Package":"example.com/a","Elapsed":1.25}
{"Time":"2026-10-14T12:00:02Z","Action":"pass","Package":"example.com/b","Elapsed":2.5}
`
	got := parseSuiteLogFacts(log)
	wantUnits := []SuiteDuration{
		{Kind: "package", Name: "example.com/a", Seconds: 1.25},
		{Kind: "package", Name: "example.com/b", Seconds: 2.5},
	}
	if !reflect.DeepEqual(got.Units, wantUnits) {
		t.Fatalf("units = %#v, want %#v", got.Units, wantUnits)
	}
	if !reflect.DeepEqual(got.Failures, []string{"example.com/a.TestBad"}) {
		t.Fatalf("failures = %q", got.Failures)
	}
	if !got.Failed {
		t.Fatal("package fail event must mark suite failed")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %q", got.Warnings)
	}
}

func TestSuiteLogFactsMalformedGoTestJSONWarnsWithoutInventingData(t *testing.T) {
	log := `{"Action":"pass","Package":"example.com/a","Elapsed":NaN}
{"Action":"pass","Package":"example.com/b","Elapsed":-2}
{"Action":"pass","Package":"example.com/c"}
{"Action":"fail","Package":"example.com/d","Elapsed":0.25}
`
	got := parseSuiteLogFacts(log)
	if !reflect.DeepEqual(got.Units, []SuiteDuration{{Kind: "package", Name: "example.com/d", Seconds: 0.25}}) {
		t.Fatalf("units = %#v", got.Units)
	}
	if len(got.Warnings) != 3 {
		t.Fatalf("warnings = %q, want malformed JSON and missing/invalid elapsed diagnostics", got.Warnings)
	}
	if !got.Failed {
		t.Fatal("valid fail event must mark suite failed")
	}
}

func TestSuiteLogFactsPlainFailureMarker(t *testing.T) {
	for _, log := range []string{"FAIL example.com/a 0.2s\n", "shard0 exit 2 12.0s 1 failed\n"} {
		if !parseSuiteLogFacts(log).Failed {
			t.Fatalf("failed suite not detected in %q", log)
		}
	}
}

func TestSuiteLogFactsJSONTestFailureWithoutPackageOutcome(t *testing.T) {
	got := parseSuiteLogFacts("{\"Action\":\"fail\",\"Package\":\"example.com/a\",\"Test\":\"TestOnly\"}\n")
	if !got.Failed || !reflect.DeepEqual(got.Failures, []string{"example.com/a.TestOnly"}) {
		t.Fatalf("facts = %#v, want failed test even without package outcome", got)
	}
}

func TestSuiteLogFactsOutcomeFormats(t *testing.T) {
	tests := []struct {
		name   string
		log    string
		failed bool
		unit   SuiteDuration
	}{
		{"plain pass", "ok example.com/pkg 1.2s\n", false, SuiteDuration{"package", "example.com/pkg", 1.2}},
		{"plain fail", "FAIL example.com/pkg 2.3s\n", true, SuiteDuration{"package", "example.com/pkg", 2.3}},
		{"JSON pass", "{\"Action\":\"pass\",\"Package\":\"example.com/pkg\",\"Elapsed\":3.4}\n", false, SuiteDuration{"package", "example.com/pkg", 3.4}},
		{"JSON fail", "{\"Action\":\"fail\",\"Package\":\"example.com/pkg\",\"Elapsed\":4.5}\n", true, SuiteDuration{"package", "example.com/pkg", 4.5}},
		{"shard pass", "shard0 exit 0 5.6s 0 failed\n", false, SuiteDuration{"shard", "shard0", 5.6}},
		{"shard fail", "shard0 exit 1 6.7s 1 failed\n", true, SuiteDuration{"shard", "shard0", 6.7}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseSuiteLogFacts(tc.log)
			if got.Failed != tc.failed || !reflect.DeepEqual(got.Units, []SuiteDuration{tc.unit}) {
				t.Fatalf("facts = %#v, want failed=%v unit=%#v", got, tc.failed, tc.unit)
			}
		})
	}
}

func TestSuiteLogFactsLineSizeBoundary(t *testing.T) {
	const lineLimit = 1 << 20
	t.Run("exact limit accepted", func(t *testing.T) {
		got := parseSuiteLogFacts(strings.Repeat("x", lineLimit) + "\nwall 1s\n")
		if got.WallSeconds == nil || *got.WallSeconds != 1 || len(got.Warnings) != 0 {
			t.Fatalf("facts = %#v", got)
		}
	})
	t.Run("over limit makes all facts unavailable", func(t *testing.T) {
		log := "wall 1s\n" + strings.Repeat("x", lineLimit+1) + "\nFAIL example.com/pkg 2s\n"
		got := parseSuiteLogFacts(log)
		if got.WallSeconds != nil || got.Failed || len(got.Failures) != 0 || len(got.Units) != 0 || len(got.Warnings) != 1 {
			t.Fatalf("facts = %#v; want unavailable measurements and one warning", got)
		}
	})
}

// A failure is attributed only to an exact test; everything else is a reason
// it is never flake-eligible (#394).
func TestUnattributedFailures(t *testing.T) {
	const pkg = "github.com/x/ariadne/cmd/sdlc"
	for _, tc := range []struct {
		name, log string
		want      []string
	}{
		{"a named test failure, test-shard form", "shard1    exit 1     3.0s  1 failed\n--- FAIL " + pkg + " TestA\n--- FAIL " + pkg + "\n", nil},
		{"a named test failure, go test form", "--- FAIL: TestA (0.00s)\nFAIL\nFAIL\t" + pkg + "\t0.123s\n", nil},
		{"a shard died without a test", "--- shard2 exited 2 with no failing test; last output:\n", []string{"shard2 exited without a failing test"}},
		{"tests never ran", "--- 3 listed cmd/sdlc test(s) never ran: TestA, TestB\n", []string{"listed tests never ran"}},
		{"a build failure", "FAIL\t" + pkg + " [build failed]\n", []string{"package " + pkg + ": build failed"}},
		{"a timeout", "panic: test timed out after 10m0s\n", []string{"a test timed out"}},
		{"a package failed alone", "--- FAIL " + pkg + "\n", []string{"package " + pkg + " failed without a failing test"}},
		{"a go test package failed alone", "FAIL\t" + pkg + "\t0.5s\n", []string{"package " + pkg + " failed without a failing test"}},
		{"a JSON package failure alone", `{"Action":"fail","Package":"` + pkg + `","Elapsed":1}` + "\n", []string{"package " + pkg + " failed without a failing test"}},
		{"a test name without a package", "--- FAIL: TestA (0.00s)\n", []string{"failure TestA names no package"}},
	} {
		got := parseSuiteLogFacts(tc.log).Unattributed
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}
