package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// laneHelper is a fake suite-lock helper that declares --lane when lane is
// set, logs each acquisition's argv, and answers acquired-light to a light
// request.
func laneHelper(t *testing.T, d localRunDeps, lane bool) string {
	t.Helper()
	argvLog := filepath.Join(t.TempDir(), "argv")
	options := `options:\n  --watch WATCH  owner\n  --slots SLOTS  weight`
	if lane {
		options += `\n  --lane LANE  the light lane`
	}
	body := fmt.Sprintf(`import sys
if sys.argv[1:] == ['acquire', '--help']:
    print('%s')
    sys.exit(0)
with open(%q, 'a') as log:
    log.write(' '.join(sys.argv[1:]) + '\n')
print('acquired-light' if '--lane' in sys.argv else 'acquired')
`, options, argvLog)
	if err := os.WriteFile(filepath.Join(d.Root, "scripts", "test-lock.py"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return argvLog
}

// seedSuiteAttempts records n finished suites of wall seconds each for d's repo.
func seedSuiteAttempts(t *testing.T, d localRunDeps, n int, wall float64) {
	t.Helper()
	dir := filepath.Join(d.Root, ".git", "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	for i := range n {
		w, q := wall, 1.0
		a := SuiteAttempt{Version: suiteAttemptVersion, ID: fmt.Sprintf("5ee0%04d", i), Repo: d.Repo, SHA: strings.Repeat("a", 40),
			Check: suiteCheckName, StartedAt: t0.Add(time.Duration(i) * time.Hour), FinishedAt: t0.Add(time.Duration(i)*time.Hour + time.Minute),
			QueueSeconds: &q, WallSeconds: &w, LogSHA256: strings.Repeat("b", 64), LogPath: filepath.Join("local-checks", strings.Repeat("a", 40), "suite.log"), Outcome: "passed"}
		data, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, a.ID+".json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func laneAcquisitions(t *testing.T, argvLog string) []string {
	t.Helper()
	b, err := os.ReadFile(argvLog)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

// #559 Done-when 2: a light run that overruns is stopped, retried once in the
// normal line, and the repo leaves the lane by measurement.
func TestRunLocalCheckLightOverrunRetriesInNormalLine(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	argvLog := laneHelper(t, d, true)
	seedSuiteAttempts(t, d, 3, 60)
	d.laneBound = time.Second
	scratch := t.TempDir()
	tried := filepath.Join(scratch, "tried")
	// The first (light) run outlasts the bound; the retry finds the marker and
	// passes at once.
	script := fmt.Sprintf(`if [ -f %[1]q ]; then echo retried; exit 0; fi; touch %[1]q; sleep 30`, tried)
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"sh", "-c", script}, TimeoutMinutes: 30}, sha)
	if err != nil || !rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("retry did not pass and record: %+v %v %v\n%s", rec, err, gh.statuses[sha], stderr)
	}
	calls := laneAcquisitions(t, argvLog)
	if len(calls) != 2 || !strings.Contains(calls[0], "--lane light") || strings.Contains(calls[1], "--lane") {
		t.Fatalf("want one light then one normal acquisition, got %q", calls)
	}
	// Only the light run is niced. (Agent sandboxes deny setpriority, so the
	// command's announced argv is the observable, not its niceness.)
	if strings.Count(stderr.String(), `command: ["nice" "-n" "10" "sh" "-c"`) != 1 || strings.Count(stderr.String(), `command: ["sh" "-c"`) != 1 {
		t.Fatalf("want one niced light run and one plain retry:\n%s", stderr)
	}
	if !strings.Contains(stderr.String(), "==> light lane: its last 3 suites ran 60s, 60s, 60s") ||
		!strings.Contains(stderr.String(), "retrying it in the normal line") {
		t.Fatalf("lane decisions not announced:\n%s", stderr)
	}
	rows, warnings, err := readSuiteAttempts(filepath.Join(d.Root, ".git", "sdlc", "test-stats"), 100)
	if err != nil || len(warnings) != 0 || len(rows) != 4 {
		t.Fatalf("attempts: %d %v %v", len(rows), warnings, err)
	}
	var retry *SuiteAttempt
	for i := range rows {
		if rows[i].SHA == sha {
			retry = &rows[i]
		}
	}
	if retry == nil || retry.Lane != "" || retry.LaneOverrunSeconds == nil || *retry.LaneOverrunSeconds < 0.5 || *retry.LaneOverrunSeconds > 10 {
		t.Fatalf("the retry does not carry the overrun: %+v", retry)
	}
	if e := laneEligibility(rows, d.Repo); e.Eligible || !strings.Contains(e.Reason, "stopped at") {
		t.Fatalf("the overrun left the repo eligible: %+v", e)
	}
}

func TestRunLocalCheckLightLaneChoice(t *testing.T) {
	run := func(t *testing.T, seeded int, wall float64, lane bool, check string) (string, []string, *SuiteAttempt) {
		d, sha, _, stderr := localRunRepo(t, "acquired")
		argvLog := laneHelper(t, d, lane)
		seedSuiteAttempts(t, d, seeded, wall)
		rec, err := runLocalCheck(context.Background(), d, localCheck{Name: check, Command: []string{"true"}, TimeoutMinutes: 30}, sha)
		if err != nil || !rec.passed() {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
		rows, _, _ := readSuiteAttempts(filepath.Join(d.Root, ".git", "sdlc", "test-stats"), 100)
		for i := range rows {
			if rows[i].SHA == sha {
				return stderr.String(), laneAcquisitions(t, argvLog), &rows[i]
			}
		}
		return stderr.String(), laneAcquisitions(t, argvLog), nil
	}
	t.Run("short repo runs light and records the lane", func(t *testing.T) {
		_, calls, a := run(t, 3, 60, true, "suite")
		if len(calls) != 1 || !strings.Contains(calls[0], "--lane light") || a == nil || a.Lane != "light" || a.LaneOverrunSeconds != nil {
			t.Fatalf("%q %+v", calls, a)
		}
	})
	t.Run("too few measurements stay normal", func(t *testing.T) {
		_, calls, _ := run(t, 2, 60, true, "suite")
		if len(calls) != 1 || strings.Contains(calls[0], "--lane") {
			t.Fatalf("%q", calls)
		}
	})
	t.Run("long suites stay normal", func(t *testing.T) {
		_, calls, _ := run(t, 3, 700, true, "suite")
		if len(calls) != 1 || strings.Contains(calls[0], "--lane") {
			t.Fatalf("%q", calls)
		}
	})
	t.Run("a helper without the lane falls back, saying so", func(t *testing.T) {
		stderr, calls, a := run(t, 3, 60, false, "suite")
		if len(calls) != 1 || strings.Contains(calls[0], "--lane") || a == nil || a.Lane != "" || !strings.Contains(stderr, "has no --lane; using the normal line") {
			t.Fatalf("%q %+v\n%s", calls, a, stderr)
		}
	})
	t.Run("other checks never ask", func(t *testing.T) {
		_, calls, _ := run(t, 3, 60, true, "lint")
		if len(calls) != 1 || strings.Contains(calls[0], "--lane") {
			t.Fatalf("%q", calls)
		}
	})
}
