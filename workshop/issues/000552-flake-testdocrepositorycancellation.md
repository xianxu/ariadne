---
id: 000552
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '08a485b4267eb10ae00c39c1b42fd531bebf4d7c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T09:21:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestDocRepositoryCancellation]
flow: {kind: quick, provenance: inferred, spec: "4ab9c5ad", done: "847448a6"}
---

# flake: TestDocRepositoryCancellation

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestDocRepositoryCancellation: Recorded suite attempt 4638dd482de33746116a1788f13ba510 failed TestDocRepositoryCancellation on exact head 0ae0016dbbe6634bfbf59132045e683bfa76862a: strconv.Atoi parsing empty readiness content. Afterward, a clean detached worktree at that exact Git-verified head ran the whole top-level test: go test ./cmd/tl-lead -run ^TestDocRepositoryCancellation$ -count=1 -v; PASS, test 0.23s, command 0.390s. Worktree status clean. Source polls file existence before reading an echo-created PID file, so creation can be observed before PID content. The suite also had a separate reproducible context-guard failure, being fixed on #551; this registration covers only the independent tl-lead fixture.

## Spec

Fix the cancellation fixture's readiness publication, keeping production process cancellation unchanged.
The shell currently opens/truncates the final readiness file before echo writes its PID; the observer's
existence check can therefore admit empty content. Write the complete PID to a sibling staging file,
close it, then rename it to the observed path on the same filesystem. File existence then denotes
completed publication. Keep the current readiness/cancellation bounds; do not add retries or sleeps.

Use native FIFO barriers, following TestExecResumeWaitsForTerminalReceipt, to hold the actual fixture
after opening its output but before writing the PID. At that barrier, assert the final readiness path
is absent and the staging bytes reproduce the recorded empty-content failure. After explicit release,
require a complete PID and retain cancellation-error, inherited-pipe and process-group termination checks.
The fixture's process group and temporary directory own all children, FIFOs and staging files; cleanup
cancels and awaits the invocation even when an assertion fails while the writer is paused.

Alternatives: waiting for parseable content would hide incomplete publication behind a retry;
a separate final-ready FIFO is sound but adds a second completion protocol when rename makes the
existing readiness path sufficient. Reuse existing FIFO fixture patterns, not a new shared API
(ARCH-DRY). This changes only test IO, with no new production logic (ARCH-PURE), configuration,
persisted product formats or user-facing states (ARCH-EVOLVE). The required interleaving and whole-group
cancellation defend the actual flake rather than raising timeouts (ARCH-PURPOSE, ARCH-ORDER).
The private fixture is bounded by existing hang guards (ARCH-CONSTRAINTS); it uses no credentials or
live fleet resources (ARCH-SECURE), and all its temporary resources die with the test (ARCH-FUNERAL).
Its real shell/process boundary is the existing stateful executable fixture (ARCH-MOCK).

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.
- A deterministic native regression reproduces the open-before-write window and rejects publishing the final path before the complete PID is written.
- The real suite failure capture is retained in testdata and consumed by the regression; cancellation still returns context.Canceled and leaves no runnable process-group members.
- Targeted verification and the recorded full suite pass; the close/landing workflow retires the flake registration, verified through the registry's read surface.

## Plan

- [x] Retain the recorded empty-PID failure and add a barrier-controlled regression; observe RED with direct publication.
- [x] Publish the fixture PID by sibling staging/close/rename, release the barrier, and verify complete readiness plus the existing cancellation controls.
- [x] Sweep sibling readiness consumers and review a mechanically sound lint guard with the hub; avoid a lexical rule that confuses safe validating readers with this race.
- [ ] Run targeted tests and the recorded pushed-head suite, retaining their proof for close and landing.

## Log

### 2026-10-22

- Claimed and started from main in canonical ariadne:5. The production adapter already owns a process
  group; the defect is the test's assumption that os.Stat success means the shell's PID write finished.
- Read the real retained attempt 4638dd482de33746116a1788f13ba510 and its raw failure log. It records
  doc_repository_test.go:280: strconv.Atoi parsing empty readiness content. The unrelated context-guard
  failure in that attempt belongs to #551 and is not part of this flake.
- The sibling TestOperatingCancellationStopsGitHookDescendants already signals a FIFO after writing
  its PID; it does not infer content completion from file existence. TestExecResumeWaitsForTerminalReceipt
  supplies the native RDWR FIFO barrier pattern. testfix.AwaitPIDFile is internal to cmd/sdlc and cannot
  be imported by tl-lead; moving that package is unnecessary for this fixture-only correction.
- Proposed quick-flow plan sent to the hub before code. Need a decision on the lint tripwire: a broad
  direct-shell-PID ban would flag safe consumers that validate content or await a post-write signal;
  the deterministic publication interleaving is the precise behavioral guard for this protocol.

- Hub receipt 7a415395-d0ad-4107-b2ce-38cf695bfc64 approved the atomic-publication plan and no new lexical
  lint; the deterministic interleaving guards this protocol. The fixture comment names the safe pattern.
- Added the native FIFO output-open/PID-write barrier with direct publication still present. RED:
  TestDocRepositoryCancellation failed in 0.56s with "readiness published before PID write: <nil>";
  failure cleanup cancelled and awaited the blocked invocation.
- Changed only the fixture writer to sibling staging/close/rename. GREEN: the same whole test passed
  in 0.41s (package 0.573s), including captured empty-content replay before release, complete PID
  afterward, context.Canceled, inherited-pipe completion and no runnable process group. No timeout changed.
- Broader targeted verification passed: doc repository tests and adjacent native FIFO/cancellation
  controls in cmd/tl-lead (13.038s); the full pkg/golint catalog including tracked files (1.274s).
  The final recorded suite is still outstanding; the last Plan row remains unchecked.

- Close round 1 returned REWORK solely for BR-1 Important, required-verification-evidence: the
  recorded full suite is still pending. The reviewer found no implementation defect; independently
  passed the cancellation regression ten times, neighboring repository/FIFO/cancellation tests and
  pkg/golint; restoring direct publication via a temporary overlay reproduced the intended RED.
  Keep the review sidecars and provide the actual suite result before the next close.
- Current recorded suite: pushed code head 1603ce1215d805c4d1d63be9c936df8bdaeba65f,
  lease-1199673725, attempt ffa80947944ff752aeb1b94024e0d3ca, retained terminal 84004.
  Its normal-FIFO waiter remains intact (position 7 at the close review). No suite pass is claimed;
  no code correction or replacement run is needed for this verification-only finding.

### 2026-10-22 — drain/lift checkpoint

- Authenticated graceful drain c28152a0-6925-45a6-a5b5-ae9c58bef2b9 arrived while the recorded suite
  was queued, not running. Left it queued and outstanding: lease-1199673725, runner 84443,
  terminal 84004, code 1603ce1215d805c4d1d63be9c936df8bdaeba65f, FIFO1 behind #551's running suite.
  No owned running check needed waiting or stopping; no queue cancellation or duplicate was started.
- Before this checkpoint finished, authenticated lift 3e5da2b2-5fa7-44a6-8de8-81016d18ae36 explicitly
  authorized continuation of current work. Continue with the same retained waiter; queued work is
  still outstanding, not restart-safety evidence. No formal acknowledgement command or courtesy reply.
- Exact next step: await terminal 84004's recorded suite result. On PASS, retain its exact attempt,
  head and log hash, tick the last verification checkbox, and close again to dispose of BR-1. Then
  mark draft PR399 ready, publish the close head and enter canonical sdlc merge; after landing verify
  TestDocRepositoryCancellation is absent from the flake registry. No full-suite pass is claimed yet.

## Revisions

### 2026-10-22 — verification checkpoint

- Clarify the last Plan checkbox as pre-close verification; close, landing and registration retirement
  remain subsequent workflow steps and explicit Done-when requirements, rather than future work that
  must be ticked before the close gate. No behavioral design or implementation changed.
- The suite waiter survived the 10:57 Couch restart and is now next to run behind #551's running suite.
  Its original lease-1199673725, runner 84443 and terminal 84004 remain retained. The registry read
  confirmed TestDocRepositoryCancellation is still registered to this working issue; verify retirement
  after landing. Draft PR399 CI is green. No new review or replacement suite was started while waiting.

### 2026-10-22 — deterministic readiness-publication fix

- Add the atomic-publication design, recorded-capture regression and explicit cancellation/retirement
  acceptance criteria to the original flake contract. Keep production behavior and all timeouts unchanged.
