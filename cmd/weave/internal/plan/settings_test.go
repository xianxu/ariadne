package plan

import (
	"encoding/json"
	"reflect"
	"testing"
)

// mergeSettings is the pure port of construct/scripts/merge-settings.sh: deep-
// merge dicts (local overrides base at matching paths), union arrays at the
// dotted paths listed in base's $merge_keys (base order, then new local items),
// apply $remove (filter base's array at a dotted path BEFORE the union), strip
// the $comment/$merge_keys/$remove meta keys from the output, and — with no
// local — emit base with meta keys stripped. No IO (ARCH-PURE); the caller reads
// the bytes off disk in Apply.
//
// We compare on PARSED JSON (semantic equality), not bytes: merge-settings.sh
// (jq/python) and weave need not agree on key ordering, and the rules are about
// structure, not serialization.

// mustParse is a test helper: parse JSON into a map[string]any or fail.
func mustParse(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse merge output: %v\n--- output ---\n%s", err, b)
	}
	return m
}

// runMerge runs mergeSettings and returns the parsed result, failing on error.
func runMerge(t *testing.T, base, local string) map[string]any {
	t.Helper()
	var localBytes []byte
	if local != "" {
		localBytes = []byte(local)
	}
	out, err := mergeSettings([]byte(base), localBytes)
	if err != nil {
		t.Fatalf("mergeSettings: %v", err)
	}
	return mustParse(t, out)
}

func TestMergeLocalAbsentStripsMeta(t *testing.T) {
	// No local → base with meta keys ($comment/$merge_keys/$remove) stripped.
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
	// Local keys override base at matching nested paths; non-overlapping keys
	// from both sides survive (deep dict merge).
	base := `{
		"a": {"x": 1, "y": 2},
		"b": 10
	}`
	local := `{
		"a": {"y": 99, "z": 3},
		"c": 20
	}`
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
	// A scalar: local replaces base.
	base := `{"n": 1, "s": "old", "keep": true}`
	local := `{"n": 2, "s": "new"}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"n":    float64(2),
		"s":    "new",
		"keep": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scalar override:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeArrayUnionAtMergeKey(t *testing.T) {
	// An array at a path in $merge_keys is UNIONED: base order first, then new
	// local items (dedup: an item already in base is not re-appended).
	base := `{
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A", "B"]}
	}`
	local := `{
		"permissions": {"allow": ["B", "C", "D"]}
	}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"permissions": map[string]any{
			// base order A,B preserved; B already present (skip); C,D appended.
			"allow": []any{"A", "B", "C", "D"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge-key union:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeArrayReplaceWhenNotMergeKey(t *testing.T) {
	// An array at a path NOT in $merge_keys is REPLACED by local wholesale.
	base := `{
		"$merge_keys": ["permissions.allow"],
		"list": ["a", "b", "c"]
	}`
	local := `{"list": ["x"]}`
	got := runMerge(t, base, local)
	want := map[string]any{"list": []any{"x"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("non-merge-key array replace:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeRemoveFiltersBeforeUnion(t *testing.T) {
	// $remove filters base's array at the dotted path BEFORE the union step, so
	// an item dropped from base then re-added to a different path is honored.
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
		"permissions": {
			"deny": ["Bash(rm:*)"]
		}
	}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"permissions": map[string]any{
			// allow: Bash(rm:*) filtered out BEFORE union; local adds nothing here.
			"allow": []any{"Bash(ls:*)", "WebSearch"},
			// deny: base order first, then the new Bash(rm:*).
			"deny": []any{"Bash(ssh *)", "Bash(rm:*)"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-before-union:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeRemoveItemNotPresentIgnored(t *testing.T) {
	// $remove items not present in base's array are silently ignored.
	base := `{
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A", "B"]}
	}`
	local := `{
		"$remove": {"permissions.allow": ["NOPE"]},
		"permissions": {"allow": ["C"]}
	}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"A", "B", "C"}, // NOPE absent; nothing removed.
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-absent-ignored:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeMetaKeysStrippedFromOutput(t *testing.T) {
	// $comment / $merge_keys / $remove never appear in the output, even when both
	// base and local carry them.
	base := `{
		"$comment": "base doc",
		"$merge_keys": ["x"],
		"x": [1]
	}`
	local := `{
		"$comment": "local doc",
		"$remove": {"x": [9]},
		"x": [2]
	}`
	got := runMerge(t, base, local)
	for _, meta := range []string{"$comment", "$merge_keys", "$remove"} {
		if _, ok := got[meta]; ok {
			t.Fatalf("meta key %q leaked into output: %#v", meta, got)
		}
	}
}

func TestMergeRemoveAtScalarPathNoop(t *testing.T) {
	// $remove pointed at a non-array (scalar) base path is a no-op — the bash's
	// `if isinstance(current, list)` guard. The scalar is then overridden by
	// local per the normal deep-merge rule.
	base := `{"s": "base"}`
	local := `{
		"$remove": {"s": ["x"]},
		"s": "local"
	}`
	got := runMerge(t, base, local)
	want := map[string]any{"s": "local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove-at-scalar:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestMergeNestedMergeKeyPath(t *testing.T) {
	// A $merge_keys path that's deeply nested (dotted) resolves correctly through
	// the deep-merge recursion — verifies path tracking matches the dotted key.
	base := `{
		"$merge_keys": ["sandbox.network.allowedDomains"],
		"sandbox": {"network": {"allowedDomains": ["a.com", "b.com"]}}
	}`
	local := `{
		"sandbox": {"network": {"allowedDomains": ["b.com", "c.com"]}}
	}`
	got := runMerge(t, base, local)
	want := map[string]any{
		"sandbox": map[string]any{
			"network": map[string]any{
				"allowedDomains": []any{"a.com", "b.com", "c.com"},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nested merge-key path:\n got=%#v\nwant=%#v", got, want)
	}
}
