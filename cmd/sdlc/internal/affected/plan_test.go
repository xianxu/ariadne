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
		// rt embeds its Lua runtime, as pair's runtimebundle does.
		Embeds: map[string][]string{"rt/assets/init.lua": {m + "/rt"}, "other/blob.bin": {m + "/leaf"}, "help/usage.md": {m + "/other"}},
	}
}

func testConfig() Config {
	return Config{
		Version: 1, Command: []string{"make", "test-affected"}, Weight: 1, TimeoutMinutes: 30,
		Full: []FullRule{{Glob: "Makefile*"}, {Glob: "go.mod"}, {Glob: ".sdlc/**"}},
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
		// Close review: a doc is ignored only when nothing claims it.
		{"an embedded doc selects its embedder", []string{"help/usage.md"}, Selection{GoPkgs: []string{m + "other"}}},
		{"a testdata doc selects its package", []string{"other/testdata/golden.md"}, Selection{GoPkgs: []string{m + "other"}}},
		{"a doc in a package dir selects the package", []string{"other/NOTES.md"}, Selection{GoPkgs: []string{m + "other"}}},
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
		{"a plain subdirectory of a package is unclassified (BR-1)", []string{"leaf/notes/n.txt"},
			Selection{Full: true, Reason: "unclassified: leaf/notes/n.txt"}},
		{"a file embedded across package dirs selects both owners (BR-7)", []string{"other/blob.bin"},
			Selection{GoPkgs: []string{m + "leaf", m + "mid", m + "other", m + "tonly", m + "top"}, Groups: []string{"cli"}}},
		{"unembedded lua in a package dir's subtree is the group only", []string{"rt/assets/extra.lua"},
			Selection{Groups: []string{"lua"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Plan(c.paths, testGraph(), testConfig(), Mapping{}, isDocTest)
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

// #473 M1 BR-1: a module whose root is a package must not swallow every path: a script outside any package
// still forces the full suite.
func TestPlanRootPackageDoesNotClaimSubdirectories(t *testing.T) {
	g := testGraph()
	g.Dir["example.com/m"], g.Deps["example.com/m"] = ".", map[string]bool{}
	if got := Plan([]string{"scripts/deploy.sh"}, g, testConfig(), Mapping{}, isDocTest); !got.Full {
		t.Fatalf("Plan = %+v; a root package claimed scripts/", got)
	}
	if got := Plan([]string{"main.go"}, g, testConfig(), Mapping{}, isDocTest); got.Full || !reflect.DeepEqual(got.GoPkgs, []string{"example.com/m"}) {
		t.Fatalf("Plan = %+v; the root's own file", got)
	}
}

func TestPlanStaleGoRoot(t *testing.T) {
	cfg := testConfig()
	cfg.Groups[1].Go = []string{"./gone"}
	got := Plan([]string{"leaf/leaf.go"}, testGraph(), cfg, Mapping{}, isDocTest)
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

// #473 M1 BR-1/BR-7: ownership is exact and a union. A file is owned by the package whose dir holds it, every
// package whose testdata/ holds it, and every package that embeds it; a plain subdirectory file, under the root
// package above all, is owned by none.
func TestOwnersAreExactAndAUnion(t *testing.T) {
	g := Graph{Module: "m", Dir: map[string]string{"m": ".", "m/a": "a", "m/a/b": "a/b"}, Deps: map[string]map[string]bool{},
		Embeds: map[string][]string{"a/assets/x.lua": {"m/a"}, "a/b/data.txt": {"m"}, "a/testdata/t.txt": {"m"}}}
	for file, want := range map[string]string{
		"main.go": "m", "a/x.go": "m/a", "a/README.txt": "m/a",
		"testdata/in.txt": "m", "a/testdata/deep/in.txt": "m/a", "a/assets/x.lua": "m/a",
		"a/b/data.txt":     "m m/a/b", // its own dir's package and the root, which embeds it
		"a/testdata/t.txt": "m m/a",   // its testdata owner and an embedder
		"a/b/testdata/x":   "m/a/b",
		"b/x.go":           "", "scripts/run.sh": "", "a/assets/other.lua": "", "a/notes/n.txt": "",
	} {
		if got := strings.Join(g.Owners(file), " "); got != want {
			t.Errorf("Owners(%q) = %q; want %q", file, got, want)
		}
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
		sel := Plan(paths, g, Config{Version: 1, Command: []string{"x"}}, Mapping{}, isDocTest)
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
