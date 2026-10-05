---
id: 000294
status: codecomplete
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: '114dd6837d1dece62bb7fd1f847b5e9e80bb19de' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-05T15:17:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.12
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

One traversal, two views (so `Walk` and `DeclaredSubstrates` cannot diverge):
`declaredGraph(fs, root)` is the existing BFS from `discoverEdges`, extended to
record every declared substrate (owner, physical path, source, present) as well
as the edge map. `Walk` keeps using the edges; `DeclaredSubstrates` returns the
declared list. `substrateTargets` returns rows (path + source, via `ParseRows`
substrate rows) instead of paths. Self-exclusion and the present-but-no-manifest
loud error stay exactly as `Walk` has them, so both views share them.

Bounded reads: `construct/deps` is read through `ReadDeclaration` when the FS
offers it (an optional `DeclarationReader` interface `OSFS` implements), else
`ReadFile` with the `DeclarationLimit` length check — weave's fs and test fakes
keep working unchanged. No fleet `construct/deps` is a symlink today (checked),
so `OSFS` refusing one changes nothing. Package stays stdlib-only.

ARCH-FUNERAL: creates nothing durable (a read-only walk). ARCH-ORDER: no state
between events.

- [x] Tests first (`walk_test.go`, existing in-memory FS fakes): transitive
  present chain; absent substrate reported, not descended; source carried; a
  malformed row errors; an unresolvable parent skipped; dedup across two owners
  (first owner kept); `Walk`'s layer set equals the present, manifest-bearing
  subset of `DeclaredSubstrates`; an oversized declaration errors.
- [x] Refactor `discoverEdges` into `declaredGraph`; `substrateTargets` returns
  rows; export `DeclaredSubstrate` / `DeclaredSubstrates`.
- [x] `DeclarationReader` + `OSFS.ReadDeclaration`; bounded fallback.
- [x] `go test ./pkg/... ./cmd/weave/... ./cmd/datatype/... ./cmd/vocabulary/...` + `make test`; atlas; close; land.

## Log

### 2026-10-05
- 2026-10-05: closed — Re-close after fixing the three close-review Minors: TestUnreadableDeclarationIsLoudAndRootsAreAbsolute (construct/deps that exists but cannot be read is an error for Walk and DeclaredSubstrates under OSFS; relative root yields absolute Path/Owner); fallback bound documented. Earlier evidence: TestDeclaredSubstrates, TestDeclaredSubstratesErrors, TestWalkIsThePresentSubsetOfDeclaredSubstrates, TestDeclarationReadsAreBounded; stdlib-only; pkg/weave/datatype/vocabulary/fleet green; make test green.; review verdict: SHIP
- 2026-10-05: closed — DeclaredSubstrates exported over the one traversal Walk uses (declaredGraph). Tests: TestDeclaredSubstrates (transitive present chain; absent reported with Present=false and source, not descended; unresolvable parent skipped; dedup across two owners, first kept; root never its own substrate), TestDeclaredSubstratesErrors (malformed row; present without manifest), TestWalkIsThePresentSubsetOfDeclaredSubstrates, TestDeclarationReadsAreBounded. go list -deps: stdlib only. pkg, weave, datatype, vocabulary, fleet tests green; make test green (processgroup sandbox-only).; review verdict: SHIP
- 2026-10-05: flow upgraded quick → full — 105 added lines in code files (limit 100)

Filed at the operator's request from pair#387 planning (its Task 1.3 imports
this; Tasks 1.1–1.2 do not wait). Related: #293 (fleet substrate sources).

Implemented. `declaredGraph` is the one BFS: `discoverEdges` (Walk) and
`DeclaredSubstrates` are its two views. `substrateTargets` returns rows with
`Source` via `ParseRows`. `DeclarationReader` (OSFS implements it with
`ReadDeclaration`); other FS fall back to `ReadFile` + `DeclarationLimit`.
Tests: `TestDeclaredSubstrates` (transitive chain, absent reported with source,
unresolvable parent skipped, dedup first-owner, root never its own substrate),
`TestDeclaredSubstratesErrors` (malformed row, present without manifest),
`TestWalkIsThePresentSubsetOfDeclaredSubstrates`, `TestDeclarationReadsAreBounded`
(oversized via OSFS and a plain FS; symlinked declaration refused by OSFS).
`go list -deps ./pkg/layergraph`: stdlib only. weave, datatype, vocabulary,
fleet tests green.

Close review Minors fixed in round: with a safe reader (OSFS) a construct/deps
that exists but cannot be read is now an error for Walk and DeclaredSubstrates
(previously "no deps" — intended: a silently dropped layer chain was #155's
failure mode), tested; the ReadFile fallback's bound is documented as a parse
bound; a relative root is made absolute, tested.
