---
id: 000569
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4a93dae95bc7ebe64cb03be2f9658261d744e79a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T01:30:48-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "1963757c", done: "0f903ae5"}
---

# Diagnose legacy actual-window refusal blocking pair#162

## Problem

The hub reports that central sdlc actual --issue 162 from pair-slot2 refuses because old landed pair#151 has no trustworthy saved adopted window: legacy completion workshop/history/issues/000151-hierarchical-thread-menu.md has no finalized whole-issue close in its saved boundary ledger. pair#162 was filed on September 1 and is closing now. Its close waits on diagnosis; no urgency before about08:00 on October23.

## Spec

Diagnose the provenance and scope of the cross-issue adopted-window dependency before changing anything. Determine whether genuine historical evidence can be restored or whether an unrelated legacy adoption should be excluded with a visible label. Never fabricate a close, window, ownership history or measured hours. Keep Pair checkout and historical evidence unchanged during diagnosis.

## Done when

- Reproduce the exact refusal with the central binary and trace the adopted-window selection to its data and production functions.
- Compare pair#162 engagement and migration bounds with pair#151 close/archive/measurement evidence, distinguishing proven disjointness from unknown overlap.
- Record a concrete recovery or root-fix recommendation, evidence gaps and acceptance controls; report the cause to the hub before implementation.

- Exact old unmarked imported archive/done generations may be skipped only when their immutable end strictly predates the earliest candidate prefix/engagement, with a visible issue/end label and no numerical/review repair.
- Equality, overlap, changed/restored generations, modern corrupt authority and unreadable history still refuse; strict legacy completion/replay verification stays unchanged.
- The captured partial legacy ledger and full hybrid measurement path have regression coverage; recorded suite, close review and central landing pass before telling pair:2 to measure.

## Plan

- File and claim; enter planning on a branch from main.
- Reproduce read-only in pair-slot2, inspect only the implicated historical records and measurement call chain.
- Verify the cause with the smallest nonmutating probe; preserve returned evidence and rejected alternatives.
- Publish diagnosis and recovery recommendation to the hub; obtain approval for any scope or evidence-policy change.

## Log

### 2026-10-23 — dispatch

Authenticated hub delivery9c270b83-96f0-4235-b536-7f809314451c assigned this after #563. #563 is now landed as PR406, so investigation can begin. The hub asks to file, claim and report; no speculative Pair evidence repair is authorized.

### 2026-10-23 — reproduced and traced; no evidence restoration performed

Central binary in pair-slot2 reports path=/Users/xianxu/workspace/ariadne/bin/sdlc, revision584d909713dbd858397f1330a23ee026861a5844, modified=false. sdlc actual --issue162 reproduces the exact warning (the command exits0 but produces no numeric measurement). The read-only sdlc issue migrate --repair-completions --issue151 exits1 with the same missing-finalized-whole-issue-close error. No apply, correction, exclusion, label, Pair edit or Pair commit was performed.

Production path: computeActualForPurpose -> resolveOwnedActualInputsFor -> loadAdoptedExclusions -> recoverUnboundLegacyAdoption -> resolveLegacyCompletion. The adopted-actual guard protects against counting time already adopted by another issue. Its calendar-date filter deliberately widens old card dates by a timezone-safe48h; #151 updated2026-08-31 still passes for a September1 candidate. Selection does not establish a shared slot: #151 has no tracker owner, completion, legacy-adoption marker or correction, and its tracker history has only the original migration-root commit. The legacy proof errors before slot comparison and before immutable temporal disjointness can be established.

A temporary Go overlay (no tracked source edit) invoked the real input resolver and exclusion loader read-only. On own source pinb10a950e2693b79fe69b1d4279f1b752c9d48162 it printed:

- candidate_since=2026-09-01T14:21:09-07:00, until=2026-10-22T22:03:04-07:00, method=active-time-v3-engagement-legacy-prefix-v1;
- legacy prefix September1 14:21:09 through September28 00:06:01 PDT;
- tracked engagement October22 18:09:43 through22:03:04 PDT, root pair-slot2/pair;
- immutable_archive_before_candidate=true and coarse_card_filter_may_overlap=true;
- the original production exclusion refusal naming #151.

The first boundary is the parent of the legacy #162 draft commit, not its created date or its much later tracker claim. Git independently shows parent7f85f5c4c1934f33ab23d0a0eee5e90096ffecc7 at September1 14:21:09; draft858a4452fbee8c329526adcb541fc18870f86d27 follows at14:22:14. Tracker rootc039dc4b8140f73ecc140f80e945beebadc4da7f is September28 00:05:53; marker commitd636b304405f9208a2341e8cffbcfb504f140a0c is00:06:01. This is hybrid legacy attribution, not an inferred old shared-slot claim.

#151's actual20.07h was already adopted before cutover. Git close commitd5d23dcc210f8b80c7cb440e3ffd8dc2e09e31bd (August31 18:17:38 PDT) says Review-Verdict: SHIP and Review-Window:9157998f..309a9717. Its own immutable close-gate ledger has24 rounds; round24 (18:17:17) records protocol_error: no valid findings block, and no finalized whole-issue reviewed head exists. Earlier accepted rounds are milestone boundaries, so they cannot substitute for a whole-issue close. The archive commitc643b4b2638a71f82e8e4cc1363fd7c023f198f9 is August31 18:20:40, strictly before even the candidate's earliest legacy prefix. No matching finalized whole-issue record was found in the implicated immutable close/archive history. Restoring a finalized review or numeric snapshot would invent authority; the strict repair refusal is appropriate.

Temporary overlay source and raw output are retained at common test-stats/evidence/000569/readonly-input-probe.go.txt and readonly-input-probe.log; the overlay and temporary build-input directory were removed. It passed9.619s; targeted existing timezone-filter, strict-disjointness and damaged-legacy-evidence controls passed1.642s. These verify the diagnosis, not an implementation fix. No full suite was started for this read-only investigation.

## Revisions

### 2026-10-23 — proposed policy boundary; awaits hub/operator approval

The cause is an overbroad dependency on trustworthy historical close replay even for a provably earlier imported generation. Recommend one narrow, separately proved archive/done end bound for exclusion relevance: only a numeric done issue present in the real tracker root and its exact pre-cutover archive, with matching identity/actual/started and generation continuity (no reopen, correction, modern completion or provenance substitution), may use that immutable end solely to prove end < candidate.Snapshot.Since. Since must include the entire legacy prefix/earliest engagement, not only today's owner. Print a visible diagnostic naming the issue and bound and saying its adopted number was not replay-verified. Keep its historical numeric actual unchanged.

ARCH-DRY/ARCH-PURE: extract/reuse the existing immutable source/generation proof rather than duplicate a history walker; keep close/replay validation strict. This is not permission to reinterpret malformed saved modern inputs as legacy, to accept #151's review as valid, or to ignore any potentially overlapping adoption. Equal/overlapping/missing bounds, source mismatch, reopened generation, unreadable/bounded-out history and malformed modern authority still refuse. A single lifetime lease cannot prove a historical generation/time relation; immutable Git object identity and bounded complete history are the appropriate authority.

Proposed acceptance strategies before implementation: reduce the captured real #151 ledger/old unmarked-import shape into a real-Git regression with #162's legacy-prefix boundary; prove the labelled disjoint skip, plus strict equality/overlap, initial/reclaimed ownership, archive/source tamper, reopened/restored generations and modern missing/corrupt completion refusal. Retain TestLegacyCompletionProofRefusesDamagedEvidence unchanged. A dedicated classifier handles only temporal relevance; no new persisted schema, hours, state store, runtime scan or production artifact family. The existing output/snapshot provenance path should carry any label. No production code has been edited; report this concrete recommendation to the hub before expanding scope.

### 2026-10-23 — approved implementation refinement

Authenticated hub delivery3934fa98-49d4-4b59-98d1-478677117b64 approves the narrow archive/done bound and asks to notify pair:2 after landing. Proceed within the quick-flow shell, one atomic close boundary, no Mx split. Preserve the already-approved policy; no further operator round is required for routine implementation.

Reuse the complete bounded immutable history reader in legacycompletion.go through a private archive-bound mode. resolveLegacyCompletion remains its strict wrapper. A shared recoverUnboundLegacyAdoption implementation takes the same purpose, preserving root membership, archive identity, immutable bytes, scalar types and full tracker generation checks; a separate temporal-bound entry point alone requests bound mode and returns only a time, never a valid close or number. loadAdoptedExclusions invokes that entry point only for unmarked imports without a completion, compares through the existing pure legacyWindowDisjoint, and appends a visible fresh preview/close warning. Existing immutable historical replay is unchanged, and no persisted field is added. Strict repair/show/completion paths remain on strict mode.

Actions: rerun measurement; repair named evidence.
| Internal state | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Exact unmarked import, proven archive/done end before earliest candidate | Measured with labelled disjoint legacy adoption | Continue normal close | None; old actual/review unchanged |
| Equal/overlapping/unknown bound or modern malformed evidence | Cannot measure, named evidence | Repair evidence and rerun | None; no adopted hours |

Test first: a real-Git archived old import with the captured round24 protocol-error ledger must let a later candidate proceed with the label, while strict completion verification still refuses it. Extend its controls across equality/overlap/legacy-prefix bounds and real tracker mutations, and retain existing source/root/reopen/modern corruption tests. Risky seams: resolveLegacyCompletion mode cannot yield completion authority in bound mode (strict-refusal control); recoverUnboundLegacyAdoption root/generation checks cannot be bypassed (real-Git provenance mutations); loadAdoptedExclusions cannot use only latest ownership (hybrid earliest-prefix control); visible label cannot disappear (public-output assertion).

Cost remains one bounded complete source-history walk plus the existing tracker-history validation per relevant old adoption; no process/topology scan, schema, durable family, cache or new scheduler. Reuse the existing counted history-scan guard. The mistake class is semantic authority/relevance ordering, so the real-Git behavioural fixture is the guard; a lexical linter cannot infer temporal disjointness. Update help/atlas and lessons, run focused normal/race and a private read-only production CLI against Pair; normal central suite/close/merge only, no staging or live queue mutation. Notify the hub and pair:2 only once the landed central binary is rebuilt and measurement actually succeeds.

### 2026-10-23 — implementation and targeted verification checkpoint

Approved bound-only mode reuses the immutable source/generation reader and cannot yield a completion. Strict completion/show/repair/replay wrappers retain strict validation; only the old unmarked-import exclusion path asks for the temporal fact and prints the label. Shipping additions remain49 lines including help. The exact original Pair#151 close-gate bytes are committed as the regression fixture; no original evidence is changed.

The first regression setup used SplitCard instead of the old migration codec and failed on an ID mismatch, so it did not count as RED. After using PlanTrackerMigration, the regression failed for the exact missing-finalized-whole-issue-ledger refusal in1.860s. The implementation made it pass2.796s, while equality/overlap and strict completion refusal stayed intact. Generation controls add numeric/start mismatch, modern completion, reopen/restore and provenance-strip/restore. A real public measurement fixture recreates an old importer with the partial ledger and a genuine legacy prefix; after correcting fixture bootstrap clocks (started cannot be later than its tracker-root commit), both its measured/labelled and overlapping-prefix controls passed6.886s. Go/shell lint passed1.034s/0.395s.

The broader legacy/adopted test cases all printed PASS, but TestMain's hermeticity guard correctly failed the187.533s run: I edited atlas/workflow/issue-tracker-migration.md while it ran. This is attributable to my concurrent edit, not a test flake or main failure; no registration or blind full retry. Freeze ALL checkout files (including atlas/Log), then repeat the targeted verification unchanged. Documentation and lessons edits are complete before the repeat. No full recorded suite has been queued yet. Next: targeted normal/race controls, clean code commit, private read-only CLI proof of pair#162, then pushed-head recorded suite/close/central landing. Keep historical #151 evidence strict and unchanged.

### 2026-10-23 — focused verification complete; prepare private CLI proof

With the entire checkout frozen, the broader TestLegacy/TestVerifyLegacy/TestAdopted/archived-migration/parked-claim controls passed138.888s. Race verification of the new disjointness/generation controls passed9.830s; an explicit public hybrid old-partial-ledger selector independently passed under race10.953s. The new production branch stays disjointness-only; strict historical completion/replay and damaged-modern controls passed unchanged. No blind suite retry or flake registration was used for the earlier concurrent atlas edit.

Implementation source remains44 added Go lines plus5 embedded help lines, inside the quick shell. README, both relevant atlas pages, help, exact captured ledger fixture and lessons are updated. No persisted schema, number, prior review record or peer checkout was changed. Commit the clean code checkpoint, build a private production CLI with verified revision/modified=false, and invoke actual162 read-only in pair-slot2. If another genuine blocker appears, report it instead of weakening evidence. After private proof, queue the central pushed-head recorded suite, close/PR and normal central merge; post-land rebuild and actual success precede the direct pair:2 readiness message.
