# Boundary Review — ariadne#346 (whole-issue close)

| field | value |
|-------|-------|
| issue | 346 — test gate: a sharded full suite counts as both suite slots (two ariadne suites at once pushed load to 185) |
| repo | ariadne |
| issue file | workshop/issues/000346-test-gate-a-sharded-full-suite-counts-as-both-suite-slots-two-ariadne-suites-at-once-pushed-load-to-185.md |
| boundary | whole-issue close |
| milestone | — |
| window | 810c69ea76578e679e34487791d6f12ca08eddb3..5b2e260be969a79d7dbb70263b4c8719ca9633c0 |
| command | sdlc close --issue 346 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:55:37-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The tests pass. Now I'm checking the other places that describe the lock (pkg/testgate, AGENTS.base.md) for statements the weight change has made stale.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. `scripts/test-lock.py` now takes all N slots in one scan or none. A queue lock means only one waiter, the head, can take slots. The head holds no slot and slot holders wait on nothing, so there is no deadlock cycle. The Makefile wiring passes the per-repo `WF_TEST_WEIGHT`, which defaults to 1, so other repos that include `Makefile.workflow` keep today's behaviour. I ran `bash scripts/test-lock.test.sh` in the pinned checkout and all cases passed, the new #346 cases included. Every Done-when clause is tested both ways: two heavy suites run one after the other, a light suite waits for a heavy one and the reverse, a waiter holds no slot, and a second heavy waiter finishes. Nothing blocks shipping. What's left is doc drift: one comment contradicts the implementor's own logged lesson, one agent-facing summary is now incomplete, and the "cannot starve" claim is stronger than the code guarantees.

**1. Strengths**
- `scan(fds, want)` (`scripts/test-lock.py:61-76`) is all-or-none: it locks every free slot it can, keeps the first `want`, and releases the rest in the same pass. A waiter never keeps a partial set.
- The no-deadlock argument holds in the code, not just the docs. The queue is closed before the fork (`test-lock.py` `os.close(queue)`), so the head moves on before it starts holding slots. The timeout and reentrant paths release the queue when the process exits.
- Moving `rec` above the loop lets the queue file and the slot files share one record shape. That makes `holders()` dedup and `next in line` work without a second format.
- The test for "the waiter holds no slot" (`free_slot "$tmp/w.1"`) checks lock state directly instead of parsing log text. The mutation check in the Log (letting a non-head take slots turns two cases red) shows the queue test has teeth.
- `WF_TEST_WEIGHT ?= 2` with `$(or $(WF_TEST_WEIGHT),1)` leaves the default unchanged for other repos, and an operator's environment override wins (tested by "a weight-1 make test runs alongside").

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- **Wrong declaration form in a comment.** `Makefile.workflow:15-17` tells repos to write `WF_TEST_WEIGHT := N`. The Makefile and the atlas use `?=`, and the issue Log says `:=` was rejected because it beats the operator's environment override. Every place in this window that states the declaration form:
  - `Makefile.workflow:16` uses `:=` (wrong);
  - `Makefile:10` uses `?=` (correct);
  - the atlas uses `?=` (correct);
  - the `test-lock.py` header names `WF_TEST_WEIGHT` without giving a form (fine).
- **Agent-facing summary not updated.** `AGENTS.base.md:62` still says "At most 2 full `make test` suites run machine-wide; a third waits". With weights, an ariadne suite waits behind any single suite, and it also blocks any second suite. This file propagates to dependent repos, so add one clause about weighted suites.
- **"Cannot starve" is stronger than the code.** The queue is a polled `flock`, not first-in-first-out. Only the current head is protected. A heavy waiter that is not yet head races light newcomers for the queue each time it frees up. The claim appears in two places, and both should say "the head waiter is never overtaken":
  - `atlas/workflow/sdlc-binary.md` ("so light suites cannot starve a heavy waiter");
  - the `test-lock.py` header ("light suites cannot starve a heavy waiter").
- **The make-wiring test can fail on an operator's environment.** `scripts/test-lock.test.sh` unsets `WF_TEST_LOCK*`, but not `WF_TEST_WEIGHT`. With `WF_TEST_WEIGHT=1` exported, "make test waits for both slots" fails. Add it to the existing `unset` line.
- **Old helpers during rollout.** A worktree still running the pre-#346 `test-lock.py` ignores the queue and takes any free slot. This lasts only until every worktree has the new helper, so it needs no code change.

**5. Test coverage notes**
- Every plan case is present and passes: weight-2 takes both slots, light waits behind heavy and heavy behind light, no partial hold, a light newcomer queues behind a heavy waiter, two heavy waiters both finish, the clamp, and the make wiring both ways.
- Not covered: a nested (reentrant) `make test` under a weight-2 holder while another waiter holds the queue. I read the code path: a non-head still scans, sees the busy records, and returns `reentrant`. It looks right but has no test.
- The tests use only `sleep` holders and `make -n`, so they generate no real load, as Done-when requires.

**6. Architecture notes**
- **ARCH-DRY: pass.** The `holders()` dedup is new. The record write moved rather than being copied, and `try_lock` is a small helper next to the inline try in `scan`. Folding the two together would be optional tidying.
- **ARCH-PURE: pass.** The script is a thin IO wrapper by nature (`flock`, `fork`). The new rules (all-or-none selection, the clamp, the dedup) are small and tested through fake lock files, which suits a script of this size.
- **ARCH-PURPOSE: pass, with the shadow sweep noted above.** The weight's consumers are `Makefile.workflow`, the root `Makefile`, `test-lock.py` and the atlas, and all of them derive from it. One hand-written restatement of the budget is out of date: `AGENTS.base.md:62`, listed under Minor. `pkg/testgate` only reads the lock files and is unaffected.

**7. Plan revision recommendations:** none needed; the plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-drifts-from-code
    title: |
      Makefile.workflow comment says WF_TEST_WEIGHT := N; the contract is ?= (env override must win)
    detail: |
      Makefile.workflow:16 says `:=`, while Makefile:10 and the atlas use `?=`, and the issue Log records that `:=` was rejected because it beats an operator's environment override. Change the comment to `?=`. Those are all the places in the window that state the declaration form; only Makefile.workflow:16 is wrong.
  - id: new
    severity: Minor
    family: doc-claim-drifts-from-code
    title: |
      AGENTS.base.md:62 machine test budget still says "at most 2 suites; a third waits", which ignores weights
    detail: |
      With WF_TEST_WEIGHT, ariadne's sharded suite takes both slots: it waits behind any single suite and blocks any second one. This agent-facing summary propagates to dependent repos; add one clause about weighted suites (ARCH-PURPOSE shadow sweep).
  - id: new
    severity: Minor
    family: doc-claim-drifts-from-code
    title: |
      "Light suites cannot starve a heavy waiter" overstates it; only the queue head is protected
    detail: |
      The queue is a polled flock, not first-in-first-out, so a heavy waiter that is not yet head races light newcomers for the queue each time it frees. The claim appears in two places: the atlas "One queue head" bullet and the test-lock.py header. Reword both as "the head waiter is never overtaken by light suites".
  - id: new
    severity: Minor
    family: test-env-hermeticity
    title: |
      test-lock.test.sh does not unset WF_TEST_WEIGHT, so the make-wiring case fails under an operator's env override
    detail: |
      The script already unsets WF_TEST_LOCK, WF_TEST_LOCK_TIMEOUT and WF_TEST_LOCK_HELD_BY; add WF_TEST_WEIGHT to that line so "make test waits for both slots" does not depend on the caller's environment.
```
