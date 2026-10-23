---
id: 000574
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '73723ceef72cb4276a425b9bca048a4ddf0f6cf7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T09:02:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "48209606", done: "ed3a978f"}
---

# Keep slot-sibling ariadne checkouts current so queue-helper rollouts finish

## Problem

## Spec

**Slot-sibling ariadne checkouts drift, so queue-helper rollouts never finish (TL, 10-23).** Each non-ariadne slot has
a sibling `worktree/<repo>-slotN/ariadne` checkout on `main` that supplies its queue helper (`test-lock.py`, kept
checkout-supplied by ariadne#563). Nothing fast-forwards these: on 10-23 they ranged from 10-21 to 10-23. After
ariadne#570 landed, every waiter from a stale sibling counted as a legacy-policy waiter, so `sdlc landing grant` kept
refusing (correctly) with no end in sight; the TL fast-forwarded all 21 clean siblings by hand at 08:2x.

Fix: keep sibling ariadne checkouts current at a natural boundary (for example `sdlc claim`/`start-plan`, or
local-check before taking a ticket, fast-forwards a clean sibling on `main` to `origin/main`, never touching a dirty or
non-main one, and says so), and have queue status name the stale sibling behind a legacy ticket.

## Done when

- A clean sibling ariadne on `main` behind `origin/main` is fast-forwarded at the chosen boundary; dirty or off-main
  siblings are left alone with one line naming them (test).
- Queue status's mixed-policy line names the supplying checkout's revision for each legacy ticket.

## Plan

- [x] Implement and verify the [supplier-refresh plan](../plans/000574-sibling-helper-refresh-plan.md) after TL approval.

## Log

### 2026-10-23

Claimed and started in ariadne:3. Read the actual Spec after noting that issue show displays headings without bodies. Traced workflowFile/prepareSuiteHelper, start-plan catch-up, weave refresh, and setup lease lifetime. The proposed suite-setup boundary handles issues active across a helper rollout; a narrow weave-owned refresh avoids broad compile effects. Text status identifies each legacy ticket's current supplier, without claiming that current HEAD is the running helper's launch revision. ARCH-DRY and ARCH-ORDER shaped the seam and lifetime choices. Durable plan awaits TL approval; no implementation, suite, live checkout update or grant. Original incident TTY unavailable: only the labelled TL report of 21 siblings exists here. #533 stays parked.

Authenticated TL approval received for all choices, including narrow weave mode and private conformance plus post-land rebuild. Approval recorded in the plan Revisions; entering change-code before implementation.

Implemented the approved boundary: weave's optional mode owns the narrow private-sibling update under its existing setup lease; the shared SDLC new-suite setup calls it before capability preflight. Ordinary and baseline runs, train caller anchors, resume launch versus proof-only repost, runner exclusion, installed argv/check-only environment, old/missing executable fallback, and cancellation are exercised. Status uses a bounded per-request owner/supplier cache and labels current HEAD/dirty state separately from unknown launch revision; grant compatibility and frozen JSON stay unchanged. Extracted the existing workflow resolver into pkg/workspace and shared the actual helper-symlink resolution. A new native red control demonstrated that refreshing the workflow directory alone could update a checkout that did not supply the actual helper; following the helper symlink makes it skip safely.

Verification so far: initial CLI red was unknown --suite-helper; preflight red ran the stale helper before refresh; status red lacked supplier revision. Green broader SDLC helper/queue/grant/catch-up controls: 77.666s. New SDLC helper/supplier race checks: 8.064s. Weave helper/refresh race checks: refresh package 14.104s and weave 32.538s; shared resolver race 1.261s. Full pkg/workspace: 36.253s. Go lint catalog: 1.479s. Counted status tests cover 8/16 repeated owners, shared suppliers and distinct suppliers. Private actual-binary qualification, recorded branch suite, close review and landing remain. No live source supplier or queue/grant was changed.

Private actual-binary conformance PASS: own sdlc and weave built with Go 1.27.3 at the committed implementation, both modified=false. Real local-check run used a disposable numbered Git environment and fake publication transport: actual weave fast-forwarded the old-policy supplier before actual Python allocated a fifo-grant-aging-v1 ticket, then the private recorded suite succeeded. A held synthetic old-format ticket showed current supplier full HEAD and launch revision unknown; frozen schema_version=1 JSON stayed unchanged. No live queue/grant, primary checkout, external VM or real GitHub publication was exercised. Reproducer, raw command/gh/suite/status logs, result and build metadata retained under common Git sdlc/test-stats/evidence/000574/private-helper-refresh; owned temporary root was removed after copying evidence. Central CLI remains the lifecycle authority. Proceeding to pushed branch suite and close; code exceeds the quick shell, so close must use the full review.

Close round 1: FIX-THEN-SHIP with Important BR-1 (README new-flag coverage) and BR-2 (native provenance controls missing despite the plan checkmark); no final close/card publication. Accepted both after inspecting the actual files. Canceled only own still-queued lease-2648555812 through local-check cancel; it never acquired suite capacity and two other checkouts were left running. Added README coverage and the real Git-backed production resolver/read matrix. RED: a redirected owner symlink manufactured current supplier proof; exact owner-path identity now yields unavailable. GREEN: all supplier tests 9.663s, including distinct real revisions, post-ticket fast-forward without changing ticket bytes, missing/nested/redirected owner, failed supplier HEAD read, actual helper symlink, dirty state, deadline and counted cache controls. Plan Revisions records the evidence correction; race, fresh private proof, corrected-head suite and round 2 remain.
