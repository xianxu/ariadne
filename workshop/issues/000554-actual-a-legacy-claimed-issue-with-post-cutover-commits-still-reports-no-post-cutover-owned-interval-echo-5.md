---
id: 000554
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '028dc9e4c92e5eb3039e7ed355f3a283de4dd56c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T11:00:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "4c6861a1", done: "173e9989"}
actual_hours: 0.95
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
