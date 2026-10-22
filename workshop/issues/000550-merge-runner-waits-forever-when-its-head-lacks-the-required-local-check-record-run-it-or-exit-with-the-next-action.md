---
id: 000550
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '44ae48f71e5b3ac2aab550cf2e2a3598c205592d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T07:53:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "b3e89876", done: "68b71c82"}
---

# merge runner waits forever when its head lacks the required local-check record; run it or exit with the next action

## Problem

## Spec

pair#779 (10-21 19:02 → 10-22 05:00): its retained `sdlc merge --yes` runner sat first in pair's landing line for ten
hours, repeatedly logging "train due; not ready: PR535 needs recorded local checks on its head: affected or its
full-suite fallback", while three later changes landed past it. Nobody acted: the slot believed admission would run
the check, and the slot was idle (its turn had ended), so the repeated message was never read. Train readiness
excludes a car with no recorded local check **before** admission, so the runner can never make progress by waiting.

Fix: when the only thing keeping a retained merge runner from readiness is a missing local-check record on the exact
pushed head, the runner starts that check itself (the same `sdlc local-check run` under the normal queue), records it,
and proceeds; or, if policy forbids that, it exits with a clear next action rather than waiting forever. Either way
it doesn't loop silently.

## Done when

- A merge runner whose head lacks the required local-check record runs (or clearly demands) it instead of waiting
  indefinitely (test with a fake queue/check runner).
- A head that already has the record is unaffected (test).

## Plan

- [x] Preserve missing-local-check facts independently of the aggregate CI verdict.
- [x] End retained train waits with the exact local-check action, joining readiness and releasing the ticket.
- [x] Cover the original pair#779 capture, mixed CI verdicts, satisfied records, and interrupted wait orderings.
- [x] Update help/README/atlas and complete focused verification.

Close and landing follow the verification checkpoint through their existing SDLC verbs.
The recorded full suite must pass before landing; its retained queue may overlap close review under the worker rules.

Durable design: [retained merge check refusal](../plans/000550-retained-merge-check-plan.md).

## Log

### 2026-10-22

- Claimed and started in ariadne:4 from landed main. Source trace confirms the cycle: `productionCarGates` calls `runPreparedCIGate` with no local runner, `carReadiness` turns its typed refusal into a one-line reason, and `vouchWhileWaiting` repeats until admission; admission itself requires a ready vouch. The admitted merge's check runner is unreachable in this state.
- Existing readiness controls pass: `go test ./cmd/sdlc -run '^(TestCarReadiness|TestEnterLandingLineVouchesWhileWaiting)$' -count=1` (5.588s). This is baseline evidence, not a fixed regression.
- Design exploration also found missing local records can be masked by pending/red/unknown server CI, no registered checks, or a missing required guard. Preserve structured missing-check evidence instead of matching the displayed aggregate reason (ARCH-DRY/ARCH-PURE).
- Authenticated pair:4 supplied `/tmp/pair779-original-readiness-capture.txt`; locally verified SHA256 `f45200e875895acaead21dc387cb79da83c83590fbb6c387b68c5efac6b501d7`. Original tool output at 02:06:34Z and 04:34:02Z on October 22 shows pending → missing affected/fallback → held on the same missing record. Historical DNS diagnostics are capture context, not a current incident. Copy this bounded text to the test fixture during implementation.

### 2026-10-22 — implementation and focused evidence

Change-code admitted the inferred quick flow. Implemented independent `ciGateError.LocalNeeded`, shared canonical command formatting, the third readiness result and scoped cancel-cause propagation. Successful admission/observed settlement/caller cancellation retain their authority, and a survivor refuses before releasing the old conductor. No automatic check runner or durable state added. Help, README and both atlas descriptions distinguish pre-admission refusal from admitted automatic execution.

Red: the original waiter called readiness twice then returned `context canceled`, instead of the actionable refusal, for both Claude/Codex labels. CI metadata tests independently failed across the masked-verdict inventory. Green: focused CI classifier/gate controls PASS1.474s; production readiness matrix PASS23.718s; full `runMerge` entry with a held ticket PASS2.565s; new regression/production/race controls PASS35.452s, and full merge-entry race control PASS4.322s after replacing its shared buffer with the existing synchronized writer. Tracked-file Go lint PASS1.059s.

The first combined broad run's assertions passed, but its hermeticity guard correctly rejected concurrent agent edits to two docs and a test file (174.212s); this is contaminated verification, not a suite pass or product/flake evidence. Freeze edits for its replacement focused run. A retry fixture also initially reused its missing-record second-poll cancellation guard after the gate passed: removed that artificial scheduling assumption from the retry path; the refusal still asserts exactly one observation. No production timing budget was changed.

## Revisions

### 2026-10-22 — proposed exit policy; awaiting hub approval

Choose the dispatch's explicit-demand option: train readiness remains read-only, and confirmed missing local-check evidence ends this invocation with the canonical commands and retry instruction. No automatic check execution inside the retained wait. Existing admitted/serial check execution remains unchanged. Carry missing-check evidence separately from the aggregate CI reason, so concurrent server CI does not hide the missing action. Caller cancellation, successful admission and observed final train outcomes retain their existing authority.

Additional Done when clauses:

- Missing required head-scoped checks produce a nonzero exit and complete command(s), also alongside pending/red/unknown/no registered CI or a missing guard; no repeated missing-record polling.
- Affected stand-ins/full-suite fallbacks and accepted docs-only ancestor records keep their existing satisfaction rules. Landing-scoped checks are not treated as missing branch prerequisites.
- Refusal joins the readiness goroutine, ends wait evidence and releases this ticket without touching another holder. A check already queued elsewhere is left running; its existing result can satisfy a subsequent merge retry.
- Cancellation, admission, settled/merged observations, survivor reformation, and trains turning on during an existing wait have deterministic regression controls. No new persisted state or check runner is introduced.

### 2026-10-22 — hub approval and actionable exit wording

Authenticated ops:0 receipt `4cec01c2-b2f4-4df2-9346-718e9a3e4c55` approves this plan and fixture plus real-lock verification; no stage. The exit must say plainly `merge stopped: run <exact command>, then re-enter sdlc merge`, so an idle slot's pane gives the TL a single next action. This is an additional Done when requirement; implementation must preserve the canonical commands and all approved cancellation/cleanup semantics.

### 2026-10-22 — clean focused verification

With all repository edits frozen, the combined readiness/missing-check/production-entry/wait-lifecycle/CI controls PASS136.615s; the staged-file Go lint catalog PASS1.077s. This replaces the contaminated broad run. Both targeted race runs passed as recorded above. Default machine cache verified as `/Users/xianxu/Library/Caches/go-build`; no overrides or private cache. Next: push the committed final build, queue the recorded full suite, close and enter merge on SHIP.

### 2026-10-22 — verification checklist wording

The last Plan checkbox originally grouped documentation, suite, close and landing. It now ends at required suite verification; close and landing remain subsequent gated lifecycle steps, so checking implementation complete cannot falsely claim that this issue has already landed. No scope or behavior change.

### 2026-10-22 — retained suite and permitted close overlap

The required recorded suite is queued on pushed `789b21c045282b828ee70d702ff286b717afab11`, PR398: lease `lease-832394143`, attempt `b518b1c09fb83e30ce69c8b8395966f6`, terminal29415, initial normal FIFO position7. Preserve this runner and its place. The source code is unchanged while it runs. The final issue Plan checkbox now names completed focused verification; the still-required full-suite pass remains the explicit pre-landing condition, permitting the documented tl-worker close-review overlap without claiming a pass. No bypass or stage; ops:0 approved the latter. Close may review now, but any code fix needs a fresh recorded head.
