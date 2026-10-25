package activetime

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

// assertUnknownOnly pins #603: activity under concurrent claims with an
// unknown, detached or ambiguous branch is recorded in Unknown, never counted.
func assertUnknownOnly(t *testing.T, got BranchSelection, err error, start, end time.Time) {
	t.Helper()
	if err != nil || len(got.Windows) != 0 || len(got.Unknown) != 1 || !got.Unknown[0].Start.Equal(start) || !got.Unknown[0].End.Equal(end) {
		t.Fatalf("unknown branch activity = %+v %v, want only Unknown %s → %s", got, err, start, end)
	}
}

func branchTime(minute int) time.Time {
	return time.Date(2026, 10, 14, 0, minute, 0, 0, time.UTC)
}

func TestSelectBranchActivityPartitionsConcurrentReservations(t *testing.T) {
	window := ActivityWindow{Start: branchTime(2), End: branchTime(31), Began: branchTime(2)}
	claims := []BranchClaim{{Start: branchTime(1), End: branchTime(31), Issue: "336"}, {Start: branchTime(2), End: branchTime(31), Issue: "335"}}
	history := []BranchInterval{{Start: branchTime(0), End: branchTime(20), Branch: "000336-a", Known: true}, {Start: branchTime(20), End: branchTime(31), Branch: "000335-b", Known: true}}
	got, err := SelectBranchActivity("335", []ActivityWindow{window}, history, claims)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Windows) != 1 || !got.Windows[0].Start.Equal(branchTime(20)) || !got.Windows[0].End.Equal(branchTime(31)) || !got.Windows[0].Began.Equal(window.Began) {
		t.Fatalf("B windows = %+v", got.Windows)
	}
	for _, branch := range []BranchInterval{{Start: branchTime(0), End: branchTime(31)}} {
		got, err := SelectBranchActivity("335", []ActivityWindow{window}, []BranchInterval{branch}, claims)
		assertUnknownOnly(t, got, err, window.Start, window.End)
	}
	solo, err := SelectBranchActivity("335", []ActivityWindow{window}, nil, claims[1:])
	if err != nil || len(solo.Windows) != 1 || !solo.Windows[0].Start.Equal(window.Start) || len(solo.Notes) != 1 || solo.Notes[0].Kind != "ownership-only" {
		t.Fatalf("sole claim planning = %+v, %v", solo, err)
	}
}

func TestSelectBranchActivityNoClaimCannotCredit(t *testing.T) {
	_, err := SelectBranchActivity("335", []ActivityWindow{{Start: branchTime(2), End: branchTime(31)}}, []BranchInterval{{Start: branchTime(0), End: branchTime(31), Branch: "000335-b", Known: true}}, nil)
	if err == nil {
		t.Fatal("a branch alone supplied ownership")
	}
}

func TestSelectBranchActivityDoesNotReassignAnUnclaimedForeignBranch(t *testing.T) {
	window := ActivityWindow{Start: branchTime(2), End: branchTime(31), Began: branchTime(2)}
	got, err := SelectBranchActivity("335", []ActivityWindow{window}, []BranchInterval{{Start: branchTime(2), End: branchTime(31), Branch: "000336-a", Known: true}}, []BranchClaim{{Start: branchTime(2), End: branchTime(31), Issue: "335"}})
	if err != nil || len(got.Windows) != 0 || len(got.Notes) != 1 || got.Notes[0].Kind != "foreign-branch" {
		t.Fatalf("sole reservation absorbed a witnessed foreign branch: %+v %v", got, err)
	}
}

func TestSelectBranchActivityTerminalOwnerIsNotAnotherReservation(t *testing.T) {
	windows := []ActivityWindow{{Start: branchTime(2), End: branchTime(31), Began: branchTime(2)}}
	claims := []BranchClaim{{Start: branchTime(1), End: branchTime(15), Issue: "336"}, {Start: branchTime(15), End: branchTime(31), Issue: "336", Terminal: true}, {Start: branchTime(2), End: branchTime(31), Issue: "335"}}
	history := []BranchInterval{{Start: branchTime(0), End: branchTime(20), Branch: "000336-a", Known: true}}
	got, err := SelectBranchActivity("335", windows, history, claims)
	if err != nil || len(got.Windows) != 1 || !got.Windows[0].Start.Equal(branchTime(20)) || got.Notes[len(got.Notes)-1].Kind != "ownership-only" {
		t.Fatalf("terminal attribution was either credited to B or retained as a lock: %+v %v", got, err)
	}
}

func TestSelectBranchActivityCountedDisjointPartition(t *testing.T) {
	previous := 0
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			start := branchTime(0)
			end := start.Add(time.Duration(n) * time.Minute)
			claims := []BranchClaim{{Start: start, End: end, Issue: "335"}, {Start: start, End: end, Issue: "336"}}
			var history []BranchInterval
			for i := range n {
				branch := "000335-b"
				if i%2 == 1 {
					branch = "000336-a"
				}
				history = append(history, BranchInterval{Start: start.Add(time.Duration(i) * time.Minute), End: start.Add(time.Duration(i+1) * time.Minute), Branch: branch, Known: true})
			}
			slices.Reverse(history) // input order cannot select another owner
			windows := []ActivityWindow{{Start: start, End: end, Began: start}}
			a, err := SelectBranchActivity("335", windows, history, claims)
			if err != nil {
				t.Fatal(err)
			}
			b, err := SelectBranchActivity("336", windows, history, claims)
			if err != nil {
				t.Fatal(err)
			}
			var total time.Duration
			for _, w := range a.Windows {
				total += w.End.Sub(w.Start)
				for _, other := range b.Windows {
					if w.Start.Before(other.End) && other.Start.Before(w.End) {
						t.Fatal("one slot interval was selected for both issues")
					}
				}
			}
			for _, w := range b.Windows {
				total += w.End.Sub(w.Start)
			}
			if total != end.Sub(start) || a.steps > 4*n+12 || previous > 0 && a.steps > 2*previous {
				t.Fatalf("n=%d selected=%s steps=%d prior=%d", n, total, a.steps, previous)
			}
			previous = a.steps
		})
	}
}

func TestSelectBranchActivityClaimAndDispatchBoundaries(t *testing.T) {
	start, end := branchTime(0), branchTime(31)
	claim := BranchClaim{Start: start, End: end, Issue: "335"}
	windows := []ActivityWindow{{Start: start, End: end, Began: start}}
	history := []BranchInterval{{Start: start, End: branchTime(10), Branch: "000335-b", Known: true}, {Start: branchTime(10), End: end}}
	got, err := SelectBranchActivity("335", windows, history, []BranchClaim{claim, {Start: branchTime(20), End: end, Issue: "336"}})
	if err != nil || len(got.Unknown) != 1 || !got.Unknown[0].Start.Equal(branchTime(20)) || !got.Unknown[0].End.Equal(end) || len(got.Windows) != 1 || !got.Windows[0].End.Equal(branchTime(20)) {
		t.Fatalf("ambiguity did not start at the second claim: %+v %v", got, err)
	}
	got, err = SelectBranchActivity("335", windows, history[:1], []BranchClaim{claim})
	if err != nil || len(got.Windows) != 1 || len(got.Notes) != 2 || got.Notes[1].Kind != "ownership-only" {
		t.Fatalf("known prefix plus sole-claim planning = %+v %v", got, err)
	}
	for _, invalid := range []string{"335-b", "000000-none", "000335-", "000335-b\n", "citation-000335-b"} {
		if BranchIssue(invalid) != "" {
			t.Fatalf("invalid branch %q named an issue", invalid)
		}
	}
}
