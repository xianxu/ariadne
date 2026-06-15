package settingsx

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Merge is the pure port of construct/scripts/merge-settings.sh: deep-merge
// dicts (local overrides base at matching paths), union arrays at the dotted
// paths listed in base's $merge_keys (base order, then new local items), apply
// $remove (filter base's array at a dotted path BEFORE the union), strip the
// $comment/$merge_keys/$remove meta keys, and — with no local — emit base with
// meta keys stripped. No IO (ARCH-PURE). We compare on PARSED JSON (semantic
// equality), not bytes.

// mustParse parses JSON into a map or fails.
func mustParse(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse merge output: %v\n--- output ---\n%s", err, b)
	}
	return m
}

// runMerge runs Merge and returns the parsed result, failing on error.
func runMerge(t *testing.T, base, local string) map[string]any {
	t.Helper()
	var localBytes []byte
	if local != "" {
		localBytes = []byte(local)
	}
	out, err := Merge([]byte(base), localBytes)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	return mustParse(t, out)
}

func TestMergeLocalAbsentStripsMeta(t *testing.T) {
	base := `{
		"$comment": "doc",
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A", "B"]},
		"scalar": 1
	}`
	got := runMerge(t, base, "")
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"A", "B"}},
		"scalar":      float64(1),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("local-absent strip:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeDeepDictMerge(t *testing.T) {
	base := `{"a": {"x": 1, "y": 2}, "b": 10}`
	local := `{"a": {"y": 99, "z": 3}, "c": 20}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"a": map[string]any{
			"x": float64(1),  // base-only nested key survives
			"y": float64(99), // local overrides at matching path
			"z": float64(3),  // local-only nested key added
		},
		"b": float64(10), // base-only top key survives
		"c": float64(20), // local-only top key added
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("deep dict merge:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeScalarOverride(t *testing.T) {
	base := `{"n": 1, "s": "old", "keep": true}`
	local := `{"n": 2, "s": "new"}`
	got := runMerge(t, base, local)
	want := map[string]any{"n": float64(2), "s": "new", "keep": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scalar override:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeArrayUnionAtMergeKey(t *testing.T) {
	// An array at a path in $merge_keys is UNIONED: base order first, then new
	// local items (an item already in base is not re-appended).
	base := `{"$merge_keys": ["permissions.allow"], "permissions": {"allow": ["A", "B"]}}`
	local := `{"permissions": {"allow": ["B", "C", "D"]}}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"A", "B", "C", "D"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge-key union:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeArrayReplaceWhenNotMergeKey(t *testing.T) {
	// An array at a path NOT in $merge_keys is REPLACED by local wholesale.
	base := `{"$merge_keys": ["permissions.allow"], "list": ["a", "b", "c"]}`
	local := `{"list": ["x"]}`
	got := runMerge(t, base, local)
	want := map[string]any{"list": []any{"x"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("non-merge-key array replace:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeRemoveFiltersBeforeUnion(t *testing.T) {
	// $remove filters base's array at the dotted path BEFORE the union step.
	// Classic case: drop "Bash(rm:*)" from permissions.allow, add it to deny.
	base := `{
		"$merge_keys": ["permissions.allow", "permissions.deny"],
		"permissions": {
			"allow": ["Bash(rm:*)", "Bash(ls:*)", "WebSearch"],
			"deny": ["Bash(ssh *)"]
		}
	}`
	local := `{
		"$remove": {"permissions.allow": ["Bash(rm:*)"]},
		"permissions": {"deny": ["Bash(rm:*)"]}
	}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Bash(ls:*)", "WebSearch"},
			"deny":  []any{"Bash(ssh *)", "Bash(rm:*)"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-before-union:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeRemoveItemNotPresentIgnored(t *testing.T) {
	base := `{"$merge_keys": ["permissions.allow"], "permissions": {"allow": ["A", "B"]}}`
	local := `{"$remove": {"permissions.allow": ["NOPE"]}, "permissions": {"allow": ["C"]}}`
	got := runMerge(t, base, local)
	want := map[string]any{"permissions": map[string]any{"allow": []any{"A", "B", "C"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-absent-ignored:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeMetaKeysStrippedFromOutput(t *testing.T) {
	base := `{"$comment": "base doc", "$merge_keys": ["x"], "x": [1]}`
	local := `{"$comment": "local doc", "$remove": {"x": [9]}, "x": [2]}`
	got := runMerge(t, base, local)
	for _, meta := range []string{"$comment", "$merge_keys", "$remove"} {
		if _, ok := got[meta]; ok {
			t.Fatalf("meta key %q leaked into output: %#v", meta, got)
		}
	}
}

func TestMergeRemoveAtScalarPathNoop(t *testing.T) {
	// $remove at a non-array (scalar) base path is a no-op; the scalar is then
	// overridden by local per the normal deep-merge rule.
	base := `{"s": "base"}`
	local := `{"$remove": {"s": ["x"]}, "s": "local"}`
	got := runMerge(t, base, local)
	want := map[string]any{"s": "local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-at-scalar:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeNestedMergeKeyPath(t *testing.T) {
	// A deeply-nested (dotted) $merge_keys path resolves through the recursion.
	base := `{
		"$merge_keys": ["sandbox.network.allowedDomains"],
		"sandbox": {"network": {"allowedDomains": ["a.com", "b.com"]}}
	}`
	local := `{"sandbox": {"network": {"allowedDomains": ["b.com", "c.com"]}}}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"sandbox": map[string]any{
			"network": map[string]any{"allowedDomains": []any{"a.com", "b.com", "c.com"}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nested merge-key path:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestSemanticEqual(t *testing.T) {
	// Key ordering + whitespace differ but the structures are equal → true.
	a := []byte(`{"b": 2, "a": [1, 2]}`)
	b := []byte("{\n  \"a\": [1, 2],\n  \"b\": 2\n}\n")
	eq, err := SemanticEqual(a, b)
	if err != nil {
		t.Fatalf("SemanticEqual: %v", err)
	}
	if !eq {
		t.Fatalf("SemanticEqual = false, want true (key order should not matter)")
	}

	// A real difference → false.
	c := []byte(`{"a": [1, 3], "b": 2}`)
	eq, err = SemanticEqual(a, c)
	if err != nil {
		t.Fatalf("SemanticEqual: %v", err)
	}
	if eq {
		t.Fatalf("SemanticEqual = true, want false (arrays differ)")
	}

	// A parse error surfaces.
	if _, err := SemanticEqual([]byte(`{bad`), b); err == nil {
		t.Fatalf("SemanticEqual on bad JSON: want error, got nil")
	}
}
