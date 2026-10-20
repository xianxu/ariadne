package affected

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// testGraph: leaf ← mid ← top, tonly depends on leaf only through its test, other is unrelated.
func testGraph() Graph {
	const m = "example.com/m"
	return Graph{
		Module: m,
		Dir: map[string]string{
			m + "/leaf":   "leaf",
			m + "/mid":    "mid",
			m + "/top":    "top",
			m + "/tonly":  "tonly",
			m + "/other":  "other",
			m + "/rt":     "rt",
			m + "/rt/sub": "rt/sub",
		},
		Deps: map[string]map[string]bool{
			m + "/leaf":   {},
			m + "/mid":    {m + "/leaf": true},
			m + "/top":    {m + "/mid": true, m + "/leaf": true},
			m + "/tonly":  {m + "/leaf": true},
			m + "/other":  {},
			m + "/rt":     {},
			m + "/rt/sub": {},
		},
	}
}

func testConfig() Config {
	return Config{
		Version: 1, Command: []string{"make", "test-affected"}, Weight: 1, TimeoutMinutes: 30,
		Full: []string{"Makefile*", "go.mod", ".sdlc/**"},
		Groups: []Group{
			{Name: "lua", Paths: []string{"**/*.lua", "tests/lua/**"}},
			{Name: "cli", Paths: []string{"tests/cli-*.sh"}, Go: []string{"./top"}},
		},
	}
}

func isDocTest(p string) bool { return strings.HasSuffix(p, ".md") }

func TestPlan(t *testing.T) {
	const m = "example.com/m/"
	cases := []struct {
		name  string
		paths []string
		want  Selection
	}{
		{"docs only is empty", []string{"README.md", "atlas/x.md"}, Selection{}},
		{"leaf selects leaf and dependents only (Done-when)", []string{"leaf/leaf.go"},
			Selection{GoPkgs: []string{m + "leaf", m + "mid", m + "tonly", m + "top"}, Groups: []string{"cli"}}},
		{"unrelated change selects itself", []string{"other/other.go"},
			Selection{GoPkgs: []string{m + "other"}}},
		{"testdata belongs to its package", []string{"other/testdata/golden.txt"},
			Selection{GoPkgs: []string{m + "other"}}},
		{"nested package is the deepest dir", []string{"rt/sub/x.go"},
			Selection{GoPkgs: []string{m + "rt/sub"}}},
		{"lua inside a package selects package and group (PQ-1)", []string{"rt/assets/init.lua"},
			Selection{GoPkgs: []string{m + "rt"}, Groups: []string{"lua"}}},
		{"group by path only", []string{"tests/lua/spec.lua"}, Selection{Groups: []string{"lua"}}},
		{"group via go root", []string{"mid/mid.go"},
			Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli"}}},
		{"group via its own script", []string{"tests/cli-run.sh"}, Selection{Groups: []string{"cli"}}},
		{"full beats everything", []string{"leaf/leaf.go", "tests/lua/a.lua", ".sdlc/affected.json", "README.md"},
			Selection{Full: true, Reason: ".sdlc/affected.json"}},
		{"unclassified falls back to full", []string{"leaf/leaf.go", "scripts/deploy.sh"},
			Selection{Full: true, Reason: "unclassified: scripts/deploy.sh"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Plan(c.paths, testGraph(), testConfig(), isDocTest)
			if c.want.Full {
				if !got.Full || !strings.Contains(got.Reason, c.want.Reason) {
					t.Fatalf("Plan = %+v; want full naming %q", got, c.want.Reason)
				}
				return
			}
			if !reflect.DeepEqual(norm(got), norm(c.want)) {
				t.Fatalf("Plan = %+v; want %+v", got, c.want)
			}
		})
	}
}

func norm(s Selection) Selection {
	if len(s.GoPkgs) == 0 {
		s.GoPkgs = nil
	}
	if len(s.Groups) == 0 {
		s.Groups = nil
	}
	return s
}

func TestPlanStaleGoRoot(t *testing.T) {
	cfg := testConfig()
	cfg.Groups[1].Go = []string{"./gone"}
	got := Plan([]string{"leaf/leaf.go"}, testGraph(), cfg, isDocTest)
	if !got.Full || got.Reason != "group cli: ./gone is not a package" {
		t.Fatalf("Plan = %+v", got)
	}
}

func TestSelectionEmpty(t *testing.T) {
	if !(Selection{}).Empty() {
		t.Fatal("zero Selection not empty")
	}
	for _, s := range []Selection{{Full: true}, {GoPkgs: []string{"x"}}, {Groups: []string{"g"}}} {
		if s.Empty() {
			t.Fatalf("%+v reported empty", s)
		}
	}
}

func TestPackageOfRoot(t *testing.T) {
	g := Graph{Module: "m", Dir: map[string]string{"m": ".", "m/a": "a"}, Deps: map[string]map[string]bool{}}
	for file, want := range map[string]string{"main.go": "m", "a/x.go": "m/a", "b/x.go": "m", "a/b/c.txt": "m/a"} {
		if got, ok := g.PackageOf(file); !ok || got != want {
			t.Errorf("PackageOf(%q) = %q, %v; want %q", file, got, ok, want)
		}
	}
	g2 := Graph{Module: "m", Dir: map[string]string{"m/a": "a"}}
	if _, ok := g2.PackageOf("b/x.go"); ok {
		t.Error("PackageOf outside any package reported ok")
	}
}

// TestPlanProperty: over random DAGs and change sets, GoPkgs is closed under dependents and contains only packages
// a changed package reaches.
func TestPlanProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(473))
	for iter := 0; iter < 500; iter++ {
		n := 1 + rng.Intn(12)
		g := Graph{Module: "m", Dir: map[string]string{}, Deps: map[string]map[string]bool{}}
		name := func(i int) string { return fmt.Sprintf("m/p%d", i) }
		direct := make([][]int, n)
		for i := 0; i < n; i++ {
			g.Dir[name(i)] = fmt.Sprintf("p%d", i)
			for j := 0; j < i; j++ {
				if rng.Intn(3) == 0 {
					direct[i] = append(direct[i], j)
				}
			}
		}
		for i := 0; i < n; i++ { // transitive closure; j < i so lower indices are complete
			deps := map[string]bool{}
			for _, j := range direct[i] {
				deps[name(j)] = true
				for d := range g.Deps[name(j)] {
					deps[d] = true
				}
			}
			g.Deps[name(i)] = deps
		}
		changed := map[string]bool{}
		var paths []string
		for i := 0; i < n; i++ {
			if rng.Intn(4) == 0 {
				changed[name(i)] = true
				paths = append(paths, fmt.Sprintf("p%d/f.go", i))
			}
		}
		sel := Plan(paths, g, Config{Version: 1, Command: []string{"x"}}, isDocTest)
		if sel.Full {
			t.Fatalf("iter %d: unexpected full: %s", iter, sel.Reason)
		}
		in := map[string]bool{}
		for _, p := range sel.GoPkgs {
			in[p] = true
		}
		for p, deps := range g.Deps {
			hit := changed[p]
			for d := range deps {
				hit = hit || changed[d]
			}
			if hit != in[p] {
				t.Fatalf("iter %d: %s selected=%v, reachable from change=%v", iter, p, in[p], hit)
			}
			if in[p] {
				continue
			}
			for d := range deps {
				if in[d] {
					t.Fatalf("iter %d: %s depends on selected %s but is not selected", iter, p, d)
				}
			}
		}
	}
}
