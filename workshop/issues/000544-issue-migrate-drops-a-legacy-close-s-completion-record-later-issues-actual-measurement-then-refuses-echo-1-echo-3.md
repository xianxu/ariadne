---
id: 000544
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 3.57
collaboration-mode: collaboration
card_mirror: '585491c70ccd512790479fee5890879e2748a20a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T00:02:32-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3)

## Problem

Tracker cutover preserves adopted numbers on archived issues but drops the evidence that bounds them. The conservative exclusion reader consequently blocks later issues, including echo#3.

## Spec

From ops-echo:0 (10-21 evening). echo#1 closed under the legacy workflow (merge 08124b3, archive 112fd43,
actual_hours 0.80) just before echo migrated to the tracker (`sdlc issue migrate --apply`, 6f0fe49). After the
migration, `sdlc issue show 1` reads `status done, completion: no close recorded`. echo#3's close now stops before
review: `sdlc actual --issue 3` refuses "cannot auto-compute (issue #1 adopted actual has no trustworthy saved
window: completion is missing)", from the adopted-actual exclusion (`cmd/sdlc/actualexclusion.go`, `adoptionRefuse`
with no saved window).

So the migration carries a legacy-closed issue's status and `actual_hours`, but not its close binding or saved
measurement window, and every later issue whose window overlaps it is blocked from measuring. Every legacy repo that
migrates with recently closed issues hits this.

Fix: `sdlc issue migrate` carries each legacy close's completion record (the close commit and saved measurement
window) onto the tracker. For already-migrated repos, `--reconcile` (or a narrow repair verb) backfills it from the
legacy close commit and archive, with the same evidence the legacy close used. Never invent a window: if none is
recoverable, mark the adopted actual as windowless-legacy, and let the exclusion treat a windowless issue whose
close predates the candidate's engagement as disjoint (skip), not refuse.

## Done when

- An issue claimed under legacy and closed after migration measures the labelled legacy repository-plus-brain prefix and verified tracker suffix once, retains both input populations, and replays the same total after transcript appends; missing/contradictory evidence refuses (test and Echo live preview).
- Prefix-only previews refuse with an explicit no-post-cutover-owned-interval-yet reason until natural post-cutover work supplies an owned edge; no timestamp is converted into an invented owner.

- An already-migrated born-archived root card lacking completion/provenance permits later actuals only through read-only proof of its exact root/archive generation, without changing the card (test).

- A born-archived legacy import with no close remains visibly unbound and warns; proven archive/done provenance strictly before the earliest new engagement permits later actual measurement, while overlap remains strict (test).

- Migrating a legacy fixture with a closed issue keeps its completion record; `sdlc issue show` shows the close (test).
- In an already-migrated fixture, the repair restores it, and a later issue's `sdlc actual` measures (test).
- A legacy-closed issue with no recoverable window, closed before the candidate's claim, doesn't block measurement
  (test); an overlapping one still refuses.

## Plan

Durable plan: [legacy completion preservation and repair](../plans/000544-legacy-completion-plan.md).

- [x] Recover and import verified legacy close bindings, retaining exact saved measurement inputs or explicit windowless provenance.
- [x] Add a digest-reviewed single-card repair for repositories already migrated, preserving current state and supporting uncertain-push retry.
- [x] Use proven windowless legacy close ordering for disjointness, with overlapping/unknown evidence still refusing.
- [x] Preserve measurement across cutover with explicit legacy-prefix and owned-suffix evidence, deterministic replay and a claimed-before/closed-after fixture.
- [x] Verify production-path migration/repair/actual controls, full suite, approved Echo conformance and documentation; ready for close and landing.

## Log

- 2026-10-22: Fixed-head recorded suite passed on Git-verifiedf425ec39fe8a7becbaf15a491526ee7e7df0cb89: all8 shards and76 other packages,2151 SDLC tests,0 failures,777.0s wall after254s FIFO wait. Lease285754682 completed normally; exact-head sdlc-local/suite success recorded, log sha256 prefix1ae4a50f672e. Raw log copied to persistent evidence/000544/recorded-fixed-suite.log. The earlier overlap attempt stopped at the unchecked Plan guard before dispatching review; no waiver or duplicate reviewer was used. Verification checklist now complete; close round2 and landing remain the release gates.

- 2026-10-22: Final fixed-source verification passed: repair authority and correction trust race matrix40.278s; cutover/replay/queue race matrix11.742s; tracked-file golint1.020s, shelllint0.352s and gatestate0.179s. git diff --cached --check is clean. No concurrent source edits during these checks. Checkpointing and pushing the fixes before recording the required new-head suite.

- 2026-10-22: BR-1 red proved the pinned base decoder accepts an old-format hybrid while discarding its prefix. Fixed hybrid roots to version2 with distinct legacy-prefix and queue-legacy-prefix methods; plain version1 inputs remain compatible. Pinned-reader codec matrix passed0.224s, production retained replay and both correction-method controls passed33.398s, actual queue promotion passed3.914s. BR-2 red proved removal/replacement histories could be repaired; deleted the weaker repair check and shared legacyGenerationHistory with both exclusion paths. Stable repair/exclusion matrix passed73.040s, including positive repair-to-actual and four authority mutation modes. Raw red/green evidence is under common Git sdlc/test-stats/evidence/000544. Both agents stopped edits before final checks. Codec race passed2.142s and touched-command vet returned0. New fixed-head suite and round2 close remain required.

- 2026-10-22: Close round1 returned REWORK, not closed: BR-1 reproduced older decoders silently discarding hybrid prefix fields; BR-2 reproduced repair accepting removed completion/provenance history that actual then refuses. TL approved an explicit hybrid method/version boundary and one shared monotone generation validator (delivery24b0ff95-3691-4347-82fa-aed92f08a7b0). Added durable plan Revisions and assigned independent bounded corrections with red-first regressions. Re-run the suite on the fixed pushed head before re-close; the prior suite remains an exact old-head pass, not approval of these fixes. #548 remains next after landing.

### 2026-10-21

### 2026-10-22

- Claimed and started planning from fresh main after #540 landed. Read current worker/SDLC contracts and issue vocabulary. Parallel read-only investigations confirmed migrationInventory and PlanTrackerMigration bind only active codecomplete records; archived done uses ArchivedCard without completion. Existing --reconcile converts branch details and does not repair this loss.
- Inspected the issue-referenced Echo evidence read-only after AGENTS.local.md (no MEMORY.md present). Git verified latest close 476b6b12e39e87ed1febcda2a4b7ef94c8f4ec73, archive 112fd43, cutover 6f0fe498956349b2c3318000f524bf93920b3b98 and tracker root 601c7b41a7d40e2adaaea4fa1d70b8798aedfa70. The imported #1 card is done with actual0.80 and no completion. The close has SHIP/Review-Window trailers and an issue1 ledger whose final reviewed head agrees; it contains no retained ActualInputs. This is a real windowless legacy case, not permission to synthesize a measurement window. Echo checkout contains unrelated feature/untracked work; none was changed.
- ARCH-DRY/PURE: use one immutable legacy-evidence resolver for cutover and repair, then feed proven earlier windowless closes into the existing windowDisjoint decision. Candidate cutoff is Snapshot.Since, the earliest admitted engagement, never the latest claim or mutable started card field. ARCH-EVOLVE: add explicit legacy provenance; malformed modern records and damaged saved inputs remain strict. Plan is awaiting TL approval; no code or Echo state changed.

- TL approved the plan and exact repair command; ops-echo:0 owns the real Echo dry-run/apply/repeat/actual3 conformance in a temporary worktree. Entering change-code, no implementation yet. Approval deliveries d46eb30c-7e6c-4da2-9368-129e6b4cdb7c and 904d435d-de8b-446e-8fea-489c1c90ddbe.

- Plan-quality round1 PQ-1 requests named function-level test strategies and compressed prose inventories. Refined that contract without changing approved mechanisms. Existing born-archived fixture records have no actual historical close; sent the hub the narrow compatibility distinction before code.

- Implementation checkpoint: change-code passed the full-flow plan-quality and estimate gates (accepted design commit 42dad88f). The additive legacy_adoption codec, pure close selector/repair planner, shared immutable-source verifier, migration binding, digest-pinned repair CAS, strict-before exclusion and issue-show provenance are implemented. Stored adoption classes are retained/windowless-close/windowless-terminal; old unbound recovery is read-only and restricted to done numeric cards present at the actual tracker root that match a canonical immutable pre-cutover archive generation. No completion is invented. Numeric active-done imports remain supported by explicit migration, but gain no old-import fallback.
- TDD evidence so far: real archived-close migration reproduced the lost binding, then passed (7.177s); old migration repair initially refused the unknown flag, then dry-run/apply/no-op repeat passed (6.974s). Real CAS stale-card, lost-reply and cancellation-before-push controls passed (11.581s). Codec package, pure planner/selector and observation package checks passed. The final resolver/verifier/old-recovery targeted check passed (19.058s), including active-done, modern missing, correction and changed-source negative controls. Full production actual measurement covers new and already-migrated terminal records, close-backed imports, and earlier parked engagement: all six cases passed (29.569s); removing only the recovery fallback makes the two later-claim cases fail while overlap stays strict. These are targeted checks, not a completed full suite.
- Observation regression first failed because the recovered class was absent, then passed (2.516s); it confirms issue show labels the provenance read-only while leaving completion absent and card bytes/remote tracker unchanged. Full internal/observe package passed (0.161s). Some earlier concurrent targeted runs tripped the repository hermeticity guard when another authorized implementation agent introduced new dirty paths; stable reruns passed. This is test scheduling during shared edits, not registered flake evidence.
- Remaining verification: stage all new files before tracked-file lint, run a stable combined targeted/race/vet pass, build an exact committed binary, coordinate ops-echo:0's approved dry-run/apply/repeat/actual3 live check, then push and record the full branch suite. Close review and landing remain outstanding. No Echo writes, PR or recorded #544 suite yet. The estimate's module rows include implementation tests; full-suite wait and owner handoff may increase measured actual and must not reduce verification. All bounded implementation agents have finished; no own recorded check is queued or running at this checkpoint.

- Resumed checkpoint at TL request; internal issue/tracker/observe packages passed (0.369s/47.004s/0.167s), tracked-file golint/shelllint passed, vet passed. Internal provenance/planner/observation race tests passed (3.886s/3.535s/1.225s). Combined production checks exposed old migration fixtures refusing source history before digest, so they correctly blocked proceeding to the recorded suite.
- Echo-owned pre-landing check on built d50fc9f42511c6d7d7698fe021d258dfc045719d (modified=false) refused read-only: source history lost its path; no apply attempted (delivery1244b5bb-b6d9-42b3-acd2-88dcd44728bc). Captured native Git output persistently at common Git test-stats/evidence/000544/echo-native-history-before.log. Root cause: --follow -M1% follows the issue beyond its creation into an unrelated project template, rather than losing the real archive rename. New real-Git template fixture reproduced that refusal. Trying Git's normal similarity threshold passed that fixture but hid the existing damaged-identity control, so it was rejected.
- Fixed the bounded reader by retaining the identity-sensitive scan and proving issue creation before stopping at non-issue template ancestry: the oldest issue path must be absent from every parent, and later parsed history may contain no other issue ancestry. Added a real merge control whose second parent already holds the issue, so first-parent-only proof fails. Native follow regression plus close/identity/damaged-evidence/merged-close/old cutover controls passed together (16.847s). Updated the legacy migration test helper to expose original errors when no digest prints. Fresh binary and Echo repeat still pending at this log entry.

- Echo owner confirmed the repair/apply/no-op sequence on fe236ba (delivery5a42758e-fde1-45f6-a108-952fd49b0ccc): correct latest close, unchanged actual/checkout, token migrate-000001. The subsequent actual3 was only0.03h over95s, exposing missing pre-cutover engagement; the TL explicitly folded that measurement gap into #544 and approved the revised prefix/suffix plan (delivery66dc716c-80f0-41d0-97d0-db740be9ebee). Echo closes remain held until the revised live measurement.
- Added the planned complete pinned tracker-card history fence, rather than comparing only equal endpoints: old and explicit imports reopened then restored no longer inherit the old bound. Shared fence verifies exact pinned raw/blob, root ancestry and bounded complete history, rejects intervening lifecycle/correction/authority changes, permits only monotone matching repair. Restored-reopen tests were RED before the fence and GREEN after; expanded targeted controls passed44.570s, explicit same-value correction-history negative passed1.790s. Different-ID templates are allowed only beyond proven issue birth; the same-ID ancestry guard remains strict.
- Cutover continuity is implemented as approved: bounded legacy-v1 child inputs plus the exact cutover commit; no invented owner. Discovery includes caller and verified primary checkout plus brain, so a temporary measuring worktree can find the original legacy sessions; unrelated linked workers are excluded. Clipped modern windows retain their true dispatch origin and queue records use the post-cutover boundary. Production claim-before/migrate/close-after measurement and replay passed4.104s, including transcript appends. Combined prefix authority/source/clipping controls passed10.809s. Legacy working-transition and first-commit fallback controls passed4.104s. Codec JSON/YAML, recursion/size/overlap, multi-harness discovery and pure join/engine controls are covered; complete stable combined verification is still pending.

- Live conformance passed on clean branch binary b1ea6825219db0cf214cef91c1e7ca3e5c740821 (ops-echo:0 deliveryf8174005-af47-42ac-af83-c57928432164): actual3 at its pushed issue head a40d090 measures0.50h, legacy20:05:50–20:25:53 plus tracker20:25:53–20:45:16, queue-aware with6 admitted sessions. Main-head preview is naturally a different pinned history; the owner repeated at the real issue head. Echo10 independently measured0.59h. TL accepted this as #544's live proof (delivery59ea794e-0b24-4d8f-85cf-c1b571d0117a).
- TL keeps Echo4/5 prefix-only previews refusing until natural reconcile/work supplies a post-cutover owned edge. Added the exact reason no post-cutover owned interval yet, with a production fixture. Echo9's cross-issue queue-attempt refusal is separately filed and reserved as ariadne#548, next after this landing; no cross-issue queue policy is changed in #544.
- Final audit reproduced another evidence-retention boundary: a positive prefix plus known-zero owned suffix measured correctly but saved no suffix sources, so replay became a telemetry gap. Hybrid-only exact empty-source retention fixes it without accepting missing telemetry; fresh/replay/missing-file controls passed18.073s and the focused race test4.728s. New refusal/replay controls passed6.569s. Full production root prefix/replay/authority and repair race controls passed29.815s. The broader combined matrix printed PASS for all behavioral cases (346.395s), then the real-repository hermeticity guard correctly failed because the authorized audit changed two tracked source paths concurrently. That run is not claimed as a passing check: repeat the stable matrix after all edits stop, then queue the recorded full suite. This is an implementation scheduling error, not registered flake evidence.

- Stable verification on Git-verified31a60351a84c70c88695af48556f634bbbbbf338 completed with no concurrent source edits: migration/legacy/repair/exclusion/cutover matrix passed248.586s; owned measurement/replay/queue matrix passed92.540s. Both include the normal real-repository hermeticity guard. Prefix refusal/replay final controls passed6.569s; tracked-file lint and codec package passed; vet returned0. Raw logs and native/live evidence are persisted under common Git sdlc/test-stats/evidence/000544 with SHA256SUMS. Checkpointing this result before queuing the recorded full suite; no further implementation edit is planned.

- Recorded branch suite completed on Git-verified912d87634143d023fab266ded7e212c53082db70: all8 SDLC shards and76 other packages passed,2149 SDLC tests,0 failures,808.9s wall after135s FIFO wait. Lease712010329 survived the capacity interruption; no duplicate or requeue. The exact-head sdlc-local/suite status is recorded. Saved log: common Git sdlc/local-checks/912d87634143d023fab266ded7e212c53082db70/suite.log. TL accepted the result and directed close/land, then #548 code (deliveryc4635b43-eb9c-41eb-9592-c4174b66fe93).

## Revisions

### 2026-10-22 — unbound terminal imports

TL delivery 82679d7d-74d5-4320-9051-4232bdd6f97f extends the same strictly-earlier rule to unbound legacy imports using a proven archive/done commit. Preserve absence of a close, never fabricate review/window evidence, and keep overlapping cases strict. The plan now uses one LegacyAdoption envelope record across close-backed and terminal-only imports.

### 2026-10-22 — existing unbound imports

TL delivery ec3a2adf-2b31-4b08-a330-56f170a32a64 adds read-only root/archive proof recovery for old born-archived imports without completion or provenance. Strict-before earliest engagement remains required, no card write or fake close occurs, and modern missing records remain strict.

### 2026-10-22 — pre-cutover measurement continuity

TL delivery9668b551-4c1c-44d3-8751-52b9a5f6a69 folds the live measurement gap into #544. An issue claimed under legacy and closed after migration must measure both its provable pre-cutover interval under the existing labelled repository-plus-brain heuristic and its tracker-owned suffix. The plan revision preserves distinct evidence and retained replay, with no invented owner/window or suffix-only adoption on partial failure. The estimate above this revision predates this additional measurement work; measured actual remains authoritative.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is stale, so provisional. Issue shaping uses 0.70 design without discount and 0.20 raw implementation. Codec is a typed-data prototype (0.50 design ×0.2 spec discount; 0.50 raw impl). Shared proof resolution is a single-concern greenfield Go module (1.20 design halved because existing contextual Git, review-window and ledger adapters provide the IO foundation, then ×0.2; 0.80 raw impl). Migration and exclusion each extend smaller modules (0.30 design ×0.2; 0.50 raw impl). Repair uses the existing tracker API/CAS integration with fault/retry tests (1.00 design ×0.2; 1.50 raw impl). Docs use 0.10 design ×0.2 and 0.20 raw impl. One close review uses 0.50 raw impl; owner-run Echo discovery uses 0.60 raw impl. Apply v3.1's 40% to implementation once, familiarity1.0 and 15% thorough-plan design buffer. Design1.26 ×1.15 + implementation2.12 =3.569, rounded3.57 focused ship hours. Queue waits are not active implementation; no tests or evidence are waived to meet the estimate.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: issue-spec design=0.70 impl=0.08
item: typed-data-prototype design=0.10 impl=0.20
item: greenfield-go-module design=0.12 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: api-integration design=0.20 impl=0.60
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
total: 3.57
```
