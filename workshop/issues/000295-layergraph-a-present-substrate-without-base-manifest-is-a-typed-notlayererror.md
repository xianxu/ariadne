---
id: 000295
status: working
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: '13255d78a45ef93e847e7ab6ec3b3ec5f5499c79' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-05T15:56:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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

`NotLayerError{Path, Owner}` in `pkg/layergraph`; `declaredGraph` returns it at
the #155 site with today's message (its `Error()` formats exactly the old
string), passed through unwrapped by `Walk` and `DeclaredSubstrates`.
ARCH-FUNERAL/ORDER: no state, nothing durable.

- [ ] Test first: `errors.As` from `Walk` and `DeclaredSubstrates` with the right
  Path/Owner; message byte-identical to the pre-change string (asserted
  literally); a malformed row and an unreadable construct/deps are not
  `NotLayerError`.
- [ ] Implement; consumers' tests (weave, datatype, vocabulary, fleet) + `make test`; atlas; close; land.

## Log

### 2026-10-05

Filed at the operator's request from pair#387 planning, as a follow-up to
#294 (`DeclaredSubstrates`). pair#387 Task 1.3 pins the sha that includes it.

Decision on the Spec's open option: typed error only. `DeclaredSubstrates`
keeps stopping at a present non-layer; the reconciler sets that clone aside and
re-runs, a short loop. The `Layer: false` variant would change #294's
just-pinned contract (and its "Walk = the present entries" invariant) without
a stated need for seeing every other dependency in the same pass.
