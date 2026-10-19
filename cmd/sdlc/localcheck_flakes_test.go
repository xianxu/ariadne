package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

const (
	gateKnown   = "o/r/pkg.TestFlaky"
	gateUnknown = "o/r/pkg.TestBroken"
)

// suiteFailing is a suite command whose log names failures the way
// test-shard reports them.
func suiteFailing(tests ...string) localCheck {
	var b strings.Builder
	for _, t := range tests {
		pkg, fn, _ := flakes.SplitTest(t)
		b.WriteString("echo '--- FAIL " + pkg + " " + fn + "'; echo '--- FAIL " + pkg + "'; ")
	}
	return localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", b.String() + "exit 1"}, TimeoutMinutes: 30}
}

func gateRegistry(t *testing.T, tests ...string) func(context.Context) (flakes.Registry, error) {
	t.Helper()
	r, err := flakes.BuildRegistry([]flakes.Card{{ID: "000448", Status: "open", Tests: tests}}, flakeIssueOpen)
	if err != nil {
		t.Fatal(err)
	}
	return func(context.Context) (flakes.Registry, error) { return r, nil }
}

// scriptRerun makes the targeted rerun run script and records its argv.
func scriptRerun(t *testing.T, script string) *[]string {
	t.Helper()
	var asked []string
	prev := knownFlakeRerunArgv
	t.Cleanup(func() { knownFlakeRerunArgv = prev })
	knownFlakeRerunArgv = func(tests []string) ([]string, error) {
		asked = append(asked, tests...)
		return []string{"sh", "-c", script}, nil
	}
	return &asked
}

func gateAttempts(t *testing.T, d localRunDeps) []SuiteAttempt {
	t.Helper()
	rows, warnings, err := readSuiteAttempts(filepath.Join(d.Root, ".git", "sdlc", "test-stats"), 100)
	if err != nil || len(warnings) != 0 {
		t.Fatalf("attempts: %v %v", warnings, err)
	}
	return rows
}

func TestKnownFlakeGate(t *testing.T) {
	ctx := context.Background()
	t.Run("a registered failure reruns and lands with its issue", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		asked := scriptRerun(t, "echo '--- PASS: TestFlaky (0.00s)'; echo 'ok  o/r/pkg 0.1s'")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakePassed {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
		if strings.Join(*asked, ",") != gateKnown {
			t.Errorf("reran %v, want only the registered test", *asked)
		}
		if gh.statuses[sha]["sdlc-local/suite"] != "success" || !strings.Contains(rec.statusDescription(), "pass · known flake #000448 rerun") {
			t.Errorf("status %v, description %q", gh.statuses[sha], rec.statusDescription())
		}
		rows := gateAttempts(t, d)
		var failed, rerun *SuiteAttempt
		for i := range rows {
			if rows[i].RerunOf != "" {
				rerun = &rows[i]
			} else {
				failed = &rows[i]
			}
		}
		if failed == nil || rerun == nil || rerun.RerunOf != failed.ID || failed.Exit != 1 || rerun.Exit != 0 || rerun.Check != suiteRerunCheck || failed.Check != suiteCheckName {
			t.Fatalf("both attempts kept and linked: %+v", rows)
		}
		if failed.Outcome != string(flakes.KnownFlakePassed) || len(failed.KnownFlakes) != 1 || failed.KnownFlakes[0].Issue != "000448" {
			t.Errorf("failed attempt %+v", failed)
		}
		if len(gh.comments[9]) != 0 {
			t.Error("a known-flake landing posted a failure comment")
		}
	})
	t.Run("the rerun fails too: blocked", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		scriptRerun(t, "echo '--- FAIL o/r/pkg TestFlaky'; exit 1")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.KnownFlakeTwice || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v %v", rec, err, gh.statuses)
		}
	})
	t.Run("an unregistered failure blocks without a rerun", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		asked := scriptRerun(t, "exit 0")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown, gateUnknown), sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || len(*asked) != 0 || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v asked %v", rec, err, *asked)
		}
		if !strings.Contains(rec.statusDescription(), "unknown-failure") {
			t.Errorf("description %q", rec.statusDescription())
		}
	})
	t.Run("policy accept lands without a rerun", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		asked := scriptRerun(t, "exit 1")
		c := suiteFailing(gateKnown)
		c.KnownFlakes = "accept"
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakeAccept || len(*asked) != 0 || gh.statuses[sha]["sdlc-local/suite"] != "success" {
			t.Fatalf("%+v %v asked %v", rec, err, *asked)
		}
	})
	t.Run("an unreadable registry blocks", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = func(context.Context) (flakes.Registry, error) {
			return flakes.Registry{}, errors.New("tracker unreachable")
		}
		asked := scriptRerun(t, "exit 0")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.RegistryUnusable || len(*asked) != 0 {
			t.Fatalf("%+v %v", rec, err)
		}
	})
	t.Run("a failure naming no test blocks", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		asked := scriptRerun(t, "exit 0")
		c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo '--- shard1 exited 2 with no failing test; last output:'; exit 1"}, TimeoutMinutes: 30}
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.Unattributable || len(*asked) != 0 {
			t.Fatalf("%+v %v", rec, err)
		}
	})
	t.Run("an interrupted rerun records nothing new and posts nothing", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		cctx, cancel := context.WithCancel(ctx)
		defer cancel()
		prev := knownFlakeRerunArgv
		t.Cleanup(func() { knownFlakeRerunArgv = prev })
		knownFlakeRerunArgv = func([]string) ([]string, error) {
			time.AfterFunc(300*time.Millisecond, cancel)
			return []string{"sleep", "30"}, nil
		}
		_, err := runLocalCheck(cctx, d, suiteFailing(gateKnown), sha)
		if err == nil || !strings.Contains(err.Error(), "known-flake rerun did not complete") || len(gh.statuses[sha]) != 0 {
			t.Fatalf("err %v statuses %v", err, gh.statuses[sha])
		}
		rows := gateAttempts(t, d)
		if len(rows) != 1 || rows[0].Exit != 1 || rows[0].Outcome != "" {
			t.Fatalf("the failed attempt stays, unjudged: %+v", rows)
		}
	})
	t.Run("a suite killed by its timeout never lands (BR-11)", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		asked := scriptRerun(t, "echo '--- PASS: TestFlaky (0.00s)'")
		c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo '--- FAIL o/r/pkg TestFlaky'; echo '--- FAIL o/r/pkg'; sleep 5"}, TimeoutMinutes: 1}
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil || rec.passed() || !rec.TimedOut || rec.Outcome != flakes.Unattributable || len(*asked) != 0 || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v asked %v statuses %v", rec, err, *asked, gh.statuses[sha])
		}
	})
	t.Run("a rerun that shows no PASS proves nothing", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		scriptRerun(t, "echo 'ok  o/r/pkg 0.1s [no tests to run]'")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.RerunInconclusive || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v", rec, err)
		}
	})
	t.Run("the flake issue closes during the rerun: blocked, and not reused (BR-12)", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "acquired")
		open := gateRegistry(t, gateKnown)
		closed := false
		d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) {
			if closed {
				return flakes.Registry{}, nil
			}
			return open(ctx)
		}
		prev := knownFlakeRerunArgv
		t.Cleanup(func() { knownFlakeRerunArgv = prev })
		knownFlakeRerunArgv = func([]string) ([]string, error) {
			closed = true // the issue closes while the rerun runs
			return []string{"sh", "-c", "echo '--- PASS: TestFlaky (0.00s)'"}, nil
		}
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || gh.statuses[sha]["sdlc-local/suite"] != "failure" || !strings.Contains(stderr.String(), "no longer a known flake") {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
	})
	t.Run("a saved known-flake pass is reused only while it holds", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		scriptRerun(t, "echo '--- PASS: TestFlaky (0.00s)'")
		c := suiteFailing(gateKnown)
		if rec, err := runLocalCheck(ctx, d, c, sha); err != nil || !rec.passed() {
			t.Fatalf("%+v %v", rec, err)
		}
		delete(gh.statuses, sha)
		d.FlakeRegistry = func(context.Context) (flakes.Registry, error) { return flakes.Registry{}, nil } // the issue closed
		if _, ok, err := republishSavedPass(ctx, d, c, sha); ok || err != nil {
			t.Fatalf("reused a saved known-flake pass whose issue closed: %v %v", ok, err)
		}
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		if rec, ok, err := republishSavedPass(ctx, d, c, sha); !ok || err != nil || !rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "success" {
			t.Fatalf("a still-valid saved known-flake pass: %+v %v %v\n%s", rec, ok, err, d.Stderr)
		}
	})
	t.Run("an inconclusive rerun is never reused as the head's pass (BR-19)", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		scriptRerun(t, "echo 'ok  o/r/pkg 0.1s [no tests to run]'")
		c := suiteFailing(gateKnown)
		if rec, err := runLocalCheck(ctx, d, c, sha); err != nil || rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v", rec, err)
		}
		if rec, ok, err := republishSavedPass(ctx, d, c, sha); ok || err != nil || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("merge reused the rerun as a suite pass: %+v %v %v %v", rec, ok, err, gh.statuses[sha])
		}
	})
	t.Run("the rerun is bounded by the check's timeout (BR-16)", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		prev := knownFlakeRerunTimeout
		t.Cleanup(func() { knownFlakeRerunTimeout = prev })
		knownFlakeRerunTimeout = time.Hour
		scriptRerun(t, "sleep 10; echo '--- PASS: TestFlaky (0.00s)'")
		c := suiteFailing(gateKnown)
		c.TimeoutMinutes = 1 // one second under the test's minute
		start := time.Now()
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.RerunInconclusive || time.Since(start) > 8*time.Second {
			t.Fatalf("%+v %v after %s", rec, err, time.Since(start))
		}
	})
	t.Run("a blind full rerun never turns an unknown failure green; registering then republishing does (M3)", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, "o/r/pkg.TestOther")
		if rec, err := runLocalCheck(ctx, d, suiteFailing(gateUnknown), sha); err != nil || rec.passed() {
			t.Fatalf("%+v %v", rec, err)
		}
		pass := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo passed"}, TimeoutMinutes: 30}
		rec, err := runLocalCheck(ctx, d, pass, sha)
		var unexplained *unexplainedFailureError
		if !rec.passed() || !errors.As(err, &unexplained) || !strings.Contains(err.Error(), gateUnknown) || !strings.Contains(err.Error(), "sdlc flake register") || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("a blind rerun: %+v %v %v", rec, err, gh.statuses[sha])
		}
		d.FlakeRegistry = gateRegistry(t, gateUnknown) // the TL registers it, with this fail→pass as evidence
		if rec, ok, err := republishSavedPass(ctx, d, pass, sha); !ok || err != nil || !rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "success" {
			t.Fatalf("after registration the saved pass republishes: %+v %v %v %v", rec, ok, err, gh.statuses[sha])
		}
	})
	t.Run("an unattributable failure stays unexplained whatever is registered", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo '--- shard1 exited 2 with no failing test; last output:'; exit 1"}, TimeoutMinutes: 30}
		_, _ = runLocalCheck(ctx, d, c, sha)
		var unexplained *unexplainedFailureError
		if _, err := runLocalCheck(ctx, d, localCheck{Name: suiteCheckName, Command: []string{"true"}, TimeoutMinutes: 30}, sha); !errors.As(err, &unexplained) || gh.statuses[sha]["sdlc-local/suite"] != "failure" || strings.Contains(err.Error(), "flake register") || !strings.Contains(err.Error(), "cannot explain it") {
			t.Fatalf("%v %v", err, gh.statuses[sha])
		}
	})
	t.Run("a baseline never classifies", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		d.FlakeRegistry = gateRegistry(t, gateKnown)
		d.SkipRecord = true
		asked := scriptRerun(t, "exit 0")
		rec, err := runLocalCheck(ctx, d, suiteFailing(gateKnown), sha)
		if err != nil || rec.passed() || rec.Outcome != "" || len(*asked) != 0 {
			t.Fatalf("%+v %v", rec, err)
		}
	})
}

// nestedFlakeRepo adds a real Go module to a local-run repo: TestScripts has
// subtests plain/dev-aliases.test.sh (flaky: fails once per $FLAKE_DIR, then
// passes) and plain/other.test.sh (fails while $FLAKE_DIR/sibling exists).
// The suite is plain `go test -v`; the rerun is the real knownFlakeRerunArgv.
func nestedFlakeRepo(t *testing.T) (localRunDeps, string, *ciFakeGH, *bytes.Buffer, string) {
	t.Helper()
	d, _, gh, stderr := localRunRepo(t, "acquired")
	dir := t.TempDir()
	t.Setenv("FLAKE_DIR", dir)
	for rel, body := range map[string]string{
		"go.mod": "module example.com/flaky\n\ngo 1.22\n",
		"pkg/nested_test.go": `package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScripts(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		t.Run("dev-aliases.test.sh", func(t *testing.T) {
			mark := filepath.Join(os.Getenv("FLAKE_DIR"), "dev")
			if _, err := os.Stat(mark); err != nil {
				os.WriteFile(mark, nil, 0o644)
				t.Fatal("flaky once")
			}
		})
		t.Run("other.test.sh", func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(os.Getenv("FLAKE_DIR"), "sibling")); err == nil {
				t.Fatal("the sibling is broken")
			}
		})
	})
}
`,
	} {
		p := filepath.Join(d.Root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "nested tests"}} {
		if out, err := d.Git(args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	out, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(out)
	gh.setHead(9, sha)
	return d, sha, gh, stderr, dir
}

// #466 done-when: a registered subtest's same-head failure takes the
// known-flake rerun and lands; the same registration never explains its
// sibling's failure.
func TestKnownFlakeGateSubtest(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a real Go package")
	}
	ctx := context.Background()
	const top = "example.com/flaky/pkg.TestScripts"
	suite := localCheck{Name: suiteCheckName, Command: []string{"go", "test", "-count=1", "-v", "./..."}, TimeoutMinutes: 30}
	t.Run("the registered subtest reruns and lands", func(t *testing.T) {
		d, sha, gh, stderr, _ := nestedFlakeRepo(t)
		d.FlakeRegistry = gateRegistry(t, top+"/plain/dev-aliases.test.sh")
		rec, err := runLocalCheck(ctx, d, suite, sha)
		if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakePassed {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
		if gh.statuses[sha]["sdlc-local/suite"] != "success" {
			t.Fatalf("statuses %v", gh.statuses)
		}
		if !strings.Contains(stderr.String(), "-run ^(TestScripts)$") {
			t.Fatalf("the rerun was not its whole top-level test:\n%s", stderr)
		}
	})
	t.Run("a sibling's failure is unknown and blocks", func(t *testing.T) {
		d, sha, gh, stderr, dir := nestedFlakeRepo(t)
		d.FlakeRegistry = gateRegistry(t, top+"/plain/dev-aliases.test.sh")
		if err := os.WriteFile(filepath.Join(dir, "dev"), nil, 0o644); err != nil { // the registered one passes
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sibling"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		rec, err := runLocalCheck(ctx, d, suite, sha)
		if err != nil || rec.passed() || rec.Outcome != flakes.Unknown || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
		if !strings.Contains(stderr.String(), "plain/other.test.sh") {
			t.Fatalf("the refusal does not name the sibling:\n%s", stderr)
		}
	})
}

// The suite log's failing subtests (#466), in each format the parser reads;
// the records' top-level failures are unchanged.
func TestFailingSubtestsInEveryLogFormat(t *testing.T) {
	plain := `=== RUN   TestX
=== RUN   TestX/plain
=== RUN   TestX/plain/dev-aliases.test.sh
    x_test.go:9: flaky
--- FAIL: TestX (0.01s)
    --- FAIL: TestX/plain (0.01s)
        --- FAIL: TestX/plain/dev-aliases.test.sh (0.00s)
FAIL
FAIL	example.com/m/pkg	0.02s
`
	sharded := "--- FAIL example.com/m/scripts/test TestConstructScripts/plain\n--- FAIL example.com/m/scripts/test TestConstructScripts\n--- FAIL example.com/m/scripts/test\n"
	jsonLog := `{"Action":"fail","Package":"example.com/m/pkg","Test":"TestX/plain/a.sh"}
{"Action":"fail","Package":"example.com/m/pkg","Test":"TestX/plain"}
{"Action":"fail","Package":"example.com/m/pkg","Test":"TestX"}
{"Action":"fail","Package":"example.com/m/pkg","Elapsed":0.1}
`
	for name, c := range map[string]struct {
		log, top string
		subs     []string
	}{
		"go test -v": {plain, "example.com/m/pkg.TestX", []string{"example.com/m/pkg.TestX/plain", "example.com/m/pkg.TestX/plain/dev-aliases.test.sh"}},
		"sharded":    {sharded, "example.com/m/scripts/test.TestConstructScripts", []string{"example.com/m/scripts/test.TestConstructScripts/plain"}},
		"json":       {jsonLog, "example.com/m/pkg.TestX", []string{"example.com/m/pkg.TestX/plain", "example.com/m/pkg.TestX/plain/a.sh"}},
	} {
		facts := parseSuiteLogFacts(c.log)
		if strings.Join(facts.Failures, ",") != c.top || strings.Join(facts.Subtests, ",") != strings.Join(c.subs, ",") {
			t.Errorf("%s: failures %v subtests %v", name, facts.Failures, facts.Subtests)
		}
	}
	data, err := json.Marshal(parseSuiteLogFacts(plain))
	if err != nil || strings.Contains(string(data), "dev-aliases") {
		t.Fatalf("subtests reached the persisted facts: %s %v", data, err)
	}
}
