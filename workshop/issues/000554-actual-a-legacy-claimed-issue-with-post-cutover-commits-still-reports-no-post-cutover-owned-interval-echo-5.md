---
id: 000554
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '89e6ccef04f309d7ac537cfc619474d9ec5cfcc6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T11:00:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "4c6861a1", done: "173e9989"}
actual_hours: 2.30
---

# actual: a legacy-claimed issue with post-cutover commits still reports 'no post-cutover owned interval' (echo#5)

## Problem

## Spec

From ops-echo:0 (10-22 ~11:00). echo#5's close refuses at measurement on fleet sdlc 1c0595d (and in a read-only preview
at pushed head 350a62d): "can't auto-compute (no post-cutover owned interval yet: measurement has no verified owned
interval before its pinned right edge; reconcile and record post-cutover work before closing)". But echo#5 has
post-cutover work: tracker cutover 10-21 20:25:53; #5 commits 20:34:31, 20:35:14 and 21:00:04; a merge of main
10-22 09:17:50; "#5: merge origin/main (echo#3 contracts); add VocabFromProto" 10:59:11. Its code-complete hold was
recorded around 21:00 on 10-21, which may be the pinned right edge, so the post-cutover owned interval is computed
as empty or isn't recognised. ariadne#544 deliberately refused echo#4 and #5 while they had no post-cutover work; now
the work exists and it still refuses.

Likely causes to check: the owned tracker interval after cutover never opened for an issue claimed pre-cutover and
reconciled (the reconcile event isn't treated as an engagement start); or the right edge is pinned at the
pre-existing hold and excludes later work.

## Done when

- A legacy-claimed issue reconciled after cutover, with commits after cutover, measures its post-cutover owned
  interval (real-git fixture of echo#5's shape).
- echo#5's 'sdlc actual --issue 5' measures (owner-run live proof by ops-echo).

## Plan

- [x] Distinguish missing tracker ownership from an owned interval outside the pinned measurement window; carry the exact claim action through preview and close diagnostics.
- [x] State the unowned cutover-to-claim gap explicitly in refusal and successful hybrid measurement output, including retained replay.
- [x] Add the captured Echo refusal and real migration, reconcile, claim, owned-work regression path; preserve historical and correction pins and all attribution arithmetic.
- [ ] Update docs and verify focused/race/recorded checks; obtain ops-echo's owner-run live actual proof before landing.

Implementation design: [explicit ownership recovery](../plans/000554-cutover-ownership-recovery-plan.md).

## Log

### 2026-10-22
- 2026-10-22: closed — Recorded suite exposed a deterministic issue-hint convention defect, not a flake: raw padded IDs in claim recovery. New independent padded-ID cases and existing TestIssueHintsRenderDecimalIDs reproduced red 0.517s; existing issue.CLIRef fixes it. Focused renderer/capture/padded-ID/source-guard controls PASS 0.447s, lint catalog PASS 0.736s, diff clean. Ownership, queue clipping, release/reclaim, native replay and prior race evidence remain unchanged; new delta is pure formatting within the approved recovery contract. Fresh recorded suite started on pushed e01e95a065207a21c29d8e74465f5e7aa69f0448 and may overlap full review under tl-worker rules. --no-plan-check acknowledges remaining recorded-suite and owner-run Echo live-proof holds before landing; no attribution/schema or scope expansion. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-22: closed — Full close reviews the approved added first-claim queue behavior and disposes BR-1 review-routing finding; no fix-only flag. Exact live zero-owner refusal reproduced in Claude/Codex fixtures, corrected queue/ownership/legacy/release controls PASS 75.433s, native lifecycle/replay race PASS 15.091s, strict queue-boundary race PASS 16.971s, lint PASS 1.711s, integrated-head queue controls PASS 53.542s. Required pre-noon main integrated without conflicts. Recorded suite lease-3755615631 retained on c7778d78873aaa2a2e61e48ef295c6fa181df6bd; later changes are review evidence and prose only. The queued suite may overlap review under tl-worker rules and must pass before landing. --no-plan-check acknowledges only remaining recorded gates and hub-coordinated owner-run Echo live proof, both pre-merge holds. Raw queue evidence, strict release/reclaim and immutable historical replay remain preserved. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-22: closed — Real Claude/Codex migrate-reconcile-claim controls, captured Echo refusal, credited-gap exclusion and immutable historical replay PASS; focused legacy/owned/Qoder controls PASS 358.689s, tracker controls PASS 4.137s, ownership race PASS 45.679s, lint catalog PASS 1.442s. Recorded suite lease-3451749813 retained on pushed bf6e38e1; queued suite may overlap this review under tl-worker rules and must pass before landing. The open final Plan row is the recorded gates and hub-coordinated post-close/pre-merge owner-run Echo proof, so --no-plan-check acknowledges that hold rather than marking it done. No attribution arithmetic, ownership date or saved hours changed. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Claimed in ariadne:4 and entered planning from landed main. Consulted ariadne:1, author of #544, per dispatch (authenticated advice `43ec058e-eb34-43a3-b009-e30bb3b1e827`). Legacy `started` never creates a tracker owner. Historical measurement replays its completion pins; a close candidate uses current code/tracker. Reconcile changes ancestry/mirrors, not ownership.
- Authenticated ops-echo evidence `c0e53bb7-de22-435e-8e0a-a53ea8dbd05e` and echo:2 capture `e5972fcd-a514-4203-97c1-59db6210817a` disprove the initial hold/right-edge suspect for this incident. The card is working, with no claimant, no completion binding and only its migration commit. Echo:2 confirms no post-cutover claim/change-code/card setter; its codecomplete hold was prose in the Log. Empty ownership intervals are correct.
- Original captured close refusal: `/Users/xianxu/workspace/worktree/echo-slot2/echo-5-close-refusal-2026-10-22.txt`, SHA256 verified locally as `b292b28128eed6b0b743217532440ea294793b24a4f98e7a851eaedbd5636f85`. It includes the full omitted-actual close invocation, exit1, and the misleading reconcile/restore recommendation. Preserve the bounded capture for the regression fixture; no Echo state was changed by this worker.

## Revisions

### 2026-10-22 — root cause and approved recovery scope

Ops:0 approved this revision in authenticated receipt `9d48a9c0-5144-4084-bb86-dd422b2a0800`. Preserve the original incident report above; its codecomplete/right-edge hypotheses are not this incident's cause. This issue fixes explicit ownership recovery and visible undercount, without weakening attribution or inventing a retroactive owner. The hub will have Echo claim #5 so future work counts.

Revised Spec: when a current working issue has no claimant and pinned history proves it has never had a tracker owner, say `no owner since cutover: run sdlc claim --issue N from the slot doing the work`. Preview and close must give that action rather than reconcile or restore evidence. State plainly that work between cutover and claim is unowned and not counted. After a real claim and subsequent owned work, retain the legacy prefix and verified owned suffix, with the intervening gap visible. Reconcile remains read-only with respect to ownership; finalized historical/correction authority remains pinned.

Revised Done when (supersedes an implication that all post-cutover Git commits are owned):

- No-owner migrated working cards still refuse measurement, with the exact claim instruction, the unowned/not-counted statement, and no misleading reconcile/restore instruction.
- A real migrate → reconcile → claim → owned-work fixture measures both the legacy prefix and only the evidenced owned suffix; gap transcript events do not count. Accepted historical replay/correction pins and other missing-evidence refusals remain unchanged.
- Successful hybrid output and replay expose the cutover-to-claim gap; no saved hours or ownership dates are rewritten.
- Ops-echo runs `sdlc actual --issue 5` as owner against the final build, after real claim and owned work, and supplies the live result. Stage/clear timing goes through the hub; no worker writes Echo's card or operator documents.

### 2026-10-22 11:20 PDT — approved continuity oracle and drain checkpoint

Authenticated hub receipt `0efb89bd-348b-4815-8190-b50f00369f34` approves replacing the overstrong gap-event total-hours equality oracle: prove that no credited segment or phase intersects the unowned cutover-to-claim window. Preserve the existing continuity arithmetic: an event before claim may shape continuity inside the owned window, without crediting the earlier unowned window. Name this effect in the test comment. Equal-hours assertions remain required for immutable historical replay. This supersedes the earlier plan wording that gap events cannot increase the measured suffix; production attribution arithmetic remains unchanged.

Drain checkpoint (authenticated supervisor receipt `8e579609-d86b-4e28-bc7f-3f9e35d37023`): implementation is unfinished. Typed ownership recovery, shared preview/close guidance and the retained hybrid-gap warning are written, along with renderer controls and real Claude/Codex migration/reconcile/claim/replay fixtures. The original Echo capture is copied into testdata but still needs an executable capture-backed assertion. No live Echo state was changed.

Verification so far: the frozen red run observed the old reconcile/restore diagnostic; after the source fix, the focused `TestActualOwnership` run failed only the Claude/Codex lifecycle total-hours invariant (0.2 versus 0.23333333333333334), because the injected pre-claim event extends continuity by two minutes inside ownership. This is the approved test-oracle correction above, not a test pass or flake. No production continuity fix is authorized or needed. Renderer controls passed in that run. No recorded suite, review, PR, stage or merge runner exists; no queued work is outstanding, and all delegated work has finished.

Stopped before changing that test oracle. On an authenticated drain lift, first replace it with an independent no-credit-in-gap check (phase labels alone are not credited activity), retain historical replay equality, and add the continuity comment. Then finish capture-backed and diagnostic eligibility controls, docs and prevention lesson; freeze edits and rerun focused/neighbor/race checks. Only after those pass, commit/push the final build, open the PR and run the recorded suite/close gates. Coordinate owner-run Echo live proof with the hub before landing. This checkpoint deliberately preserves unverified implementation and a known failing test; it is not SHIP.

### 2026-10-22 11:38 PDT — resumed implementation and focused verification

Authenticated hub lift `f028e174-7c34-4966-a802-59f3feb1fa63` ended the drain; resumed from the checkpoint. Completed the approved credit-window oracle: positively allocated segments intersected with counted phases never overlap cutover→claim; their independent credit sum matches adopted hours. The test names pre-claim continuity into owned time, and retained historical replay still requires identical hours and pins. Corrected two newly reached fixture setup errors (title uses its supported setter; real unclaim requires committing the refreshed mirror); no attribution code changed in response.

The preserved Echo terminal capture is consumed by a normal-suite control. Replaying its untyped old diagnostic cannot invent ownership authority; the proven typed failure replaces reconcile/restore guidance in preview and close. Real Claude/Codex migration→reconcile→claim fixtures prove the owner transition, genuine owned work, excluded gap and immutable completion replay. Real later-claim/release controls retain the pinned-right-edge refusal instead of falsely claiming never-owned history. Existing malformed-history, release/reclaim and Qoder controls remain green.

Verified frozen source: `go test ./cmd/sdlc -run '^(TestActualOwnership|TestOwnedActual|TestActualAcrossLegacyCutover|TestLegacyPrefix|TestQoderActualClose|TestClipOwnedAfterCutover)' -count=1` PASS (358.689s); tracker `Test(DeriveEngagements|EngagementHistory)` PASS (4.137s); `go test -race ./cmd/sdlc -run '^TestActualOwnership' -count=1` PASS (45.679s); Go lint catalog/hygiene PASS (1.442s); `git diff --check` clean. Default GOCACHE verified as `/Users/xianxu/Library/Caches/go-build`, with no override. README, actual help, atlas and prevention lesson document strict ownership and visible undercount. No saved schema, owner date, historical number, continuity arithmetic or usage policy changed.

Next: publish the verified branch and run the required recorded suite and fresh close review. Live Echo proof remains a post-close/pre-merge hold coordinated with the hub on the final staged build; no Echo/operator state has been touched. The final Plan row remains open until recorded checks and that live proof finish.

Timing correction: the preceding focused-verification entry was recorded at 11:34 PDT; its 11:38 heading is a transcription error. PR #400 is open; recorded-suite and close evidence follow separately.

### 2026-10-22 — SHIP; recorded suite retained in FIFO

PR #400: https://github.com/xianxu/ariadne/pull/400. Close round 1 returned SHIP, high confidence, with zero findings; the reviewer independently ran ownership/race, engagement/history and selected legacy/replay controls and accepted the revised Done when plus explicit post-close live-proof hold. Evidence commit verified with Git: `50fa503cc50a92b6d8cb9296bb26f3a72f1e22f3`; codecomplete mirror `e1971ee53282b9b5f4c1603b75d4e97069c1c95d`. Close adopted measured 0.95h automatically. No hours were hand-entered.

Recorded suite is retained on source head `bf6e38e1ebc285ef74315d90849b2abe2a485db6`: lease `lease-3451749813`, attempt `170d1711e572ce1b452eead85cdddb05`, terminal session 31357; latest observed FIFO position 10. Log: `.git/sdlc/test-stats/inflight/170d1711e572ce1b452eead85cdddb05/run.log` in the common Git directory. Keep this runner and FIFO place; do not start a duplicate, cancel or restart it. Close runner session 64102 completed successfully; no close is running. Only close evidence/mirror prose changed after the suite source head.

Next gate is the retained recorded-suite result. On PASS, coordinate final stage/build with the hub and obtain ops-echo's owner-run `sdlc actual --issue 5` evidence showing measured output and the explicit excluded gap. Append that live proof, finish Plan checkboxes, clear stage and enter `sdlc merge --yes`; do not enter landing before the live hold is satisfied. A suite failure needs same-head classification and normal repair/resume, never a duplicate queue entry or guessed attribution.

### 2026-10-22 — approved first-claim suite queue boundary

Live Echo follow-up (authenticated hub receipt `d9d4fbbf-adb2-43fa-b2a6-90d60e75c6b0`) reached a second consumer of the same first-ownership boundary: suite attempt `21fd6681014fb81d65bb1ff403b31d51` queued before claim and acquired after it. The hybrid attachment path resets queue selection to cutover, so its selected queue interval includes unowned time and the unchanged strict selector refuses with zero full matching engagements. Modern inputs already begin at ownership; this is the hybrid left-edge mismatch, not evidence of retroactive ownership.

Hub approval `4f847ff8-9947-4df2-a754-f5ff80f98819`: bound hybrid queue selection at max(cutover, first proven ownership start). Preserve exact raw endpoints/duration, immutable tested-head attribution, and strict release/reclaim-crossing refusals. The existing cutover→claim warning labels the unowned prefix; no guessed time, earlier owner, new stored field or retrospective hours rewrite. ARCH-DRY/ORDER: use the existing pinned engagements and queue-window clipping at the attachment seam; the selector's ownership policy stays unchanged.

Revision Plan:

- [x] Add the unchanged real Echo suite capture and a native Claude/Codex real migration→reconcile→claim fixture. Observe the straddling attempt fail under the old attachment edge, then prove only claim→pinned-right-edge wait is selected; retain raw duration and excluded-gap warning.
- [x] Apply the bounded attachment-edge fix and preserve release/reclaim, foreign/unknown head, malformed evidence and immutable replay controls.
- [ ] Verify corrected head, publish, queue its recorded suite and re-close with the fixed-then-ship review. Retain the owner-run Echo live-proof hold before landing.

The original queued suite `lease-3451749813` on the superseded source head was cancelled through `sdlc local-check cancel --run lease-3451749813` before code changes; the CLI confirmed cancellation. Do not revive or duplicate that obsolete run. Copied the live JSON read-only from Echo's common Git evidence into `cmd/sdlc/testdata/actual/echo5-claim-straddling-suite.json`; SHA256 `830e2848c584132315e093e2b4699a8667891ed7ce03ee257fdd342a1ff53b5e`. It records exact queued/acquired endpoints, raw queue 5204.94009 seconds and passing outcome. No peer card, tracker, attempt or checkout was changed.

### 2026-10-22 — first-claim queue regression red/green

The frozen native Claude/Codex regression reproduced the exact live refusal (`suite attempt 21fd6681014fb81d65bb1ff403b31d51 has 0 matching pinned slot engagements`) before the production fix, exit1, 16.541s. Initial fixture setup errors were corrected before accepting that red evidence: a rebased SHA requires its generated log path to be rebased too, and effective local Git transport remains a file-scoped repository. The test uses the existing hybrid-queue fixture adapter at the production attachment seam, retaining genuine file-scoped ownership and replay authority; no new transport or production seam was added.

The fix bounds only hybrid queue selection at max(cutover, first pinned engagement start). The pure selector, raw queue endpoints/duration, tested-head membership, release/reclaim refusal and historical pins stay unchanged. Real captured queue/setup/runtime durations are preserved after shifting the capture clock into the disposable fixture. It selects exactly claim20→pinned-right-edge30 (600 seconds) from the captured raw 5204.94009 seconds, preserves the visible unowned-gap warning, and replays identical hours/pins/wait after the ambient attempt is malformed and the card changes. The wholly preclaim captured failed attempt has an unknown fixture SHA and is excluded before immutable head attribution. A cancelled invocation with no complete record is not fabricated.

Second read-only capture: `cmd/sdlc/testdata/actual/echo5-preclaim-failed-suite.json`, SHA256 `72e41127f628c34b102cf47bc2ff480909002af16491bc9ead3f196aed2560ca`. Both live captures remain byte-identical. No Echo checkout/card/evidence was changed.

Corrected frozen source verification: focused ownership, queue, legacy-prefix and release/reclaim controls PASS (75.433s); race selector/issue-head/release-crossing controls PASS (16.971s); new native claim-straddling lifecycle/replay race control PASS (15.091s); tracked Go lint catalog/hygiene PASS (1.711s); diff check clean. README/help/atlas and the prevention lesson explain owned-only queue exclusion. Next: publish corrected source, retain one new recorded suite and run fixed-then-ship re-close; SHIP on the previous source is not final approval of this correction. Owner-run Echo proof remains a pre-merge hold.

### 2026-10-22 — corrected head integrated before its recorded suite

The re-close now falls after noon. Current main contains the #551 code landing from 11:27, so integrated that required baseline once before queuing the corrected-source suite. Git merged README, atlas and lessons automatically; no conflicts. On the integrated source, the captured ownership-queue lifecycle plus tested-issue/selector/last-commit-boundary controls PASS (53.542s), and the diff check is clean. The correction's race verification remains above; the required recorded suite on the pushed integrated source is still outstanding. No unreviewed integrated code is authorized to land.

### 2026-10-22 — full review required for the approved queue scope

Boundary round 2 (BR-1, post-close-scope-review) correctly rejected the fix-only review route: first-claim queue admission is approved added behavior, not disposal of an earlier finding. The reviewer independently passed the queue lifecycle and neighboring controls and reported no implementation defect, but the expanded contract requires a full `sdlc close`. This supersedes the revision plan's fixed-then-ship instruction. Run the full close without that flag; do not waive the finding or alter production code to dispose of a review-routing error. The corrected-source suite lease and owner-run live-proof hold remain unchanged.

### 2026-10-22 — expanded contract SHIP; corrected suite retained

Full close round 3 returned SHIP/high confidence, disposed BR-1 as addressed and reported no new findings. The reviewer independently passed ownership controls (52.673s), removed the queue fix and reproduced the exact zero-owner refusal for both native harnesses, restored it and passed queue/release-reclaim/tested-head/legacy-replay controls (62.485s), plus tracker engagement controls (2.777s). This is a full-range review of the expanded contract, not a fix-only approval. Git-verified evidence commit: `bb31782cd295b8e9f14dc09b95a273db65352f25`; codecomplete mirror: `40140c115a4922ee37f6866ea6473eeae892e42b`. The close automatically adopted measured 2.30h for its current generation; no number was typed.

Corrected-source recorded suite: pushed integrated head `c7778d78873aaa2a2e61e48ef295c6fa181df6bd`, lease `lease-3755615631`, attempt `87523898ae0044a31e8087b679815bc0`, retained terminal session 74284. Latest observed FIFO position 10. Common Git log: `.git/sdlc/test-stats/inflight/87523898ae0044a31e8087b679815bc0/run.log`. Keep this runner; do not revive cancelled lease-3451749813 or start a duplicate. Full-close session 51757 completed exit0; fix-only session 87304 completed REWORK and was superseded by this full close. Since the suite source head, only workshop review evidence, plan/Log and lesson prose changed, verified by Git's path diff.

Next: inspect this retained suite's actual result. On PASS, coordinate final stage/build with the hub for ops-echo's owner-run `sdlc actual --issue 5`, preserving the unowned-prefix warning and the claim-straddling queue behavior. Append live evidence, finish the remaining Plan hold, clear stage and enter the normal merge gate. No landing or stage has been started yet; no Echo/operator files were touched.

### 2026-10-22 — recorded suite found decimal issue-hint convention bug

The retained suite actually ran and completed: 2283 cmd/sdlc tests over eight shards plus 76 other packages, wall 920.4s. It failed deterministically only on `TestIssueHintsRenderDecimalIDs`, which names `actualownership.go`'s new claim hint interpolating raw issue ID instead of `issue.CLIRef`. Other shards/packages passed. Failed attempt `87523898ae0044a31e8087b679815bc0`, log `.git/sdlc/test-stats/failed/87523898ae0044a31e8087b679815bc0.log`, recorded log SHA prefix `9ab199a09430`. This is our formatter defect, not infrastructure or a flake. Runner 74284 finished exit1; lease-3755615631 is no longer running. Do not resume/register a flake or retry that unchanged source.

Added a pure padded-ID regression for 000005, 000008 and 000253; it and the existing source guard failed before the fix (0.517s), proving the raw zero-padded command would be unsafe for flag parsing. The formatter now uses the existing `issue.CLIRef`, preserving the reviewed recovery contract and all attribution behavior. Focused renderer/capture/padded-ID and source-guard controls PASS (0.447s), Go lint catalog/hygiene PASS (0.736s). The existing issue-hint guard already owns this mechanical mistake class; no duplicate lint mechanism is introduced. No race rerun is needed for this pure formatting change; prior ownership/queue race evidence remains above.

Next: publish the corrected source, queue exactly one fresh recorded suite, and re-close this bugfix. Expanded queue-contract approval remains recorded, while the final source still needs its own passing suite and accepted close. Owner-run Echo proof stays a pre-merge hold. Hub incident report `ce9fad07-2996-4e05-a575-1574f57bd51e` was submitted.
