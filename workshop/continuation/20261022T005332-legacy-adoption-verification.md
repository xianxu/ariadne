---
type: continuation
slug: legacy-adoption-verification
agent: codex
session_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
created: 2026-10-22T00:53:32
branch: 000544-issue-migrate-drops-a-legacy-close-s-completion-record-later-issues-actual-measurement-then-refuses-echo-1-echo-3
worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
issues: [000544]
---

# Continuation: legacy-adoption-verification

## NEXT ACTION

Run stable combined targeted verification for #544 on the committed implementation, with all bounded agents finished and no concurrent source edits. Start with `go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/tracker ./cmd/sdlc/internal/observe -count=1`, then a selected main-package run covering legacy migration, repair, proof, exclusion, real actual measurement and issue show. Read the issue Log and durable plan first. All new files are now tracked, so run the Go lint catalog too; prior untracked-file lint was not sufficient. Persist logs under the common Git test-stats evidence/000544 directory, not the RAM disk. No recorded check is queued or running and no retained terminal needs preserving.

After targeted/race/vet verification, build an exact committed binary and coordinate directly with ops-echo:0 for its approved temporary-worktree dry-run, digest-pinned apply, no-op repeat and `sdlc actual --issue 3`. The Echo owner performs the writes; our slot must not mutate Echo. Then push/open PR, run the recorded full suite on the pushed issue head, close with the binary-owned fresh review and measured actual, and enter `sdlc merge --yes` promptly on SHIP. A branch suite pass does not cover integrated code; let the landing runner own integration. Report the landed PR or blocker to hub.

## State of play

#540 is landed as PR383 and already reported to ops:0 with a submitted receipt. Do not redo it. #544 is claimed by ariadne:1, full-flow change-code admitted, implementation checkpoint committed but not yet reviewed or fully suited. Run `sdlc state` to recover exact branch state. No #544 PR, full recorded suite or live Echo check has been done. All bounded implementation agents finished and committed nothing; root integrated and committed their shared edits.

## Thread arc & user model

The operator wants historical accounting restored without weakening modern evidence. TL approved one shared immutable close verifier and explicit digest-pinned repair. Existing born-archived records forced a deliberate distinction: keep additive adoption provenance instead of inventing a close, and allow a proven terminal bound only when it strictly predates the earliest admitted engagement. The final TL addition explicitly covers already-migrated unbound cards at the real tracker root, with no card write or fake completion. The user values progress through landing, precise evidence and owner-controlled live recovery over progress chatter or repeated permission requests.

## Artifact map

Read workshop/issues/000544-issue-migrate-drops-a-legacy-close-s-completion-record-later-issues-actual-measurement-then-refuses-echo-1-echo-3.md for approvals, current Done when, revisions, targeted red/green evidence and remaining verification. The durable plan is workshop/plans/000544-legacy-completion-plan.md; its gate artifact records clean plan-quality after naming risky functions. Estimate3.57h was derived after admission; actual must be measured, not typed. The hub owns the project file; use --no-project at close and send outcome rather than editing it.

cmd/sdlc/internal/issue/legacyadoption.go owns the additive codec (retained/windowless-close/windowless-terminal); completion and adoption are separate. internal/tracker/legacycompletion.go selects final terminal-generation close evidence, and legacyrepair.go provides the pure replacement/digest decision. Main legacycompletion.go shares immutable history proof across migration, repair, exclusion and read-only old-root recovery; its sibling test file covers cancellation, truncation, branch ambiguity, quotes and correction/modern controls. Main issuemigrate_repair.go resolves cutover/source/root authority and publishes one card through existing CAS, retaining lost-reply recovery. Main actualexclusion.go gates strict-before against Snapshot.Since. observe.go and internal/observe display stored or recovered class without fabricating a completion. README, helptext and atlas/workflow/issue-tracker-migration.md describe recovery.

The real Echo checkout is /Users/xianxu/workspace/echo; its :0 is busy with unrelated work. AGENTS.local.md was read and MEMORY.md absent. We inspected only the referenced issue history read-only. Echo#1 has legitimate older closes: select newest valid close of final terminal generation, do not reject merely because history contains several ancestor closes. Its latest close has valid SHIP/review evidence but no retained ActualInputs, so windowless-close is expected. ops-echo:0 runs conformance itself in a disposable worktree off origin/main; tell that exact slot when the built binary is ready.

## Decisions & dead ends

Completion-only metadata cannot represent born-archived records honestly; blanket ignoring missing completions would hide modern corruption. The approved additive legacy_adoption record covers both, with positive root/source proof. Old read-only fallback is narrower: canonical archived terminal-only generation, done numeric card present at the actual tracker root, matching immutable pre-cutover fields, no correction, completion or explicit adoption. Active-done source support remains explicit migration-only. Modern missing records and malformed/overlapping evidence stay strict. Source-card matching uses true migration output in fixtures, not synthetic card formatting that never existed in production.

## Lessons learned

Current targeted evidence is in the issue Log: production actual six-case fixture passed29.569s, final resolver/proof controls19.058s, real repair CAS controls11.581s, observer integration2.516s and observe package0.161s. Earlier concurrent source introductions tripped the hermeticity guard; stable reruns passed, so this is not registered as a flake. Preserve all future cited raw logs on SSD. Tests exist for design discoveries, including old-root terminal cards and parked earlier engagement. Review still needs to challenge authority/cancellation/CAS and operating-cost boundaries; do not claim overall completion from targeted checks.
