package plan

import (
	"encoding/json"
	"fmt"
)

// settings.go is the pure port of construct/scripts/merge-settings.sh (ARCH-DRY:
// the bash is the source of truth; this reproduces its semantics, not its bytes).
// mergeSettings deep-merges a base (settings.ariadne.json) and an optional local
// (settings.local.json) into the composed settings.json content. No IO
// (ARCH-PURE): Apply reads base + local off disk and writes the result; this
// function only transforms in-memory bytes.
//
// Semantics, ported line-for-line from the bash's embedded python:
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
//   - Local absent → base with meta keys stripped.
//
// Output is indent-2 JSON with a trailing newline, matching the bash's
// json.dump(indent=2) + print(). The golden classifier compares on PARSED JSON
// (semantic equality), so byte-level key ordering need not match the bash.
func mergeSettings(base, local []byte) ([]byte, error) {
	var baseObj map[string]any
	if err := json.Unmarshal(base, &baseObj); err != nil {
		return nil, fmt.Errorf("mergeSettings: parse base: %w", err)
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
			return nil, fmt.Errorf("mergeSettings: parse local: %w", err)
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
		return nil, fmt.Errorf("mergeSettings: marshal result: %w", err)
	}
	out = append(out, '\n') // match the bash's trailing print().
	return out, nil
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

// containsValue reports whether list contains an element value-equal to v. JSON
// values decode to comparable scalars (string/float64/bool/nil) or composite
// types; we compare via JSON-canonical equality so list/dict items also work
// (the bash uses python's `in`, which is deep value equality). Mirrors the
// `item not in combined` / `x not in drop` checks.
func containsValue(list []any, v any) bool {
	for _, x := range list {
		if valueEqual(x, v) {
			return true
		}
	}
	return false
}

// valueEqual is deep value equality for decoded-JSON values, matching python's
// `==`. Scalars compare directly; lists/dicts compare structurally. Comparing
// the marshaled forms would be order-sensitive for dicts, so we recurse.
func valueEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			w, ok := bv[k]
			if !ok || !valueEqual(v, w) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valueEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
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
