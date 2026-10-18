---
collaboration-mode: discussion
---

# Annotated wire-message decode lint Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to
> choose an execution approach. Check off steps as they pass.

**Goal:** Prevent a strict JSON decoder from rejecting newer fields of a
marked, independently evolving wire message, while preserving strict decoding
for local persisted contracts (pair#594, ariadne#440).

**Architecture:** Keep `pkg/golint` syntax-only. Parse tracked Go sources once,
index `// wire:` type declarations by package/import path, and pass that pure
index to the catalog's matchers. A bounded direct decoder-chain matcher reports
the `Decode` call only when both the strict decoder and annotated target are
established in the same lexical flow. Reuse the existing finding and
`//lint:allow` mechanisms.

**Tech Stack:** Go stdlib `go/parser`, `go/ast`, `go/token`; the existing
`pkg/golint` catalog and `go test` runner. No new dependency or runtime data.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| `WireTypeIndex` | `pkg/golint/wire.go` | new |
| `wireStrictDecode` matcher | `pkg/golint/wire.go` | new |
| `Rule.Match` context | `pkg/golint/check.go`, `rules.go` | modified |

`WireTypeIndex` owns only declared type identity: package directory/import path
plus type name. It is built once from the source set, not once per candidate
decoder (`ARCH-DRY`, `ARCH-CONSTRAINTS`). A comment belongs to the type when it
is the `TypeSpec` doc or the immediately enclosing single-type declaration doc;
the prefix is exactly `wire:` and must include a reason. An imported selector
resolves through that file's import alias to the index; unrelated names do not
match. Type aliases need their own annotation. The matcher owns local decoder
and target binding identity and direct statement order. It does not infer
helper semantics or control-flow joins, avoiding false positives from guesses.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| tracked-source collector | `pkg/golint/repo_test.go` | modified | Git's tracked-file list and filesystem reads |
| atlas convention | `atlas/workflow/sdlc-binary.md` | modified | contributor guidance |

The collector remains the thin IO seam. It gathers the files, passes bytes to
one pure `CheckSources`/index pass, and prints findings with their paths. Tests
of the matcher feed source maps directly, with no command or filesystem mocks.
This issue creates no durable runtime artifact; source annotations and atlas
text live with the code and end by normal repository history.

## Design decisions and limits

Two simpler approaches fail the purpose: banning every
`DisallowUnknownFields` breaks deliberately strict local-file validators;
looking for a wire type name in the same function guesses at a destination.
`go/types` or `go/packages` would follow arbitrary helper flows, but adds a
dependency and expensive module-wide type-checking to a catalog that currently
scans ~1,140 files under a second. The chosen rule covers the direct standard
decoder chain. Unsupported or ambiguous flows stay unflagged and documented;
the rule never reports a destination it cannot identify.

The initial annotations are `cmd/sdlc/internal/judge/stream.go`'s
`streamEvent`, `cmd/sdlc/teststats_parse.go`'s `goTestJSONEvent`,
`cmd/sdlc/internal/usage/read.go`'s `record`, and
`cmd/sdlc/internal/usage/codex.go`'s `codexUsage`. All are read tolerantly from
independently evolving producers. Persisted `tracker.receiptWire`,
`fleet.fleetPolicyWire`, and `observe.Observation` remain unmarked, since
unknown fields are schema violations for those version-pinned stores.

For F tracked files, T type declarations, and D decoder calls, index work is
one file parse plus one type-declaration visit per item; matching is one
bounded walk over each file, O(source nodes + D), with map lookups for types.
A counted fixture at N=8 and 2N=16 files will assert parse/visit counts grow
at most 2x, not use elapsed time. No artifact retention or cleanup is added.

## Task 1: prove the annotation and index contract

**Files:** `pkg/golint/wire_test.go`, `pkg/golint/wire.go`, `pkg/golint/check.go`,
`pkg/golint/repo_test.go`.

- [ ] Write red source-map tests: `// wire:` on a named type in the same file,
  a second file of the package, and an imported package with an alias all
  resolve; a local type of the same name, an unmarked type, an empty or
  unrelated comment, and a generated file do not. Add a counted 8/16-file
  fixture that checks one parse and one index visit per file/declaration.
- [ ] Run `go test ./pkg/golint -run 'TestWire(TypeIndex|IndexCost)' -count=1`
  and capture the expected red failure.
- [ ] Add a pure multi-source parse/index/check path. Keep `Check`'s one-file
  API for existing callers; have it use the same core with a one-file source
  set. Update existing rules to receive the shared index context without
  duplicating suppression logic. The repo test supplies tracked sources once.
- [ ] Run the focused tests green and commit the index increment.

## Task 2: prove the decoder rule and annotate the real types

**Files:** `pkg/golint/wire_test.go`, `pkg/golint/rules.go`, the four input files
named above.

- [ ] Add red table cases for `json.NewDecoder` →
  `DisallowUnknownFields` → `Decode(&wire)` and `Decode(&Wire{})`, including
  cross-file/imported targets and an aliased `encoding/json` import. Cases
  without the flag, with a local persisted type, with a different decoder,
  with a rebound decoder, or with ambiguous control flow have no finding.
  Verify the finding points to `Decode`, names `pair#594`, and a reasoned
  `//lint:allow wire-strict-decode` works through existing suppression rules.
- [ ] Run `go test ./pkg/golint -run TestWireStrictDecode -count=1` red.
- [ ] Implement the conservative direct-chain matcher, add it to `Catalog`,
  and add the four `// wire: <producer/reason>` type-doc comments. Keep the
  normal local-file `DisallowUnknownFields` calls untouched.
- [ ] Run focused tests green and commit the rule and annotations.

## Task 3: document, verify, close

**Files:** `atlas/workflow/sdlc-binary.md`, `workshop/issues/000440-*.md`.

- [ ] Explain when to mark a type, the direct-chain coverage boundary, why
  local/versioned stores stay strict, and how to suppress a proven exception.
  Run `go test ./pkg/golint -count=1`, `make -k`, and `git diff --check`.
- [ ] Push the issue head through `sdlc pr`, then run
  `sdlc local-check run --name suite` on that head. Record exact head, log and
  result in the issue Log. Close on the reviewed head and answer findings.
- [ ] Land via `sdlc merge --yes` after SHIP, report the PR and real merge
  commit to the hub. The TL owns any project update.

## Revisions

- 2026-10-18, initial plan: chose syntax-only annotated types and a direct
  decoder chain after evaluating the broad flag and type-checking alternatives.
  The hub required annotating existing Ariadne wire types now, with Pair's
  adoption in pair#665.
