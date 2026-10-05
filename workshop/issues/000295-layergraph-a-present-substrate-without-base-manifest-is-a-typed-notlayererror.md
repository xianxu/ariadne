---
id: 000295
status: open
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: '45fe449fde04ac1c22cd22d4121a296f8d02c1da' # card fields mirrored from issue-cards; edit via sdlc
---

# layergraph: a present substrate without base.manifest is a typed NotLayerError

## Problem

`declaredGraph` (`pkg/layergraph/walk.go`, shared by `Walk` and #294's
`DeclaredSubstrates`) fails when a declared substrate is present on disk but has
no `construct/base.manifest`. That's the #155 rule, and it should stay loud.

The failure is an untyped `fmt.Errorf`, so a caller can't tell "this one
dependency directory is not a layer" from any other failure (a malformed row, an
unreadable `construct/deps`) without matching message text.

pair#387 (the Couch slot reconciler) needs exactly that distinction. A dependency
clone that was interrupted mid-clone, or gutted, is present without a
base.manifest. The reconciler should set that directory aside (moved whole into
saved work) and re-clone it. With an untyped error it can only report "unknown"
and hand off to a human.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- Export `type NotLayerError struct { Path, Owner string }` (Path: the
  physical substrate dir; Owner: the layer root whose `construct/deps` declares
  it). Its `Error()` returns today's message unchanged, so weave's output and
  any test or user who reads it see no change.
- `declaredGraph` returns `&NotLayerError{…}` at that site; `Walk` and
  `DeclaredSubstrates` pass it through unwrapped (or wrapped with `%w`), so
  `errors.As(err, &nle)` works for every caller.
- No other error changes type. The package stays stdlib-only.
- Open for the implementer to weigh (not required): `DeclaredSubstrates` could
  instead report such an entry (e.g. `Present: true, Layer: false`), stop
  descending at it, and keep walking the rest, while `Walk` stays loud. That
  would let a consumer see every other dependency too. If chosen, record why in
  the Log; the typed error is still exported for `Walk`'s callers.

## Done when

- `NotLayerError` is exported; a test asserts `errors.As` succeeds with the
  right `Path`/`Owner` from both `Walk` and `DeclaredSubstrates`, and that the
  message is byte-identical to before.
- A malformed row and an unreadable `construct/deps` are *not*
  `NotLayerError` (tested).
- Landed on `main` so pair can pin it.

## Plan

- [ ]

## Log

### 2026-10-05

Filed at the operator's request from pair#387 planning, as a follow-up to
#294 (`DeclaredSubstrates`). pair#387 Task 1.3 pins the sha that includes it.
