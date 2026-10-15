---
id: 000383
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '70f5509b1909f08bb06abcb85ec5ee73077f57b2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T22:12:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc merge holds the machine suite slot until the process exits, through GitHub polling, the PR merge and archiving

## Problem

On 2026-10-14 (21:56–22:08) ariadne#377's `sdlc merge --yes` held the machine-wide suite slot for ~12 minutes with no
test process running, while parley.nvim#330's suite waited 2nd in the queue (reported by parley.nvim:2). Cause:
`acquireSuiteSlot` (`cmd/sdlc/localcheck_run.go:99`, `:202–212`) runs `scripts/test-lock.py acquire --watch
<os.Getpid()>`, so the slot is tied to the `sdlc merge` process and released only when it exits; `test-lock.py` has
no release. After the suite finishes, the slot stays held through:
- GitHub status/comment writes;
- the CI re-read and the 30 s pending-poll loop (`cigate_wait.go:190–198`);
- possible re-integration pushes;
- the PR merge (`landing.go:508–529` / `merge.go:641`), archive commits and pushes, and branch cleanup.

Every landing therefore blocks the fleet's suites for minutes beyond its own test run.

## Spec

- The suite slot covers **test execution only:** acquire right before the suite process, release right after it exits
  (e.g. a short-lived child process as the `--watch` pid, or an explicit release in `test-lock.py`).
- A reentrant rerun in the same merge (re-integration, baseline diagnosis) acquires again, in ticket order like anyone
  else.
- No change to admission order, slot counts or the suite command.

## Done when

- A test with the real `test-lock.py` shows the slot is free while `sdlc merge` polls GitHub and merges (a fake GitHub
  client that blocks), and a second waiter is admitted then.
- A rerun inside the same merge re-acquires correctly.
- The local-check docs say what the slot covers.

## Plan

**Design.** Each run gets its own slot holder: a child process that lives only as long as the suite. `acquireSuiteSlot` starts `cat` with its stdin on a pipe from `sdlc` and its stdout on /dev/null, in its own process group, and passes **that child's pid** as `--watch`. `test-lock.py`'s watcher releases the flock when the watched pid exits. So the release is: close the pipe's write end, then `Wait` on the child. `runLocalCheck` calls it right after the suite exits and its process group is terminated. That comes before the log hash, the status post, the PR comment and everything `sdlc merge` does after the run: the CI poll, re-integration, the PR merge, archiving and cleanup. A `defer` also releases on every early return.
- **No helper change.** Every copy of `test-lock.py` already takes `--watch`, including other repos' substrates found through `Makefile.workflow`. So there's no capability probe and no mixed-fleet case, and admission order, slot counts and the suite command stay as they are.
- **Crash safety is kept.** If `sdlc` dies by any route, the kernel closes the pipe, `cat` reads EOF and exits, and the watcher releases the slot. The pipe's write end is close-on-exec, so the suite never inherits it and can't keep the holder alive.
- **A rerun re-acquires.** Each `runLocalCheck` (a re-integrated head, the baseline diagnosis, another named check) starts a fresh holder with a fresh pid. The helper sees no slot held by that pid, so the run takes a ticket and waits in FIFO order like anyone else. `suiteLockOwners`, the per-process cache that made reruns pass through as reentrant, goes away. `reentrant` now means only that this `sdlc` is nested under a locked make: the run exports the inherited `WF_TEST_LOCK_HELD_BY`, and the holder is released at once.
- `WF_TEST_LOCK_HELD_BY` exported to the suite is the holder's pid, which is the pid in the slot record, so a nested `make test` still passes through as reentrant.

**ARCH-ORDER.** The holder has three states: started, holding, released. Events and outcomes:
- The helper answers `acquired`: holding.
- The helper answers `reentrant` or `skipped`, or times out or fails: release the holder at once and return as today.
- The suite exits normally or by timeout: release.
- The context is cancelled: release by `defer`.
- `sdlc` crashes: the kernel releases.
- Release twice: a no-op (`sync.Once`).

The event most likely to be mishandled is an early return between acquire and run; the `defer` covers it. **ARCH-FUNERAL:** the holder is the slot's lifetime, and nothing new is persisted. **ARCH-MOCK:** the tests use the real helper and real flocks, as the reentry tests already do.

- [ ] `acquireSuiteSlot` returns a holder with a `release`; `runLocalCheck` releases right after the suite and `defer`s it; `suiteLockOwners` removed
- [ ] Test (real `test-lock.py`): a fake GitHub whose `PostStatus` blocks runs an outside `acquire --slots 2` while blocked, and that outside acquire answers `acquired`; and from inside the suite, an outside acquire answers `timeout` (the slot is held for the suite's whole run, per the TL). That is, the slot is free during the GitHub phase of `runLocalCheck`, which every later merge step comes after.
- [ ] Test: a rerun goes through the line. Between run 1 and run 2 an outside process holds the slots, so run 2 times out with `errUnboundedSuite`; without the outside holder, run 2 acquires and its nested make sees `reentrant` with run 2's holder pid. The existing reentry tests keep passing.
- [ ] Docs: atlas ci-merge-check.md (Recording) and sdlc-binary.md (the suite lock) say the slot covers the suite's execution only

## Log

### 2026-10-14
- Plan: a per-run slot holder (`cat` on a pipe) as the `--watch` pid, released right after the suite; no `test-lock.py` change (every copy already takes `--watch`). Quick flow.
- The TL approved the plan and added: test (1) also asserts that no waiter is admitted while the suite runs.
