---
id: 000421
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '7392b00984bd7ade5f93894aff667d0a08250145' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T22:20:44-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ea3c6446", done: "85448d29"}
---

# sdlc stage: one stage at a time in :0, a lease naming the holder; staging over another issue's live stage refuses

## Problem

## Spec

2026-10-16: ariadne:0 (the checkout whose `bin/` every slot's sdlc runs) was staged with #418 for pair:2's live merge of pair#618. ariadne:5 then ran `sdlc stage --build` for #416 on top of it. `stage` replaced HEAD and rebuilt the bin tools without noticing that another issue's stage was live and in use. `bin/sdlc` only survived by accident, because #417's sentinel skip left it alone. The TL had been granting :0 to one slot at a time by message. A rule kept by messages fails.

What makes "one stage at a time" a single check is one resource held for the stage's lifetime: a stage lease in :0 naming the issue, branch, slot and time. Then:
- `sdlc stage` takes the lease, and refuses while another issue holds it, naming the holder and when it staged.
- Clearing the stage releases the lease. Re-staging the same issue refreshes it.
- A recovery verb, shown in the refusal, frees an abandoned lease with an explicit acknowledgement.

## Done when

- With :0 staged for issue A, staging issue B refuses, naming A, its slot and its time; a test covers it.
- Clearing A then staging B succeeds; re-staging A while A holds the lease succeeds.
- The refusal names the recovery verb for an abandoned lease.

## Plan

**The lease is the staged record, extended.** `:N`'s per-worktree `sdlc-staged` file (#344) already lives for exactly
the stage's lifetime: `stage` writes it and `--clear` or a move removes it. It gains the holder, so one file is both
the record and the lease (ARCH-DRY: no second file to keep in step). Format, one `key value` per line: `commit`,
`holder` (the issue's CLI ref when the branch is an issue branch, else `branch:<name>`), `branch`, `slot`, `at`
(RFC 3339). A legacy one-line record is a commit with no holder, so it is unheld. Any other unparsable content refuses
and names the file and the recovery verb (ARCH-SECURE: never read a damaged lease as free).

**Acquiring is one step under a flock.** `stage`, `stage --clear` and `move` into `:N` take a non-blocking flock on
`sdlc-staged.lock` (`internal/flock`) in `:N`'s git dir. They hold it across read lease → decide → write lease →
switch → record the commit, and release it before `--build`. A second concurrent stage or move refuses ("another
stage or move is changing :N now"). Nothing is decided from observing `:N`'s HEAD.

**The rules, pure, in stageplan.go** (`checkLease(held, me)`, unit-tested):
- No holder, or the same holder: allowed. The same holder refreshes the lease (slot, branch, `at`).
- Another holder: `stage` and `move` refuse, naming the holder, branch, slot and time. The refusal shows both exits:
  the holder's `sdlc stage --clear :N`, and for an abandoned stage `sdlc stage --clear :N --abandoned <holder>`.
- `--clear` from a slot that isn't the holder refuses the same way. Clearing someone else's live stage is as
  destructive as staging over it.
- `--clear --abandoned H` clears only when `H` is the current holder. Naming the holder is the explicit
  acknowledgement, and it can't free a lease that has since changed hands.
- `--clear` with `:N` already resting but a lease still held (`:N` switched by hand) releases the lease.

**Write order (ARCH-ORDER):**
1. The holder is written before the switch, keeping the old `commit`, and `commit` is updated after the verified
   switch.
2. A crash after the switch still leaves `:N` held.
3. A crash before the switch leaves a held lease on an unchanged `:N`, which the holder's rerun or `--clear` resolves.

**Lifetime (ARCH-FUNERAL):** one file per slot, removed by `--clear` (holder or abandoned) or by a move into the slot,
as today.

- [x] fleet/staged.go: `StageLease` with parse and format; `ReadStageLease`, `WriteStageLease` (temp file then
  rename) and `LockStage`. `StagedCommit` reads the lease's commit. Unit tests: round trip, legacy, damaged.
- [x] stageplan.go: `checkLease` and `checkClearLease`, plus the holder of a branch. Table tests.
- [x] stage.go and move.go: the flock, lease reads and writes in that order; `--abandoned`.
- [x] Integration tests for each Done-when line, run on `moveFixture`'s three slots:
  - A staged, B refused (naming A, its slot and its time), and `:0` unchanged;
  - B's `--clear` refused;
  - `--abandoned` with the wrong holder refused, and with the right one clearing;
  - A clears, then B stages;
  - A restages;
  - a move from B into A's stage refused;
  - a damaged lease refused;
  - a busy flock refused.
- [x] Recovery catalog entry, atlas (sdlc-binary staging), and `sdlc stage --help` text.

## Log

### 2026-10-16
- Coordination with #417 (ariadne:1, which builds and verifies every binary in `stagebuild.go`): this issue touches
  `stage.go` (`runStage`, `runUnstage`), `stageplan.go`, `fleet/staged.go` and `move.go`, not `stagebuild.go`. The one
  shared line is the `stageBuild` call at the end of `runStage`.
- Built: the lease in `fleet/staged.go`, the rules in `stageplan.go` (`checkLease`, `checkClearLease`, `stageHolder`), the
  flock and lease writes in `stage.go`, and the same check in `move.go` through `takeStageLease`. Targeted runs pass:
  `TestStageLease*`, `TestCheckLease`, `TestStageLeaseParse`, and every existing Stage and Move test.
- The refusals print the target's full address (`sample:0`, `ariadne:0`), not the `:0` the operator typed, so the
  recovery command can be copied from any slot.

### 2026-10-17
- Suite 1 on d260ae08 failed one test: `TestEveryCommitSiteStamps` read `leaseKeys`' "commit" field name as a commit
  site. Exempted it as a field name (48a2bc9f).
- Merged main with #417 (42c2184e). `stage.go` and `helptext/stage.md` conflicted; kept #417's rebuild on `--clear`,
  and a resting target with a held lease now frees the lease, then rebuilds. Suite green on 42c2184e.
