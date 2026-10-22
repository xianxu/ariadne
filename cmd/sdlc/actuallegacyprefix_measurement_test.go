package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func TestJoinLegacyActualResultsClipsBoundaryAndPreservesInputs(t *testing.T) {
	start := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	cut := start.Add(10 * time.Minute)
	segment := func(a, b int, source string) activetime.Segment {
		return activetime.Segment{Start: start.Add(time.Duration(a) * time.Minute), End: start.Add(time.Duration(b) * time.Minute), Source: source, Active: float64(b - a), Alloc: map[string]float64{"335": float64(b - a)}}
	}
	prefix := activetime.Result{Status: activetime.Measured, Segments: []activetime.Segment{segment(0, 15, "shared"), segment(1, 3, "prefix-only")}, SourceIDs: []string{"prefix-only", "shared"}, IssueCommitTimes: []time.Time{cut}}
	suffix := activetime.Result{Status: activetime.Measured, Segments: []activetime.Segment{segment(5, 20, "shared"), segment(10, 10, "boundary")}, SourceIDs: []string{"shared"}, IssueCommitTimes: []time.Time{cut, start.Add(20 * time.Minute)}}
	before := prefix.Segments[0].Alloc["335"]
	got, err := joinLegacyActualResults(prefix, suffix, cut)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalActive != 22 || got.PerIssue["335"] != 22 {
		t.Fatalf("boundary counted twice: %+v", got)
	}
	if len(got.Segments) != 3 || len(got.IssueCommitTimes) != 2 {
		t.Fatalf("zero-boundary segment or duplicate commit retained: %+v", got)
	}
	if prefix.Segments[0].Alloc["335"] != before || !prefix.Segments[0].End.Equal(start.Add(15*time.Minute)) {
		t.Fatal("join mutated component evidence")
	}
	prefix.Segments[0], prefix.Segments[1] = prefix.Segments[1], prefix.Segments[0]
	again, err := joinLegacyActualResults(prefix, suffix, cut)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatalf("input order changed result: %+v %v", again, err)
	}
}

func TestJoinLegacyActualResultsRefusesPartialAndAcceptsKnownEmpty(t *testing.T) {
	cut := time.Date(2026, 10, 14, 0, 10, 0, 0, time.UTC)
	for _, status := range []activetime.Status{0, activetime.TelemetryGap} {
		for _, prefixBad := range []bool{true, false} {
			a, b := activetime.Result{Status: activetime.EmptyWindow}, activetime.Result{Status: activetime.EmptyWindow}
			if prefixBad {
				a.Status = status
			} else {
				b.Status = status
			}
			_, err := joinLegacyActualResults(a, b, cut)
			label := "tracker ownership"
			if prefixBad {
				label = "legacy prefix"
			}
			if err == nil || !strings.Contains(err.Error(), label) {
				t.Fatalf("partial status %v lacks component diagnostic: %v", status, err)
			}
		}
	}
	empty := activetime.Result{Status: activetime.EmptyWindow}
	got, err := joinLegacyActualResults(empty, empty, cut)
	if err != nil || got.Status != activetime.EmptyWindow || got.TotalActive != 0 {
		t.Fatalf("known empty refused: %+v %v", got, err)
	}
	if _, err := joinLegacyActualResults(empty, empty, time.Time{}); err == nil {
		t.Fatal("missing boundary accepted")
	}
	bad := activetime.Result{Status: activetime.Measured, Segments: []activetime.Segment{{Start: cut.Add(-time.Minute), End: cut, Active: math.NaN(), Alloc: map[string]float64{"335": 1}}}}
	if _, err := joinLegacyActualResults(bad, empty, cut); err == nil {
		t.Fatal("nonfinite component accepted")
	}
}

func TestLegacyPrefixMeasurementUsesOneSplitAndDistinctSourceEvidence(t *testing.T) {
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit())
	ownedFixtureAt(t, 10)
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: legacy work")
	legacyHead := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	ownedFixtureAt(t, 20)
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: owned suffix")
	head := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	start := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	cut, end := start.Add(10*time.Minute), start.Add(20*time.Minute)
	source := func(id, cwd string, minutes ...int) transcripts.SessionEvidence {
		file := filepath.Join(t.TempDir(), id+".jsonl")
		var data string
		for _, minute := range minutes {
			data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"user\",\"message\":{\"content\":\"#335\"}}\n", minute)
		}
		if err := os.WriteFile(file, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		return transcripts.SessionEvidence{Harness: "claude", ID: id, CWD: cwd, File: file, Started: start, Digest: fmt.Sprintf("%x", sha256.Sum256([]byte(data))), Length: int64(len(data))}
	}
	shared := source("shared", root, 0, 5, 10, 15, 20)
	brain := source("brain", t.TempDir(), 1, 4)
	legacy := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "repo", CodeHead: legacyHead, Since: start.Format(time.RFC3339), Until: cut.Format(time.RFC3339), GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true},
		Files: []string{shared.File, brain.File}, Sessions: []transcripts.SessionEvidence{shared, brain}, Scope: activetime.Scope{Head: legacyHead, Issue: "335"}, Prefixes: map[string]activetime.SourcePrefix{shared.File: {Bytes: shared.Length, Digest: shared.Digest}, brain.File: {Bytes: brain.Length, Digest: brain.Digest}}, Windows: map[string][]activetime.ActivityWindow{shared.File: {{Start: start, End: cut}}, brain.File: {{Start: start, End: cut}}}}
	legacy.Scope.Issue = ""
	prefixOnly, _, err := measureLegacyPrefix(context.Background(), root, "335", legacy)
	if err != nil || len(prefixOnly.IssueCommitTimes) != 1 {
		t.Fatalf("legacy issue commit timing lost: %+v %v", prefixOnly, err)
	}
	inputs := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodEngagementV1, Repository: "repo", CodeHead: head, TrackerHead: head, Since: start.Format(time.RFC3339), Until: end.Format(time.RFC3339), GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true, LegacyPrefix: &legacy.Snapshot, LegacyCutover: legacyHead, Engagements: []gatestate.MeasurementEngagement{{Start: cut.Format(time.RFC3339), End: end.Format(time.RFC3339), Began: start.Format(time.RFC3339), Roots: []string{root}}}},
		Files: []string{shared.File}, Sessions: []transcripts.SessionEvidence{shared, brain}, Scope: activetime.Scope{Head: head, BranchPoint: legacyHead, Issue: "335"}, Prefixes: map[string]activetime.SourcePrefix{shared.File: {Bytes: shared.Length, Digest: shared.Digest}}, Windows: map[string][]activetime.ActivityWindow{shared.File: {{Start: cut, End: end}}}, SourceSlots: map[string]string{shared.File: "ariadne:1"}, LegacyPrefix: legacy}
	got := computeOwnedActualFor(context.Background(), root, "335", "", inputs)
	if got.Status != actualMeasured || math.Abs(got.Hours-23.0/60) > 1e-9 {
		t.Fatalf("legacy prefix omitted or boundary duplicated: %+v", got)
	}
	if got.Inputs == nil || got.Inputs.LegacyPrefix == nil || len(got.Inputs.LegacyPrefix.Sources) != 2 || len(got.Inputs.Sources) != 1 || got.Inputs.Sources[0].ID != "shared" {
		t.Fatalf("component source evidence conflated: %+v", got.Inputs)
	}
	if !strings.Contains(strings.Join(got.Warnings, "\n"), "legacy repository-plus-brain heuristic: 2026-10-14T00:00:00Z → 2026-10-14T00:10:00Z; tracker ownership: 2026-10-14T00:10:00Z → 2026-10-14T00:20:00Z") {
		t.Fatalf("heuristic population not disclosed: %v", got.Warnings)
	}
	partial := *inputs
	partialChild := *legacy
	partialChild.Sessions = []transcripts.SessionEvidence{shared}
	partial.LegacyPrefix = &partialChild
	partial.Snapshot.LegacyPrefix = &partialChild.Snapshot
	if broken := computeOwnedActualFor(context.Background(), root, "335", "", &partial); broken.Status != actualError {
		t.Fatalf("partial prefix source evidence adopted: %+v", broken)
	}
	replay, err := ownedReplaySources(*got.Inputs, []transcripts.SessionEvidence{shared, brain})
	if err != nil {
		t.Fatal(err)
	}
	replay.Scope.Issue = "335"
	child, err := ownedReplaySources(*got.Inputs.LegacyPrefix, []transcripts.SessionEvidence{shared, brain})
	if err != nil {
		t.Fatal(err)
	}
	child.Scope.Issue = "335"
	replay.LegacyPrefix = child
	again := computeOwnedActualFor(context.Background(), root, "335", "", replay)
	if again.Status != actualMeasured || again.Hours != got.Hours {
		t.Fatalf("component replay changed hours: %+v original=%+v", again, got)
	}
	// A missing prefix file is never silently replaced by the owned suffix.
	if err := os.Remove(brain.File); err != nil {
		t.Fatal(err)
	}
	broken := computeOwnedActualFor(context.Background(), root, "335", "", inputs)
	if broken.Status != actualError {
		t.Fatalf("partial suffix adopted after prefix loss: %+v", broken)
	}
}

func TestLegacyPrefixKnownEmptyRequiresExactEvidenceAndRejectsTelemetryGap(t *testing.T) {
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit())
	head := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	file := filepath.Join(t.TempDir(), "empty.jsonl")
	data := []byte("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"empty\",\"cwd\":\"/repo\"}}\n")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 14, 0, 5, 0, 0, time.UTC)
	end := start.Add(5 * time.Minute)
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	source := transcripts.SessionEvidence{Harness: "codex", ID: "empty", CWD: "/repo", File: file, Started: start, Digest: digest, Length: int64(len(data))}
	inputs := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "repo", CodeHead: head, Since: start.Format(time.RFC3339), Until: end.Format(time.RFC3339), GapMinutes: 15}, Files: []string{file}, Sessions: []transcripts.SessionEvidence{source}, Prefixes: map[string]activetime.SourcePrefix{file: {Bytes: source.Length, Digest: digest}}, Windows: map[string][]activetime.ActivityWindow{file: {{Start: start, End: end}}}, Scope: activetime.Scope{Head: head, Issue: "335"}}
	// Fresh discovery can exceed the retained bound; only inspected sources
	// chosen by the engine form the saved population.
	inputs.Snapshot.Sources = make([]gatestate.MeasurementSource, gatestate.SnapshotLimitSources+1)
	out, saved, err := measureLegacyPrefix(context.Background(), root, "335", inputs)
	if err != nil || out.Status != activetime.EmptyWindow || saved == nil || len(saved.Sources) != 1 {
		t.Fatalf("known empty evidence lost: %+v %+v %v", out, saved, err)
	}
	missing := *inputs
	missing.Files = nil
	if _, _, err := measureLegacyPrefix(context.Background(), root, "335", &missing); err == nil {
		t.Fatal("missing telemetry treated as known empty")
	}
	ownedFixtureAt(t, 7)
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: work without telemetry")
	inputs.Snapshot.CodeHead = strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	inputs.Scope.Head = inputs.Snapshot.CodeHead
	if _, _, err := measureLegacyPrefix(context.Background(), root, "335", inputs); err == nil {
		t.Fatal("prefix telemetry gap admitted as zero")
	}
	// One observed event can prove zero elapsed activity without being a
	// telemetry gap. Retain the inspected bytes for that exact zero too.
	data = append(data, []byte("{\"timestamp\":\"2026-10-14T00:08:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335\"}}\n")...)
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	source.Length, source.Digest = int64(len(data)), fmt.Sprintf("%x", sha256.Sum256(data))
	inputs.Sessions = []transcripts.SessionEvidence{source}
	inputs.Prefixes[file] = activetime.SourcePrefix{Bytes: source.Length, Digest: source.Digest}
	out, saved, err = measureLegacyPrefix(context.Background(), root, "335", inputs)
	if err != nil || out.TotalActive != 0 || saved == nil || len(saved.Sources) != 1 {
		t.Fatalf("known isolated zero lost: %+v %+v %v", out, saved, err)
	}

}
