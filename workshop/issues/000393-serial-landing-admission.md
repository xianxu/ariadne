---
id: 000393
status: working
deps: [392, 380]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'abceb82d523b1fc0851167a186147f3d1556e4d0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T23:44:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

## Log

### 2026-10-15

Filed from #390's approved design. Depends on the weekly baseline (#392) and saved local-check status recovery (#380). Batching and doc-only reuse are separate later choices.
