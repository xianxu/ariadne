package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func TestOwnedActualEngineReplayAndExcludedBrainProvenance(t *testing.T) {
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-14T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-14T00:00:00Z")
	root := testfix.Repo(t, testfix.InitialCommit())
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-14T00:30:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-14T00:30:00Z")
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: measured work")
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	start := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	source := func(id, cwd string) transcripts.SessionEvidence {
		file := filepath.Join(t.TempDir(), id+".jsonl")
		data := []byte("{\"timestamp\":\"2026-10-14T00:05:00Z\",\"type\":\"user\",\"message\":{\"content\":\"#335\"}}\n{\"timestamp\":\"2026-10-14T00:15:00Z\",\"type\":\"user\",\"message\":{\"content\":\"#335\"}}\n")
		if err := os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
		return transcripts.SessionEvidence{Harness: "claude", ID: id, CWD: cwd, File: file, Digest: fmt.Sprintf("%x", sha256.Sum256(data)), Length: int64(len(data)), Started: start}
	}
	owner := source("owner", root)
	brain := source("unbound", t.TempDir())
	m := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "repo", CodeHead: head,
		Since: start.Format(time.RFC3339), Until: start.Add(30 * time.Minute).Format(time.RFC3339), BrainRoot: brain.CWD, GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true}
	in := &ownedActualInputs{Snapshot: m, Files: []string{owner.File}, Prefixes: map[string]activetime.SourcePrefix{owner.File: {Bytes: owner.Length, Digest: owner.Digest}},
		Windows: map[string][]activetime.ActivityWindow{owner.File: {{Start: start, End: start.Add(30 * time.Minute)}}}, Scope: activetime.Scope{Head: head, Issue: "335"}, Sessions: []transcripts.SessionEvidence{owner}, ExcludedBrain: []transcripts.SessionEvidence{brain}}
	result := computeOwnedActualFor(context.Background(), root, "335", "", in)
	if result.Status != actualMeasured || math.Abs(result.Hours-1.0/6) > .00001 {
		t.Fatalf("result=%+v", result)
	}
	if result.Inputs == nil || result.Inputs.ExcludedBrainSessions != 1 || result.Inputs.ExcludedBrainMinutes != 10 || len(result.Inputs.Sources) != 1 {
		t.Fatalf("invisible excluded evidence: %+v", result.Inputs)
	}
	// New HEAD and appended events must not change replay's pinned code or bytes.
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-14T00:10:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-14T00:10:00Z")
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#999: another task")
	f, err := os.OpenFile(owner.File, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("{\"timestamp\":\"2026-10-14T00:25:00Z\",\"type\":\"user\",\"message\":{\"content\":\"#335\"}}\n")
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ownedReplaySources(*result.Inputs, []transcripts.SessionEvidence{owner})
	if err != nil {
		t.Fatal(err)
	}
	replay.Scope.Issue = "335"
	again := computeOwnedActualFor(context.Background(), root, "335", "", replay)
	if again.Status != actualMeasured || again.Hours != result.Hours || again.Inputs.ExcludedBrainMinutes != 10 {
		t.Fatalf("changed replay: %+v original=%+v", again, result)
	}
	raw, err := os.ReadFile(owner.File)
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = ' '
	if err := os.WriteFile(owner.File, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if changed := computeOwnedActualFor(context.Background(), root, "335", "", replay); changed.Status != actualError {
		t.Fatalf("changed prefix auto-adopted: %+v", changed)
	}
}

func TestActualQueueMeasurementReplay(t *testing.T) {
	t.Setenv("GIT_AUTHOR_DATE", "2026-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-01-01T00:00:00Z")
	root := testfix.Repo(t, testfix.InitialCommit())
	t.Setenv("GIT_AUTHOR_DATE", "2026-01-01T00:40:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-01-01T00:40:00Z")
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: queue work")
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	file := filepath.Join(t.TempDir(), "owner.jsonl")
	data := []byte(strings.Join([]string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"message","role":"user","content":"#335 start"}}`,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"suite","arguments":"{\"cmd\":\"sdlc local-check run --name suite\"}"}}`,
		`{"timestamp":"2026-01-01T00:25:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"done"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":"done"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"response_item","payload":{"type":"message","role":"user","content":"#335 done"}}`,
	}, "\n") + "\n")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	s := transcripts.SessionEvidence{Harness: "codex", ID: "owner", CWD: root, File: file, Digest: fmt.Sprintf("%x", sha256.Sum256(data)), Length: int64(len(data)), Started: start}
	brainFile := filepath.Join(t.TempDir(), "brain.jsonl")
	brainData := []byte("{\"timestamp\":\"2026-01-01T00:02:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335\"}}\n{\"timestamp\":\"2026-01-01T00:08:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335\"}}\n")
	if err := os.WriteFile(brainFile, brainData, 0600); err != nil {
		t.Fatal(err)
	}
	brain := transcripts.SessionEvidence{Harness: "codex", ID: "brain", CWD: t.TempDir(), File: brainFile, Digest: fmt.Sprintf("%x", sha256.Sum256(brainData)), Length: int64(len(brainData)), Started: start}
	m := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodQueueV1, Repository: "x/ariadne", CodeHead: head, TrackerHead: head, Since: start.Format(time.RFC3339), Until: start.Add(time.Hour).Format(time.RFC3339), GapMinutes: 120, CommitWeight: 1, IncludeAssistant: true,
		Engagements: []gatestate.MeasurementEngagement{{Start: start.Format(time.RFC3339), End: start.Add(time.Hour).Format(time.RFC3339), Began: start.Format(time.RFC3339), Roots: []string{root}}},
		Cuts: []gatestate.MeasurementCut{
			{Dispatched: start.Add(20 * time.Minute).Format(time.RFC3339), Finalized: start.Add(30 * time.Minute).Format(time.RFC3339)},
			{Dispatched: start.Add(50 * time.Minute).Format(time.RFC3339), Finalized: start.Add(time.Hour).Format(time.RFC3339), Finalizing: true},
		},
		QueueWaits: []gatestate.MeasurementQueueWait{{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: start.Add(10 * time.Minute).Format(time.RFC3339), End: start.Add(50 * time.Minute).Format(time.RFC3339), QueueSeconds: 2400, RawSeconds: 2400, Exact: true}}}
	in := &ownedActualInputs{Snapshot: m, Files: []string{file}, Prefixes: map[string]activetime.SourcePrefix{file: {Bytes: s.Length, Digest: s.Digest}}, Windows: map[string][]activetime.ActivityWindow{file: {{Start: start, End: start.Add(time.Hour), Began: start}}}, SourceSlots: map[string]string{file: "ariadne:1"}, Scope: activetime.Scope{Head: head, Issue: "335"}, TrackerRefs: []string{head}, Sessions: []transcripts.SessionEvidence{s}, ExcludedBrain: []transcripts.SessionEvidence{brain}}
	got := computeOwnedActualFor(context.Background(), root, "335", "", in)
	if got.Status != actualMeasured || got.Inputs == nil || len(got.Inputs.QueueWaits) != 1 || got.Inputs.Sources[0].Slot != "ariadne:1" {
		t.Fatalf("queue candidate: %+v", got)
	}
	if got.Split == nil || math.Abs(got.Split.Build-10.0/60) > .00001 || math.Abs(got.Split.Converge-5.0/60) > .00001 || math.Abs(got.Split.Tail-10.0/60) > .00001 || math.Abs(got.Hours-.25) > .00001 {
		t.Fatalf("queue crossing phase cuts did not partition build/converge/tail: %+v", got)
	}
	replay, err := ownedReplaySources(*got.Inputs, []transcripts.SessionEvidence{s})
	if err != nil {
		t.Fatal(err)
	}
	replay.Scope.Issue = "335"
	again := computeOwnedActualFor(context.Background(), root, "335", "", replay)
	if again.Status != actualMeasured || again.Hours != got.Hours || len(again.Inputs.QueueWaits) != 1 {
		t.Fatalf("queue replay changed: %+v candidate=%+v", again, got)
	}
}

func TestOwnedActualProductionConsumersShareBrainAndOwnership(t *testing.T) {
	r, _, _, _, path := ownedResolverFixture(t)
	appendEvents := func(path string, minutes ...int) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		for _, minute := range minutes {
			if _, err := fmt.Fprintf(f, "{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":[{\"type\":\"input_text\",\"text\":\"#335\"}]}}\n", minute); err != nil {
				t.Fatal(err)
			}
		}
	}
	appendEvents(path, 2, 8)
	brain := t.TempDir()
	brainSource := filepath.Join(filepath.Dir(path), "brain.jsonl")
	raw := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"brain\",\"cwd\":%q}}\n", brain)
	if err := os.WriteFile(brainSource, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	appendEvents(brainSource, 1, 9)
	measured := computeActualFor(context.Background(), r.root, brain, "335", "", "")
	if measured.Status != actualMeasured || math.Abs(measured.Hours-.1) > .00001 {
		t.Fatalf("measured=%+v", measured)
	}
	var output bytes.Buffer
	command := NewActualCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--issue", "335", "--brain-dir", brain})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "0.10h") {
		t.Fatalf("actual command disagrees: %s", output.String())
	}
	for _, milestone := range []string{"", "M1"} {
		f := (&milestoneCloseFlags{Issue: 335, Milestone: milestone, BrainDir: brain}).closeFlags()
		result, ok := adoptOmittedActual(&bytes.Buffer{}, f, "335", "issue")
		if !ok || f.Actual != "0.10" || result.Inputs == nil || result.Inputs.BrainRoot != brain || result.Inputs.CodeHead != measured.Inputs.CodeHead || result.Inputs.ExcludedBrainSessions != 1 || result.Inputs.ExcludedBrainMinutes != 8 {
			t.Fatalf("consumer %q lost shared inputs: result=%+v flags=%+v", milestone, result, f)
		}
	}
}
