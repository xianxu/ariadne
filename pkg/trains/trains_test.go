package trains

import (
	"reflect"
	"strings"
	"testing"
)

func refs(t *testing.T, ss ...string) []IssueRef {
	t.Helper()
	var out []IssueRef
	for _, s := range ss {
		var last string
		got := parseRefs(s, &last)
		if len(got) != 1 {
			t.Fatalf("bad test ref %q", s)
		}
		out = append(out, got[0])
	}
	return out
}

// The ops ariadne-robustness-1 table and edges as authored on 10-12: the
// parser's real input, not an idealized one.
const opsTrains = `Some lede prose about trains.

| Train | Cars, in order | Slot |
|---|---|---|
| Main integration | ariadne#183 → #269 → #303 | ariadne:4 |
| Ownership and tracker | ariadne#320 → #324 (borrowed from test infra, for load) → #326 → #307 | ariadne:4 (from #274, 10-11) |
| Project datatype | ariadne#329 | ariadne:3, after #330 |
| Merge and CI | pair#324 ⇒ ariadne#196 → #309 → #228 | ariadne:2 (from #196, 10-12) |
| Tail | ariadne#240 / #235 / #311 | done |

Cross-train edges (⇒ hard, must land first; ~ overlap, coordinate through the TL):
- ariadne#183 and ariadne#320 ⇒ ariadne#269/#303 (requirement 6 applies only after requirement 5 is complete).
- ariadne#183 ~ ariadne#234: #183's narrow re-close is a new converge segment.
- pair#324 ⇒ ariadne#309: merge refusing red CI would block every pair#99 landing.
`

func TestParseTrainsOpsShape(t *testing.T) {
	p := ParseTrains(opsTrains, "ariadne")
	if len(p.Warnings) != 0 {
		t.Fatalf("warnings: %v", p.Warnings)
	}
	want := []Train{
		{Name: "Main integration", Slot: "ariadne:4", Cars: refs(t, "ariadne#183", "ariadne#269", "ariadne#303")},
		{Name: "Ownership and tracker", Slot: "ariadne:4 (from #274, 10-11)", Cars: refs(t, "ariadne#320", "ariadne#324", "ariadne#326", "ariadne#307")},
		{Name: "Project datatype", Slot: "ariadne:3, after #330", Cars: refs(t, "ariadne#329")},
		{Name: "Merge and CI", Slot: "ariadne:2 (from #196, 10-12)", Cars: refs(t, "pair#324", "ariadne#196", "ariadne#309", "ariadne#228")},
		{Name: "Tail", Slot: "done", Cars: refs(t, "ariadne#240", "ariadne#235", "ariadne#311")},
	}
	if !reflect.DeepEqual(p.Trains, want) {
		t.Fatalf("trains:\n got %+v\nwant %+v", p.Trains, want)
	}
	r := func(s string) IssueRef { return refs(t, s)[0] }
	wantEdges := []Edge{
		{r("ariadne#183"), r("ariadne#269")}, {r("ariadne#183"), r("ariadne#303")},
		{r("ariadne#320"), r("ariadne#269")}, {r("ariadne#320"), r("ariadne#303")},
		{r("pair#324"), r("ariadne#309")}, // the reason's pair#99 is not an edge
	}
	if !reflect.DeepEqual(p.Edges, wantEdges) {
		t.Fatalf("edges:\n got %v\nwant %v", p.Edges, wantEdges)
	}
	tr, car, ok := p.Locate(r("ariadne#326"))
	if !ok || tr.Name != "Ownership and tracker" || car != 3 {
		t.Fatalf("Locate(#326) = %q car %d ok %v", tr.Name, car, ok)
	}
	if _, _, ok := p.Locate(r("ariadne#999")); ok {
		t.Fatal("Locate found an issue in no train")
	}
}

func TestParseTrainsGrammar(t *testing.T) {
	cases := []struct {
		name, body, repo string
		cars             []string // first train's cars, as repo#N
		warnings         int
	}{
		{"bare first ref takes the project repo", "| Train | Cars | Slot |\n|-|-|-|\n| A | #7 → #0008 | x |", "metis", []string{"metis#7", "metis#8"}, 0},
		{"repo inherits across a qualified ref", "| Train | Cars |\n|-|-|\n| A | #1, pair#2, #3 |", "ariadne", []string{"ariadne#1", "pair#2", "pair#3"}, 0},
		{"dotted repo name", "| Train | Cars |\n|-|-|\n| A | parley.nvim#314 → #316 |", "ariadne", []string{"parley.nvim#314", "parley.nvim#316"}, 0},
		{"nested parens stripped", "| Train | Cars |\n|-|-|\n| A | ariadne#1 (see (#99)) → #2 |", "", []string{"ariadne#1", "ariadne#2"}, 0},
		{"row with no cars warns", "| Train | Cars |\n|-|-|\n| A | tbd |\n| B | ariadne#5 |", "", []string{"ariadne#5"}, 1},
		{"short row warns", "| Train | Slot | Cars |\n|-|-|-|\n| A |\n| B | s | ariadne#5 |", "", []string{"ariadne#5"}, 1},
		{"table without a Cars header is not trains", "| Name | Refs |\n|-|-|\n| A | ariadne#1 |", "", nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := ParseTrains(c.body, c.repo)
			if len(p.Warnings) != c.warnings {
				t.Fatalf("warnings = %v, want %d", p.Warnings, c.warnings)
			}
			var got []string
			if len(p.Trains) > 0 {
				for _, r := range p.Trains[0].Cars {
					got = append(got, r.String())
				}
			}
			if strings.Join(got, ",") != strings.Join(c.cars, ",") {
				t.Fatalf("cars = %v, want %v", got, c.cars)
			}
		})
	}
}

func TestParseTrainsEdgeChainAndMalformed(t *testing.T) {
	p := ParseTrains("- a#1 ⇒ a#2 ⇒ #3\n- ⇒ a#9\n- a#4 ~ a#5\n", "")
	want := []Edge{{IssueRef{"a", 1}, IssueRef{"a", 2}}, {IssueRef{"a", 2}, IssueRef{"a", 3}}}
	if !reflect.DeepEqual(p.Edges, want) {
		t.Fatalf("edges = %v, want %v", p.Edges, want)
	}
	if len(p.Warnings) != 1 {
		t.Fatalf("warnings = %v, want the one-sided ⇒", p.Warnings)
	}
}

// FuzzParseTrains: hand-authored markdown never panics the parser, and every
// parsed train carries a name and at least one car.
func FuzzParseTrains(f *testing.F) {
	for _, seed := range []string{opsTrains, "| Train | Cars |\n|", "- ⇒ ⇒", "| Train | Cars |\n|-|-|\n| | #|", "((((#1"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		p := ParseTrains(body, "r")
		for _, tr := range p.Trains {
			if tr.Name == "" || len(tr.Cars) == 0 {
				t.Fatalf("train without name or cars: %+v", tr)
			}
		}
	})
}
