package main

import (
	"context"
	"errors"
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
		asked := scriptRerun(t, "echo 'ok  o/r/pkg 0.1s'")
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
		if failed == nil || rerun == nil || rerun.RerunOf != failed.ID || failed.Exit != 1 || rerun.Exit != 0 {
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
