package flakes

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestUnitRunnersLiteralProjection(t *testing.T) {
	input := `{"script":{"command":["bash","{path}"],"future":{"nested":[1,null]}},"lua":{"command":["spec-runner","{path}"],"results":"plenary-events-v1"}}`
	runners, err := ParseUnitRunners([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	for _, unit := range []string{"script:tests/-x ; $(no).sh", "lua:tests/a b.lua"} {
		id, err := ParseUnit(unit)
		if err != nil {
			t.Fatal(err)
		}
		argv, err := runners[id.Kind()].Project(id)
		if err != nil || len(argv) != 2 || argv[1] != id.Path() {
			t.Fatalf("literal projection %q: %q, %v", unit, argv, err)
		}
		argv[0] = "mutated"
		if reflect.DeepEqual(argv, runners[id.Kind()].Command) {
			t.Fatal("projection aliases declaration")
		}
	}
	got, err := json.Marshal(runners)
	if err != nil {
		t.Fatal(err)
	}
	var before, after any
	_ = json.Unmarshal([]byte(input), &before)
	_ = json.Unmarshal(got, &after)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("lost opaque nested members: %s", got)
	}
}

func TestUnitRunnersRejectMalformedAuthority(t *testing.T) {
	bad := []string{
		`null`, `{}`, `[]`,
		`{"python":{"command":["python","{path}"]}}`,
		`{"script":null}`,
		`{"script":{"command":["bash"]}}`,
		`{"script":{"command":["{path}"]}}`,
		`{"script":{"command":["bash","prefix{path}"]}}`,
		`{"script":{"command":["bash","{path}","{path}"]}}`,
		`{"script":{"command":["bash","{path}"],"command":["evil","{path}"]}}`,
		`{"script":{"command":["bash","{path}"],"Command":["evil","{path}"]}}`,
		`{"lua":{"command":["runner","{path}"]}}`,
		`{"script":{"command":["bash","{path}"],"results":"plenary-events-v1"}}`,
		`{"script":{"command":["bash","{path}"],"results":null}}`,
		`{"script":{"command":["bash","{path}","` + strings.Repeat("x", 4097) + `"]}}`,
	}
	for _, input := range bad {
		if _, err := ParseUnitRunners([]byte(input)); err == nil {
			t.Fatalf("accepted malformed authority: %.200s", input)
		}
	}
}

func FuzzUnitRunnersRoundTrip(f *testing.F) {
	f.Add(`{"script":{"command":["bash","{path}"]}}`)
	f.Add(`{"lua":{"command":["runner","{path}"],"results":"plenary-events-v1","future":[null]}}`)
	f.Fuzz(func(t *testing.T, input string) {
		r, err := ParseUnitRunners([]byte(input))
		if err != nil {
			return
		}
		raw, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseUnitRunners(raw); err != nil {
			t.Fatalf("cannot reread accepted authority: %v", err)
		}
	})
}
