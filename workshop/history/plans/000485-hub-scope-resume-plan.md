---
collaboration-mode: discussion
---

# Hub-scoped status and serial fleet resume implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach. Steps use checkboxes; this is one atomic close boundary, with no Mx tags.

**Goal:** Show a hub its own work and resume its eligible slots one at a time, stopping visibly when any attempt fails or becomes uncertain.

**Architecture:** Reuse `.hub`, `ownerOf`, the fleet inventory and existing Couch seams. Share one scope value across slot rendering, resume planning and landing freshness. Couch's CLI already polls operation receipts; tl-lead waits for that CLI and stops the batch on any unsuccessful return.

**Tech Stack:** Go, Cobra, real Git fixtures, existing stateful Couch/SDLC/session-sync fakes.

**Approval:** Proposed; awaiting the owning TL. No implementation authorized by this document alone.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| hubScope | cmd/tl-lead/scope.go | new |
| ownership / ownerOf | cmd/tl-lead/team.go | existing, reused |
| resumeInputs / planResume | cmd/tl-lead/fleet.go | modified |
| resumeRow | cmd/tl-lead/fleet.go | modified |

`hubScope` holds self and the parsed hubs, and classifies slots using `ownerOf`.
One scope covers many fleet slots. It also supplies the hub's claimed repository names
for freshness (an exact slot claim still brings that repository's landings into scope).
An absent local `.hub` preserves the existing machine view for status; malformed or
unreadable scope refuses rather than silently widening. `fleet resume` still requires a hub.

`planResume` remains pure. Default output includes sole-owned slots and ownership ties
involving self; unclaimed and foreign slots become summary counts. With `--all`, show
every slot, but foreign/unclaimed/tied rows remain ineligible for automatic resume.
Use a distinct excluded display class for sole foreign owners; retain ambiguous for ties
and unclaimed slots. Never infer Couch's main hub from alphabetical discovery order.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| readHubScope | cmd/tl-lead/scope.go | new | Git root and existing readHubs |
| gatherSlots / renderedSections | cmd/tl-lead/slots.go, status.go | modified | fleet, Couch actors, issue observations, guarded session writes |
| gitNewestLanding | cmd/tl-lead/freshness.go | modified | local Git main refs |
| execResume / fleetResume | cmd/tl-lead/fleet.go | modified | Couch candidates and resume CLI |
| fakeCouch | cmd/tl-lead/fleet_test.go | modified | stateful resume completion and failures |

No new persistent format, service, ledger, lock or lifecycle is introduced.
Existing guarded section writes remain the sole session-file writer.

## Decisions and alternatives

1. **Use existing ownership semantics.** A shared scope adapter prevents three divergent
   filters (ARCH-DRY/PURE). A new Couch ownership API would require a separate pair change;
   current actors expose no owner or started-hub ordering. Report that limitation.
2. **Let Couch own receipt polling.** Its `exchangeSlotOperation` already waits for a terminal
   receipt, polling every 500 ms for up to five minutes. Reimplementing this in tl-lead would
   duplicate the protocol (ARCH-DRY). The currently synchronous `Resume` seam is sufficient.
3. **Stop after any error.** Continuing after a refusal gives a partial batch without a clear
   stopping point. Stop, return nonzero, and name current and untouched slots. Never retry
   the uncertain mutation. Read-only observation can inform recovery but never turn an
   uncertain operation into success (ARCH-SECURE/ORDER).
4. **Align the timeout with Couch.** Replace tl-lead's three-minute production bound with a
   documented six-minute outer bound: Couch's five-minute receipt budget plus one minute
   for admission/process teardown. Retain injected short bounds in tests. Keep both output
   streams in diagnostics so request IDs and Couch's recovery guidance survive.

## State design

Actions: inspect (`couch --operation-status ID`, `couch --peek ADDR`), correct refusal,
rerun dry-run, render Slots with `--all` (optionally releasing a hold), explicitly apply
again, or manually resume through Couch/the switcher.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| All eligible resumes return success | resumed per slot; batch succeeds | rerun dry-run to inspect | none |
| One resume still running | current slot in progress; no next attempt | wait for bounded Couch outcome | none |
| Couch explicitly refuses/fails | stopped; current reason and untouched list; nonzero | correct refusal, rerun dry-run | earlier slots remain resumed |
| Receipt/admission/outer timeout is uncertain | stopped, uncertain; request diagnostics, observed actor state, untouched list | inspect receipt and slot before deciding | current slot may have resumed; earlier slots remain resumed |
| Actor recheck fails after uncertainty | stopped, observation unavailable | inspect through Couch/switcher | same partial completion; no inferred rollback |
| Actors unavailable before planning | one top-level note; owned rows unknown; no resumes | restore access, rerun dry-run | none |
| Ownership missing/tied/foreign | counted or explicitly excluded with --all | correct .hub or use manual Couch path | none |
| Scoped Slots rewrite would hide an existing hold | refuse write, naming held slots | render with --all; release explicitly if obsolete | none; old section retained |
| TL process dies mid-resume | no further sequential attempts from that process | inspect current slot, rerun dry-run | latest outcome/output may be unavailable; prior resumes remain |

The active synchronous Couch process is the lifetime resource: the loop cannot call
the next slot until it returns successfully. An error exits the loop even if a remote
operation could continue. No cross-process batch mutex is claimed: Couch still owns
mutation admission. External/manual operations and runtime moves are not inferred from `.hub`.

## Tasks 1–3: implementation and test strategy

Each task follows test first → observed failure → implementation → targeted pass → commit.
All three close atomically with Task 4. The prior enumerated test inventories are replaced
by this function-level contract (PQ-2); this is a presentation change, not a scope change.

### Task 1: shared scope and displays

**Files:** create `cmd/tl-lead/scope.go`, `scope_test.go`; modify `slots.go`, `status.go`,
`fleet.go`, `fleet_test.go`, `trains_test.go`, `status_test.go`.

- [ ] `hubScope` / `ownerOf`: exercise adversarial overlapping claims and missing/unreadable
  ownership with pure ownership assertions; absent local .hub retains legacy status scope,
  unreadable ownership never widens authority.
- [ ] `gatherSlots`: use the two-hub stateful fleet fixture and counted N/2N selected issue
  reads to prove filtering occurs before IO, with zero excluded-slot issue reads.
- [ ] `renderedSections` / `applyHolds`: exercise changing scope through the session fake;
  scoped rewrites that would drop existing holds refuse before writes, directing --all
  or explicit release; no hidden hold store.
- [ ] `planResume` / command flags: adversarial ownership resolves to own-only rows or
  summary counts by default; --all exposes all rows without foreign candidate/resume calls.
- [ ] `gatherSlots` / `fleetResume`: failed actors seam produces one diagnostic and unknown
  rows, with no mutation; individual claims errors retain their own reasons.

Run `go test ./cmd/tl-lead -run 'Test(HubScope|GatherSlots|StatusSlots|FleetResume)' -count=1`
red, then green. Implement one shared scope adapter and wire rendering/flags before committing.

### Task 2: freshness shares scope

**Files:** `cmd/tl-lead/freshness.go`, `freshness_test.go`; reuse `scope.go`.

- [ ] `gitNewestLanding`: real-Git multi-hub fixtures with adversarial newer unrelated
  histories prove only claimed repo merges affect freshness; preserve first-parent,
  brain and direct-publication exclusions. Scope errors remain unknown, not global fallback.
- [ ] `staleLines`: stamp/merge comparisons exercise the scoped landing result through
  the session fake; freshness remains read-only.

Run `go test ./cmd/tl-lead -run 'Test(GitNewestLanding|.*Stale)' -count=1` red, then green;
apply shared scope before Git history reads and commit.

### Task 3: serial resume with terminal evidence

**Files:** `cmd/tl-lead/fleet.go`, `fleet_test.go`, plus a focused receipt test file if needed.

- [ ] `classifyResumeResult`: malformed/truncated/contradictory child responses and exit
  modes must never prove completion; table/fuzz strategies enforce the confirmation rule below.
- [ ] `execResume.Resume`: stateful fake executable uses completion barriers to prove
  synchronous waiting and stdout/stderr retention through the real subprocess seam.
- [ ] `fleetResume`: stateful fake time models 10–30-second operations; completion barriers
  reject any next-slot attempt before completion. A first unsuccessful result stops with
  nonzero status, retained earlier successes, exact untouched list, and no automatic retry.
  Read-only actor recheck failure/changes cannot promote uncertain receipt evidence.
- [ ] `newFleetCmd`: assert operator-started/never-at-startup help remains present.

Run `go test ./cmd/tl-lead -run 'Test(FleetResume|ExecResume|ClassifyResume)' -count=1`
red, then green. Keep Couch as the polling owner, implement conservative classification,
retain diagnostics, stop on every unsuccessful outcome and commit. These semantic protocol
properties use state/subprocess tests, not a lexical lint rule guessing from loop syntax.

### Confirming evidence (PQ-1)

`execResume.Resume` requests `couch --resume ADDR --json`; Couch already supports this
form in `cmd/internal/couchcmd/cli.go:378-413` and emits one terminal OperationReceipt
from `messages.go`'s `reportSlotOperation`. Pure `classifyResumeResult` consumes the expected
address, stdout, stderr, whether the process started, normal/signal exit and timeout facts.
A known process-start failure means not attempted. Otherwise confirm only a complete JSON
receipt with nonempty ID, Op=resume, matching Target, recognized terminal Status and a
consistent normal exit: succeeded with zero; refused/failed with nonzero. Everything else
(including signal death, timeout, absent/unrecognized/truncated output or inconsistent exit)
is uncertain. Retain both streams, request ID when present, and manual inspection guidance.
Unknown JSON fields are tolerated because this reader never rewrites Couch's receipt.
No second receipt poller is introduced. Actor observation is diagnostic only.

## Task 4: docs, verification and live proof

**Files:** update `cmd/tl-lead/help.md`, `atlas/workflow/tl-skills.md`,
`atlas/workflow/sandbox.md`; update atlas index only if adding a page.

- [ ] Document own-only defaults, `--all` as visibility rather than expanded mutation
  authority, unclaimed counts, no-hub compatibility and runtime ownership limitations.
- [ ] Document the bare `tl-lead` invocation required for Claude sandbox exclusions.
  Codex has no equivalent command exclusion; the live dry run in this session worked.
  Do not claim that spawning a child can escape a parent sandbox or change harness settings.
- [ ] Document serial resume, stop/nonzero behavior, partial completion and inspection paths.
- [ ] Run `go test ./cmd/tl-lead -count=1` and relevant `pkg/tlrules`/catalog checks if touched.
  Do not commit while source-checkout tests are active.
- [ ] Push with `sdlc pr`; queue `sdlc local-check run --name suite` on the pushed head.
  Close via `sdlc close --issue 485 --verified '<actual evidence>'`; the binary owns review.
- [ ] Coordinate `sdlc stage --build` with the hub before touching :0. Ask the hub to
  verify each live hub's scoped Slots and stale check; compare dry-run `fleet resume`
  with `--all`. No unsolicited live `--apply`: use the fake to prove mutation ordering
  and perform any real resume only with the hub/operator's chosen down slots.
- [ ] After successful verification and SHIP, run `sdlc merge --yes`, then report outcome.

## Architectural envelope

- ARCH-DRY/PURE: one scope adapter, reuse ownership and Couch polling; pure selection stays separate from reads.
- ARCH-PURPOSE: own-work visibility and honest partial-batch recovery; no new ownership service.
- ARCH-MOCK: stateful fakes plus real subprocess/Git fixtures; live read-only conformance.
- ARCH-CONSTRAINTS: per render O(slots × existing ownership comparisons), O(selected busy slots)
  issue calls; freshness scans sibling directory entries once, Git only for scoped SDLC repos.
  Resume remains one heavy operation at a time. Count N/2N reads; do not rely on timing for growth.
- ARCH-SECURE/ORDER: ownership ambiguity excludes mutation; receipt uncertainty stops; actor presence never proves request completion.
- ARCH-FUNERAL: creates no new durable runtime objects; existing temporary-body cleanup, section history
  and Couch receipt retention remain authoritative. Test files use t.TempDir.
- ARCH-EVOLVE: no persisted-format change. Refuse scoped rewrites that would discard a hidden hold.

## Revisions

- 2026-10-20 13:44 PDT — TL approved this plan (Couch delivery bac6759f-4e2e-4a0d-a272-4c8409efe48e); the proposed approval label above is historical. All ownership, terminal polling, stop/nonzero and hidden-hold decisions are accepted. Add explicit acceptance: fleet resume remains operator-started and never runs automatically at startup; retain that wording in --help and test it. No further operator review is required for these settled recovery behaviors.

- 2026-10-20 13:47 PDT — plan-quality PQ-1: define pure classifyResumeResult over the existing Couch --json terminal receipt, with absent/mismatched/nonterminal/abnormal evidence uncertain. PQ-2: compress Tasks 1–3 to named functions and adversarial strategies; scope and approved visible recovery states unchanged.

- 2026-10-20 14:06 PDT — BR-1: actor observation uses a 10-second command context and one-second pipe cleanup; stopped/untouched output and inspection guidance precede it. BR-2: modeled synctest sleeps carry reasoned catalog annotations. BR-3: README documents both --all flags. BR-4: isolated real-Couch receipt conformance is opt-in, run at tl-lead closes and after Couch receipt-protocol changes; it never targets the live supervisor. Publication gate ordering moved out of issue implementation checkboxes, preserving suite/live/review requirements before landing.
