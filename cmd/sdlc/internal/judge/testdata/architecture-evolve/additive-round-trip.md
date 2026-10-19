# Add an optional display label to a stored bookmark

## Spec

This is a self-contained proposed change in a small, otherwise unspecified Go
bookmark utility. All relevant code context is defined here, rather than asserted
about the review checkout. There are no external dependencies or other workstreams.

`bookmark.go` defines `DecodeBookmark([]byte) (Bookmark, error)` and
`EncodeBookmark(Bookmark) ([]byte, error)`. The stored JSON object has required
`version` (integer 1), `id` (nonempty string), and `target` (object with required
nonempty string `url`). The decoder checks duplicate keys recursively before
decoding known fields and validates their exact types and presence. Both object
types already retain unknown members as `map[string]json.RawMessage`; encoding
merges those untouched values with known fields, including inside `target`.

Add optional `label` to the top-level bookmark, a string used only for display.
An older utility ignoring `label` still opens the same URL and preserves it on
rewrite. Retain version 1: the new field cannot alter execution or interpretation.
Absent labels remain valid. On decode, keep every unrecognized member's raw JSON
value at its original object level; on every rewrite, emit those values unchanged.
Known members take precedence only after removal from the unknown-member map,
so encoding cannot introduce duplicate keys. Object ordering is not significant.

## Done when

Newly written bookmarks can show an optional label without requiring a version
bump. Both current and older codec paths preserve future members, including nested
ones, through edits to known members. Wrong known types, missing required fields,
and duplicate keys remain errors. Existing version admission remains unchanged.

## Plan

- In `bookmark.go`, extend `Bookmark` with optional `Label`, teach both named
  codec functions about it, and keep version 1. Reuse the current recursive
  duplicate check and raw-member merge rather than introducing another codec.
- In `bookmark_test.go`, exercise `DecodeBookmark` and `EncodeBookmark` with a
  round-trip property through old and new codec paths over arbitrary nested
  unknown members and known-field edits; fuzz malformed inputs and assert strict
  duplicate/type/required-member rejection.
- Run `go test ./...` and a bounded decoder fuzz run. Update `README.md` with
  additive label support. Land as one atomic change.

Non-goals: migrations, network access, URL semantics, and general serializer
refactoring; the task changes only display metadata.
