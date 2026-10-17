package observe

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestReviewRoundsSelectCompletionOrLatest(t *testing.T) {
	head := strings.Repeat("a", 40)
	first := sidecar("REWORK", "base..old")
	second := "\n---\n\n## Re-review — then (SHIP)\n\n" + strings.TrimPrefix(sidecar("SHIP", "base.."+head), "# Boundary Review\n\n")
	third := "\n---\n\n## Re-review — later (REWORK)\n\n" + strings.TrimPrefix(sidecar("REWORK", "base..later"), "# Boundary Review\n\n")
	for _, tc := range []struct {
		name, status, text, verdict, window string
		round, history                      int
	}{
		{"latest", "working", first + second, "SHIP", "base.." + head, 2, 1},
		{"stale absent", "working", first + second, "SHIP", "base.." + head, 2, 1},
		{"bound", "codecomplete", first + second + third, "SHIP", "base.." + head, 2, 2},
		{"stale bound", "codecomplete", first + second + third, "SHIP", "base.." + head, 2, 2},
		{"latest can rework", "working", first + second + third, "REWORK", "base..later", 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := base(card(t, tc.status, &me, false))
			if strings.HasPrefix(tc.name, "stale") {
				in.TrackerStale = true
				in.TrackerErr = errors.New("offline")
			}
			in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: tc.text}}}
			o := Assemble(in)
			r, _ := reviewFor(o, "close")
			if r.Verdict != tc.verdict || r.Window != tc.window {
				t.Fatalf("selected stale or unbound review: %+v", r)
			}
			raw, _ := json.Marshal(r)
			var fields struct {
				Round   int               `json:"round"`
				History []json.RawMessage `json:"history"`
			}
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if fields.Round != tc.round || len(fields.History) != tc.history {
				t.Fatalf("round/history absent: %s", raw)
			}
			var b bytes.Buffer
			RenderText(&b, o)
			if !strings.Contains(b.String(), "round 2") || !strings.Contains(b.String(), "earlier round 1: REWORK") {
				t.Fatalf("text lost round history: %s", b.String())
			}
		})
	}
}

func TestReviewRoundsNeverSubstituteUnboundEvidence(t *testing.T) {
	in := base(card(t, "codecomplete", &me, false))
	in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: sidecar("SHIP", "base..wrong")}}}
	r, _ := reviewFor(Assemble(in), "close")
	if r.State != Unknown || r.Verdict != "" || !strings.Contains(r.Error, "completion") {
		t.Fatalf("unbound review presented as current: %+v", r)
	}
}

func TestReviewRoundsMetadataAndMissingLatest(t *testing.T) {
	head := strings.Repeat("a", 40)
	quoted := "\n```markdown\n## Re-review — quoted (REWORK)\n\n| verdict | REWORK |\n```\n\n| verdict | REWORK |\n"
	in := base(card(t, "working", &me, false))
	in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: sidecar("SHIP", "base.."+head) + quoted}}}
	r, _ := reviewFor(Assemble(in), "close")
	if r.Verdict != "SHIP" || r.Round != 1 {
		t.Fatalf("quoted prose overrode metadata: %+v", r)
	}
	a := in.Evidence.Artifacts["close"]
	a.Sidecar += "\n---\n\n## Re-review — interrupted\n\n| window | base..new |\n"
	in.Evidence.Artifacts["close"] = a
	r, _ = reviewFor(Assemble(in), "close")
	if r.State != Unknown || r.Verdict != "" || r.Round != 2 {
		t.Fatalf("missing latest became earlier success: %+v", r)
	}
	in.Card = card(t, "codecomplete", &me, false)
	r, _ = reviewFor(Assemble(in), "close")
	if r.State != Present || r.Verdict != "SHIP" || r.Round != 1 || len(r.History) != 1 || r.History[0].State != Unknown {
		t.Fatalf("completion did not retain its bound review and unknown history: %+v", r)
	}
}

func TestReviewRoundsKeepLedgerNumberAndCurrentFindings(t *testing.T) {
	in := base(card(t, "codecomplete", &me, false))
	first := sidecar("SHIP", "base.."+strings.Repeat("a", 40))
	second := "\n---\n\n## Re-review — later\n\n" + strings.TrimPrefix(strings.ReplaceAll(sidecar("REWORK", "base..later"), "08:30:00", "09:30:00"), "# Boundary Review\n\n")
	in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: first + second, OpenBlocking: 1, Rounds: map[string][]int{"2026-10-02T08:30:00-07:00": {4}, "2026-10-02T09:30:00-07:00": {6}}}}}
	r, _ := reviewFor(Assemble(in), "close")
	if r.Round != 4 || r.OpenBlocking != 1 || len(r.History) != 1 || r.History[0].Round != 6 || r.History[0].Relation != "later, not completion-bound" {
		t.Fatalf("lost ledger provenance/current findings: %+v", r)
	}
	raw, err := json.Marshal(Assemble(in))
	if err != nil {
		t.Fatal(err)
	}
	var back Observation
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
}

func TestReviewRoundsKeepDistinctNumbersWithinOneSecond(t *testing.T) {
	first := sidecar("REWORK", "base..old")
	second := "\n---\n\n## Re-review — same second\n\n" + strings.TrimPrefix(sidecar("SHIP", "base..new"), "# Boundary Review\n\n")
	r := review("close", ArtifactFacts{Found: true, Sidecar: first + second, Rounds: map[string][]int{"2026-10-02T08:30:00-07:00": {4, 6}}}, false, "plans")
	if r.Round != 6 || len(r.History) != 1 || r.History[0].Round != 4 {
		t.Fatalf("duplicate timestamp conflated rounds: %+v", r)
	}
}
