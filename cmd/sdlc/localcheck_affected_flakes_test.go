package main

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func asAffected(c localCheck) localCheck { c.Name = affectedCheckName; return c }

// #473 D4: the affected suite is a test run. A failure only registered flakes
// explain reruns them and lands as a known flake, like the suite's.
func TestAffectedKnownFlakeReruns(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	asked := scriptRerun(t, gatePassJSON)
	rec, err := runLocalCheck(context.Background(), d, asAffected(suiteFailing(gateKnown)), sha)
	if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakePassed {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	if strings.Join(*asked, ",") != gateKnown || gh.statuses[sha]["sdlc-local/affected"] != "success" {
		t.Fatalf("reran %v, statuses %v", *asked, gh.statuses[sha])
	}
}

// #586: an affected run stopped at a registered flake may never have reached
// its groups, so its known-flake child runs them too; a failing group means
// no known-flake pass, and a selection without groups reruns only the tests.
func TestAffectedKnownFlakeRunsGroups(t *testing.T) {
	for _, tc := range []struct {
		name   string
		groups []string
		group  string
		pass   bool
	}{
		{"groups pass", []string{"a", "b"}, "exit 0", true},
		{"a group fails", []string{"a", "b"}, "exit 1", false},
		{"no groups", nil, "exit 1", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, sha, gh, stderr := localRunRepo(t, "acquired")
			d.FlakeRegistry = gateRegistry(t, gateKnown)
			scriptRerun(t, gatePassJSON)
			marker := t.TempDir() + "/groups"
			c := asAffected(suiteFailing(gateKnown))
			c.Rerun = &affectedRerun{Command: []string{"sh", "-c", `printf '%s|%s' "$SDLC_AFFECTED_GO_PKGS" "$SDLC_AFFECTED_GROUPS" > ` + marker + "; " + tc.group}, GoPkgs: []string{"o/r/pkg"}, Groups: tc.groups}
			rec, _ := runLocalCheck(context.Background(), d, c, sha)
			got, readErr := os.ReadFile(marker)
			if len(tc.groups) > 0 && (readErr != nil || string(got) != "|"+strings.Join(tc.groups, " ")) {
				t.Fatalf("groups command ran with %q (%v)\n%s", got, readErr, stderr)
			}
			if len(tc.groups) == 0 && readErr == nil {
				t.Fatalf("no groups selected, yet the groups command ran: %q", got)
			}
			passed := rec.passed() && rec.Outcome == flakes.KnownFlakePassed
			if passed != tc.pass || (gh.statuses[sha]["sdlc-local/affected"] == "success") != tc.pass {
				t.Fatalf("pass %v, want %v: %+v statuses %v\n%s", passed, tc.pass, rec, gh.statuses[sha], stderr)
			}
		})
	}
}

// #473 D4: an unexplained failure on a head, in either test run, blocks a
// success of either: a passing affected run must not green-wash a failed
// suite on the same commit, nor a passing suite an affected failure.
func TestTestRunGuardSpansSuiteAndAffected(t *testing.T) {
	pass := localCheck{Command: []string{"sh", "-c", "echo passed"}, TimeoutMinutes: 30}
	for _, tc := range []struct{ failed, passed string }{
		{suiteCheckName, affectedCheckName},
		{affectedCheckName, suiteCheckName},
		{affectedCheckName, affectedCheckName},
	} {
		t.Run(tc.failed+" then "+tc.passed, func(t *testing.T) {
			ctx := context.Background()
			d, sha, gh, _ := localRunRepo(t, "acquired")
			d.FlakeRegistry = gateRegistry(t, "o/r/pkg.TestOther")
			failing := suiteFailing(gateUnknown)
			failing.Name = tc.failed
			if rec, err := runLocalCheck(ctx, d, failing, sha); err != nil || rec.passed() {
				t.Fatalf("%+v %v", rec, err)
			}
			p := pass
			p.Name = tc.passed
			_, err := runLocalCheck(ctx, d, p, sha)
			var unexplained *unexplainedFailureError
			if !errors.As(err, &unexplained) || !strings.Contains(err.Error(), gateUnknown) || !strings.Contains(err.Error(), "local check "+tc.passed) {
				t.Fatalf("a pass over an unexplained %s failure: %v", tc.failed, err)
			}
			if gh.statuses[sha][localStatusContext(tc.passed)] == "success" {
				t.Fatal("the pass was posted")
			}
		})
	}
}

// #473 D4: a subset run's pass never confirms a fail→pass: it may not have
// run the failed test. A full suite pass, or a targeted rerun, still does.
func TestSubsetPassNeverConfirms(t *testing.T) {
	at := func(m int) time.Time { return time.Date(2026, 10, 20, 9, m, 0, 0, time.UTC) }
	failed := flakes.Attempt{ID: "f", SHA: "s", Failures: []string{gateKnown}, FinishedAt: at(1)}
	subset := flakes.Attempt{ID: "a", SHA: "s", Passed: true, Subset: true, FinishedAt: at(2)}
	full := flakes.Attempt{ID: "p", SHA: "s", Passed: true, FinishedAt: at(3)}
	if _, _, ok := flakes.SameHeadFailPass([]flakes.Attempt{failed, subset}, gateKnown); ok {
		t.Fatal("an affected pass confirmed a fail→pass")
	}
	if f, p, ok := flakes.SameHeadFailPass([]flakes.Attempt{failed, subset, full}, gateKnown); !ok || f != "f" || p != "p" {
		t.Fatalf("a full pass after it: %s %s %v", f, p, ok)
	}
	if a := flakeAttempt(SuiteAttempt{Check: affectedCheckName}); !a.Subset {
		t.Fatal("an affected attempt is not marked a subset")
	}
	if a := flakeAttempt(SuiteAttempt{Check: suiteRerunCheck}); a.Subset {
		t.Fatal("a targeted rerun was marked a subset: it runs exactly the failed tests")
	}
}

// #473 D4: affected runs get their own test-stats row; the suite's numbers
// are what they were without them.
func TestAffectedStatsAreTheirOwnRow(t *testing.T) {
	t0 := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	wall := func(s float64) *float64 { return &s }
	suite := []SuiteAttempt{
		{ID: "s1", Repo: "pair", SHA: "a", Check: suiteCheckName, StartedAt: t0, FinishedAt: t0.Add(12 * time.Minute), WallSeconds: wall(720)},
		{ID: "s2", Repo: "pair", SHA: "b", Check: suiteCheckName, StartedAt: t0, FinishedAt: t0.Add(13 * time.Minute), WallSeconds: wall(780)},
	}
	affected := []SuiteAttempt{
		{ID: "a1", Repo: "pair", SHA: "c", Check: affectedCheckName, StartedAt: t0, FinishedAt: t0.Add(2 * time.Minute), WallSeconds: wall(120)},
	}
	alone := aggregateSuiteAttempts(suite, t0.Add(-time.Hour), "pair")
	mixed := aggregateSuiteAttempts(append(append([]SuiteAttempt(nil), suite...), affected...), t0.Add(-time.Hour), "pair")
	if !reflect.DeepEqual(alone, mixed) {
		t.Fatalf("affected runs moved the suite row:\n%+v\n%+v", alone, mixed)
	}
	rows := aggregateCheckAttempts(append(suite, affected...), t0.Add(-time.Hour), "pair", affectedCheckName)
	if len(rows) != 1 || rows[0].Check != affectedCheckName || rows[0].Runs != 1 || rows[0].Wall == nil {
		t.Fatalf("affected row %+v", rows)
	}
}

// #473 M2: a repo whose window holds only affected runs reports one affected
// row and no empty suite row; with both, the suite row comes first.
func TestTestStatsRowsNeedRuns(t *testing.T) {
	t0 := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	aff := SuiteAttempt{ID: "a1", Repo: "pair", SHA: "c", Check: affectedCheckName, StartedAt: t0, FinishedAt: t0.Add(time.Minute)}
	suite := SuiteAttempt{ID: "s1", Repo: "pair", SHA: "a", Check: suiteCheckName, StartedAt: t0, FinishedAt: t0.Add(time.Minute)}
	if rows := testStatsRows([]SuiteAttempt{aff}, t0.Add(-time.Hour), "pair"); len(rows) != 1 || rows[0].Check != affectedCheckName {
		t.Fatalf("affected only: %+v", rows)
	}
	if rows := testStatsRows([]SuiteAttempt{aff, suite}, t0.Add(-time.Hour), "pair"); len(rows) != 2 || rows[0].Check != "" || rows[1].Check != affectedCheckName {
		t.Fatalf("both: %+v", rows)
	}
}
