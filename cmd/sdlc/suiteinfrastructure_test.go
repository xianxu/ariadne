package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Recorded verbatim from #508's failed candidate suite on 2026-10-21. The
// diagnostic main failed the same way. No volume-filling fixture is needed.
func recordedCacheFailure(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/suite-enospc-gocache.log")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func infrastructureCheck(t *testing.T, output string) localCheck {
	t.Helper()
	p := filepath.Join(t.TempDir(), "startup.log")
	if err := os.WriteFile(p, []byte(output), 0o600); err != nil {
		t.Fatal(err)
	}
	return localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `cat "$1"; exit 1`, "fixture", p}, TimeoutMinutes: 1}
}

func TestSuiteInfrastructureGrammar(t *testing.T) {
	capture := recordedCacheFailure(t)
	for _, tc := range []struct {
		name, log string
	}{
		{"recorded cache and go-list wrapper", capture},
		{"quota and wrapper", strings.ReplaceAll(capture, "no space left on device", "disk quota exceeded")},
		{"direct cache write", "pattern ./...: write /cache/go-build/ab/object-a: no space left on device\n"},
		{"cache initialization", "go: failed to initialize build cache at /cache/go-build: mkdir /cache/go-build/00: no space left on device\n"},
		{"build output", "go: copying /tmp/go-build123/exe/app: write /build/bin/app: disk quota exceeded\n"},
		{"missing toolchain", "go: cannot find GOROOT directory: /missing/go\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !suiteEnvironmentLog(tc.log) {
				t.Fatalf("startup infrastructure became code failure: %q", tc.log)
			}
			for _, activity := range []string{
				"=== RUN   TestActual\n",
				"--- PASS: TestActual (0.01s)\n",
				"--- FAIL: TestActual (0.01s)\n",
				"ok  example/pkg 0.1s\n",
				`{"Action":"run","Package":"example/pkg","Test":"TestActual"}` + "\n",
				"x.go:3: undefined: broken\n",
				"unknown diagnostic\n",
			} {
				if suiteEnvironmentLog(activity+tc.log) || suiteEnvironmentLog(tc.log+activity) {
					t.Errorf("mixed/test-started output escaped failure authority: %q", activity)
				}
			}
			if suiteEnvironmentLog(strings.TrimSuffix(tc.log, "\n")) || suiteEnvironmentLog("    "+strings.ReplaceAll(strings.TrimSuffix(tc.log, "\n"), "\n", "\n    ")+"\n") {
				t.Error("truncated or quoted startup diagnostic accepted")
			}
		})
	}
}

func TestSuiteInfrastructureFailureHasNoCodeAuthority(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	check := infrastructureCheck(t, recordedCacheFailure(t))
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	var infra *suiteEnvironmentError
	if !errors.As(err, &infra) {
		t.Fatalf("want could-not-run, got rec=%+v err=%v\n%s", rec, err, stderr)
	}
	if !strings.Contains(err.Error(), "could-not-run") || !strings.Contains(err.Error(), "/Volumes/AgentScratch/go-build/") {
		t.Fatalf("resource/path missing from recovery: %v", err)
	}
	if len(gateAttempts(t, d)) != 0 || len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("infrastructure published code authority: attempts=%+v statuses=%v", gateAttempts(t, d), gh.statuses)
	}
	if b, err := os.ReadFile(rec.LogPath); err != nil || string(b) != recordedCacheFailure(t) {
		t.Fatalf("captured output not retained: %q %v", b, err)
	}
	if diagnoseSuiteFailure(context.Background(), err, sha, d) != err {
		t.Fatal("startup failure entered diagnostic baseline")
	}
	check.Command = []string{"true"}
	if pass, err := runLocalCheck(context.Background(), d, check, sha); err != nil || !pass.passed() || gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("genuine same-head repair blocked: %+v %v\n%s", pass, err, stderr)
	}
}

func TestSuiteInfrastructureBaselineCannotBlameCars(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	check := infrastructureCheck(t, recordedCacheFailure(t))
	log := filepath.Join(t.TempDir(), "candidate.log")
	if err := os.WriteFile(log, []byte("--- FAIL: TestActual (0.01s)\nFAIL example/pkg 0.01s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	failed := localRunRecord{Exit: 1, LogPath: log}
	err := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: failed, Head: sha}, sha, d)
	var infra *suiteEnvironmentError
	if !errors.As(err, &infra) || strings.Contains(err.Error(), "main is red") {
		t.Fatalf("baseline infrastructure became main-red: %v", err)
	}
	c := &conductor{ctx: context.Background(), heads: []string{sha}, base: sha, suite: check, local: d}
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return failed, nil }
	if event, err := c.runSuite(1); err == nil || event != nil || c.baseline != nil {
		t.Fatalf("baseline infrastructure blamed cars: event=%+v err=%v baseline=%v", event, err, c.baseline)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("baseline infrastructure posted a result: %+v %+v", gh.statuses, gh.comments)
	}
}
