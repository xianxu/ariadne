package affected

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// The declared-ownership contract (#598): affected.json's owners, full excepts and sidecar conventions, compiled into
// a Mapping that Plan consults in its claim step.

const declaredConfig = `{"version": 1, "command": ["make", "test-affected"],
 "full": ["go.mod", ".sdlc/**", "scripts/test-affected.py", {"glob": "cmd/couch/**", "except": ["cmd/couch/*_test.go"]}],
 "groups": [{"name": "lua", "paths": ["**/*.lua"]}, {"name": "cli", "paths": ["tests/cli-*.sh"], "go": ["./top"]}],
 "owners": [
   {"paths": ["docs/leaf/**"], "go": ["./leaf"]},
   {"paths": ["fixtures/shared.txt"], "as": ["mid/mid_test.go", "nvim/init.lua"]},
   {"paths": ["golden/**/*.out"], "strip": {"prefix": "golden/", "suffix": ".out"}}
 ],
 "sidecars": [{"prefix": "tests/readers/", "suffix": ".json"}]}`

func TestParseConfigDeclared(t *testing.T) {
	cfg, err := ParseConfig([]byte(declaredConfig))
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if want := (FullRule{Glob: "cmd/couch/**", Except: []string{"cmd/couch/*_test.go"}}); !reflect.DeepEqual(cfg.Full[3], want) || !reflect.DeepEqual(cfg.Full[0], FullRule{Glob: "go.mod"}) {
		t.Fatalf("full = %+v", cfg.Full)
	}
	if len(cfg.Owners) != 3 || cfg.Owners[2].Strip == nil || cfg.Sidecars[0].Prefix != "tests/readers/" {
		t.Fatalf("owners = %+v, sidecars = %+v", cfg.Owners, cfg.Sidecars)
	}
}

func TestParseConfigDeclaredRefuses(t *testing.T) {
	const head = `{"version":1,"command":["x"],`
	cases := []struct{ name, json, want string }{
		{"full object unknown key", `"full":[{"glob":"a","extra":1}]}`, "unknown field"},
		{"full object duplicate key", `"full":[{"glob":"a","glob":"b"}]}`, "duplicate object key"},
		{"full object no glob", `"full":[{"except":["a"]}]}`, "bad glob"},
		{"full wrong type", `"full":[3]}`, "glob or"},
		{"full bad except", `"full":[{"glob":"a","except":["b/["]}]}`, "except"},
		{"owner no paths", `"owners":[{"go":["./a"]}]}`, "paths is empty"},
		{"owner bad glob", `"owners":[{"paths":["a/["],"go":["./a"]}]}`, "bad glob"},
		{"owner none set", `"owners":[{"paths":["a"]}]}`, "exactly one"},
		{"owner two set", `"owners":[{"paths":["a"],"go":["./a"],"as":["b"]}]}`, "exactly one"},
		{"owner bad root", `"owners":[{"paths":["a"],"go":["../a"]}]}`, "go root"},
		{"owner traversal target", `"owners":[{"paths":["a"],"as":["../etc/x"]}]}`, "clean"},
		{"owner absolute target", `"owners":[{"paths":["a"],"as":["/etc/x"]}]}`, "clean"},
		{"owner empty strip", `"owners":[{"paths":["a"],"strip":{"prefix":"","suffix":""}}]}`, "strip"},
		{"owner unknown key", `"owners":[{"paths":["a"],"go":["./a"],"run":"x"}]}`, "unknown field"},
		{"owner duplicate key", `"owners":[{"paths":["a"],"paths":["b"],"go":["./a"]}]}`, "duplicate object key"},
		{"sidecar no slash", `"sidecars":[{"prefix":"tests/readers","suffix":".json"}]}`, "prefix"},
		{"sidecar traversal", `"sidecars":[{"prefix":"../readers/","suffix":".json"}]}`, "prefix"},
		{"sidecar empty suffix", `"sidecars":[{"prefix":"r/","suffix":""}]}`, "suffix"},
		{"sidecar slash suffix", `"sidecars":[{"prefix":"r/","suffix":"/x.json"}]}`, "suffix"},
		{"sidecar unknown key", `"sidecars":[{"prefix":"r/","suffix":".json","reads":[]}]}`, "unknown field"},
		{"traceability not an array", `"traceability":"atlas/t.yaml"}`, "cannot unmarshal"},
		{"traceability glob", `"traceability":["atlas/*.yaml"]}`, "exact"},
		{"traceability traversal", `"traceability":["../t.yaml"]}`, "exact"},
		{"traceability duplicate", `"traceability":["t.yaml","t.yaml"]}`, "duplicate"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseConfig([]byte(head + c.json))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// FuzzParseConfig: a config either refuses or reparses from its own fields to the same rules; never a panic.
func FuzzParseConfig(f *testing.F) {
	for _, s := range []string{exampleConfig, declaredConfig, `{"version":1,"command":["x"],"full":[{"glob":"a/**","except":["a/b"]}]}`,
		`{"version":1,"command":["x"],"owners":[{"paths":["a"],"strip":{"prefix":"../","suffix":""}}]}`,
		`{"version":1,"command":["x"],"traceability":["atlas/traceability.yaml"]}`,
		`{"version":1,"command":["x"],"traceability":["atlas/*.yaml","../t.yaml","a","a"]}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		cfg, err := ParseConfig(data)
		if err != nil {
			return
		}
		for _, r := range cfg.Full {
			if !validGlob(r.Glob) {
				t.Fatalf("accepted bad full glob %q", r.Glob)
			}
		}
		for _, o := range cfg.Owners {
			if err := validOwner(o); err != nil {
				t.Fatalf("accepted owner %+v: %v", o, err)
			}
		}
		for _, sc := range cfg.Sidecars {
			if err := validSidecar(sc.Prefix, sc.Suffix); err != nil {
				t.Fatalf("accepted sidecar %+v: %v", sc, err)
			}
		}
		seen := map[string]bool{}
		for _, p := range cfg.Traceability {
			if !cleanRelPath(p) || hasGlobMeta(p) || seen[p] {
				t.Fatalf("accepted traceability path %q", p)
			}
			seen[p] = true
		}
	})
}

func TestParseSidecar(t *testing.T) {
	cases := []struct {
		name, json string
		want       []string
		err        string
	}{
		{"pair today: no reads declares nothing", `{"version": 1, "disposition": "local_or_literal_fixture", "sha256": "92f7"}`, nil, ""},
		{"reads, opaque members ignored", `{"version": 1, "reads": ["tests/fixtures/**", "nvim/init.lua"], "future": {"x": [1]}}`, []string{"tests/fixtures/**", "nvim/init.lua"}, ""},
		{"empty reads", `{"reads": []}`, []string{}, ""},
		{"null reads", `{"reads": null}`, nil, "array"},
		{"reads not array", `{"reads": "a"}`, nil, "array"},
		{"reads not strings", `{"reads": [1]}`, nil, "array"},
		{"traversal read", `{"reads": ["../x"]}`, nil, "clean"},
		{"absolute read", `{"reads": ["/etc/passwd"]}`, nil, "clean"},
		{"bad glob read", `{"reads": ["a/["]}`, nil, "clean"},
		{"duplicate key", `{"reads": ["a"], "reads": ["b"]}`, nil, "duplicate"},
		{"array", `[]`, nil, "object"},
		{"null", `null`, nil, "object"},
		{"trailing", `{} {}`, nil, "trailing"},
		{"truncated", `{"reads": ["a"`, nil, "unexpected end"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseSidecar([]byte(c.json))
			if c.err != "" {
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("err = %v, want %q", err, c.err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %q, %v; want %q", got, err, c.want)
			}
		})
	}
}

// FuzzSidecarDerivation: whatever a sidecar's bytes or a strip's input, every read and every derived target is a
// clean relative path, or the input refuses.
func FuzzSidecarDerivation(f *testing.F) {
	for _, s := range []struct{ body, path string }{
		{`{"reads":["a/**"]}`, "tests/readers/cmd/x_test.go.json"},
		{`{"reads":["../a"]}`, "tests/readers/../x_test.go.json"},
		{`{"reads":["a//b"]}`, "tests/readers//x.json"},
		{`{"reads":["a\\b"]}`, "tests/readers/.json"},
		{"{\"reads\":[\"a\x00\"]}", "tests/readers/\x00.json"},
		{`{"reads":["/abs"]}`, "tests/readers/./x.json"},
	} {
		f.Add([]byte(s.body), s.path)
	}
	f.Fuzz(func(t *testing.T, body []byte, p string) {
		if reads, err := ParseSidecar(body); err == nil {
			for _, r := range reads {
				if !cleanRelPath(r) || !validGlob(r) {
					t.Fatalf("accepted read %q", r)
				}
			}
		}
		if target, ok := stripTarget(Strip{Prefix: "tests/readers/", Suffix: ".json"}, p); ok && !cleanRelPath(target) {
			t.Fatalf("derived %q from %q", target, p)
		}
	})
}

// declaredGraph extends testGraph with the packages declared ownership names.
func declaredGraph() Graph {
	g := testGraph()
	const m = "example.com/m"
	g.Dir[m+"/cmd/couch"] = "cmd/couch"
	g.Deps[m+"/cmd/couch"] = map[string]bool{m + "/leaf": true}
	return g
}

func declaredMapping(t *testing.T, cfg Config, sidecars ...SidecarFile) Mapping {
	t.Helper()
	m, err := Compile(cfg, Sources{Sidecars: sidecars})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return m
}

func TestPlanDeclared(t *testing.T) {
	const m = "example.com/m/"
	cfg, err := ParseConfig([]byte(declaredConfig))
	if err != nil {
		t.Fatal(err)
	}
	mp := declaredMapping(t, cfg,
		SidecarFile{Path: "tests/readers/mid/mid_test.go.json", Content: []byte(`{"version":1,"reads":["fixtures/mid/**"]}`)},
		SidecarFile{Path: "tests/readers/leaf/leaf_test.go.json", Content: []byte(`{"version":1,"disposition":"non_go_full_fallback"}`)})
	leafUp := []string{m + "cmd/couch", m + "leaf", m + "mid", m + "tonly", m + "top"}
	cases := []struct {
		name  string
		paths []string
		want  Selection
	}{
		{"declared go root selects the package and its dependents, and the group on it", []string{"docs/leaf/notes.txt"}, Selection{GoPkgs: leafUp, Groups: []string{"cli"}}},
		{"as targets union a package and a group", []string{"fixtures/shared.txt"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli", "lua"}}},
		{"strip derives the target's owner", []string{"golden/leaf/x.go.out"}, Selection{GoPkgs: leafUp, Groups: []string{"cli"}}},
		{"an edited sidecar selects its test's package", []string{"tests/readers/mid/mid_test.go.json"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli"}}},
		{"an added sidecar main has never seen is path-only", []string{"tests/readers/top/new_test.go.json"}, Selection{GoPkgs: []string{m + "top"}, Groups: []string{"cli"}}},
		{"a sidecar for a deleted test in a living package", []string{"tests/readers/leaf/gone_test.go.json"}, Selection{GoPkgs: leafUp, Groups: []string{"cli"}}},
		{"a renamed sidecar: both paths", []string{"tests/readers/mid/a_test.go.json", "tests/readers/top/a_test.go.json"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli"}}},
		{"a declared read selects the reader's package", []string{"fixtures/mid/case1.txt"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli"}}},
		{"full wins over a declaration", []string{"docs/leaf/notes.txt", "go.mod"}, full("go.mod")},
		{"except falls through to the test's package", []string{"cmd/couch/main_test.go"}, Selection{GoPkgs: []string{m + "cmd/couch"}, Groups: []string{}}},
		{"production code under the except's glob stays full", []string{"cmd/couch/main.go"}, full("cmd/couch/main.go")},
		{"the selector stays full", []string{"scripts/test-affected.py"}, full("scripts/test-affected.py")},
		{"a target matching full is full", []string{"golden/cmd/couch/main.go.out"}, full("cmd/couch/main.go (declared by affected.json owners[2] for golden/cmd/couch/main.go.out)")},
		{"a sidecar for a deleted package claims nothing", []string{"tests/readers/gone/x_test.go.json"}, full("affected.json sidecars[0]: target gone/x_test.go of tests/readers/gone/x_test.go.json claims nothing")},
		{"a sidecar path no target derives from", []string{"tests/readers/.json"}, full("affected.json sidecars[0]: no target derives from tests/readers/.json")},
		{"an unmapped path is unchanged", []string{"tools/x.py"}, full("unclassified: tools/x.py")},
		{"an unmapped doc is unchanged", []string{"atlas/x.md"}, Selection{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Plan(c.paths, declaredGraph(), cfg, mp, isDocTest); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Plan(%q)\n got  %+v\n want %+v", c.paths, got, c.want)
			}
		})
	}
}

func TestPlanDeclaredStaleRoot(t *testing.T) {
	cfg := testConfig()
	cfg.Owners = []OwnerDecl{{Paths: []string{"docs/**"}, Go: []string{"./gone"}}}
	got := Plan([]string{"docs/a.txt"}, testGraph(), cfg, declaredMapping(t, cfg), isDocTest)
	if want := full("affected.json owners[0]: ./gone is not a package"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestCompileRefuses(t *testing.T) {
	cfg, err := ParseConfig([]byte(declaredConfig))
	if err != nil {
		t.Fatal(err)
	}
	for name, f := range map[string]SidecarFile{
		"malformed sidecar":   {Path: "tests/readers/mid/mid_test.go.json", Content: []byte(`{"reads": "x"}`)},
		"outside conventions": {Path: "elsewhere/x.json", Content: []byte(`{}`)},
		"no derivable target": {Path: "tests/readers/.json", Content: []byte(`{}`)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Compile(cfg, Sources{Sidecars: []SidecarFile{f}}); err == nil {
				t.Fatal("Compile accepted it")
			}
		})
	}
}

// TestOwnedProperty: declarations only add — Owned holds every graph owner of every path, declared or not, with every
// source compiled in (owners, sidecars, sidecar reads and, #616, traceability entries).
func TestOwnedProperty(t *testing.T) {
	cfg, err := ParseConfig([]byte(declaredConfig))
	if err != nil {
		t.Fatal(err)
	}
	g := declaredGraph()
	cfg.Traceability = []string{"atlas/traceability.yaml"}
	mp, err := Compile(cfg, Sources{
		Sidecars:     []SidecarFile{{Path: "tests/readers/mid/mid_test.go.json", Content: []byte(`{"reads":["fixtures/**"]}`)}},
		Traceability: []TraceabilityFile{{Path: "atlas/traceability.yaml", Content: []byte(traceDoc)}}})
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"leaf/leaf.go", "mid/testdata/x", "rt/assets/init.lua", "docs/leaf/a", "fixtures/shared.txt", "fixtures/mid/a",
		"golden/top/t.out", "tests/readers/mid/mid_test.go.json", "tests/readers/gone/x.json", "cmd/couch/x_test.go", "other/blob.bin",
		"ui/toolbar.txt", "ui/icons/a.png", "ui/gone.txt", "ui/couch.txt", "ui/notes.txt"}
	for _, p := range paths {
		owned := map[string]bool{}
		for _, pkg := range Owned(p, g, cfg, mp) {
			owned[pkg] = true
		}
		for _, pkg := range g.Owners(p) {
			if !owned[pkg] {
				t.Errorf("Owned(%s) lost graph owner %s", p, pkg)
			}
		}
	}
	if got := Owned("fixtures/mid/a", g, cfg, mp); !reflect.DeepEqual(got, []string{"example.com/m/mid"}) {
		t.Fatalf("declared read: Owned = %q", got)
	}
}

// BenchmarkPlanDeclared is the operating envelope: 2000 changed paths against 600 sidecar reads rules (exact),
// 600 globbed owners rules and a 65-entry traceability file.
func BenchmarkPlanDeclared(b *testing.B) {
	g := testGraph()
	const m = "example.com/m"
	for i := range 600 {
		dir := fmt.Sprintf("p%d", i)
		g.Dir[m+"/"+dir] = dir
		g.Deps[m+"/"+dir] = map[string]bool{}
	}
	cfg := testConfig()
	cfg.Sidecars = []SidecarDecl{{Prefix: "tests/readers/", Suffix: ".json"}}
	var sidecars []SidecarFile
	for i := range 600 {
		cfg.Owners = append(cfg.Owners, OwnerDecl{Paths: []string{fmt.Sprintf("assets/g%d/**/*.dat", i)}, Go: []string{fmt.Sprintf("./p%d", i)}})
		sidecars = append(sidecars, SidecarFile{Path: fmt.Sprintf("tests/readers/p%d/x_test.go.json", i), Content: []byte(fmt.Sprintf(`{"reads":["inputs/p%d.txt"]}`, i))})
	}
	// #616: a parley-sized traceability file, 65 entries of 20 exact code paths each.
	var trace strings.Builder
	trace.WriteString("version: 1\natlas:\n")
	for e := range 65 {
		fmt.Fprintf(&trace, "  f%d:\n    code:\n", e)
		for c := range 20 {
			fmt.Fprintf(&trace, "      - src/f%d/c%d.lua\n", e, c)
		}
		fmt.Fprintf(&trace, "    tests: [p%d/x_test.go]\n", e)
	}
	cfg.Traceability = []string{"atlas/traceability.yaml"}
	mp, err := Compile(cfg, Sources{Sidecars: sidecars, Traceability: []TraceabilityFile{{Path: "atlas/traceability.yaml", Content: []byte(trace.String())}}})
	if err != nil {
		b.Fatal(err)
	}
	var paths []string
	for i := range 2000 {
		switch i % 4 {
		case 0:
			paths = append(paths, fmt.Sprintf("inputs/p%d.txt", i%600))
		case 1:
			paths = append(paths, fmt.Sprintf("tests/readers/p%d/x_test.go.json", i%600))
		case 2:
			paths = append(paths, fmt.Sprintf("src/f%d/c%d.lua", i%65, i%20))
		default:
			paths = append(paths, fmt.Sprintf("p%d/f%d.go", i%600, i))
		}
	}
	b.ResetTimer()
	for b.Loop() {
		if sel := Plan(paths, g, cfg, mp, isDocTest); sel.Full {
			b.Fatalf("full: %s", sel.Reason)
		}
	}
}

// TestLiteralPrefix: the prefilter never rejects a path the glob matches.
func TestLiteralPrefix(t *testing.T) {
	for glob, want := range map[string]string{"a/b/*.go": "a/b/", "**/x": "", "a/**/b": "a/", "a/b": "a/", "x*": "", "a/[b]/c": "a/"} {
		if got := literalPrefix(glob); got != want {
			t.Errorf("literalPrefix(%q) = %q, want %q", glob, got, want)
		}
	}
	for _, c := range [][2]string{{"a/**/b", "a/b"}, {"a/**/b", "a/x/y/b"}, {"**/x", "x"}, {"a/b/*.go", "a/b/c.go"}} {
		if ok, _ := Match(c[0], c[1]); !ok || !strings.HasPrefix(c[1], literalPrefix(c[0])) {
			t.Errorf("%q vs %q: match %v, prefix %q", c[0], c[1], ok, literalPrefix(c[0]))
		}
	}
}

// TestOwnedOrderIndependent: a declaration that forces full does not hide a later declaration's packages from Owned.
func TestOwnedOrderIndependent(t *testing.T) {
	cfg := testConfig()
	cfg.Owners = []OwnerDecl{{Paths: []string{"docs/**"}, Go: []string{"./gone"}}, {Paths: []string{"docs/**"}, Go: []string{"./other"}}}
	mp := declaredMapping(t, cfg)
	if got := Owned("docs/a.txt", testGraph(), cfg, mp); !reflect.DeepEqual(got, []string{"example.com/m/other"}) {
		t.Fatalf("Owned = %q", got)
	}
	if got := Plan([]string{"docs/a.txt"}, testGraph(), cfg, mp, isDocTest); !reflect.DeepEqual(got, full("affected.json owners[0]: ./gone is not a package")) {
		t.Fatalf("Plan = %+v", got)
	}
}
