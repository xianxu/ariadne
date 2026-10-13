package project

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

func r(repo string, id int) IssueRef { return IssueRef{Repo: repo, ID: id} }

var unitCal = estimate.ForecastCalibration{N: 5, Point: 1, Low: 0.5, High: 2, K: 0.2, Prior: 3}

func eps(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestForecastTrainsParallelAndEdges(t *testing.T) {
	p := TrainPlan{Trains: []Train{
		{Name: "A", Cars: []IssueRef{r("x", 1), r("x", 2)}},
		{Name: "B", Cars: []IssueRef{r("x", 3)}},
	}}
	cars := map[IssueRef]CarEstimate{r("x", 1): {Estimate: 2}, r("x", 2): {Estimate: 1}, r("x", 3): {Estimate: 4}}
	f, err := ForecastTrains(p, cars, unitCal)
	if err != nil {
		t.Fatal(err)
	}
	// parallel: max(2+1, 4) = 4, ends in B
	if !eps(f.Path, 4) || f.LastTrain != "B" || !eps(f.Remaining, 7) || f.ActiveTrains != 2 {
		t.Fatalf("forecast = %+v", f)
	}
	if !eps(f.PathLow, 2) || !eps(f.PathHigh, 8) {
		t.Fatalf("range = %v–%v", f.PathLow, f.PathHigh)
	}
	// x#2 ⇒ x#3 serializes B behind A: 2+1+4 = 7
	p.Edges = []Edge{{r("x", 2), r("x", 3)}}
	f, _ = ForecastTrains(p, cars, unitCal)
	if !eps(f.Path, 7) || f.LastTrain != "B" {
		t.Fatalf("with edge = %+v", f)
	}
	// an edge from outside the plan is ignored
	p.Edges = []Edge{{r("other", 9), r("x", 1)}}
	if f, _ = ForecastTrains(p, cars, unitCal); !eps(f.Path, 4) {
		t.Fatalf("out-of-plan edge changed the path: %+v", f)
	}
}

func TestForecastTrainsLandedPriorAndMissing(t *testing.T) {
	p := TrainPlan{Trains: []Train{{Name: "A", Cars: []IssueRef{r("x", 1), r("x", 2), r("x", 3), r("x", 4)}}}}
	cars := map[IssueRef]CarEstimate{
		r("x", 1): {Estimate: 5, Terminal: true},
		r("x", 2): {Unestimated: true},
		r("x", 3): {LookupErr: "no card"},
		// x#4 absent from the map
	}
	f, err := ForecastTrains(p, cars, unitCal)
	if err != nil {
		t.Fatal(err)
	}
	if !eps(f.Remaining, 9) { // three cars at the 3h prior; the landed car counts 0
		t.Fatalf("remaining = %v", f.Remaining)
	}
	out := RenderTrainForecast(f)
	for _, want := range []string{"unestimated, at the 3.0h prior: x#2", "card not read", "x#3 (no card)", "x#4 (not resolved)"} {
		if !strings.Contains(out, want) {
			t.Errorf("statement missing %q:\n%s", want, out)
		}
	}
	all := map[IssueRef]CarEstimate{r("x", 1): {Terminal: true}, r("x", 2): {Terminal: true}, r("x", 3): {Terminal: true}, r("x", 4): {Terminal: true}}
	f, _ = ForecastTrains(p, all, unitCal)
	if f.Remaining != 0 || f.LastTrain != "" || !strings.Contains(RenderTrainForecast(f), "every car of 1 train(s) has landed") {
		t.Fatalf("all landed = %+v", f)
	}
}

func TestForecastTrainsCycleAndEmpty(t *testing.T) {
	p := TrainPlan{
		Trains: []Train{{Name: "A", Cars: []IssueRef{r("x", 1), r("x", 2)}}},
		Edges:  []Edge{{r("x", 2), r("x", 1)}},
	}
	if _, err := ForecastTrains(p, nil, unitCal); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle = %v", err)
	}
	if _, err := ForecastTrains(TrainPlan{}, nil, unitCal); err == nil {
		t.Fatal("empty plan forecast")
	}
}

// TestForecastTrainsInvariants: on random acyclic plans, the critical path is
// at least every train's own sum and at most the total, a ⇒ edge never
// shortens it, and low ≤ point ≤ high.
func TestForecastTrainsInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(329))
	for iter := 0; iter < 200; iter++ {
		var p TrainPlan
		cars := map[IssueRef]CarEstimate{}
		id := 0
		for tn := 0; tn < 1+rng.Intn(4); tn++ {
			tr := Train{Name: fmt.Sprintf("T%d", tn)}
			for c := 0; c < 1+rng.Intn(4); c++ {
				id++
				ref := r("x", id)
				tr.Cars = append(tr.Cars, ref)
				cars[ref] = CarEstimate{Estimate: float64(rng.Intn(5)), Terminal: rng.Intn(4) == 0}
			}
			p.Trains = append(p.Trains, tr)
		}
		base, err := ForecastTrains(p, cars, unitCal)
		if err != nil {
			t.Fatal(err)
		}
		if !(base.PathLow <= base.Path && base.Path <= base.PathHigh) {
			t.Fatalf("range order: %+v", base)
		}
		if base.Path > base.Remaining+1e-9 {
			t.Fatalf("path %v > total %v", base.Path, base.Remaining)
		}
		for _, tr := range p.Trains {
			sum := 0.0
			for _, c := range tr.Cars {
				ce := cars[c]
				if !ce.Terminal {
					est := ce.Estimate
					if est <= 0 {
						est = unitCal.Prior
					}
					sum += est
				}
			}
			if base.Path+1e-9 < sum {
				t.Fatalf("path %v < train %s sum %v", base.Path, tr.Name, sum)
			}
		}
		// a forward edge (lower id ⇒ higher id) keeps the graph acyclic
		if id >= 2 {
			a, b := 1+rng.Intn(id-1), 0
			b = a + 1 + rng.Intn(id-a)
			p.Edges = append(p.Edges, Edge{r("x", a), r("x", b)})
			withEdge, err := ForecastTrains(p, cars, unitCal)
			if err != nil {
				t.Fatal(err)
			}
			if withEdge.Path+1e-9 < base.Path {
				t.Fatalf("edge shortened the path: %v < %v", withEdge.Path, base.Path)
			}
		}
	}
}
