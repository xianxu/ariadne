package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// ownedActualInputs is the command IO seam's pinned authority and evidence.
// Nil alone denotes an untracked repository eligible for labeled legacy input.
type ownedActualInputs struct {
	Snapshot        gatestate.MeasurementInputs
	Files           []string
	Prefixes        map[string]activetime.SourcePrefix
	Windows         map[string][]activetime.ActivityWindow
	DispatchWindows map[string][]activetime.ActivityWindow
	Scope           activetime.Scope
	TrackerRefs     []string
	Sessions        []transcripts.SessionEvidence
	ExcludedBrain   []transcripts.SessionEvidence
	Ledger          *gatestate.Ledger
	Warnings        []string
	SourceSlots     map[string]string
	Replay          bool
	Protected       []activetime.CreditedInterval
	LegacyPrefix    *ownedActualInputs             // separately labelled, unowned pre-cutover evidence
	CarWaits        []gatestate.MeasurementCarWait // descriptive; never passed to activity subtraction
	CarWaitNotes    []string
	// Unattributed is every uncounted record this resolve saw (#603); the
	// measurement totals it and retains a bounded prefix in the snapshot.
	Unattributed []gatestate.MeasurementUnattributed
}

type actualInputPurpose uint8

const (
	actualHistorical actualInputPurpose = iota
	actualCloseCandidate
	actualCorrectionCandidate
)

func resolveOwnedActualInputs(ctx context.Context, repoTop, brainAbs, issueNum, plans string, milestones ...string) (*ownedActualInputs, error) {
	return resolveOwnedActualInputsFor(ctx, repoTop, brainAbs, issueNum, plans, actualHistorical, milestones...)
}

func resolveOwnedActualInputsFor(ctx context.Context, repoTop, brainAbs, issueNum, plans string, purpose actualInputPurpose, milestones ...string) (*ownedActualInputs, error) {
	return resolveOwnedActualInputsLimited(ctx, repoTop, brainAbs, issueNum, plans, purpose, nil, milestones...)
}

// Explicit standalone inspection limits apply after source provenance admission
// and before branch selection. They never substitute for either authority.
func resolveOwnedActualInputsLimited(ctx context.Context, repoTop, brainAbs, issueNum, plans string, purpose actualInputPurpose, limits *ownedActualLimits, milestones ...string) (*ownedActualInputs, error) {
	ctx = ownedBranchRequestContext(ctx)
	if purpose != actualHistorical && purpose != actualCloseCandidate && purpose != actualCorrectionCandidate {
		return nil, fmt.Errorf("unsupported measurement purpose %d", purpose)
	}
	milestone := ""
	if len(milestones) > 0 {
		milestone = milestones[0]
	}
	id, err := strconv.Atoi(issueNum)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("invalid measurement issue %q", issueNum)
	}
	repo, err := recordsRepository(ctx, repoTop)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, nil
	}
	rs, err := actualInputRecords(ctx, repo, repoTop, filepath.Join(repoTop, envOr("WF_ISSUES_DIR", "workshop/issues")))
	if err != nil {
		return nil, err
	}
	if !rs.Tracker {
		return nil, nil
	}
	if rs.Stale {
		return nil, fmt.Errorf("measurement tracker evidence is stale: %v", rs.FetchErr)
	}
	rec, ok, err := rs.Require(fmt.Sprintf("%06d", id))
	if err != nil {
		return nil, err
	}
	if !ok || rec.Card == nil {
		return nil, fmt.Errorf("measurement issue #%d has no authoritative tracker card", id)
	}
	identity, err := resolveWorkspace(repoTop)
	if err != nil {
		return nil, err
	}
	repoTop = identity.WorktreeRoot
	resting := "main"
	if identity.RestingBranch != nil {
		resting = *identity.RestingBranch
	}
	target, err := gitx.ResolvePublicationTarget(ctx, repoTop, resting)
	if err != nil {
		return nil, err
	}
	head, err := ownedActualGit(ctx, repoTop, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return nil, err
	}
	completion, completed, err := issue.CardCompletion(rec.Card.Raw)
	if err != nil {
		return nil, err
	}
	var ledger gatestate.Ledger
	var warnings []string
	trackerPin := rs.Ref
	if completed && completion.Repository != target.Repository {
		return nil, fmt.Errorf("completion evidence belongs to another repository")
	}
	// A new close reviews today's candidate, even when the card still carries
	// an earlier completion or correction. Historical consumers must instead
	// replay that completion's immutable measurement generation.
	completed = completed && purpose != actualCloseCandidate
	if purpose == actualCloseCandidate {
		warnings = append(warnings, "measurement purpose: close candidate (fresh code/tracker inputs)")
	}
	if completed {
		selected, selectedLedger, correctionToken, err := completedMeasurementSnapshot(ctx, repoTop, rec, completion, target.Repository, plans, purpose == actualHistorical)
		if err != nil {
			return nil, err
		}
		ledger = selectedLedger
		if correctionToken != "" {
			out, err := resolveOwnedReplay(ctx, repoTop, issueNum, *selected, ledger)
			if err != nil {
				return nil, err
			}
			out.Warnings = append(out.Warnings, "replaying current card actual correction "+correctionToken)
			return out, nil
		}
		if milestone != "" {
			if saved, err := ownedBoundarySnapshot(ledger, milestone, target.Repository); err != nil {
				return nil, err
			} else if saved != nil {
				return resolveOwnedReplay(ctx, repoTop, issueNum, *saved, ledger)
			}
		}
		if selected != nil {
			saved := *selected
			if purpose == actualCorrectionCandidate {
				out, err := resolveOwnedReplay(ctx, repoTop, issueNum, saved, ledger)
				if err != nil {
					return nil, err
				}
				out.Warnings = append(out.Warnings, "measurement purpose: correction candidate (completion-pinned authority with fresh queue evidence)")
				if err := attachOwnedQueueEvidence(ctx, repoTop, repo, rec.Card.Path, target.Repository, id, out, purpose, gitx.RemoteTrackingRef(target.Remote, "main")); err != nil {
					return nil, err
				}
				out.Replay = false
				return out, nil
			}
			return resolveOwnedReplay(ctx, repoTop, issueNum, saved, ledger)
		}
		head = completion.ReviewedHEAD
		trackerPin, err = ownedCompletionTrackerPin(ctx, repoTop, rs.Ref, rec.Card.Path, completion)
		if err != nil {
			return nil, err
		}
		if purpose == actualHistorical {
			warnings = append(warnings, "completed issue has no retained measurement inputs; reconstructed historical authority from its exact completion binding")
		}
		if purpose == actualCorrectionCandidate {
			warnings = append(warnings, "measurement purpose: correction candidate (completion-pinned authority with fresh queue evidence)")
		}
	} else {
		var found bool
		ledger, found, err = issueBoundaryLedger(repoTop, plans, id)
		if err != nil {
			return nil, err
		}
		_ = found
		if milestone != "" {
			if saved, err := ownedBoundarySnapshot(ledger, milestone, target.Repository); err != nil {
				return nil, err
			} else if saved != nil {
				return resolveOwnedReplay(ctx, repoTop, issueNum, *saved, ledger)
			}
		}
	}
	if _, err := ownedActualGit(ctx, repoTop, "cat-file", "-e", head+"^{commit}"); err != nil {
		return nil, fmt.Errorf("measurement code pin unavailable: %w", err)
	}
	_, _, lastISO, err := gitx.CommitWindowAt(ctx, repoTop, head, issueNum, trackerPin)
	if err != nil {
		return nil, err
	}
	if lastISO == "" {
		return nil, fmt.Errorf("measurement has no issue commit window at its pinned history")
	}
	until, err := time.Parse(time.RFC3339, lastISO)
	if err != nil {
		return nil, err
	}
	claimIssue := ""
	if !completed && rec.Status() == "working" {
		_, hasOwner, err := issue.CardClaimant(rec.Card.Raw)
		if err != nil {
			return nil, err
		}
		if !hasOwner {
			claimIssue = issueNum
		}
	}
	sourceWindows, err := ownershipSourceWindows(ctx, repo, trackerPin, rec.Card.Path, target.Repository, until, claimIssue)
	if err != nil {
		return nil, err
	}
	legacyPrefix, legacyCutover, err := resolveLegacyActualPrefix(ctx, repoTop, head, trackerPin, rec.Card.Path, target.Repository, brainAbs, issueNum)
	if err != nil {
		return nil, err
	}
	if legacyPrefix != nil {
		cutover, err := time.Parse(time.RFC3339Nano, legacyPrefix.Snapshot.Until)
		if err != nil {
			return nil, err
		}
		if cutover.After(until) {
			return nil, fmt.Errorf("legacy cutover is after the measurement right edge")
		}
		sourceWindows = clipOwnedAfterCutover(sourceWindows, cutover)
		if len(sourceWindows) == 0 {
			return nil, fmt.Errorf("measurement has no owned suffix after legacy cutover")
		}
	}
	var scope activetime.Scope
	if completed {
		scope, err = ownedHistoricalScope(ctx, repoTop, issueNum, completion, ledger, target.Remote)
	} else {
		scope, err = ownedActualScope(ctx, repoTop, head, issueNum, target.Remote, identity.Branch)
	}
	if err != nil {
		return nil, err
	}
	snapshot := gatestate.MeasurementInputs{Version: gatestate.MeasurementCompactBranchInputsVersion, Method: gatestate.MeasurementMethodBranchV1, Repository: target.Repository, CodeHead: head, TrackerHead: trackerPin, BranchPoint: scope.BranchPoint, Since: sourceWindows[0].Start.Format(time.RFC3339Nano), Until: until.Format(time.RFC3339Nano), BrainRoot: brainAbs, GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true}
	if legacyPrefix != nil {
		snapshot.Version = gatestate.MeasurementCompactBranchInputsVersion
		snapshot.Method = gatestate.MeasurementMethodBranchHybridV1
		snapshot.LegacyPrefix, snapshot.LegacyCutover = &legacyPrefix.Snapshot, legacyCutover
		snapshot.Since = legacyPrefix.Snapshot.Since
	}
	roots := []string{}
	for _, w := range sourceWindows {
		snapshot.Engagements = append(snapshot.Engagements, gatestate.MeasurementEngagement{Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano), Began: w.Began.Format(time.RFC3339Nano), Roots: append([]string(nil), w.Roots...)})
		roots = append(roots, w.Roots...)
	}
	roots = uniqueStrings(append(roots, nonEmpty(brainAbs)...))
	sessions, notes, err := transcripts.DiscoverEvidence(ctx, roots, transcripts.DefaultHarnesses())
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, notes...)
	windows, notes, err := transcripts.EligibleSources(sessions, sourceWindows)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, notes...)
	inventory := rs.CardSnapshot()
	if inventory.Ref() != trackerPin {
		inventory, err = repo.SnapshotAt(trackerPin)
		if err != nil {
			return nil, err
		}
	}
	var captures []ownedBranchCapture
	dispatchWindows := windows
	if limits != nil {
		windows, err = limitOwnedSourceWindows(windows, *limits)
		if err != nil {
			return nil, err
		}
	}
	var unattributed []gatestate.MeasurementUnattributed
	windows, captures, notes, unattributed, err = selectOwnedBranchWindows(ctx, repo, inventory, issueNum, sessions, windows, sourceWindows[0].Start, until)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, notes...)
	snapshot.Branches = measurementBranches(captures)
	out := &ownedActualInputs{Snapshot: snapshot, Prefixes: map[string]activetime.SourcePrefix{}, Windows: windows, Scope: scope, TrackerRefs: []string{trackerPin}, Ledger: &ledger, Warnings: warnings, LegacyPrefix: legacyPrefix}
	out.DispatchWindows = dispatchWindows
	out.Unattributed = unattributed
	for _, s := range sessions {
		if len(windows[s.File]) == 0 {
			if brainAbs != "" && filepath.Clean(s.CWD) == filepath.Clean(brainAbs) {
				out.ExcludedBrain = append(out.ExcludedBrain, s)
			}
			continue
		}
		out.Files = append(out.Files, s.File)
		out.Sessions = append(out.Sessions, s)
		out.Prefixes[s.File] = activetime.SourcePrefix{Bytes: s.Length, Digest: s.Digest}
		pin := ownedMeasurementSource(s, windows[s.File])
		pin.DispatchWindows = ownedMeasurementSource(s, dispatchWindows[s.File]).Windows
		out.Snapshot.Sources = append(out.Snapshot.Sources, pin)
	}
	rounds, notes, err := freshActualRounds(ctx, repoTop, head, issueNum, milestone, splitRounds(ledger))
	if err != nil {
		return nil, err
	}
	out.Warnings = append(out.Warnings, notes...)
	out.Snapshot.Cuts = measurementCuts(rounds)
	if err := attachOwnedQueueEvidence(ctx, repoTop, repo, rec.Card.Path, target.Repository, id, out, purpose, gitx.RemoteTrackingRef(target.Remote, "main")); err != nil {
		return nil, err
	}
	return out, nil
}

// completedMeasurementSnapshot selects a completed actual's immutable inputs
// without touching transcript files. Exclusion can compare its exact time range
// before asking whether those files and their source-slot pins remain available.
// Historical replay uses this same selection, so both paths name one generation.
func completedMeasurementSnapshot(ctx context.Context, root string, rec tracker.IssueRecord, completion issue.Completion, repository, plans string, useCorrection bool) (*gatestate.MeasurementInputs, gatestate.Ledger, string, error) {
	id, err := strconv.Atoi(rec.ID)
	if err != nil {
		return nil, gatestate.Ledger{}, "", err
	}
	if completion.Repository != repository {
		return nil, gatestate.Ledger{}, "", fmt.Errorf("completion evidence belongs to another repository")
	}
	if useCorrection {
		correction, corrected, err := issue.CardActualCorrection(rec.Card.Raw)
		if err != nil {
			return nil, gatestate.Ledger{}, "", err
		}
		if corrected {
			if correction.CompletionToken != completion.Token || correction.Repository != completion.Repository || correction.EvidenceCommit != completion.EvidenceCommit || correction.ReviewedHEAD != completion.ReviewedHEAD {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual correction belongs to a different completion binding")
			}
			if rec.Text("actual_hours") != correction.NewActual {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual correction numeric authority differs from the current card actual_hours")
			}
			var saved gatestate.MeasurementInputs
			if err := strictCorrectionJSON([]byte(correction.Inputs), &saved); err != nil {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual correction measurement inputs: %w", err)
			}
			if err := saved.Validate(); err != nil {
				return nil, gatestate.Ledger{}, "", err
			}
			if saved.Repository != completion.Repository || saved.CodeHead != completion.ReviewedHEAD {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual correction measurement pins differ from the completion binding")
			}
			if correction.Exclude {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual sample explicitly excluded by correction %s: %s", correction.Token, correction.Reason)
			}
			if _, err := ownedActualGit(ctx, root, "cat-file", "-e", completion.EvidenceCommit+"^{commit}"); err != nil {
				return nil, gatestate.Ledger{}, "", fmt.Errorf("actual correction completion evidence is unavailable: %w", err)
			}
			return &saved, gatestate.Ledger{Gate: "boundary-review", IssueNum: id, IDPrefix: "BR"}, correction.Token, nil
		}
	}
	ledger, err := ownedLedgerAt(ctx, root, completion.EvidenceCommit, id, plans)
	if err != nil {
		return nil, gatestate.Ledger{}, "", err
	}
	for i := len(ledger.Rounds) - 1; i >= 0; i-- {
		round := ledger.Rounds[i]
		if round.Boundary != "" || round.Reviewed == "" {
			continue
		}
		if round.Reviewed != completion.ReviewedHEAD {
			return nil, gatestate.Ledger{}, "", fmt.Errorf("completion boundary ledger reviewed a different code generation")
		}
		if round.ActualInputs == nil {
			return nil, ledger, "", nil
		}
		saved := *round.ActualInputs
		if err := saved.Validate(); err != nil {
			return nil, gatestate.Ledger{}, "", err
		}
		if saved.Repository != repository {
			return nil, gatestate.Ledger{}, "", fmt.Errorf("retained measurement belongs to another repository")
		}
		if saved.CodeHead != completion.ReviewedHEAD {
			return nil, gatestate.Ledger{}, "", fmt.Errorf("retained measurement code head differs from the completion's reviewed generation")
		}
		return &saved, ledger, "", nil
	}
	return nil, ledger, "", nil
}

func attachOwnedQueueEvidence(ctx context.Context, repoTop string, repo *tracker.Repository, cardPath, repository string, issueID int, out *ownedActualInputs, purpose actualInputPurpose, mainRefs ...string) error {
	if issueID <= 0 {
		return fmt.Errorf("queue evidence requires a canonical issue ID")
	}
	until, err := time.Parse(time.RFC3339Nano, out.Snapshot.Until)
	if err != nil {
		return err
	}
	engagementEvents, err := loadOwnedBranchHistory(ctx, repo, out.Snapshot.TrackerHead, cardPath)
	if err != nil {
		return err
	}
	engagements, err := tracker.DeriveEngagements(engagementEvents, until)
	if err != nil {
		return err
	}
	out.SourceSlots = ownedQueueSourceSlots(out.Sessions, engagements)
	for i := range out.Snapshot.Sources {
		for _, session := range out.Sessions {
			if session.Harness == out.Snapshot.Sources[i].Harness && session.ID == out.Snapshot.Sources[i].ID {
				out.Snapshot.Sources[i].Slot = out.SourceSlots[session.File]
			}
		}
	}
	common, err := ownedActualGit(ctx, repoTop, "rev-parse", "--git-common-dir")
	if err != nil {
		return err
	}
	common = strings.TrimSpace(common)
	if !filepath.IsAbs(common) {
		common = filepath.Join(repoTop, common)
	}
	evidence, err := readSuiteEvidence(filepath.Join(common, "sdlc", "test-stats"), testStatsRecordLimit)
	attempts, attemptWarnings := evidence.Attempts, evidence.Warnings
	if err != nil {
		return err
	}
	since, err := time.Parse(time.RFC3339Nano, out.Snapshot.Since)
	if err != nil {
		return err
	}
	// Incomplete queue evidence degrades instead of refusing (#603), except
	// for a correction, which must stand on exact evidence. The queue time of
	// an unknown attempt cannot be subtracted, so the hours become an upper
	// bound: each attempt is recorded, with its interval when it has one.
	degrade := purpose != actualCorrectionCandidate
	unknownQueue := func(a *SuiteAttempt, reason string) {
		out.Unattributed = append(out.Unattributed, queueUnattributed(a, reason, since, until))
	}
	if len(attemptWarnings) > 0 {
		if !degrade {
			return &queueEvidenceError{Blockers: attemptWarnings}
		}
		for _, w := range attemptWarnings {
			unknownQueue(nil, w)
		}
	}
	if out.Snapshot.LegacyPrefix != nil {
		since, err = time.Parse(time.RFC3339Nano, out.Snapshot.LegacyPrefix.Until)
		if err != nil {
			return err
		}
		// The legacy prefix can precede the first owner. Queue idle may
		// exclude only the owned suffix, even when queuing began earlier.
		if len(engagements) > 0 && since.Before(engagements[0].Start) {
			since = engagements[0].Start
		}
	}
	attachCarWaitInputs(ctx, repoTop, common, repository, issueID, engagements, out)
	attempts, err = potentialIssueQueueAttempts(attempts, engagements, repository, since, until)
	if err != nil {
		return err
	}
	var heads []string
	for _, a := range attempts {
		heads = append(heads, a.SHA)
	}
	memberships, err := readQueueHeadMemberships(ctx, heads, path.Base(repository), gitQueueHeadReader(repoTop, mainRefs...))
	if err != nil {
		var blockers []string
		for _, a := range attempts {
			blockers = append(blockers, fmt.Sprintf("suite attempt %s SHA %s: %v; restore immutable commit evidence and retry sdlc actual --issue %d", a.ID, a.SHA, err, issueID))
		}
		if len(blockers) == 0 {
			return err
		}
		if !degrade {
			return &queueEvidenceError{Blockers: blockers}
		}
		for i := range attempts {
			unknownQueue(&attempts[i], blockers[i])
		}
		out.Warnings = append(out.Warnings, blockers...)
		return nil
	}
	if degrade {
		var known []SuiteAttempt
		for i, a := range attempts {
			if m := memberships[a.SHA]; !m.Known {
				unknownQueue(&attempts[i], fmt.Sprintf("suite attempt %s SHA %s has unknown issue-head attribution: %s", a.ID, a.SHA, m.Reason))
				continue
			}
			known = append(known, a)
		}
		attempts = known
	}
	attempts, err = filterIssueQueueAttempts(attempts, memberships, issueID)
	if err != nil {
		return err
	}
	waits, notes, err := selectOwnedQueueAttempts(attempts, engagements, repository, since, until)
	if err != nil {
		return err
	}
	if len(notes) > 0 {
		if !degrade {
			return &queueEvidenceError{Blockers: notes}
		}
		for _, note := range notes {
			unknownQueue(queueAttemptNamed(attempts, note), note)
		}
	}
	if purpose == actualCorrectionCandidate && len(waits) == 0 {
		return fmt.Errorf("correction candidate has no complete, owned suite queue records for the completion window")
	}
	// Opaque future records are not queue samples. Preserve supported owned
	// measurements and name the unavailable coverage instead of poisoning
	// every issue that reads this shared store.
	out.Warnings = append(out.Warnings, evidence.coverageNotes(strings.TrimPrefix(repository, "github.com/"))...)
	out.Warnings = append(out.Warnings, attemptWarnings...)
	out.Warnings = append(out.Warnings, notes...)
	if len(waits) > 0 {
		switch {
		case gatestate.IsBranchMethod(out.Snapshot.Method):
			out.Snapshot.Method = gatestate.MeasurementMethodBranchQueueV1
			if out.LegacyPrefix != nil {
				out.Snapshot.Method = gatestate.MeasurementMethodBranchHybridQueueV1
			}
		case out.LegacyPrefix != nil:
			out.Snapshot.Method = gatestate.MeasurementMethodHybridQueueV1
		default:
			out.Snapshot.Method = gatestate.MeasurementMethodQueueV1
		}
		out.Snapshot.QueueWaits = waits
	}
	return nil
}

// queueUnattributed records one queue attempt whose wait cannot be
// subtracted: bounded by its queue interval inside the sample when it has
// one, else unbounded.
func queueUnattributed(a *SuiteAttempt, reason string, since, until time.Time) gatestate.MeasurementUnattributed {
	reason = oneLine(reason)
	if len(reason) > 400 {
		reason = reason[:397] + "..."
	}
	u := gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedQueueUnknown, Reason: reason}
	if a == nil || a.QueuedAt == nil || a.AcquiredAt == nil {
		return u
	}
	start, end := *a.QueuedAt, *a.AcquiredAt
	if start.Before(since) {
		start = since
	}
	if end.After(until) {
		end = until
	}
	if end.After(start) {
		u.Source, u.Start, u.End, u.Bounded = "suite/"+a.ID, start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano), true
	}
	return u
}

// queueAttemptNamed finds the attempt a selection note names by its ID.
func queueAttemptNamed(attempts []SuiteAttempt, note string) *SuiteAttempt {
	for i := range attempts {
		if attempts[i].ID != "" && strings.Contains(note, " "+attempts[i].ID+" ") {
			return &attempts[i]
		}
	}
	return nil
}

func ownedMeasurementSource(s transcripts.SessionEvidence, windows []activetime.ActivityWindow) gatestate.MeasurementSource {
	out := gatestate.MeasurementSource{Harness: s.Harness, ID: s.ID, CWD: s.CWD, ParentID: s.ParentID, Digest: s.Digest, Started: s.Started.Format(time.RFC3339Nano), Spawned: s.Spawned, Bytes: s.Length}
	for _, w := range windows {
		out.Windows = append(out.Windows, gatestate.MeasurementInterval{Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano)})
	}
	return out
}

func ownedMeasurementCuts(l gatestate.Ledger) []gatestate.MeasurementCut {
	return measurementCuts(splitRounds(l))
}

func measurementCuts(rounds []activetime.BoundaryRound) []gatestate.MeasurementCut {
	var out []gatestate.MeasurementCut
	for _, r := range rounds {
		cut := gatestate.MeasurementCut{Boundary: r.Boundary, Finalized: r.Finalized.Format(time.RFC3339Nano), Finalizing: r.Finalizing}
		if !r.Dispatched.IsZero() {
			cut.Dispatched = r.Dispatched.Format(time.RFC3339Nano)
		}
		out = append(out, cut)
	}
	return out
}

type localMeasurementContextKey struct{}

func actualInputRecords(ctx context.Context, repo *tracker.Repository, root, details string) (tracker.Records, error) {
	if local, _ := ctx.Value(localMeasurementContextKey{}).(bool); !local {
		return tracker.LoadRecords(ctx, repo, details, tracker.PreferFresh)
	}
	snap, ok, err := repo.LocalSnapshot()
	if err != nil {
		return tracker.Records{}, err
	}
	if !ok {
		tracked, e := localTrackerDeclared(root)
		if e != nil {
			return tracker.Records{}, e
		}
		if !tracked {
			return tracker.Records{}, nil
		}
		return tracker.Records{}, fmt.Errorf("local tracker snapshot unavailable")
	}
	return tracker.ComposeRecords(tracker.Records{Tracker: true, Ref: snap.Ref()}, snap.Records(), snap.Unreadable(), nil), nil
}

// ownershipSourceWindows is shared authority for active time and token usage.
// Each caller supplies its own explicit right edge; tokens include the review tail.
func ownershipSourceWindows(ctx context.Context, repo *tracker.Repository, pin, cardPath, repository string, until time.Time, claimIssue string) ([]transcripts.SourceWindow, error) {
	events, err := loadOwnedBranchHistory(ctx, repo, pin, cardPath)
	if err != nil {
		return nil, err
	}
	engagements, err := tracker.DeriveEngagements(events, until)
	if err != nil {
		return nil, err
	}
	if len(engagements) == 0 {
		neverOwned := true
		for _, event := range events {
			if event.Owner != nil {
				neverOwned = false
				break
			}
		}
		if neverOwned && claimIssue != "" {
			return nil, &unownedActualError{Issue: claimIssue}
		}
		return nil, fmt.Errorf("no post-cutover owned interval yet: measurement has no verified owned interval before its pinned right edge; reconcile and record post-cutover work before closing")
	}
	machine, _, err := localMachine()
	if err != nil {
		return nil, err
	}
	return ownedSourceWindows(ctx, engagements, repository, machine)
}

// Roots are validated within each continuous engagement, including retained
// relocated paths whose successor still proves repository and machine identity.
func ownedSourceWindows(ctx context.Context, engagements []tracker.EngagementWindow, repository, machine string) ([]transcripts.SourceWindow, error) {
	verified := map[string]map[string]string{}
	missing := map[string][]issue.Claimant{}
	for _, w := range engagements {
		key := w.Began.Format(time.RFC3339Nano)
		if verified[key] == nil {
			verified[key] = map[string]string{}
		}
		for _, owner := range w.Owners {
			if owner.Machine != machine {
				return nil, fmt.Errorf("measurement owner %s is on another machine; local transcript provenance is unavailable", owner.MachineName)
			}
			if owner.Repository != repository {
				return nil, fmt.Errorf("measurement owner path %s belongs to another repository", owner.Worktree)
			}
			if _, ok := verified[key][owner.Worktree]; ok {
				continue
			}
			path, err := workspace.CanonicalPath(owner.Worktree)
			if os.IsNotExist(err) {
				missing[key] = append(missing[key], owner)
				continue
			}
			if err != nil {
				return nil, err
			}
			identity, err := resolveWorkspace(path)
			if err != nil {
				return nil, fmt.Errorf("verify measurement owner %s: %w", path, err)
			}
			resting := "main"
			if identity.RestingBranch != nil {
				resting = *identity.RestingBranch
			}
			target, err := gitx.ResolvePublicationTarget(ctx, identity.WorktreeRoot, resting)
			if err != nil {
				return nil, err
			}
			if target.Repository != repository {
				return nil, fmt.Errorf("measurement owner path %s now belongs to another repository", path)
			}
			verified[key][owner.Worktree] = identity.WorktreeRoot
		}
	}
	for key, owners := range missing {
		if len(verified[key]) == 0 {
			return nil, fmt.Errorf("measurement owner path %s is unavailable with no verified relocation successor", owners[0].Worktree)
		}
		for _, owner := range owners {
			verified[key][owner.Worktree] = filepath.Clean(owner.Worktree)
		}
	}
	out := make([]transcripts.SourceWindow, 0, len(engagements))
	for _, w := range engagements {
		window := transcripts.SourceWindow{Start: w.Start, End: w.End, Began: w.Began}
		for _, owner := range w.Owners {
			window.Roots = append(window.Roots, verified[w.Began.Format(time.RFC3339Nano)][owner.Worktree])
		}
		window.Roots = uniqueStrings(window.Roots)
		out = append(out, window)
	}
	return out, nil
}

func ownedReplaySources(saved gatestate.MeasurementInputs, found []transcripts.SessionEvidence) (*ownedActualInputs, error) {
	out := &ownedActualInputs{Snapshot: saved, Prefixes: map[string]activetime.SourcePrefix{}, Windows: map[string][]activetime.ActivityWindow{}, SourceSlots: map[string]string{}, Scope: activetime.Scope{Head: saved.CodeHead, BranchPoint: saved.BranchPoint}, TrackerRefs: nonEmpty(saved.TrackerHead), Replay: true}
	for _, credit := range saved.Protected {
		slot := parseAdoptedSlot(credit.Slot)
		if !slot.known {
			return nil, fmt.Errorf("retained protected issue %s has no canonical pinned slot", credit.Issue)
		}
		start, err := time.Parse(time.RFC3339Nano, credit.Start)
		if err != nil {
			return nil, err
		}
		end, err := time.Parse(time.RFC3339Nano, credit.End)
		if err != nil {
			return nil, err
		}
		out.Protected = append(out.Protected, activetime.CreditedInterval{Issue: credit.Issue, Slot: slot.String(), Start: start, End: end})
	}
	byID := map[[2]string]transcripts.SessionEvidence{}
	for _, s := range found {
		key := [2]string{s.Harness, s.ID}
		if previous, ok := byID[key]; ok && previous.File != s.File {
			return nil, fmt.Errorf("ambiguous retained source %s/%s", s.Harness, s.ID)
		}
		byID[key] = s
	}
	for _, pin := range saved.Sources {
		if len(pin.DispatchWindows) > 0 && out.DispatchWindows == nil {
			out.DispatchWindows = map[string][]activetime.ActivityWindow{}
		}
		slot := parseAdoptedSlot(pin.Slot)
		if pin.Slot != "" && !slot.known {
			return nil, fmt.Errorf("retained source %s/%s has no canonical pinned slot", pin.Harness, pin.ID)
		}
		if len(saved.Protected) > 0 && !slot.known {
			return nil, fmt.Errorf("retained source %s/%s has no pinned slot for protected intervals", pin.Harness, pin.ID)
		}
		s, ok := byID[[2]string{pin.Harness, pin.ID}]
		if !ok {
			return nil, fmt.Errorf("retained measurement source %s/%s is unavailable", pin.Harness, pin.ID)
		}
		if filepath.Clean(s.CWD) != filepath.Clean(pin.CWD) || s.ParentID != pin.ParentID || s.Spawned != pin.Spawned {
			return nil, fmt.Errorf("retained source %s/%s provenance changed", pin.Harness, pin.ID)
		}
		s.Digest, s.Length = pin.Digest, pin.Bytes
		s.Started, _ = time.Parse(time.RFC3339Nano, pin.Started)
		out.Files = append(out.Files, s.File)
		out.Sessions = append(out.Sessions, s)
		out.SourceSlots[s.File] = slot.String()
		out.Prefixes[s.File] = activetime.SourcePrefix{Bytes: pin.Bytes, Digest: pin.Digest}
		for _, selection := range []struct {
			windows []gatestate.MeasurementInterval
			target  map[string][]activetime.ActivityWindow
		}{{pin.Windows, out.Windows}, {pin.DispatchWindows, out.DispatchWindows}} {
			for _, interval := range selection.windows {
				start, err := time.Parse(time.RFC3339Nano, interval.Start)
				if err != nil {
					return nil, err
				}
				end, err := time.Parse(time.RFC3339Nano, interval.End)
				if err != nil {
					return nil, err
				}
				var began time.Time
				if gatestate.IsEngagementMethod(saved.Method) {
					for _, engagement := range saved.Engagements {
						left, err := time.Parse(time.RFC3339Nano, engagement.Start)
						if err != nil {
							return nil, err
						}
						right, err := time.Parse(time.RFC3339Nano, engagement.End)
						if err != nil {
							return nil, err
						}
						if start.Before(left) || end.After(right) {
							continue
						}
						origin, err := time.Parse(time.RFC3339Nano, engagement.Began)
						if err != nil {
							return nil, err
						}
						if !began.IsZero() && !began.Equal(origin) {
							return nil, errors.New("retained source window has ambiguous dispatch engagement")
						}
						began = origin
					}
					if began.IsZero() {
						return nil, errors.New("retained source window has no dispatch engagement")
					}
				}
				selection.target[s.File] = append(selection.target[s.File], activetime.ActivityWindow{Start: start, End: end, Began: began})
			}
		}
	}
	return out, nil
}

func resolveOwnedReplay(ctx context.Context, root, issueNum string, saved gatestate.MeasurementInputs, ledger gatestate.Ledger) (*ownedActualInputs, error) {
	if gatestate.IsSummaryMethod(saved.Method) && saved.Summary != nil {
		return nil, fmt.Errorf("measurement is summarised and not replayable (%s)", saved.Summary.Reason)
	}
	if err := validateOwnedBranchProof(saved, issueNum); err != nil {
		return nil, err
	}
	var legacyPrefix *ownedActualInputs
	if saved.LegacyPrefix != nil {
		if err := validateRetainedLegacyPrefix(ctx, root, issueNum, saved); err != nil {
			return nil, err
		}
		var err error
		legacyPrefix, err = resolveOwnedReplay(ctx, root, issueNum, *saved.LegacyPrefix, gatestate.Ledger{})
		if err != nil {
			return nil, fmt.Errorf("retained legacy prefix: %w", err)
		}
	}
	for _, pin := range nonEmpty(saved.CodeHead, saved.TrackerHead, saved.BranchPoint) {
		if _, err := ownedActualGit(ctx, root, "cat-file", "-e", pin+"^{commit}"); err != nil {
			return nil, fmt.Errorf("retained measurement Git pin %s unavailable: %w", pin, err)
		}
	}
	var roots []string
	for _, source := range saved.Sources {
		roots = append(roots, source.CWD)
	}
	found, warnings, err := transcripts.DiscoverEvidence(ctx, uniqueStrings(roots), transcripts.DefaultHarnesses())
	if err != nil {
		return nil, err
	}
	out, err := ownedReplaySources(saved, found)
	if err != nil {
		return nil, err
	}
	out.LegacyPrefix = legacyPrefix
	for _, s := range out.Sessions {
		if err := verifyOwnedPrefix(ctx, s.File, out.Prefixes[s.File]); err != nil {
			return nil, err
		}
	}
	out.Scope.Issue = issueNum
	out.Warnings = append(warnings, "replaying retained measurement inputs from its finalized boundary evidence")
	out.Warnings = append(out.Warnings, measurementBranchWarnings(saved)...)
	// The snapshot owns the cuts. Later ledger edits cannot move old phases.
	ledger.Rounds = nil
	for i, cut := range saved.Cuts {
		reviewed := ""
		if cut.Finalizing {
			reviewed = saved.CodeHead
		}
		ledger.Rounds = append(ledger.Rounds, gatestate.Round{N: i + 1, Boundary: cut.Boundary, Timestamp: cut.Finalized, Dispatched: cut.Dispatched, Reviewed: reviewed})
	}
	out.Ledger = &ledger
	return out, nil
}

func ownedBoundarySnapshot(ledger gatestate.Ledger, boundary, repository string) (*gatestate.MeasurementInputs, error) {
	for i := len(ledger.Rounds) - 1; i >= 0; i-- {
		round := ledger.Rounds[i]
		if round.Boundary != boundary || round.Reviewed == "" {
			continue
		}
		if round.ActualInputs == nil {
			return nil, nil
		}
		saved := *round.ActualInputs
		if err := saved.Validate(); err != nil {
			return nil, err
		}
		if saved.Repository != repository {
			return nil, fmt.Errorf("retained milestone measurement belongs to another repository")
		}
		if saved.CodeHead != round.Reviewed {
			return nil, fmt.Errorf("retained milestone measurement code head differs from its finalized reviewed generation")
		}
		return &saved, nil
	}
	return nil, nil
}

func verifyOwnedPrefix(ctx context.Context, path string, pin activetime.SourcePrefix) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha256.New()
	reader := io.LimitReader(f, pin.Bytes)
	buffer := make([]byte, 32*1024)
	var copied int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := reader.Read(buffer)
		if n > 0 {
			copied += int64(n)
			_, _ = hash.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if copied != pin.Bytes || fmt.Sprintf("%x", hash.Sum(nil)) != pin.Digest {
		return fmt.Errorf("retained source %s changed or shortened since measurement", path)
	}
	return nil
}

func ownedLedgerAt(ctx context.Context, root, pin string, id int, plans string) (gatestate.Ledger, error) {
	if plans == "" {
		plans = plansDir()
	}
	if filepath.IsAbs(plans) {
		relative, err := filepath.Rel(root, plans)
		if err != nil || relative == ".." || strings.HasPrefix(relative, "../") {
			return gatestate.Ledger{}, fmt.Errorf("measurement ledger must be inside repository")
		}
		plans = relative
	}
	listing, err := ownedActualGit(ctx, root, "ls-tree", "-r", "--name-only", pin, "--", filepath.ToSlash(plans), "workshop/history/plans")
	if err != nil {
		return gatestate.Ledger{}, err
	}
	var paths []string
	for _, path := range strings.Split(listing, "\n") {
		name := filepath.Base(path)
		if strings.HasPrefix(name, fmt.Sprintf("%06d-", id)) && strings.HasSuffix(name, "-"+boundaryGateSuffix+".md") {
			paths = append(paths, path)
		}
	}
	if len(paths) == 0 {
		return gatestate.Ledger{Gate: "boundary-review", IssueNum: id, IDPrefix: "BR"}, nil
	}
	if len(paths) != 1 {
		return gatestate.Ledger{}, fmt.Errorf("completion evidence has ambiguous boundary ledgers for #%d", id)
	}
	raw, err := ownedActualGit(ctx, root, "show", pin+":"+paths[0])
	if err != nil {
		return gatestate.Ledger{}, err
	}
	ledger, err := gatestate.ParseSidecar(raw)
	if err != nil {
		return gatestate.Ledger{}, err
	}
	if ledger.IssueNum != id || ledger.Gate != "boundary-review" {
		return gatestate.Ledger{}, fmt.Errorf("completion evidence boundary ledger identity mismatch")
	}
	return ledger, nil
}

func ownedCompletionTrackerPin(ctx context.Context, root, pin, cardPath string, completion issue.Completion) (string, error) {
	history, err := ownedActualGit(ctx, root, "log", "--format=%H", "--max-count="+strconv.Itoa(tracker.ClaimHistoryLimit+1), pin, "--", cardPath)
	if err != nil {
		return "", err
	}
	commits := strings.Fields(history)
	if len(commits) > tracker.ClaimHistoryLimit {
		return "", fmt.Errorf("historical completion tracker history is incomplete")
	}
	for i := len(commits) - 1; i >= 0; i-- {
		raw, err := ownedActualGit(ctx, root, "show", commits[i]+":"+cardPath)
		if err != nil {
			return "", err
		}
		candidate, ok, err := issue.CardCompletion([]byte(raw))
		if err != nil {
			return "", err
		}
		if ok && candidate.Token == completion.Token && candidate.ReviewedHEAD == completion.ReviewedHEAD && candidate.EvidenceCommit == completion.EvidenceCommit {
			return commits[i], nil
		}
	}
	return "", fmt.Errorf("historical completion tracker pin unavailable")
}

func ownedActualScope(ctx context.Context, root, head, issueNum, remote string, branch *string) (activetime.Scope, error) {
	scope := activetime.Scope{Head: head, Issue: issueNum}
	var base string
	for _, ref := range []string{gitx.RemoteTrackingRef(remote, "main"), "refs/heads/main"} {
		candidate, err := ownedActualGit(ctx, root, "merge-base", ref, head)
		if err != nil {
			if ctx.Err() != nil {
				return scope, ctx.Err()
			}
			continue
		}
		if base == "" {
			base = candidate
			continue
		}
		if _, err := ownedActualGit(ctx, root, "merge-base", "--is-ancestor", base, candidate); err == nil {
			base = candidate
		}
	}
	id, _ := strconv.Atoi(issueNum)
	onBranch := branch != nil && strings.HasPrefix(*branch+"-", fmt.Sprintf("%06d-", id))
	if base != head || onBranch {
		scope.BranchPoint = base
	}
	return scope, nil
}

// ownedReviewWindow accepts one concrete recorded window, never an ambient
// ref such as HEAD whose meaning changes after a close.
func ownedReviewWindow(message string) (string, string, error) {
	var window string
	count := 0
	for _, line := range strings.Split(message, "\n") {
		if strings.HasPrefix(line, "Review-Window:") {
			count++
			window = strings.TrimSpace(strings.TrimPrefix(line, "Review-Window:"))
		}
	}
	if count != 1 {
		return "", "", fmt.Errorf("historical measurement requires one exact Review-Window trailer")
	}
	base, head, ok := strings.Cut(window, "..")
	concrete := func(oid string) bool {
		if len(oid) < 4 || len(oid) > 64 {
			return false
		}
		for _, c := range oid {
			if !strings.ContainsRune("0123456789abcdef", c) {
				return false
			}
		}
		return true
	}
	if !ok || !concrete(base) || !concrete(head) {
		return "", "", fmt.Errorf("historical Review-Window must name concrete commit identities")
	}
	return base, head, nil
}

func ownedHistoricalScope(ctx context.Context, root, issueNum string, completion issue.Completion, ledger gatestate.Ledger, remotes ...string) (activetime.Scope, error) {
	scope := activetime.Scope{Head: completion.ReviewedHEAD, Issue: issueNum}
	finalized := 0
	multipleBoundaries := false
	for _, round := range ledger.Rounds {
		if _, _, ok := cutTimes(round); !ok {
			continue
		}
		if round.Boundary != "" {
			multipleBoundaries = true
		}
		if round.Reviewed != "" {
			finalized++
		}
	}
	message, err := ownedActualGit(ctx, root, "show", "-s", "--format=%B", completion.EvidenceCommit)
	if err != nil {
		return scope, err
	}
	baseAbbrev, headAbbrev, err := ownedReviewWindow(message)
	if err != nil {
		return scope, err
	}
	base, err := ownedActualGit(ctx, root, "rev-parse", "--verify", baseAbbrev+"^{commit}")
	if err != nil {
		return scope, fmt.Errorf("historical review base cannot resolve unambiguously: %w", err)
	}
	head, err := ownedActualGit(ctx, root, "rev-parse", "--verify", headAbbrev+"^{commit}")
	if err != nil {
		return scope, fmt.Errorf("historical review head cannot resolve unambiguously: %w", err)
	}
	if head != completion.ReviewedHEAD {
		return scope, fmt.Errorf("historical Review-Window head differs from completion reviewed generation")
	}
	if _, err := ownedActualGit(ctx, root, "merge-base", "--is-ancestor", base, head); err != nil {
		return scope, fmt.Errorf("historical Review-Window base is not an ancestor of the reviewed head: %w", err)
	}
	if multipleBoundaries || finalized > 1 {
		remote := ""
		if len(remotes) > 0 {
			remote = remotes[0]
		}
		proven, err := ownedMainIntegrationBase(ctx, root, issueNum, head, base, remote)
		if err != nil {
			return scope, err
		}
		if !proven {
			return scope, fmt.Errorf("historical multi-boundary or re-close measurement lacks its original pinned branch point; exclude the sample explicitly")
		}
	}
	scope.BranchPoint = base
	return scope, nil
}

// A recorded main integration at the reviewed HEAD proves that a Review-Window
// base is the issue's main boundary, rather than an earlier milestone head.
// Both topology and the explicit publication-remote subject must agree.
func ownedMainIntegrationBase(ctx context.Context, root, issueNum, head, base, remote string) (bool, error) {
	if remote == "" {
		return false, nil
	}
	metadata, err := ownedActualGit(ctx, root, "show", "-s", "--format=%P%x00%s", head)
	if err != nil {
		return false, err
	}
	parentsRaw, subject, ok := strings.Cut(metadata, "\x00")
	if !ok {
		return false, fmt.Errorf("historical reviewed merge metadata is malformed")
	}
	parents := strings.Fields(parentsRaw)
	if len(parents) != 2 || parents[1] != base {
		return false, nil
	}
	id, err := strconv.Atoi(issueNum)
	if err != nil {
		return false, err
	}
	prefix := "Merge remote-tracking branch '" + remote + "/main' into "
	if !strings.HasPrefix(subject, prefix) {
		return false, nil
	}
	branch := strings.TrimPrefix(subject, prefix)
	return strings.HasPrefix(branch, fmt.Sprintf("%06d-", id)) && !strings.ContainsAny(branch, " \t\r\n"), nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

// All measurement Git reads share the bounded process-group boundary (#332).
func ownedActualGit(ctx context.Context, root string, args ...string) (string, error) {
	stdout, stderr, err := gitx.RunGitInContext(ctx, root, nil, args...)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if len(stdout) > 4<<20 || len(stderr) > 4<<20 {
		return "", fmt.Errorf("measurement Git output exceeds bound")
	}
	if err != nil {
		return "", fmt.Errorf("measurement git %s: %w: %s", args[0], err, strings.TrimSpace(string(stderr)))
	}
	return strings.TrimSuffix(string(stdout), "\n"), nil
}

func localTrackerDeclared(root string) (bool, error) {
	_, e := os.Stat(filepath.Join(root, tracker.CutoverMarkerPath))
	if os.IsNotExist(e) {
		return false, nil
	}
	return e == nil, e
}
