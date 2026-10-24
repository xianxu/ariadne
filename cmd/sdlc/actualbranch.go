package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// ownedBranchCapture is a bounded observation, never a replacement for saved
// measurement windows. A failed capture is unknown and can support only the
// explicitly labelled sole-reservation fallback.
type ownedBranchCapture struct {
	Root, Path, Head, Branch, Digest string
	Bytes                            int64
	Intervals                        []activetime.BranchInterval
	Unavailable                      string
	Notes                            []gatestate.MeasurementBranchInterval
	raw                              []byte
}

func captureOwnedBranchEvidence(ctx context.Context, root string, since, until time.Time) (capture ownedBranchCapture) {
	git := func(args ...string) (string, error) {
		return ownedActualGit(ctx, root, append([]string{"--no-optional-locks"}, args...)...)
	}
	return captureOwnedBranchEvidenceWithGit(ctx, root, since, until, git)
}

func captureOwnedBranchEvidenceWithGit(ctx context.Context, root string, since, until time.Time, git func(...string) (string, error)) (capture ownedBranchCapture) {
	capture.Root = root
	capture.Unavailable = "HEAD history unavailable"
	observe := func() (string, string, string, error) {
		head, err := git("rev-parse", "--verify", "HEAD^{commit}")
		if err != nil {
			return "", "", "", err
		}
		branch, err := git("rev-parse", "--symbolic-full-name", "HEAD")
		if err != nil {
			return "", "", "", err
		}
		path, err := git("rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
		return head, strings.TrimPrefix(branch, "refs/heads/"), path, err
	}
	head, branch, path, err := observe()
	if err != nil {
		return
	}
	capture.Path, capture.Head, capture.Branch = path, head, branch
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return
	}
	if before.Size() > activetime.HeadHistoryByteLimit {
		capture.Unavailable = "HEAD history exceeds capture byte limit"
		return
	}
	if ctx.Err() != nil {
		return
	}
	raw, err := io.ReadAll(io.LimitReader(f, activetime.HeadHistoryByteLimit+1))
	if err != nil || ctx.Err() != nil {
		return
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || int64(len(raw)) != before.Size() || !before.ModTime().Equal(after.ModTime()) {
		capture.Unavailable = "HEAD history changed during capture"
		return
	}
	nextHead, nextBranch, nextPath, err := observe()
	if err != nil || head != nextHead || branch != nextBranch || path != nextPath {
		capture.Unavailable = "HEAD identity changed during capture"
		return
	}
	// A checkout can leave HEAD unchanged (A -> B -> A). Recheck the file
	// after the final Git observations too, not just before those subprocesses.
	final, err := os.Stat(path)
	if err != nil || !os.SameFile(before, final) || before.Size() != final.Size() || !before.ModTime().Equal(final.ModTime()) {
		capture.Unavailable = "HEAD history changed during final identity observation"
		return
	}
	capture.Bytes = int64(len(raw))
	capture.Digest = fmt.Sprintf("%x", sha256.Sum256(raw))
	capture.raw = raw
	capture.Intervals, err = activetime.ParseHeadHistory(raw, branch, head, since, until)
	if err != nil {
		capture.Unavailable = err.Error()
		return
	}
	capture.Unavailable = ""
	return
}

func branchOwnerKey(owner issue.Claimant) string {
	// A relocation changes its physical root, not the owning slot. An unnumbered
	// workspace has no slot identity and therefore uses its verified worktree.
	slot := owner.Workspace
	if slot == "" {
		slot = filepath.Clean(owner.Worktree)
	}
	return owner.Machine + "\x00" + owner.Repository + "\x00" + slot
}

// resolveSlotClaimWindows reads each immutable card history once. Historical
// owner changes matter even when today's owner is a different slot.
// A root whose historical owners disagree maps to ambiguousOwner: its sources
// are recorded as unattributed instead of failing the measurement (#603).
func resolveSlotClaimWindows(ctx context.Context, repo *tracker.Repository, inventory tracker.Snapshot, candidate string, since, until time.Time) (map[string][]activetime.BranchClaim, map[string]string, []string, error) {
	var unknown []string
	for _, bad := range inventory.Unreadable() {
		unknown = append(unknown, "unreadable card "+bad.Path)
	}
	claims, roots := map[string][]activetime.BranchClaim{}, map[string]string{}
	for _, card := range inventory.Records() {
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		events, err := loadOwnedBranchHistory(ctx, repo, inventory.Ref(), card.Path)
		// Validate the same owner chronology without changing its historical
		// retained-terminal semantics. Any competing uncertainty blocks fallback.
		if err == nil {
			_, err = tracker.DeriveEngagements(events, until)
		}
		if err == nil {
			for _, event := range events {
				if event.Owner != nil && !vocab.Issue().CanHoldOwner(event.Status) && !vocab.Issue().IsTerminal(event.Status) {
					err = fmt.Errorf("unknown ownership lifecycle status %q", event.Status)
					break
				}
			}
		}
		if err != nil {
			if strings.TrimLeft(card.ID, "0") == candidate {
				return nil, nil, nil, fmt.Errorf("branch claim history %s: %w", card.ID, err)
			}
			unknown = append(unknown, fmt.Sprintf("card %s: %v", card.ID, err))
			continue
		}
		id, err := strconv.Atoi(card.ID)
		if err != nil {
			return nil, nil, nil, err
		}
		firstClaim := true
		for i, event := range events {
			if event.Owner == nil {
				continue
			}
			start, end := event.At, until
			if firstClaim && event.Operation == tracker.OpClaim && !event.Started.IsZero() && !event.Started.Before(events[0].At) {
				start = event.Started
			}
			firstClaim = false
			if i+1 < len(events) && events[i+1].At.Before(end) {
				end = events[i+1].At
			}
			if !start.Before(end) {
				continue
			}
			key := branchOwnerKey(*event.Owner)
			if strconv.Itoa(id) == candidate {
				root := filepath.Clean(event.Owner.Worktree)
				if prior, seen := roots[root]; seen && prior != key {
					key = ambiguousOwner
				}
				roots[root] = key
			}
			model := vocab.Issue()
			claims[key] = append(claims[key], activetime.BranchClaim{Start: start, End: end, Issue: strconv.Itoa(id), Terminal: model.IsTerminal(event.Status)})
		}
	}
	// Unknown competing history is never absence. It blocks the sole-claim
	// fallback, while a directly witnessed candidate branch still selects its
	// already verified owner. This also preserves unrelated legacy cutovers.
	if len(unknown) > 0 {
		keys := map[string]bool{}
		for _, key := range roots {
			if key != ambiguousOwner {
				keys[key] = true
			}
		}
		for key := range keys {
			claims[key] = append(claims[key], activetime.BranchClaim{Start: since, End: until, Issue: "unknown-competing-history"})
		}
	}
	return claims, roots, unknown, nil
}

// ambiguousOwner marks a worktree root with more than one historical owner.
const ambiguousOwner = "\x00ambiguous"

// selectOwnedBranchWindows runs only after source lineage admission. Spawned
// sessions inherit their root author's worktree; a child cwd (e.g. brain) is
// never mistaken for the checkout whose HEAD selects this issue. Activity it
// cannot attribute (an unknown branch under concurrent claims, a source with no
// root lineage or no single owner) is returned as unattributed and not counted
// (#603); only a lineage cycle, which corrupts every source, still fails.
func selectOwnedBranchWindows(ctx context.Context, repo *tracker.Repository, inventory tracker.Snapshot, id string, sessions []transcripts.SessionEvidence, admitted map[string][]activetime.ActivityWindow, since, until time.Time) (map[string][]activetime.ActivityWindow, []ownedBranchCapture, []string, []gatestate.MeasurementUnattributed, error) {
	claims, roots, unknown, err := resolveSlotClaimWindows(ctx, repo, inventory, id, since, until)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var warnings []string
	var unattributed []gatestate.MeasurementUnattributed
	skip := func(s transcripts.SessionEvidence, kind, reason string, windows []activetime.ActivityWindow) {
		for _, w := range windows {
			unattributed = append(unattributed, gatestate.MeasurementUnattributed{Kind: kind, Source: s.Harness + "/" + s.ID, Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano), Bounded: true, Reason: reason})
		}
		warnings = append(warnings, fmt.Sprintf("branch attribution %s/%s: %s; %d windows not counted", s.Harness, s.ID, reason, len(windows)))
	}
	byID := map[string]transcripts.SessionEvidence{}
	for _, s := range sessions {
		byID[s.Harness+":"+s.ID] = s
	}
	captures := map[string]ownedBranchCapture{}
	out := map[string][]activetime.ActivityWindow{}
	for _, reason := range unknown {
		warnings = append(warnings, "branch competing history unknown: "+reason)
	}
	for _, s := range sessions {
		if len(admitted[s.File]) == 0 {
			continue
		}
		ancestor := s
		seen := map[string]bool{}
		for ancestor.Spawned {
			key := ancestor.Harness + ":" + ancestor.ID
			if seen[key] {
				return nil, nil, nil, nil, fmt.Errorf("branch source lineage cycle")
			}
			seen[key] = true
			parent, ok := byID[ancestor.Harness+":"+ancestor.ParentID]
			if !ok {
				break
			}
			ancestor = parent
		}
		if ancestor.Spawned {
			skip(s, gatestate.UnattributedSourceAttribution, "branch source lacks admitted root lineage", admitted[s.File])
			continue
		}
		root := filepath.Clean(ancestor.CWD)
		owner, ok := roots[root]
		if !ok {
			skip(s, gatestate.UnattributedSourceAttribution, "branch source has no pinned worktree owner for "+root, admitted[s.File])
			continue
		}
		if owner == ambiguousOwner {
			skip(s, gatestate.UnattributedSourceAttribution, "ambiguous historical owner for "+root, admitted[s.File])
			continue
		}
		capture, ok := captures[root]
		if !ok {
			capture = captureOwnedBranchCached(ctx, root, since, until)
			captures[root] = capture
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, nil, err
		}
		selected, err := activetime.SelectBranchActivity(id, admitted[s.File], capture.Intervals, claims[owner])
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("branch attribution for %s (%s): %w; %s", s.ID, root, err, strings.Join(append([]string{capture.Unavailable}, unknown...), "; "))
		}
		out[s.File] = selected.Windows
		if len(selected.Unknown) > 0 {
			reason := "concurrent claims have unknown, detached or ambiguous active branch"
			if capture.Unavailable != "" {
				reason += " (" + capture.Unavailable + ")"
			}
			skip(s, gatestate.UnattributedUnknownBranch, reason, selected.Unknown)
		}
		for _, note := range selected.Notes {
			capture.Notes = append(capture.Notes, gatestate.MeasurementBranchInterval{Source: s.Harness + "/" + s.ID, Start: note.Start.Format(time.RFC3339Nano), End: note.End.Format(time.RFC3339Nano), Branch: note.Branch, Kind: note.Kind})
			if note.Kind == "branch" {
				continue
			}
			warnings = append(warnings, measurementBranchWarning(capture.Notes[len(capture.Notes)-1]))
		}
		captures[root] = capture
	}
	var ordered []ownedBranchCapture
	for _, capture := range captures {
		ordered = append(ordered, capture)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Root < ordered[j].Root })
	return out, ordered, warnings, unattributed, nil
}

func measurementBranches(captures []ownedBranchCapture) []gatestate.MeasurementBranchEvidence {
	var out []gatestate.MeasurementBranchEvidence
	for _, c := range captures {
		if len(c.Notes) == 0 {
			continue
		}
		out = append(out, gatestate.MeasurementBranchEvidence{Root: c.Root, Path: c.Path, Head: c.Head, Branch: c.Branch, Digest: c.Digest, Bytes: c.Bytes, Unavailable: c.Unavailable, Intervals: c.Notes})
	}
	return out
}

func measurementBranchWarnings(m gatestate.MeasurementInputs) []string {
	var notes []string
	for _, capture := range m.Branches {
		for _, w := range capture.Intervals {
			if w.Kind == "branch" {
				continue
			}
			notes = append(notes, measurementBranchWarning(w))
		}
	}
	return notes
}

// Durations describe excluded attribution windows, not measured focused activity.
func measurementBranchWarning(w gatestate.MeasurementBranchInterval) string {
	if w.Kind == "resting-excluded" || w.Kind == "stage-excluded" {
		start, _ := time.Parse(time.RFC3339Nano, w.Start)
		end, _ := time.Parse(time.RFC3339Nano, w.End)
		return fmt.Sprintf("branch attribution %s: excluded and unassigned %.2f branch-window minutes on %s during %s → %s (not focused activity minutes)", w.Source, end.Sub(start).Minutes(), w.Branch, w.Start, w.End)
	}
	return fmt.Sprintf("branch attribution %s: %s %s during %s → %s", w.Source, w.Kind, w.Branch, w.Start, w.End)
}

// The audit facts and final eligibility must agree before a v3 result can be
// measured or retained. This checks saved inputs only, never today's checkout.
func validateOwnedBranchProof(m gatestate.MeasurementInputs, id string) error {
	if !gatestate.IsBranchMethod(m.Method) {
		return nil
	}
	fail := func(reason string) error { return fmt.Errorf("measurement branch proof: %s", reason) }
	// Shape only: the retention bounds are checked once, after trimming to
	// contributing sources (#603; pair#838 and #851 refused here before).
	if err := m.ValidateShape(); err != nil {
		return fail(err.Error())
	}
	type interval struct{ start, end time.Time }
	selected := map[string][]interval{}
	for _, capture := range m.Branches {
		for _, w := range capture.Intervals {
			branchID := activetime.BranchIssue(w.Branch)
			if w.Kind == "branch" && branchID != id || w.Kind == "foreign-branch" && (branchID == "" || branchID == id) || w.Kind == "ownership-only" && branchID != "" {
				return fail("branch witness does not match its issue attribution")
			}
			if w.Kind == "foreign-branch" || w.Kind == "resting-excluded" || w.Kind == "stage-excluded" {
				continue
			}
			start, _ := time.Parse(time.RFC3339Nano, w.Start)
			end, _ := time.Parse(time.RFC3339Nano, w.End)
			selected[w.Source] = append(selected[w.Source], interval{start, end})
		}
	}
	for source, windows := range selected {
		sort.Slice(windows, func(i, j int) bool { return windows[i].start.Before(windows[j].start) })
		var merged []interval
		for _, w := range windows {
			last := len(merged) - 1
			if last >= 0 && w.start.Before(merged[last].end) {
				return fail("overlapping branch witnesses for " + source)
			}
			if last >= 0 && w.start.Equal(merged[last].end) {
				merged[last].end = w.end
			} else {
				merged = append(merged, w)
			}
		}
		selected[source] = merged
	}
	for _, s := range m.Sources {
		witnesses := selected[s.Harness+"/"+s.ID]
		cursor := 0
		var previous time.Time
		for _, w := range s.Windows {
			start, _ := time.Parse(time.RFC3339Nano, w.Start)
			end, _ := time.Parse(time.RFC3339Nano, w.End)
			if !end.After(start) || start.Before(previous) {
				return fail("selected windows are not ordered")
			}
			previous = end
			for cursor < len(witnesses) && !witnesses[cursor].end.After(start) {
				cursor++
			}
			if cursor == len(witnesses) || start.Before(witnesses[cursor].start) || end.After(witnesses[cursor].end) {
				return fail("selected window exceeds its branch witness")
			}
		}
	}
	return nil
}
