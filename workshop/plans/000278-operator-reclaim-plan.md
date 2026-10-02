# Operator-Directed Reclaim Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc reclaim` lets an operator, after coordinating out of band, move an issue's recorded responsibility (#277's `claimant`) to the workspace that runs it. It is guarded against stale and concurrent updates. It leaves the old owner, the new owner and the reason discoverable in the tracker's history.

**Architecture:** There are two steps, both headless.
- **Inspect.** `sdlc reclaim --issue N` only reads. It shows the card's current owner, the proposed owner (this workspace), the card revision, and past reclaims. It prints the exact confirm command.
- **Confirm.** `sdlc reclaim --issue N --expect <rev> --reason '<why>'` runs a pure decision against the card it reads. It publishes by compare-and-swap on that same card.
  - `--expect` pins the revision the operator inspected, so any change since then refuses.
  - The tracker commit carries `Reclaim-From` / `Reclaim-To` / `Reclaim-Reason` trailers. That history is where the record lives.
  - A retry, including one after a lost publication response, is decided from the card. It is a no-op success when the card already names this workspace.

**Tech Stack:** Go (`cmd/sdlc`), the tracker CAS (`tracker.Repository.UpdateCard`), real-git tests on the #277 harness (`closeReady`, `withClaimant`, `raceBuiltBinary`, `ownershipGate`).

**Operator decisions (2026-10-01, issue Log):**
- Inspect then confirm, pinned by `--expect`.
- The new owner is only the workspace running reclaim, whose identity is resolved locally.

## Non-goals

- No inference of abandonment from timeouts, shutdown, parking or reachability. Nothing but the `reclaim` command invokes the transfer, and a test guards that.
- No change to worktrees, branches or uncommitted work anywhere. No killing or fencing of an old worker. No replaying prompts.
- `sdlc move` keeps only its #277 same-owner relocation. It gains no reclaim semantics.
- Unattributed cards stay with `claim --adopt` (#277). Reclaim needs a recorded owner to transfer from.
- Structured observation of assignments is #279. Inspect is a human view.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `reclaimDecision` | `cmd/sdlc/reclaim.go` | new |
| `reclaimTrailers` / `parseReclaimTrailers` | `cmd/sdlc/reclaim.go` | new |

- **reclaimDecision(card []byte, rev, expect, reason string, me issue.Claimant) (next []byte, from issue.Claimant, err)** decides:
  - status `working`, `blocked` or `codecomplete` is required. `open` points to claim; a terminal status has nothing to own.
  - a recorded claimant is required; without one, use `--adopt`.
  - Mine gives `errAlreadyMine`. That is the identical-retry and lost-response reconciliation: the card already shows the result.
  - `rev != expect` gives a stale refusal naming the current revision.
  - an empty or multi-line reason is refused.
  - otherwise the result is the card with `me` as claimant, plus the previous claimant for the trailers.

  It is pure and table-tested over status × ownership × expect × reason.
  - **Relationships:** one transfer per call. It reuses `issue.SetCardClaimant` and `MatchClaimant`.
  - **DRY rationale:** the claim, adopt, set-status and reclaim decisions all stamp through `SetCardClaimant`. Reclaim is the only one that may replace a Foreign owner, and only with `expect` and `reason`.
- **reclaimTrailers(from, to, reason)** returns `Reclaim-From: <describeClaimant>`, `Reclaim-To: …` and `Reclaim-Reason: …`. The parser reads them back for inspect's history. One function is used for both writing and reading.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `NewReclaimCmd` / `runReclaim` | `cmd/sdlc/reclaim.go`, `helptext/reclaim.md` | new | tracker snapshot + CAS, identity seam |
| `Repository.UpdateCard` (+ message trailers) | `cmd/sdlc/internal/tracker/repository.go` | modified | gitx `UpdateManyPrepared` |
| `reclaimHistory` | `cmd/sdlc/reclaim.go` | new | `git log` of the card path on the snapshot's tracker ref (the last 20 reclaims, selected by their trailer) |

- **UpdateCard** gains an optional message detail: trailers appended after `Tracker-Operation`. Existing callers pass none, so their message is unchanged.
- **runReclaim** works as follows:
  - It opens the tracker, reads one fresh snapshot, and computes `rev` = the card's blob OID.
  - With no `--expect` it prints the inspect view and writes nothing.
  - With `--expect` it requires `--reason`, runs `reclaimDecision`, then `UpdateCard(card, next, token "reclaim", trailers)` against the card it read.
  - `ErrCardChanged` means a concurrent change: refuse and re-inspect.
  - `ErrPublicationUncertain` means the response was lost: tell the operator to rerun the same command. The rerun decides from the card, giving a no-op if the reclaim landed and a CAS retry if it didn't.
  - After success it refreshes the local details mirror (never on rest) and prints the transfer.
  - Mutating command: it takes the repo lock (`markMutatingCommand`).

### Lifecycle and ordering (ARCH-FUNERAL / ARCH-ORDER)

- It creates nothing durable except one tracker commit per reclaim. That commit is the record, bounded by use, and lives as long as tracker history. There is no local state.
- Ordering: the decision and the CAS key on the same card read. `--expect` pins the operator's earlier observation. A concurrent reclaim, claim or set-status changes the blob, which refuses the loser. A retry decides from the authoritative card.

## Steps (single pass; one close)

- [x] **Pure decision and trailers (TDD).** `reclaim_test.go` holds two tables:
  - `reclaimDecision` over {open, working, blocked, codecomplete, done} × {mine, foreign, unattributed} × {expect matches, stale} × {reason ok, empty, multi-line};
  - a trailer round trip, including a reason containing `:` and unicode.
- [x] **Tracker message trailers.** Extend `UpdateCard`. A test asserts that the commit message carries the trailers and that existing callers' messages are byte-identical.
- [x] **Command.** Add `runReclaim` with inspect and confirm, `helptext/reclaim.md`, and registration on the root command.
- [x] **Real-git tests** (`reclaim_test.go`, on the #277 harness):
  1. Inspect writes nothing (the tracker tip is unchanged) and prints the current owner, the proposed owner, the rev and a confirm command that works when run.
  2. Confirm transfers. The old workspace's start-plan / change-code / close then refuse as Foreign, and the new one passes: wrong-owner resume refusal.
  3. Dirty files in both worktrees are byte-identical before and after.
  4. A stale `--expect`, after an intervening card change, refuses and leaves the card unchanged.
  5. An identical rerun after success is a no-op with no new tracker commit. A simulated lost response is handled the same way: the CAS publishes, the error is injected, and the rerun reconciles.
  6. Two clones each inspect the same rev and confirm concurrently (`raceBuiltBinary`): exactly one wins and the loser refuses.
  7. History: after two reclaims, inspect lists both transfers with their reasons, read from the trailers.
- [x] **No-automation guard.** An AST test asserts that the reclaim effect is reachable only from `NewReclaimCmd`, modeled on `context_guard_test.go`'s walk. It also asserts that `move.go` does not reference reclaim.
- [x] **Docs.**
  - `reclaim.md` help covers out-of-band coordination, the two steps, the guarantees (stale/concurrent/retry), what reclaim does not do, and the instruction that an agent runs it only on the operator's direction.
  - Point the #277 refusals ("reassignment is operator-directed reclaim (#278)") to `sdlc reclaim`.
  - Add an atlas `issue-tracker.md` Reclaim subsection and a verb-table row. Add `atlas/index.md` if needed.
  - `TestEveryFlagAppearsInItsHelp` covers the new flags.
- [ ] Close.

## Revisions
