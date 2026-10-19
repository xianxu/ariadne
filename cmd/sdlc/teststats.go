package main

import (
	"sort"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// SuiteAttempt is a completed suite invocation, including the exact status
// evidence #380 needs if status publication fails. Legacy is reader-only:
// historical per-SHA logs cannot establish retry or queue history.
type SuiteAttempt struct {
	Version      int             `json:"version"`
	ID           string          `json:"id"`
	Repo         string          `json:"repo"`
	SHA          string          `json:"sha"`
	Check        string          `json:"check"`
	Slot         string          `json:"slot,omitempty"`
	Agent        *string         `json:"agent,omitempty"`
	Model        *string         `json:"model,omitempty"`
	Effort       *string         `json:"effort,omitempty"`
	StartedAt    time.Time       `json:"started_at"`
	QueuedAt     *time.Time      `json:"queued_at,omitempty"`
	AcquiredAt   *time.Time      `json:"acquired_at,omitempty"`
	FinishedAt   time.Time       `json:"finished_at"`
	MainBaseline bool            `json:"main_baseline"`
	QueueSeconds *float64        `json:"queue_seconds"`
	WallSeconds  *float64        `json:"wall_seconds"`
	Exit         int             `json:"exit"`
	TimedOut     bool            `json:"timed_out"`
	LogSHA256    string          `json:"log_sha256"`
	LogPath      string          `json:"log_path"`
	Failures     []string        `json:"failures,omitempty"`
	Units        []SuiteDuration `json:"units,omitempty"`
	// Version 2 (#394): the gate's outcome, the registered tests it ran on,
	// and, for a targeted rerun, the failed attempt it reran.
	Outcome     string         `json:"outcome,omitempty"`
	KnownFlakes []flakes.Entry `json:"known_flakes,omitempty"`
	RerunOf     string         `json:"rerun_of,omitempty"`
	Legacy      bool           `json:"-"`
}

type SuiteQuantiles struct {
	P50     float64 `json:"p50_seconds"`
	P90     float64 `json:"p90_seconds"`
	Samples int     `json:"samples"`
}

type SuiteSlowUnit struct {
	Kind    string  `json:"kind"`
	Name    string  `json:"name"`
	P90     float64 `json:"p90_seconds"`
	Samples int     `json:"samples"`
}

type SuiteFlake struct {
	Name     string    `json:"name"`
	SHA      string    `json:"sha"`
	Kind     string    `json:"kind"`
	FailedAt time.Time `json:"failed_at"`
}

type RepoStats struct {
	Repo                   string          `json:"repo"`
	Runs                   int             `json:"runs"`
	FullyRecorded          int             `json:"fully_recorded"`
	WallCoverage           int             `json:"wall_coverage"`
	QueueCoverage          int             `json:"queue_coverage"`
	FlakeDenominator       int             `json:"flake_denominator"`
	ConfirmedFlakeAttempts int             `json:"confirmed_flake_attempts"`
	Wall                   *SuiteQuantiles `json:"wall"`
	Queue                  *SuiteQuantiles `json:"queue"`
	Slow                   []SuiteSlowUnit `json:"slow,omitempty"`
	Flakes                 []SuiteFlake    `json:"flakes,omitempty"`
}

func suiteQuantiles(samples []float64) *SuiteQuantiles {
	if len(samples) == 0 {
		return nil
	}
	sort.Float64s(samples)
	nearest := func(p int) float64 {
		// ceil(p*n/100)-1, clamped to the first sample.
		i := (p*len(samples)+99)/100 - 1
		if i < 0 {
			i = 0
		}
		return samples[i]
	}
	return &SuiteQuantiles{P50: nearest(50), P90: nearest(90), Samples: len(samples)}
}

// aggregateSuiteAttempts computes stats from validated records plus explicitly
// partial legacy entries. The caller owns persisted-record validation.
func aggregateSuiteAttempts(attempts []SuiteAttempt, since time.Time, repo string) []RepoStats {
	byRepo := make(map[string][]SuiteAttempt)
	for _, a := range attempts {
		if a.Check != suiteCheckName || suiteAttemptCompletedAt(a).Before(since) || (repo != "" && a.Repo != repo) {
			continue
		}
		byRepo[a.Repo] = append(byRepo[a.Repo], a)
	}
	names := make([]string, 0, len(byRepo))
	for name := range byRepo {
		names = append(names, name)
	}
	sort.Strings(names)
	stats := make([]RepoStats, 0, len(names))
	for _, name := range names {
		rows := byRepo[name]
		sort.Slice(rows, func(i, j int) bool {
			iAt, jAt := suiteAttemptCompletedAt(rows[i]), suiteAttemptCompletedAt(rows[j])
			if iAt.Equal(jAt) {
				return rows[i].ID < rows[j].ID
			}
			return iAt.Before(jAt)
		})
		s := RepoStats{Repo: name}
		for _, a := range rows {
			if a.RerunOf == "" {
				s.Runs++
			}
		}
		var wall, queue []float64
		evidence := flakeAttempts(rows)
		unitTimes := map[string][]float64{}
		units := map[string]SuiteDuration{}
		for i, a := range rows {
			if a.RerunOf != "" {
				// A known-flake rerun (#394) runs a few named tests, not the
				// suite: it is flake evidence (above), never a suite run's
				// count, wall, queue or slow unit.
				continue
			}
			if !a.Legacy {
				s.FullyRecorded++
				s.FlakeDenominator++
			}
			if a.WallSeconds != nil {
				s.WallCoverage++
				wall = append(wall, *a.WallSeconds)
			}
			if a.QueueSeconds != nil {
				s.QueueCoverage++
				queue = append(queue, *a.QueueSeconds)
			}
			for _, u := range a.Units {
				key := u.Kind + "\x00" + u.Name
				unitTimes[key] = append(unitTimes[key], u.Seconds)
				units[key] = u
			}
			if a.Legacy || a.Exit == 0 || len(a.Failures) == 0 {
				continue
			}
			kind := ""
			if a.MainBaseline {
				kind = "suspected-main-baseline"
			} else if _, ok := flakes.Confirmed(evidence, i); ok {
				kind = "confirmed"
			}
			if kind == "" {
				continue
			}
			if kind == "confirmed" {
				s.ConfirmedFlakeAttempts++
			}
			for _, failure := range a.Failures {
				s.Flakes = append(s.Flakes, SuiteFlake{Name: failure, SHA: a.SHA, Kind: kind, FailedAt: a.StartedAt})
			}
		}
		s.Wall, s.Queue = suiteQuantiles(wall), suiteQuantiles(queue)
		for key, samples := range unitTimes {
			u := units[key]
			s.Slow = append(s.Slow, SuiteSlowUnit{Kind: u.Kind, Name: u.Name, P90: suiteQuantiles(samples).P90, Samples: len(samples)})
		}
		sort.Slice(s.Slow, func(i, j int) bool {
			if s.Slow[i].P90 != s.Slow[j].P90 {
				return s.Slow[i].P90 > s.Slow[j].P90
			}
			if s.Slow[i].Kind != s.Slow[j].Kind {
				return s.Slow[i].Kind < s.Slow[j].Kind
			}
			return s.Slow[i].Name < s.Slow[j].Name
		})
		if len(s.Slow) > 5 {
			s.Slow = s.Slow[:5]
		}
		sort.Slice(s.Flakes, func(i, j int) bool {
			if s.Flakes[i].Kind != s.Flakes[j].Kind {
				return s.Flakes[i].Kind == "suspected-main-baseline"
			}
			if s.Flakes[i].FailedAt.Equal(s.Flakes[j].FailedAt) {
				return s.Flakes[i].Name < s.Flakes[j].Name
			}
			return s.Flakes[i].FailedAt.Before(s.Flakes[j].FailedAt)
		})
		stats = append(stats, s)
	}
	return stats
}

func suiteAttemptCompletedAt(a SuiteAttempt) time.Time {
	if !a.FinishedAt.IsZero() {
		return a.FinishedAt
	}
	return a.StartedAt // historical logs and older in-memory fixtures
}

// flakeAttempt is a recorded attempt as the known-flake rules read it (#394).
func flakeAttempt(a SuiteAttempt) flakes.Attempt {
	return flakes.Attempt{ID: a.ID, SHA: a.SHA, Passed: a.Exit == 0 && !a.TimedOut, Baseline: a.MainBaseline,
		Legacy: a.Legacy, Failures: a.Failures, Kind: flakes.Kind(a.Outcome), FinishedAt: suiteAttemptCompletedAt(a)}
}

func flakeAttempts(as []SuiteAttempt) []flakes.Attempt {
	out := make([]flakes.Attempt, len(as))
	for i, a := range as {
		out[i] = flakeAttempt(a)
	}
	return out
}
