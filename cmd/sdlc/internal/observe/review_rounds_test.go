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
	first := sidecar("REWORK", "base.."+strings.Repeat("b", 40))
	second := "\n---\n\n## Re-review — then (SHIP)\n\n" + strings.TrimPrefix(sidecar("SHIP", "base.."+head), "# Boundary Review\n\n")
	third := "\n---\n\n## Re-review — later (REWORK)\n\n" + strings.TrimPrefix(sidecar("REWORK", "base.."+strings.Repeat("c", 40)), "# Boundary Review\n\n")
	for _, tc := range []struct {
		name, status, text, verdict, window string
		round, history                      int
	}{
		{"latest", "working", first + second, "SHIP", "base.." + head, 2, 1},
		{"stale absent", "working", first + second, "SHIP", "base.." + head, 2, 1},
		{"bound", "codecomplete", first + second + third, "SHIP", "base.." + head, 2, 2},
		{"stale bound", "codecomplete", first + second + third, "SHIP", "base.." + head, 2, 2},
		{"latest can rework", "working", first + second + third, "REWORK", "base.." + strings.Repeat("c", 40), 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := base(card(t, tc.status, &me, false))
			if strings.HasPrefix(tc.name, "stale") {
				in.TrackerStale = true
				in.TrackerErr = errors.New("offline")
			}
			in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: tc.text, Rounds: []ReviewRound{{Number: 1, Head: strings.Repeat("b", 40)}, {Number: 2, Head: head}, {Number: 3, Head: strings.Repeat("c", 40)}}}}}
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
	in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: sidecar("SHIP", "base.."+head) + quoted, Rounds: []ReviewRound{{Number: 1, Head: head}, {Number: 2, Head: strings.Repeat("d", 40)}}}}}
	r, _ := reviewFor(Assemble(in), "close")
	if r.Verdict != "SHIP" || r.Round != 1 {
		t.Fatalf("quoted prose overrode metadata: %+v", r)
	}
	a := in.Evidence.Artifacts["close"]
	a.Sidecar += "\n---\n\n## Re-review — interrupted\n\n| window | base..dddddddddddddddddddddddddddddddddddddddd |\n"
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
	second := "\n---\n\n## Re-review — later\n\n" + strings.TrimPrefix(strings.ReplaceAll(sidecar("REWORK", "base.."+strings.Repeat("c", 40)), "08:30:00", "09:30:00"), "# Boundary Review\n\n")
	in.Evidence = Evidence{Source: "branch:plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: first + second, OpenBlocking: 1, Rounds: []ReviewRound{{Number: 4, Head: strings.Repeat("a", 40)}, {Number: 6, Head: strings.Repeat("c", 40)}}}}}
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
	first := sidecar("REWORK", "base.."+strings.Repeat("b", 40))
	second := "\n---\n\n## Re-review — same second\n\n" + strings.TrimPrefix(sidecar("SHIP", "base.."+strings.Repeat("d", 40)), "# Boundary Review\n\n")
	r := review("close", ArtifactFacts{Found: true, Sidecar: first + second, Rounds: []ReviewRound{{Number: 4, Head: strings.Repeat("b", 40)}, {Number: 6, Head: strings.Repeat("d", 40)}}}, false, "plans")
	if r.Round != 6 || len(r.History) != 1 || r.History[0].Round != 4 {
		t.Fatalf("duplicate timestamp conflated rounds: %+v", r)
	}
}

func TestReviewRoundsTimestampDrift(t *testing.T) {
	in := base(card(t, "codecomplete", &me, false))
	first := sidecar("REWORK", "base.."+strings.Repeat("b", 40))
	second := "\n---\n\n## Re-review — later\n\n" + strings.TrimPrefix(strings.ReplaceAll(sidecar("SHIP", "base.."+strings.Repeat("a", 40)), "08:30:00", "09:30:00"), "# Boundary Review\n\n")
	in.Evidence = Evidence{Source: "plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: first + second, Rounds: []ReviewRound{{Number: 4, Head: strings.Repeat("b", 40)}, {Number: 6, Head: strings.Repeat("a", 40)}}}}}
	r, _ := reviewFor(Assemble(in), "close")
	if r.Round != 6 || len(r.History) != 1 || r.History[0].Round != 4 || r.History[0].Relation != "earlier" {
		t.Fatalf("timestamp drift lost stable round identity/order: %+v", r)
	}
}

func TestReviewRoundAbsentLedgerStaysUnknown(t *testing.T) {
	in := base(card(t, "working", &me, false))
	in.Evidence = Evidence{Source: "plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: sidecar("SHIP", "base.."+strings.Repeat("a", 40))}}}
	o := Assemble(in)
	r, _ := reviewFor(o, "close")
	raw, _ := json.Marshal(r)
	if r.Verdict != "SHIP" || r.Round != 0 || !strings.Contains(string(raw), "round_error") {
		t.Fatalf("invented round without ledger: %s", raw)
	}
	var b bytes.Buffer
	RenderText(&b, o)
	if !strings.Contains(b.String(), "round unknown") {
		t.Fatalf("text hides unknown round: %s", b.String())
	}
}

func TestAssignReviewRounds(t *testing.T) {
	a, b := strings.Repeat("a", 40), strings.Repeat("b", 40)
	for _, tc := range []struct {
		name   string
		heads  []string
		rounds []ReviewRound
		want   []int
	}{
		{"anchored legacy", []string{b, a}, []ReviewRound{{Number: 4}, {Number: 6, Head: a}}, []int{4, 6}},
		{"incompatible legacy", []string{a, b}, []ReviewRound{{Number: 4}, {Number: 6, Head: a}}, []int{0, 0}},
		{"unanchored legacy", []string{b, a}, []ReviewRound{{Number: 4}, {Number: 6}}, []int{0, 0}},
		{"partial legacy", []string{a}, []ReviewRound{{Number: 4}, {Number: 6, Head: b}}, []int{0}},
		{"complete duplicate group", []string{a, a}, []ReviewRound{{Number: 4, Head: a}, {Number: 6, Head: a}}, []int{4, 6}},
		{"missing duplicate", []string{a}, []ReviewRound{{Number: 4, Head: a}, {Number: 6, Head: a}}, []int{0}},
		{"extra duplicate", []string{a, a}, []ReviewRound{{Number: 6, Head: a}}, []int{0, 0}},
		{"malformed", []string{"garbage"}, []ReviewRound{{Number: 6, Head: "garbage"}}, []int{0}},
		{"conflicting", []string{a}, []ReviewRound{{Number: 6, Head: a, Error: "conflict"}}, []int{0}},
		{"bad number", []string{a}, []ReviewRound{{Number: 0, Head: a}}, []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries := make([]*Review, len(tc.heads))
			for i, h := range tc.heads {
				entries[i] = &Review{Window: "base.." + h}
			}
			assignReviewRounds(entries, tc.rounds)
			for i, e := range entries {
				if e.Round != tc.want[i] || (e.Round == 0) != (e.RoundError != "") {
					t.Fatalf("entry %d: %+v, want %d", i, e, tc.want[i])
				}
			}
		})
	}
}

// Exhaustively omit entries and ledger identities from a known history. A
// missing duplicate must never acquire an arbitrary number; unique identities
// retain their own provenance, and a fully aligned legacy sequence can recover.
func TestAssignReviewRoundsGeneratedPartialHistories(t *testing.T) {
	heads := []string{strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("a", 40)}
	for entryMask := 0; entryMask < 8; entryMask++ {
		for roundMask := 0; roundMask < 8; roundMask++ {
			var entries []*Review
			var rounds []ReviewRound
			countsE, countsR := map[string]int{}, map[string]int{}
			for i, h := range heads {
				if entryMask&(1<<i) != 0 {
					entries = append(entries, &Review{Window: "base.." + h})
					countsE[h]++
				}
				if roundMask&(1<<i) != 0 {
					rounds = append(rounds, ReviewRound{Number: 4 + i*2, Head: h})
					countsR[h]++
				}
			}
			assignReviewRounds(entries, rounds)
			for _, e := range entries {
				h := reviewHead(e.Window)
				if countsE[h] != countsR[h] && e.Round != 0 {
					t.Fatalf("ambiguous duplicate mapped: masks %d/%d %+v", entryMask, roundMask, e)
				}
				if e.Round > 0 {
					found := false
					for _, r := range rounds {
						found = found || (r.Head == h && r.Number == e.Round)
					}
					if !found {
						t.Fatalf("invented identity: %+v", e)
					}
				}
			}
		}
	}
}

func TestReviewHistoryLabelsUseEntryOrder(t *testing.T) {
	in := base(card(t, "codecomplete", &me, false))
	first := sidecar("REWORK", "base.."+strings.Repeat("b", 40))
	second := "\n## Re-review — later\n" + sidecar("SHIP", "base.."+strings.Repeat("a", 40))
	in.Evidence = Evidence{Source: "plans", Details: details(""), Artifacts: map[string]ArtifactFacts{"close": {Found: true, Sidecar: first + second, Rounds: []ReviewRound{{Number: 6, Head: strings.Repeat("b", 40)}, {Number: 4, Head: strings.Repeat("a", 40)}}}}}
	r, _ := reviewFor(Assemble(in), "close")
	if len(r.History) != 1 || r.History[0].Relation != "earlier" {
		t.Fatalf("numeric order reversed chronology: %+v", r)
	}
}
