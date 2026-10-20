package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func stubPlanInputs(costs ...planCost) planInputs {
	return func() ([]planCost, []string, error) { return costs, nil, nil }
}

func cost(repo string, weight int, mins float64, samples int) planCost {
	return planCost{RepoCost: landingline.RepoCost{Repo: repo, Weight: weight, Duration: time.Duration(mins * float64(time.Minute))}, Interval: time.Hour, Samples: samples}
}

// #475: `landing plan` prints each repository's slot and basis; --write
// writes the plan file beside the suite lock, keeping unknown members and
// other repositories' entries; an unreadable plan is never overwritten.
func TestLandingPlanCommand(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("WF_TEST_LOCK_FILE", "")
	t.Setenv("WF_TEST_SUITES", "")
	path := filepath.Join(state, "ariadne", "landing-plan.json")
	inputs := stubPlanInputs(cost("pair", 1, 12.5, 9), cost("ariadne", 2, 9.5, 14), cost("parley.nvim", 1, 2.7, 0))

	var out bytes.Buffer
	if err := runLandingPlan(&out, false, inputs); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"2 suite slot(s), every 60m", ":00  ariadne", "p90 9.5m x weight 2 (14 samples)", ":15  pair", ":15  parley.nvim", "no samples", "dry run"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("dry run lacks %q:\n%s", want, out.String())
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("a dry run wrote the plan")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"slots":2,"owner":"steward","repos":{"echo":{"offset_minutes":40,"basis":"kept","pin":true}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runLandingPlan(&out, true, inputs); err != nil {
		t.Fatal(err)
	}
	p, err := readMachinePlan(path)
	if err != nil {
		t.Fatal(err)
	}
	if off, ok := p.Offset("pair", time.Hour); !ok || off != 15*time.Minute {
		t.Errorf("pair at %v %v", off, ok)
	}
	if off, ok := p.Offset("echo", time.Hour); !ok || off != 40*time.Minute {
		t.Errorf("another repository's entry was dropped: %v %v", off, ok)
	}
	data, _ := os.ReadFile(path)
	for _, want := range []string{`"owner":"steward"`, `"pin":true`} {
		if !strings.Contains(strings.ReplaceAll(string(data), " ", ""), want) {
			t.Errorf("--write lost %s:\n%s", want, data)
		}
	}

	if err := os.WriteFile(path, []byte(`{"version":9}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runLandingPlan(io.Discard, true, inputs); err == nil || !strings.Contains(err.Error(), "unreadable plan") {
		t.Fatalf("overwrote an unreadable plan: %v", err)
	}
}

// #475: the schedule source places the repository on its plan offset, and a
// bad plan departs on the interval, warned once.
func TestScheduleSourceReadsThePlan(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("WF_TEST_LOCK_FILE", "")
	g := func(args ...string) (string, error) {
		switch args[0] {
		case "ls-tree":
			return landingConfigPath + "\n", nil
		case "show":
			return `{"train_interval_minutes": 60}`, nil
		}
		return "upstream", nil
	}
	path := filepath.Join(state, "ariadne", "landing-plan.json")
	procedureWrite(t, filepath.Dir(path), "landing-plan.json", `{"version":1,"slots":2,"repos":{"sample":{"offset_minutes":15}}}`)
	if s := landingScheduleSource(g, "upstream", "sample", io.Discard)(); s.Offset == nil || *s.Offset != 15*time.Minute {
		t.Fatalf("offset %v", s.Offset)
	}
	if s := landingScheduleSource(g, "upstream", "other", io.Discard)(); s.Offset != nil {
		t.Fatalf("a repository without an entry got %v", *s.Offset)
	}
	procedureWrite(t, filepath.Dir(path), "landing-plan.json", `{"version":1,"slots":2,"repos":{"sample":{"offset_minutes":"x"}}}`)
	var stderr bytes.Buffer
	src := landingScheduleSource(g, "upstream", "sample", &stderr)
	oldTTL := landingScheduleTTL
	landingScheduleTTL = 0
	t.Cleanup(func() { landingScheduleTTL = oldTTL })
	for i := 0; i < 3; i++ {
		if s := src(); s.Offset != nil || s.Interval != time.Hour {
			t.Fatalf("bad plan: %+v", s)
		}
	}
	if n := strings.Count(stderr.String(), "machine plan unreadable"); n != 1 {
		t.Fatalf("warned %d times:\n%s", n, stderr.String())
	}
}

// #475 Done-when: two repositories planned from their costs; stepping a fake
// clock through two hours after each one's last train, each one's
// slot-triggered departure falls in its own window, never inside the other's
// reserved window.
func TestPlannedReposNeverDepartInEachOthersWindow(t *testing.T) {
	costs := []planCost{cost("ariadne", 2, 9.5, 10), cost("pair", 1, 12.5, 10)}
	entries, interval, _ := planFromCosts(costs, 2)
	start := time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)
	window := func(c planCost) (time.Duration, time.Duration) {
		off := time.Duration(entries[c.Repo].OffsetMinutes) * time.Minute
		span := (c.Duration+landingline.PlanStep-1)/landingline.PlanStep*landingline.PlanStep + landingline.PlanStep
		return off, off + span
	}
	departures := map[string][]time.Time{}
	for _, c := range costs {
		off := time.Duration(entries[c.Repo].OffsetMinutes) * time.Minute
		s := landingline.Schedule{Interval: interval, MaxCars: 8, Offset: &off}
		events := []landingline.Event{{Kind: landingline.EventTrain, Result: "landed", At: start, Departed: start, NextDue: landingline.NextDue(start, s, false)}}
		for m := 1; m <= 120; m++ {
			now := start.Add(time.Duration(m) * time.Minute)
			if due, why := landingline.DepartureDue(now, s, events, landingline.Waiters{}); due && why == "train due (slot)" {
				departures[c.Repo] = append(departures[c.Repo], now)
				events = append(events, landingline.Event{Kind: landingline.EventTrain, Result: "landed", At: now, Departed: now, NextDue: landingline.NextDue(now, s, false)})
			}
		}
	}
	for _, c := range costs {
		if len(departures[c.Repo]) != 2 {
			t.Fatalf("%s departed %v; want once an hour", c.Repo, departures[c.Repo])
		}
		for _, other := range costs {
			if other.Repo == c.Repo {
				continue
			}
			lo, hi := window(other)
			for _, at := range departures[c.Repo] {
				m := time.Duration(at.Minute()) * time.Minute
				if m >= lo && m < hi {
					t.Errorf("%s departs at :%02d, inside %s's window [%v, %v)", c.Repo, at.Minute(), other.Repo, lo, hi)
				}
			}
		}
	}
}
