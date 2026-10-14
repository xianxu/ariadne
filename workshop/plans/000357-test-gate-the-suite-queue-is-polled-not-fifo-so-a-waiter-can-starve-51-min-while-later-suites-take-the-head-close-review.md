# Boundary Review — ariadne#357 (whole-issue close)

| field | value |
|-------|-------|
| issue | 357 — test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head |
| repo | ariadne |
| issue file | workshop/issues/000357-test-gate-the-suite-queue-is-polled-not-fifo-so-a-waiter-can-starve-51-min-while-later-suites-take-the-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | b913e05ad2aaa9e97738f5a6dc064bde1ce30dc7..83634985882f1566e69c81d746218cacb9ce1d50 |
| command | sdlc close --issue 357 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T20:19:22-07:00 |
| verdict | SHIP |

## Review

Checking how the merge check discovers the moved test, then running the test suite.

```verdict
verdict: SHIP
confidence: high
```

This boundary delivers what the Spec and the Done-when for ariadne#357 ask for. I ran the stat, name-status and full-diff recipes on the pinned range and found nothing blocking. Admission is now first-in-first-out: each waiter takes a numbered ticket, holds a flock on it while it waits, and only the lowest live ticket can become the head. Weights, the take-all-slots-or-none rule and the no-deadlock property (the head holds no slot) are unchanged. The tests pass at head: `bash scripts/test/test-lock.test.sh` printed "all passed" in about 32s. To check they actually catch the bug, I swapped in the base commit's `scripts/test-lock.py` and re-ran them. Three of the new FIFO cases failed: positions, "a later waiter that polls first does not overtake", and "an older copy's head goes first". I then restored the file. The remaining findings are Minor: a timing-sensitive test, a liveness gap that strict FIFO now makes wider, and a small style nit.

**1. Strengths**
- **Ticket allocation is race-safe.** Tickets are created, locked and pruned under the `alloc` mutex (`scripts/test-lock.py:116-123`, `82-108`). So a new ticket is never mistaken for a dead one, and pruning can't race a ticket being created. Because flock is held per open file, a waiter checking its own ticket correctly sees it as live. A new ticket is numbered "highest live + 1", so a newer waiter can never sort ahead of an older one.
- **One cleanup path for every exit.** Taking slots, timing out and nested runs all remove the ticket through `try/finally` (`scripts/test-lock.py:246-280`). A crashed waiter's ticket is pruned the next time anyone allocates or checks the line. This settles ARCH-FUNERAL: the only file that persists is the single `alloc` file.
- **Mixed fleets keep working.** The head still takes `.queue` (`:251`), so an older copy of the script that holds it goes first instead of both acting as head at once. A test pins this (`test-lock.test.sh:143-149`).
- **The tests actually reach the waiter process.** Waiters start as `python3` directly, so `$!` is the real waiter and `kill -STOP`/`kill -9` hit it. The stopped-waiter case also asserts that neither later waiter took a slot. The FIFO cases fail on the old code, as shown above.
- **The test now has a runner.** Moving it to `scripts/test/` puts it in the loop that `20-ci-workflow-tests.sh` runs. A grep finds no references to the old path outside the issue file, and the atlas is updated.

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
- **Tests depend on timing** (`scripts/test/test-lock.test.sh:119-121`, `134-135`). Arrival order between waiters is set only by `sleep 0.5`. Each waiter starts Python and runs `git rev-parse` before it takes a ticket, so on a loaded CI machine two waiters could swap order and the position checks would fail spuriously. The fix is to poll until the count of entries in `$tmp/fifo.tickets` (excluding `alloc`) reaches N before starting the next waiter.
- **A stopped waiter now blocks everyone behind it** (`scripts/test-lock.py:251`). Before, only a stopped head could block the line. Now any waiter that is alive but suspended, such as a foreground `make test` suspended with Ctrl-Z, blocks every later waiter until it resumes. That is inherent to strict FIFO, but the atlas only says "a crash never blocks the line". Either document that a suspended waiter holds its place, or skip tickets whose waiters stop updating them (e.g. a stale ticket mtime).
- **Style nit:** `Tickets.locked()` hand-writes a small context-manager class. `contextlib.contextmanager` would be shorter.
- **Untested path:** no test checks that the timeout path leaves no ticket behind. The release in `finally` covers it, but only the success path is checked.

**5. Test coverage notes**
- Cases covered: arrival order with a later waiter polling first, positions in the wait message, a dead waiter skipped, no ticket left behind, an older copy's head going first, then the slot handoff.
- I confirmed the cases fail without the fix, and they use only private lock files and sleep holders.
- Not covered: FIFO combined with weights, i.e. a weight-2 head with weight-1 waiters behind it. The existing weight tests partly cover this.

**6. Architectural notes**
- **ARCH-DRY: pass.** `live_tickets` repeats the try-flock pattern, but its handling differs (prune vs keep), so a separate function is justified.
- **ARCH-PURE: pass, given the scope.** This is a thin flock/IO script, and the queue order is the only logic.
- **ARCH-PURPOSE: pass.** It delivers the full Spec: tickets, dead waiters skipped, position in the message, the old properties kept, and the test given a runner.
- **ARCH-MOCK: N/A.** The only dependency is kernel flock, and the tests use it directly through private files.
- **ARCH-CONSTRAINTS: pass.** Each waiter takes the mutex and scans the directory once a second, and there are only a few waiters on one machine. The test runtime (about 32s) is in line with sibling tests.
- **ARCH-SECURE: pass.** Ticket and queue records are read through `holder()`, which checks their shape, and file names are checked with `isdigit`.
- **ARCH-ORDER: pass, with the timing caveat above.** The state is still `head`/`first` flags, but the legal states are few and visible in the code. The STOP/CONT approach does give the tests control over interleaving. Only the arrival order still relies on sleeps.
- **ARCH-FUNERAL: pass.** Each ticket lives exactly as long as its wait, dead waiters' tickets are pruned on the next scan, and `alloc` is a single persistent file.

**7. Plan revision recommendations**
None. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: sleep-ordered-test-arrival
    title: |
      FIFO tests order waiter arrival with sleep 0.5, which can flake under CI load
    detail: |
      scripts/test/test-lock.test.sh:119-121,134-135 start waiters 0.5s apart; each runs python startup plus git rev-parse before taking a ticket. Poll until the tickets dir holds N entries before starting the next waiter.
  - id: new
    severity: Minor
    family: strict-fifo-liveness
    title: |
      A stopped (Ctrl-Z) non-head waiter now blocks every later waiter indefinitely
    detail: |
      Before, only the head could stall the line. Either document in the atlas that a suspended waiter holds its place, or skip tickets whose waiters stop updating them (e.g. a stale mtime heartbeat).
  - id: new
    severity: Minor
    family: hand-rolled-context-manager
    title: |
      Tickets.locked() defines an ad-hoc context-manager class; contextlib.contextmanager is simpler
```
