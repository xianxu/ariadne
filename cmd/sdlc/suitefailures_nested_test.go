package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode"

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

func TestSuiteQuotedOverlappingNames(t *testing.T) {
	const top = "--- FAIL: TestOuter (0.01s)\n"
	for _, realChild := range []bool{false, true} {
		log := top
		wantSubs := []string(nil)
		if realChild {
			log += "    --- FAIL: TestOuter/real (0.01s)\n"
			wantSubs = []string{"example.com/outer.TestOuter/real"}
		}
		log += "        outer_test.go:9: nested command failed:\n" +
			"            --- FAIL: TestOuter/helper (0.01s)\n" +
			"                --- FAIL: TestOuter/helper/deep (0.01s)\n" +
			"            FAIL example.com/helper 0.1s\nFAIL example.com/outer 0.2s\n"
		for name, text := range map[string]string{
			"plain":   log,
			"CRLF":    strings.ReplaceAll(log, "\n", "\r\n"),
			"sharded": "--- FAIL example.com/outer TestOuter\n" + log + "--- FAIL example.com/outer\n",
		} {
			got := parseSuiteLogFacts(text)
			if !reflect.DeepEqual(got.Failures, []string{"example.com/outer.TestOuter"}) || !reflect.DeepEqual(got.Subtests, wantSubs) || len(got.Unattributed) != 0 {
				t.Errorf("%s real child=%v: quoted shared-name helper became evidence: %+v", name, realChild, got)
			}
		}
	}
}

func TestSuiteIndentedFailureAncestry(t *testing.T) {
	log := "--- FAIL: TestOuter (0.01s)\n" +
		"    --- FAIL: TestOuter/literal/a/b (0.01s)\n" +
		"        --- FAIL: TestOuter/literal/a/b/grand (0.01s)\n" +
		"            --- FAIL: TestOuter/helper (0.01s)\n" +
		"    --- FAIL: TestOuter/second (0.01s)\n" +
		"            --- FAIL: TestOuter/second/skipped (0.01s)\n" +
		"        --- FAIL: TestOuter/second/deep (0.01s)\n" +
		"      --- FAIL: TestOuter/misaligned (0.01s)\n" +
		"\t--- FAIL: TestOuter/tabbed (0.01s)\n" +
		"FAIL example.com/outer 0.2s\n"
	want := []string{"example.com/outer.TestOuter/literal/a/b", "example.com/outer.TestOuter/literal/a/b/grand",
		"example.com/outer.TestOuter/second", "example.com/outer.TestOuter/second/deep"}
	if got := parseSuiteLogFacts(log); !reflect.DeepEqual(got.Subtests, want) {
		t.Fatalf("ancestry must use observed levels and complete parent names: %+v", got)
	}
}

func TestKnownFlakeGateCannotRegisterAQuotedLeaf(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	d.FlakeRegistry = gateRegistry(t, "example.com/outer.TestOuter/helper")
	asked := scriptRerun(t, "printf -- '--- PASS: TestOuter (0.01s)\\n    --- PASS: TestOuter/helper (0.01s)\\nok example.com/outer 0.1s\\n'")
	log := "--- FAIL: TestOuter (0.01s)\n    x_test.go:9: nested command failed:\n" +
		"            --- FAIL: TestOuter/helper (0.01s)\nFAIL example.com/outer 0.2s\n"
	c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `printf '%s' "$1"; exit 1`, "fixture", log}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || len(*asked) != 0 || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
		t.Fatalf("quoted leaf registration must not explain real outer failure: %+v %v, reran %v", rec, err, *asked)
	}
}

func TestSuiteWholeWhitespacePrefix(t *testing.T) {
	// Enumerate Go's entire whitespace category, rather than a blacklist of
	// the two characters that exposed BR-1. LF frames records in the parser,
	// so its invalid single-line prefix is tested directly below.
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if !unicode.IsSpace(r) || r == ' ' {
			continue
		}
		for _, prefix := range []string{string(r) + "    ", "    " + string(r), "  " + string(r) + "  "} {
			if got := suiteIndentLevel(prefix + "--- FAIL: TestOuter/helper (0.01s)"); got != -1 {
				t.Errorf("U+%04X prefix %q: level=%d, want invalid", r, prefix, got)
			}
			if r == '\n' {
				continue // LF is a line boundary, not an in-line indentation rune.
			}
			for mode := 0; mode < 3; mode++ {
				log := "--- FAIL: TestOuter (0.01s)\n" + prefix + "--- FAIL: TestOuter/helper (0.01s)\nFAIL example.com/outer 0.2s\n"
				if mode == 1 {
					log = strings.ReplaceAll(log, "\n", "\r\n")
				} else if mode == 2 {
					log = "--- FAIL example.com/outer TestOuter\n" + log + "--- FAIL example.com/outer\n"
				}
				got := parseSuiteLogFacts(log)
				if !reflect.DeepEqual(got.Failures, []string{"example.com/outer.TestOuter"}) || len(got.Subtests)+len(got.Unattributed) != 0 {
					t.Errorf("U+%04X prefix %q mode=%d: mixed indentation became a leaf: %+v", r, prefix, mode, got)
				}
			}
		}
	}
	for _, tc := range []struct {
		prefix string
		level  int
	}{{"", 0}, {"    ", 1}, {"        ", 2}, {" ", -1}, {"     ", -1}, {"      ", -1}} {
		if got := suiteIndentLevel(tc.prefix + "--- FAIL: TestOuter/real (0.01s)"); got != tc.level {
			t.Errorf("ASCII prefix %q: level=%d, want %d", tc.prefix, got, tc.level)
		}
	}
}

func TestKnownFlakeGateMixedWhitespaceQuotation(t *testing.T) {
	for _, prefix := range []string{"    \t", "    \u00a0", "  \u2003  "} {
		t.Run(fmt.Sprintf("%x", prefix), func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "acquired")
			d.FlakeRegistry = gateRegistry(t, "example.com/outer.TestOuter/helper")
			asked := scriptRerun(t, "printf -- '--- PASS: TestOuter (0.01s)\\n    --- PASS: TestOuter/helper (0.01s)\\nok example.com/outer 0.1s\\n'")
			log := "--- FAIL: TestOuter (0.01s)\n" + prefix + "--- FAIL: TestOuter/helper (0.01s)\nFAIL example.com/outer 0.2s\n"
			c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `printf '%s' "$1"; exit 1`, "fixture", log}, TimeoutMinutes: 30}
			rec, err := runLocalCheck(context.Background(), d, c, sha)
			if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || len(*asked) != 0 || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
				t.Fatalf("mixed-whitespace leaf explained outer failure: %+v %v, reran %v", rec, err, *asked)
			}
		})
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
