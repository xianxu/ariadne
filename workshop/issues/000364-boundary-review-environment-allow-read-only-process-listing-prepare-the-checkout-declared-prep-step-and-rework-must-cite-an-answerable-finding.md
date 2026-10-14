---
id: 000364
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours: 3.608
card_mirror: '20895a41ea90cb4a6d7efec93a5220f13605820d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T23:48:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

- [x] Add and prove the literal Darwin ps exception without weakening protected writes.
- [x] Model unverifiable findings and validate answerable boundary responses.
- [x] Prepare every disposable-checkout attempt and reuse bounded retry admission.
- [x] Document and verify with pair's runtime generation and the full suite.

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

### 2026-10-14 — implementation checks

Change-code full flow passed (plan-quality CLEAN; estimate-quality INFO). The
estimate includes integration design for concrete snapshot/sidecar threading in
existing lock paths, and verification/ship work in integration plus boundary review;
docs are small extensions to existing pages. Its implementation allowance may be
tight, as the advisory review noted; actuals will be measured rather than adjusted
to fit the estimate.

Literal ps rule: live Darwin direct and shell-child known-PID checks pass, as do
scratch writes and protected-write refusal. Removing the rule reproduced EPERM
and failed the conformance test; restored package passes. No Linux behavior changed.
Preparation/retry tests first failed for missing assets, unchecked failures and
accepted empty REWORK, then passed after lifecycle integration. A boundary test
proves both invalid attempts are retained as diagnostics without an admitted round.
Parser/finding fuzzing and four admission guard mutations passed their checks.
Combined judge/reviewprepare/reviewbox/gatestate/vocab tests passed. Full suite and final boundary review remain pending.

Live downstream proof: pair origin/main `b63f4bd76e3e81d1433cfeda544bb3bed8501522`
failed its runtimebundle test before generation. In a disposable clone with a
tracked declaration, production fenced preparation took 655.738 ms and the real
runtimebundle test passed (0.181 s); source HEAD/refs/status/diff stayed unchanged.
Installed Codex v0.162.1, gpt-6.1-sol, approval never, danger-full-access beneath
Seatbelt executed `/bin/ps -p 1 -o pid=,comm=` and returned PID 1 `/sbin/launchd`.
Its SHIP/findings handoff passed the boundary validator; the live smoke passed
in 9.385 s. Both credential/platform-sensitive checks are opt-in tests, skipped
by default. Linux code is unchanged; no live Linux runner is available locally.
The hub was asked to assign pair's real declaration separately, with operator-held
verification for its `.sdlc` configuration change.

Full-suite attempt 1 at `ee488c91982297fa4b34ddaa01bb5c00013f94d5` waited
449 s and ran 501.1 s, failing. Remaining raw SHIP/REWORK test fixtures needed
valid findings/detail. A concurrency fixture's empty findings fence triggered
retry and a double-close panic, aborting its shard (the 12 missing-test report).
Corrected fixtures pass targeted tests, including all 14 concurrency schedules
(41.398 s). A production regression was also caught: dispatch errors must never
persist semantic review text. The sidecar now retains only bounded diagnostics
for failures; cancellation/no-evidence launch failures retain no artifact. The
interruption and four-call diagnostic retention tests pass (10.745 s for retention).
Weave's unrelated setup-lease test failed once in the suite but passed 20 isolated
repetitions; the next full suite retains raw JSON so any recurrence has diagnostics.

A pre-restart probe showed Git textconv can reduce a real tracked-content diff
to zero bytes. Cancelled the still-queued second run before any tests started.
Added a red regression (reviewer incorrectly launched after a hidden mutation),
then disabled textconv alongside external diffs in the tracked-state comparison.
All preparation tests pass (11.523 s). Restarting the full run on this corrected
commit, with raw JSON retained.

Full suite PASS at `74bd16708c717eeafdcf599ad9842180c09c6ed4`:
`make -k test ARGS='--json /tmp/ariadne-364-full-suite-2.jsonl'` waited 70 s,
then ran 449.2 s; all 12 shards over 1,249 SDLC tests plus 53 other packages
exited zero, with no missing tests or failures. Log:
`/tmp/ariadne-364-full-suite-2.log`. The weave lease failure did not recur.
Implementation acceptance evidence is complete; close review and landing are next.

Boundary review round 1 REWORK: BR-1 Critical, `review-input-integrity`.
The reviewer reproduced a preparation command using assume-unchanged to change
tracked bytes without detection and identified clean-filter normalization as the
same class. Accepted; the guard must fingerprint actual filesystem contents,
types and modes independently of Git's normalized diff. Durable plan revised;
implementation acceptance reopened for this correction. Other review checks passed;
live Seatbelt checks skipped under the reviewer parent sandbox, with the host
conformance evidence retained above.


BR-1 correction: replaced normalized Git diff with a framed SHA-256 fingerprint
of actual tracked bytes, modes, types and leaf symlink targets, retaining HEAD
and index identity. Initialized gitlinks recurse into their own repository;
ambiguous ancestor symlinks and populated non-repository gitlinks fail explicitly.
Four old-code bypasses (assume-unchanged, skip-worktree, clean filter, chmod)
failed red and now stop before reviewer execution. Dirty overlays and generated
ignored/untracked assets remain allowed. All preparation tests pass (6.793 s),
the full judge package passes (7.616 s), and targeted boundary lifecycle tests
plus the live pair proof pass (14.814 s). Pair origin/main remains
b63f4bd76e3e81d1433cfeda544bb3bed8501522: the missing-assets failure reproduces,
fenced preparation takes 1.532 s, runtimebundle tests pass, and source is unchanged.
Fetched origin/main; it is already an ancestor. Full suite rerun follows.

Full suite after BR-1 correction PASS at
`bd486507794b44ba6bd094ed85c96fd6d77da30f`:
`make -k test ARGS='--json /tmp/ariadne-364-full-suite-3.jsonl'` completed
in 401.2 s, all 12 shards over 1,249 SDLC tests plus 53 other packages,
zero failures or missing tests. Log: `/tmp/ariadne-364-full-suite-3.log`.
Implementation acceptance restored; submitting BR-1 correction for re-review.

Boundary review round 2 REWORK: BR-1 addressed (reviewer independently restored
the prior guard and observed the regression fail). BR-2 Important identifies a
missing production close/dispute lifecycle proof for unverifiable findings;
accepted. Adding the promised lifecycle and mutation regression using existing
fixtures. Core implementation and targeted reviewer checks otherwise pass.

## Revisions

- 2026-10-14 — The final checklist item now describes implementation verification;
  close/land remain subsequent SDLC actions recorded in the Log and durable plan,
  so their future execution is not falsely marked complete before the close gate.

- 2026-10-14 — Spec 1 corrected with hub approval: the existing fence denies source
  writes, not arbitrary signals/network/outside writes. Preserve that boundary and
  add only the measured literal ps exception; a general confinement redesign is out
  of scope. Evidence and approved execution direction are recorded above.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. The existing process runner, JSON decoder,
ledger and fence provide the needed libraries; no new dependency. Thorough plan
uses ×0.2 design discount, 40% v2 implementation hours and 15% design buffer.
Four smaller modules cover fence, vocabulary, diagnostics, and boundary wiring;
two greenfield concerns cover declaration and admission. Integration includes
retry/cancellation and downstream live proof. Familiarity is 1.0 for this Go stack.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: greenfield-go-module design=0.10 impl=0.32
item: greenfield-go-module design=0.10 impl=0.32
item: api-integration design=0.60 impl=0.60
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 3.608
```

Plan-quality round 2 CLEAN: PQ-1/PQ-2 addressed; no new findings.
