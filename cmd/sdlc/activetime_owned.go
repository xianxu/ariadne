package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// computeActiveTime routes tracked repositories through the same authority as
// actual and close. Explicit manual inputs narrow that authority; they cannot
// replace it. No-tracker repositories retain the standalone engine contract.
func computeActiveTime(opts activetime.Options) (activetime.Result, []string, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = ownedBranchRequestContext(ctx)
	root, err := ownedActualGit(ctx, manualHomePath(opts.GitRepo), "--no-optional-locks", "rev-parse", "--show-toplevel")
	if err != nil {
		return activetime.Result{}, nil, err
	}
	tracked, err := localTrackerDeclared(root)
	if err != nil {
		return activetime.Result{}, nil, err
	}
	if !tracked {
		r, e := activetime.Compute(opts)
		return r, nil, e
	}
	dirs := make([]string, 0, len(opts.Dirs))
	for _, dir := range opts.Dirs {
		p, e := workspace.CanonicalPath(manualHomePath(dir))
		if e != nil {
			return activetime.Result{}, nil, e
		}
		dirs = append(dirs, p)
	}
	result := activetime.Result{Status: activetime.EmptyWindow, PerIssue: map[string]float64{}}
	var notes []string
	notes = append(notes, "tracked repository: branch/claim attribution by default; --dir, --since and --until only narrow verified source windows")
	type selectedWindow struct {
		start, end time.Time
		issue      string
	}
	selected := map[string][]selectedWindow{}
	seen := map[string]bool{}
	var codeHead, trackerHead string
	// Resolve each issue before allocation. The source intervals are checked
	// for cross-issue overlap; they are never unioned into a run whose later
	// commit could claim another branch's activity.
	type prepared struct {
		id      string
		options activetime.Options
	}
	var batches []prepared
	for _, rawID := range opts.Issues {
		n, e := strconv.Atoi(rawID)
		if e != nil || n <= 0 {
			return activetime.Result{}, nil, fmt.Errorf("invalid issue %q", rawID)
		}
		id := strconv.Itoa(n)
		if seen[id] {
			continue
		}
		seen[id] = true
		inputs, e := resolveOwnedActualInputs(ctx, root, "", id, "")
		if e != nil {
			return activetime.Result{}, nil, e
		}
		if inputs == nil {
			return activetime.Result{}, nil, fmt.Errorf("tracker disappeared during active-time measurement")
		}
		m := inputs.Snapshot
		if !inputs.Replay && codeHead != "" && (codeHead != m.CodeHead || trackerHead != m.TrackerHead) {
			return activetime.Result{}, nil, fmt.Errorf("active-time issue authorities differ; inspect completed snapshots or retry a moving tracker")
		}
		if !inputs.Replay {
			codeHead, trackerHead = m.CodeHead, m.TrackerHead
		}
		if inputs.LegacyPrefix != nil {
			return activetime.Result{}, nil, fmt.Errorf("active-time issue #%s has a legacy prefix; inspect its separately labelled sdlc actual instead", id)
		}
		if !inputs.Replay {
			inputs.Protected, e = loadAdoptedExclusions(ctx, root, id, "", inputs)
			if e != nil {
				return activetime.Result{}, nil, e
			}
		}
		if e := validateProtectedSlotIdentities(inputs); e != nil {
			return activetime.Result{}, nil, e
		}
		options := opts
		options.Context = ctx
		options.GitRepo = root
		options.Dirs = nil
		options.Files = nil
		options.ActivityWindows = inputs.Windows
		options.DispatchWindows = inputs.DispatchWindows
		options.SourcePrefixes = inputs.Prefixes
		options.SourceSlots = inputs.SourceSlots
		options.ProtectedIntervals = inputs.Protected
		options.ExtraRefs = inputs.TrackerRefs
		options.Scope = inputs.Scope
		options.SinceISO, e = narrowManualTime(m.Since, opts.SinceISO, true)
		if e != nil {
			return activetime.Result{}, nil, e
		}
		options.UntilISO, e = narrowManualTime(m.Until, opts.UntilISO, false)
		if e != nil {
			return activetime.Result{}, nil, e
		}
		if opts.Scope.BranchPoint != "" {
			pin, e := ownedActualGit(ctx, root, "rev-parse", "--verify", opts.Scope.BranchPoint+"^{commit}")
			if e != nil {
				return activetime.Result{}, nil, e
			}
			if _, e = ownedActualGit(ctx, root, "merge-base", "--is-ancestor", pin, m.CodeHead); e != nil {
				return activetime.Result{}, nil, fmt.Errorf("explicit branch point is outside the measured history")
			}
			if inputs.Scope.BranchPoint != "" {
				if _, e = ownedActualGit(ctx, root, "merge-base", "--is-ancestor", inputs.Scope.BranchPoint, pin); e != nil {
					return activetime.Result{}, nil, fmt.Errorf("explicit branch point would widen owned history")
				}
			}
			options.Scope.BranchPoint = pin
		}
		for _, source := range inputs.Files {
			path, e := workspace.CanonicalPath(source)
			if e != nil {
				return activetime.Result{}, nil, e
			}
			allowed := false
			for _, dir := range dirs {
				rel, e := filepath.Rel(dir, path)
				if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
			options.Files = append(options.Files, source)
			for _, w := range inputs.Windows[source] {
				selected[path] = append(selected[path], selectedWindow{w.Start, w.End, id})
			}
		}
		for _, q := range m.QueueWaits {
			start, _ := time.Parse(time.RFC3339Nano, q.Start)
			end, _ := time.Parse(time.RFC3339Nano, q.End)
			options.QueueIntervals = append(options.QueueIntervals, activetime.QueueInterval{Start: start, End: end, Slot: q.Slot, AttemptID: q.AttemptID})
		}
		notes = append(notes, inputs.Warnings...)
		batches = append(batches, prepared{id, options})
	}
	for source, windows := range selected {
		sort.Slice(windows, func(i, j int) bool { return windows[i].start.Before(windows[j].start) })
		for i := 1; i < len(windows); i++ {
			if windows[i].start.Before(windows[i-1].end) && windows[i].issue != windows[i-1].issue {
				return activetime.Result{}, nil, fmt.Errorf("active-time retained inputs overlap for issues %s/%s in %s; inspect immutable adopted evidence", windows[i-1].issue, windows[i].issue, source)
			}
		}
	}
	for _, batch := range batches {
		r, e := activetime.Compute(batch.options)
		if e != nil {
			return activetime.Result{}, nil, e
		}
		if r.Status == activetime.TelemetryGap {
			result.Status = activetime.TelemetryGap
		} else if r.Status == activetime.Measured && result.Status != activetime.TelemetryGap {
			result.Status = activetime.Measured
		}
		result.NumEvents += r.NumEvents
		result.NumCommits += r.NumCommits
		result.Warnings = append(result.Warnings, r.Warnings...)
		for _, segment := range r.Segments {
			// Preserve the weight calculation, but a different issue's share
			// cannot become credited time outside that issue's selected partition.
			own := segment.Alloc[batch.id]
			segment.Alloc = map[string]float64{batch.id: own}
			result.PerIssue[batch.id] += own
			if remainder := segment.Active - own; remainder > 1e-9 {
				segment.Alloc[activetime.UnattributedKey] = remainder
				result.PerIssue[activetime.UnattributedKey] += remainder
			}
			result.Segments = append(result.Segments, segment)
			result.TotalActive += segment.Active
		}
	}
	sort.SliceStable(result.Segments, func(i, j int) bool { return result.Segments[i].Start.Before(result.Segments[j].Start) })
	return result, notes, nil
}

func manualHomePath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, e := os.UserHomeDir(); e == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~"), "/"))
		}
	}
	return path
}

func narrowManualTime(owned, explicit string, left bool) (string, error) {
	bound, e := time.Parse(time.RFC3339Nano, owned)
	if e != nil {
		return "", e
	}
	if explicit == "" {
		return owned, nil
	}
	requested, e := time.Parse(time.RFC3339Nano, explicit)
	if e != nil {
		return "", e
	}
	if left && requested.After(bound) || !left && requested.Before(bound) {
		return explicit, nil
	}
	return owned, nil
}
