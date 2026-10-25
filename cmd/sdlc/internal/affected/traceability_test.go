package affected

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// The golden is parley.nvim's atlas/traceability.yaml at ee938b5a (#616): the real file must parse unchanged.
func TestParseTraceabilityParleyGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/parley-traceability.yaml")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ParseTraceability(data)
	if err != nil {
		t.Fatalf("parley's file refused: %v", err)
	}
	// 65 entries have code; one has `tests: []` and so declares nothing.
	if len(entries) != 64 {
		t.Fatalf("%d entries, want 64", len(entries))
	}
	got := entries[0]
	want := TraceEntry{Key: "chat/spell_typeahead",
		Code: []string{"lua/parley/spell.lua", "lua/parley/spell_state.lua", "lua/parley/spell_blink.lua", "lua/parley/spell_source.lua",
			"scripts/check-spell-compatibility.sh", "tests/packaging/spell_compatibility.lua", "tests/helpers/fake_blink.lua"},
		Tests: []string{"tests/unit/spell_spec.lua", "tests/unit/spell_state_spec.lua", "tests/integration/spell_chat_spec.lua",
			"tests/integration/spell_source_spec.lua", "tests/integration/spell_blink_spec.lua", "tests/integration/neighborhood_completion_spec.lua"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("first entry = %+v", got)
	}
}

// Every hand-edit form the walk owns refuses, at each level it can appear.
func TestParseTraceabilityRefuses(t *testing.T) {
	const ok = "version: 1\natlas:\n  a:\n    code: [x.go]\n    tests: [x_test.go]\n"
	if _, err := ParseTraceability([]byte(ok)); err != nil {
		t.Fatalf("baseline refused: %v", err)
	}
	cases := []struct{ name, doc, want string }{
		{"empty", "", "empty"},
		{"two documents", ok + "---\nversion: 1\n", "more than one"},
		{"version 2", strings.Replace(ok, "version: 1", "version: 2", 1), "version"},
		{"version string", strings.Replace(ok, "version: 1", `version: "1"`, 1), "version"},
		{"no version", strings.Replace(ok, "version: 1\n", "", 1), "version"},
		{"unknown root key", ok + "extra: 1\n", "unknown key"},
		{"duplicate root key", ok + "version: 1\n", "duplicate key"},
		{"duplicate entry", ok + "  a:\n    code: [y.go]\n    tests: [y_test.go]\n", "duplicate key"},
		{"duplicate field", ok + "    code: [y.go]\n", "duplicate key"},
		{"unknown field", ok + "    owner: me\n", "unknown key"},
		{"int path", strings.Replace(ok, "[x_test.go]", "[1]", 1), "strings"},
		{"null path", strings.Replace(ok, "[x_test.go]", "[~]", 1), "strings"},
		{"empty path", strings.Replace(ok, "[x_test.go]", `[""]`, 1), "clean"},
		{"traversal", strings.Replace(ok, "[x_test.go]", "[../x_test.go]", 1), "clean"},
		{"glob test", strings.Replace(ok, "[x_test.go]", "[\"*_test.go\"]", 1), "clean"},
		{"scalar list", strings.Replace(ok, "[x_test.go]", "x_test.go", 1), "list"},
		{"alias", "version: 1\natlas:\n  a: &e\n    code: [x.go]\n    tests: [x_test.go]\n  b: *e\n", "aliases"},
		{"merge key", "version: 1\natlas:\n  a:\n    <<: {code: [x.go]}\n    tests: [x_test.go]\n", "merge"},
		{"atlas a list", "version: 1\natlas: [a]\n", "mapping"},
		{"int key", "version: 1\natlas:\n  1:\n    code: [x.go]\n    tests: [x_test.go]\n", "keys must be"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseTraceability([]byte(c.doc))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// FuzzParseTraceability: hand-edited YAML either refuses or yields only clean paths, never a panic.
func FuzzParseTraceability(f *testing.F) {
	golden, _ := os.ReadFile("testdata/parley-traceability.yaml")
	f.Add(golden)
	for _, s := range []string{
		"version: 1\natlas:\n  a:\n    code: [x.go]\n    tests: [x_test.go]\n",
		"version: 1\nversion: 1\natlas: {}\n",
		"version: 1\natlas:\n  a: {code: [x], tests: [y]}\n  a: {code: [x], tests: [y]}\n",
		"version: 1\natlas:\n  a: {code: [x], code: [z], tests: [y]}\n",
		"version: 1\natlas:\n  a: &e {code: [x], tests: [y]}\n  b: *e\n",
		"version: 1\natlas:\n  a: {<<: {code: [x]}, tests: [y]}\n",
		"version: 1\natlas:\n  a: {code: [1, ~, ../x, /abs, a//b], tests: [y]}\n",
		"version: 1\n---\nversion: 1\n",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		entries, err := ParseTraceability(data)
		if err != nil {
			return
		}
		keys := map[string]bool{}
		for _, e := range entries {
			if keys[e.Key] || len(e.Code) == 0 || len(e.Tests) == 0 {
				t.Fatalf("bad entry %+v", e)
			}
			keys[e.Key] = true
			for _, p := range e.Code {
				if !cleanRelPath(p) || !validGlob(p) {
					t.Fatalf("unclean code path %q", p)
				}
			}
			for _, p := range e.Tests {
				if !cleanRelPath(p) || hasGlobMeta(p) {
					t.Fatalf("unclean test path %q", p)
				}
			}
		}
	})
}

const traceDoc = `version: 1
atlas:
  ui/toolbar:
    code: [ui/toolbar.txt, "ui/icons/**"]
    tests: [mid/mid_test.go, nvim/spec/toolbar_spec.lua]
  ui/gone:
    code: [ui/gone.txt]
    tests: [gone/x_test.go]
  ui/couch:
    code: [ui/couch.txt]
    tests: [cmd/couch/main.go]
  ui/notes:
    code: [ui/notes.txt]
`

func traceConfig(t *testing.T) (Config, Mapping) {
	t.Helper()
	cfg, err := ParseConfig([]byte(declaredConfig))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Traceability = []string{"atlas/traceability.yaml"}
	mp, err := Compile(cfg, Sources{Traceability: []TraceabilityFile{{Path: "atlas/traceability.yaml", Content: []byte(traceDoc)}}})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return cfg, mp
}

// A traceability entry's code paths select what its tests select (#616), through the one claim step.
func TestPlanTraceability(t *testing.T) {
	const m = "example.com/m/"
	cfg, mp := traceConfig(t)
	cases := []struct {
		name  string
		paths []string
		want  Selection
	}{
		{"a code path selects a Go test's package and a spec's group", []string{"ui/toolbar.txt"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli", "lua"}}},
		{"a code glob does too", []string{"ui/icons/a/b.png"}, Selection{GoPkgs: []string{m + "mid", m + "top"}, Groups: []string{"cli", "lua"}}},
		{"a test that claims nothing is full, naming the entry", []string{"ui/gone.txt"}, full("atlas/traceability.yaml ui/gone: target gone/x_test.go of ui/gone.txt claims nothing")},
		{"a test in full is full", []string{"ui/couch.txt"}, full("cmd/couch/main.go (declared by atlas/traceability.yaml ui/couch for ui/couch.txt)")},
		{"an entry without tests declares nothing", []string{"ui/notes.txt"}, full("unclassified: ui/notes.txt")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Plan(c.paths, declaredGraph(), cfg, mp, isDocTest); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Plan(%q)\n got  %+v\n want %+v", c.paths, got, c.want)
			}
		})
	}
	// Declarations only add: Owned keeps every graph owner and gains the traceability tests' packages.
	g := declaredGraph()
	for _, p := range []string{"ui/toolbar.txt", "mid/testdata/x", "ui/gone.txt", "leaf/leaf.go"} {
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
	if got := Owned("ui/toolbar.txt", g, cfg, mp); !reflect.DeepEqual(got, []string{m + "mid"}) {
		t.Fatalf("Owned(ui/toolbar.txt) = %q", got)
	}
}

// A declared traceability file that main lacks, or that does not parse, refuses: its inputs are unknown.
func TestCompileTraceabilityRefuses(t *testing.T) {
	cfg := testConfig()
	cfg.Traceability = []string{"atlas/traceability.yaml"}
	if _, err := Compile(cfg, Sources{}); err == nil || !strings.Contains(err.Error(), "traceability atlas/traceability.yaml is missing") {
		t.Fatalf("missing: %v", err)
	}
	bad := Sources{Traceability: []TraceabilityFile{{Path: "atlas/traceability.yaml", Content: []byte("version: 2\natlas: {}\n")}}}
	if _, err := Compile(cfg, bad); err == nil || !strings.Contains(err.Error(), "traceability atlas/traceability.yaml: version") {
		t.Fatalf("malformed: %v", err)
	}
	// An undeclared file is ignored: only what main's config names counts.
	cfg.Traceability = nil
	if _, err := Compile(cfg, bad); err != nil {
		t.Fatalf("undeclared file: %v", err)
	}
}
