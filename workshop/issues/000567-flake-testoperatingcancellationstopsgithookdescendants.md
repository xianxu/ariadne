---
id: 000567
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8dc3747fc6535d094bfd3aff87c2c649aff8be65' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T05:17:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants]
flow: {kind: quick, provenance: inferred, spec: "cbf89813", done: "c4b06df2"}
---

# flake: TestOperatingCancellationStopsGitHookDescendants

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants: Recorded #563 suite attempt 2cb913014a018f68422fdf9ef52180cf on Git-verified head 6a0b35faea2b351b5b1ec19a310f6ea1d19f1948 failed at doc_operating_test.go:419: Git descendant remains runnable: ?E (0.87s). Unchanged exact detached checkout targeted command go test ./cmd/tl-lead -run ^TestOperatingCancellationStopsGitHookDescendants$ -count=1 -timeout 2m returned exit0 PASS in0.986s. Captured original JSON events and targeted result are retained in common test-stats/evidence/000563/cancellation-failure.jsonl and cancellation-targeted-pass.txt. Full original suite is still running; no full retry or source edit. Investigate whether transient Darwin exiting state E is misclassified as runnable, without weakening cleanup proof.

## Spec

The cancellation assertion races asynchronous kernel teardown. The publisher
already kills its private process group on cancellation and return. Its Git
leader and output pipes can finish while a killed descendant is still visible
to Darwin `ps` as `?E` (the native manual defines `E` as trying to exit).
The current one-shot oracle accepts only absence or a leading `Z`, so it calls
this terminal transition runnable. An unchanged 30-run targeted reproduction
on the planning base produced 29 passes and the identical `?E` failure once.

Fix the test's completion boundary, not the publisher: capture the descendant's
owned group before cancellation, reject the test's own group, and wait until
`procgroup.Kill(pgid, 0)` reports `os.ErrProcessDone`, under the existing test
deadline. This shared Unix boundary already handles Darwin exiting/zombie-only
groups and absent groups. A live group remains a failure when the deadline
expires, and unexpected probe errors fail immediately. This is a condition
wait for the original cancellation, never a retry of the operation or test.
The hook child itself publishes its PID and readiness before blocking, so the
assertion covers an executing descendant rather than only the parent's fork.

ARCH-DRY/ORDER: reuse the same group-lifetime semantics as the neighboring
repository-cancellation fixture; do not invent a `ps` state parser or accept
`E` text as proof. A process group is the existing owner-lifetime resource;
no new lease or production state is needed. ARCH-PURE/SECURE: the only new
shipping logic is a pure AST lint rule. Native fixtures use private temporary
Git repositories/FIFOs and signal only the captured owned group. ARCH-PURPOSE:
sweep found this is the only direct `ps stat=` test query in the repository.
ARCH-MOCK: keep real local Git and kernel conformance; no service adapter is
added. ARCH-CONSTRAINTS: preserve all existing command/test deadlines and add
no production hot-path work; lint visits source calls once, O(AST nodes + argv
literals), with counted N/2N checks. ARCH-FUNERAL: immediate fixture cleanup
cancels/joins the publisher and kills its own blocked child on failures;
`t.TempDir` removes readiness files and repositories. ARCH-EVOLVE: no persisted
format, public state or lifecycle contract changes.

Alternatives rejected: accepting an `E` suffix merely adds another textual
state guess; a fixed sleep or longer timeout masks scheduling. Changing the
production group runner would widen scope despite successful group signalling.

## Done when

- The captured native failure is retained with provenance; the unchanged original test reproduces the same `?E` rejection.
- Cancellation waits on the existing owned-group terminal condition, uses child-origin readiness, and preserves the existing deadline and error/WaitDelay assertions.
- A native leader-only-kill mutation fails the descendant guard; original/corrected lint source fixtures reject the one-shot `ps` oracle without banning unrelated commands or production diagnostics.
- Focused normal and race checks plus a recorded full suite pass; closing this issue unregisters its flake.

## Plan

- [ ] Retain the original JSON capture and test-source excerpt; make the narrow `test-ps-liveness` catalog rule red on that source, with import-alias/context/absolute-binary and unrelated-command controls plus counted N/2N visits.
- [ ] Fix `TestOperatingCancellationStopsGitHookDescendants` to obtain child-origin readiness, capture and validate its group before cancel, and observe the shared terminal condition under its existing guard; register cleanup before failure-prone waits.
- [ ] Verify the real Git fixture normally and with race instrumentation; temporarily change only the operating command kill to leader-only, observe the negative control fail, restore it, and retain the result. The test must not accept a surviving owned group or an unexpected probe error.
- [ ] Run the tracked-file catalog and relevant cancellation controls; record the lesson, push, run the recorded suite, close for fresh review and land through central SDLC.

Risky function strategies: `TestOperatingCancellationStopsGitHookDescendants`
uses a ready blocked native child and the independent leader-only-kill mutation;
`psLivenessQueries` receives the recorded old oracle and adversarial command
syntax, checked through both Check and CheckSources. Finite case inventories
live in executable tests. No new user-visible state or production state machine
is introduced.

## Log

### 2026-10-22

### 2026-10-23 — diagnosis and proposed quick-flow fix

Claimed and started on the fresh main base. Native original repetition:
`go test ./cmd/tl-lead -run '^TestOperatingCancellationStopsGitHookDescendants$' -count=30 -timeout 3m -json`
returned exit 1 in 26.605s: 29 pass, one failure with descendant `?E`, exactly
matching the original #563 capture. Raw results are under the Git common dir,
`sdlc/test-stats/evidence/000567/original-repeat.jsonl`; no code was edited.
The native `ps(1)` state section documents `E` as trying to exit; existing
`pkg/procgroup.Kill` explicitly covers XNU's exiting/zombie-only group result.
Production `operatingCommandGroup` already installs Setpgid, group SIGKILL and
one-second WaitDelay; adjacent the repository-cancellation regression uses a bounded
group completion probe instead of an immediate `ps` snapshot. The proposed
patch changes the fixture/oracle and adds the mechanical lint backstop only.
