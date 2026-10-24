package activetime

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Replay the checkout/merge/return reported by pair:6 in pair#830. Object
// identities are synthetic: the peer supplied abbreviated IDs, not raw records.
func TestPair830StagingReplayExcludes172Seconds(t *testing.T) {
	start := time.Date(2026, 10, 24, 2, 54, 2, 0, time.UTC)
	enter, leave, end := start.Add(time.Minute), start.Add(232*time.Second), start.Add(292*time.Second)
	var raw strings.Builder
	previous := strings.Repeat("0", 40)
	for _, row := range []struct {
		at  time.Time
		msg string
	}{
		{start, "commit: work"},
		{enter, "checkout: moving from 000849-cast to stage/000849-000850"},
		{enter, "merge origin/000850-cast: Merge made by the 'ort' strategy."},
		{leave, "checkout: moving from stage/000849-000850 to 000849-cast"},
	} {
		head := strings.Repeat("a", 40)
		fmt.Fprintf(&raw, "%s %s Fixture <fixture@example.invalid> %d +0000\t%s\n", previous, head, row.at.Unix(), row.msg)
		previous = head
	}
	history, err := ParseHeadHistory([]byte(raw.String()), "000849-cast", previous, start, end)
	if err != nil {
		t.Fatal(err)
	}
	claims := []BranchClaim{{Start: start, End: end, Issue: "849"}, {Start: start, End: end, Issue: "830"}}
	for _, id := range []string{"849", "830"} {
		got, err := SelectBranchActivity(id, []ActivityWindow{{Start: start, End: end, Began: start}}, history, claims)
		if err != nil {
			t.Fatalf("pair#%s refused witnessed stage: %v", id, err)
		}
		var excluded, credited time.Duration
		for _, note := range got.Notes {
			if note.Kind == "stage-excluded" {
				if note.Branch != "stage/000849-000850" || !note.Start.Equal(enter) || !note.End.Equal(leave) {
					t.Fatal(note)
				}
				excluded += note.End.Sub(note.Start)
			}
		}
		for _, w := range got.Windows {
			credited += w.End.Sub(w.Start)
		}
		want := time.Duration(0)
		if id == "849" {
			want = 120 * time.Second
		}
		if excluded != 172*time.Second || credited != want {
			t.Fatalf("%s: excluded=%s credited=%s", id, excluded, credited)
		}
	}
}

func TestStagingSelectionRequiresAttachedValidWitness(t *testing.T) {
	start, end := branchTime(0), branchTime(10)
	windows := []ActivityWindow{{Start: start, End: end, Began: start}}
	claims := []BranchClaim{{Start: start, End: end, Issue: "335"}, {Start: start, End: end, Issue: "336"}}
	for _, name := range []string{"stage/combined", "stage/000849-000850", "stage/team/nested"} {
		history := []BranchInterval{{Start: start, End: end, Branch: name, Known: true}}
		got, err := SelectBranchActivity("335", windows, history, claims)
		if err != nil || len(got.Windows) != 0 || len(got.Notes) != 1 || got.Notes[0].Kind != "stage-excluded" {
			t.Fatalf("%s: %+v %v", name, got, err)
		}
		solo, err := SelectBranchActivity("335", windows, history, claims[:1])
		if err != nil || len(solo.Windows) != 1 || solo.Notes[0].Kind != "ownership-only" {
			t.Fatalf("solo: %+v %v", solo, err)
		}
	}
	for _, name := range []string{"", "HEAD", "stage", "stage/", "stages/combined", "feature", "stage/a..b", "stage/a.lock", "stage/a//b", "stage/.hidden", "stage/a b", "stage/a\x01b", "refs/heads/stage/a"} {
		got, err := SelectBranchActivity("335", windows, []BranchInterval{{Start: start, End: end, Branch: name, Known: true}}, claims)
		assertUnknownOnly(t, got, err, start, end)
	}
	got, err := SelectBranchActivity("335", windows, []BranchInterval{{Start: start, End: end, Branch: "stage/combined", Known: false}}, claims)
	assertUnknownOnly(t, got, err, start, end)
}
