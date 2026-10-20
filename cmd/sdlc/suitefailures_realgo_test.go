package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

const realGoFixturePkg = "example.com/ariadne476output"

var realGoFixtureTops = []string{realGoFixturePkg + ".TestNestedReturnAndSibling", realGoFixturePkg + ".TestParentAfter", realGoFixturePkg + ".TestParentBefore"}

func realGoOutput(t *testing.T, mode string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/476-parent-output-" + mode + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Captured by running the committed .go.txt source with Go 1.27.1, both with
// and without -v. No Go compiler is launched by these fixture tests.
func TestSuiteRealGoOutputCoarsens(t *testing.T) {
	for _, fixture := range []string{"plain", "verbose"} {
		for _, mode := range []string{"raw", "CRLF", "sharded"} {
			log := realGoOutput(t, fixture)
			if mode == "CRLF" {
				log = strings.ReplaceAll(log, "\n", "\r\n")
			} else if mode == "sharded" {
				log = "--- FAIL " + realGoFixturePkg + " TestParentAfter\n" + log
			}
			got := parseSuiteLogFacts(log)
			if !reflect.DeepEqual(got.Failures, realGoFixtureTops) || len(got.Subtests)+len(got.Unattributed) != 0 || !got.Failed {
				t.Errorf("%s/%s: real parent logging must coarsen, never invent a leaf: %+v", fixture, mode, got)
			}
			if !strings.Contains(strings.Join(got.Warnings, " "), "coars") {
				t.Errorf("%s/%s: lost coarsening note: %+v", fixture, mode, got)
			}
			store := t.TempDir()
			if err := os.MkdirAll(filepath.Join(store, "failed"), 0o700); err != nil {
				t.Fatal(err)
			}
			a := SuiteAttempt{ID: "replay", LogPath: filepath.Join("failed", "replay.log")}
			if err := os.WriteFile(filepath.Join(store, a.LogPath), []byte(log), 0o600); err != nil {
				t.Fatal(err)
			}
			if subs, err := storeSubtests(store)(a); err != nil || len(subs) != 0 {
				t.Fatalf("historical replay invented a precise leaf: %v %v", subs, err)
			}
		}
	}
}

func TestSuiteStructuredLeafRemainsProven(t *testing.T) {
	for _, log := range []string{
		"--- FAIL " + realGoFixturePkg + " TestParentAfter/child\n" + realGoOutput(t, "plain"),
		`{"Action":"fail","Package":"example.com/ariadne476output","Test":"TestParentAfter/child"}` + "\n" +
			`{"Action":"fail","Package":"example.com/ariadne476output","Test":"TestParentAfter"}` + "\n" +
			`{"Action":"fail","Package":"example.com/ariadne476output","Elapsed":0.01}` + "\n",
	} {
		got := parseSuiteLogFacts(log)
		if !reflect.DeepEqual(got.Subtests, []string{realGoFixturePkg + ".TestParentAfter/child"}) || len(got.Unattributed) != 0 {
			t.Fatalf("explicit leaf lost or raw helper added: %+v", got)
		}
	}
}

func TestKnownFlakeGateRealGoProvenGranularity(t *testing.T) {
	for _, leafOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "proven parents", true: "subtests only"}[leafOnly], func(t *testing.T) {
			d, sha, gh, stderr := localRunRepo(t, "acquired")
			registered := append([]string(nil), realGoFixtureTops...)
			if leafOnly {
				for i := range registered {
					registered[i] += "/child"
				}
			}
			d.FlakeRegistry = gateRegistry(t, registered...)
			asked := scriptRerun(t, "printf -- '--- PASS: TestNestedReturnAndSibling (0.01s)\\n--- PASS: TestParentAfter (0.01s)\\n--- PASS: TestParentBefore (0.01s)\\nok example.com/ariadne476output 0.03s\\n'")
			c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `printf '%s' "$1"; exit 1`, "fixture", realGoOutput(t, "plain")}, TimeoutMinutes: 30}
			rec, err := runLocalCheck(context.Background(), d, c, sha)
			if err != nil || rec.passed() == leafOnly || (gh.statuses[sha]["sdlc-local/suite"] == "success") == leafOnly {
				t.Fatalf("real Go attribution at wrong granularity: %+v %v\n%s", rec, err, stderr)
			}
			if leafOnly && (rec.Outcome != flakes.Unknown || len(*asked) != 0) ||
				!leafOnly && (rec.Outcome != flakes.KnownFlakePassed || !reflect.DeepEqual(*asked, realGoFixtureTops)) {
				t.Fatalf("wrong retry: %+v asked=%v", rec, *asked)
			}
			if !strings.Contains(stderr.String(), "coars") || !strings.Contains(stderr.String(), "subtest-only") {
				t.Fatalf("missing precision note:\n%s", stderr)
			}
		})
	}
}
