---
id: 000598
status: done
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours: 1.48
card_mirror: '3d78bb4f07e8503f1497c40b47007c302c8a9ec8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T15:03:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.89
---

# Affected planning: declared non-Go path ownership and audited reader inputs

## Problem

Pair#853 cannot safely narrow reader-certificate changes with the current affected declaration. Pair has ~560 certificates under tests/affected-readers/<Go test path>.json. Its broad tests/** full rule forces a suite for each update. Removing the rule leaves certificates unclassified. Group.go only triggers non-Go groups; it never adds Go packages. Certificates currently contain version, disposition, sha256, with no declared read paths.

## Spec

Provide declared path-to-Go-package ownership in affected.Plan: a non-Go changed path can add one or more exact Go package roots to the existing ownership union, dependency closure and group selection. Reuse the existing planner as authority (ARCH-DRY); no Pair-side duplicate planner or per-certificate test-group forest. Declaration must come from the pinned trusted base, as existing affected policy does. Preserve explicit full overrides, multi-owner union, and conservative full fallback for unclassified or missing package owners.

Support Pair's certificate identity tests/affected-readers/<test path>.json mapping to the package holding that test, including added, edited, removed and renamed certificates. The exact generic schema/representation belongs to this issue's approved design; avoid requiring a manually maintained row per test if a bounded canonical mapping can derive it. Strictly reject traversal or ambiguous mappings.

Certificates must also record their audited declared reads (repository-relative paths or explicit patterns), alongside the current disposition/hash. Define where this consumer format is owned: ariadne owns generic planner/config capability; Pair#853 owns its inventory schema migration/audit and consumer wiring unless the approved design requires an upstream seam. Declared reads must have explicit selection semantics: retain the owner's tests and all packages/groups genuinely needed by those inputs, preserving full fallbacks for policy, selector, shared harness or unclassified inputs. Do not treat reverse import closure as a substitute for black-box executable/fixture reads. local_or_literal_fixture, fixed_scanner and non_go_full_fallback remain auditable classifications; missing or invalid read declarations must never authorize skipping coverage.

Pair#853 also narrows cmd/couch/*_test.go changes, which are currently caught by cmd/couch/**. Keep production command/harness changes conservative; the input-reader audit and fixed source preflight remain in force. A test edit should select the owning command package and audited input implications, without requiring the whole repository merely because it is in the main package.

This is a prerequisite for Pair#853, not stacked work. ops:0 assigned ariadne:5 after ariadne#592 lands. Filed unclaimed as requested; no implementation or estimate implied.

## Done when

- Production affected.Plan plus config/graph integration can map a non-Go path to exact Go package owners; tests cover union, dependencies, groups, full precedence, malformed/unknown owners and no-op unmapped input.
- A real Pair certificate case (request_lifetime_test.go.json from Pair#825) qualifies against the new provider and selects its owning package; arbitrary certificates cannot silently select nothing.
- Certificate read semantics and migration boundary are documented, with invalid/older data conservatively handled and opaque future metadata preserved on rewrites.
- Policy/selector/shared harness changes retain full selection; cmd/couch test narrowing has a qualified consumer path without reducing native fixture coverage.
- Compatibility, trusted-policy provenance, deleted/renamed path behavior and bounded mapping cost are tested. No new per-file subprocess or unbounded group multiplication.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.3  impl=0.32
item: smaller-go-module      design=0.1  impl=0.2
item: smaller-go-module      design=0.05 impl=0.16
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0  impl=0.16
design-buffer: 0.15
total: 1.48
```

M1's rows: mapping, sidecar and config parse; Plan integration with the loader and the gitx extraction; the qualification fixture and benchmark; atlas; review. The M2 and M3 rows (originally 3.1h in total) moved to ariadne#616 with their scope.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: [workshop/plans/000598-affected-declared-input-ownership-plan.md](../plans/000598-affected-declared-input-ownership-plan.md).

- [x] Design the generic ownership seam and certificate-read consumer contract (plan above; sent to ops:0 for approval).
- [x] M1 — owners, full exceptions and sidecars (what pair#853 needs)
- M2 (parley-style traceability input): moved to ariadne#616; see Revisions.
- M3 (missed mappings at merge): moved to ariadne#616; see Revisions.

## Log

### 2026-10-24
- 2026-10-24: closed — M1 scope (M2/M3 moved to ariadne#616). Pre-close suite passed on a31433be (sdlc-local/suite, 24m14s, 2479 cmd/sdlc tests + 77 packages). TestPairQualification: pair 93a57a380 real affected.json with pair#853 edits + real pair#825 request_lifetime certificate selects cmd/internal/testcouch/agentfixture and dependents; couch test edits select couch; policy, selector, harness, couch production stay full. Declared, stale, malformed, deleted/renamed sidecar, cancellation (ordinary+boarding) and splitImpact tests pass; fuzz clean; BenchmarkPlanDeclared 33ms/op at 2000 paths x 1200 rules.; review verdict: SHIP
- 2026-10-24: closed M1 — go test ./cmd/sdlc/internal/affected ./cmd/sdlc/internal/gitx pass; go test ./cmd/sdlc -run Affected|SplitImpact|LocalCheckResume|Boarding|GraphPrepare|CarCIGate|PreClose|Attribution passes, plus TestAffectedDeclaredMappingCancelled in ordinary and boarding modes; FuzzParseConfig and FuzzSidecarDerivation 20s each clean; TestPairQualification on pair 93a57a380 affected.json plus the real pair#825 certificate; BenchmarkPlanDeclared 33ms/op at 2000 paths x 1200 rules; review verdict: SHIP

Filed by pair:1 at ops:0 request. Source anchors: cmd/sdlc/internal/affected/plan.go Plan (full overrides, owners plus groups, then Dependents); config.go Group.Go is a trigger only; graph.go Graph.Owners has directory/testdata/embed ownership only. Pair source anchors: .sdlc/affected.json full tests/** and cmd/couch/**; tests/affected-go-readers-test.py load_inventory; scripts/test-affected.py run_selection fixed source preflight. No source code changed.

Design (ariadne:5). The TL relayed the operator's direction: a curated, low-tech mapping as the backbone (parley's traceability spirit, "change the toolbar, run these"), with Go-import selection where it works; plus two separable parts, parley-format input and missed mappings at merge.
- Pair facts (read 10-24, pair 93a57a380): 583 certificates, each {version, disposition, sha256} with no reads; `full` holds tests/** and cmd/couch/**. The Done-when certificate is tests/affected-readers/cmd/internal/testcouch/agentfixture/request_lifetime_test.go.json; it is not under cmd/couch.
- ARCH-DRY: all three declaration sources compile into one Mapping, and Plan and splitImpact share one ownership function.
- Trust: declarations come from main's rev only. A sidecar's path-only alias covers added, renamed and deleted certificates; its `reads` take effect once they land.
- Missed mappings are recorded only when the main baseline passed, so load storms such as today's add no noise.

M1 implemented (ariadne:5):
- `FullRule` with `except`, inline `owners`, `sidecars`, the pure `Mapping` (`Compile`/`ParseSidecar`/`Owned`), and Plan's claim step over it.
- `gitx.TreeFilesSelected` extracted; `loadAffectedMapping` added behind the `affectedDeps.mapping` seam; `splitImpact` loads per endpoint.
- The `traceability` member moved from M1's config row to M2, since M2 waits on the operator's go. M1 parses no field it can't act on (YAGNI).
- Perf (ARCH-CONSTRAINTS): the first benchmark took 710 ms/op for 2,000 paths × 1,200 rules, because every globbed rule was Matched. A literal-prefix prefilter brought it to 33 ms/op on today's loaded machine.
- Qualification: `TestPairQualification` runs pair's real affected.json (pair 93a57a380, copied to testdata) with pair#853's edits, plus pair#825's real request_lifetime certificate. The certificate selects agentfixture and its dependents; couch test edits select couch; policy, selector, harness and couch production code stay full.
- Tests: the affected package with fuzz targets (20s each, clean); gitx; `cmd/sdlc` -run Affected|SplitImpact|LocalCheckResume|Boarding|GraphPrepare|CarCIGate|PreClose|Attribution, all passing.

M1 boundary review, round 1 (REWORK). All five findings fixed in this round:
- BR-1: the plan's table and checkboxes had drifted from the code; a plan Revisions entry now records what was built.
- BR-2: the README's affected paragraph now covers owners, sidecars, `except` and the older-binary fallback.
- BR-3: a cancelled mapping load is now an interruption (ctx.Err check), tested in ordinary and boarding modes.
- BR-4: `claimPath` keeps collecting claims after the first full reason, so `Owned` doesn't depend on rule order; tested.
- BR-5: a shared `hasGlobMeta`.
- Lesson added to workshop/lessons.md.

M1 closed SHIP at round 2 (d595e3b01). Its two advisory Minors are fixed in the same round:
- The README and atlas now name each caller's outcome for an older binary: pre-close and cars require the suite; an explicit affected run and its resume refuse.
- The malformed-sidecar test covers ordinary and boarding modes.

## Revisions

### 2026-10-24: scope widened by the operator's direction (via ops:0)

Reason: the operator wants a curated low-tech mapping as the backbone, plus two separable additions: parley-style traceability input (M2) and missed-mapping recording at merge (M3). The TL approved M1. M2 and M3 wait on the operator's go.
Delta: two Done-when bullets (M2, M3). The Plan's M2/M3 rows and the durable plan's sections for them. M1's scope is unchanged.

### 2026-10-24: M2 and M3 moved to ariadne#616 (ops:0)

Reason: the TL asked for M1 to land alone so pair#853 can proceed. M2 (traceability input) and M3 (the missed-mapping ledger) still wait on the operator's go.
Delta:
- Removed the M2 and M3 Done-when bullets and Plan rows; their acceptance criteria now live in ariadne#616, filed unclaimed.
- #598 closes on M1's scope, the original issue's Done-when.
- The durable plan keeps M2's and M3's design as #616's reference.
- Estimate re-derived to M1's own rows: 3.1h → 1.48h, the same derivation minus the moved rows.

