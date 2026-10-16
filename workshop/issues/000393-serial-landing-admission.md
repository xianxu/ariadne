---
id: 000393
status: codecomplete
deps: [392, 380]
github_issue:
created: 2026-10-15
updated: 2026-10-16
estimate_hours: 3.86
card_mirror: 'd1cd02597cca4b556f13b3390986948d5b8913cc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T23:44:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.95
---

# Serial per-repo landing admission with fleet-blocking priority

## Problem

Concurrent `sdlc merge` calls integrate against the same main; after one lands, the others reintegrate and rerun a roughly nine-minute suite plus queue wait. #376 required three integrations. A fleet-blocking fix (#391) also waited fourth for a suite with no priority lane. #390's operator-approved design chooses serial, per-repo landing admission before any batching.

## Spec

Keep `sdlc merge` as the agent verb. Admit one ready PR per repository at a time; the admitted candidate integrates current main, runs one full suite on its exact head, verifies review/CI/local gates and pinned PR/main OIDs, then lands. Later PRs wait; a stale head/main re-enters admission without falsely reusing suite evidence. Use an explicit crash-recoverable state model and uncertain-publication reconciliation. An operator or TL may grant a logged, expiring `fleet-blocking` priority for a named issue; it moves ahead of waiting normal candidates at the next admission, FIFO within each class, but never preempts a running suite or waives a gate. Preserve #370's short transaction-lock scope. Queue files live outside Git in Couch's storage directory when Couch is configured and in the repository Git common directory otherwise; `sdlc` reads/writes them directly with atomic writes and locking so Couch may be down. Couch only displays state. Cross-repo task dependencies remain the TL's pre-start sequencing responsibility. Follow #390's approved plan and side doc 01; rollout is an ariadne default for every SDLC repo after the ariadne/pair pilot.

## Done when

- [ ] Concurrent ready merges in one repo admit only one active integration; different repos are independent and the full-suite gate remains exact-head.
- [ ] Main/head drift, failed suite, process death in each phase, duplicate invocation and uncertain GitHub outcome recover or refuse without duplicate landing or lost evidence.
- [ ] Granted fleet-blocking work overtakes waiting normal work; same-class FIFO holds; running suites continue; expired, unauthorized or unlogged grants confer no priority. Grant and displacement are auditable.
- [ ] Landing works and recovers while Couch is down, with atomic queue files in its storage directory; a plain-shell setup uses the Git common directory.
- [ ] Stateful Git/GitHub fake and live read-only conformance cover the consumed interactions; pilot reports full suites per landed issue and p90 ready-to-land delay before/after.

## Plan

Durable plan: [000393-serial-landing-admission-plan.md](../plans/000393-serial-landing-admission-plan.md) (TL-approved design, 2026-10-16).

- [x] M1 — landingline package: pure NextHead, grant validity and Authorize; flock-held tickets and head admission; events and grants logs with a 30-day sweep.
- [x] M2 — runDurableMerge holds the line for its lifetime; sdlc landing status, grant and revoke; integration tests on landingFakeGH; atlas and merge help.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.4 impl=0.32
item: greenfield-go-module design=0.4 impl=0.32
item: smaller-go-module design=0.1 impl=0.2
item: smaller-go-module design=0.2 impl=0.2
item: smaller-go-module design=0.1 impl=0.2
item: smaller-go-module design=0.05 impl=0.05
item: atlas-docs design=0.1 impl=0.08
item: milestone-review design=0.1 impl=0.2
item: milestone-review design=0.1 impl=0.2
item: issue-spec design=0.2 impl=0.08
design-buffer: 0.15
total: 3.86
```

Derived against `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`, Method A only, and provisional because the source reports stale. The items follow #376's pattern. Two greenfield Go concerns: the pure ordering and grant model, and the flock-held line with helper-process death tests. Four smaller modules, each extending existing code: the events and grants log with its sweep, merge wiring with the PQ-1 integration cases on `landingFakeGH`, the `sdlc landing` verbs, and the reviewbox flock lift. Then atlas, two milestone reviews, and the issue spec and plan already written.

## Log

### 2026-10-15

Filed from #390's approved design. Depends on the weekly baseline (#392) and saved local-check status recovery (#380). Batching and doc-only reuse are separate later choices.
- The TL approved the lock-shaped plan (72fd2309) on 2026-10-16. The proposal text that went to the TL:
  Proposed to the TL on 2026-10-16; waiting for approval. This is the lock-shaped design #404 asks to weigh first.

  **Seams found** (cmd/sdlc):
  - The durable landing path, `runDurableMerge` (landing.go:336), already models integration with a transition function (integrationstate.go): integrate main with an `Integrated-Main` trailer, push with a lease, wait for the gates, then `gh pr merge --match-head-commit`. Its recovery uses only Git and GitHub state: the trailer acts as the receipt, an already-merged PR resumes, and a lost merge response re-reads the PR (convergent retry, #280).
  - Main moving loops back to integration (landing.go:397, 456, 471, 511), capped at 3 attempts (`integrationLimit`).
  - The repo transaction lock (#370) lives in the git common dir. Every slot of a repo shares that directory (the slots of ariadne all resolve to ~/workspace/ariadne/.git).
  - scripts/test-lock.py already keeps a FIFO line of flock-held tickets for suites, pruning dead waiters' tickets.

  **Design: a per-repo landing line, held for the merge's lifetime.**
  - `runDurableMerge` takes a ticket in the repo's landing line before its integration loop and holds the line until it returns: landed, refused, or died. Each ticket is a file flock-held by the merge process itself, so process death releases it and no daemon or helper process is needed. A PR that is already merged takes no ticket.
  - Only the head of the line integrates, so its suite runs once on a main nobody else is testing against. Waiters integrate only after the head has landed. Main moving under the head (a direct push) uses the existing loop and cap.
  - Crash recovery, a duplicate invocation, an uncertain GitHub outcome and drift stay what they are today, because the line adds ordering only. A second `sdlc merge` of the same PR waits in the line behind the first. #390's durable eight-state queue is replaced by two states, a live waiting ticket and the head. The landing's own state machine stays in integrationstate.go.
  - **Order:** a pure function, `nextHead(tickets, grants, now)`. Live fleet-blocking tickets go first, then normal ones, FIFO within each class. The head is never preempted: priority decides only who is admitted next, at each poll (every 2 s, under the line's allocation mutex).
  - **Priority lane:** `sdlc landing grant --repo DIR --issue N --reason R [--hours H ≤ 24]` appends a grant to the line's `grants.jsonl`. It refuses from a worker slot (a numbered slot), so only the operator's shell or a :0 slot (the TL's hub) can grant. The log is the only source of grants: a valid grant is the latest logged entry for the issue that has not expired, is not revoked, and records a grantor. Landing ends a grant.
  - **Audit and pilot data:** `events.jsonl` records enqueue, admit (with the tickets it displaced), and landed or refused. `sdlc landing status` shows the line and the live grants. p90 ready-to-land delay is enqueue to landed; full suites per landed issue come from test-stats (#376).
  - **Storage:** `<git-common-dir>/sdlc/landing/`, beside #370's lock (the location is a TL/operator choice; see below).
  - **Lifetime (ARCH-FUNERAL):** a ticket is removed when its process exits, and a dead waiter's ticket is pruned by the next scan. Events and grants older than 30 days are swept on write.
  - **Out of scope:** the legacy path for ordinary worktrees, which never integrates and so pays no integration tax. Batching, a later revision of this issue. Doc-only reuse (#398).

  **Tests**, against the existing `landingFakeGH` with real temp git repos (ARCH-MOCK; the live read-only conformance check `TestLandingGHLiveConformance` already exists):
  - two concurrent merges in one repo: the second waits, integrates after the first lands, and each runs exactly one suite;
  - different repos don't wait on each other;
  - killing the head lets the next one through;
  - a dead waiter is pruned;
  - a duplicate invocation waits rather than landing twice;
  - direct-push drift re-integrates once;
  - `nextHead` table cases: fleet-blocking overtakes normal, same-class FIFO, no preemption, and an expired, unauthorized or unlogged grant confers nothing;
  - grant authorization;
  - the event and grant sweeps.

  **Decisions for the TL:**
  1. Replace #390's durable queue state machine with the lock-shaped line, which is a delta from the operator-approved design.
  2. Storage location. I recommend the git common dir rather than `COUCH_STORE_DIR`: that variable is set per shell, so an operator shell without it and a slot with it would use two different lines for one repo. Couch can still display the line by reading the git common dir.
  3. Build now while #380 and #392 are open. #380 isn't needed: a failed status post at worst reruns a suite inside the head. The pilot's before/after comparison waits for #392's baseline.

### 2026-10-16
- 2026-10-16: closed — Close-review Minor fixed: readTicket reads the whole ticket bounded at 1 MiB instead of 4096 bytes; TestLineLongTicketReads (8 KiB branch name) reads it back intact. landingline go test -race green. (fixed-then-ship); review verdict: SHIP
- 2026-10-16: closed — Per-repo landing line (lock-shaped, TL-approved over #390s eight-state queue). Done-when 1: one admitted merge per repo via flock-held tickets and head.lock in the shared git common dir; TestLandingLineWaitsThenIntegratesOnce shows a waiting merge does nothing while another holds the line, then integrates the moved main once with exactly one suite and lands; separate repos are separate dirs (TestLineIndependentDirectories); exact-head CI gate unchanged. Done-when 2: SIGKILLed helper processes as waiter and as head release the line; duplicate invocation resumes as merged with no second merge; fix pushed while waiting is what lands; refusal frees the line; lagging listing refuses having changed nothing; drift and uncertain-outcome recovery unchanged and all existing TestLanding/TestMerge pass. Done-when 3: NextHead/Order/ValidGrants tables (fleet-blocking first, FIFO within class, never preempts, expired/unauthorized/unlogged/future-dated confer nothing), grants checked on read, randomized invariant that a sweep never changes validity, admit events record displacement, grant refused from worker slots. Done-when 4: no Couch dependency; git common dir is the only location (TL decision). Done-when 5: landingFakeGH on real temp repos plus existing live TestLandingGHLiveConformance; events.jsonl gives ready-to-land delay p50/p90 in sdlc landing status, test-stats gives suites per landed issue; the 7-day pilot report is the TLs (Revisions). Full make test green (12 shards, exit 0). M1 and M2 closed SHIP after review rounds with every Minor fixed.; review verdict: SHIP
- 2026-10-16: closed M2 — M2 review round 2 answered. I-A (gate-skip-predicate-divergence, rule: a merge without a turn acts only on an already-merged PR): resumeOnly skips the line only for a merged PR with no open one, an empty listing queues (TestLandingLineEmptyListingQueues), and the selected PR is held to the rule; TestLandingLineLaggingListingRefusesOutsideTheLine (first listing shows only an old merged PR, the next shows the open one) refuses with nothing changed, and disabling the check makes it fail. I-B: README and the atlas subcommand table cover sdlc landing. Minors: status grant order tested across runs, lineLog warns once on a log write failure and leave warns when grants are unreadable (TestLandingLineLogFailuresWarn), logger owned by the line (no global), plan Task 4 paragraph and sketch match the code. Targeted TestLandingLine*/Grant/Revoke/Status/EndsThe green.; review verdict: SHIP
- 2026-10-16: closed M1 — M1 review round 2 answered: retention is pure (retention.go KeepGrants/KeepEvents), sweep only applies it; TestKeepGrantsPreservesValidity checks ValidGrants(kept)==ValidGrants(all) and nothing outside retention kept over 2000 random logs with out-of-order, old and future entries; TestSweepKeepsGrantsInForce is the round-2 regression (a sweep while grants are in force keeps all three); ticket counter is untrusted (malformed refuses, set back cannot overwrite a live ticket). Plan Task 1 sketch and Task 2 Status signature marked as built; Revisions and atlas updated. go test -race -count=3 green.; review verdict: FIX-THEN-SHIP

- M1 build notes (deviations from the durable plan, each a simplification):
  - `NextHead(tickets, grants)` has no `now`, because `ValidGrants` already applied the clock.
  - A pure `Order` gives the full admission order, so the position line and `NextHead` derive from one function.
  - `DelayStats` became `Delays`, which returns the raw seconds; `sdlc landing status` passes them to test-stats' `suiteQuantiles`, so there is still one nearest-rank implementation (ARCH-DRY).
  - `Ticket` carries no PR, because the PR is chosen after admission (PQ-1); the `admit` event records it.
  - The admitted head deletes its ticket and records itself in `head.lock`, so it can never be passed to `NextHead`.
  - `OnAdmit` runs outside the line mutex: flock is per open file, so a hook that logged under the mutex would deadlock against its own process.
- Two test pitfalls hit:
  - a helper process with `select {}` dies as a Go deadlock, so it now sleeps until killed;
  - the test grants map needed a mutex for the race detector.
- landingline: `go test -race -count=3` green.


## Revisions

- 2026-10-16, TL-approved design deltas from #390:
  - (1) A per-repo landing line replaces #390's durable eight-state queue. A ticket is flock-held by the merge process and admission is the head lock. The landing's recovery stays in integrationstate.go and in Git and GitHub state (#404 lock-first). Done when's "process death in each phase" therefore means a waiter's death and a head's death. The landing phases recover as they do today.
  - (2) Storage is <git-common-dir>/sdlc/landing/ in every setup, not COUCH_STORE_DIR, which is set per shell and would split one repository's line. Couch reads the line there for display. Done when bullet 4's "atomic queue files in its storage directory" becomes "in the git common dir, which every slot shares".
  - (3) #380 and #392 aren't prerequisites; build now.
  - (4) Pilot: this issue delivers the pilot's data (events.jsonl for the p90 ready-to-land delay, test-stats for suites per landed issue). The seven-day before/after report is the TL's, recorded in test-economy.
  - (5) The line is on for every repository's durable merge as soon as it ships. It only orders landings within one repository, so the "ariadne/pair pilot, then default" rollout reduces to measuring ariadne and pair.
