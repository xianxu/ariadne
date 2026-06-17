package activetime

import "time"

// Status mirrors active-time-v3.py's exit-code contract minus the CLI-layer
// misinvoke (exit 2, validated before Compute runs):
//
//	Measured     — events present; per-issue totals produced (v3 exit 0 + table).
//	TelemetryGap — commits in window but 0 transcript events (v3 exit 3): the
//	               work's transcripts aren't under the given dirs or aged out;
//	               callers must NOT read 0 as a measured value (#68).
//	EmptyWindow  — no events and no commits — nothing to measure (v3 exit 0).
type Status int

const (
	Measured Status = iota
	TelemetryGap
	EmptyWindow
)

// Options configures a Compute run. PrefixWeight is a *float64 (nil = unset, fall
// back to CommitWeight) — NOT a float sentinel — so an explicit
// --prefix-commit-weight 0 is honored, matching active-time-v3.py:282.
type Options struct {
	Dirs             []string
	GitRepo          string
	SinceISO         string
	UntilISO         string
	Issues           []string
	CommitWeight     float64
	PrefixWeight     *float64
	ThresholdMin     int
	IncludeAssistant bool
}

// Result is the structured outcome of a Compute run — the single value consumed
// by both computeActual (reads PerIssue[issue]) and the active-time renderer
// (prints Segments + PerIssue). PerIssue and TotalActive are in MINUTES; callers
// divide by 60 for hours.
type Result struct {
	Status      Status
	PerIssue    map[string]float64
	TotalActive float64
	Segments    []Segment
	NumEvents   int
	NumCommits  int
}

// Compute is the engine: load transcript events + window commits, then attribute
// active time per the v3 rule. Mirrors active-time-v3.py main()'s branching
// (minus printing). The only IO is the two loaders; everything below is pure.
func Compute(opts Options) (Result, error) {
	pat := issuePattern(opts.Issues)
	events, err := loadEvents(opts.Dirs, pat, opts.IncludeAssistant, opts.SinceISO, opts.UntilISO)
	if err != nil {
		return Result{}, err
	}
	commits, err := loadWindowCommits(opts.GitRepo, opts.SinceISO, opts.UntilISO, pat)
	if err != nil {
		return Result{}, err
	}

	res := Result{
		PerIssue:   map[string]float64{},
		NumEvents:  len(events),
		NumCommits: len(commits),
	}

	if len(events) == 0 {
		if len(commits) > 0 {
			res.Status = TelemetryGap
		} else {
			res.Status = EmptyWindow
		}
		return res, nil
	}

	prefixWeight := opts.CommitWeight
	if opts.PrefixWeight != nil {
		prefixWeight = *opts.PrefixWeight
	}

	if len(commits) == 0 {
		// No commit signal in the window → whole-window mention attribution
		// (active-time-v3.py lines 309–319).
		times := make([]time.Time, len(events))
		mentions := map[string]int{}
		for i, e := range events {
			times[i] = e.Time
			for iss, n := range e.Mentions {
				mentions[iss] += n
			}
		}
		active := activeMinutes(times, opts.ThresholdMin)
		res.TotalActive = active
		res.PerIssue = attributeSegment(active, nil, mentions, opts.CommitWeight)
		res.Status = Measured
		return res, nil
	}

	res.Segments = buildSegments(events, commits, opts.CommitWeight, prefixWeight, opts.ThresholdMin)
	for _, s := range res.Segments {
		res.TotalActive += s.Active
		for iss, m := range s.Alloc {
			res.PerIssue[iss] += m
		}
	}
	res.Status = Measured
	return res, nil
}
