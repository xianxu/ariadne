---
collaboration-mode: discussion
---

# Merge Intents Survive a Full Shutdown: Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to choose between superpowers-subagent-driven-development and superpowers-executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** after a full shutdown, every slot whose `sdlc merge` was in flight is named in `sdlc landing status` and `sdlc state`, each with one exact re-entry command. Re-entering that command lands on the same head and reuses suites that already passed.

**Architecture:**
- A durable **merge intent** record per branch lives in the landing directory.
- The running `sdlc merge` writes it and holds a flock on it for its lifetime. Only a terminal outcome removes it; an interruption or a kill leaves it behind.
- A reader classifies a record as live (locked) or orphaned (unlocked) with a pure function.
- Status surfaces print an orphan's re-entry command.
- Re-entry itself needs no new machinery. `sdlc merge --branch B --yes` already resumes from Git and GitHub evidence, and the #332 gate reuses a passing `sdlc-local/suite` commit status on an unchanged head.

**Tech Stack:** Go (`cmd/sdlc`), flock (`internal/landingline` patterns), JSON records.

---

## What exists (measured 10-23)

- **Nothing on the landing line survives a shutdown, by design.** Tickets and `head.lock` are flock-only, and `scan()` prunes dead tickets (`internal/landingline/line.go:501`). The atlas says "Nothing is left that a crash can make inconsistent".
- **The queue op survives a Couch restart, but not as a re-entry record.**
  - It survives as `unresolved/<id>.json` (phase `unknown`), but only once something observes it (`internal/queuegate/store.go:435`).
  - It has no branch and no re-entry command.
  - A machine reboot deletes it (`store.go:458`).
- **Re-entry is already correct.**
  - `sdlc merge --branch B --yes` enrolls a fresh op and ticket.
  - `resumeOnly` (`landingqueue.go:273`) skips the line for an already-merged PR, and `integrationResume` (`landingintegrate.go:53`) picks up an unpushed integration.
  - The suite is reused when `decideCIGate` finds a passing `sdlc-local/suite` status on the head (`cigate.go:640`). That status lives on GitHub, so it survives any shutdown.
- **The recovery command is printed today only to the dead process's stderr** (`landing.go:485`).

So the gap is the record and the surface, not re-entry.

## Scope choices (for the TL)

- **(a) Record:** a per-branch intent file, `<git-common-dir>/sdlc/landing/intents/<branch-key>.json`. It is written at merge start and flock-held by the runner.
  - I'm not reconstructing intents from `events.jsonl`: those events are a 30-day metrics log with no lock, and they cannot tell a live runner from a dead one.
  - I'm not using the queue op either: it has no branch, and a reboot erases it.
- **(b) What ends an intent:**
  - A **terminal outcome** removes it: landed, or refused for any reason other than interruption. The slot saw that error, so nothing is left to re-enter.
  - An **interruption** keeps it: cancellation, SIGTERM/SIGHUP from a Couch shutdown, or `queueInterruptionReason(err) != ""`. A kill -9 or power loss also keeps it, because the process never returns.
  - Without this split, a graceful Couch shutdown would erase exactly the intents this issue exists for.
- **(c) Surfaces:**
  - `sdlc landing status` lists every orphaned intent in the repo.
  - `sdlc state` (the post-compaction/restart primitive a worker already runs) prints the current slot's own orphan.
  - The `--json` of both carries the same fields, so Couch's `--recover-plan-from-sdlc` restart summary (#564) can read them without parsing prose.
  - The Couch side itself is out of this repo.
- **(d) A second runner for the same branch:** re-entry takes the lock and rewrites the intent with its own pid and head. If another runner already holds the lock, the new `sdlc merge` refuses, naming the live pid. Today a second runner would queue a second ticket for the same branch, which is never what anyone wants.
- **(e) Paused queue:** when the machine queue is paused (the shutdown freeze), the orphan line says so: "the queue is paused by X; re-enter after it resumes". The slot then doesn't have to decide whether the freeze still applies.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `MergeIntent` | `cmd/sdlc/internal/landingline/intent.go` | new |
| `ClassifyIntent` | `cmd/sdlc/internal/landingline/intent.go` | new |
| `intentReentry` | `cmd/sdlc/mergeintent.go` | new |

- **MergeIntent**: `{version, branch, issue, head, workspace, root, pid, pid_started, boot_id, started_at}`, plus preserved unknown members.
  - **Relationships:** 1:1 with a branch while it is in flight; N per repo.
  - **ARCH-EVOLVE:** the decoder keeps unknown keys in a raw map and writes them back unchanged on rewrite. Known fields stay strictly typed, duplicate keys are refused, and branch, head and root are required. Tests cover an older reader rewriting a newer record.
  - **ARCH-SECURE:** the file sits under a shared directory, so it is parsed into a typed value at the boundary. The branch must pass `git check-ref-format`, the head must be a 40-hex SHA, and the root must be absolute. An invalid record is reported as "unreadable intent <file>: <why>" and never yields a command. Every printed field is `%q`-quoted. The re-entry command is built from the validated branch, never copied from the file.
- **ClassifyIntent(intent, locked bool, branchExists bool) → {Live | Orphaned | Stale}**:
  - Live: the lock is held.
  - Orphaned: the lock is free and the branch still exists.
  - Stale: the lock is free and the branch is gone, so the work landed or was deleted elsewhere.
- **intentReentry(intent, paused \*pauseOwner) → (command, note)**: in the slot at `root`, run `sdlc merge --branch B --yes`; the note covers the pause.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `IntentStore` | `cmd/sdlc/internal/landingline/intent.go` | new | flock + files under `landing/intents/` |
| `runDurableMergeAttempt` | `cmd/sdlc/landing.go` | modified | holds the intent for its lifetime |
| `runLandingStatus` / `runState` | `landingcmd.go`, `state.go` | modified | print and serialize orphans |

- **IntentStore**: `Hold(intent) (release func(outcome), err)` and `List() ([]observed, []unreadable)`. It uses the real filesystem in a temp dir: flock is local and deterministic, so the stateful fake *is* the real store at a temp root (ARCH-MOCK). The tests kill a real child process that holds the lock to simulate a shutdown.

## ARCH-ORDER: intent lifecycle

| State | Event | Next | Effect |
|---|---|---|---|
| none | merge starts | live | write the record, then flock it; a lock already held means refuse, naming the live pid (d) |
| live | landed / refused (not interruption) | none | remove the file, then unlock |
| live | interrupted (ctx cancelled, signal) | orphaned | unlock and keep the file |
| live | process death (kill -9, power) | orphaned | the kernel drops the flock; the file stays |
| orphaned | re-entry (`sdlc merge --branch B`) | live | take the lock, rewrite pid, head and started_at, keep unknown members |
| orphaned | branch deleted or landed elsewhere | stale | `landing status` sweeps it (removes the file) after a try-lock |
| any | unreadable record | — | reported and never removed automatically; the operator deletes it (the path is named) |

- **The event most likely to be mishandled** is the graceful interruption. A shutdown's SIGTERM unwinds through the same deferred cleanup as a refusal, so the release callback must take the outcome and decide from `queueInterruptionReason`.
- **Ordering of the write:** write to a temp file, take the flock, then rename into place. A reader that finds a file without a lock therefore never sees a half-written record. Alternatively, open and flock first and write in place; I'll pick whichever `landingline` already uses.
- **Lock order:** two runners racing on the same branch are serialized by the flock. The loser refuses, so there is no duplicate ticket.

## ARCH-FUNERAL

- **Creator:** `sdlc merge`.
- **Remover:**
  - the terminal outcome;
  - re-entry rewrites the record in place, so it doesn't accumulate;
  - the stale sweep (branch gone) in `landing status` and in `sdlc merge` start;
  - an age bound of 30 days, matching `events.jsonl`, for an orphan nobody re-enters.
- **Size:** about 300 bytes per in-flight branch, so a few kilobytes for a fleet.

## ARCH-CONSTRAINTS

- **Status path:** a directory listing plus one try-lock per file, and one `git show-ref` per orphan. Fewer than 20 files means well under 50 ms.
- **Merge path:** one small file write and one flock. Both are negligible.

## Tasks

### Task 1: `MergeIntent` codec and `ClassifyIntent` (pure)

- [ ] **Write failing tests** in `internal/landingline/intent_test.go`:
  - a decode/encode round trip that preserves unknown members (including nested ones);
  - refusals of a bad branch, a short head, a relative root, a duplicate key and a mistyped field;
  - the full `ClassifyIntent` table.
- [ ] **Implement and pass.** Commit `#565: merge intent record and classification`.

### Task 2: `IntentStore` (hold, release by outcome, list)

- [ ] **Write failing tests:**
  - Hold, then List shows it Live.
  - A terminal release removes the file.
  - An interrupted release keeps it as Orphaned.
  - A **child process** holding the intent is SIGKILLed, and List shows it Orphaned (the shutdown simulation).
  - A second Hold on a held branch refuses with the holder's pid.
  - A re-Hold of an orphan keeps unknown members.
  - The stale sweep removes it once the branch is gone.
  - The age bound removes a record older than 30 days.
- [ ] **Implement and pass.** Commit.

### Task 3: `sdlc merge` holds the intent

- [ ] **Write failing tests** in `cmd/sdlc/mergeintent_test.go`, using the existing landing fakes (`landing_test.go` harness):
  - a landed merge leaves no intent;
  - a CI-red refusal leaves none;
  - a merge whose context is cancelled mid-wait (`testfix.DeadlineOnCue`) leaves an Orphaned intent with the branch and head;
  - a concurrent second merge on the same branch refuses.
- [ ] **Wire it in.** In `runDurableMergeAttempt`, after the branch is resolved and before queue enrollment, call `Hold`. Its release goes in the existing deferred block, with an outcome from `queueInterruptionReason(err)`.
- [ ] **Pass.** Commit.

### Task 4: surfaces (`landing status`, `sdlc state`, both `--json`)

- [ ] **Write failing tests:**
  - After a SIGKILLed runner, `sdlc landing status` prints `orphaned merge: #565 branch "B" head abcd1234 (runner pid P gone since T); re-enter in /path/slot: sdlc merge --branch B --yes`.
  - With the queue paused, the line adds the pause note.
  - `sdlc state` in that slot prints the same line for its own branch only.
  - The `--json` field `merge_intents[]` carries `{branch, issue, head, root, state, reentry}`.
  - An unreadable intent is reported with its path and no command.
- [ ] **Implement and pass.** Commit.

### Task 5: re-entry reuses the passed suite (Done-when 2)

- [ ] **Write the end-to-end test:**
  - A recorded passing `sdlc-local/suite` status sits on head H in the fake GitHub.
  - The merge is cancelled mid-landing, leaving an orphan.
  - The test then runs the printed command.
  - It asserts that the suite command never ran (its marker file is absent), that the PR landed, and that the intent was removed.
  - A control case: the head changed, so the suite reruns.
- [ ] **Pass.** This should need no production change; if it needs one, I log why. Commit.

### Task 6: atlas and verification

- [ ] **Atlas:** in `atlas/workflow/sdlc-binary.md` "Landing line (#393)", add a "Merge intents (#565)" paragraph. Amend the "nothing is left behind" sentence: the intent is the one durable record, and it is lock-classified.
- [ ] **Recovery catalog:** in `internal/recovery/catalog.go`, the merge LostResponse names `sdlc landing status` as the place to find the command.
- [ ] **Verify:** targeted tests, then the recorded suite on the pushed head.

## Out of scope

- Couch reading the `--json` field (Couch repo; #564 owns the restart summary).
- Non-suite landing-check proofs. They are held in memory by design (`landingchecks_gate.go:12` refuses stored statuses as unbound), so they rerun on re-entry. That is correct and cheap.
- Automatic re-entry. The spec says no.

## Revisions

- 2026-10-23: TL approved scope choices (a)–(e) as written. Follow-up: an unclaimed pair issue for Couch's recover-plan reading `merge_intents`.
- 2026-10-23 (implementation): four details differ from the plan.
  - The intent has no `pid_started` or `boot_id`. Liveness comes from the flock alone, which dies with the process on a kill or a reboot, so neither field would be read.
  - The record is written in place under its lock, not by temp file and rename. A rename would move the path to a new inode, away from the lock. A reader that finds the file locked treats it as live, even half-written.
  - `sdlc landing status` has no JSON mode, so `merge_intents` lives in `sdlc state --json` only, scoped to the slot. Couch's restart plan reads it per slot, and a status JSON carrying intents alone would be half a format.
  - The live-intent refusal is `IntentHeldError`; the line package already has a `HeldError` for train holds.
