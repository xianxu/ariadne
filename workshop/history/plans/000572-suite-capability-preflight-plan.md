---
collaboration-mode: discussion
---

# Suite host capability preflight implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. One atomic change, one close review.

**Goal:** Refuse an execution environment that cannot supply the suite's host capabilities before it can queue, launch, or publish evidence.

**Architecture:** Probe independent host capabilities, not test-output wording. Reuse the existing nonpassing `suiteEnvironmentError`; keep attempt formats, failure classification, reclassification authority, and publication rules unchanged.

**Tech Stack:** Go 1.27.3, existing kernel boot-identity reader, private Unix listener, existing local-run and resume fixtures.

## Decisions and evidence

The original #561 log (attempt `c5fff316216ec03a781b4796f770a3e5`, head `94aa0864e648e752b90c7fbe354b19c1569a9126`) contains many EPERM failures **and** an `invalid attempt record` failure. Its SHA256 is `b59b0ca759609006000b60cfa623460af4aa9821e76c550e1237fec0806ae646`. The unpiped same-head pass log hashes to `3f8788ebb35242129e880cee61024b6bb6c4b155d236c1d0094735a94432f02d`. Both are retained under the common Git directory; no original evidence is edited.

Options considered:

1. Majority-denial log classification: rejected because it can erase an unrelated code failure.
2. Strict all-denial log classification: cannot explain the original mixed capture, and the sharded reporter retains only the last 40 output events per failing test. A complete outer log is not complete diagnostic evidence.
3. Independent host preflight: selected and approved by ops:0 on 2026-10-23 (receipt `d01f12b8-070e-4d91-ab10-7056c270e6d4`). It prevents this environment from executing the suite in the first place.

No new post-hoc log classifier or invalidate verb. The original mixed attempt stays a failure. EPERM in code running on a healthy host remains an ordinary failure. A preflight pass proves only the named capabilities at that instant, not that all later failures are code defects or that nested sandboxes are usable.

## Core concepts

| Pure entity | Lives in | Status |
| --- | --- | --- |
| `suiteEnvironmentError` formatting | `cmd/sdlc/suiteenvironment.go` | modified |

The pure `suiteEnvironmentError.Error` formatter takes the stored operation diagnosis and wrapped error. It distinguishes EPERM (`errors.Is`, not message matching) from other unavailable-capability errors. Only independent EPERM yields the sandbox/whole-call advice; other errors name the failed capability without asserting sandbox causation. Its colocated tests need no IO. The existing startup-only log diagnosis remains unchanged.

| Integration point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| `probeSuiteCapabilities` / `suiteHostCapabilities` | `cmd/sdlc/suitecapability.go` | new | `queueBootID`, private temporary directory, `net.ListenUnix`, close/remove |
| `preflightSuite` | `cmd/sdlc/suitecapability.go` | new | invokes probe and binds check/head to its result |
| Ordinary execution preflight | `cmd/sdlc/localcheck_run.go` | modified | shared probe before helper/worktree/attempt allocation |
| Affected graph preflight | `cmd/sdlc/localcheck_affected.go` | modified | shared probe before optional generator and graph worktree |
| Eligible resume preflight | `cmd/sdlc/localcheck_resume.go` | modified | shared probe before worktree, child writer or immutable link |
| Automatic flake rerun preflight | `cmd/sdlc/localcheck_flakes.go` | modified | shared probe before child writer/queue |

One probe owner, injected through `localRunDeps` for deterministic tests (ARCH-DRY, ARCH-PURE). Production has no environment variable to bypass it. The existing local-run Git repository, status store, and resume fixtures provide stateful external doubles. The capability fixture tracks acquired/closed listeners and owned scratch removal, including errors.

## Ordering, scope, and limits

Ordinary `suite` and `affected` checks, including typed-runner execution and diagnostic baselines, preflight before any execution resource. Unrelated named checks keep their existing contract. Resume first resolves its read-only proof: already-completed saved-pass reposts need no execution capabilities. Only an eligible new execution probes before creating its child or link. Automatic reruns probe before their child: a refusal preserves the genuine parent failure, records no invalid child and posts no status.

The sequence is fixed: context check -> boot identity read -> create private short scratch -> bind private Unix listener -> close listener -> remove owned scratch -> authorize existing execution. An error refuses; cancellation between steps refuses with cleanup. There is no concurrent probe, child process, background goroutine, retry, durable flag, or state machine added. Scope-local ownership and deferred cleanup bound its lifetime (ARCH-ORDER).

Use a short, canonical, mode-0700 temporary directory: `scratchdisk.Roots("tmp", "/tmp")` selects TestScratch's tmp directory or the short system temporary root, with a documented Unix-path-length exception. Name it `sdlc-run-probe-PID-RANDOM`, recognized by the existing `runDirRE` and `sweepStaleTemp`; call that existing sweep on the selected roots when preparing a subsequent probe. Do not inherit an arbitrarily long TMPDIR socket path. No connection/accept loop is needed: successful bind is the capability. Directory and listener are removed on every normal return; report cleanup errors rather than claiming readiness. A process crash can leave one tiny temporary directory/socket, reclaimed by the existing PID/dead-or-24-hour sweep policy (ARCH-FUNERAL). No new registry or scratch family.

This is a local preflight with two fixed kernel operations and scratch work; no network request, polling, suite queue slot, or test execution. Probe CPU/memory are constant apart from the existing temporary-directory sweep, whose cost follows the directory population; disk use is one empty private directory/socket. Existing synchronous kernel calls are used without goroutine timeout wrappers. Unavailable capabilities fail closed. No credentials or persisted input are consumed; a test log cannot grant authority (ARCH-SECURE). No wire-format or vocabulary lifecycle changes (ARCH-EVOLVE).

Failure prints `environment failure: could-not-run`, the denied capability and cause, and on EPERM: `run sdlc local-check run --name CHECK --rev SHA as the whole Bash call, without a pipe, redirect, or command wrapper`. Resume advice preserves its exact parent retry command. Omit the log clause when preflight created no log; do not print an empty path. Existing older failed records are neither erased nor promoted.

## Implementation steps

- [x] Capture the original mixed log unchanged into a text fixture, plus a clearly labelled all-denial projection for a control. Bind fixture comments to the original SHA256; no claim that the projection is the real run.
- [x] Add red pure tests for wrapped EPERM versus EACCES, unrelated errors and quoted denial text. Add native probe tests for boot denial, listener denial, ordinary error, cancellation, successful cleanup, cleanup failure, and long TMPDIR. Use deterministic fake resources for error paths.
- [x] Add red ordinary-run integration cases with injected capability outcomes: suite/affected/baseline/typed checks refuse before helper, worktree, writer, queue, registry or status effects. Successful repair on the same head records normally. Non-test checks retain existing behavior.
- [x] Add red controls replaying captured mixed and projected all-denial outputs on a capability-healthy host: both remain genuine failed attempts. Existing failures/statuses survive a later capability refusal; no flake registration or reason-only disposition occurs.
- [x] Add red automatic-rerun and manual-resume controls: failed preflight leaves no child/link/new queue work; genuine parent preserved. Saved-pass repost remains available without a probe. Confirm resume retry works after capability restoration.
- [x] Implement the shared probe and thin entry-point calls; improve environment error formatting without changing startup classifier semantics. Keep production changes within this boundary.
- [x] Run focused checks: `go test ./cmd/sdlc -run 'TestSuiteCapability|TestSuiteEnvironment|TestSuiteInfrastructure|TestResume' -count=1 -timeout 20m` and `go test ./pkg/golint -count=1`. Keep checkout HEAD/status fixed during cmd/sdlc tests.
- [x] Add an opt-in Darwin conformance child under a minimal Seatbelt profile denying the boot read or Unix bind. Run it outside the agent sandbox; prove could-not-run, exact guidance, no queue/attempt/status effects, then repeat the probe on the host. No full suite inside this conformance fixture. If native sandbox execution is unavailable, retain raw evidence and report the limitation instead of a PASS.
- [ ] Update `atlas/workflow/ci-merge-check.md` and local-check help with the independent proof contract and existing-record limit. Run focused checks, commit/push, then `sdlc local-check run --name suite` on the final head.
- [ ] After recorded suite PASS, run `sdlc close --issue 572 --verified '…'`, address its review, and use the normal landing queue. Report outcome to hub. No :0 stage or live failed-attempt mutation is needed.

## Risky-function test strategy

`probeSuiteCapabilities`: resource-owning fake plus real private listener, injected boot failures, cleanup errors, and native Seatbelt conformance. `runLocalCheck`: existing real-repository/status fixture asserts absent side effects and same-head repair. `knownFlakeStep`: preserve parent/no invalid child fixture. `resumeSavedSuite`: existing proof/link fixture asserts preflight precedes link and saved-pass repost bypasses execution. `suiteEnvironmentError.Error`: exact command and absent-log formatting table. These test behavior and state, not source-text matching (ARCH-MOCK, ARCH-PURPOSE).

## Revisions

### 2026-10-23 — Include affected graph generation in the execution boundary

The entry-point sweep found affected's optional graph generator precedes `runLocalCheck`, even for a docs-only selection. Apply the same approved preflight before that graph call and test no graph/execution/status effects on refusal. No new mechanism. Correct the cost scope: the reused stale-temp sweep scales with its existing directory population, while the new capability operations are fixed.

### 2026-10-23 — Close round 1 corrections

BR-1: correct the Core concepts table to the implementation: the pure EPERM formatter is `suiteEnvironmentError.Error` in `suiteenvironment.go`; the three functions in `suitecapability.go` invoke/own IO and belong in the integration table. No additional pure helper is needed. Add the already-implemented affected graph boundary to that table. BR-2: retain the short-path explanation as an ordinary comment, removing an unused lint suppression. BR-3: document the new refusal/retry contract in README as well as help and atlas. Recorded-suite and landing gates remain outstanding.

### 2026-10-23 — Enforce the approved short path on configured scratch roots

The recorded suite found that `WF_TEST_SCRATCH` can name a long root inside a test, independently of TMPDIR. Enforce the existing short-path contract on the canonical directory plus socket name: at most 103 bytes. Try the selected scratch root and then the existing short system root, with at most two directory creations. Only path length permits this selection; any creation, canonicalization, bind, close or removal error refuses immediately. Remove the rejected owned directory before trying the short root. The existing sweep recognizes both placements. Tests cover long TMPDIR, long configured scratch roots and symlink-resolved roots; native bind denial also uses a long configured root to prove it stays EPERM rather than EINVAL. This repairs an implementation gap in the approved envelope; no new classifier, state or recovery mechanism.
