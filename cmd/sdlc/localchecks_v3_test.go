package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLocalChecksV3TypedRunnerContract(t *testing.T) {
	input := `{"version":3,"future":{"keep":true},"checks":[{"name":"suite","command":["make","test"],"future_check":[1],"flake_runners":{"script":{"command":["bash","{path}"],"future_runner":{"keep":true}},"lua":{"command":["scripts/spec.sh","{path}"],"results":"plenary-events-v1"}}}]}`
	d, err := parseLocalChecks([]byte(input))
	if err != nil {
		t.Fatalf("typed declaration refused: %v", err)
	}
	if d.Version != 3 || len(d.Checks) != 1 {
		t.Fatalf("wrong typed declaration: %+v", d)
	}
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var before, after any
	if err := json.Unmarshal([]byte(input), &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &after); err != nil {
		t.Fatal(err)
	}
	// Defaults may be serialized, so assert opaque values at each nesting level.
	m := after.(map[string]any)
	c := m["checks"].([]any)[0].(map[string]any)
	r := c["flake_runners"].(map[string]any)["script"].(map[string]any)
	if m["future"].(map[string]any)["keep"] != true || c["future_check"].([]any)[0] != float64(1) || r["future_runner"].(map[string]any)["keep"] != true {
		t.Fatalf("lost opaque data: %s", raw)
	}
	if _, err := parseLocalChecks(raw); err != nil {
		t.Fatalf("cannot reread encoded typed declaration: %v", err)
	}
}

func TestLocalChecksV3RejectsRunnerAuthorityMutation(t *testing.T) {
	for _, runners := range []string{
		`null`,
		`{"python":{"command":["python","{path}"]}}`,
		`{"script":{"command":["bash","prefix{path}"]}}`,
		`{"script":{"command":["bash","{path}","{path}"]}}`,
		`{"script":{"command":["bash","{path}"],"Command":["evil","{path}"]}}`,
		`{"lua":{"command":["runner","{path}"]}}`,
		`{"script":{"command":["bash","{path}"],"results":"plenary-events-v1"}}`,
	} {
		t.Run(runners, func(t *testing.T) {
			input := `{"version":3,"checks":[{"name":"suite","command":["make","test"],"flake_runners":` + runners + `}]}`
			if _, err := parseLocalChecks([]byte(input)); err == nil {
				t.Fatalf("accepted invalid runner authority: %s", input)
			}
		})
	}
	for _, name := range []string{"lint", "landing-check"} {
		input := `{"version":3,"checks":[{"name":"` + name + `","command":["true"],"flake_runners":{"script":{"command":["bash","{path}"]}}}]}`
		if _, err := parseLocalChecks([]byte(input)); err == nil || !strings.Contains(err.Error(), "suite") {
			t.Fatalf("non-suite runner declaration: %v", err)
		}
	}
}
