---
collaboration-mode: discussion
---

# Scratch fallback and environment failure plan

> **For agentic workers:** Consult AGENTS.md Section 3. Keep implementation
> coordinated in the main session; the binary owns the fresh close review.
> Never commit or create new dirty paths while cmd/sdlc targeted tests run.

**Goal:** An inherited unusable temp directory must not defeat scratch fallback,
and a Go work-directory startup failure must not become evidence that main is red.

**Architecture:** Reuse the existing scratch admission and local-check execution
seams. Select a writable normal fallback without changing process-global env;
override both child temp variables. A conservative pure startup classifier returns
a typed execution error before flake classification, status posting or baseline
blame. Retain a separate diagnostic log, not a failed code-suite attempt.

**Tech stack:** Go, existing scratchdisk filesystem seam, real temporary Git
repositories and the existing stateful CI/status fake; no live sandbox changes.

## Decisions and alternatives

1. Recommended: repair temp selection in scratch preparation and normal checkout
   fallback, and treat recognized startup failure as a could-not-run error.
   Existing run leases, slots and cleanup retain their owners.
2. Only overriding GOTMPDIR in the final child is insufficient: an unusable
   inherited TMPDIR can prevent the detached checkout from being created first.
3. Adding an environment outcome to normal SuiteAttempt records would require
   every acceptance, flake, republish and compatibility reader to interpret it.
   Avoid that persisted schema change: environment diagnostics are a separate
   log file with no suite result or commit status. They cannot poison a later
   same-head pass, erase an earlier real failure, or become passing evidence.

No timeout bumps, retry policy changes, machine sandbox edits, shared staging,
private landing binary, new grants or live sandbox reproduction. #476 owns
quoted nested Go-test attribution; do not repair that parser here.

## Core concepts

| Name | Kind | Location | Status |
|---|---|---|---|
| Normal temp selection | IO helper | cmd/sdlc/internal/scratchdisk | new |
| Scratch preparation | owned scoped resource | cmd/sdlc/testscratch.go | modified |
| Work-dir startup evidence | PURE predicate | cmd/sdlc/suiteenvironment.go | new |
| Suite environment error | typed execution error | cmd/sdlc/suiteenvironment.go | new |

| Seam | Location | Wraps / tests |
|---|---|---|
| Scratch and normal mkdir | internal/scratchdisk | actual private folders; controlled candidate/probe fixtures |
| Child execution | localcheck_run.go / runUnderSlot | real shell/Go child through existing runner |
| Status and attempt evidence | localcheck_run.go | existing stateful CI fake and on-disk store |
| Main comparison | suitebaseline.go, landingtrain.go | ordinary and train baseline fixtures |

## Temp selection and ownership

- Keep #474's Darwin APFS requirement and other-platform overrides intact.
- Keep normal inherited TMPDIR when usable. If it is unusable, try the platform's
  normal candidates independently of that variable (Darwin per-user temp, then
  Unix default). Validate by creating/removing an owned probe, not mode bits;
  ACLs and sandbox policy can defeat a permission-bit test. Deduplicate candidates.
- Do not unset or mutate os.Environ globally. Go and shell children get explicit
  TMPDIR and GOTMPDIR assignments at the existing final-env seam.
- prepareScratch owns one per-run directory on scratch or the selected normal
  root and removes it after the child stops. Return explicit scratch provenance
  rather than using non-nil env as the proxy for RAM placement, so fallback
  temp churn does not trigger the RAM-disk leak report.
- Return a visible error if no normal fallback can be created. Report replaced
  inherited TMPDIR/GOTMPDIR values in one-line notes. Never silently inherit the
  rejected value. Do not read or alter the operator's current sandbox settings.
- Repair the existing scratchdisk.MkdirTemp normal-temp fallback lazily after
  scratch/current fallback creation fails. Preserve explicit custom fallback
  errors; broaden recovery only for the os.TempDir-derived fallback used by
  local detached checkouts. Thus TMPDIR failure before prepareScratch is covered.
- The Unix default is an intentional exception to tmp-literal: this helper must
  escape an unusable TMPDIR. Give that one call a reasoned catalog suppression;
  ordinary callers still use the shared helper. This is a runtime permission
  invariant, not a mechanically provable new Go lint rule.

## Environment evidence and retention

- Read the complete bounded run log once using the existing 64 MiB limit; share
  that read with normal suite facts. Missing, oversized or malformed evidence
  does not prove environment failure.
- Recognize direct Go startup diagnostics beginning with `go: creating work dir:`
  and an actual creation/access failure. Reject zero-exit or timed-out runs,
  named test failures, test execution/pass evidence, unrelated compiler/setup
  failures, and mixed/unknown failure output. Handle the real test-shard.py
  compile-before-report form with its Python/make wrapper explicitly. Quoted
  copies or incidental mentions confer no authority.
- Pure tests cover all-startup failures, the real compile-startup wrapper,
  multiple failing processes, mixed code/test failure, successful activity,
  malformed/truncated evidence and quoted output controls. N/2N fixtures vary
  startup diagnostic count; classification is one linear pass with no subprocesses.
- In runLocalCheck, recognize the error before knownFlakeStep and before normal
  SuiteAttempt publication. Save the raw log as `<head>/<check>.environment.log`
  under the existing local-check log tree, then return a typed environment error
  naming GOTMPDIR/TMPDIR and that log. No success or failure commit status, PR
  failure comment, flake classification or failed code-suite record is published.
- That diagnostic is deliberately a separate evidence channel. At most one
  latest environment log per check/head is retained; existing newest-20-head
  pruning removes it. The existing private inflight owner is cleaned normally.
  A failure to retain the log is explicit; never point at silently deleted proof.
- Normal suites and main baselines use this same seam. A same-head repaired-env
  run can post success; an earlier genuine unexplained failure still blocks it.
  No persisted attempt fields or version change, and no lossy record rewrites.

## Caller behavior and state

Actions: repair temp-directory access and rerun the printed check/merge command;
or fix a genuine reported test/code failure and follow normal close/merge.

| Failure mode | Visible result | Recovery | Possible loss |
|---|---|---|---|
| Scratch denied; normal fallback writable | run with explicit writable temp env and note | none | no code/history loss; temp directory removed after run |
| Inherited temp denied before checkout | choose verified normal fallback | none | same existing disposable checkout lifetime |
| No usable temp root | environment could-not-run error; no acceptance record | repair access and rerun | no passing test evidence is claimed |
| All recognized Go work-dir startup failures | environment error naming variables and retained log | repair access and rerun same head | latest diagnostic replaces earlier diagnostic only |
| Mixed, test, compiler, timeout or ambiguous failure | existing ordinary failure/flake rules | normal fix/diagnosis | no failure is hidden |
| Environment failure only on main baseline | baseline could-not-run; main health unknown | repair access and rerun | no red-main claim or car blame/splitting |
| Cancellation | existing unwind/no result | rerun exact command | existing interruption behavior |

Ordinary diagnoseSuiteFailure already distinguishes run errors from failed test
records; retain the environment explanation. Train mainIsRed must return a
baseline-run error to runSuite rather than swallow it as false and blame/split
cars. Preserve existing train cleanup/refusal ownership; no new train events or
state machine. A lock cannot establish filesystem write permission or classify
process output; the existing run lease remains the resource-lifetime owner.

## Architecture and operating envelope

- ARCH-DRY/PURE: one normal-root selector and one startup predicate serve suites,
  known-flake reruns and baselines; filesystem operations stay at existing IO seams.
- ARCH-PURPOSE/MOCK: exercise actual child env and a real temp write, plus stateful
  status/store assertions and both baseline callers, not only helper output.
- ARCH-CONSTRAINTS: batch setup, fixed small candidate set, constant mkdir probes
  per candidate, no polling or new processes per package; log work O(bytes),
  existing 64 MiB and line bounds. Linux and Darwin fixtures; no live deny probe.
- ARCH-SECURE: inherited paths/logs are observations, not authority; positive
  creation proof and narrow startup grammar; unknown/mixed logs remain failures.
- ARCH-ORDER/FUNERAL: no new asynchronous actor or durable state machine. Existing
  leases own runs; per-run temp cleanup follows child termination. Diagnostic
  logs have the existing 20-head retention bound and never act as suite evidence.
- ARCH-EVOLVE: no attempt schema changes. Existing code-failure records remain
  readable and untouched; a separate log cannot reinterpret their outcomes.

## Tasks (one atomic close boundary)

- [x] Add failing scratch fixtures for unusable inherited TMPDIR and GOTMPDIR
  with rejected/unwritable scratch, off mode, usable inherited roots, no usable
  candidate, and cleanup. Use a regular-file path or injected permission failure,
  not the live sandbox or chmod assumptions. Run targeted tests and observe red.
- [x] Implement normal-root repair and explicit scratch preparation/error/provenance.
  Update existing off/HFS expectations while preserving admission rules; prove a
  real child writes successfully under the returned env and checkout fallback.
- [x] Add failing pure startup classifier tables and caller/store tests. Include
  same-SHA environment error then success, prior genuine failure still blocking,
  no status/flake/baseline call on direct env error, and baseline-only environment
  failure in ordinary and train paths. Observe the expected reds before fixes.
- [x] Wire conservative classification and retained diagnostic error; propagate
  train baseline errors. Run relevant scratch, local-check, republish/guard,
  ordinary landing, train and golint tests from a stable checkout.
- [x] Update local-check help, README, atlas and a concise lesson about inherited
  temp fallback and evidence separation. Commit verified implementation.
- [ ] Open PR, run recorded suite with normal central sdlc, close via its mandatory
  fresh review, then enter merge on SHIP. No manual main chasing or private binary.
- [ ] Verify Git merge, card/archive/clean rest and report PR plus Git-derived SHA
  to hub. Only after #477 lands, claim and plan separately authorized #476.

## Revisions

### 2026-10-19 — implementation details after temp fixtures

Normal roots now own run directories too, so their existing stale sweep must
include the reserved sdlc-run prefix. The intentional Unix default is explained
in an ordinary comment: the tmp-literal catalog does not flag this selector
call, so an unused suppression would itself be invalid. No new lint exception
or change to the selected fallback policy is needed.

### 2026-10-19 — preserve evidence separation on exceptional IO paths

The existing optional-store fallback now writes to a bounded per-check pending
log until classification, then publishes the ordinary legacy filename. This
keeps a failed environment-log rename from exposing raw startup output as
legacy suite evidence. Known-flake reruns use the same classifier and retained
diagnostic helper; their genuine parent failure remains intact. These implement
the approved single environment-evidence boundary without a persisted schema
or additional acceptance state.

### 2026-10-19 — BR-1 checkout creation/cleanup parity

The first close review exposed a missed cleanup path: repaired detached
checkouts can live outside the inherited root. Extend scratchdisk.Roots for
os.TempDir-based families to enumerate the inherited, Darwin per-user and Unix
default candidates, including when the inherited root later becomes healthy.
Normalize/deduplicate without probing or creating cleanup roots. Explicit custom
fallback families (including review checkouts) retain their existing root set.
This adds one bounded per-user lookup and at most three normal candidates to
the existing sweep, with no polling or new state. The normal fallback selector
keeps its lazy write probes.

Regression: a private Darwin per-user root contains old and fresh sdlc-local
checkouts; cleanup must remove only the old one, both with inherited TMPDIR
broken and after it becomes healthy. Pure candidate tests cover the Unix
default and deduplication on every platform.

### 2026-10-19 — BR-2 creation/cleanup rule across both artifact families

BR-1 fixed checkout cleanup, but its sibling per-run directories still used
current placement as the cleanup root set. Enforce one rule for every disposable
family changed here: sweep every root creation can select, independently of
which root is selected now. Local checkouts use Roots(local-checks, os.TempDir);
per-run directories use Roots(tmp, os.TempDir), both sharing the same normal
candidate enumeration. Explicit review-cache families remain unchanged because
creation never gained normal fallback there. No new mechanism or state.

The private Darwin recovery fixture covers old removal, fresh/live preservation
and unrelated-tool preservation, first during fallback and then after inherited
TMPDIR recovers. Unix default enumeration is covered by the platform-neutral
candidate fixture. Also pin recovery text for an environment-failed targeted
rerun to the declared suite command; suite-rerun is only a record/log name.

If another review proposes a further change to this cleanup mechanism, stop
and take a complete re-plan to the hub before changing it again.
