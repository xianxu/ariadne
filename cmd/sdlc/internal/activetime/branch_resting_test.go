package activetime

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRecordedPair856RestingClaimTransition(t *testing.T) {
	raw, err := os.ReadFile("testdata/pair856/head.log")
	if err != nil {
		t.Fatal(err)
	}
	at := func(s string) time.Time {
		t.Helper()
		v, e := time.Parse(time.RFC3339, s)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	start, end := at("2026-10-23T21:35:40-07:00"), at("2026-10-24T00:24:00-07:00")
	rows, err := decodeHeadHistory(raw)
	if err != nil {
		t.Fatal(err)
	}
	branch := "000856-cast-tool-one-notes-panel-for-notes-and-sketches-and-one-toolbar-column-with"
	history, err := ParseHeadHistory(raw, branch, rows[len(rows)-1].next, start, end)
	if err != nil {
		t.Fatal(err)
	}
	claims := []BranchClaim{{Start: start, End: end, Issue: "850"}, {Start: start, End: end, Issue: "856"}}
	windows := []ActivityWindow{{Start: start, End: end, Began: start}}
	a, err := SelectBranchActivity("856", windows, history, claims)
	if err != nil {
		t.Fatalf("recorded ordinary planning transition refused: %v", err)
	}
	b, err := SelectBranchActivity("850", windows, history, claims)
	if err != nil {
		t.Fatal(err)
	}
	var credited time.Duration
	for _, result := range []BranchSelection{a, b} {
		var resting time.Duration
		for _, n := range result.Notes {
			if n.Kind == "resting-excluded" {
				resting += n.End.Sub(n.Start)
				if n.Branch != "main-slot4" {
					t.Fatal(n)
				}
			}
		}
		if resting != 119*time.Second {
			t.Fatalf("excluded resting span = %s, want 68s planning + 51s return", resting)
		}
		for _, w := range result.Windows {
			credited += w.End.Sub(w.Start)
		}
	}
	if credited != end.Sub(start)-119*time.Second {
		t.Fatalf("credited %s of %s", credited, end.Sub(start))
	}
	for _, x := range a.Windows {
		for _, y := range b.Windows {
			if x.Start.Before(y.End) && y.Start.Before(x.End) {
				t.Fatal("issues double credited")
			}
		}
	}
	if len(a.Windows) != 2 || !a.Windows[0].Start.Equal(at("2026-10-23T21:36:48-07:00")) || !a.Windows[1].Start.Equal(at("2026-10-24T00:23:20-07:00")) {
		t.Fatalf("856 windows = %+v", a.Windows)
	}
}

func TestSelectBranchActivityRestingDoesNotInferOtherBranches(t *testing.T) {
	start, end := branchTime(0), branchTime(10)
	windows := []ActivityWindow{{Start: start, End: end, Began: start}}
	claims := []BranchClaim{{Start: start, End: end, Issue: "335"}, {Start: start, End: end, Issue: "336"}}
	for _, branch := range []string{"main", "main-slot1", "main-slot4"} {
		got, err := SelectBranchActivity("335", windows, []BranchInterval{{Start: start, End: end, Branch: branch, Known: true}}, claims)
		if err != nil || len(got.Windows) != 0 || len(got.Notes) != 1 || got.Notes[0].Kind != "resting-excluded" {
			t.Fatalf("%s: %+v %v", branch, got, err)
		}
		solo, err := SelectBranchActivity("335", windows, []BranchInterval{{Start: start, End: end, Branch: branch, Known: true}}, claims[:1])
		if err != nil || len(solo.Windows) != 1 || solo.Notes[0].Kind != "ownership-only" {
			t.Fatalf("sole planning changed: %+v %v", solo, err)
		}
	}
	for _, branch := range []BranchInterval{{Start: start, End: end}, {Start: start, End: end, Branch: "main-slot4", Known: false}, {Start: start, End: end, Branch: "feature", Known: true}, {Start: start, End: end, Branch: "main-slot04", Known: true}} {
		if _, err := SelectBranchActivity("335", windows, []BranchInterval{branch}, claims); err == nil || !strings.Contains(err.Error(), "concurrent claims") {
			t.Fatalf("unknown/detached/ambiguous became unassigned: %+v %v", branch, err)
		}
	}
}

func TestRecordedPair812RestingClaimTransition(t *testing.T) {
	raw, err := os.ReadFile("testdata/pair812/head.log")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := decodeHeadHistory(raw)
	if err != nil {
		t.Fatal(err)
	}
	start, _ := time.Parse(time.RFC3339, "2026-10-23T05:01:56-07:00")
	split, _ := time.Parse(time.RFC3339, "2026-10-23T05:02:21-07:00")
	end := split.Add(time.Minute)
	_, after, _ := historyTransition(rows[len(rows)-1].msg)
	history, err := ParseHeadHistory(raw, after.name, rows[len(rows)-1].next, start, end)
	if err != nil {
		t.Fatal(err)
	}
	got, err := SelectBranchActivity("812", []ActivityWindow{{Start: start, End: end, Began: start}}, history, []BranchClaim{{Start: start, End: end, Issue: "812"}, {Start: start, End: end, Issue: "other-held-issue"}})
	if err != nil {
		t.Fatalf("second incident refused: %v", err)
	}
	if len(got.Windows) != 1 || !got.Windows[0].Start.Equal(split) || len(got.Notes) != 2 || got.Notes[0].Kind != "resting-excluded" || got.Notes[0].End.Sub(got.Notes[0].Start) != 25*time.Second {
		t.Fatalf("pair812 = %+v", got)
	}
}

func TestRestingExcludedCountedDisjointPartition(t *testing.T) {
	previous := 0
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			start, end := branchTime(0), branchTime(n)
			claims := []BranchClaim{{Start: start, End: end, Issue: "335"}, {Start: start, End: end, Issue: "336"}}
			var history []BranchInterval
			for i := range n {
				branch := []string{"000335-b", "000336-a", "main-slot1"}[i%3]
				history = append(history, BranchInterval{Start: branchTime(i), End: branchTime(i + 1), Branch: branch, Known: true})
			}
			var selected []ActivityWindow
			for _, id := range []string{"335", "336"} {
				got, err := SelectBranchActivity(id, []ActivityWindow{{Start: start, End: end}}, history, claims)
				if err != nil {
					t.Fatal(err)
				}
				for _, a := range got.Windows {
					for _, b := range selected {
						if a.Start.Before(b.End) && b.Start.Before(a.End) {
							t.Fatal("overlapping credit")
						}
					}
				}
				selected = append(selected, got.Windows...)
				if got.steps > 4*n+12 || previous > 0 && got.steps > 2*previous {
					t.Fatalf("n=%d steps=%d prior=%d", n, got.steps, previous)
				}
				if id == "336" {
					previous = got.steps
				}
			}
			var total time.Duration
			for _, w := range selected {
				total += w.End.Sub(w.Start)
			}
			if total != time.Duration(n-n/3)*time.Minute {
				t.Fatalf("resting time credited: %s", total)
			}
		})
	}
}
