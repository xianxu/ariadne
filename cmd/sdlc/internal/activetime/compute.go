package activetime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	// Context optionally cancels IO and parsing. Nil preserves legacy runners.
	Context context.Context
	Dirs    []string
	Files   []string
	GitRepo string
	// ExtraRefs are histories read beside HEAD, e.g. the issue tracker (#252).
	ExtraRefs []string
	// Scope keeps another session's commits, integrated from main, from
	// bounding this branch's segments (#270). Zero = every commit counts.
	Scope            Scope
	SinceISO         string
	UntilISO         string
	Issues           []string
	CommitWeight     float64
	PrefixWeight     *float64
	ThresholdMin     int
	IncludeAssistant bool
	// ActivityWindows, when non-nil, admits only confirmed ownership activity
	// for each transcript source. The gap/span calculation runs before clipping.
	ActivityWindows map[string][]ActivityWindow
	SourcePrefixes  map[string]SourcePrefix
	// QueueIntervals are validated exact attempt evidence; Compute applies each
	// wait to admitted activity sources of the attempt's owning slot.
	QueueIntervals []QueueInterval
	SourceSlots    map[string]string
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
	// IssueCommitTimes are the author times of the window commits whose lead
	// names Scope.Issue, oldest first (#234): the evidence for where each
	// boundary's converge ends.
	IssueCommitTimes []time.Time
	// SourceIDs identifies actual contributing runs, excluding old same-cwd
	// sessions with no activity in the admitted sample.
	SourceIDs []string
	// QueueBindingSourceIDs identifies transcripts required to replay selected
	// queue waits, even when subtraction removes all their activity.
	QueueBindingSourceIDs []string
	QueueAttribution      QueueAttribution
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
	if err := contextError(opts.Context); err != nil {
		return Result{}, err
	}
	// The mention scope carries the tracked set plus the LOCAL repo name, so a `pair#127` in
	// transcript prose is not counted as this repo's #127 (ariadne#190). The qualifier comes
	// from GitRepo — the repo whose text is being attributed — for the same reason
	// selfQualifier does on the commit path.
	self, err := selfQualifierContext(opts.Context, opts.GitRepo)
	if err != nil {
		return Result{}, err
	}
	sc := newMentionScope(self, opts.Issues)
	events, spans, err := loadEventsWithFilesContext(opts.Context, opts.Dirs, opts.Files, sc, opts.IncludeAssistant, opts.SinceISO, opts.UntilISO, opts.SourcePrefixes)
	if err != nil {
		return Result{}, err
	}
	commits, err := loadWindowCommitsContext(opts.Context, opts.GitRepo, opts.SinceISO, opts.UntilISO, opts.Scope, opts.ExtraRefs...)
	if err != nil {
		return Result{}, err
	}

	if err := contextError(opts.Context); err != nil {
		return Result{}, err
	}
	res := Result{
		PerIssue:   map[string]float64{},
		NumEvents:  len(events),
		NumCommits: len(commits),
	}
	for _, c := range commits {
		if opts.Scope.Issue != "" && slices.Contains(c.Issues, opts.Scope.Issue) {
			res.IssueCommitTimes = append(res.IssueCommitTimes, c.Time)
		}
	}

	if len(events) == 0 {
		if len(commits) > 0 {
			res.Status = TelemetryGap
		} else {
			res.Status = EmptyWindow
		}
		if err := contextError(opts.Context); err != nil {
			return Result{}, err
		}
		return res, nil
	}

	prefixWeight := opts.CommitWeight
	if opts.PrefixWeight != nil {
		prefixWeight = *opts.PrefixWeight
	}

	spans, err = AdmitTaskSpans(spans, opts.ActivityWindows)
	if err != nil {
		return Result{}, err
	}
	runs, err := ClipActivity(activityRuns(events, spans, opts.ThresholdMin), opts.ActivityWindows)
	if err != nil {
		return Result{}, err
	}
	if len(opts.QueueIntervals) > 0 {
		tools, err := loadToolSpans(opts)
		if err != nil {
			return Result{}, err
		}
		bound := make([]QueueInterval, 0, len(opts.QueueIntervals))
		bindingSources := map[string]bool{}
		for _, wait := range opts.QueueIntervals {
			forSources, err := bindQueueSources(wait, runs, opts.SourceSlots)
			if err != nil {
				return Result{}, err
			}
			for _, b := range forSources {
				for _, tool := range tools {
					if tool.Kind == toolUncertainWork && tool.Source == b.Source && tool.Start.Before(b.End) {
						return Result{}, fmt.Errorf("queue attempt %s overlaps yielded or incomplete work in source %s", b.AttemptID, b.Source)
					}
				}
				bound = append(bound, b)
				if !bindingSources[b.Source] {
					bindingSources[b.Source] = true
					res.QueueBindingSourceIDs = append(res.QueueBindingSourceIDs, b.Source)
				}
			}
		}
		slices.Sort(res.QueueBindingSourceIDs)
		// A subagent dispatch is independent work only within its admitted
		// engagement. It cannot restore any time after release/reclaim.
		for _, sp := range spans {
			tools = append(tools, ToolSpan{Start: sp.Start, End: sp.End, Source: sp.Source, Kind: toolWork})
		}
		runs, res.QueueAttribution = subtractQueueIdle(runs, bound, tools)
	}
	res.Segments = claimActivityRuns(runs, commits, opts.CommitWeight, prefixWeight)
	seenSources := map[string]bool{}
	for _, run := range runs {
		if !seenSources[run.Source] {
			seenSources[run.Source] = true
			res.SourceIDs = append(res.SourceIDs, run.Source)
		}
	}
	slices.Sort(res.SourceIDs)
	for _, s := range res.Segments {
		res.TotalActive += s.Active
		for iss, m := range s.Alloc {
			res.PerIssue[iss] += m
		}
	}
	res.Warnings = attributionWarnings(res.Segments, res.PerIssue)
	res.Warnings = append(res.Warnings, foreignRefWarnings(commits, sc.selfRepo)...)
	res.Status = Measured
	if err := contextError(opts.Context); err != nil {
		return Result{}, err
	}
	return res, nil
}

func loadToolSpans(opts Options) ([]ToolSpan, error) {
	paths := map[string]bool{}
	for _, path := range opts.Files {
		paths[path] = true
	}
	for _, dir := range opts.Dirs {
		matches, err := filepath.Glob(filepath.Join(expandUser(dir), "*.jsonl"))
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			paths[path] = true
		}
	}
	var ordered []string
	for path := range paths {
		ordered = append(ordered, path)
	}
	slices.Sort(ordered)
	var out []ToolSpan
	for _, path := range ordered {
		var pins []SourcePrefix
		if opts.SourcePrefixes != nil {
			pin, ok := opts.SourcePrefixes[path]
			if !ok {
				return nil, fmt.Errorf("source %s lacks pinned queue evidence", path)
			}
			pins = append(pins, pin)
		}
		data, err := readSessionPrefixContext(opts.Context, expandUser(path), pins)
		if err != nil {
			if opts.SourcePrefixes != nil || !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		out = append(out, workSpansFromTranscript(data, path)...)
	}
	return out, nil
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
