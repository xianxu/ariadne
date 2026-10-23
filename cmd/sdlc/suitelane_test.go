package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func laneAttempt(repo, check string, finished int, wall float64) SuiteAttempt {
	t0 := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	return SuiteAttempt{Repo: repo, Check: check, FinishedAt: t0.Add(time.Duration(finished) * time.Minute), WallSeconds: &wall}
}

// The limits are written here as literals (3 samples, 120 s), independent of
// the constants, so changing a constant fails these cases.
func TestLaneEligibility(t *testing.T) {
	short := func(at int) SuiteAttempt { return laneAttempt("x/echo", "suite", at, 70) }
	long := func(at int) SuiteAttempt { return laneAttempt("x/echo", "suite", at, 717) }
	at := func(at int, wall float64) SuiteAttempt { return laneAttempt("x/echo", "suite", at, wall) }
	overrun := func(at int) SuiteAttempt {
		a := short(at)
		stopped := 240.0
		a.LaneOverrunSeconds = &stopped
		return a
	}
	timedOut := func(at int) SuiteAttempt { a := short(at); a.TimedOut = true; return a }
	baseline := func(at int) SuiteAttempt { a := short(at); a.MainBaseline = true; return a }
	noWall := func(at int) SuiteAttempt { a := short(at); a.WallSeconds = nil; return a }
	cases := []struct {
		name     string
		attempts []SuiteAttempt
		eligible bool
		reason   string
	}{
		{"none", nil, false, "needs 3 recorded suites, has 0"},
		{"two short", []SuiteAttempt{short(1), short(2)}, false, "has 2"},
		{"three short", []SuiteAttempt{short(1), short(2), short(3)}, true, "ran 70s, 70s, 70s, each under 120s"},
		{"just under", []SuiteAttempt{at(1, 119.9), short(2), short(3)}, true, ""},
		{"exactly 120", []SuiteAttempt{at(1, 120), short(2), short(3)}, false, "ran 120s (limit 120s)"},
		{"one long recent", []SuiteAttempt{short(1), long(2), short(3)}, false, "ran 717s"},
		{"long pushed out by three short", []SuiteAttempt{long(1), short(2), short(3), short(4)}, true, ""},
		{"listing order is not recency", []SuiteAttempt{short(4), long(1), short(3), short(2)}, true, ""},
		{"overrun recent", []SuiteAttempt{short(1), overrun(2), short(3)}, false, "stopped at 240s"},
		{"overrun pushed out", []SuiteAttempt{overrun(1), short(2), short(3), short(4)}, true, ""},
		{"timed out", []SuiteAttempt{short(1), timedOut(2), short(3)}, false, "ran 70s"},
		{"other repo and check ignored", []SuiteAttempt{short(1), short(2), laneAttempt("x/ariadne", "suite", 3, 700), laneAttempt("x/echo", "affected", 4, 700)}, false, "has 2"},
		{"baseline and unmeasured ignored", []SuiteAttempt{short(1), short(2), baseline(3), noWall(4)}, false, "has 2"},
	}
	for _, c := range cases {
		got := laneEligibility(c.attempts, "x/echo")
		if got.Eligible != c.eligible || !strings.Contains(got.Reason, c.reason) {
			t.Errorf("%s: eligible=%v reason=%q, want %v containing %q", c.name, got.Eligible, got.Reason, c.eligible, c.reason)
		}
	}
}

func TestNextSuiteLaneStep(t *testing.T) {
	legal := map[[2]int]suiteLaneState{
		{int(laneStart), int(laneEligible)}:       laneWaitLight,
		{int(laneStart), int(laneIneligible)}:     laneWaitNormal,
		{int(laneWaitLight), int(laneAcquired)}:   laneRunLight,
		{int(laneWaitLight), int(laneCancelled)}:  laneDone,
		{int(laneRunLight), int(laneExited)}:      laneDone,
		{int(laneRunLight), int(laneOverran)}:     laneWaitNormal,
		{int(laneRunLight), int(laneCancelled)}:   laneDone,
		{int(laneWaitNormal), int(laneAcquired)}:  laneRunNormal,
		{int(laneWaitNormal), int(laneCancelled)}: laneDone,
		{int(laneRunNormal), int(laneExited)}:     laneDone,
		{int(laneRunNormal), int(laneCancelled)}:  laneDone,
	}
	for s := laneStart; s <= laneDone; s++ {
		for e := laneEligible; e <= laneCancelled; e++ {
			got, err := nextSuiteLaneStep(s, e)
			want, ok := legal[[2]int{int(s), int(e)}]
			switch {
			case ok && (err != nil || got != want):
				t.Errorf("(%d,%d) = %d, %v; want %d", s, e, got, err, want)
			case !ok && (!errors.Is(err, errLaneEvent) || got != s):
				t.Errorf("(%d,%d) = %d, %v; want refused, state kept", s, e, got, err)
			}
		}
	}
	// An overrun leads to the normal line, never back to the light one.
	s := laneStart
	for _, e := range []suiteLaneEvent{laneEligible, laneAcquired, laneOverran, laneAcquired, laneOverran} {
		s, _ = nextSuiteLaneStep(s, e)
	}
	if s != laneRunNormal {
		t.Fatalf("a second overrun moved a normal run: %d", s)
	}
}

func TestSuiteAttemptLaneFieldsRoundTrip(t *testing.T) {
	stopped := 240.5
	a := storedAttemptFixture("a1")
	a.Lane, a.LaneOverrunSeconds = suiteLaneLight, &stopped
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeSuiteAttempt(data)
	if err != nil || got.Lane != "light" || got.LaneOverrunSeconds == nil || *got.LaneOverrunSeconds != 240.5 {
		t.Fatalf("lane evidence lost: %+v %v", got, err)
	}
	for name, bad := range map[string]string{
		"unknown lane":     strings.Replace(string(data), `"lane":"light"`, `"lane":"heavy"`, 1),
		"zero overrun":     strings.Replace(string(data), `"lane_overrun_seconds":240.5`, `"lane_overrun_seconds":0`, 1),
		"null overrun":     strings.Replace(string(data), `"lane_overrun_seconds":240.5`, `"lane_overrun_seconds":null`, 1),
		"aliased lane key": strings.Replace(string(data), `"lane":"light"`, `"Lane":"light"`, 1),
	} {
		if _, err := decodeSuiteAttempt([]byte(bad)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
