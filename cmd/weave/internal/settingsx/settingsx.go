// Package settingsx is the ONE home for weave's pure settings-merge reasoning
// (ARCH-DRY, ARCH-PURE), the port of construct/scripts/merge-settings.sh. Two
// consumers need it: plan.Apply (the IO seam reads base + local, calls Merge,
// writes the target) and the golden classifier (it recomputes Merge from the
// observed base + local and asks SemanticEqual whether the live settings.json
// matches). It sits below plan and golden with no internal imports, so both
// import it without a cycle. No IO: it transforms in-memory bytes only.
//
// merge-settings.sh is the source of truth; this reproduces its embedded
// python's deep_merge / get_nested / set_nested / strip_meta semantics
// line-for-line. SemanticEqual compares PARSED JSON (not bytes) because the bash
// (jq/python) and weave need not agree on key ordering.
package settingsx

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Merge deep-merges a base (settings.ariadne.json) and an optional local
// (settings.local.json) into the composed settings.json content. local == nil
// is the local-absent case (base with meta keys stripped). Semantics, ported
// from the bash:
//
//   - Dicts deep-merge: at a matching key, recurse; local-only keys are added;
//     base-only keys are kept. ($-prefixed meta keys are skipped on both sides.)
//   - Arrays at a dotted path listed in base's $merge_keys are UNIONED: base
//     order first, then each new local item not already present (value equality).
//   - $remove (in local): {"$remove": {"<dotted.path>": [items]}} filters base's
//     array at that path — dropping matching items — BEFORE the union step. A
//     non-array target is left untouched. Items not in base are ignored.
//   - Arrays at any other path are REPLACED by local wholesale.
//   - Scalars: local replaces base.
//   - The $comment / $merge_keys / $remove meta keys are stripped from output.
//
// Output is indent-2 JSON with a trailing newline, matching the bash's
// json.dump(indent=2) + print().
func Merge(base, local []byte) ([]byte, error) {
	var baseObj map[string]any
	if err := json.Unmarshal(base, &baseObj); err != nil {
		return nil, fmt.Errorf("settingsx.Merge: parse base: %w", err)
	}

	// merge_keys = set(base.get('$merge_keys', [])) — the dotted paths whose
	// arrays union rather than replace.
	mergeKeys := map[string]bool{}
	if raw, ok := baseObj["$merge_keys"].([]any); ok {
		for _, k := range raw {
			if s, ok := k.(string); ok {
				mergeKeys[s] = true
			}
		}
	}

	var result map[string]any
	if local == nil {
		// Local absent → base with meta keys stripped.
		result = stripMeta(baseObj).(map[string]any)
	} else {
		var localObj map[string]any
		if err := json.Unmarshal(local, &localObj); err != nil {
			return nil, fmt.Errorf("settingsx.Merge: parse local: %w", err)
		}

		// Apply $remove against base BEFORE merging (the bash filters a deep copy
		// of base, then merges strip_meta(base_filtered) with local).
		baseForMerge := baseObj
		if removals, ok := localObj["$remove"].(map[string]any); ok && len(removals) > 0 {
			baseForMerge = applyRemovals(baseObj, removals)
		}
		merged := deepMerge(stripMeta(baseForMerge), localObj, "", mergeKeys)
		// deepMerge over two dicts always yields a dict here (both are objects).
		result, _ = merged.(map[string]any)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("settingsx.Merge: marshal result: %w", err)
	}
	out = append(out, '\n') // match the bash's trailing print().
	return out, nil
}

// SemanticEqual reports whether two JSON byte slices decode to deeply-equal
// values, ignoring key ordering and formatting. Used by the golden classifier to
// compare weave's Merge output against the live settings.json (which the bash
// produced with possibly-different key ordering — a semantically-equal file is
// not a divergence). Returns an error if either side fails to parse.
func SemanticEqual(a, b []byte) (bool, error) {
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		return false, fmt.Errorf("settingsx.SemanticEqual: parse a: %w", err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false, fmt.Errorf("settingsx.SemanticEqual: parse b: %w", err)
	}
	return reflect.DeepEqual(av, bv), nil
}

// stripMeta returns obj with every $-prefixed key removed recursively from
// dicts (ports strip_meta). Non-dicts pass through unchanged.
func stripMeta(obj any) any {
	m, ok := obj.(map[string]any)
	if !ok {
		return obj
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if len(k) > 0 && k[0] == '$' {
			continue
		}
		out[k] = stripMeta(v)
	}
	return out
}

// applyRemovals returns a deep copy of base with each $remove dotted path's
// array filtered to drop the listed items (ports the $remove block). A path
// pointing at a non-array (or absent) is left untouched; items not present are
// ignored. The bash deep-copies base via json round-trip before mutating; we
// likewise copy so the caller's base is not mutated.
func applyRemovals(base map[string]any, removals map[string]any) map[string]any {
	filtered := deepCopy(base).(map[string]any)
	for path, raw := range removals {
		items, ok := raw.([]any)
		if !ok {
			continue
		}
		current := getNested(filtered, path)
		arr, ok := current.([]any)
		if !ok {
			continue // not an array — no-op (the bash's isinstance(current, list) guard)
		}
		drop := make([]any, 0, len(arr))
		for _, x := range arr {
			if !containsValue(items, x) {
				drop = append(drop, x)
			}
		}
		setNested(filtered, path, drop)
	}
	return filtered
}

// deepMerge ports the bash's deep_merge(b, l, path):
//
//   - both dicts → merge key-wise, skipping $-keys on both sides; recurse at a
//     shared key (extending path with .key), keep base-only keys, add local-only;
//   - both lists → union (base order, then new local items by value) iff path is
//     in mergeKeys, else local replaces base;
//   - otherwise → local replaces base.
func deepMerge(b, l any, path string, mergeKeys map[string]bool) any {
	bDict, bIsDict := b.(map[string]any)
	lDict, lIsDict := l.(map[string]any)
	if bIsDict && lIsDict {
		out := map[string]any{}
		for k, bv := range bDict {
			if len(k) > 0 && k[0] == '$' {
				continue
			}
			sub := k
			if path != "" {
				sub = path + "." + k
			}
			if lv, ok := lDict[k]; ok {
				out[k] = deepMerge(bv, lv, sub, mergeKeys)
			} else {
				out[k] = bv
			}
		}
		for k, lv := range lDict {
			if (len(k) > 0 && k[0] == '$') || hasKey(bDict, k) {
				continue
			}
			out[k] = lv
		}
		return out
	}

	bList, bIsList := b.([]any)
	lList, lIsList := l.([]any)
	if bIsList && lIsList {
		if mergeKeys[path] {
			combined := make([]any, len(bList))
			copy(combined, bList)
			for _, item := range lList {
				if !containsValue(combined, item) {
					combined = append(combined, item)
				}
			}
			return combined
		}
		return lList
	}

	return l
}

// getNested walks a dotted path through nested dicts, returning nil if any
// segment is missing or not a dict (ports get_nested).
func getNested(obj map[string]any, path string) any {
	var cur any = obj
	for _, p := range splitDots(path) {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		v, ok := m[p]
		if !ok {
			return nil
		}
		cur = v
	}
	return cur
}

// setNested sets value at a dotted path, creating intermediate dicts as needed
// (ports set_nested's obj.setdefault(p, {})).
func setNested(obj map[string]any, path string, value any) {
	parts := splitDots(path)
	cur := obj
	for _, p := range parts[:len(parts)-1] {
		next, ok := cur[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[p] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value
}

// splitDots splits a dotted path into segments (path.split('.')).
func splitDots(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}

// hasKey reports whether m contains key k.
func hasKey(m map[string]any, k string) bool {
	_, ok := m[k]
	return ok
}

// containsValue reports whether list contains an element value-equal to v.
// Mirrors python's `in` (deep value equality), so list/dict items also compare
// correctly — matching the bash's `item not in combined` / `x not in drop`.
func containsValue(list []any, v any) bool {
	for _, x := range list {
		if reflect.DeepEqual(x, v) {
			return true
		}
	}
	return false
}

// deepCopy returns a structural copy of a decoded-JSON value, so applyRemovals
// can mutate base's arrays without touching the caller's object (ports the
// json.loads(json.dumps(base)) deep copy).
func deepCopy(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(vv))
		for k, val := range vv {
			out[k] = deepCopy(val)
		}
		return out
	case []any:
		out := make([]any, len(vv))
		for i, val := range vv {
			out[i] = deepCopy(val)
		}
		return out
	default:
		return v
	}
}
