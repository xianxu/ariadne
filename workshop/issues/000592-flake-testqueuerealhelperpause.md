---
id: 000592
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '257d0bff731e42b79f1f00ddb9945ead8388dc57' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T04:45:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause]
flow: {kind: quick, provenance: inferred, spec: "096ee224", done: "a0bbe98b"}
---

# flake: TestQueueRealHelperPause

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause: failed in attempt c5fff316216ec03a781b4796f770a3e5, passed on the same head in attempt f3122a02c4fdbbaba2e5dd1f03b2b89f

## Spec

It isn't a timeout. Both recorded failures (attempt 81a77a06 on b9611a8a, and 7e509591 on the
#588 merge head 7108f40d) fail at `test-lock-pause.test.py:255`, `assertFalse(path.exists())`:
the operation file is still there.

`QueueLifetime.retire` (`scripts/test-lock.py:574`) runs one transaction under `control.lock`:
write `unresolved/<id>.json`, write `interrupted/<id>.json`, unlink `operations/<id>.json`. The
test's wait (`:246`) polls for the two evidence files without that lock, so it can observe the
transaction half done; load widens the window between the second write and the unlink.
Production readers don't have the race: queuegate `Observe` reads under `WithGate`, the same
lock.

Reproduced deterministically: a `sitecustomize` shim on `PYTHONPATH` that sleeps 0.5s before
the helper unlinks an `operations/*.json` fails line 255 in 6 of 6 runs.

The same class, found by a sweep of the queue-helper tests' unlocked reads of state the helper
writes in one locked transaction:
- `test-lock-exit.test.py:129` (`unknown()`) waits for `unresolved/` only, then `:273` reads
  `interrupted/`, which the same transaction writes second.
- `test-lock-pause.test.py:229-232`: the child owner writes its pid with a non-atomic
  `write_text`, and the test reads it as soon as the file exists. A slowed run read it empty
  (`ValueError: invalid literal for int()`), 1 in 10.

Other waits read files written atomically or ordered by a barrier (`make.pid` before `ready`;
`priority:436` guards with `isdigit`).

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ] Fixture waits that observe queue state take the helper's `control.lock` without blocking
      (`lock.try_lock`): a predicate is evaluated only while the lock is held, so it sees
      committed transactions only. Used at `pause:246`, where it waits for the settled state (both
      evidence files present, operation file gone), and in `exit` `unknown()`, where the
      unresolved and interrupted files then appear together.
- [ ] The child-owner fixture writes its pid file atomically (temporary file + replace).
- [ ] Regression: the pause test case also runs with the slowed-unlink shim on the helper's
      `PYTHONPATH`; it fails before the fix and passes after.
- [ ] Verify: the shim case repeated 20 times, both scripts plain, and `TestQueueRealHelperPause`
      and `TestQueueRealHelperCompletion`.

## Log

### 2026-10-24
- Root cause: an unlocked fixture read of the helper's locked retire transaction, not an overrun; reproduced with a slowed-unlink shim (see Spec).
