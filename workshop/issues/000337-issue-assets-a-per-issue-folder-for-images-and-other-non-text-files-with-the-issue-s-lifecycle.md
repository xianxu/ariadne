---
id: 000337
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours: 3.403
card_mirror: 'b3805816f8745a7b181430ff1319546b52ba044e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T13:03:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle

## Problem

Issues are text only. Couch's DOM submission (pair#483) and the terminal note submission (pair#455) produce screenshot crops that an agent may turn into a new ticket or attach to an existing one ("this demonstrates XXX#NNN, update that ticket"). Today there's nowhere in the issue format for those images to live with the ticket and follow it through its lifecycle. The operator, 10-12: extend the ticket format to hold non-text files, following parley.nvim's chat assets.

## Spec

Unshaped; a starting sketch from parley.nvim's `lua/parley/assets.lua` (#231):
- A per-issue sidecar folder keyed by the issue's **id**, not its slug, so a rename never orphans it (e.g. `workshop/issues/assets/<id>/`).
- **The issue body is the index:** the folder holds only bytes the body links (`![](assets/<id>/<file>)`); nothing scans the folder to discover content.
- **Same lifecycle as the issue:** archive to `workshop/history/` moves the folder too; `sdlc migrate` and peer moves carry it; one writer owns create/move/delete (ARCH-DRY).
- **Git implications (the operator, 10-12):** binary files change how the repos use git: repository size and history weight, diffs, encrypted repos (gcrypt brain), publication through `sdlc issue publish` and the issue-tracker branch (#252). Decide in the design: plain git with size limits, LFS, or keeping bytes outside git with a pointer.

### Approved design — 2026-10-13

The original sketch above is retained as history. The approved design is
[the durable plan](../plans/000337-issue-assets-plan.md): one body-indexed bundle
per issue, stored in plain Git with immutable content names, bounded at 1 MiB
per file, 8 MiB and 32 files per issue. The binary resolves operator-owned
allowed/text-only policy before attaching; unset policy derives from host
visibility, with public/unknown becoming text-only. Help prints policy and
source; every refusal gives its clearing action. Source sharing eligibility
remains independently required by the Couch rule from pair#526.

Assets move with the issue through all supported in-repo lifecycle paths,
including publish, archive, restore, abandon/reopen and slot move, using one
shared bundle model and existing exact-byte Git transactions. Ownership and
retry proofs cover asset-only changes. Cross-repo migration remains refused
and is explicitly deferred to #347. Agents never change policy to admit bytes.

## Done when

- An agent can attach an image or other bounded file through an sdlc verb; the body links it and the bundle follows publish, archive, restore, abandonment/reopen and slot move within the repository.
- The attach verb enforces operator-owned allowed/text-only repository policy, derived from host visibility when unset; public or unknown visibility defaults to text-only. Help prints the policy and source. Refusals write nothing.
- Publication and lifecycle retries preserve exact linked bytes and existing ownership protections, including asset-only changes and legacy repositories.
- Cross-repo issue migration remains safely refused; its asset support is tracked by dependent #347.
- The git storage decision is recorded with its reasons.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Calibration marked stale by estimate-source;
values are provisional ship-wall-clock hours.*

Decomposition, before summing: issue shaping uses the low v2 issue-spec design
0.5 and impl 0.2. Bundle/parser and attachment transition/writer are two new Go
concerns, each v2 design 1.0 and impl 0.8. Policy/visibility reuses gh and JSON
rather than a new stack (smaller module design 0.3, impl 0.5). Publication,
archive, restore/abandon, and bounded discovery are four existing-module
extensions, each design 0.3, impl 0.5. Docs use design 0.1, impl 0.2; the single
close boundary uses design 0, impl 0.5. Existing Git transaction/gh libraries
short-circuit integration; a Markdown parser library may halve the bundle
primitive's design, but availability has not yet been verified, so the full
1.0 design basis is retained rather than claiming that discount. No other
library replaces the new attachment transition.

Apply the accepted plan's 0.2 design discount to implementation components
(shaping is already its own primitive), v3.1's 0.4 to all impl values, familiar
Go/Cobra/Git multiplier 1.0, and thorough-plan design buffer 0.15. Rows follow
that decomposition, in the same order. Design subtotal 1.22; impl subtotal 2.00;
1.22 × 1.15 + 2.00 = 3.403 hours, recorded as 3.40.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.50 impl=0.08
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 3.403
```

## Plan

Approved [implementation plan](../plans/000337-issue-assets-plan.md); next gate is change-code.

- [x] Typed asset policy, bounded bundle model, attach verb and dynamic policy help.
- [x] Bundle publication, ownership, recovery and restore.
- [x] Archive, abandon/reopen, legacy lifecycle and slot-move verification.
- [ ] Documentation and acceptance preparation: targeted tests and full suite started below load 40; then enter the binary-owned close review. Shipping remains subject to hub SHIP.

## Log

### 2026-10-12

Filed from pair#483's shaping (pair:2), at the operator's request. Not a dependency of #483: a submission is written to Couch's store first; attaching it to a ticket is step two.

### 2026-10-13 — design exploration

Claimed in ariadne:5 and entered planning. Existing cross-repo migrate explicitly
refuses issues because destination renumbering is deferred; hub narrowed this
issue to the in-repo lifecycle and requested unassigned follow-up #347.
Read pair origin/main's canonical couch skill: only explicit shareable submission
content is eligible, private/unknown yields sanitized text without crops, and
recordings remain local and expendable. The installed couch skill predates that
rule. No implementation or tests run yet. ARCH-DRY and ARCH-PURPOSE motivate one
bundle model across publication, all archive paths, abandon/reopen and restore.

### 2026-10-13 — implementation checkpoint

Change-code accepted the full-flow plan after clean review; estimate-quality noted
integration/recovery and shipping allocations may be optimistic. Implementation
is in progress across attach/policy, bounded snapshot readers, and publication
and archive lifecycle adapters. Pure bundle/transition/policy tests, race tests,
and four bounded fuzz runs passed. Targeted attach/policy/contract tests passed
(26.950s), including fixes first demonstrated RED for missing linked bytes,
symlinked source ancestors, issue ID mismatch, and standalone help recursion.
Lifecycle RED tests exposed missing archive retry proofs, stray archived members,
and recreated active bytes; fixes await combined verification. Publication
receipt tests passed; merged-candidate and recovery coverage is still in progress.
The full suite has not run: machine load remains above the hub's limit of 40.
No completion claim; remaining work includes combined quiet-checkout GREEN,
full suite, binary-owned review, and hub shipping authorization.

### 2026-10-13 — integrated regression evidence

Combined asset, attachment, ownership, slot-move, symlink, recovery and bounded
reader tests passed (145.797s). All internal/tracker and internal/gitx package
tests passed (21.095s and 39.124s). Explicit CLI attach→publish on a tracker
resting branch passed (14.838s); slug-rename preservation passed (2.194s).
Malformed host visibility with duplicate keys first failed RED, then passed
following strict parsing (1.117s). The lifecycle-focused asset/attribute/privacy
run passed (63.288s). The older broad command run timed out at 600s while
scanning live transcript sources under machine load above 300; it also exposed
macOS path canonicalization and repeated-abandon compatibility regressions.
Those failures were addressed; the successor compatibility/privacy/generation
run passed (150.148s). Final card-race/successor/retry checks passed (46.376s), and attribute/legacy-residue/compatibility checks passed (14.353s).
Full make test and the close review are still required. Core implementation
checkpoint: c7c74af75d74224df94f7e3d109f90526d6f653b.

### 2026-10-13 — full-suite findings and corrections

Full suite at 1c5ef2e12ff6afe9b59fd6bd4731ee40b379fa7a started at
21:07:58Z with load 36.96 and finished in 485.8s with five failures.
Two concern this implementation: the new attach verb lacked its recovery
catalog contract, and legacy publication discarded the caller context.
A cancellation regression demonstrated the latter pushing despite cancellation;
context now reaches both the Git transaction and asset policy validation.
Context/publication/cancellation regressions passed (72.115s); attach and
recovery-contract checks passed (63.991s), with all recovery package tests
passing separately (0.077s). The other three failures are existing review
fixtures inheriting PAIR_AGENT=codex despite Claude-specific inputs/assertions;
all pass with PAIR_AGENT=claude (21.947s and 4.558s). Follow-up #349 tracks
pinning those fixtures independently. The next full run will explicitly use
PAIR_AGENT=claude; no test or gate is skipped.

### 2026-10-13 — full-suite acceptance evidence

PAIR_AGENT=claude make test ARGS=--shards 6 passed at committed HEAD
35eccf5153f31923d7f2c371ad1a8a618350d8f4: all six shards (1155 cmd/sdlc
tests) and 47 other packages, zero failures, 431.2s test wall time.
Invocation began at 21:31:57Z with load 33.61, then waited 1082s for both
machine-wide suite slots; completed 21:57:10Z. The checkout remained unchanged
throughout. Full log: /tmp/ariadne337-full-rerun-20261013.log. Follow-up #349
is published and unassigned. Entering the binary-owned close review next;
shipping still requires hub SHIP.

### 2026-10-13 — first close review

The binary-owned close review returned REWORK (BR-1 through BR-4), leaving the
issue working. Its temporary regressions demonstrated stale checkout policy
admitting new bytes after destination tightening, and archive-only Git attributes
converting bytes during reopen. Correcting both classes across their callers,
plus the pure/integration table and live visibility conformance coverage. The
prior full-suite pass remains recorded; acceptance is reopened for these fixes.

### 2026-10-13 — close-review corrections verified

BR-1: RED reproduced stale-local allowance in republish, first publication,
legacy publication and abandonment, plus a real CAS retry race. New admission
intersects local restrictions with each attempt's pinned destination policy;
first-publication recovery uses its prepared candidate's bound base. Added
recovery, same-candidate relaxation, malformed/missing policy and nonregular
or oversized policy regressions. Combined asset/attach/context tests pass
46.623s; all internal/tracker tests pass 74.688s.

BR-2: production RED reproduced archive-only LF-to-CRLF conversion during
reopen and converting attributes during slot moves. Source/destination
preflight now covers reopen, slot moves, resting fast-forwards and publication
cleanup; actual materialized bytes are compared with pinned snapshots.
Conversion refusal and successful unchanged reopen checks pass 10.328s.
The broad existing move/reopen run exposed an inactive hook test fixture,
then the repaired hook-tamper/fast-forward checks passed 5.464s. Custom issue
roots are preserved and IO refusals name recovery actions.

BR-3: the plan distinguishes pure planLandingArchive from integration
confirmLandingArchive, and now uses the implemented core type/function names.
BR-4: TestLiveAssetVisibilityConformance passed through the real production
adapter (0.40s test, 1.171s package), matching independently observed PUBLIC
visibility for xianxu/ariadne. Default non-opt-in skip passed separately.
Atlas documents the explicit read-only invocation and cadence. Full suite and
another close review remain required for this corrected revision.

### 2026-10-13 — corrected full-suite guard finding

The corrected full run at 123491afad386e9176804b80787335beca29c2c2
finished in 490.1s: all behavioral tests passed, with only
TestIssueHintsRenderDecimalIDs failing on the new slot-move recovery hint.
It used a filename-only ID extractor on a branch name, producing an empty
claim argument. The hint now parses the branch convention explicitly and
formats through issue.CLIRef; non-issue branches omit the claim instruction.
The static guard and both runtime branches failed RED before this fix and
passed GREEN in 2.132s. Full-suite acceptance will be rerun once more.

### 2026-10-13 — corrected acceptance suite passed

PAIR_AGENT=claude make test ARGS=--shards 6 passed at
ff3e3811b841091430163f44ad2812066ebf0746: all six shards, 1169 command
tests and 47 other packages, zero failures, 482.1s. Invoked at 22:39:57Z
with load 7.57, acquired both test slots after 483s, completed 22:56:03Z.
Log: /tmp/ariadne337-full-final-20261013.log. Checkout stayed unchanged.
BR-1 through BR-4 corrections and recovery-hint regressions are included;
returning to the binary-owned close review. Hub SHIP is still required.

### 2026-10-13 — second close review

Round 2 returned REWORK. BR-1, BR-3 and BR-4 were accepted with independently
verified regressions/conformance. BR-2 remains open because issue restore can
commit post-checkout-hook-modified bytes. New BR-5 demonstrates GH_HOST selecting
a different repository's visibility after hostname stripping; BR-6 demonstrates
legacy body-only missing-asset links bypassing validation. All three were
reproduced by reviewer overlays. Correction plan appended; acceptance reopened.

- Round 2 targeted evidence: explicit-host visibility regressions passed (3.503s),
  and the real read-only GitHub query passed with conflicting GH_HOST (0.37s).
  Body-index/creation tests passed in cmd/sdlc (21.374s) and tracker (10.734s).
  Tracker reconciliation conversion/content-name regressions reproduced RED,
  then passed with existing reconcile tests (14.473s). Planning/handoff hook,
  conversion and retry tests passed (62.955s); branch/worktree tests passed
  (5.727s), and the context-threading guard passed (0.289s). Full acceptance
  awaits integration of newly landed main paths. The restore/reopen/archive and
  retry regression sweep passed (137.475s); after tightening duplicate cleanup
  to the bounded reader, the reopen pair passed again (68.677s). Reopen now
  keeps every pending merge uncommitted until indexed bundles validate, with
  byte-proven duplicate cleanup before the archive becomes the sole index.

- Main integration verification: stage/clear conversion and hook regressions
  reproduced RED, then passed with existing staging and context guards (26.562s).
  Completed-issue reopen RED (82.426s) became GREEN (25.174s); retained-branch
  reopen RED (30.386s) became GREEN (27.223s). A quiet combined body/visibility/
  reconciliation/recovery/static-guard run passed cmd/sdlc (145.374s) and tracker
  (4.836s). The broad reopen compatibility run timed out at 600.337s while its
  fixture scanned real session transcripts under high load; it is not a pass.
  All code is now frozen for full acceptance, which must start below load 40.

## Revisions

- 2026-10-12 — the operator: issue assets are where submitted web-cast note crops (pair#483) land when notes become an issue
  (text first in the issue body; only the crops that matter as assets; recordings stay local, never S3 yet). Add a
  **privacy gate enforced by the binary**, not agent judgment: a per-repo asset policy (`allowed` | `text-only`), defaulting
  to `text-only` for a public repository (from its host visibility) and to `allowed` for a private one; the attach verb
  refuses images under `text-only` and says why; the policy and its source are printed by `sdlc issue --help`. The
  operator changes a repo's policy; agents never do.

- 2026-10-13T20:06Z — Hub scope direction (ops:0, delivery d0338359-494f-4951-838c-4e18ae676bf6): deliver attach/privacy and the complete in-repository bundle lifecycle; retain migrate's issue-family refusal. Cross-repo issue migration needs the renumbering capability deferred by #179 and is separated into #347, dependent on #337 and left unassigned. This revises the original migrate acceptance criterion explicitly.
- 2026-10-13T20:06Z — Proposed design, pending plan review: plain Git with 1 MiB/file, 8 MiB/issue and 32-file limits; body-indexed immutable content-named assets, shared lifecycle planner, strict operator-owned repository policy, and unknown visibility treated as text-only. Full details and alternatives are in the durable plan; limits remain proposed choices until approved.
- 2026-10-13T20:09Z — Ops:0 approved plan commit 399e38374dd9d769c9d21748e53b15757ee70b16 and the proposed limits, directing change-code. Added the operator's actionable-refusal requirement: each refusal names the action that clears it. The plan records this addition.
- 2026-10-13T20:12Z — Plan-quality round 1 refused with PQ-1 (binary subtree readers), PQ-2 (attachment rollback/restart), PQ-3 (function-level test strategy). Verified existing Files consumers, refined the plan across each class, and recorded prevention rules in lessons. No code changes; rerunning plan-quality.

- 2026-10-13T20:43Z — Byte-preservation validation now refuses converting Git
  attributes with an explicit per-path override; abandon checks new evidence
  before its archive-ref push. A repeat abandonment after reopening replaces the
  previous archive only with proof of successor lineage; same-generation retries
  retain exact-byte proofs. These resolve implementation findings within the
  approved lifecycle/privacy scope.

- 2026-10-13T21:20Z — Clarified the final issue Plan checkbox as acceptance preparation; close and hub-authorized shipping are subsequent workflow gates, so the checkbox does not claim them before they run. Added the attach recovery catalog contract and cancellation threading after full-suite findings, without changing the approved feature scope.

- 2026-10-13T22:06Z — Close review correction plan appended to the durable plan: destination policy authority, all materialization paths, accurate core-concept classification, and live visibility conformance. No feature-scope expansion; acceptance checkbox reopened pending corrected verification.

- 2026-10-13T23:04Z — Round 2 correction plan strengthens complete destination identity, pre-commit materialized-byte verification for every caller, and body-indexed legacy validation. Existing scope retained; repeated class omissions addressed with an explicit caller inventory.
