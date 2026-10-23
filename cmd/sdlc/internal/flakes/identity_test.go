package flakes

import (
	"reflect"
	"strings"
	"testing"
)

func TestIdentityParsing(t *testing.T) {
	for _, tc := range []struct {
		input                      string
		kind                       IdentityKind
		unit, path, label, display string
	}{
		{"github.com/x/pkg.TestA/sub", KindGo, "github.com/x/pkg.TestA", "", "", "TestA/sub"},
		{"script:scripts/a check.sh", KindScript, "script:scripts/a check.sh", "scripts/a check.sh", "", "script:scripts/a check.sh"},
		{"lua:tests/单元.lua# case/a # nested ", KindLua, "lua:tests/单元.lua", "tests/单元.lua", " case/a # nested ", "lua:tests/单元.lua# case/a # nested "},
	} {
		t.Run(tc.input, func(t *testing.T) {
			id, err := ParseIdentity(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if id.Kind() != tc.kind || id.Unit() != tc.unit || id.Path() != tc.path || id.Case() != tc.label || id.Display() != tc.display || id.String() != tc.input {
				t.Errorf("wrong components: kind=%q unit=%q path=%q case=%q display=%q string=%q", id.Kind(), id.Unit(), id.Path(), id.Case(), id.Display(), id.String())
			}
			n, ok := id.GoName()
			if ok != (tc.kind == KindGo) || (ok && n.String() != tc.input) {
				t.Errorf("GoName() = %+v, %v", n, ok)
			}
			again, err := ParseIdentity(id.String())
			if err != nil || again != id {
				t.Errorf("round trip = %+v, %v", again, err)
			}
		})
	}
}

func TestIdentityRejectsMalformed(t *testing.T) {
	badPaths := []string{"", "/a.sh", "-a.sh", "a//b.sh", "./a.sh", "a/../b.sh", "a/./b.sh", "a\\b.sh", "a#b.sh", "a\nb.sh", "a\x7fb.sh", "a\u0085b.sh", "a\xff.sh", "a.lua", strings.Repeat("x", 510) + ".sh"}
	for _, p := range badPaths {
		if _, err := ParseIdentity("script:" + p); err == nil {
			t.Errorf("accepted path %q", p)
		}
	}
	for _, input := range []string{"", "script:a.sh#case", "lua:a.lua", "lua:a.lua#", "lua:a.lua#bad\ncase", "lua:a.lua#bad\x7fcase", "lua:a.lua#\xff", "lua:a.lua#" + strings.Repeat("é", 1025), "python:a.py#case", "Script:a.sh"} {
		if _, err := ParseIdentity(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	for _, input := range []string{"script:" + strings.Repeat("a", 509) + ".sh", "script:a/-b.sh", "lua:a.lua#" + strings.Repeat("é", 1024), "lua:a.lua# "} {
		if _, err := ParseIdentity(input); err != nil {
			t.Errorf("boundary %q: %v", input, err)
		}
	}
}

func TestTypedClassificationPreservesCaseIdentity(t *testing.T) {
	a, b := "lua:tests/unit.lua#case", "lua:tests/unit.lua#case/child"
	r, err := BuildRegistry([]Card{{ID: "1", Tests: []string{a}}}, func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	got := Classify(Input{Attempt: Result{Failures: []string{a}, Subtests: []string{b}}, Registry: r, Policy: PolicyAccept})
	if got.Kind != KnownFlakeAccept {
		t.Errorf("Lua case slash became ancestry: %+v", got)
	}
	got = Classify(Input{Attempt: Result{Failures: []string{b}}, Registry: r, Policy: PolicyAccept, MainBaseline: &Result{Failures: []string{a}}})
	if got.Kind != Unknown {
		t.Errorf("different Lua case treated as same baseline: %+v", got)
	}
	got = Classify(Input{Attempt: Result{Failures: []string{b}}, Registry: r, Policy: PolicyAccept, MainBaseline: &Result{Failures: []string{b}}})
	if got.Kind != PersistentMain {
		t.Errorf("exact Lua case missing baseline: %+v", got)
	}
	if got := RerunTests([]Entry{{Test: a}, {Test: b}, {Test: a}, {Test: "pkg.TestA/sub"}}); !reflect.DeepEqual(got, []string{a, b, "pkg.TestA"}) {
		t.Errorf("RerunTests = %q", got)
	}
}

func TestTypedCoverageIsExactAcrossKinds(t *testing.T) {
	for _, tc := range []struct {
		registered, failure string
		want                bool
	}{
		{"script:tests/check.sh", "script:tests/check.sh", true},
		{"script:tests/check.sh", "script:tests/check.sh/sub", false},
		{"lua:tests/unit.lua#x", "lua:tests/unit.lua#x", true},
		{"lua:tests/unit.lua#x", "lua:tests/unit.lua#x/y", false},
		{"lua:tests/unit.lua#x/y", "lua:tests/unit.lua#x", false},
		{"lua:tests/unit.lua#x", "lua:tests/unit.lua#x#other", false},
		{"lua:tests/unit.lua#x", "lua:tests/other.lua#x", false},
		{"lua:tests/unit.lua#x", "lua:tests/unit.lua# x", false},
		{"pkg.TestA", "script:pkg.TestA.sh", false},
		{"pkg.TestA", "lua:pkg.TestA.lua#x", false},
		{"script:tests/check.sh", "lua:tests/check.lua#x", false},
		{"lua:tests/unit.lua#x", "pkg.TestA/x", false},
	} {
		if got := Covers(tc.registered, tc.failure); got != tc.want {
			t.Errorf("Covers(%q, %q) = %v, want %v", tc.registered, tc.failure, got, tc.want)
		}
	}
}

func TestParseExecutionUnit(t *testing.T) {
	for _, input := range []string{"script:scripts/check.sh", "lua:tests/unit.lua"} {
		id, err := ParseUnit(input)
		if err != nil {
			t.Fatal(err)
		}
		if id.Unit() != input || id.String() != input || id.Case() != "" {
			t.Errorf("unit: %+v", id)
		}
	}
	for _, input := range []string{"pkg.TestA", "lua:tests/unit.lua#case", "script:../check.sh", "lua:/unit.lua"} {
		if _, err := ParseUnit(input); err == nil {
			t.Errorf("accepted unit %q", input)
		}
	}
}

func FuzzIdentityRoundTrip(f *testing.F) {
	for _, input := range []string{"pkg.TestA/sub", "script:scripts/check.sh", "lua:tests/unit.lua#case/a # b", "lua:a.lua# ", "script:../bad.sh"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseIdentity(input)
		if err != nil {
			return
		}
		if id.String() != input {
			t.Fatalf("canonical identity changed %q to %q", input, id.String())
		}
		again, err := ParseIdentity(id.String())
		if err != nil || again != id {
			t.Fatalf("roundtrip %q: %+v, %v", input, again, err)
		}
		if id.Kind() != KindGo {
			unit, err := ParseUnit(id.Unit())
			if err != nil || unit.Unit() != id.Unit() || unit.Path() != id.Path() {
				t.Fatalf("unit roundtrip %q: %+v, %v", input, unit, err)
			}
		}
	})
}

func TestTypedRegistrationAndCoverage(t *testing.T) {
	for _, name := range []string{"script:scripts/check.sh", "lua:tests/unit.lua#case/with # punctuation"} {
		if err := ValidTest(name); err != nil {
			t.Errorf("ValidTest(%q): %v", name, err)
		}
		if !Covers(name, name) {
			t.Errorf("identity does not cover itself: %q", name)
		}
		if Covers(name, name+"/child") {
			t.Errorf("typed identity covers an apparent descendant: %q", name)
		}
		if got := Leaves(name, []string{name + "/child"}); !reflect.DeepEqual(got, []string{name}) {
			t.Errorf("Leaves(%q) = %q", name, got)
		}
	}
	if !Covers("pkg.TestA", "pkg.TestA/非ASCII case") {
		t.Error("valid Go ancestor must cover unrestricted emitted descendant labels")
	}
	if Covers("script:tests/check.sh", "lua:tests/check.lua#case") {
		t.Error("cross-kind coverage")
	}
}

func TestTypedRegistryAmbiguity(t *testing.T) {
	a, b := "lua:tests/unit.lua#case", "lua:tests/unit.lua#case/child"
	cards := []Card{{ID: "1", Status: "open", Tests: []string{a}}, {ID: "2", Status: "open", Tests: []string{b}}}
	r, err := BuildRegistry(cards, func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	for name, issue := range map[string]string{a: "1", b: "2"} {
		if got, ok := r.Cover(name); !ok || got.Issue != issue {
			t.Errorf("Cover(%q) = %+v, %v", name, got, ok)
		}
	}
	if _, ok := r.Cover(b + "/other"); ok {
		t.Error("registry must not cover an unregistered Lua case")
	}
	cards[1].Tests = []string{a}
	if _, err := BuildRegistry(cards, func(string) bool { return true }); err == nil {
		t.Error("duplicate exact typed identity accepted")
	}
}
