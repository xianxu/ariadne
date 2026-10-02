---
id: 000278
status: working
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '6bd7daaa803b59bf95092eced08f9f98e1d40935' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T17:25:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3eac6248", done: "db328d5b"}
---

# Add operator-directed reclaim for recovery

## Problem

After a crash or deliberate relocation, responsibility can need reassignment. Neither unreachable-machine inference nor the publishing-oriented sdlc move operation expresses an operator-authorized transfer.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Add sdlc reclaim as an operator-directed recovery operation assisted by an LLM. Read and show current and proposed operator/machine/slot/worktree assignment; perform a guarded tracker update against the observed revision. Record the operator’s recovery reason and transfer in Git history. Exact CLI shape is design work, not settled here.

Do not infer abandonment from a timeout, shutdown, parking or unreachable machine. Operators coordinate out of band, then issue reclaim on one machine. Do not overload sdlc move or automatically transfer responsibility as a side effect of publishing movement. Reclaim changes assignment; it must not reset worktrees, replay prompts, kill remote workers or pretend to fence off an old machine. Specify how an identical retry and a lost publication response are reconciled from tracker evidence.

## Done when

- An operator can inspect and reclaim a recorded assignment, with the old and new responsibility and reason discoverable afterward.
- Concurrent/stale reclaim attempts cannot silently overwrite a newer assignment; lost-response recovery observes the authoritative result.
- No automated crash/reachability path invokes reclaim, and publishing-oriented sdlc move remains distinct.
- Tests demonstrate wrong-owner resume refusal after reclaim and preservation of dirty work; help explains out-of-band operator coordination.

## Plan

Durable plan: `workshop/plans/000278-operator-reclaim-plan.md` (single pass).

- [x] Pure `reclaimDecision` and the trailers; `UpdateCard` message trailers.
- [x] `sdlc reclaim`: inspect (read-only), then confirm with `--expect` and
      `--reason` (CAS; a retry decides from the card).
- [x] Real-git tests: transfer and wrong-owner refusal, dirty work preserved,
      stale expect, identical retry and lost response, concurrent reclaim,
      history; a no-automation guard.
- [x] Docs: reclaim help, #277 refusal pointers, atlas.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-01 (implementation session)

- Operator authorized the work ("work on #278"). Claimed in ariadne:1 with
  :0's pre-#277 binary, deliberately: a claimant card would start the #277
  flag day before the fleet is rebuilt. This card is therefore unattributed
  and will need `--adopt` after the rollout. Ran start-plan; the branch
  `000278-operator-reclaim` sits at main, which includes #277.
- Tension to resolve in design: the Spec says "do not overload sdlc move",
  but #277 (operator decision) already has move relocate the owner's *own*
  work on the same machine, using positive move evidence. Reclaim is the
  general, operator-directed transfer across workspaces and machines; move's
  relocation stays the narrow same-owner case.
- Operator decisions:
  - **CLI shape:** inspect, then confirm. `sdlc reclaim --issue N` shows the
    current owner, the proposed owner and past reclaims, and prints the
    confirm command. `--expect <card-rev> --reason '...'` performs it as a
    CAS against that revision.
  - **New owner:** only the running workspace.
- side-quest: corrected #277's rollout procedure in the atlas and README.
  Each environment builds `sdlc` from its own ariadne checkout, so the
  procedure is: update ariadne:0, then `weave refresh` in every :1+ slot.
  The earlier `weave compile` / `make weave-all` advice would not have
  advanced any Git revision. The operator is running the refresh.

- **Correction** to the claim note above. The claim did *not* run pre-#277
  code. `sdlc merge` for #277 had already fast-forwarded `main-slot1` to
  include #277 (17:19), and `sdlc` builds from this slot's own checkout. So
  the 17:25 claim wrote the first `claimant` card (tracker `4be6376a`), and
  the #277 flag day began then, not at a later claim.
  - The operator was told and has since refreshed the fleet: every live slot
    is on #277 (checked with `couch --actors`). The idle ariadne-slot2,
    ariadne-slot3 and pair-slot2 still need `weave refresh` before reuse.
  - `sdlc claim --adopt` just now was a no-op; the card already names this
    workspace.
- Wrote the durable plan (single pass, no milestones). Awaiting operator approval.
- Operator approved the plan. change-code inferred the quick flow (no plan-quality, no estimate); close will upgrade to the full review if the diff leaves the shell. Implementing.
- Done: `UpdateCardWithTrailers` (tracker commit trailers; plain messages unchanged) and the pure `reclaimDecision` plus trailer round trip. Both are table-tested.
- Done: the `sdlc reclaim` command (inspect/confirm, history from trailers, lost-response message, mirror refresh) and `helptext/reclaim.md`. Next: real-git tests and the no-automation guard.
- Done:
  - Real-git tests: transfer and wrong-owner refusal, dirty work preserved,
    stale expect, identical rerun and lost response, a two-clone race, and
    history.
  - The no-automation AST guard. Its mutation check (a `move.go` reference)
    goes red.
  - The #277 refusal texts and help now name `sdlc reclaim`.
  - Atlas: a Reclaim subsection plus the verb row.
- Regenerated the process manual (reclaim help). Full sharded suite: all cmd/sdlc tests pass; only the sandbox-only processgroup failure remains. Closing.
- Close review round 1 returned FIX-THEN-SHIP. Fixed every finding:
  - BR-1: the old workspace's start-plan, change-code and close are each
    refused after reclaim, before any review. The new owner's start-plan
    passes once the old worktree lets go of the branch.
  - Minors:
    - an "already yours" rerun refreshes the details mirror (lost-response
      case; mutation-checked);
    - history is bounded to the last 20 reclaims, selected by trailer, and the
      help and plan say so;
    - `--reason` without `--expect` refuses;
    - the guard now also scans package-level initializers (mutation-checked).
