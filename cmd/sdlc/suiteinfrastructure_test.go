package main

import (
	"context"
	"errors"
	"io"
	"iter"
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

func advanceInfrastructureCallerHead(t *testing.T, d localRunDeps, executed string) {
	t.Helper()
	if out, err := d.Git("commit", "--allow-empty", "-qm", "advance caller beyond executed head"); err != nil {
		t.Fatalf("advance caller: %s %v", out, err)
	}
	if head, err := d.Git("rev-parse", "HEAD"); err != nil || strings.TrimSpace(head) == executed {
		t.Fatalf("fixture did not separate executed head: %s %v", head, err)
	}
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
	for _, name := range []string{suiteCheckName, affectedCheckName} {
		t.Run(name, func(t *testing.T) { testSuiteInfrastructureFailureHasNoCodeAuthority(t, name) })
	}
}

func testSuiteInfrastructureFailureHasNoCodeAuthority(t *testing.T, name string) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	advanceInfrastructureCallerHead(t, d, sha)
	check := infrastructureCheck(t, recordedCacheFailure(t))
	check.Name = name
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	var infra *suiteEnvironmentError
	if !errors.As(err, &infra) {
		t.Fatalf("want could-not-run, got rec=%+v err=%v\n%s", rec, err, stderr)
	}
	if !strings.Contains(err.Error(), "could-not-run") || !strings.Contains(err.Error(), "/Volumes/AgentScratch/go-build/") {
		t.Fatalf("resource/path missing from recovery: %v", err)
	}
	if !strings.Contains(err.Error(), "--name "+name+" --rev "+sha) {
		t.Fatalf("live recovery targets caller HEAD instead of executed head: %v", err)
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
	if pass, err := runLocalCheck(context.Background(), d, check, sha); err != nil || !pass.passed() || gh.statuses[sha][localStatusContext(name)] != "success" {
		t.Fatalf("genuine same-head repair blocked: %+v %v\n%s", pass, err, stderr)
	}
}

func TestSuiteInfrastructureRecordedCacheKnownFlakeRerun(t *testing.T) {
	for _, name := range []string{suiteCheckName, affectedCheckName} {
		t.Run(name, func(t *testing.T) { testSuiteInfrastructureRecordedCacheKnownFlakeRerun(t, name) })
	}
}

func testSuiteInfrastructureRecordedCacheKnownFlakeRerun(t *testing.T, name string) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	advanceInfrastructureCallerHead(t, d, sha)
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	argv := infrastructureCheck(t, recordedCacheFailure(t)).Command
	previous := knownFlakeRerunCommands
	t.Cleanup(func() { knownFlakeRerunCommands = previous })
	knownFlakeRerunCommands = func([]string) ([][]string, error) { return [][]string{argv}, nil }
	check := suiteFailing(gateKnown)
	check.Name = name
	_, err := runLocalCheck(context.Background(), d, check, sha)
	var infra *suiteEnvironmentError
	if !errors.As(err, &infra) {
		t.Fatalf("recorded rerun startup failure lost execution provenance: %v", err)
	}
	if !strings.Contains(err.Error(), "--name "+name+" --rev "+sha) {
		t.Fatalf("rerun recovery dropped parent check/executed head: %v", err)
	}
	rows := gateAttempts(t, d)
	if len(rows) != 1 || rows[0].Check != name || len(gh.statuses[sha]) != 0 {
		t.Fatalf("startup rerun published code authority: %+v %+v", rows, gh.statuses)
	}
	if log, err := os.ReadFile(infra.LogPath); err != nil || string(log) != recordedCacheFailure(t) {
		t.Fatalf("rerun raw capture changed: %q %v", log, err)
	}
}

func TestSuiteInfrastructureBaselineCannotBlameCars(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	advanceInfrastructureCallerHead(t, d, sha)
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
	if !strings.Contains(err.Error(), "--name suite --rev "+sha) {
		t.Fatalf("baseline recovery targets caller instead of baseline: %v", err)
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

func TestSuiteInfrastructureMissingGoNeverStarts(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	advanceInfrastructureCallerHead(t, d, sha)
	missing := filepath.Join(t.TempDir(), "go")
	check := localCheck{Name: suiteCheckName, Command: []string{missing, "test", "./..."}, TimeoutMinutes: 1}
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	var infra *suiteEnvironmentError
	if !errors.As(err, &infra) || !strings.Contains(err.Error(), missing) {
		t.Fatalf("missing Go was not could-not-run: %+v %v\n%s", rec, err, stderr)
	}
	if !strings.Contains(err.Error(), "--name suite --rev "+sha) {
		t.Fatalf("unstarted Go recovery lost executed head: %v", err)
	}
	if len(gateAttempts(t, d)) != 0 || len(gh.statuses[sha]) != 0 {
		t.Fatalf("unstarted tool published code authority: %+v %v", gateAttempts(t, d), gh.statuses)
	}
	if log, err := os.ReadFile(rec.LogPath); err != nil || len(log) != 0 {
		t.Fatalf("raw empty output changed: %q %v", log, err)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	if leases, err := os.ReadDir(localCheckRunsDir(logDir)); err != nil || len(leases) != 0 {
		t.Fatalf("unstarted Go retained lease: %v %v", leases, err)
	}
	if trees, err := d.Git("worktree", "list", "--porcelain"); err != nil || strings.Count(trees, "worktree ") != 1 {
		t.Fatalf("unstarted Go retained checkout: %s %v", trees, err)
	}
}

func TestSuiteInfrastructureMissingGoBatchControls(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "go")
	interpreter := filepath.Join(t.TempDir(), "go")
	if err := os.WriteFile(interpreter, []byte("#!/missing-530-interpreter\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	denied := filepath.Join(t.TempDir(), "go")
	if err := os.WriteFile(denied, []byte("#!/bin/sh\nexit 127\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	exited := filepath.Join(t.TempDir(), "go")
	if err := os.WriteFile(exited, []byte("#!/bin/sh\nexit 127\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		commands [][]string
		startup  bool
	}{
		{"first Go", [][]string{{missing, "test", "./..."}}, true},
		{"missing interpreter", [][]string{{interpreter}}, true},
		{"permission denied", [][]string{{denied}}, false},
		{"started and exited 127", [][]string{{exited}}, false},
		{"after earlier command", [][]string{{"true"}, {missing, "test", "./..."}}, false},
		{"unrelated executable", [][]string{{missing + "-other"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := localRunDeps{Stderr: io.Discard}
			err, cleanup := runManagedCommandBatch(context.Background(), d, t.TempDir(), suiteCheckName, tc.commands, os.Environ(), io.Discard, "")
			var infra *suiteEnvironmentError
			if err == nil || cleanup != nil || errors.As(err, &infra) != tc.startup {
				t.Fatalf("startup=%v error=%v cleanup=%v", tc.startup, err, cleanup)
			}
		})
	}
}

func TestSuiteInfrastructureLinearScan(t *testing.T) {
	capture := recordedCacheFailure(t)
	for _, n := range []int{8, 16} {
		visits := 0
		lines := func(yield func(string) bool) {
			for range n {
				for line := range strings.Lines(capture) {
					visits++
					if !yield(line) {
						return
					}
				}
			}
		}
		_, ok := diagnoseSuiteStartupLines(iter.Seq[string](lines))
		if want := n * strings.Count(capture, "\n"); !ok || visits != want {
			t.Fatalf("N=%d visits=%d want=%d classified=%v", n, visits, want, ok)
		}
	}
}

func FuzzDiagnoseSuiteStartup(f *testing.F) {
	capture, err := os.ReadFile("testdata/suite-enospc-gocache.log")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(string(capture))
	f.Add(workDirFailure + "\n")
	f.Add("go: cannot find GOROOT directory: /missing/go\n")
	f.Add("")
	f.Fuzz(func(t *testing.T, log string) {
		diagnosis, ok := diagnoseSuiteStartup(log)
		if !ok {
			return
		}
		if diagnosis.Resource == "" || diagnosis.Path == "" || diagnosis.Reason == "" {
			t.Fatalf("incomplete diagnosis: %+v", diagnosis)
		}
		started := `{"Action":"run","Package":"example/pkg","Test":"TestActuallyStarted"}` + "\n"
		if _, accepted := diagnoseSuiteStartup(log + started); accepted {
			t.Fatal("test activity accepted after startup diagnostic")
		}
		if _, accepted := diagnoseSuiteStartup(started + log); accepted {
			t.Fatal("test activity accepted before startup diagnostic")
		}
		if _, accepted := diagnoseSuiteStartup(strings.TrimSuffix(log, "\n")); accepted {
			t.Fatal("truncated diagnostic accepted")
		}
	})
}
