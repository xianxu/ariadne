package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const (
	testStatsRecordLimit = 10000
	testStatsLogLimit    = 64 << 20
)

type testStatsFlags struct {
	Since string
	Repo  string
	JSON  bool
}

type testStatsDeps struct {
	now       func() time.Time
	identity  func() (workspace.Identity, error)
	fleetDirs func(string) ([]string, error)
	commonDir func(string) (string, error)
}

type testStatsReport struct {
	Since    time.Time   `json:"since"`
	Repos    []RepoStats `json:"repos"`
	Warnings []string    `json:"warnings"`
}

func defaultTestStatsDeps() testStatsDeps {
	return testStatsDeps{
		now: time.Now,
		identity: func() (workspace.Identity, error) {
			return workspace.Resolve(execGitRunner{}, ".", "")
		},
		fleetDirs: project.FleetRepoDirs,
		commonDir: func(root string) (string, error) {
			out, err := execGitRunner{}.GitInDir(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
			return strings.TrimSpace(string(out)), err
		},
	}
}

func NewTestStatsCmd() *cobra.Command {
	flags := testStatsFlags{}
	cmd := &cobra.Command{
		Use: "test-stats", Args: cobra.NoArgs,
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTestStats(cmd.OutOrStdout(), flags, defaultTestStatsDeps())
		},
	}
	cmd.Flags().StringVar(&flags.Since, "since", "7d", "time window ending now (for example 7d or 12h)")
	cmd.Flags().StringVar(&flags.Repo, "repo", "", "show one sibling repository by name")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "emit JSON with explicit nulls for missing metrics")
	cmd.AddCommand(newWeeklySnapshotCmd())
	return cmd
}

func parseTestStatsSince(raw string, now time.Time) (time.Time, error) {
	if strings.HasSuffix(raw, "d") {
		days := strings.TrimSuffix(raw, "d")
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err != nil || n <= 0 || fmt.Sprint(n) != days || n > 3650 {
			return time.Time{}, fmt.Errorf("test-stats: invalid --since %q (use a positive duration such as 7d or 12h)", raw)
		}
		return now.Add(-time.Duration(n) * 24 * time.Hour), nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return time.Time{}, fmt.Errorf("test-stats: invalid --since %q (use a positive duration such as 7d or 12h)", raw)
	}
	return now.Add(-d), nil
}

func runTestStats(out io.Writer, flags testStatsFlags, deps testStatsDeps) error {
	report, err := buildTestStatsReport(flags, deps)
	if err != nil {
		return err
	}
	if flags.JSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}
	return renderTestStatsText(out, report)
}

// fleetRepoPaths is every fleet repository's primary checkout, by name:
// fleet discovery plus this workspace's own primary (a numbered slot lives
// under worktree/, which discovery excludes).
func fleetRepoPaths(deps testStatsDeps) (map[string]string, error) {
	id, err := deps.identity()
	if err != nil {
		return nil, fmt.Errorf("test-stats: resolve workspace: %w", err)
	}
	paths, err := deps.fleetDirs(id.FleetRoot)
	if err != nil {
		return nil, fmt.Errorf("test-stats: discover fleet: %w", err)
	}
	paths = append(paths, id.PrimaryRoot)
	byName := map[string]string{}
	for _, p := range paths {
		if p != "" {
			byName[filepath.Base(p)] = p
		}
	}
	return byName, nil
}

// buildTestStatsReport measures every fleet repository's suite (or the one
// flags.Repo names) since flags.Since; the landing plan (#475) reads its
// cost basis from the same report.
func buildTestStatsReport(flags testStatsFlags, deps testStatsDeps) (testStatsReport, error) {
	if deps.now == nil || deps.identity == nil || deps.fleetDirs == nil || deps.commonDir == nil {
		return testStatsReport{}, fmt.Errorf("test-stats: missing command dependency")
	}
	since, err := parseTestStatsSince(flags.Since, deps.now().UTC())
	if err != nil {
		return testStatsReport{}, err
	}
	byName, err := fleetRepoPaths(deps)
	if err != nil {
		return testStatsReport{}, err
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		if flags.Repo == "" || flags.Repo == name {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	report := testStatsReport{Since: since, Repos: []RepoStats{}, Warnings: []string{}}
	for _, name := range names {
		common, err := deps.commonDir(byName[name])
		if err != nil || common == "" {
			// The fleet-name filter is intentionally looser than Git discovery.
			// Non-repository siblings contribute no measurements.
			continue
		}
		root := filepath.Join(common, "sdlc", "test-stats")
		attempts, warnings, err := readSuiteAttempts(root, testStatsRecordLimit)
		if err != nil {
			report.Warnings = append(report.Warnings, name+": "+err.Error())
			continue
		}
		for _, warning := range warnings {
			report.Warnings = append(report.Warnings, name+": "+warning)
		}
		known := map[string]bool{}
		for i := range attempts {
			// A different local check at this head does not cover suite.log.
			if isSuiteRun(attempts[i]) {
				known[attempts[i].SHA] = true
			}
			attempts[i].Repo = name
		}
		legacy, legacyWarnings := readLegacySuiteLogs(common, name, known)
		attempts = append(attempts, legacy...)
		for _, warning := range legacyWarnings {
			report.Warnings = append(report.Warnings, name+": "+warning)
		}
		rows := aggregateSuiteAttempts(attempts, since, name)
		if len(rows) != 0 {
			report.Repos = append(report.Repos, rows[0])
		}
	}
	return report, nil
}

func readLegacySuiteLogs(common, repo string, known map[string]bool) ([]SuiteAttempt, []string) {
	base := filepath.Join(common, "sdlc", "local-checks")
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, []string{"legacy logs: " + err.Error()}
	}
	var out []SuiteAttempt
	var warnings []string
	for _, entry := range entries {
		sha := entry.Name()
		if !entry.IsDir() || known[sha] || !suiteHex.MatchString(sha) || len(sha) != 40 && len(sha) != 64 {
			continue
		}
		path := filepath.Join(base, sha, "suite.log")
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			warnings = append(warnings, path+": "+err.Error())
			continue
		}
		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			warnings = append(warnings, path+": "+err.Error())
			continue
		}
		data, err := io.ReadAll(io.LimitReader(file, testStatsLogLimit+1))
		_ = file.Close()
		if err != nil || len(data) > testStatsLogLimit {
			warnings = append(warnings, path+": unreadable or over 64 MiB; partial legacy run omitted")
			continue
		}
		facts := parseSuiteLogFacts(string(data))
		exit := 0
		if facts.Failed {
			exit = 1
		}
		for _, warning := range facts.Warnings {
			warnings = append(warnings, path+": "+warning)
		}
		out = append(out, SuiteAttempt{Repo: repo, SHA: sha, Check: suiteCheckName, StartedAt: info.ModTime().UTC(), Exit: exit, WallSeconds: facts.WallSeconds, Failures: facts.Failures, Units: facts.Units, Legacy: true})
	}
	return out, warnings
}

func renderTestStatsText(out io.Writer, report testStatsReport) error {
	if _, err := fmt.Fprintf(out, "Suite attempts since %s (local records; historical logs are partial)\n", report.Since.Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "repo      runs  wall p50/p90  queue p50/p90  confirmed flakes  coverage"); err != nil {
		return err
	}
	for _, s := range report.Repos {
		if _, err := fmt.Fprintf(out, "%-9s %4d  %-13s %-14s %d/%d  wall %d/%d queue %d/%d\n", s.Repo, s.Runs, formatTestStatsQuantiles(s.Wall), formatTestStatsQuantiles(s.Queue), s.ConfirmedFlakeAttempts, s.FlakeDenominator, s.WallCoverage, s.Runs, s.QueueCoverage, s.Runs); err != nil {
			return err
		}
		for _, unit := range s.Slow {
			if _, err := fmt.Fprintf(out, "  %s slow: %s %s %.1fs (n=%d)\n", s.Repo, unit.Kind, unit.Name, unit.P90, unit.Samples); err != nil {
				return err
			}
		}
		for _, flake := range s.Flakes {
			if _, err := fmt.Fprintf(out, "  flake %s: %s on %s\n", flake.Kind, flake.Name, abbrevSHA(flake.SHA)); err != nil {
				return err
			}
		}
	}
	for _, warning := range report.Warnings {
		if _, err := fmt.Fprintf(out, "warning: %s\n", warning); err != nil {
			return err
		}
	}
	return nil
}

func formatTestStatsQuantiles(q *SuiteQuantiles) string {
	if q == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.1fs/%.1fs", q.P50, q.P90)
}
