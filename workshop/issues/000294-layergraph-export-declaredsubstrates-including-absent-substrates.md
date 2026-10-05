---
id: 000294
status: open
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: 'ccb79f725c7dc69c40a046249c8f5c54e330c5d4' # card fields mirrored from issue-cards; edit via sdlc
---

# layergraph: export DeclaredSubstrates, including absent substrates

## Problem

pair#387 (the Couch slot reconciler) has to know whether every dependency a slot
declares is actually checked out, so it can re-run setup when a clone is missing.
The operator decided pair imports `pkg/layergraph` rather than copying weave's
`construct/deps` rules.

The exported surface doesn't answer the question:
- `Walk` (`walk.go`) present-skips an absent substrate, which is exactly the case
  the reconciler needs to see.
- `substrateTargets` (the row-to-physical-path resolution, ported from
  `lib-deps.sh`) is unexported.
- `ParseRows`/`ParseDeps` give rows, not resolved paths.

A consumer would therefore have to re-implement resolution and the transitive
walk, which is the duplication the import was meant to avoid.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

Export one function built on the same internals `Walk` uses, so the two cannot
diverge:

```go
type DeclaredSubstrate struct {
    Path    string // physical, absolute (substrateTargets' resolution)
    Owner   string // the layer root whose construct/deps declares it
    Source  string // the row's source column, "" when absent
    Present bool   // exists on disk
}

// DeclaredSubstrates walks construct/deps from root over present layers
// (transitively, as Walk does) and returns every declared substrate,
// present or absent, deduplicated by Path, in discovery order.
func DeclaredSubstrates(fs FS, root string) ([]DeclaredSubstrate, error)
```

- Resolution, physical canonicalization and the unresolvable-parent skip are
  `substrateTargets`', shared rather than copied. `substrateTargets` returns
  rows that keep `Source`.
- Unlike `Walk`, an absent target is reported with `Present: false` instead of
  being skipped. Only present targets are descended into.
- A present target without a base.manifest keeps `Walk`'s loud error.
- Reads stay bounded (`ReadDeclaration`'s limit) and the package stays
  stdlib-only, so pair can import it without new transitive dependencies.

## Done when

- `DeclaredSubstrates` is exported with tests for: transitive present chain;
  an absent substrate (reported, not descended); a source column carried
  through; a malformed row (error, as `ParseRows`); an unresolvable parent
  (skipped, as `Walk`); dedup across two owners declaring the same substrate.
- `Walk` and `DeclaredSubstrates` share `substrateTargets`; a test pins that
  `Walk`'s layer set equals the present, manifest-bearing subset of
  `DeclaredSubstrates`.
- Landed on `main` so pair can pin it (`go get github.com/xianxu/ariadne@<sha>`).

## Plan

- [ ]

## Log

### 2026-10-05

Filed at the operator's request from pair#387 planning (its Task 1.3 imports
this; Tasks 1.1–1.2 do not wait). Related: #293 (fleet substrate sources).
