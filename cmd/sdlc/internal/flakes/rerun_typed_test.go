package flakes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRerunPlanMixedInventory(t *testing.T) {
	selection := []string{
		"lua:tests/z.lua# case/a # suffix ", "pkg/z.TestZ/sub", "script:scripts/z.sh",
		"lua:tests/a.lua#other", "pkg/a.TestB", "lua:tests/z.lua#different",
		"pkg/a.TestA/sub", "script:scripts/a.sh", "script:scripts/z.sh", "pkg/a.TestA/other",
	}
	prefix := []string{"custom-sdlc", "--root", "a root", "flake", "run"}
	commands, err := RerunPlan(selection, prefix)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"go", "test", "-count=1", "-json", "-run", "^(TestA|TestB)$", "pkg/a"},
		{"go", "test", "-count=1", "-json", "-run", "^(TestZ)$", "pkg/z"},
		{"custom-sdlc", "--root", "a root", "flake", "run", "--unit", "lua:tests/a.lua"},
		{"custom-sdlc", "--root", "a root", "flake", "run", "--unit", "lua:tests/z.lua"},
		{"custom-sdlc", "--root", "a root", "flake", "run", "--unit", "script:scripts/a.sh"},
		{"custom-sdlc", "--root", "a root", "flake", "run", "--unit", "script:scripts/z.sh"},
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
	commands[2][0] = "changed"
	if prefix[0] != "custom-sdlc" || commands[3][0] != "custom-sdlc" {
		t.Fatal("returned argv aliases prefix or another command")
	}
	if selection[0] != "lua:tests/z.lua# case/a # suffix " {
		t.Fatal("mutated exact selected identity")
	}
}

func TestRerunPlanInputValidation(t *testing.T) {
	for _, prefix := range [][]string{nil, {}, {""}, {"sdlc", ""}, {"sdlc\x00"}, {"sdlc", "bad\x00arg"}} {
		if _, err := RerunPlan([]string{"script:scripts/a.sh"}, prefix); err == nil {
			t.Errorf("accepted prefix %q", prefix)
		}
	}
	for _, selection := range [][]string{nil, {"lua:tests/a.lua"}, {"script:../a.sh"}, {"pkg.TestA", "bad"}, {"script:scripts/a.sh", "lua:tests/a.lua#"}} {
		if _, err := RerunPlan(selection, []string{"sdlc", "flake", "run"}); err == nil {
			t.Errorf("accepted selection %q", selection)
		}
	}
	if commands, err := RerunPlan([]string{"pkg.TestA/sub"}, nil); err != nil || !reflect.DeepEqual(commands, [][]string{{"go", "test", "-count=1", "-json", "-run", "^(TestA)$", "pkg"}}) {
		t.Fatalf("Go-only = %q, %v", commands, err)
	}
}

// Independent fixtures state the required case inventory instead of deriving
// the oracle with the same grouping or proof helpers as production.
func TestRerunMissingProofExactInventory(t *testing.T) {
	log := typedRerunProofFixture(t, "lua:tests/unit.lua", []CaseResult{{Name: "case", Result: "pass"}, {Name: "case/child", Result: "pass"}, {Name: " spaced # case ", Result: "pass"}}, true)
	log += typedRerunProofFixture(t, "script:scripts/check.sh", nil, true)
	log += "{\"Action\":\"pass\",\"Package\":\"pkg\",\"Test\":\"TestA/sub\"}\n{\"Action\":\"pass\",\"Package\":\"pkg\",\"Test\":\"TestA\"}\n{\"Action\":\"pass\",\"Package\":\"pkg\"}\n"
	selection := []string{"lua:tests/unit.lua#case", "lua:tests/unit.lua#case/child", "lua:tests/unit.lua# spaced # case ", "script:scripts/check.sh", "pkg.TestA/sub", "lua:tests/unit.lua#case/child/missing", "lua:tests/other.lua#case", "lua:tests/unit.lua#spaced # case", "script:scripts/other.sh", "pkg.TestA/missing", "bad", "bad", "lua:tests/unit.lua"}
	want := []string{"bad", "lua:tests/other.lua#case", "lua:tests/unit.lua", "lua:tests/unit.lua#case/child/missing", "lua:tests/unit.lua#spaced # case", "pkg.TestA/missing", "script:scripts/other.sh"}
	if got := RerunMissingProof(log, selection); !reflect.DeepEqual(got, want) {
		t.Fatalf("missing = %q, want %q", got, want)
	}
}

func TestRerunMissingProofRejectsIncompleteOrFailedUnit(t *testing.T) {
	selected := []string{"lua:tests/unit.lua#case", "script:scripts/check.sh"}
	for _, log := range []string{"", "--- PASS: case\nscript:scripts/check.sh passed\n", typedRerunProofFixture(t, "lua:tests/unit.lua", []CaseResult{{Name: "case", Result: "pass"}}, false), typedRerunProofFixture(t, "lua:tests/unit.lua", []CaseResult{{Name: "case", Result: "fail"}}, true)} {
		if got := RerunMissingProof(log, selected); !reflect.DeepEqual(got, selected) {
			t.Errorf("missing = %q for log %q", got, log)
		}
	}
}

func TestRerunPlanUnitInventoryAtScale(t *testing.T) {
	for _, count := range []int{8, 16} {
		var selection []string
		for i := count - 1; i >= 0; i-- {
			selection = append(selection, fmt.Sprintf("lua:tests/%02d.lua#case/a", i), fmt.Sprintf("lua:tests/%02d.lua#different", i))
		}
		commands, err := RerunPlan(selection, []string{"runner"})
		if err != nil {
			t.Fatal(err)
		}
		if len(commands) != count {
			t.Fatalf("%d files produced %d commands", count, len(commands))
		}
		for i, cmd := range commands {
			want := []string{"runner", "--unit", fmt.Sprintf("lua:tests/%02d.lua", i)}
			if !reflect.DeepEqual(cmd, want) {
				t.Errorf("command %d = %q, want %q", i, cmd, want)
			}
		}
	}
}

func typedRerunProofFixture(t *testing.T, unit string, cases []CaseResult, finish bool) string {
	t.Helper()
	invocation := strings.Repeat("a", 32)
	events := []UnitEvent{{Version: 1, Invocation: invocation, Unit: unit, Event: "start"}}
	exit := 0
	for _, c := range cases {
		events = append(events, UnitEvent{Version: 1, Invocation: invocation, Unit: unit, Event: "case", Case: c.Name, Result: c.Result})
		if c.Result == "fail" {
			exit = 1
		}
	}
	if finish {
		events = append(events, UnitEvent{Version: 1, Invocation: invocation, Unit: unit, Event: "finish", Exit: &exit, Complete: true})
	}
	var log strings.Builder
	for _, event := range events {
		raw, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		log.Write(raw)
		log.WriteByte('\n')
	}
	return log.String()
}
