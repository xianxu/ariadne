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
rewrite. This change deliberately bumps `version` to 2 and makes the new decoder
reject version 1; previously deployed readers reject version 2. The reason for
the bump is simply that adding any field should require a new version, even when
it has no effect on interpretation or safety. No migration is planned.

## Done when

Newly written bookmarks can show an optional display label. The new reader accepts
only version 2; absent labels remain valid. Existing strict validation and unknown
member preservation at both object levels remain intact.

## Plan

- In `bookmark.go`, extend `Bookmark` with optional `Label`, teach both named
  codec functions about it, write version 2, and replace the version-1 admission
  check with version-2-only admission. Keep the shared recursive duplicate check
  and raw unknown-member retention unchanged. No new seam or storage abstraction.
- In `bookmark_test.go`, exercise `DecodeBookmark` and `EncodeBookmark` using a
  round-trip property over arbitrary retained members and malformed known fields;
  fuzz decoding and assert strict duplicate/type/required-member rejection.
- Run `go test ./...` and a bounded decoder fuzz run. Update `README.md` with the
  label and version-2 admission rule. Land as one atomic change.

Non-goals: migrations, network access, URL semantics, and general serializer
refactoring; the task changes only display metadata and the stated version policy.
