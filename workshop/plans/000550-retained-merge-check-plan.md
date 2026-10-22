---
collaboration-mode: discussion
---

# Retained merge missing-check refusal implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for execution strategy. The boundary review belongs to `sdlc close`; do not dispatch a second boundary review.

**Goal:** A retained train merge that cannot become ready because its pushed head lacks a required local record exits with the exact check command instead of waiting indefinitely.

**Architecture:** Preserve the existing CI gate's missing-check evidence through its typed in-process error. Readiness carries a terminal error separately from its persisted summary; the existing cancellation owner ends and joins the wait, then releases its ticket. Checks still run through the existing local-check command and queue when explicitly requested, or through the unchanged admitted-merge path.

**Tech stack:** Go, existing CI classifier, landingline flock handles, context cancellation, existing real-Git/stateful-GitHub fixtures.

## Decision and alternatives

Recommend explicit refusal. It satisfies the issue's “run or clearly demand” contract without putting a second check owner inside the readiness goroutine. Automatic execution there would need check-run lifetime, lease/deduplication, queue attribution and check-failure handling while admission can concurrently succeed. Moving every check before ticket acquisition would change serial/grant behavior and delay receipt of a retained ticket. Neither is needed for this fix.

No timer guesses that a record will never arrive. Confirmed missing evidence is actionable immediately. If an existing check is already queued/running, leave it alone and wait for that result before retrying merge; otherwise run the printed command under its normal machine queue. Refusal forfeits this merge ticket's position, not the separate check's FIFO place. Hub approval is required for this policy.

## Core concepts

### Pure entities

| Name | Lives in | Status |
| --- | --- | --- |
| `ciVerdict` | `cmd/sdlc/cigate.go` | modified |
| `ciGateError` | `cmd/sdlc/cigate_wait.go` | modified |
| `localCheckCommand` | `cmd/sdlc/cigate.go` | new |

`ciVerdict.LocalNeeded` already names unsatisfied head-scoped local checks. Preserve that fact even when server CI determines the aggregate reason, including no-check/workflow registration. Its existing refusal formatter remains the single owner of canonical check commands. `ciGateError` gains copied missing-check names so `errors.As` can recover them through wrappers. This is in-process data only, with no persisted schema or new lifecycle. One verdict/error contains zero or more check names; future classifier reasons retain this same structured evidence rather than adding string matching.

### Integration points

| Name | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| `runPreparedCIGate` | `cmd/sdlc/cigate.go` | modified | pinned Git/GitHub CI observations |
| `carReadiness` | `cmd/sdlc/landingcar.go` | modified | Git/PR identity and gate checks |
| `vouchWhileWaiting` | `cmd/sdlc/landingcar.go` | modified | readiness publication and scoped goroutine |
| `enterLandingLine` | `cmd/sdlc/landingqueue.go` | modified | wait cancellation, outcome and held ticket |

Reuse `carFixture`, `integrationFixture`, the existing stateful fake GitHub checks/statuses, and real portable landingline locks. Their state models hold exact PR heads, PR state, checks/status records and held tickets. No new external dependency or fake service is introduced. Existing Git and GitHub seams remain the only effect boundaries; tests use disposable repositories and no live fleet state.

## Implementation contracts

1. `ciVerdict.refusal` copies `LocalNeeded` into `ciGateError` for every nonpassing aggregate verdict. The required-guard `missing(checks)` path must preserve the just-observed missing-local names rather than constructing an empty fresh verdict. Preserve existing aggregate reasons and diagnostics for other CI consumers. No-head/read errors cannot prove record absence and do not manufacture this metadata.
2. `carReadiness` and its injected callback gain a third `error` result. A typed CI refusal with nonempty missing-local evidence returns an actionable error plus a not-ready summary. Compose the existing canonical local-needed diagnostic with the observed aggregate diagnostic, so pending/red context survives. Include the verified pushed head in context; retain declaration-controlled suite/affected commands and full fallback instruction. Do not interpret human-readable reason strings.
3. `vouchWhileWaiting` publishes the not-ready summary once, then signals its cancellation owner with the terminal error and returns. Generalize the existing merged callback to accept a cause, retaining the merged sentinel. If publication returns `ErrNotWaiting`, admission/release already won: return without terminal cancellation. Ordinary not-ready results continue existing polling. Check parent cancellation before publication/signalling.
4. `enterLandingLine` joins readiness before `EndWait`, as today. When `Wait` failed because of the terminal readiness cause and the caller remains live, return that cause with its full commands. Successful admission, an observed `Settled` outcome and caller cancellation take precedence over a late observation. Do not claim atomic resolution of unobserved settlement: retain landingline's existing outcome/cancellation protocol. Normal refusal records and `turn.leave` release the ticket. The survivor-handoff readiness read must consume the error before invoking the prior conductor release callback; the caller's existing reformation cleanup remains responsible for its old turn.
5. Trains off, dry run, resume-only and `--no-ci` preserve existing behavior. A grant can win admission and then use the existing admitted check runner; the readiness refusal cannot revoke an acquired turn. No side effect is authorized from a stale pre-admission observation.

## State, failures and recovery

Actions: run the named check, retain an already running check, retry merge, cancel merge, fix/push the branch.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Pinned head missing required record, with any aggregate CI reason | refused; missing-check command(s), head and retry instruction | run named check, then retry merge | this merge's landing position |
| Same head already has a separate queued/running check | same refusal; separate check remains owned by its existing invocation | retain the check, retry merge after pass | this merge's landing position only |
| Head/ancestor/stand-in satisfies policy | normal readiness/admission | retain merge | none |
| Read fails or PR head is unconfirmed | existing not-ready diagnostic | retain/cancel merge or fix/push as today | no new loss |
| Successful admission wins readiness race | normal admitted merge; fresh gates own effects | retain merge | none |
| Final train outcome observed | existing landed/held/refused result | existing retry/cleanup instruction | no new loss |
| Caller cancellation/process death | existing cancellation; handle/kernel releases ticket | retry merge | landing position; unsaved terminal tail |

Covered: deterministic missing-local proof, masked verdicts, callback scope, ticket retirement, and existing satisfaction rules. Not covered: automatic check execution/deduplication, general timeouts for every unrelated readiness failure, changes to CI bypass policy, or a new durable recovery mechanism. No code or user work is removed.

## Constraints and principles

- ARCH-DRY/PURE: reuse `decideCIGate`, canonical refusal formatting and existing `ciGateError`; retain only typed policy at the readiness boundary.
- ARCH-PURPOSE: enumerate all aggregate CI reasons that can mask a required local record, not just the green-server case that motivated the report. Do not change unrelated retry policy.
- ARCH-ORDER: one existing ticket flock authorizes line participation for its lifetime. The readiness goroutine is cancelled/joined before return; no background check is started. Channels and held locks force race orderings in tests rather than short sleeps.
- ARCH-CONSTRAINTS: control-plane path; no extra network/Git probes or polling. Existing classifier/search costs remain unchanged; added work is O(K) copying/formatting K needed names. A terminal missing result is handled on its first confirmed observation, not after a wall-clock timeout. Count N and 2N missing names (8/16): preserve exactly K names, one terminal signal and zero additional per-name probes or repeated readiness passes. Existing read deadlines continue to bound IO.
- ARCH-MOCK/SECURE: stateful GitHub fixtures and real disposable Git/lock directories; missing evidence is trusted only after existing exact pushed-head and before/after PR-head checks. No untrusted persisted reason string authorizes exit or effects. No new credentials or live-process effects.
- ARCH-FUNERAL/EVOLVE: no new persisted format, file family or background owner. Existing ticket and 30-day wait/audit cleanup owns the additional refusal lifecycle; in-memory error metadata dies with the invocation. Older snapshots/readers are unchanged.

## Tasks and verification

### Task 1: preserve missing-check evidence

Files: `cmd/sdlc/cigate.go`, `cigate_wait.go`, `cigate_test.go`, `cigate_fake_test.go`.

- [ ] Add red tests for independent missing-local metadata alongside green, pending, red, unknown, workflow/no-check registration and missing guard. Valid-record controls retain existing verdicts.
- [ ] Run `go test ./cmd/sdlc -run '^(TestDecideCIGate.*|TestRunCIGate.*)$' -count=1`; confirm the new assertions fail on lost metadata.
- [ ] Implement typed propagation and required-guard preservation; rerun to PASS and commit.

Risky functions: `decideCIGate`/`ciVerdict.refusal` — table-driven reason × record matrix, with unchanged aggregate-verdict controls as independent guard. `runPreparedCIGate` — stateful fake observations, required-guard absence and head-lag/read-error negatives; prove only a coherent exact-head observation carries actionable absence. Executable tests own the case inventory.

### Task 2: end retained waits with the action

Files: `cmd/sdlc/landingcar.go`, `landingqueue.go`, `landing.go`; tests in `landingcar_test.go` and new `landingcar_missingcheck_test.go`; recorded text at `cmd/sdlc/testdata/landing/pair779-missing-check.txt`.

- [ ] Copy the verified original bounded capture verbatim, retaining provenance and historical DNS context. Add a capture-backed production-flow test reproducing its pending → missing affected/fallback sequence through the stateful gate fixture, asserting full refusal rather than parsing prose to make the decision.
- [ ] Add red held-line tests for terminal refusal/release and record-satisfied admission; use an existing fake queue/check runner to assert zero automatic check effects and unchanged separate waiter.
- [ ] Run `go test ./cmd/sdlc -run '^(TestCar.*|TestEnterLandingLine.*|TestVouching.*)$' -count=1`; establish red without shrinking deadlines.
- [ ] Add the callback error result and cause plumbing, preserving reformation and merged handling; update existing callback fixtures mechanically.
- [ ] Force cancellation-before-publication, admission-before-terminal, observed-settlement-before-terminal, trains-enabled-midwait and survivor-handoff-error orderings with channels/real handles. Assert readiness joins, ticket release, no effects from late errors and retry success after record arrival.
- [ ] Rerun focused tests and `go test -race ./cmd/sdlc -run '^(TestCar.*|TestEnterLandingLine.*|TestVouching.*)$' -count=1`, then commit.

Risky functions: `carReadiness` — production-gate fixture matrix plus affected/full-suite/ancestor and landing-scoped controls; independent assertion that no run/status/merge effect occurs. `vouchWhileWaiting`/`enterLandingLine` — barrier-controlled interleavings and actual ticket status, with callback-join and fresh retry assertions guarding against an error-only oracle. Count terminal signalling once for 8/16 names; no repeated readiness pass after terminal publication.

### Task 3: document, verify and land

Files: `cmd/sdlc/helptext/merge.md`, `README.md`, `atlas/workflow/sdlc-binary.md`, `workshop/lessons.md`, this issue/plan.

- [ ] Document read-only train readiness and refusal/retry, versus automatic checks after serial/granted admission. Explain preserving an already queued check rather than starting another blindly. Update the lesson; this is typed temporal policy, not a lexical lint pattern.
- [ ] Run surrounding landing/CI controls and the tracked-file Go lint catalog after staging; run `git diff --check`.
- [ ] Commit/push the final head with `sdlc pr`, run the recorded `sdlc local-check run --name suite`, then `sdlc close --issue 550 --verified '...'` and address findings. Suite pass and SHIP precede landing.
- [ ] Enter `sdlc merge --yes` promptly; its integration suite owns current-main verification. Report the Git-verified merge/PR outcome to the hub.

Agent-facing assertions cover the same CLI result under Claude/Codex-labelled fixture invocations (Qoder when available); no harness command-text heuristics. No operator-environment mutation is required for this refusal path. Ask the hub to accept the capture-backed fixture and disposable real-lock proof in place of a live stage before landing; do not stage or manipulate a live retained merge unasked.

## Revisions

### 2026-10-22 — approved exit text and no stage

Ops:0 approves this design (`4cec01c2-b2f4-4df2-9346-718e9a3e4c55`) with the additional required sentence: `merge stopped: run <exact command>, then re-enter sdlc merge`. Compose it from the same typed missing-check names as the canonical CI commands, never parse the existing message. For multiple commands, name each required action. Fixture plus real-lock verification is approved in place of a live stage. No code or effect-owner changes beyond the existing plan.

### 2026-10-22 — canonical command owner

The shared formatter is named `localCheckCommand` in `cigate.go`; both the existing CI refusal and the terminal readiness message derive commands from it. This small extraction implements the approved single-owner command wording without a second command grammar. The callback's third error result also reaches `maybeConductTrain` in `landingtrain.go`; after admission it remains a readiness summary for choosing serial versus train, while fresh landing gates continue to own effects.
