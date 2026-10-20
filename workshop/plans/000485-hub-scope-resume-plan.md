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

## Task 1: shared scope and default displays

**Files:** create `cmd/tl-lead/scope.go`, `scope_test.go`; modify `slots.go`, `status.go`,
`fleet.go`, `fleet_test.go`, `trains_test.go`, `status_test.go`.

- [ ] Add failing pure scope cases: two hubs, exact override, tie involving self,
  unclaimed slots, foreign hub :0, absent `.hub`, and malformed `.hub`.
- [ ] Add command/session-fake cases proving own-only Slots by default and full Slots
  with `--all`; preserve hold/release semantics, including holds hidden by filtering.
  If a scoped rewrite would drop an existing held address, refuse before writing and
  name `status slots FILE --all` (and explicit `--release` when appropriate). No new
  hidden representation or second hold store is introduced.
- [ ] Run `go test ./cmd/tl-lead -run 'Test(HubScope|GatherSlots|StatusSlots|FleetResume)' -count=1`;
  confirm failures identify missing filtering/flags, then implement using `ownerOf`.
- [ ] Filter before per-issue reads. Add counted N/2N inventory cases proving one
  issue observation per selected busy slot and zero observations for excluded slots.
- [ ] Add `--all` to Slots and resume; show excluded foreign rows without consulting their
  candidates or resuming them. Preserve no-hub status behavior and explicit resume refusal.
- [ ] Make actors-unavailable a single top-level diagnostic, with rows `?`/`unknown`
  and no per-row repetition. Failures reading individual claims retain their own reasons.
- [ ] Rerun targeted checks and commit the coherent scope/display change.

## Task 2: freshness shares hub scope

**Files:** modify `cmd/tl-lead/freshness.go`, `freshness_test.go`; reuse scope tests.

- [ ] Extend real-Git `TestGitNewestLanding` with two hubs and an unrelated newer merge.
  Assert only claimed repositories affect freshness; exact-slot claims include their repo.
  Assert no-hub fallback, broken local `.hub` and claimed repos with missing refs.
- [ ] Run `go test ./cmd/tl-lead -run 'Test(GitNewestLanding|.*Stale)' -count=1`;
  prove the unrelated merge fails the old implementation.
- [ ] Apply shared scope before per-repository Git history reads; preserve existing
  first-parent merge semantics, brain exclusion and direct-publication exclusion.
- [ ] Rerun targeted checks and commit.

## Task 3: terminal serial resume and explicit partial completion

**Files:** modify `cmd/tl-lead/fleet.go`, `fleet_test.go`; add a focused resume test file
if the existing one becomes unwieldy.

- [ ] Extend stateful fakeCouch with pending/completed transitions, call order and
  completion barriers. Model 10/20/30-second durations with fake time; prove that
  slot N+1 is not attempted before N completes, rather than merely checking final order.
- [ ] Add stop cases for first/middle refusal, explicit uncertainty, local timeout,
  failed actor recheck and a changed actor view. Assert nonzero, exact untouched list,
  earlier successes retained, no retries and no later resume/candidate mutation.
- [ ] Add subprocess conformance using a fake executable that blocks until a test-controlled
  barrier, then succeeds/refuses/emits uncertain diagnostics including a request ID.
  It exercises the same execResume seam and proves waiting and retained output.
- [ ] Run `go test ./cmd/tl-lead -run 'Test(FleetResume|ExecResume)' -count=1` and confirm
  old continue-on-error behavior fails before changing the runner.
- [ ] Keep CLI-owned polling; wait synchronously, retain stdout/stderr, distinguish
  timeout/explicit uncertainty from confirmed refusal, and stop after every error.
  On uncertainty, read actors once and report observation without declaring success.
  Print exact read-only/manual recovery commands and return nonzero.
- [ ] Set the documented outer timeout, retaining injectable short test durations;
  include the default-budget relationship in a regression assertion.
- [ ] Rerun targeted checks and commit. These are semantic protocol guarantees;
  integration/state tests enforce them rather than a lexical lint rule that guesses from loop syntax.

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
