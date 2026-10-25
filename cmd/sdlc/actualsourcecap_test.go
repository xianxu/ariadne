package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// writeOwnedSessions adds sessions under the fixture owner's worktree: active
// ones carry two user messages a minute apart, idle ones only their metadata.
// pad lengthens each session id, growing the snapshot without more sources.
func writeOwnedSessions(t *testing.T, cwd string, active, idle, pad int) {
	t.Helper()
	dir := filepath.Join(os.Getenv("HOME"), ".codex", "sessions")
	for i := 0; i < active+idle; i++ {
		id := fmt.Sprintf("cap%03d%s", i, strings.Repeat("x", pad))
		data := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", id, cwd)
		if i < active {
			at := time.Date(2026, 10, 14, 0, 2, 0, 0, time.UTC).Add(time.Duration(i%60) * time.Second)
			for _, ts := range []time.Time{at, at.Add(time.Minute)} {
				data += fmt.Sprintf("{\"timestamp\":%q,\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335 step %d\"}}\n", ts.Format(time.RFC3339), i)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("cap%03d.jsonl", i)), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// pair#838 and pair#851 refused close with "measurement branch proof:
// measurement sources exceed snapshot limit 128": the proof checked the bound
// on untrimmed sources. Both closes are reconstructed in shape (their raw
// transcripts are not in this repository): over 128 admitted sources, of which
// at most 128 contribute (#838), or more than 128 contribute (#851). pair#866
// refused under the source cap on bytes ("measurement snapshot is 182432 bytes,
// over the version 4 bound of 131072 (119 sources, ...)"); its shape is 119
// contributing sources whose evidence outgrows the byte bound.
func TestSourceCapMeasuresInsteadOfRefusing(t *testing.T) {
	for _, tc := range []struct {
		name         string
		active, idle int
		pad          int
		summarised   bool
		reason       string
	}{
		{name: "pair#838: over the cap admitted, under it contributing", active: 40, idle: 100},
		{name: "pair#851: over the cap contributing", active: gatestate.SnapshotLimitSources + 2, summarised: true, reason: "snapshot limit 128"},
		{name: "pair#866: under the cap, over the byte bound", active: 119, pad: 600, summarised: true, reason: "over the version 4 bound of 131072 (119 sources"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, _, owner, _ := ownedResolverFixture(t)
			writeOwnedSessions(t, owner.Worktree, tc.active, tc.idle, tc.pad)
			res := computeActualForPurpose(context.Background(), r.root, "", "335", "", "", actualCloseCandidate)
			if res.Status != actualMeasured || res.Inputs == nil {
				t.Fatalf("source cap refused the measurement: %v %s", res.Status, res.Detail)
			}
			if c := classifyActual(res); c.Level != confidenceExact {
				t.Fatalf("confidence = %s, want exact: the hours were all attributed", c)
			}
			if got := gatestate.IsSummaryMethod(res.Inputs.Method); got != tc.summarised {
				t.Fatalf("summarised = %v, want %v (%d sources retained)", got, tc.summarised, len(res.Inputs.Sources))
			}
			if err := res.Inputs.Validate(); err != nil {
				t.Fatalf("retained evidence is invalid: %v", err)
			}
			if !tc.summarised {
				if len(res.Inputs.Sources) > gatestate.SnapshotLimitSources {
					t.Fatalf("retained %d sources", len(res.Inputs.Sources))
				}
				return
			}
			s := res.Inputs.Summary
			if s.Sources != tc.active || s.Hours != res.Hours || len(s.Credited) == 0 || !strings.Contains(s.Reason, tc.reason) {
				t.Fatalf("summary = %+v, hours %.4f", s, res.Hours)
			}
			if _, err := resolveOwnedReplay(context.Background(), r.root, "335", *res.Inputs, gatestate.Ledger{}); err == nil || !strings.Contains(err.Error(), "not replayable") {
				t.Fatalf("a summarised close replayed: %v", err)
			}
			credited, widened, err := summaryCredits("335", *res.Inputs, res.Hours)
			if err != nil || len(credited) == 0 || len(widened) != 0 {
				t.Fatalf("summary credits = %d, widened %d, %v", len(credited), len(widened), err)
			}
			if _, _, err := summaryCredits("335", *res.Inputs, res.Hours+0.1); err == nil {
				t.Fatal("a summary disagreeing with the adopted hours was trusted")
			}
		})
	}
}
