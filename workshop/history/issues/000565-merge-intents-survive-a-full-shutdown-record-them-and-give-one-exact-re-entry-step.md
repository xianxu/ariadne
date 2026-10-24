---
id: 000565
status: done
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'fa4f21966f2df2b0493761b8f2ffcd9aa92b41f8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T10:46:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.92
---

# Merge intents survive a full shutdown: record them and give one exact re-entry step

## Problem

## Spec

From the 10-22 full-shutdown recovery: **merge runners don't survive a full Couch shutdown,** so every landing in
flight had to be re-entered by hand afterwards (each slot then also had to decide whether the shutdown freeze still
applied). Make re-entry automatic: on restart, a slot whose issue has a durable merge intent (an admitted merge or a
landing ticket recorded before the shutdown, and no live runner) gets a single clear next step from `sdlc` (or the
worker's restart prompt) to re-enter `sdlc merge` on the same head, without duplicating suites that already passed.
No automatic merge without the slot re-entering it; just a durable record and an exact recovery command.

## Done when

- After a simulated full shutdown with an admitted merge, `sdlc landing status`/the worker restart summary names the
  orphaned merge intent and the exact re-entry command (test).
- Re-entering reuses a passed recorded suite on an unchanged head (test).

## Plan

Durable plan: `workshop/plans/000565-merge-intents-survive-a-full-shutdown-record-them-and-give-one-exact-re-entry-step-plan.md` (scope choices a–e for the TL).

- [x] `MergeIntent` record and `ClassifyIntent` (pure)
- [x] `IntentStore`: hold, release by outcome, list, stale and age sweep
- [x] `sdlc merge` holds the intent; an interruption keeps it, a terminal outcome removes it
- [x] Surfaces: `sdlc landing status`, `sdlc state`, `--json` `merge_intents`
- [x] End-to-end: a SIGKILLed runner, the printed command, the suite reused on an unchanged head
- [x] Atlas, recovery catalog, recorded suite

## Log

### 2026-10-23
- 2026-10-23: closed — post-close fix a714e4d6 removes the unused ctx parameter of observeMergeIntents, which TestVerbContextsReachTrackerReads refused in the recorded suite on ffd53cf0; recorded suite sdlc-local/suite passed on a714e4d6 (exit 0 in 18m38s, log sha256:9727c73162e9); origin/main integrated at 0455eba5, builds and vets clean; the suite reruns on the integrated head before merge (fixed-then-ship); review verdict: SHIP
- 2026-10-23: closed — Round-2 Minors fixed in 5f609079: first intent linked into place locked and complete, dead temps swept (TestIntentCreationLeavesNoWindowOrResidue); shared-directory path quoted (TestReportMergeIntent, TestIntentHoldNeverOverwritesAnUnreadableRecord); plan Revisions record entity names. landingline intent tests and cmd/sdlc merge-intent tests pass. (fixed-then-ship); review verdict: FIX-THEN-SHIP
- 2026-10-23: closed — Round-1 Important fixed as a class: readers share a lock, the sweep decides unlocked and removes only an unchanged record, Hold waits out brief locks (TestIntentReadersAreNotOwners, mutation-verified); Minors fixed (TestIntentHoldNeverOverwritesAnUnreadableRecord, quoted fields, one branch-head helper). Done-when 1: TestOrphanedMergeOnStatusSurfaces, TestIntentOfAKilledRunnerIsOrphaned. Done-when 2: TestMergeReentryReusesThePassedSuite. landingline/queuegate/recovery packages pass; cmd/sdlc Landing|Merge|State|Intent|Train|Queue pass except two sandbox-only failures identical on main. Recorded suite runs on the final head before merge.; review verdict: FIX-THEN-SHIP
- 2026-10-23: flow upgraded quick → full — 626 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Claimed and planned (ariadne:5). What exists today:
  - Nothing on the landing line survives a shutdown: tickets and head.lock are flock-only.
  - The queue op survives as `unresolved`, but it has no branch and a reboot erases it.
  - The recovery command goes only to the dead runner's stderr.
  - Re-entry is already correct (it resumes from Git/GitHub), and #332 suite reuse reads GitHub statuses, so it survives a shutdown.
  - The gap is a durable, lock-classified intent record plus a surface that prints the command.
- TL approved scope choices (a)–(e), plus the stale sweep and the 30-day bound. After this lands, file the Couch side (recover-plan reading `merge_intents`) as an unclaimed pair issue.
- Implemented in five commits: the record, the merge hold, the surfaces, re-entry suite reuse, and the atlas/catalog.
  - The intent tests include a SIGKILLed helper process holding an intent, which reads as orphaned afterwards.
  - Mutation check: making release always remove the intent fails the lifecycle test and both reuse cases.
  - Suite reuse is end to end: the context is cancelled right after the suite records a pass, re-entry lands with 1 suite run, and it reruns (2 runs) when main moved.
  - Deviations are recorded in the plan's Revisions.
- Close review round 1, FIX-THEN-SHIP.
  - Important: readers took the owner's exclusive try-lock, so a re-entry during a status read or sweep refused as "already running", naming the dead pid. The fix covers the class, a reader's lock never reading as an owner:
    - Readers take a shared lock, so they no longer see each other as live.
    - The sweep decides from git with no lock held, then removes under a momentary exclusive lock, and only if the record still names the same run.
    - `Hold` waits up to 2 s before refusing.
    - Regression tests use an event hook, not a sleep. The test is mutation-checked: switching readers back to the exclusive lock fails it.
  - Minors fixed:
    - `Hold` refuses (naming the path) rather than overwriting an unreadable or later-version record.
    - Root, pause owner and pause reason are printed `%q`-quoted.
    - One `localBranchHead` helper replaces three copies of the lookup.
- Close review round 2, FIX-THEN-SHIP with no blocking findings. Three Minors, all fixed:
  - The unreadable record's path is now quoted, so every string from the shared directory prints `%q`.
  - A first record is linked into place already locked and complete, which closes the window in which a reader found an empty file and advised removing it. The sweep collects temp files whose writer died.
  - Plan Revisions now record the entity names as built.
- Close review round 3 (fixes only), FIX-THEN-SHIP with no blocking findings. Two Minors, both fixed:
  - The sweep could remove a running merge's temp record between its creation and its lock; that was the third finding in the reader-lock family. Under the rule "never judge an owner dead while its file can exist unlocked", the sweep skips temp records younger than the hold patience, and `create` retries when a sweep wins the race. The fresh-temp test is mutation-checked.
  - BR-7: the creation window is now pinned by a seam just before the link. At that point a reader sees no intent, and a sweep leaves the locked temp alone.

### 2026-10-22
