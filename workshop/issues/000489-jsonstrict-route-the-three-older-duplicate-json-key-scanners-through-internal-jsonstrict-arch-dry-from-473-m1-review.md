---
id: 000489
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'eea7579ed0b61cf1969d840dd54d2bc0f0bf88ec' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T02:02:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.18
---

# jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review)

## Problem

## Spec

Four copies of a "reject duplicate JSON object keys" scanner exist:
- `cmd/sdlc/internal/observe/json.go` `rejectDuplicateKeys`
- `cmd/sdlc/internal/fleet/types.go` `rejectDuplicateJSONKeys`/`scanJSONNoDuplicateKeys`
- `cmd/sdlc/teststats_store.go` `validateJSONNoDuplicates`
- the canonical `cmd/sdlc/internal/jsonstrict.RejectDuplicateKeys` (from #473 M1)

Move the three older sites onto `jsonstrict`, keeping each caller's error wrapping (ARCH-DRY).

## Done when

- `grep` finds one duplicate-key scanner in `cmd/sdlc`.
- Each former caller's existing duplicate-key test still passes.

## Plan

Quick flow. Plan posted to the hub, proceeding unless they object.

- [x] `jsonstrict`: `Scan(dec, maxDepth)` (one value, the caller's decoder, an optional nesting bound) under
  `RejectDuplicateKeys`, which now uses `UseNumber` (test-stats needs `1e400` to stay opaque).
- [x] Fold every generic recursive duplicate-key scanner into it, keeping each caller's error prefix: observe
  (`rejectDuplicateKeys`), fleet (`rejectDuplicateJSONKeys` / `scanJSONNoDuplicateKeys`), teststats_store
  (`validateJSONNoDuplicates`), tracker receipt (`uniqueReceiptKeys`, depth 12), reviewusage (`uniqueKeys`, depth 16)
  and actualcorrection (`strictCorrectionJSON`'s walk).
- [x] The issueassets policy and the estimate correction keep their exact-name allowlists; their own `seen` maps are
  replaced by `jsonstrict.RejectDuplicateKeys`.
- [x] `pkg/golint` rule `json-recursive-scanner`: a self-recursive Token/More walk with a string-keyed set, outside
  jsonstrict. Listed in the atlas lint catalog.

## Log

### 2026-10-20
- 2026-10-20: closed — jsonstrict Scan/RejectDuplicateKeys (UseNumber, depth bound, case fold) tests + FuzzFoldKey; 6 recursive scanners and 2 allowlist seen-maps folded; affected packages green (jsonstrict, estimate, issueassets, observe, fleet, tracker, reviewusage, affected) and cmd/sdlc targeted TestStats|SuiteAttempt|Correction|Observ|Receipt|Fleet ok; golint json-recursive-scanner passes the tree and flags 5/6 deleted scanners on main copies; recorded suite queued on the pushed head.; review verdict: SHIP
- 2026-10-20: flow upgraded quick → full — 117 added lines in code files (limit 100)
- The sweep found more than the three named scanners. Three more were generic recursive scanners (receipt,
  reviewusage, actualcorrection) and two were top-level allowlists with their own `seen` maps (issueassets,
  estimate); all five are folded. Left on purpose: `landingline.decodeStrict`, `qodertranscript.object`,
  `reviewprepare.Parse`, and teststats `decodeSuiteAttempt` / `validateSuiteFactObjects`. Each is a schema decoder
  whose map also tracks required or canonical fields, or returns raw unknown members (qodertranscript must keep user
  field names out of diagnostics). None of them recurses.
- Behavior: every folded caller is now case-fold strict (`command` and `COMMAND` are duplicates, as `encoding/json`
  treats them). That is only ever stricter.
- Lint mutation: run against main's copies, `json-recursive-scanner` flags 5 of the 6 deleted scanners.
  observe's was an iterative stack walk, which a syntactic rule without recursion can't tell apart from an ordinary
  token loop; the rule doesn't cover that shape.

- Close review: SHIP with 2 Minors, fixed in the same round:
  - `TestIdentityStrictProtocol` now rejects a duplicate or case-variant session key. A mutation that drops the scan
    fails it.
  - The lint rule also sees methods and calls through a selector (`s.f()`), with a table row and a mutation check.
    The rule still can't see an iterative stack walk, and its comment says so.

## Revisions

- 2026-10-20: scope grew from "the three older scanners" to the whole class found by sweep (ARCH-PURPOSE): four more
  sites, and a lint rule. Done-when's "one duplicate-key scanner" now means one *recursive* scanner, guarded by
  `json-recursive-scanner`; schema decoders that walk one object's members are a different job.

