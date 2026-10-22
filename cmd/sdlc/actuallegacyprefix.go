package main

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// legacyActualPrefixAuthority proves only legacy timing and repository scope.
// A pre-cutover started stamp is never converted into a tracker owner.
func legacyActualPrefixAuthority(ctx context.Context, root, head, trackerPin, cardPath, repository, brain string) (*gatestate.MeasurementInputs, string, error) {
	sourceMain, trackerRoot, err := legacyMigrationSource(ctx, root, head)
	if err != nil {
		return nil, "", err
	}
	if _, err = ownedActualGit(ctx, root, "merge-base", "--is-ancestor", trackerRoot, trackerPin); err != nil {
		return nil, "", fmt.Errorf("legacy prefix tracker root is outside pinned history: %w", err)
	}
	listed, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "--name-only", trackerRoot, "--", cardPath)
	if err != nil {
		return nil, "", err
	}
	if listed == "" {
		return nil, "", nil
	}
	raw, err := archiveReadContext(ctx, root, "cat-file", "blob", trackerRoot+":"+cardPath)
	if err != nil {
		return nil, "", err
	}
	card, err := issue.ParseCard(raw)
	if err != nil {
		return nil, "", err
	}
	status, err := issue.FieldString(card.Frontmatter, "status")
	if err != nil {
		return nil, "", err
	}
	if vocab.Issue().IsTerminal(status) {
		return nil, "", nil
	}
	started, err := issue.FieldString(card.Frontmatter, "started")
	if err != nil {
		return nil, "", err
	}
	_, firstISO, _, err := gitx.CommitWindowAt(ctx, root, sourceMain, strings.TrimLeft(card.ID, "0"))
	if err != nil {
		return nil, "", err
	}
	// Repositories born tracked may seed cards and mirrored details at their
	// initial marker; those are not legacy imports even if a setup commit names N.
	sourcePath := path.Join(vocab.Issue().Discovery().Home, path.Base(cardPath))
	sourceListed, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "--name-only", sourceMain, "--", sourcePath)
	if err != nil {
		return nil, "", err
	}
	if sourceListed != "" {
		sourceRaw, e := archiveReadContext(ctx, root, "cat-file", "blob", sourceMain+":"+sourcePath)
		if e != nil {
			return nil, "", e
		}
		if issue.HasMirror(sourceRaw) {
			return nil, "", nil
		}
	}
	source, err := legacyImportedSource(ctx, root, sourceMain, repository, tracker.Record{ID: card.ID, Path: cardPath, Raw: raw, Card: card})
	if err != nil {
		return nil, "", fmt.Errorf("legacy prefix imported source: %w", err)
	}
	if path.Dir(source.Path) != vocab.Issue().Discovery().Home {
		return nil, "", fmt.Errorf("legacy prefix requires an active imported issue")
	}
	fm, _, err := issue.Parse(string(source.Raw))
	if err != nil {
		return nil, "", err
	}
	sourceStarted, err := issue.FieldString(fm, "started")
	if err != nil {
		return nil, "", err
	}
	if sourceStarted != started {
		return nil, "", fmt.Errorf("legacy prefix claim differs from imported root")
	}
	anchor := sourceStarted
	if anchor == "" {
		// Same legacy working-transition fallback, pinned before the cutover and
		// bounded before picking its earliest transition.
		history, e := ownedActualGit(ctx, root, "--literal-pathspecs", "log", "-G^status: *working", "--reverse", "--format=%aI", "--max-count="+strconv.Itoa(legacyCompletionHistoryLimit+1), sourceMain, "--", source.Path)
		if e != nil {
			return nil, "", e
		}
		transitions := strings.Fields(history)
		if len(transitions) > legacyCompletionHistoryLimit {
			return nil, "", fmt.Errorf("legacy prefix claim history exceeds complete-history bound")
		}
		if len(transitions) > 0 {
			anchor = transitions[0]
		} else {
			anchor = firstISO
		}
	}
	if anchor == "" {
		return nil, "", nil // no pre-cutover engagement evidence
	}
	cutover, err := ownedActualGit(ctx, root, "--literal-pathspecs", "log", "--diff-filter=A", "--format=%H", "--max-count=1", head, "--", tracker.CutoverMarkerPath)
	if err != nil {
		return nil, "", err
	}
	untilISO, err := ownedActualGit(ctx, root, "show", "-s", "--format=%cI", cutover)
	if err != nil {
		return nil, "", err
	}
	since, err := time.Parse(time.RFC3339Nano, anchor)
	if err != nil {
		return nil, "", fmt.Errorf("legacy prefix claim timestamp: %w", err)
	}
	until, err := time.Parse(time.RFC3339Nano, untilISO)
	if err != nil {
		return nil, "", err
	}
	if since.After(until) {
		return nil, "", fmt.Errorf("legacy prefix claim is after the cutover")
	}
	if since.Equal(until) {
		return nil, "", nil
	}
	snapshot := &gatestate.MeasurementInputs{Version: gatestate.MeasurementInputsVersion, Method: gatestate.MeasurementMethodLegacyV1, Repository: repository, CodeHead: sourceMain, Since: since.Format(time.RFC3339Nano), Until: until.Format(time.RFC3339Nano), BrainRoot: brain, GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true}
	return snapshot, cutover, nil
}

func resolveLegacyActualPrefix(ctx context.Context, root, head, trackerPin, cardPath, repository, brain, issueNum string) (*ownedActualInputs, string, error) {
	snapshot, cutover, err := legacyActualPrefixAuthority(ctx, root, head, trackerPin, cardPath, repository, brain)
	if err != nil || snapshot == nil {
		return nil, cutover, err
	}
	prefix, err := discoverLegacyPrefixInputs(ctx, root, brain, *snapshot)
	if err != nil {
		return nil, "", fmt.Errorf("legacy pre-cutover attribution: %w", err)
	}
	prefix.Scope = activetime.Scope{Head: snapshot.CodeHead, Issue: issueNum}
	return prefix, cutover, nil
}

// clipOwnedAfterCutover retains the real dispatch origin while excluding
// bootstrap time already measured under the legacy attribution rule.
func clipOwnedAfterCutover(windows []transcripts.SourceWindow, cutover time.Time) []transcripts.SourceWindow {
	out := make([]transcripts.SourceWindow, 0, len(windows))
	for _, w := range windows {
		if !w.End.After(cutover) {
			continue
		}
		if w.Start.Before(cutover) {
			w.Start = cutover
		}
		out = append(out, w)
	}
	return out
}

// validateRetainedLegacyPrefix binds saved heuristic bounds back to the same
// immutable cutover and imported card; retained source hashes alone do not
// establish authority to create a historical interval.
func validateRetainedLegacyPrefix(ctx context.Context, root, issueNum string, saved gatestate.MeasurementInputs) error {
	if saved.LegacyPrefix == nil {
		return nil
	}
	_, trackerRoot, err := legacyMigrationSource(ctx, root, saved.CodeHead)
	if err != nil {
		return err
	}
	listed, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "-r", "--name-only", trackerRoot, "--", vocab.Issue().Discovery().Cards)
	if err != nil {
		return err
	}
	var cardPath string
	for _, candidate := range strings.Split(listed, "\n") {
		id, _, ok := issue.ParseFilename(path.Base(candidate))
		if !ok || strings.TrimLeft(id, "0") != strings.TrimLeft(issueNum, "0") {
			continue
		}
		if cardPath != "" {
			return fmt.Errorf("retained legacy prefix has ambiguous imported card")
		}
		cardPath = candidate
	}
	if cardPath == "" {
		return fmt.Errorf("retained legacy prefix has no imported card")
	}
	expected, cutover, err := legacyActualPrefixAuthority(ctx, root, saved.CodeHead, saved.TrackerHead, cardPath, saved.Repository, saved.BrainRoot)
	if err != nil {
		return err
	}
	p := saved.LegacyPrefix
	if expected == nil || cutover != saved.LegacyCutover || p.CodeHead != expected.CodeHead || p.Since != expected.Since || p.Until != expected.Until || p.BrainRoot != expected.BrainRoot || p.GapMinutes != expected.GapMinutes || p.CommitWeight != expected.CommitWeight || p.IncludeAssistant != expected.IncludeAssistant {
		return fmt.Errorf("retained legacy prefix differs from immutable cutover authority")
	}
	return nil
}
