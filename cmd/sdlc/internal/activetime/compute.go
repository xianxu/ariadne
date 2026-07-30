package activetime

import (
	"fmt"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueref"
)

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
	// statusInvalid is the zero value: an unset / never-computed Result. It
	// exists so a forgotten Result{} does NOT read as "measured 0 hours" — the
	// exact silent-zero footgun this package exists to prevent (#68). Compute
	// always sets a real Status and pairs Result{} with a non-nil error; a caller
	// that checks err first never observes statusInvalid, but shifting Measured
	// off zero removes the trap entirely.
	statusInvalid Status = iota
	Measured
	TelemetryGap
	EmptyWindow
)

// Options configures a Compute run. PrefixWeight is a *float64 (nil = unset, fall
// back to CommitWeight) — NOT a float sentinel — so an explicit
// --prefix-commit-weight 0 is honored, matching the Python original's
// --prefix-commit-weight defaulting (None → commit-weight).
type Options struct {
	Dirs             []string
	Files            []string
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
	Warnings    []AttributionWarning
	NumEvents   int
	NumCommits  int
}

const suspiciousSpanMin = 120.0
const suspiciousShare = 0.50

type AttributionWarning struct {
	Issue  string
	Start  time.Time
	End    time.Time
	Active float64
	Share  float64
	Reason string
}

// Compute is the engine: load transcript events + window commits, then attribute
// active time per the v3 rule. The only IO is the two loaders; everything below
// is pure.
func Compute(opts Options) (Result, error) {
	// The mention scope carries the tracked set plus the LOCAL repo name, so a `pair#127` in
	// transcript prose is not counted as this repo's #127 (ariadne#190). The qualifier comes
	// from GitRepo — the repo whose text is being attributed — for the same reason
	// selfQualifier does on the commit path.
	sc := newMentionScope(selfQualifier(opts.GitRepo), opts.Issues)
	events, spans, err := loadEventsWithFiles(opts.Dirs, opts.Files, sc, opts.IncludeAssistant, opts.SinceISO, opts.UntilISO)
	if err != nil {
		return Result{}, err
	}
	commits, err := loadWindowCommits(opts.GitRepo, opts.SinceISO, opts.UntilISO)
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

	res.Segments = buildSegments(events, commits, spans, opts.CommitWeight, prefixWeight, opts.ThresholdMin)
	for _, s := range res.Segments {
		res.TotalActive += s.Active
		for iss, m := range s.Alloc {
			res.PerIssue[iss] += m
		}
	}
	res.Warnings = attributionWarnings(res.Segments, res.PerIssue)
	res.Warnings = append(res.Warnings, foreignRefWarnings(commits, sc.selfRepo)...)
	res.Status = Measured
	return res, nil
}

// foreignRefWarnings reports cross-repo refs the window contained and this engine deliberately
// did NOT attribute (ariadne#190).
//
// Without it the exclusion is invisible, and "correctly excluded" reads identically to "never
// seen" — which is how the original defect survived: the numbers looked plausible. Since a
// foreign ref is real work on a real issue elsewhere, naming it also tells the operator where
// the missing time went.
//
// Scoped to COMMIT SUBJECTS, which Commit retains. Transcript-side foreign mentions are not
// counted here because the event text is not kept past mention extraction, and threading it
// through purely for a diagnostic would cost more than it informs — the commit subjects are the
// durable, greppable record, and in the motivating case (#187's replay of pair#127) they
// carried the foreign ref too.
func foreignRefWarnings(commits []Commit, selfRepo string) []AttributionWarning {
	counts := map[string]int{}
	var order []string
	for _, c := range commits {
		for _, r := range issueref.Find(c.Subject) {
			if r.IsLocal(selfRepo) {
				continue
			}
			key := r.Qualifier + "#" + r.Num
			if counts[key] == 0 {
				order = append(order, key)
			}
			counts[key]++
		}
	}
	var out []AttributionWarning
	for _, key := range order {
		out = append(out, AttributionWarning{
			Issue:  key,
			Active: 0, Share: 0,
			Reason: fmt.Sprintf("foreign ref ignored — another repo's issue, not attributable here (×%d)", counts[key]),
		})
	}
	return out
}

func attributionWarnings(segs []Segment, perIssue map[string]float64) []AttributionWarning {
	var warnings []AttributionWarning
	for _, s := range segs {
		spanMin := s.End.Sub(s.Start).Minutes()
		for iss, mins := range s.Alloc {
			total := perIssue[iss]
			if total <= 0 || mins <= 0 {
				continue
			}
			share := mins / total
			if spanMin > suspiciousSpanMin && share > suspiciousShare {
				warnings = append(warnings, AttributionWarning{
					Issue:  iss,
					Start:  s.Start,
					End:    s.End,
					Active: mins, Share: share,
					Reason: "dominant long attribution segment",
				})
			}
			if s.Commit == nil {
				reason := "mention fallback without issue commit boundary"
				if iss == UnattributedKey {
					reason = "unattributed fallback without issue commit boundary"
				}
				warnings = append(warnings, AttributionWarning{
					Issue:  iss,
					Start:  s.Start,
					End:    s.End,
					Active: mins, Share: share,
					Reason: reason,
				})
			}
		}
	}
	return warnings
}
