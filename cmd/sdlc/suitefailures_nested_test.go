package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

const nestedOuter = "github.com/xianxu/pair/cmd/internal/couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees"

func nestedSuiteLog(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/nested-go-test-476.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSuiteQuotedNestedOutput(t *testing.T) {
	plain := nestedSuiteLog(t)
	for name, log := range map[string]string{
		"plain": plain,
		"CRLF":  strings.ReplaceAll(plain, "\n", "\r\n"),
		"sharded": "--- FAIL github.com/xianxu/pair/cmd/internal/couchcmd TestContinuationWriterPublishesExactCheckpointAcrossWorktrees\n" + plain +
			"--- FAIL github.com/xianxu/pair/cmd/internal/couchcmd\n",
	} {
		t.Run(name, func(t *testing.T) {
			got := parseSuiteLogFacts(log)
			if !reflect.DeepEqual(got.Failures, []string{nestedOuter}) ||
				!reflect.DeepEqual(got.Subtests, []string{nestedOuter + "/warm-reattached"}) || len(got.Unattributed) != 0 || !got.Failed {
				t.Fatalf("quoted helper changed outer attribution: %+v", got)
			}
			if !reflect.DeepEqual(got.Units, []SuiteDuration{{Kind: "package", Name: "github.com/xianxu/pair/cmd/internal/couchcmd", Seconds: 3.25}}) {
				t.Fatalf("quoted helper became a measured package: %+v", got.Units)
			}
		})
	}
}

func TestSuiteQuotedControlRecords(t *testing.T) {
	controls := []string{
		"FAIL example.com/helper [build failed]", "FAIL example.com/helper [setup failed]",
		"panic: test timed out after 10m0s", "--- shard3 exited 2 with no failing test",
		"--- 2 listed tests never ran", "--- FAIL example.com/helper TestHelper",
		`{"Action":"fail","Package":"example.com/helper","Test":"TestHelper","Elapsed":1}`,
		`{"Action":"fail","Package":"example.com/helper","Elapsed":1}`,
		"ok example.com/helper 1.2s", "shard3 exit 1 1.2s 1 failed", "wall 999s", "FAIL",
	}
	for _, record := range controls {
		for _, indent := range []string{"    ", "\t"} {
			log := "ok example.com/outer 0.5s\n" + indent + record + "\n"
			got := parseSuiteLogFacts(log)
			if got.Failed || len(got.Failures)+len(got.Subtests)+len(got.Unattributed) != 0 || len(got.Units) != 1 || got.WallSeconds != nil {
				t.Errorf("quoted %q became suite evidence: %+v", indent+record, got)
			}
		}
	}
	// Column-zero failures must remain blocking, even beside a registered outer.
	for _, record := range []string{"FAIL example.com/helper 0.2s", "FAIL example.com/helper [build failed]", "panic: test timed out after 10m0s"} {
		if got := parseSuiteLogFacts(nestedSuiteLog(t) + record + "\n"); len(got.Unattributed) == 0 {
			t.Errorf("genuine failure lost: %s: %+v", record, got)
		}
	}
}

func TestKnownFlakeGateQuotedNestedOutput(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	d.FlakeRegistry = gateRegistry(t, nestedOuter+"/warm-reattached")
	asked := scriptRerun(t, "printf -- '--- PASS: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees (0.01s)\\n    --- PASS: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached (0.01s)\\nok github.com/xianxu/pair/cmd/internal/couchcmd 0.02s\\n'")
	c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `printf '%s' "$1"; exit 1`, "fixture", nestedSuiteLog(t)}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakePassed || gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("outer registered leaf did not rerun: %+v %v\n%s", rec, err, stderr)
	}
	if !reflect.DeepEqual(*asked, []string{nestedOuter}) {
		t.Fatalf("rerun must select only the owning parent (including its siblings): %v", *asked)
	}
}

func TestFailingNamesLineVisits(t *testing.T) {
	for _, n := range []int{8, 16} {
		var b strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "--- FAIL: TestOuter%d (0.01s)\n    --- FAIL: TestOuter%d/a/b (0.01s)\n        --- FAIL: TestHelper (0.01s)\n        FAIL example.com/helper 0.1s\n", i, i)
		}
		b.WriteString("FAIL example.com/outer 1s\n")
		for _, subtests := range []bool{false, true} {
			visits := 0
			lines := func(yield func(string) bool) {
				for line := range strings.Lines(b.String()) {
					visits++
					if !yield(line) {
						return
					}
				}
			}
			got := failingNameLines(lines, subtests)
			if visits != 4*n+1 || len(got) != n {
				t.Fatalf("N=%d subtests=%v: %d visits, %d names", n, subtests, visits, len(got))
			}
		}
	}
}
