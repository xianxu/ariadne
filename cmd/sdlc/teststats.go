package main

import (
	"fmt"
	"os"
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
	// Version 3 binds the nonpassing disposition to immutable recovery proof.
	ReclassificationSHA256 string    `json:"reclassification_sha256,omitempty"`
	ReclassifiedAt         time.Time `json:"-"` // reader projection of the verified audit, never an independent authority
	// Version 4 binds typed runner identities to the declared runner contract.
	RunnerContractSHA256 string `json:"runner_contract_sha256,omitempty"`
	// Reader-only projection of digest-verified raw evidence; never persisted.
	TypedFacts *flakes.UnitFacts `json:"-"`
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
	Check                  string          `json:"check,omitempty"` // "" is the suite; else the test check (#473)
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

// isSuiteRun reports whether a recorded attempt is a run of the suite itself.
// Run counts, wall and queue samples, slow units and saved-pass reuse read
// only these (#394 M2 review): a known-flake rerun is never one.
func isSuiteRun(a SuiteAttempt) bool { return a.Check == suiteCheckName }

// isTestCheck names the checks that run the repo's tests and so carry flake
// evidence and the known-flake rule: the suite and the affected suite (#473
// D4). Their timings stay apart: isSuiteRun still means the full suite.
func isTestCheck(name string) bool { return name == suiteCheckName || name == affectedCheckName }

// isFlakeEvidence reports whether an attempt is same-head flake evidence: a
// suite run, or a known-flake rerun that passed (its tests passed on the head
// where they had failed). Every reader of fail→pass evidence uses this one
// predicate, so a rerun's proof survives its parent's outcome being revised.
func isFlakeEvidence(a SuiteAttempt) bool {
	if a.ReclassificationSHA256 != "" {
		return false
	}
	return isTestCheck(a.Check) || a.Check == suiteRerunCheck && a.Exit == 0 && !a.TimedOut
}

// aggregateSuiteAttempts computes the full suite's stats from validated records
// plus explicitly partial legacy entries. The caller owns persisted-record
// validation.
func aggregateSuiteAttempts(attempts []SuiteAttempt, since time.Time, repo string) []RepoStats {
	return aggregateCheckAttempts(attempts, since, repo, suiteCheckName)
}

// aggregateCheckAttempts is aggregateSuiteAttempts for one test check's runs:
// the affected suite gets rows of its own (#473 D4), so its shorter runs never
// move the suite's numbers. Flake evidence still spans every test run.
func aggregateCheckAttempts(attempts []SuiteAttempt, since time.Time, repo, check string) []RepoStats {
	byRepo := make(map[string][]SuiteAttempt)
	for _, a := range attempts {
		if !isFlakeEvidence(a) || suiteAttemptCompletedAt(a).Before(since) || (repo != "" && a.Repo != repo) {
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
		if check != suiteCheckName {
			s.Check = check
		}
		var wall, queue []float64
		evidence := flakeAttempts(rows)
		unitTimes := map[string][]float64{}
		units := map[string]SuiteDuration{}
		for i, a := range rows {
			if a.Check != check {
				continue // another check's run, or a passing rerun: evidence for Confirmed, not a run of check
			}
			s.Runs++
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
				kind = "confirmed" // a later suite pass, or a passing known-flake rerun (#394)
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
	if a.ReclassificationSHA256 != "" {
		return flakes.Attempt{ID: a.ID, SHA: a.SHA, Kind: flakes.CouldNotRun, FinishedAt: suiteAttemptCompletedAt(a)}
	}
	passed := a.Exit == 0 && !a.TimedOut
	if a.Version == 4 && a.Outcome != "" && !flakes.Kind(a.Outcome).Lands() {
		passed = false
	}
	return flakes.Attempt{ID: a.ID, SHA: a.SHA, Passed: passed, Baseline: a.MainBaseline,
		Legacy: a.Legacy, Failures: a.Failures, Kind: flakes.Kind(a.Outcome), FinishedAt: suiteAttemptCompletedAt(a),
		Subset: a.Check == affectedCheckName, TypedFacts: a.TypedFacts, RunnerContractSHA256: a.RunnerContractSHA256}
}

func flakeAttempts(as []SuiteAttempt) []flakes.Attempt {
	out := make([]flakes.Attempt, len(as))
	for i, a := range as {
		out[i] = flakeAttempt(a)
	}
	return out
}

// withLogSubtests adds each failed attempt's failing subtests, read from its
// saved log (#466), to the attempts whose coverage a registry will judge. A
// log that can't be read adds none: the failure is then its own leaf, which
// a registered subtest never covers (fail closed).
func withLogSubtests(out []flakes.Attempt, rows []SuiteAttempt, subtestsOf func(SuiteAttempt) ([]string, error)) []flakes.Attempt {
	if subtestsOf == nil {
		return out
	}
	byID := map[string]SuiteAttempt{}
	for _, a := range rows {
		byID[a.ID] = a
	}
	for i := range out {
		if out[i].Passed || len(out[i].Failures) == 0 {
			continue
		}
		if subs, err := subtestsOf(byID[out[i].ID]); err == nil {
			out[i].Subtests = subs
		}
	}
	return out
}

// storeSubtests reads a stored attempt's failing subtests from its saved log.
func storeSubtests(store string) func(SuiteAttempt) ([]string, error) {
	return func(a SuiteAttempt) ([]string, error) {
		p := suiteAttemptLogFile(store, a)
		if p == "" {
			return nil, fmt.Errorf("attempt %s kept no log", a.ID)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return parseSuiteLogFacts(string(data)).Subtests, nil
	}
}

// withTypedLogFacts hydrates only supported typed attempts. A failed read or
// digest check leaves proof unavailable; exit status cannot replace it.
func withTypedLogFacts(out []flakes.Attempt, rows []SuiteAttempt, load func(SuiteAttempt) (flakes.UnitFacts, error)) []flakes.Attempt {
	byID := make(map[string]SuiteAttempt, len(rows))
	for _, a := range rows {
		byID[a.ID] = a
	}
	for i := range out {
		a, ok := byID[out[i].ID]
		if !ok || a.Version != 4 {
			continue
		}
		out[i].TypedFacts = nil
		if load != nil {
			if facts, err := load(a); err == nil {
				out[i].TypedFacts = &facts
			}
		}
	}
	return out
}

// withSuiteTypedFacts feeds the same reader-only facts to the stats projection.
func withSuiteTypedFacts(rows []SuiteAttempt, load func(SuiteAttempt) (flakes.UnitFacts, error)) []SuiteAttempt {
	out := append([]SuiteAttempt(nil), rows...)
	for i := range out {
		if out[i].Version != 4 {
			continue
		}
		out[i].TypedFacts = nil
		if load != nil {
			if facts, err := load(out[i]); err == nil {
				out[i].TypedFacts = &facts
			}
		}
	}
	return out
}

// storeTypedFacts reads bounded regular-file logs and verifies the immutable
// attempt digest before interpreting any unit/case event.
func storeTypedFacts(store string) func(SuiteAttempt) (flakes.UnitFacts, error) {
	return func(a SuiteAttempt) (flakes.UnitFacts, error) {
		if a.Version != 4 {
			return flakes.UnitFacts{}, fmt.Errorf("attempt %s has no typed evidence", a.ID)
		}
		data, err := readResumeStoredLog(store, a)
		if err != nil {
			return flakes.UnitFacts{}, err
		}
		if a.LogSHA256 == "" || resumeRecordDigest(data) != a.LogSHA256 {
			return flakes.UnitFacts{}, fmt.Errorf("attempt %s log digest mismatch", a.ID)
		}
		return flakes.ParseUnitEvidence(string(data)), nil
	}
}
