---
id: 000364
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '5bccab1db5d8ec12359a52c74bf54a9a4d2278c0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T23:48:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Boundary review environment: allow read-only process listing, prepare the checkout (declared prep step), and REWORK must cite an answerable finding

## Problem

pair#528 M1's mandatory boundary review returned REWORK ten rounds running with zero open findings (2026-10-13). The reviewer stated the revised M1 contract was implemented, but would not pass it without verifying process behaviour itself, and couldn't:
1. its sandbox denies `/bin/ps` (the change and its tests inspect live processes);
2. its disposable checkout lacks generated, untracked build outputs (pair's `cmd/internal/runtimebundle/assets/runtime/files`), so some packages don't build or test there.

Host evidence (full suites, the blocked checks) was supplied in the Log, but a REWORK with no finding ID can't be disputed (`sdlc finding dispute` refuses never-issued IDs), so the author had no action: a dead end. The operator approved a `--no-judge` close for M1 as a one-off.

## Spec

1. **Read-only process listing in the review sandbox:** retain the existing protected-source write fence and add the literal Darwin `/bin/ps` execution exception measured below. Do not introduce a broader confinement model or broaden reviewer permissions generally.
2. **Prepared checkout:** a repo declares a preparation step (next to `.sdlc/local-checks.json`, e.g. `prepare` running `make build` or its asset generation); the review harness runs it in the disposable checkout before the reviewer starts, and records its result. Absent a declaration, behaviour is unchanged.
3. **REWORK must cite an answerable finding:** a verdict that blocks must carry at least one finding ID. "Could not verify X" becomes a named finding (kind: unverifiable) that the author can answer by fixing the environment or by `sdlc finding dispute` with host evidence, and the next round must rule on that evidence. A REWORK with no findings is rejected by the harness and re-run (bounded) rather than recorded.

## Done when

- A review of a change whose tests use `ps` can run them in the sandbox.
- With pair's declared prepare step, the reviewer's checkout builds and tests the runtimebundle packages.
- A reviewer response of REWORK with no findings is refused by the harness; an "unverifiable" finding can be disputed, and the next round answers it (tests for each).

## Plan

- [ ] Add and prove the literal Darwin ps exception without weakening protected writes.
- [ ] Model unverifiable findings and validate answerable boundary responses.
- [ ] Prepare every disposable-checkout attempt and reuse bounded retry admission.
- [ ] Document, verify with pair's runtime generation, run the full suite, close and land.

Durable plan: [review environment](../plans/000364-review-environment-plan.md).

## Log

### 2026-10-13

### 2026-10-14 — design and measured root cause

Read pair#528 M1's review sidecar: blocked rounds were marked fenced. Current fenced
Codex already launches with danger-full-access beneath the outer fence. Direct ps
works; even an allow-default Seatbelt profile refuses exec of this host's setuid
/bin/ps. Adding `(allow process-exec (literal "/bin/ps") (with no-sandbox))` works.
A temporary fixture proved ps and scratch writes succeed while a protected write
remains denied. The reviewer-agent-only hypothesis was disproved, not assumed.

Hub approved the measured narrow rule, declared prep using pair's
`make runtimebundle-generate`, and answerable REWORK. Authorization includes writing
the plan then proceeding through change-code without another approval. Pair#547
separately replaces pair's ps subprocesses. No implementation changes yet.

### 2026-10-14 — plan review and historical corroboration

Plan-quality round 1 raised PQ-1 (function-level test strategies) and PQ-2
(diagnostic retention). The durable plan now names each risky function and its
mechanical test strategy, and makes permanent retention plus per-dispatch cost
explicit. No implementation started before the gate clears.

pair:2 located the raw round-10 reviewer session
`01a13908-b083-7751-a0f4-4345bf6222ff`: its turn context explicitly records
`danger-full-access` and approval policy `never`; direct shell/test execution still
reported `/bin/ps` EPERM. This corroborates the measured outer Seatbelt failure,
without claiming that the transcript identifies the historical outer argv.

## Revisions

- 2026-10-14 — Spec 1 corrected with hub approval: the existing fence denies source
  writes, not arbitrary signals/network/outside writes. Preserve that boundary and
  add only the measured literal ps exception; a general confinement redesign is out
  of scope. Evidence and approved execution direction are recorded above.
