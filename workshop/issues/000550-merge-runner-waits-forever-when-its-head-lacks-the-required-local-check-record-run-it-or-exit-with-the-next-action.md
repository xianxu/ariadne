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

- [ ] Preserve missing-local-check facts independently of the aggregate CI verdict.
- [ ] End retained train waits with the exact local-check action, joining readiness and releasing the ticket.
- [ ] Cover the original pair#779 capture, mixed CI verdicts, satisfied records, and interrupted wait orderings.
- [ ] Update help/README/atlas, run the recorded suite, close and land.

Durable design: [retained merge check refusal](../plans/000550-retained-merge-check-plan.md).

## Log

### 2026-10-22

- Claimed and started in ariadne:4 from landed main. Source trace confirms the cycle: `productionCarGates` calls `runPreparedCIGate` with no local runner, `carReadiness` turns its typed refusal into a one-line reason, and `vouchWhileWaiting` repeats until admission; admission itself requires a ready vouch. The admitted merge's check runner is unreachable in this state.
- Existing readiness controls pass: `go test ./cmd/sdlc -run '^(TestCarReadiness|TestEnterLandingLineVouchesWhileWaiting)$' -count=1` (5.588s). This is baseline evidence, not a fixed regression.
- Design exploration also found missing local records can be masked by pending/red/unknown server CI, no registered checks, or a missing required guard. Preserve structured missing-check evidence instead of matching the displayed aggregate reason (ARCH-DRY/ARCH-PURE).
- Authenticated pair:4 supplied `/tmp/pair779-original-readiness-capture.txt`; locally verified SHA256 `f45200e875895acaead21dc387cb79da83c83590fbb6c387b68c5efac6b501d7`. Original tool output at 02:06:34Z and 04:34:02Z on October 22 shows pending → missing affected/fallback → held on the same missing record. Historical DNS diagnostics are capture context, not a current incident. Copy this bounded text to the test fixture during implementation.

## Revisions

### 2026-10-22 — proposed exit policy; awaiting hub approval

Choose the dispatch's explicit-demand option: train readiness remains read-only, and confirmed missing local-check evidence ends this invocation with the canonical commands and retry instruction. No automatic check execution inside the retained wait. Existing admitted/serial check execution remains unchanged. Carry missing-check evidence separately from the aggregate CI reason, so concurrent server CI does not hide the missing action. Caller cancellation, successful admission and observed final train outcomes retain their existing authority.

Additional Done when clauses:

- Missing required head-scoped checks produce a nonzero exit and complete command(s), also alongside pending/red/unknown/no registered CI or a missing guard; no repeated missing-record polling.
- Affected stand-ins/full-suite fallbacks and accepted docs-only ancestor records keep their existing satisfaction rules. Landing-scoped checks are not treated as missing branch prerequisites.
- Refusal joins the readiness goroutine, ends wait evidence and releases this ticket without touching another holder. A check already queued elsewhere is left running; its existing result can satisfy a subsequent merge retry.
- Cancellation, admission, settled/merged observations, survivor reformation, and trains turning on during an existing wait have deterministic regression controls. No new persisted state or check runner is introduced.
