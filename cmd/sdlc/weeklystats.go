package main

import (
	"errors"
	"sort"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

const weeklySnapshotSchemaVersion = 2

// WeeklyWindow is a fixed, half-open seven-day interval in UTC.
type WeeklyWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func newWeeklyWindow(end, now time.Time) (WeeklyWindow, error) {
	_, offset := end.Zone()
	if end.IsZero() || offset != 0 {
		return WeeklyWindow{}, errors.New("weekly end must be a UTC instant")
	}
	if end.After(now) {
		return WeeklyWindow{}, errors.New("weekly end cannot be in the future")
	}
	end = end.UTC()
	return WeeklyWindow{Start: end.Add(-7 * 24 * time.Hour), End: end}, nil
}

func (w WeeklyWindow) Contains(at time.Time) bool {
	return !at.Before(w.Start) && at.Before(w.End)
}

// WeeklySnapshot is the stable, versioned brain artifact. Later schema changes
// increment SchemaVersion rather than silently changing historical meaning.
type WeeklySnapshot struct {
	SchemaVersion int                    `json:"schema_version"`
	Window        WeeklyWindow           `json:"window"`
	Repos         []WeeklyRepoStats      `json:"repos"`
	Resources     *WeeklyResourceSection `json:"resources"`
}

// WeeklyRepoStats keeps physical in-window suite measurements separate from
// cohort attribution, whose attempts may straddle the week's boundaries.
type WeeklyRepoStats struct {
	Repo                    string                 `json:"repo"`
	Headline                WeeklyHeadline         `json:"headline"`
	Source                  WeeklySource           `json:"source"`
	Suite                   RepoStats              `json:"suite"`
	SlowPackages            []SuiteSlowUnit        `json:"slow_packages"`
	PackageDurationCoverage int                    `json:"package_duration_coverage"`
	FailedTests             []SuiteFlake           `json:"failed_tests"`
	LandedIssues            int                    `json:"landed_issues"`
	AttributedAttempts      int                    `json:"attributed_attempts"`
	AttributedPassed        int                    `json:"attributed_passed"`
	AttributedFailed        int                    `json:"attributed_failed"`
	AttributedRepeats       int                    `json:"attributed_repeats"`
	UnassignableAttempts    int                    `json:"unassignable_attempts"`
	OutsideWeekAttributed   int                    `json:"outside_week_attributed"`
	InWeekOutsideCohort     int                    `json:"in_week_outside_cohort"`
	PreEnqueue              int                    `json:"pre_enqueue"`
	AttemptsPerLandedIssue  *float64               `json:"attempts_per_landed_issue"`
	UnknownAttempts         []weeklyUnknownAttempt `json:"unknown_attempts"`
	CoverageUnknown         []string               `json:"coverage_unknown,omitempty"`
	// TrainFilter is the week's train measurements (#561), from the same
	// fold and formatter as landing status. A snapshot written before it has
	// none: unmeasured, never a healthy zero.
	TrainFilter *landingline.TrainFilterStats `json:"train_filter,omitempty"`
}

// WeeklyHeadline makes the metric's evidence coverage visible beside its value.
// Observed landings are a lower bound when source coverage is incomplete.
type WeeklyHeadline struct {
	Text                   string                `json:"text"`
	AttemptsPerLandedIssue *float64              `json:"attempts_per_landed_issue"`
	ObservedLandedIssues   int                   `json:"observed_landed_issues"`
	CoveredLandedIssues    int                   `json:"covered_landed_issues"`
	UnknownLandedIssues    int                   `json:"unknown_landed_issues"`
	TopUnknownReasons      []WeeklyUnknownReason `json:"top_unknown_reasons"`
}

type WeeklyUnknownReason struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type WeeklySource struct {
	MainRef        string   `json:"main_ref,omitempty"`
	TrackerRef     string   `json:"tracker_ref,omitempty"`
	AttemptRecords int      `json:"attempt_records"`
	LandingEvents  int      `json:"landing_events"`
	Warnings       []string `json:"warnings"`
}

// aggregateWeeklyRepo measures completed physical attempts in the window.
// It retains unknown failed-test names instead of silently dropping them as
// the interactive test-stats display does. Classification may use earlier
// retained records and later records before the fixed window end.
func aggregateWeeklyRepo(repo string, attempts []SuiteAttempt, window WeeklyWindow) WeeklyRepoStats {
	week := make([]SuiteAttempt, 0)
	classifiable := make([]SuiteAttempt, 0)
	for _, a := range attempts {
		if a.Repo != repo || !isSuiteRun(a) || !suiteAttemptCompletedAt(a).Before(window.End) {
			continue
		}
		classifiable = append(classifiable, a)
		if window.Contains(suiteAttemptCompletedAt(a)) {
			week = append(week, a)
		}
	}
	result := WeeklyRepoStats{Repo: repo, Suite: RepoStats{Repo: repo}, FailedTests: []SuiteFlake{}}
	if stats := aggregateSuiteAttempts(week, window.Start, repo); len(stats) > 0 {
		result.Suite = stats[0]
	}
	packageTimes := map[string][]float64{}
	for _, a := range week {
		hasPackage := false
		for _, unit := range a.Units {
			if unit.Kind != "package" {
				continue
			}
			packageTimes[unit.Name] = append(packageTimes[unit.Name], unit.Seconds)
			hasPackage = true
		}
		if hasPackage {
			result.PackageDurationCoverage++
		}
	}
	for name, samples := range packageTimes {
		result.SlowPackages = append(result.SlowPackages, SuiteSlowUnit{Kind: "package", Name: name, P90: suiteQuantiles(samples).P90, Samples: len(samples)})
	}
	sort.Slice(result.SlowPackages, func(i, j int) bool {
		if result.SlowPackages[i].P90 != result.SlowPackages[j].P90 {
			return result.SlowPackages[i].P90 > result.SlowPackages[j].P90
		}
		return result.SlowPackages[i].Name < result.SlowPackages[j].Name
	})
	if len(result.SlowPackages) > 5 {
		result.SlowPackages = result.SlowPackages[:5]
	}
	classes := make(map[SuiteFlake]string)
	if stats := aggregateSuiteAttempts(classifiable, time.Time{}, repo); len(stats) > 0 {
		for _, f := range stats[0].Flakes {
			classes[SuiteFlake{Name: f.Name, SHA: f.SHA, FailedAt: f.FailedAt}] = f.Kind
		}
	}
	for _, a := range week {
		if a.Legacy || a.Exit == 0 {
			continue
		}
		for _, name := range a.Failures {
			key := SuiteFlake{Name: name, SHA: a.SHA, FailedAt: a.StartedAt}
			kind := classes[key]
			if kind == "" {
				kind = "unknown"
			}
			result.FailedTests = append(result.FailedTests, SuiteFlake{Name: name, SHA: a.SHA, FailedAt: a.StartedAt, Kind: kind})
		}
	}
	sort.Slice(result.FailedTests, func(i, j int) bool {
		a, b := result.FailedTests[i], result.FailedTests[j]
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if !a.FailedAt.Equal(b.FailedAt) {
			return a.FailedAt.Before(b.FailedAt)
		}
		if a.SHA != b.SHA {
			return a.SHA < b.SHA
		}
		return a.Kind < b.Kind
	})
	return result
}
