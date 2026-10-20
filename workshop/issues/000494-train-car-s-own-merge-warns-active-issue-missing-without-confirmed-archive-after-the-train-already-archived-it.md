---
id: 000494
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c12c8f52b22f7094cfeff79316d8ec78f48d375e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T06:34:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "a3b7f1ea", done: "a427ade0"}
---

# Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it

## Problem

In pair's first multi-car train (176, 10-20 05:48), car pair#717's own `sdlc merge` (PR 483) ended with "active issue
missing without confirmed archive", although the train had already archived the issue (it's under history on main
and its card is done). Harmless, but it reads as an error at the end of a successful landing, and a worker can't tell.

## Spec

When a car's merge finishes after the train landed it, treat "archived by the train" as the confirmed outcome (the
train's `outcomes/<seq>` record says so), and print it as such, not as a missing-issue warning.

Retain the validated landed outcome for this exact waiting ticket through cleanup. It is context for the confirmation, not proof that archival happened: the train publishes it before cleanup. Keep the ordinary per-PR provenance path unchanged. Only for this confirmed train outcome, and only for tracked completions owned by the car's exact PR, allow an already-published sibling archive to be discovered from bounded history of its destination path on fresh main. Reuse the existing archive-commit verifier: one ordinary commit, the exact issue generation and done-card binding, complete planned issue/plan/asset/side moves, unchanged current archived blobs, and no active residue. Missing, ambiguous, malformed or changed evidence still fails; never infer success from a missing active issue or OutcomeLanded alone. No new persisted fields, archive writes, retries, authority bypass or conductor scheduling changes.

| Observed state | Visible result | Recovery | Possible loss |
|---|---|---|---|
| Own PR archive provenance confirms completion | Existing successful cleanup | None | None |
| Confirmed train outcome plus exact sibling archive transition | “archived by the train” with train number | None | None |
| Train outcome but active details remain | Existing ordinary archive/cleanup path | Wait for completion, or retry the printed merge command after a failure | None; preserve checkout on failure |
| Outcome absent, invalid, or archive proof missing/changed | Existing explicit error; no success claim | Reconcile the named evidence, then retry the printed merge command | None; no cleanup authorized by uncertain proof |

ARCH-DRY/PURE: extract and reuse the existing archive transition verifier, with IO in the history lookup. ARCH-ORDER/SECURE: outcome confirms landing only; immutable Git/card proof confirms archive, and only then permits local cleanup. ARCH-MOCK: real Git plus the existing stateful train GitHub fake, with a tracked two-car fixture and deterministic cleanup ordering. ARCH-CONSTRAINTS: at most the existing archive history limit per selected issue; this is terminal cleanup, no polling/hot-path work. ARCH-EVOLVE/FUNERAL: no format changes or new durable runtime artifacts.

## Done when

- A tracked 2-car train fixture: after the conductor archives the second car through its done-card sweep, the second real merge ends with “archived by the train”, no warning, and completes safe checkout/ref cleanup.
- Missing, ambiguous, malformed, tampered or wrong-generation archive evidence remains a failure, and ordinary serial archive confirmation is unchanged.

## Plan

- [x] Reproduce the tracked two-car cleanup ordering and confirm archive provenance.
- [x] Retain the train outcome and recognize its already archived exact completion without weakening serial archive checks.
- [ ] Verify focused controls, recorded suite, fresh close review and landing.

## Log

### 2026-10-20

- Investigation: pair landing events record train 176 landing tickets 176/178 (pair#724/#717) with train PR #484. Git history identifies #717 archive commit 7d4e0c3ab as the generic “archive after a merge outside sdlc” sweep, without Landing-PR/Landing-Head trailers. The conductor completes all reachable cards, archives its own PR, then finishLandedLeftovers archives sibling done cards. The waiting merge discards its landed outcome, and confirmLandingArchive requires its own PR trailers; missing active details then fail planLandingArchive. Existing outcome confirms landed only, not archive completion, and precedes cleanup. Preserve this distinction (ARCH-ORDER/SECURE); no production edits yet. A tracked two-real-car fixture is required: legacy train fixtures do not run the done-card sweep.

- The deterministic two-real-merge regression now reproduces the incident: after the conductor completes both cards and its generic sweep archives #2, the blocked waiter returns “active issue missing without confirmed archive: workshop/issues/000002-second.md”. Fixture setup first exposed invalid card/detail separation, a synthetic arbitrary clone unsuitable for a real durable merge, and a pre-cutover resting baseline; corrected all before observing the behavioral red run. With only the production proof correction, that exact regression passes in 19.570s (test body 19.14s), including the train-archive output, unchanged main-slot2 and deleted issue ref.
- Implementation retains the exact settled outcome and enables fallback only for a failed no-per-PR-provenance case with a named train PR. It discovers one bounded path-history add commit per owned tracked issue, proves integration precedes it, and reuses the complete existing archive-transition/current-generation verifier. Existing or ambiguous per-PR provenance never falls back. No persisted format or archive writer changes; the typed wrapper preserves the ordinary error text and chain. ARCH-DRY/ORDER/SECURE shaped these boundaries. This is a semantic confirmation bug defended by stateful integration and proof controls, not a lexical mistake class suitable for a Go lint rule.

- Direct proof controls pass in 13.835s, covering positive generic archive, unchanged serial rejection, no archive, missing/tampered/malformed destinations, active issue/plan residue, ambiguous add history, malformed introducing transition, different completion token and untracked refusal; each negative pins its expected fault rather than accepting any error. Existing archive controls, all original train mergeCommit variants, no-outcome recovery, tracked archive idempotency and merge/squash/rebase controls also passed in the focused run; its only failure was the new control fixture using a file transport identity as PR.Repo, corrected to test/repo before the clean control pass. The two-real-merge fixture now reuses the existing synchronized watchWriter for concurrent stderr.

## Revisions

- 2026-10-20: refine the original outcome-only wording after investigation showed that landed outcomes precede archival. Require the existing immutable archive-transition proof in addition to the retained outcome, preserving the requested successful train-archive report without treating landing as archive success. Proof design awaiting hub approval before implementation.

- Hub approved the concrete proof design (receipt 6bd5e522-3ca0-4274-8575-a454b85901dc, verified submitted). change-code admitted the quick flow. The existing archive verifier will be reused; a read-only second pass confirmed its generation/asset/side checks and the tracked two-car fixture seams. Regression implementation is isolated to a new test file; production remains unchanged until the observed red run.
