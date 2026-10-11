package activetime

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func at(m int) time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute) }

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// seg is one segment wholly allocated to #9.
func seg(from, to int) Segment {
	return Segment{Start: at(from), End: at(to), Active: float64(to - from), Alloc: map[string]float64{"9": float64(to - from)}}
}

// #234 D1: build runs to a boundary's first dispatch, converge from there to
// the last #N commit at or before its final dispatch, and the tail after that
// commit is uncounted at every boundary.
func TestBuildTimelineAndSplit(t *testing.T) {
	cases := []struct {
		name            string
		rounds          []BoundaryRound
		commits         []int
		current         string
		since, until    int
		build, converge float64
		tail            float64
		source          string
	}{
		{
			name:    "no rounds: all build",
			commits: []int{30}, since: 0, until: 60,
			build: 60, source: SplitNoReview,
		},
		{
			// The last commit is at 50; make test runs; dispatch at 55; finalize at 58.
			name:    "single-round pass: no converge, verification after the last commit is tail",
			rounds:  []BoundaryRound{{Boundary: "", Dispatched: at(55), Finalized: at(58), Finalizing: true}},
			commits: []int{10, 50}, since: 0, until: 58,
			build: 50, tail: 8, source: SplitDispatch,
		},
		{
			name: "REWORK with fix commits: converge from the first dispatch to the last fix commit",
			rounds: []BoundaryRound{
				{Boundary: "", Dispatched: at(30), Finalized: at(35)},
				{Boundary: "", Dispatched: at(55), Finalized: at(58), Finalizing: true},
			},
			commits: []int{25, 40, 50}, since: 0, until: 58,
			build: 30, converge: 20, tail: 8, source: SplitDispatch,
		},
		{
			name: "REWORK without fix commits: no converge",
			rounds: []BoundaryRound{
				{Boundary: "", Dispatched: at(30), Finalized: at(35)},
				{Boundary: "", Dispatched: at(40), Finalized: at(45), Finalizing: true},
			},
			commits: []int{25}, since: 0, until: 60,
			build: 25 + 15, tail: 20, source: SplitDispatch,
		},
		{
			// M1 passes in one round; its tail (20, 24] is uncounted and M2's
			// build starts at M1's finalize.
			name:    "single-round M1 then M2 build",
			rounds:  []BoundaryRound{{Boundary: "M1", Dispatched: at(22), Finalized: at(24), Finalizing: true}},
			commits: []int{20, 40}, current: "M2", since: 0, until: 60,
			build: 20 + 36, tail: 4, source: SplitDispatch,
		},
		{
			name:    "open REWORK at measurement: converge to the window end",
			rounds:  []BoundaryRound{{Boundary: "", Dispatched: at(30), Finalized: at(35)}},
			commits: []int{25, 50}, since: 0, until: 50,
			build: 30, converge: 20, source: SplitDispatch,
		},
		{
			name:    "no dispatch stamp: the finalize time is the cut",
			rounds:  []BoundaryRound{{Boundary: "", Finalized: at(35)}},
			commits: []int{25, 50}, since: 0, until: 50,
			build: 35, converge: 15, source: SplitFinalize,
		},
		{
			name:    "a dispatch after its finalize is not trusted",
			rounds:  []BoundaryRound{{Boundary: "", Dispatched: at(40), Finalized: at(35)}},
			commits: []int{25, 50}, since: 0, until: 50,
			build: 35, converge: 15, source: SplitFinalize,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var commits []time.Time
			for _, m := range c.commits {
				commits = append(commits, at(m))
			}
			tl := BuildTimeline(c.rounds, commits, c.current, at(c.since), at(c.until))
			if tl.Source != c.source {
				t.Errorf("source = %q, want %q", tl.Source, c.source)
			}
			got := Split([]Segment{seg(c.since, c.until)}, "9", tl)
			if !near(got.Build, c.build) || !near(got.Converge, c.converge) || !near(got.Tail, c.tail) {
				t.Errorf("split = %+v, want build %v converge %v tail %v\nphases %+v", got, c.build, c.converge, c.tail, tl.Phases)
			}
		})
	}
}

// E1: a milestone's increment is the phases of that boundary only.
func TestSplitForBoundary(t *testing.T) {
	rounds := []BoundaryRound{
		{Boundary: "M1", Dispatched: at(20), Finalized: at(25)},
		{Boundary: "M1", Dispatched: at(32), Finalized: at(34), Finalizing: true},
	}
	tl := BuildTimeline(rounds, []time.Time{at(10), at(30), at(50)}, "M2", at(0), at(60))
	m2 := Split([]Segment{seg(0, 60)}, "9", tl.For("M2"))
	if !near(m2.Build, 26) || m2.Converge != 0 {
		t.Errorf("M2 = %+v, want build 26 (34 → 60)", m2)
	}
	m1 := Split([]Segment{seg(0, 60)}, "9", tl.For("M1"))
	if !near(m1.Build, 20) || !near(m1.Converge, 10) {
		t.Errorf("M1 = %+v, want build 20, converge 10", m1)
	}
}

// Properties over random rounds and segments: the phases are ordered, never
// overlap and cover the window, and the split conserves every minute.
func TestSplitProperties(t *testing.T) {
	r := rand.New(rand.NewSource(234))
	for i := 0; i < 500; i++ {
		var rounds []BoundaryRound
		for _, b := range []string{"M1", "M2", ""}[:r.Intn(4)] {
			for k := r.Intn(3); k >= 0; k-- {
				d := r.Intn(120)
				rounds = append(rounds, BoundaryRound{Boundary: b, Dispatched: at(d), Finalized: at(d + r.Intn(10)), Finalizing: k == 0 && r.Intn(4) > 0})
			}
		}
		var commits []time.Time
		for k := r.Intn(8); k > 0; k-- {
			commits = append(commits, at(r.Intn(120)))
		}
		tl := BuildTimeline(rounds, commits, "", at(0), at(120))
		prev := at(0)
		for _, p := range tl.Phases {
			if p.Start.Before(prev) || !p.End.After(p.Start) {
				t.Fatalf("phases overlap or are empty: %+v", tl.Phases)
			}
			if p.Start.After(prev) {
				t.Fatalf("gap before %+v in %+v", p, tl.Phases)
			}
			prev = p.End
		}
		if !prev.Equal(at(120)) {
			t.Fatalf("phases end at %v, not the window end: %+v", prev, tl.Phases)
		}
		var segs []Segment
		total := 0.0
		for k := r.Intn(6); k >= 0; k-- {
			s := r.Intn(120)
			e := s + r.Intn(120-s+1)
			segs = append(segs, seg(s, e))
			total += float64(e - s)
		}
		got := Split(segs, "9", tl)
		if !near(got.Build+got.Converge+got.Tail, total) {
			t.Fatalf("split %+v does not conserve %v", got, total)
		}
	}
}
