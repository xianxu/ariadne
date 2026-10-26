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
	"slices"
	"strings"
	"testing"
	"time"

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
		// The build failure sits in an alarmed process, so only the unattributed-reason rule can refuse it.
		{"a build failure", lines(
			"packages  exit 1    300.0s  0 failed",
			"--- packages hit its 5m0s test alarm in github.com/x/y/pkg/slow; running: TestHang",
			"FAIL\tgithub.com/x/y/pkg/z [build failed]"), nil},
		{"an alarmed package in the packages job", lines(
			"packages  exit 1    300.0s  1 failed",
			"--- packages hit its 5m0s test alarm in github.com/x/y/pkg/slow; running: TestHang",
			"--- FAIL github.com/x/y/pkg/other TestReal",
			"shard0    exit 0   1100.0s  0 failed"), []string{"github.com/x/y/pkg/other", "github.com/x/y/pkg/slow"}},
		{"a passing run", lines(packagesOK, "shard0    exit 0   1100.0s  0 failed"), nil},
		// Each reason the alarm explains is accepted only for an alarmed process or package (#630 close review).
		{"an alarmed package failed without a test", lines(
			"packages  exit 1    300.0s  0 failed",
			"--- packages hit its 5m0s test alarm in github.com/x/y/pkg/slow; running: TestHang",
			"--- FAIL github.com/x/y/pkg/slow"), []string{"github.com/x/y/pkg/slow"}},
		{"an unalarmed package failed without a test", lines(
			"packages  exit 1    300.0s  0 failed",
			"--- packages hit its 5m0s test alarm in github.com/x/y/pkg/slow; running: TestHang",
			"--- FAIL github.com/x/y/pkg/other"), nil},
		{"an alarmed shard exited without a test", lines(append(append([]string{packagesOK}, alarmed("shard0")...), "--- shard0 exited 1 with no failing test", neverRan)...), []string{sdlc}},
		{"an unalarmed shard exited without a test", lines(append(append([]string{packagesOK}, alarmed("shard0")...), "--- shard1 exited 1 with no failing test", neverRan)...), nil},
		{"the alarm's panic line survives", lines(append(append([]string{packagesOK}, alarmed("shard0")...), "panic: test timed out after 30m0s", neverRan)...), []string{sdlc}},
		// Well-formed typed evidence leaves no unattributed reason, so only the Typed.Seen guard refuses it.
		{"a typed runner's evidence", lines(append(append([]string{packagesOK,
			`{"sdlc_unit":1,"invocation":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","unit":"script:a.sh","event":"start","complete":false,"errors":0}`,
			`{"sdlc_unit":1,"invocation":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","unit":"script:a.sh","event":"finish","exit":0,"complete":true,"errors":0}`},
			alarmed("shard0")...), neverRan)...), nil},
		{"a parse warning", lines(append(append([]string{packagesOK, `{"Action":"fail","Package":`}, alarmed("shard0")...), neverRan)...), nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, ok := classifyIncomplete(parseSuiteLogFacts(c.log))
			if ok != (c.want != nil) || ok && !reflect.DeepEqual(run.Packages, c.want) {
				t.Fatalf("incomplete=%v packages=%v, want %v\nfacts: %+v", ok, run.Packages, c.want, parseSuiteLogFacts(c.log))
			}
		})
	}
	run, _ := classifyIncomplete(parseSuiteLogFacts(cases[0].log))
	if want := "the Go test alarm stopped shard0 in " + sdlc + ", shard1 in " + sdlc + " (running TestSlow, TestSlow); unfinished or failed: " + sdlc; run.String() != want {
		t.Fatalf("report %q, want %q", run.String(), want)
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

// incompleteRepo is localRunRepo for a run that may retry (#630): the resume
// path reads main's policy through this seam, as the resume tests do.
func incompleteRepo(t *testing.T) (localRunDeps, string, *ciFakeGH, *bytes.Buffer) {
	t.Helper()
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyRerun, nil }
	return d, sha, gh, stderr
}

// A run the alarm cuts short twice: recorded with the load, retried once on
// its unfinished and failed packages, then reported, never posted (#630).
func TestRunLocalCheckRecordsIncompleteTimeout(t *testing.T) {
	d, sha, gh, stderr := incompleteRepo(t)
	d.FlakeRegistry = gateRegistry(t, gateKnown) // registered, yet an incomplete run never reruns it
	asked := scriptRerun(t, gatePassJSON)
	d.loadAverage = func() ([3]float64, error) { return [3]float64{31.5, 46.4, 52}, nil }
	rec, err := runLocalCheck(context.Background(), d, suiteTimedOut(), sha)
	var incomplete *suiteIncompleteError
	if !errors.As(err, &incomplete) || rec.Outcome != flakes.IncompleteTimeout {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	for _, want := range []string{"incomplete (timeout) again on its one retry", "shard0 in o/r/pkg", "host load 31.5 46.4 52.0", "no main baseline was run"} {
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
	var parent, child SuiteAttempt
	for _, a := range rows {
		if a.RerunOf == "" {
			parent = a
		} else {
			child = a
		}
	}
	if len(rows) != 2 || parent.Version != suiteAttemptIncompleteVersion || parent.Outcome != string(flakes.IncompleteTimeout) || child.RerunOf != parent.ID || child.Check != suiteRerunCheck {
		t.Fatalf("want the incomplete attempt and its one retry: %+v", rows)
	}
	if !reflect.DeepEqual(parent.LoadAverage, []float64{31.5, 46.4, 52}) {
		t.Fatalf("load %v", parent.LoadAverage)
	}
	store := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	// A reader of version 4 at most (every binary before #630) refuses the
	// record's version, then keeps it as an unsupported proof of this head
	// alone; load_average is an unknown key it tolerates, so the rest of the
	// store stays readable.
	raw, err := os.ReadFile(filepath.Join(store, "attempts", parent.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if proof, err := decodeUnsupportedProofAt(raw, parent.ID+".json", 4); err != nil || proof.SHA != sha {
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
	// Version 5 carries only what version 2 does: no typed runner identity, no known flakes.
	for name, set := range map[string]func(){
		"a typed runner's case": func() { fields["failures"] = []string{"lua:spec/a_spec.lua#case"} },
		"known flakes":          func() { fields["known_flakes"] = []map[string]string{{"test": "o/r/pkg.TestFlaky", "issue": "000448"}} },
	} {
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		set()
		bad, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeSuiteAttempt(bad); err == nil {
			t.Errorf("a version 5 attempt carrying %s decoded", name)
		}
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
	d, sha, gh, stderr := incompleteRepo(t)
	d.loadAverage = func() ([3]float64, error) { return [3]float64{40, 45, 50}, nil }
	_, err := runLocalCheck(context.Background(), d, suiteTimedOut(), sha)
	var incomplete *suiteIncompleteError
	if !errors.As(err, &incomplete) {
		t.Fatalf("%v\n%s", err, stderr)
	}
	if _, reused, err := republishSavedPass(context.Background(), d, suiteTimedOut(), sha); reused || gh.statuses[sha]["sdlc-local/suite"] != "" {
		t.Fatalf("a saved incomplete attempt was republished: %v %v", err, gh.statuses[sha])
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

// FuzzClassifyIncomplete: over arbitrary shard output, the classifier never
// selects a package the log does not name, and never calls a run incomplete
// unless some process hit the alarm, every failed process did, and every
// process or package that failed without a test is an alarmed one.
func FuzzClassifyIncomplete(f *testing.F) {
	for _, seed := range []string{
		"shard0    exit 1   1800.4s  1 failed\n--- shard0 hit its 30m0s test alarm in a/b; running: TestX\n--- FAIL a/b TestY\n",
		"shard0    exit 1   1800.4s  1 failed\n--- shard0 hit its 30m0s test alarm in a/b; running:",
		"--- shard0 hit its 30m0s test alarm in a/b; running: TestX\nshard0    exit 1   1800.4s  1 failed\n",
		"packages  exit 1    5.0s  0 failed\nshard0    exit 1   1800.4s  0 failed\n--- shard0 hit its 1s test alarm in a/b; running: none listed\n",
		"shard0    exit 1   1800.4s  1 failed\n--- shard0 hit its 30m0s test alarm in a/b; running: TestX\nFAIL\ta/c [build failed]\n",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, log string) {
		facts := parseSuiteLogFacts(log)
		run, ok := classifyIncomplete(facts)
		if !ok {
			return
		}
		alarmed := false
		for _, p := range facts.Processes {
			if p.Exit != 0 && len(p.Alarmed) == 0 {
				t.Fatalf("a non-alarmed failed process classified incomplete: %+v", p)
			}
			alarmed = alarmed || len(p.Alarmed) != 0
		}
		if !alarmed || len(run.Packages) == 0 {
			t.Fatalf("incomplete without an alarm or a selection: %+v", run)
		}
		alarmedPkg, alarmedLabel := map[string]bool{}, map[string]bool{}
		for _, p := range facts.Processes {
			for _, pkg := range p.Alarmed {
				alarmedPkg[pkg], alarmedLabel[p.Label] = true, true
			}
		}
		for _, why := range facts.Unattributed {
			if n := packageNoTestReason.FindStringSubmatch(why); n != nil && !alarmedPkg[n[1]] {
				t.Fatalf("an unalarmed package's failure classified incomplete: %q", why)
			}
			if m := alarmExplainedReason.FindStringSubmatch(why); m != nil && !alarmedLabel[m[1]] {
				t.Fatalf("an unalarmed process's failure classified incomplete: %q", why)
			}
		}
		for _, pkg := range run.Packages {
			if !strings.Contains(log, pkg) {
				t.Fatalf("selected %q, which the log never names", pkg)
			}
		}
	})
}

// Each guard on which runs the rule judges is pinned on its own (#630, BR-5).
func TestIncompleteEligible(t *testing.T) {
	failed := localRunRecord{Exit: 1}
	for _, c := range []struct {
		name, check, contract string
		rec                   localRunRecord
		want                  bool
	}{
		{"a failed untyped suite", suiteCheckName, "", failed, true},
		{"an affected run", affectedCheckName, "", failed, true},
		{"a typed runner's suite keeps today's classification", suiteCheckName, strings.Repeat("a", 64), failed, false},
		{"sdlc's own deadline", suiteCheckName, "", localRunRecord{Exit: 1, TimedOut: true}, false},
		{"no exit reported", suiteCheckName, "", localRunRecord{Exit: -1}, false},
		{"a pass", suiteCheckName, "", localRunRecord{}, false},
		{"another check", "lint", "", failed, false},
	} {
		if got := incompleteEligible(c.check, c.contract, c.rec); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

// Statistics read flake evidence only (isFlakeEvidence), so an incomplete run
// is neither a run nor a flake, like a reclassified one: a later pass on the
// head confirms no flake (#630).
func TestAggregateSuiteAttemptsIncompleteIsNoFlake(t *testing.T) {
	wall := 1800.0
	at := time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)
	incomplete := SuiteAttempt{Version: 5, ID: "a1", Repo: "o/r", SHA: "s1", Check: suiteCheckName, Exit: 1, Outcome: string(flakes.IncompleteTimeout),
		Failures: []string{"o/r/pkg.TestBarrier"}, StartedAt: at, FinishedAt: at.Add(time.Minute), WallSeconds: &wall}
	pass := SuiteAttempt{Version: 2, ID: "a2", Repo: "o/r", SHA: "s1", Check: suiteCheckName, Outcome: string(flakes.Passed),
		StartedAt: at.Add(2 * time.Minute), FinishedAt: at.Add(3 * time.Minute), WallSeconds: &wall}
	stats := aggregateSuiteAttempts([]SuiteAttempt{incomplete, pass}, at.Add(-time.Hour), "")
	if len(stats) != 1 || stats[0].Runs != 1 || stats[0].FlakeDenominator != 1 || stats[0].ConfirmedFlakeAttempts != 0 || len(stats[0].Flakes) != 0 {
		t.Fatalf("%+v", stats)
	}
}

// A real failure beside an alarmed shard is a test result: the merge gate
// diagnoses it against main as before (#630).
func TestSuiteMixedAlarmAndRealFailureDiagnoses(t *testing.T) {
	d, sha, _, stderr := incompleteRepo(t)
	out := strings.Join([]string{
		"shard0    exit 1   1800.4s  1 failed",
		"--- shard0 hit its 30m0s test alarm in o/r/pkg; running: TestSlow",
		"shard1    exit 1   300.0s  1 failed",
		"--- FAIL o/r/other TestReal",
		"--- FAIL o/r/other",
	}, "\n")
	check := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "printf '%s\\n' '" + out + "'; exit 1"}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	var incomplete *suiteIncompleteError
	if errors.As(err, &incomplete) || rec.passed() || rec.Outcome == flakes.IncompleteTimeout {
		t.Fatalf("a real failure was called incomplete: %+v %v\n%s", rec, err, stderr)
	}
	got := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: rec, Head: sha}, sha, d)
	if got == nil || !strings.Contains(got.Error(), "suite comparison") {
		t.Fatalf("the mixed run was not diagnosed against main: %v", got)
	}
}

// suiteRetried times out like suiteTimedOut until the retry's package list
// arrives (#630); the retry records the list it got in marker, then runs retry.
func suiteRetried(marker, retry string) localCheck {
	c := suiteTimedOut()
	first := c.Command[2]
	c.Command[2] = `if [ -n "$SDLC_AFFECTED_GO_PKGS" ]; then printf '%s' "$SDLC_AFFECTED_GO_PKGS" > ` + marker + `; ` + retry + `; else ` + first + `; fi`
	return c
}

// The operator's Done when: the retry runs only the packages that did not
// finish or had a named failure, and a retry that passes posts green.
func TestIncompleteRetryRunsOnlyItsPackagesAndLands(t *testing.T) {
	d, sha, gh, stderr := incompleteRepo(t)
	prev := knownFlakeRerunTimeout // a targeted rerun's cap, set below the check's timeout so the two differ
	knownFlakeRerunTimeout = 5 * d.minute
	t.Cleanup(func() { knownFlakeRerunTimeout = prev })
	var deadlines []time.Duration
	d.executionTimeout = func(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
		deadlines = append(deadlines, timeout)
		return context.WithTimeout(ctx, timeout)
	}
	marker := filepath.Join(t.TempDir(), "pkgs")
	rec, err := runLocalCheck(context.Background(), d, suiteRetried(marker, "echo ok"), sha)
	// The retry reruns whole packages under the check's own 30 minutes, not a
	// targeted rerun's 15.
	if len(deadlines) != 2 || deadlines[1] != 30*d.minute {
		t.Fatalf("execution deadlines %v, want the run's and the retry's, both the check's 30 minutes (minute=%v), not the targeted cap", deadlines, d.minute)
	}
	if err != nil || !rec.passed() || rec.Outcome != flakes.IncompleteRetried || gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("%+v %v %v\n%s", rec, err, gh.statuses[sha], stderr)
	}
	if got, _ := os.ReadFile(marker); string(got) != "o/r/pkg" {
		t.Fatalf("the retry ran %q, want only o/r/pkg", got)
	}
	if !strings.Contains(rec.statusDescription(), "pass · incomplete run retried") {
		t.Errorf("status %q", rec.statusDescription())
	}
	// The posted composite is re-verified like any other: a later reuse lands it.
	if reused, ok, err := republishSavedPass(context.Background(), d, suiteRetried(marker, "echo ok"), sha); err != nil || !ok || reused.Outcome != flakes.IncompleteRetried {
		t.Fatalf("the retried pass did not verify for reuse: %+v %v %v", reused, ok, err)
	}
}

// A retry that fails for real posts a failure the merge gate diagnoses
// against main, from the retry's own log.
func TestIncompleteRetryRealFailureDiagnoses(t *testing.T) {
	d, sha, gh, stderr := incompleteRepo(t)
	marker := filepath.Join(t.TempDir(), "pkgs")
	check := suiteRetried(marker, "printf '%s\\n' '--- FAIL o/r/pkg TestReal' '--- FAIL o/r/pkg'; exit 1")
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
		t.Fatalf("%+v %v %v\n%s", rec, err, gh.statuses[sha], stderr)
	}
	log, err := os.ReadFile(rec.LogPath)
	if err != nil || !strings.Contains(string(log), "TestReal") {
		t.Fatalf("the failure does not point at the retry's log: %s %v", rec.LogPath, err)
	}
	// The gate goes on to main's baseline: here the same fixture times out on
	// main too, so the comparison reports main's baseline as unknown.
	got := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: rec, Head: sha}, sha, d)
	if got == nil || !strings.Contains(got.Error(), "baseline on main could not run") || !strings.Contains(got.Error(), "main's baseline is unknown") {
		t.Fatalf("a real failure on the retry was not diagnosed against main: %v", got)
	}
}

// A main baseline that is itself incomplete never retries and says so.
func TestIncompleteBaselineNeverRetries(t *testing.T) {
	d, sha, _, stderr := incompleteRepo(t)
	d.SkipRecord = true
	marker := filepath.Join(t.TempDir(), "pkgs")
	_, err := runLocalCheck(context.Background(), d, suiteRetried(marker, "echo ok"), sha)
	var incomplete *suiteIncompleteError
	if !errors.As(err, &incomplete) || !incomplete.Baseline || incomplete.Retried || !strings.Contains(err.Error(), "main's baseline is unknown") {
		t.Fatalf("%v\n%s", err, stderr)
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatal("a main baseline retried")
	}
}

func TestIncompleteRetryCommands(t *testing.T) {
	suite, aff := []string{"make", "test"}, []string{"make", "affected"}
	if got := incompleteRetryCommands(suite, nil, []string{"a", "b"}); !reflect.DeepEqual(got, [][]string{affectedEnvCommand(suite, []string{"a", "b"}, nil)}) {
		t.Errorf("no affected declaration: %v", got)
	}
	if got := incompleteRetryCommands(suite, aff, []string{"a"}); !reflect.DeepEqual(got, [][]string{affectedEnvCommand(aff, []string{"a"}, nil)}) {
		t.Errorf("an affected declaration: %v", got)
	}
}

// The proof that lets an incomplete parent's retry land (#630) binds the
// selection the parent's own log implies, the link's version, and a receipt
// to its link; each tampering is refused.
func TestResolveIncompleteResumeProof(t *testing.T) {
	d, sha, _, stderr := incompleteRepo(t)
	marker := filepath.Join(t.TempDir(), "pkgs")
	if _, err := runLocalCheck(context.Background(), d, suiteRetried(marker, "echo ok"), sha); err != nil {
		t.Fatalf("%v\n%s", err, stderr)
	}
	store := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	snapshot, err := readSuiteEvidence(store, testStatsRecordLimit)
	if err != nil {
		t.Fatal(err)
	}
	var parent SuiteAttempt
	for _, a := range snapshot.Attempts {
		if a.RerunOf == "" {
			parent = a
		}
	}
	base, _, err := readResumeProofInput(store, snapshot, parent)
	if err != nil {
		t.Fatal(err)
	}
	base.Policy = flakes.PolicyRerun
	if p := resolveResumeProof(base); p.State != ResumePassed || p.Outcome.Kind != flakes.IncompleteRetried || !reflect.DeepEqual(p.Selection.GoPkgs, []string{"o/r/pkg"}) {
		t.Fatalf("the real proof: %+v", p)
	}
	withLink := func(f func(*ResumeLink)) ResumeProofInput {
		in := base
		l := *base.Link
		sel := *l.Selection
		l.Selection = &sel
		f(&l)
		in.Link = &l
		return in
	}
	for name, in := range map[string]ResumeProofInput{
		"a wider selection":  withLink(func(l *ResumeLink) { l.Selection.GoPkgs = []string{"o/r/other", "o/r/pkg"} }),
		"a group":            withLink(func(l *ResumeLink) { l.Selection.Groups = []string{"lua"} }),
		"a known-flake link": withLink(func(l *ResumeLink) { l.Version, l.Selection = 3, nil }),
		// The parent's own log must still say incomplete: the selection is
		// derived from it, never from the link.
		"a parent log no longer incomplete": func() ResumeProofInput {
			in := base
			in.ParentFacts.Processes = slices.Clone(base.ParentFacts.Processes)
			for i := range in.ParentFacts.Processes {
				in.ParentFacts.Processes[i].Alarmed = nil
			}
			return in
		}(),
		// Before any link exists, the guard alone stops an empty retry.
		"a parent log no longer incomplete, unlinked": func() ResumeProofInput {
			in := base
			in.Link, in.LinkSHA256, in.Child, in.ChildLog, in.ChildLogSHA256 = nil, "", nil, "", ""
			in.ParentFacts.Processes = slices.Clone(base.ParentFacts.Processes)
			for i := range in.ParentFacts.Processes {
				in.ParentFacts.Processes[i].Alarmed = nil
			}
			return in
		}(),
		"a receipt version 3": func() ResumeProofInput {
			in := base
			in.Child = nil
			in.Receipt = &ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "ee", Repo: base.Link.Repo, SHA: base.Link.SHA, Check: base.Link.Check, RerunOf: base.Link.RerunOf,
				LinkID: base.Link.ID, LinkSHA256: base.LinkSHA256, ChildID: base.Link.ChildID, Actor: "uid=1", Note: "stopped", Outcome: "interrupted", FinishedAt: base.Link.CreatedAt.Add(1)}
			return in
		}(),
	} {
		if p := resolveResumeProof(in); p.State != ResumeRefused {
			t.Errorf("%s: %+v", name, p)
		}
	}
}

func TestIncompleteRetryOutcome(t *testing.T) {
	retry := incompleteRun{Packages: []string{"o/r/pkg"}}
	alarmLog := "shard0    exit 1   1800.4s  0 failed\n--- shard0 hit its 30m0s test alarm in o/r/pkg; running: TestSlow\n"
	for _, c := range []struct {
		name     string
		exit     int
		timedOut bool
		log      string
		want     flakes.Kind
	}{
		{"a clean pass lands", 0, false, "ok\n", flakes.IncompleteRetried},
		{"exit 0 with a named failure does not", 0, false, "--- FAIL o/r/pkg TestX\n--- FAIL o/r/pkg\n", flakes.Unknown},
		{"cut short again", 1, false, alarmLog, flakes.IncompleteTimeout},
		{"sdlc's deadline is not the alarm", 1, true, alarmLog, flakes.Unattributable},
		{"a named failure", 1, false, "--- FAIL o/r/pkg TestX\n--- FAIL o/r/pkg\n", flakes.Unknown},
		{"an unnamed failure", 1, false, "FAIL\to/r/pkg [build failed]\n", flakes.Unattributable},
	} {
		if got := incompleteRetryOutcome(retry, c.exit, c.timedOut, parseSuiteLogFacts(c.log)); got.Kind != c.want {
			t.Errorf("%s: %s (%s), want %s", c.name, got.Kind, got.Reason, c.want)
		}
	}
}

// A known-flake suite bundle stays version 3: a version 5 selection link may
// fence only an incomplete parent's retry (#630).
func TestKnownFlakeSuiteLinkStaysVersion3(t *testing.T) {
	in := resumeCorePassed(t)
	if in.Parent.Check != suiteCheckName || resolveResumeProof(in).State != ResumePassed {
		t.Fatalf("fixture: %+v", resolveResumeProof(in))
	}
	l := *in.Link
	l.Version, l.Selection = 5, &ResumeSelection{GoPkgs: []string{"example.org/p"}, Groups: []string{}}
	in.Link = &l
	if p := resolveResumeProof(in); p.State != ResumeRefused {
		t.Fatalf("a known-flake suite parent accepted a version 5 selection link: %+v", p)
	}
}

// #630 inside a landing train (the evidence #635 waits on): the production
// conductor path, runSuite over newLocalCheckRunner exactly as the train wires
// it. An incomplete suite retries only its unfinished and failed packages and
// never diagnoses main; a real failure on the retry still does.
func TestLandingTrainIncompleteSuite(t *testing.T) {
	train := func(t *testing.T, retry string) (*conductor, *ciFakeGH, string, string) {
		t.Helper()
		d, sha, gh, _ := incompleteRepo(t)
		marker := filepath.Join(t.TempDir(), "pkgs")
		c := &conductor{ctx: context.Background(), heads: []string{sha}, base: sha, suite: suiteRetried(marker, retry), local: d}
		c.run = newLocalCheckRunner(c.local)
		return c, gh, sha, marker
	}
	t.Run("the retry passes: a pass, of only its packages, with no main baseline", func(t *testing.T) {
		c, gh, sha, marker := train(t, "echo ok")
		ev, err := c.runSuite(1, landingline.ExecInitial)
		done, ok := ev.(landingline.SuiteDone)
		if err != nil || !ok || done.Outcome != landingline.SuitePass || c.baseline != nil || gh.statuses[sha]["sdlc-local/suite"] != "success" {
			t.Fatalf("%+v %v baseline=%v", ev, err, c.baseline)
		}
		if got, _ := os.ReadFile(marker); string(got) != "o/r/pkg" {
			t.Fatalf("the retry ran %q, want only o/r/pkg", got)
		}
	})
	t.Run("incomplete again: could not run, no split, no main baseline", func(t *testing.T) {
		c, gh, sha, _ := train(t, suiteTimedOut().Command[2]) // the retry is cut short the same way
		ev, err := c.runSuite(1, landingline.ExecInitial)
		if err == nil || ev != nil || !strings.Contains(err.Error(), "could not run") || !strings.Contains(err.Error(), "incomplete (timeout) again") || c.baseline != nil || gh.statuses[sha]["sdlc-local/suite"] != "" {
			t.Fatalf("%+v %v baseline=%v", ev, err, c.baseline)
		}
	})
	t.Run("control: a real failure on the retry diagnoses against main", func(t *testing.T) {
		c, gh, sha, _ := train(t, "printf '%s\\n' '--- FAIL o/r/pkg TestReal' '--- FAIL o/r/pkg'; exit 1")
		_, err := c.runSuite(1, landingline.ExecInitial)
		// main's baseline runs: here the same fixture times out on main, so it reports main unknown.
		if err == nil || !strings.Contains(err.Error(), "train baseline on main could not run") || !strings.Contains(err.Error(), "main's baseline is unknown") || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("the real failure did not reach main's baseline: %v %v", err, gh.statuses[sha])
		}
	})
}
