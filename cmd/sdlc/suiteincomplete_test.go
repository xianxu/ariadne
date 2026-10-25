package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestClassifyIncomplete(t *testing.T) {
	const sdlc = "github.com/x/y/cmd/sdlc"
	lines := func(l ...string) string { return strings.Join(l, "\n") + "\n" }
	alarmed := func(shard string) []string {
		return []string{shard + "    exit 1   1800.4s  1 failed", "--- " + shard + " hit its 30m0s test alarm in " + sdlc + "; running: TestSlow"}
	}
	var (
		packagesOK  = "packages  exit 0    269.9s  0 failed"
		neverRan    = "--- 760 listed cmd/sdlc test(s) never ran: TestA, TestB"
		namedInSdlc = "--- FAIL " + sdlc + " TestBarrier"
	)
	cases := []struct {
		name string
		log  string
		want []string // the retry's packages; nil means not incomplete
	}{
		{"every failed shard alarmed", lines(append(append([]string{packagesOK}, alarmed("shard0")...), append(append(alarmed("shard1"), namedInSdlc), neverRan)...)...), []string{sdlc}},
		{"a shard failed without the alarm", lines(append(append([]string{packagesOK}, alarmed("shard0")...), "shard1    exit 1   300.0s  1 failed", namedInSdlc)...), nil},
		{"no alarm at all", lines(packagesOK, "shard0    exit 1   300.0s  1 failed", namedInSdlc), nil},
		{"a build failure", lines(append(append([]string{"packages  exit 1    10.0s  0 failed", "FAIL\tgithub.com/x/y/pkg/z [build failed]"}, alarmed("shard0")...), neverRan)...), nil},
		{"an alarmed package in the packages job", lines(
			"packages  exit 1    300.0s  1 failed",
			"--- packages hit its 5m0s test alarm in github.com/x/y/pkg/slow; running: TestHang",
			"--- FAIL github.com/x/y/pkg/other TestReal",
			"shard0    exit 0   1100.0s  0 failed"), []string{"github.com/x/y/pkg/other", "github.com/x/y/pkg/slow"}},
		{"a passing run", lines(packagesOK, "shard0    exit 0   1100.0s  0 failed"), nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, ok := classifyIncomplete(parseSuiteLogFacts(c.log))
			if ok != (c.want != nil) || ok && !reflect.DeepEqual(run.Packages, c.want) {
				t.Fatalf("incomplete=%v packages=%v, want %v\nfacts: %+v", ok, run.Packages, c.want, parseSuiteLogFacts(c.log))
			}
		})
	}
}

// suiteTimedOut is a sharded suite whose one shard the Go test alarm stopped,
// with a named failure that finished before it.
func suiteTimedOut() localCheck {
	out := strings.Join([]string{
		"packages  exit 0    10.0s  0 failed",
		"shard0    exit 1   1800.4s  1 failed",
		"--- shard0 hit its 30m0s test alarm in o/r/pkg; running: TestSlow",
		"--- FAIL o/r/pkg TestFlaky",
		"--- FAIL o/r/pkg",
		"--- 3 listed cmd/sdlc test(s) never ran: TestA, TestB, TestC",
	}, "\n")
	return localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "printf '%s\\n' '" + out + "'; exit 1"}, TimeoutMinutes: 30}
}

func TestRunLocalCheckRecordsIncompleteTimeout(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	d.FlakeRegistry = gateRegistry(t, gateKnown) // registered, yet an incomplete run never reruns it
	asked := scriptRerun(t, gatePassJSON)
	d.loadAverage = func() ([3]float64, error) { return [3]float64{31.5, 46.4, 52}, nil }
	rec, err := runLocalCheck(context.Background(), d, suiteTimedOut(), sha)
	var incomplete *suiteIncompleteError
	if !errors.As(err, &incomplete) || rec.Outcome != flakes.IncompleteTimeout {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	for _, want := range []string{"incomplete (timeout)", "shard0 in o/r/pkg", "host load 31.5 46.4 52.0", "no main baseline was run"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
	if len(*asked) != 0 {
		t.Errorf("an incomplete run reran known flakes: %v", *asked)
	}
	if got := gh.statuses[sha]["sdlc-local/suite"]; got != "" {
		t.Errorf("an incomplete run posted %q", got)
	}
	rows := gateAttempts(t, d)
	if len(rows) != 1 || rows[0].Version != suiteAttemptIncompleteVersion || rows[0].Outcome != string(flakes.IncompleteTimeout) {
		t.Fatalf("attempts %+v", rows)
	}
	if !reflect.DeepEqual(rows[0].LoadAverage, []float64{31.5, 46.4, 52}) {
		t.Fatalf("load %v", rows[0].LoadAverage)
	}
	store := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	// A reader of version 4 at most (every binary before #630) refuses the
	// record's version, then keeps it as an unsupported proof of this head
	// alone; load_average is an unknown key it tolerates, so the rest of the
	// store stays readable.
	raw, err := os.ReadFile(filepath.Join(store, "attempts", rows[0].ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if proof, err := decodeUnsupportedProofAt(raw, rows[0].ID+".json", 4); err != nil || proof.SHA != sha {
		t.Fatalf("an older reader would not scope this record to its head: %+v %v", proof, err)
	}
	// load_average belongs to an incomplete attempt alone.
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["version"], fields["outcome"] = 2, string(flakes.Unknown)
	stray, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeSuiteAttempt(stray); err == nil {
		t.Fatal("a version 2 attempt carrying load_average decoded")
	}
}

// An incomplete run is never an unexplained failure, but a failure that
// finished before the alarm stays listed until it is registered (#630).
func TestFlakeListIncompleteRunFailures(t *testing.T) {
	w := &flakeWorld{cards: []flakes.Card{{ID: "000448", Status: "working", Tests: []string{flakeT1}, Owner: "ariadne:3", Created: "2026-10-16"}}}
	barrier := "github.com/xianxu/ariadne/pkg/z.TestBarrier"
	incomplete := attemptAt("i1", "s3", 4, 1, flakeT1, barrier)
	incomplete.Version, incomplete.Outcome = suiteAttemptIncompleteVersion, string(flakes.IncompleteTimeout)
	w.attempts = []SuiteAttempt{incomplete}
	var out bytes.Buffer
	if err := runFlakeList(context.Background(), &out, w.deps()); err != nil {
		t.Fatal(err)
	}
	section := out.String()[strings.Index(out.String(), "incomplete-run failures"):]
	if !strings.Contains(section, barrier+"  attempt i1 on s3") || strings.Contains(section, flakeT1+"  attempt i1") {
		t.Errorf("incomplete-run failures (registered ones omitted):\n%s", out.String())
	}
	if strings.Contains(out.String(), "attempt i1 on s3, 2026-10-18T10:04:00Z: ") {
		t.Errorf("an incomplete run listed as unexplained:\n%s", out.String())
	}
}

// An incomplete run reaches the merge gate and the train as execution
// evidence: no main baseline, no split, no blame; and a later pass on the
// same head is not refused because of it (#630).
func TestSuiteIncompleteIsNotATestResult(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	d.loadAverage = func() ([3]float64, error) { return [3]float64{40, 45, 50}, nil }
	_, err := runLocalCheck(context.Background(), d, suiteTimedOut(), sha)
	var incomplete *suiteIncompleteError
	if !errors.As(err, &incomplete) {
		t.Fatalf("%v\n%s", err, stderr)
	}
	gateErr := fmt.Errorf("CI gate: %w", err) // how waitCIGate returns a local run's error; the very same error back means no baseline ran
	if got := diagnoseSuiteFailure(context.Background(), gateErr, sha, d); got != gateErr {
		t.Fatalf("the merge gate diagnosed an incomplete run: %v", got)
	}
	c := &conductor{ctx: context.Background(), heads: []string{sha}, base: sha, suite: suiteTimedOut(), local: d}
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return localRunRecord{}, err }
	if ev, runErr := c.runSuite(1, landingline.ExecInitial); runErr == nil || ev != nil || !strings.Contains(runErr.Error(), "incomplete (timeout)") || c.baseline != nil {
		t.Fatalf("the train split or blamed an incomplete run: %+v %v", ev, runErr)
	}
	pass := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo ok"}, TimeoutMinutes: 30}
	if rec, err := runLocalCheck(context.Background(), d, pass, sha); err != nil || !rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("a pass after an incomplete run was refused: %+v %v %v\n%s", rec, err, gh.statuses[sha], stderr)
	}
}
