---
id: 000574
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a09c1d1a707c61e39ac04c22b37e0007ca7f387e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T09:02:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.80
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
- 2026-10-23: closed — Post-SHIP correction enforces the already approved read-only supplier contract. Native unchanged-content/changed-stat index-byte regression failed without the fix, then helper/supplier race checks passed 10.945s and catalog passed 1.214s. Shared observation adapter now disables optional Git writes on every query. Fresh modified=false actual sdlc/weave qualification passed refresh before current-policy ticket, recorded suite success, legacy current-versus-unknown-launch provenance, redirected-owner unavailable, frozen JSON and byte-identical supplier index; evidence common test-stats/evidence/000574/private-helper-refresh-readonly. Final pushed-head recorded branch suite is queued; it must pass before central merge. (fixed-then-ship); review verdict: SHIP
- 2026-10-23: closed — BR-1 README now documents mode, suite boundary, eligibility and skip/recovery. BR-2 native production resolver/read matrix covers distinct real supplier revisions, source ff after unchanged legacy ticket, missing/nested/redirected owner, failed Git read and helper symlink; redirected-owner regression red then green. Corrected SDLC race checks passed 13.357s, catalog 1.277s; fresh actual sdlc/weave private qualification passed refresh before current-policy ticket, recorded success, current versus unknown launch provenance, redirected-owner unavailable and frozen JSON. Evidence in common test-stats/evidence/000574/private-helper-refresh-round2. Corrected-head recorded branch suite is queued and must pass before landing; close overlaps under tl-worker rules.; review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 466 added lines in code files (limit 100); an earlier round of this close already ran the full review

Claimed and started in ariadne:3. Read the actual Spec after noting that issue show displays headings without bodies. Traced workflowFile/prepareSuiteHelper, start-plan catch-up, weave refresh, and setup lease lifetime. The proposed suite-setup boundary handles issues active across a helper rollout; a narrow weave-owned refresh avoids broad compile effects. Text status identifies each legacy ticket's current supplier, without claiming that current HEAD is the running helper's launch revision. ARCH-DRY and ARCH-ORDER shaped the seam and lifetime choices. Durable plan awaits TL approval; no implementation, suite, live checkout update or grant. Original incident TTY unavailable: only the labelled TL report of 21 siblings exists here. #533 stays parked.

Authenticated TL approval received for all choices, including narrow weave mode and private conformance plus post-land rebuild. Approval recorded in the plan Revisions; entering change-code before implementation.

Implemented the approved boundary: weave's optional mode owns the narrow private-sibling update under its existing setup lease; the shared SDLC new-suite setup calls it before capability preflight. Ordinary and baseline runs, train caller anchors, resume launch versus proof-only repost, runner exclusion, installed argv/check-only environment, old/missing executable fallback, and cancellation are exercised. Status uses a bounded per-request owner/supplier cache and labels current HEAD/dirty state separately from unknown launch revision; grant compatibility and frozen JSON stay unchanged. Extracted the existing workflow resolver into pkg/workspace and shared the actual helper-symlink resolution. A new native red control demonstrated that refreshing the workflow directory alone could update a checkout that did not supply the actual helper; following the helper symlink makes it skip safely.

Verification so far: initial CLI red was unknown --suite-helper; preflight red ran the stale helper before refresh; status red lacked supplier revision. Green broader SDLC helper/queue/grant/catch-up controls: 77.666s. New SDLC helper/supplier race checks: 8.064s. Weave helper/refresh race checks: refresh package 14.104s and weave 32.538s; shared resolver race 1.261s. Full pkg/workspace: 36.253s. Go lint catalog: 1.479s. Counted status tests cover 8/16 repeated owners, shared suppliers and distinct suppliers. Private actual-binary qualification, recorded branch suite, close review and landing remain. No live source supplier or queue/grant was changed.

Private actual-binary conformance PASS: own sdlc and weave built with Go 1.27.3 at the committed implementation, both modified=false. Real local-check run used a disposable numbered Git environment and fake publication transport: actual weave fast-forwarded the old-policy supplier before actual Python allocated a fifo-grant-aging-v1 ticket, then the private recorded suite succeeded. A held synthetic old-format ticket showed current supplier full HEAD and launch revision unknown; frozen schema_version=1 JSON stayed unchanged. No live queue/grant, primary checkout, external VM or real GitHub publication was exercised. Reproducer, raw command/gh/suite/status logs, result and build metadata retained under common Git sdlc/test-stats/evidence/000574/private-helper-refresh; owned temporary root was removed after copying evidence. Central CLI remains the lifecycle authority. Proceeding to pushed branch suite and close; code exceeds the quick shell, so close must use the full review.

Close round 1: FIX-THEN-SHIP with Important BR-1 (README new-flag coverage) and BR-2 (native provenance controls missing despite the plan checkmark); no final close/card publication. Accepted both after inspecting the actual files. Canceled only own still-queued lease-2648555812 through local-check cancel; it never acquired suite capacity and two other checkouts were left running. Added README coverage and the real Git-backed production resolver/read matrix. RED: a redirected owner symlink manufactured current supplier proof; exact owner-path identity now yields unavailable. GREEN: all supplier tests 9.663s, including distinct real revisions, post-ticket fast-forward without changing ticket bytes, missing/nested/redirected owner, failed supplier HEAD read, actual helper symlink, dirty state, deadline and counted cache controls. Plan Revisions records the evidence correction; race, fresh private proof, corrected-head suite and round 2 remain.

Corrected-head verification: SDLC helper/supplier race tests PASS 13.357s; full Go lint catalog PASS 1.277s. Rebuilt private sdlc/weave at the committed correction (Go 1.27.3, modified=false), then repeated the actual binary qualification successfully, adding a held synthetic redirected-owner ticket that reports revision unavailable and launch revision unknown. Raw evidence and reproducible script retained under common test-stats/evidence/000574/private-helper-refresh-round2; owned temporary root removed. README inspection confirms the new mode, automatic boundary, eligibility and skip/recovery behavior. Requeue the recorded branch suite on this corrected head and run full close round 2; no unrelated check, lease or queue order was changed.

Full close round 2 SHIP: BR-1 and BR-2 addressed, zero new findings; reviewer independently removed the exact-owner guard via a temporary overlay and reproduced the redirected-owner regression, then removed the overlay. Card is codecomplete. Before landing, an additional private read-only check proved Git status can rewrite the supplier's index cache for unchanged contents with changed stat metadata. Native TestSuiteQueueSupplierReadDoesNotRewriteIndexCache RED reproduced that write. Fixed the shared observation adapter to use --no-optional-locks for every query; new/helper/supplier race tests GREEN 10.945s and full catalog GREEN 1.214s. Canceled only own still-queued lease-3030667962 before any suite capacity; three other checkouts were left running. Plan Revisions records this enforcement of the existing read-only contract; no admission or JSON policy change. Fresh private proof, final recorded suite and mandatory post-close fix review remain.

Read-only correction private qualification PASS: rebuilt own sdlc/weave with modified=false and repeated the actual local-check refresh/marked-ticket/recorded-success, legacy provenance, redirected-owner unavailable and frozen-JSON controls. Added a real queue-status assertion that changed stat metadata with unchanged helper contents leaves supplier index bytes identical. Raw result records index_unchanged=true; evidence retained at common test-stats/evidence/000574/private-helper-refresh-readonly and owned temporary root removed. Required next steps: pushed final-head suite, post-close fix review, then central merge. There are no running private proof processes.

2026-10-23 — Final-head recorded suite passed 2,410 cmd/sdlc tests and 76 other packages with zero failures in 17m45s. The retained central merge waited behind #573, #572 and the operator priority grant for #575, then exited on integration conflicts. Recovery merges current origin/main and preserves both additive fields: refreshSuiteHelper plus suiteCapabilities in localRunDeps, and supplier diagnostics plus the light-lane observation in queue status. No policy change. Targeted combined checks, a fresh recorded full suite, mandatory fixed-then-ship review and central merge re-entry follow; #578 remains next after landing.
