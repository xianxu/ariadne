---
id: 000546
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 2.11
collaboration-mode: collaboration
card_mirror: '947aab14d0e6aa1f5f4204726b5fc68e4f16425d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T03:02:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.71
---

# issue migrate inventory omits archive-only branches and other-worktree history evidence

## Problem

Pre-existing issue migrate inventory omits archived issue evidence outside pinned main. Discovered by #543 close review round 2 (BR-2, authoritative-inventory-completeness); TL explicitly moved importer completeness out of new-repo init's scope.

Reproductions against real Git/bare origin: (1) clone B creates and pushes a branch containing workshop/history/issues/000001-old.md while main has no issues; clone A migrationInventory scans branch edits only under workshop/issues and plans an empty tracker. (2) another linked live worktree contains uncommitted workshop/history/issues/000001-old.md; dirtyIssuePaths scans only workshop/issues, again allowing an empty plan. The same roots/scope rule must examine flat legacy history and detached/inherited branch/worktree evidence, rather than patching only these two sites.

## Spec

Main remains the only source of imported cards. Use the direct-child issue filename rule across the three roots from `issue.IDDirs`: active `workshop/issues`, flat legacy `workshop/history`, and canonical `workshop/history/issues`. Plans, nested side bundles and assets are excluded. Observation is complete before any consumer policy; init retains its independent presence-only policy.

| Scope | Active root | Both archive roots |
|---|---|---|
| Pinned publication main | Import and mirror, using existing card/close proof rules | Import cards using existing duplicate selection; never rewrite archived details |
| Every pinned local/current publication-remote tip and live worktree HEAD, including detached | Refuse a present ID absent from main's imported population, even if inherited. For related tips retain existing merge-base changed-file policy for card fields, active rename/deletion and close anchors. For unrelated tips require exact published active records; otherwise refuse rather than infer ancestry. | A present archive must match a published archive's filename ID and blob bytes (path may differ); otherwise refuse, naming its carrier and path. Main remains authoritative for absences: missing archives on stale tips are not invented deletions. |
| Every live worktree index and filesystem | Refuse uncommitted issue edits, additions and deletions | Same rule, including ignored/untracked direct-child documents |
| Unknown, unreadable, malformed or non-directory root observation | Refuse before tracker/main publication | Same rule |

Every branch name is observed, including main, resting and reserved names. Deduplicate immutable reads by OID while retaining carrier names. Verified partial tracker bootstrap and completed migration recovery remain the existing exact-root/marker proofs; no new durable state or publication phase. Bare/prunable worktrees retain their explicit skip policy. The local worktree HEAD is pinned, including detached and unrelated histories. Root components and candidate file kinds cannot silently hide evidence.

The archive rule is intentionally conservative: even an old branch's differing archive copy refuses until the operator lands, aligns or drops that evidence under the legacy workflow. It does not import branch-only cards or decide which differing historical record should win. Related inherited active copies of already represented IDs keep their existing semantics; they are not fabricated edits. Archive observations never enter the active `Branches` rename/reconcile input.

For worktrees, retain Git's tracked-content comparison (including attribute normalization), use NUL-delimited worktree/index/status parsing, and inspect direct filesystem entries against the index so ignored issue documents cannot disappear. Conflicted indexes and issue entries whose assume-unchanged/skip-worktree flags prevent trustworthy observation refuse with the named remedy. No raw-byte filesystem-versus-index comparison that mistakes CRLF normalization for lost work.

The operator's existing freeze of all legacy SDLC writers remains a precondition during dry run/apply. A local lock cannot freeze other clones or editors; this work does not claim a distributed absence guarantee. Apply recollects the complete inventory before checking the reviewed digest. Existing tracker-root CAS, main cutover CAS, exact retry and post-cutover reconcile remain unchanged.

## Done when

- Real-Git regressions reproduce the remote archive-only branch and other-worktree dirty archive cases and prevent silent empty migration.
- A reviewed root-by-scope inventory rule states what is imported, what refuses, and how unknown observations fail closed.
- Existing active branch rename/archive and provable legacy-close migration behavior remains verified.

## Plan

Durable plan: [complete migration inventory](../plans/000546-migration-inventory-plan.md). Full flow, one atomic close boundary.

- [x] Implement typed evidence and pure matching while preserving related active reconciliation and close proofs.
- [x] Reuse the complete observer across init/migrate; cover all roots, tips and live worktree evidence with main-only imports.
- [x] Establish counted growth, private production CLI/native Git conformance, unchanged migration recovery and current documentation. The recorded suite, close and landing remain binary-owned acceptance gates.

## Log

Filed unclaimed per TL direction, Couch 5119a959-0128-4420-89ee-9a91fb0b08d9. Review reproduction was against #543 pinned head 96e8ff1cfaf539a80c5c174ce4ba7bbf83992bae; the omission predates #543 in issuemigrate.go (branch/worktree scopes use d.Home only). No importer expansion authorized in #543.

### 2026-10-22 — planning inventory completeness
- 2026-10-22: closed — Bounded post-SHIP correction: deterministic asset-conversion fixture RED reproduced (1.659s); common published archive reference is now established before fixture fork, retaining main-only asset/attribute arrival. Both asset-reconcile refusal/no-mutation oracles passed (9.962s); combined asset/inventory/ref-category tests passed (124.191s); Go lint/recovery passed. Removed unused inventory helpers and ported tests to the live complete observer, addressing Minor BR-1 without changing production policy. New exact-head recorded suite is queued/running and must pass before landing; previous failed attempt remains failure. TL owns project updates; private CLI/native conformance gate approved with no stage. (fixed-then-ship); review verdict: SHIP
- 2026-10-22: closed — Native completeness regressions passed after behavioral RED; root growth/uncertainty/unusual-filename group and corrected space-path fixture passed (53.381s); pure issue/tracker packages and Go lint/recovery contracts passed. Private production CLI on implementation 956c1869c proved both refusals preserve refs/HEAD/index, main-only archive imports, idempotent apply and next-ID/claim; final-head proof refresh is running. Recorded suite on pushed 345fdc368 is queued/running, not claimed passed, and must pass before landing. TL approved private CLI/native Git acceptance with no stage and owns project updates; no-project preserves that ownership.; review verdict: SHIP

Claimed and started on a fresh main-based issue branch after #543's confirmed landing/cleanup. Read the current importer and real-Git regression contracts. Main's archive scan omits the flat root; branch deltas omit archive and inherited/unrelated presence; worktree dirtiness omits both archive roots and detached committed evidence. A bounded read-only survey independently confirmed these gaps and the active rename trap.

Alternatives: (A, recommended) complete observation, main-only imports and conservative refusal of unrepresented non-main evidence; (B) import/merge records from every carrier, rejected because it needs new authority and conflict-selection rules; (C) patch the two reproduced scan sites, rejected because inherited/detached/name-excluded evidence remains invisible. ARCH-PURPOSE requires the root × scope class; ARCH-DRY keeps one path/root observer and separate init/migrate policy; ARCH-PURE keeps archive/presence decisions in the existing pure migration planner. No implementation begun; spec sent to TL for approval before the detailed plan.

### 2026-10-22 — spec approved; durable plan frozen for review

TL approved the complete root-by-scope spec and private production CLI/native bare-Git conformance with no stage (Couch 0fc506bf-9240-416e-b647-898eb9f309a1). A wording check explicitly preserved related-active merge-base/card-field reconciliation and allowed prose differences; byte matching applies to archive evidence and unrelated active snapshots only (7a42fef5-cd44-4876-b333-aff944ce8657). Detailed plan now names pure entities, shared observation extraction, finite refusal/recovery table, the exact preserved authority boundary, counted growth and adversarial guards. No implementation or recorded tests started. Await TL plan approval before change-code.

### 2026-10-22 — plan-quality round 1

TL approved the durable plan (Couch c377850d-b344-4086-9dfc-3587e136bee2). First change-code inferred quick from design size; reran with the approved full-flow pin before implementation. Full plan-quality review accepted the architecture/scope and raised PQ-1 (Important, function-level-test-strategy): name production functions and compress repeated test inventories. Revised the strategy table/tasks accordingly without a design change; rerunning the same gate. No implementation begun.

### 2026-10-22 — implementation checkpoint: pure and observation adapters

Full-flow change-code admitted the plan after PQ-1 correction; estimate-quality INFO accepted the derivation while noting optimistic verification budgets. Four module units include their adversarial tests, the native discovery unit covers production CLI acceptance, and the review unit carries close/landing coordination; recorded suite/integration duration remains a calibration risk rather than excluded active time.

Pure issue and tracker packages passed in full (0.371s / 65.192s), including unchanged migration/close contracts. Gitx bounded-reader and NUL stage/flag parser targeted tests passed after RED. Native completeness regressions first failed behaviorally on the old adapter (48.413s), including both reviewer reproductions, inherited/unrelated/detached omissions, name exemptions, flat-main import, ignored/index-hidden dirtiness and changed-evidence apply. After shared observation, complete main imports and worktree adapter wiring, the same new native regression group passed (34.009s). No test failure classified as infrastructure or flake. Next: old migration/init regression sweep, growth/error guards, private CLI conformance, atlas, recorded suite and binary close review.

### 2026-10-22 — acceptance and compatibility sweep

Private production CLI conformance passed on committed implementation 956c1869c58d69fc097621d9a3a006893ab31e1b. Artifact `/Volumes/TestScratch/ariadne-546-cli.odlwjZ/result.json` links exact commands/results, executable harness and preservation snapshots. Both original reproductions refused without changing origin refs, caller/worktree HEAD, raw index, status/diffs or issue bytes. Legacy Git publication resolved evidence; apply imported both archive roots, repeated apply was idempotent, and normal issue new allocated #43 then claim succeeded.

The broader migration/init native sweep (265.258s) found only the existing path-with-spaces fixture used a noncanonical scratch filename; updated it to `000099-scratch notes.md` so its original byte-exact path oracle remains under the approved filename grammar. A self-audit caught a separate false-absence bug in the new classifier: literal POSIX backslashes are accepted by ParseFilename, so they must stay visible. Behavioral RED then GREEN corrected the classifier and metadata validator without a design change; pure related/unrelated/role matching cases now defend it. Added native unusual-filename, resource-growth and uncertainty tests before final acceptance. No recorded suite has run yet.

### 2026-10-22 — final targeted acceptance

Additional native growth/uncertainty/filename tests and the corrected legacy space fixture passed together (53.381s). Git process / blob-batch counts: documents 8/16 = 44/44 and 4/4; distinct tips 8/16 = 147/267 and 3/3; duplicate refs 8/16 = 42/42 and 3/3; live worktrees 0/8/16 = 27/43/59 and 3/3/3; nested assets 8/16 = 27/27 and 3/3. All three roots were exercised for unsupported root/document kinds, unreadability and unusual filename bytes; cancellation refused. Go lint/recovery tests passed (0.754s / 0.135s), diff check clean. Atlas and README now describe the complete observation and conservative refusal rules.

Implementation Plan checkboxes now describe demonstrated implementation/targeted acceptance outcomes; recorded suite, mandatory binary review and landing are still outstanding gates, not falsely completed checklist actions. Next exact step: push final committed implementation, queue its recorded suite, refresh private CLI proof on that head, then binary close/merge. No stage per TL.

### 2026-10-22 — close SHIP; final-head private proof

Close review round 1 returned SHIP with no Critical/Important findings and independently reran the full targeted migration/init/inventory group (256.929s), existing pure migration cases, lint and recovery tests successfully. One Minor advisory (obsolete-production-paths) records unused compatibility helpers; retained in the ledger rather than changing accepted code and abandoning the queued exact-head suite. The card is codecomplete, bound to close evidence; landing still requires the recorded suite.

Final required private production CLI conformance passed at 345fdc36883e03cca3b98d68586010ce77926e03: `/Volumes/TestScratch/ariadne-546-cli-final.oGwUNw/result.json` with exact commands, executable harness and preservation snapshots. All 13 required assertions passed. A separate supplemental main archive with a literal backslash was observed correctly but its derived card path was refused by the existing bootstrap writer validator; main was unchanged and tracker absent. This is not import support or a passing supplementary case. Evidence: `/Volumes/TestScratch/ariadne-546-cli-refresh.OULaB4/supplemental-ref-preservation.json`; resolve unsupported writer filenames under the legacy workflow before retrying.

Recorded suite is queued on the accepted implementation head, run lease-2685531763, attempt 1a507cf3e73d3dd4381dd4ada1284890. Never infer PASS from targeted review/conformance; preserve that runner/FIFO position through PR and landing preparation. No private Go cache was created.

### 2026-10-22 — recorded suite fixture failure and bounded correction

Recorded attempt `1a507cf3e73d3dd4381dd4ada1284890` on 345fdc368 failed after 11m44s: 2,159 cmd/sdlc tests and 76 other packages ran; only `TestAssetMigrateReconcileRefusesConversionBeforeMerge` failed (plus its package failure event). Saved log hash prefix `56a45adc9312`. This is a deterministic fixture incompatibility, not infrastructure or a flake: targeted RED reproduced it in 1.659s. The fixture forked an older archive body before main added its evidence reference, so the new approved archive-bytes rule correctly refused before the intended asset-conversion guard.

Corrected setup publishes the common archive reference before the fork, while the asset bytes/attributes still arrive only on main for the conversion case. Both asset-reconcile tests passed (9.962s), preserving their exact conversion/content-name refusal and no-mutation oracles. No production policy was relaxed. Since the failed suite requires a fresh head anyway, also addressed the review's Minor obsolete-production-paths class: removed the unused dirty wrapper and filtered branch helper, retained one live dirty observer, and moved old helper tests to the actual complete observer. The combined asset/inventory/ref-category target group passed (124.191s).

Next: fresh recorded suite on this committed correction, plus narrow fixed-then-ship review. The earlier full-suite failure remains recorded as failure; it is not resumed, republished as pass or registered as a flake.


### 2026-10-22 — corrected-head conformance and re-close SHIP

Final private production CLI conformance passed all 13 required assertions on 8797b4a4a7e57b60ff83e29f56a245c40926d4d3. Artifact `/Volumes/TestScratch/ariadne-546-cli-close.Rw2Sg9/result.json` retains commands, executable harness and preservation snapshots. Both original refusal reproductions preserve raw index and refs; ordinary legacy publication resolves them, migration imports both archive roots, repeated apply is idempotent, and issue allocation/claim works.

Narrow fixed-then-ship review round 2 returned SHIP: BR-1 addressed, no new findings or remaining advisories. Independent affected native tests passed (126.465s). Fresh recorded suite on the corrected source head is queued as lease-2642570106, attempt a100f3e29cc43b3710f507575726fcfb; preserve this runner. Landing remains contingent on its PASS.


### 2026-10-22 — final recorded suite PASS; entering landing

Recorded attempt a100f3e29cc43b3710f507575726fcfb passed on 8797b4a4a7e57b60ff83e29f56a245c40926d4d3: all eight shards and 76 other packages passed, covering 2,159 cmd/sdlc tests with zero failures, wall 843.4s. Suite status recorded as sdlc-local/suite; log SHA256 prefix 8a3c47dc6fd4. Queue wait was 2,347s, preserved through admission. Close review round 2 is SHIP with BR-1 addressed and no remaining findings. Entering sdlc merge for authoritative integration, archive and branch cleanup; no manual main chase.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; calibration reports stale and values remain provisional. The approved root × scope contract resolves design decisions (×0.2); Git/vocab/normalization libraries already exist, so no greenfield API is needed. Four smaller-module units cover pure evidence, shared observation, bulk main/delta adaptation and worktree uncertainty; the extraction is one cross-cutting refactor. Native conformance is one real-API discovery unit. Familiarity is 1.0 after #543, with +15% design buffer. Implementation values are 40% of the selected v2 ranges. Design sum 0.36h × 1.15 plus implementation sum 1.70h = 2.114h, rounded 2.11h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.10 impl=0.08
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.03 impl=0.20
item: cross-cutting-refactor design=0.10 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 2.11
```

### 2026-10-22 — plan quality clean

Round 2 CLEAN: PQ-1 addressed, no new findings. Derived the estimate only after acceptance; no implementation before the estimate/admission gate finishes.
