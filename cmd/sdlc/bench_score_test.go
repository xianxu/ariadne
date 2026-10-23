package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func benchScoreFixture(t *testing.T) (bench.Task, string, string, string) {
	t.Helper()
	task := preparedFixtureTask(t)
	// Replace the initial fixture's test-only patch with a runnable hidden
	// acceptance script. The tests exercise literal argv, not a shell wrapper
	// invented by the scorer.
	prepareTestGit(t, task.SourceRepo, "switch", "--detach", task.ParentOID)
	prepareTestWrite(t, task.SourceRepo, "tests/acceptance_test.sh", "#!/bin/sh\ngrep -q '^fixed' src/main.txt\n", 0755)
	prepareTestGit(t, task.SourceRepo, "add", "tests/acceptance_test.sh")
	patch := prepareTestGit(t, task.SourceRepo, "diff", "--cached", "--", "tests/acceptance_test.sh") + "\n"
	prepareTestWrite(t, task.SourceRepo, "src/main.txt", "fixed\n", 0644)
	prepareTestGit(t, task.SourceRepo, "add", "-A")
	prepareTestGit(t, task.SourceRepo, "commit", "-qm", "oracle landing")
	task.LandingOID = prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD")
	task.Acceptance.LandingOID = task.LandingOID
	task.Acceptance.Paths = []string{"tests/acceptance_test.sh"}
	task.Acceptance.Argv = []string{"sh", "tests/acceptance_test.sh"}
	sum := sha256.Sum256([]byte(patch))
	task.Acceptance.PatchSHA256 = hex.EncodeToString(sum[:])
	observer := filepath.Join(t.TempDir(), "observer")
	if err := os.Mkdir(observer, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(observer, "acceptance.patch"), []byte(patch), 0600); err != nil {
		t.Fatal(err)
	}
	// The source patch path is intentionally invalid: scoring must read only
	// the immutable captured copy above.
	task.Acceptance.Patch = filepath.Join(t.TempDir(), "must-not-be-read.patch")
	work := t.TempDir()
	prepareTestGit(t, work, "init", "-q", "-b", "main")
	prepareTestGit(t, work, "config", "user.name", "Fixture")
	prepareTestGit(t, work, "config", "user.email", "fixture@example.invalid")
	prepareTestWrite(t, work, "src/main.txt", "candidate\n", 0644)
	prepareTestGit(t, work, "add", "-A")
	prepareTestGit(t, work, "commit", "-qm", "candidate")
	candidate := prepareTestGit(t, work, "rev-parse", "HEAD")
	return task, work, candidate, observer
}

func TestBenchAcceptanceOracleOrderAndCandidateFailure(t *testing.T) {
	task, work, candidate, observer := benchScoreFixture(t)
	var roles []string
	execute := func(ctx context.Context, c benchScoreCommand) (benchScoreRun, error) {
		if _, bounded := ctx.Deadline(); !bounded {
			t.Fatal("acceptance execution has no outer wall-clock budget")
		}
		roles = append(roles, c.Role)
		if !runnerArtifactProtected(c.Root) {
			t.Fatal("scoring copy lacks a native protected Git worktree lease")
		}
		if _, err := os.Stat(filepath.Join(c.Root, "tests", "acceptance_test.sh")); err != nil {
			t.Fatal("hidden acceptance absent", err)
		}
		if c.WorkRoot != work || len(c.Argv) != 2 || c.Argv[0] != "sh" {
			t.Fatal("execution authority/argv changed", c)
		}
		exit := 0
		if c.Role != "landing" {
			exit = 1
		}
		return benchScoreRun{Exit: exit, Quiet: true}, nil
	}
	a, err := evaluateBenchAcceptanceWith(context.Background(), task, work, candidate, observer, benchScoreDeps{Execute: execute})
	if err != nil || a.Status != "command" || a.CommandPassed == nil || *a.CommandPassed || !a.ParentRed || !a.LandingGreen || a.Total != nil || strings.Join(roles, ",") != "parent,landing,candidate" {
		t.Fatalf("acceptance %+v %v roles %v", a, err, roles)
	}
}

func TestBenchAcceptanceRealManagedCommands(t *testing.T) {
	for _, kind := range []string{"candidate-fails", "candidate-passes", "missing-tool", "output-budget"} {
		t.Run(kind, func(t *testing.T) {
			task, work, candidate, observer := benchScoreFixture(t)
			prepareTestWrite(t, work, "Makefile.workflow", "# fixture helper authority\n", 0644)
			prepareTestWrite(t, work, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"), 0755)
			if kind == "candidate-passes" {
				prepareTestWrite(t, work, "src/main.txt", "fixed candidate\n", 0644)
				prepareTestGit(t, work, "add", "-A")
				prepareTestGit(t, work, "commit", "-qm", "good candidate")
				candidate = prepareTestGit(t, work, "rev-parse", "HEAD")
			}
			if kind == "missing-tool" {
				task.Acceptance.Argv = []string{"__bench_missing_acceptance_executable__"}
			}
			if kind == "output-budget" {
				task.Acceptance.Argv = []string{"python3", "-c", "import sys; sys.stdout.write('x' * (9 << 20))"}
			}
			a, err := evaluateBenchAcceptance(context.Background(), task, work, candidate, observer)
			if kind == "missing-tool" || kind == "output-budget" {
				if err == nil || a.Status != "unavailable" || a.CommandPassed != nil || a.ParentRed {
					t.Fatalf("startup fabricated test failure %+v %v", a, err)
				}
				return
			}
			if err != nil || a.Status != "command" || a.CommandPassed == nil || *a.CommandPassed != (kind == "candidate-passes") || !a.ParentRed || !a.LandingGreen {
				t.Fatalf("managed score %+v %v", a, err)
			}
			entries, err := os.ReadDir(observer)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if !strings.HasPrefix(entry.Name(), "score-") {
					continue
				}
				for _, role := range []string{"parent", "landing", "candidate"} {
					if _, err := os.Stat(filepath.Join(observer, entry.Name(), role)); !os.IsNotExist(err) {
						t.Fatal("confirmed scoring copy not cleaned", role, err)
					}
				}
			}
		})
	}
}

func TestBenchAcceptanceRefusesNonCommitPins(t *testing.T) {
	task, work, candidate, observer := benchScoreFixture(t)
	task.ParentOID = prepareTestGit(t, task.SourceRepo, "rev-parse", task.ParentOID+"^{tree}")
	task.Acceptance.ParentOID = task.ParentOID
	execute := func(_ context.Context, c benchScoreCommand) (benchScoreRun, error) {
		exit := 0
		if c.Role == "parent" {
			exit = 1
		}
		return benchScoreRun{Quiet: true, Exit: exit}, nil
	}
	a, err := evaluateBenchAcceptanceWith(context.Background(), task, work, candidate, observer, benchScoreDeps{Execute: execute})
	if err == nil || a.Status != "unavailable" || a.CommandPassed != nil {
		t.Fatalf("tree object accepted as commit pin: %+v %v", a, err)
	}
}

func TestBenchAcceptanceNeverScoresInvalidOraclesOrUnknownExecution(t *testing.T) {
	for _, kind := range []string{"green-parent", "red-landing", "startup", "unknown-cleanup", "timeout", "patch-scope", "patch-digest", "patch-collision"} {
		t.Run(kind, func(t *testing.T) {
			task, work, candidate, observer := benchScoreFixture(t)
			var roles []string
			switch kind {
			case "patch-scope":
				task.Acceptance.Paths = []string{"tests/other_test.sh"}
			case "patch-digest":
				task.Acceptance.PatchSHA256 = strings.Repeat("0", 64)
			case "patch-collision":
				prepareTestWrite(t, work, "tests/acceptance_test.sh", "collision\n", 0644)
				prepareTestGit(t, work, "add", "-A")
				prepareTestGit(t, work, "commit", "-qm", "collision")
				candidate = prepareTestGit(t, work, "rev-parse", "HEAD")
			}
			execute := func(_ context.Context, c benchScoreCommand) (benchScoreRun, error) {
				roles = append(roles, c.Role)
				r := benchScoreRun{Quiet: true, Exit: 0}
				if c.Role == "parent" {
					r.Exit = 1
				}
				switch kind {
				case "green-parent":
					r.Exit = 0
				case "red-landing":
					r.Exit = 1
				case "startup":
					r.Exit = -1
				case "unknown-cleanup":
					r.Quiet = false
					return r, errors.New("cleanup unknown")
				case "timeout":
					r.TimedOut = true
				}
				return r, nil
			}
			a, err := evaluateBenchAcceptanceWith(context.Background(), task, work, candidate, observer, benchScoreDeps{Execute: execute})
			if err == nil || a.Status != "unavailable" || a.CommandPassed != nil || a.Total != nil {
				t.Fatalf("invented score %+v %v", a, err)
			}
			if kind != "patch-collision" {
				for _, role := range roles {
					if role == "candidate" {
						t.Fatal("candidate executed without valid oracle proofs")
					}
				}
			}
			if kind == "unknown-cleanup" {
				var custody *runnerCleanupUnconfirmed
				if !errors.As(err, &custody) || !strings.Contains(custody.Detail, observer) || !strings.Contains(custody.Detail, "locked") {
					t.Fatalf("missing exact retained bundle/lease guidance: %v", err)
				}
				entries, _ := os.ReadDir(observer)
				retained := false
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "score-") {
						if _, err := os.Stat(filepath.Join(observer, entry.Name(), "parent", "anchor")); err == nil {
							retained = true
						}
					}
				}
				if !retained {
					t.Fatal("unconfirmed scoring inputs removed")
				}
			}
		})
	}
}
