package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The machine plan's command (#475): `sdlc landing plan` shows where each
// train-enabled repository departs on the machine's grid and why; `--write`
// recomputes the offsets from measured suite cost and writes the plan file,
// keeping its unknown members. It is the steward's, run after costs change
// (a suite's shard count, a new train-enabled repository).

// planCost is one train-enabled repository's measured cost.
type planCost struct {
	landingline.RepoCost
	Interval time.Duration
	Samples  int // p90 samples; 0: no measurement, a default stood in
}

// noSampleDuration stands in for a repository with no measured suite.
const noSampleDuration = 15 * time.Minute

// planInputs gathers the costs; warnings name what was skipped and why.
type planInputs func() (costs []planCost, warnings []string, err error)

// planFromCosts packs the costs on the longest interval among them (the
// repositories share one grid; differing intervals are warned) into the
// machine's slots, and returns each repository's entry.
func planFromCosts(costs []planCost, slots int) (map[string]landingline.PlanEntry, time.Duration, []string) {
	var interval time.Duration
	var warnings []string
	for _, c := range costs {
		if interval != 0 && c.Interval != interval {
			warnings = append(warnings, "train intervals differ; the plan packs on the longest")
		}
		interval = max(interval, c.Interval)
	}
	rc := make([]landingline.RepoCost, len(costs))
	for i, c := range costs {
		rc[i] = c.RepoCost
	}
	offsets, over := landingline.PackOffsets(rc, interval, slots)
	names := make([]string, 0, len(over))
	for repo := range over {
		names = append(names, repo)
	}
	sort.Strings(names)
	for _, repo := range names {
		warnings = append(warnings, fmt.Sprintf("%s does not fit: its reservation peaks at %d of %d suite slots (the costs exceed one interval; lengthen it or shorten a suite)", repo, over[repo], slots))
	}
	out := map[string]landingline.PlanEntry{}
	for _, c := range costs {
		basis := fmt.Sprintf("p90 %s x weight %d (%d samples)", minutes(c.Duration), c.Weight, c.Samples)
		if c.Samples == 0 {
			basis = fmt.Sprintf("no samples: %s x weight %d", minutes(c.Duration), c.Weight)
		}
		out[c.Repo] = landingline.PlanEntry{OffsetMinutes: int(offsets[c.Repo] / time.Minute), Basis: basis}
	}
	return out, interval, dedupe(warnings)
}

func minutes(d time.Duration) string {
	return strconv.FormatFloat(d.Minutes(), 'f', 1, 64) + "m"
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// renderPlan prints the plan as a table, by offset then repository.
func renderPlan(out io.Writer, path string, slots int, interval time.Duration, entries map[string]landingline.PlanEntry) {
	fmt.Fprintf(out, "machine plan %s: %d suite slot(s), every %dm\n", path, slots, int(interval/time.Minute))
	names := make([]string, 0, len(entries))
	for n := range entries {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		a, b := entries[names[i]], entries[names[j]]
		if a.OffsetMinutes != b.OffsetMinutes {
			return a.OffsetMinutes < b.OffsetMinutes
		}
		return names[i] < names[j]
	})
	for _, n := range names {
		e := entries[n]
		fmt.Fprintf(out, "  :%02d  %-16s %s\n", e.OffsetMinutes, n, e.Basis)
	}
	if len(names) == 0 {
		fmt.Fprintln(out, "  no repository has trains on")
	}
}

// machineSuiteSlots is the machine's full-suite slots (scripts/test-lock.py).
func machineSuiteSlots() int {
	if n, err := strconv.Atoi(os.Getenv("WF_TEST_SUITES")); err == nil && n > 0 {
		return n
	}
	return 2
}

// runLandingPlan prints the plan computed from today's costs; with write it
// replaces the plan file, keeping its unknown members and its entries for
// repositories that no longer have trains on.
func runLandingPlan(out io.Writer, write bool, inputs planInputs) error {
	costs, warnings, err := inputs()
	if err != nil {
		return err
	}
	slots := machineSuiteSlots()
	entries, interval, more := planFromCosts(costs, slots)
	path := machinePlanPathFn()
	renderPlan(out, path, slots, interval, entries)
	for _, w := range append(warnings, more...) {
		fmt.Fprintln(out, "  [!] "+w)
	}
	if !write {
		fmt.Fprintln(out, "dry run: --write replaces the plan file (the steward's)")
		return nil
	}
	old, err := readMachinePlan(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("refusing to overwrite an unreadable plan (%v); fix or remove it first", err)
	}
	data, err := old.WithOffsets(slots, entries).MarshalJSON()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", path)
	return nil
}

// fleetPlanInputs measures each fleet repository whose main turns trains on:
// its train interval, its declared suite weight and its suite's p90 wall time
// over the last 7 days (test-stats). Its sources are injected: the fleet's
// repositories, the test-stats report, and git run in a repository.
func fleetPlanInputs(repos func() (map[string]string, error), report func() (testStatsReport, error), git func(root string, args ...string) (string, error)) planInputs {
	return func() ([]planCost, []string, error) {
		paths, err := repos()
		if err != nil {
			return nil, nil, err
		}
		stats, err := report()
		if err != nil {
			return nil, nil, err
		}
		p90 := map[string]*SuiteQuantiles{}
		for _, r := range stats.Repos {
			p90[r.Repo] = r.Wall
		}
		var costs []planCost
		var warnings []string
		names := make([]string, 0, len(paths))
		for n := range paths {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			root := paths[name]
			g := func(args ...string) (string, error) { return git(root, args...) }
			main := landingMainRef(g, "")
			s, err := readLandingSchedule(g, main) // the setting alone: the plan is what this computes
			if err != nil {
				warnings = append(warnings, name+": "+err.Error())
				continue
			}
			if s.Off() {
				continue
			}
			weight := 1
			if decl, err := readLocalChecks(g, main); err == nil {
				if c, ok := decl.check(suiteCheckName); ok && c.Weight > 0 {
					weight = c.Weight
				}
			}
			c := planCost{RepoCost: landingline.RepoCost{Repo: name, Weight: weight, Duration: noSampleDuration}, Interval: s.Interval}
			if q := p90[name]; q != nil && q.Samples > 0 {
				c.Duration, c.Samples = time.Duration(q.P90*float64(time.Second)), q.Samples
			}
			costs = append(costs, c)
		}
		return costs, warnings, nil
	}
}

// defaultPlanInputs wires fleetPlanInputs to the fleet, test-stats and git.
func defaultPlanInputs() planInputs {
	deps := defaultTestStatsDeps()
	return fleetPlanInputs(
		func() (map[string]string, error) { return fleetRepoPaths(deps) },
		func() (testStatsReport, error) { return buildTestStatsReport(testStatsFlags{Since: "7d"}, deps) },
		func(root string, args ...string) (string, error) { return landingGit(execGitRunner{}, root, args...) },
	)
}

func newLandingPlanCmd(inputs func() planInputs) *cobra.Command {
	var write bool
	c := &cobra.Command{
		Use:   "plan [--write]",
		Short: "Show (or, with --write, recompute) the machine's train departure plan: each repository's slot and its cost basis",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLandingPlan(cmd.OutOrStdout(), write, inputs())
		},
	}
	c.Flags().BoolVar(&write, "write", false, "replace the machine plan file (the steward's)")
	return c
}
