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

- [x] Write red source-map tests: `// wire:` on a named type in the same file,
  a second file of the package, and an imported package with an alias all
  resolve; a local type of the same name, an unmarked type, an empty or
  unrelated comment, and a generated file do not. Add a counted 8/16-file
  fixture that checks one index visit per file/declaration.
- [x] Run `go test ./pkg/golint -run TestWireIndexCost -count=1`
  and capture the expected red failure.
- [x] Add a pure multi-source parse/index/check path. Keep `Check`'s one-file
  API for existing callers; have it use the same core with a one-file source
  set. Update existing rules to receive the shared index context without
  duplicating suppression logic. The repo test supplies tracked sources once.
- [x] Run the focused tests green and commit the index and rule increment.

## Task 2: prove the decoder rule and annotate the real types

**Files:** `pkg/golint/wire_test.go`, `pkg/golint/rules.go`, the four input files
named above.

- [x] Add red table cases for `json.NewDecoder` →
  `DisallowUnknownFields` → `Decode(&wire)` and `Decode(&Wire{})`, including
  cross-file/imported targets and an aliased `encoding/json` import. Cases
  without the flag, with a local persisted type, with a different decoder,
  with a rebound decoder, or with ambiguous control flow have no finding.
  Verify the finding points to `Decode`, names `pair#594`, and a reasoned
  `//lint:allow wire-strict-decode` works through existing suppression rules.
- [x] Run `go test ./pkg/golint -run TestWireStrictDecode -count=1` red.
- [x] Implement the conservative direct-chain matcher, add it to `Catalog`,
  and add the four `// wire: <producer/reason>` type-doc comments. Keep the
  normal local-file `DisallowUnknownFields` calls untouched.
- [x] Run focused tests green and commit the rule and annotations.

## Task 3: document, verify, close

**Files:** `atlas/workflow/sdlc-binary.md`, `workshop/issues/000440-*.md`.

- [x] Explain when to mark a type, the direct-chain coverage boundary, why
  local/versioned stores stay strict, and how to suppress a proven exception.
  Run `go test ./pkg/golint -count=1`, `make -k tools`, and `git diff --check`.
- [x] Push the issue head through `sdlc pr`, then run
  `sdlc local-check run --name suite` on that head. Record exact head, log and
  result in the issue Log.

After the evidence is complete, close on the reviewed head, answer findings,
land via `sdlc merge --yes` after SHIP, and report the PR and real merge commit
to the hub. The TL owns any project update.

## Revisions

- 2026-10-18, initial plan: chose syntax-only annotated types and a direct
  decoder chain after evaluating the broad flag and type-checking alternatives.
  The hub required annotating existing Ariadne wire types now, with Pair's
  adoption in pair#665.
- 2026-10-18, evidence accounting: the counted fixture checks index visits at
  8 and 16 files; the `CheckSources` parse loop itself has one parse call per
  map entry. Index and matcher were committed as one coherent increment after
  red/green rounds. The close checklist now stops at pre-close evidence so the
  gate need not check its own future outcome.
- 2026-10-18, close review round 1: BR-1/BR-2 exposed stale provenance after
  writes. Keep the conservative contract: an explicitly unmarked interface
  type cannot inherit wire identity from its initializer; direct, multiple,
  and conditional target reassignments clear that identity; direct and
  conditional writes through a decoder pointer clear its strictness. Red
  fixtures for each case precede the correction. BR-3 adds the annotation
  convention to README beside its atlas detail.
- 2026-10-18, close review round 2, decoder-proof re-plan: BR-2 remains open
  because aliases and range assignments can overwrite a decoder outside the
  few assignment forms modeled so far. Use a single-owner proof instead of an
  alias graph: a tracked decoder is usable only while every reference to its
  local name is a direct `Decode` or `DisallowUnknownFields` receiver. Any
  other reference, including an assignment RHS, method argument, closure
  capture, pointer use, or branch use, discards the proof before a later
  decode. A fresh direct `json.NewDecoder` assignment can establish a new
  non-strict proof. Range LHS bindings also invalidate decoder and target
  identity. This deliberately misses some true hazards after an alias but
  cannot report from stale strictness. An alias graph is the tempting
  alternative; it would need to model escapes, helper calls, and joins, which
  defeats this bounded syntax-only rule. Red fixtures will enumerate aliases
  before and after the strict call, helper/closure escapes, and direct,
  conditional, multiple, pointer, and range overwrites.
- 2026-10-18, close review round 3, binding-escape correction: the previous
  model incorrectly let `d = json.NewDecoder(...)` restore proof after `&d`
  escaped; an alias to the variable can overwrite the new object too. Use one
  up-front function scan to permanently disqualify any local name whose
  binding address is taken or whose name occurs inside a closure. The latter
  is deliberately conservative when a closure shadows the name. Such a name
  never enters the decoder proof map, even if the escape appears before its
  first constructor or after a would-be finding. Other uses that copy only a
  decoder object still discard current proof, but a later direct assignment
  may establish a fresh, initially non-strict object when its variable
  binding never escaped. This replaces temporal alias-state tracking with a
  small invariant: eligible binding + direct constructor + direct flag +
  direct decode. Red fixtures cover address-taking before initial
  construction, before and after reassignment, closure capture across
  reassignment, object alias followed by safe fresh assignment, and the
  original `&d` runtime counterexample. Any uncertainty stays unflagged.
- 2026-10-18, hub decision after round 3: simplify further. A decoder local
  qualifies only if its name has exactly one assignment in the entire
  function, that assignment is a direct `json.NewDecoder`, its address is
  never taken, it never occurs in a closure, and every other use is a direct
  receiver of `DisallowUnknownFields` or `Decode`. Any reassignment,
  including a second fresh constructor, permanently disqualifies that name;
  there is no proof restoration. A pre-scan decides eligibility before the
  ordered matcher sees a call. This supersedes round 2's fresh-assignment
  restoration and round 3's eligible-binding restoration. It loses some
  findings but has no temporal alias state. The reviewer's address-alias
  runtime sequence stays a no-finding fixture, alongside construction after
  a prior escape, closure capture, range/pointer writes, ordinary aliases,
  and a clean one-assignment positive fixture.
- 2026-10-18, close review round 4, target-identity matrix: BR-2 is
  addressed. BR-4 found that an import equal to the module path never maps
  to the index's root directory (`"."`). Resolve imports by one rule:
  exact module path → `"."`; module plus slash → the suffix directory;
  every other import → outside this source index. Then require the selected
  directory's declared package name (or its explicit import alias) and a
  marked type in that package. A qualifier resolved to a local binding cannot
  stand for an import. Red source-map fixtures cover same-file, cross-file,
  root and subpackage imports, explicit/default aliases, a lookalike module
  prefix, and shadowed identities in one matrix. This is a rule for the
  whole supported identity space, not a special case at one Decode site.
