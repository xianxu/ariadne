# Boundary Review — ariadne#393 (milestone M1)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 655bda2e67d4fcb701adda3c88f3bc526fae1f05..17b06388290cedca6a6b94f80652583481b448df |
| command | sdlc milestone-close --issue 393 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T00:07:32-07:00 |
| verdict | REWORK |

## Review

Reviewing the M1 landingline package; running its tests and probing two suspected grant-validity edge cases now.

```verdict
verdict: REWORK
confidence: high
```

M1 delivers the landingline package as planned and its tests pass. `go test -race -count=1` is green for `landingline` and `reviewbox`, and `go vet` is clean. The ordering core is pure. Tickets and the head are held by flock and released by the kernel, and the death tests kill a real helper process with SIGKILL. The flock helper was lifted cleanly out of reviewbox. Three things block SHIP:

- **Plan table drift.** The plan's Core concepts table no longer matches the code. The deviations are written in the issue's `## Log` but not in the plan, and there is no `## Revisions` entry.
- **Future-dated grants.** `ValidGrants` accepts a grant whose `Granted` time is in the future. Such a grant gives priority right away, and the sweep never removes it. I confirmed this with a throwaway probe test.
- **Revokes the sweep can drop.** `AddGrant` doesn't stamp `Granted`, but the sweep keys grant age on `Granted`. A revoke written with a zero `Granted` is swept away, and the grant it revoked becomes valid again. I reproduced this with the same probe.

The second and third are cheap to fix and are the kind of bug M2's `revoke` will hit.

**1. Strengths**
- `order.go` is fully pure with no `time.Now`. `Order` is the single source for both `NextHead` and the waiting position (`order.go:69-97`), so the position line and admission can't disagree (ARCH-DRY).
- Tickets are created already locked: written under `.tmp-`, locked, then renamed (`line.go:95-118`). An abandoned temp file is pruned by the next scan, so no window exists where a live ticket looks dead.
- The death tests use a real helper process killed with SIGKILL (`line_test.go:150-217`) for both a waiter and a head. They prove the kernel releases the locks, not cleanup code.
- `OnAdmit` runs outside the mutex (`line.go:191-195`), and `TestAdmitHookMayRecord` pins that a hook which logs doesn't deadlock.
- The flock lift into `internal/flock` keeps reviewbox's behaviour, and the reviewbox tests still pass.

**2. Critical findings**
- **The Core concepts table contradicts the code** (`workshop/plans/000393-serial-landing-admission-plan.md:19-34`):
  - The table lists `DelayStats` in `stats.go`, which doesn't exist; the code has `Delays` in `log.go:68`.
  - The plan's `NextHead` takes `now`; the code's doesn't.
  - The plan's `Ticket` has `PR`; the code's doesn't.
  - The new pure `Order` isn't in the table at all.
  - Task 3's `stats.go` and its "DelayStats p50/p90" test aren't delivered as written.
  - **Fix:** update the table and add a `## Revisions` entry (see section 7).

**3. Important findings**
- **Future-dated grants are valid and never swept** (`order.go:59-60`, `log.go:122`). `ValidGrants` checks `now.Before(Expires)` and the 24 h span, but not `!Granted.After(now)`. A grant with `Granted` set to 2099 is valid now and stays valid until 2099. The sweep test `now.Sub(t) <= Retention` is always true for future times, so the entry is never removed either. That breaks both the "expiring ≤24 h" rule and ARCH-FUNERAL, and it is ARCH-SECURE: a hand-edited log entry is trusted.
  - **Fix:** require `!g.Granted.After(now)` in `ValidGrants`, add a table case for it, and have the sweep drop entries dated in the future.
- **A swept revoke brings its grant back** (`log.go:52-57`). `Record` stamps `At` when it is zero, but `AddGrant` doesn't stamp `Granted`. A revoke with a zero `Granted` is treated as 2,000 years old by the next daily sweep and dropped. The original grant, still within its 24 h, becomes the latest entry and is valid again.
  - **Fix:** in `AddGrant`, stamp `Granted` from the line clock when it is zero. Better still, key grant retention on `max(Granted, Expires)`.
  - **Test:** grant, then revoke with a zero `Granted`, then force a sweep; the grant must stay invalid.

**4. Minor findings**
- `scan` trusts the `Seq` in a ticket's JSON over its file name (`line.go:262-272`). The file name should be authoritative (ARCH-SECURE).
- `sweep` silently drops malformed lines (`log.go:122`). The plan says "skipped and counted"; at least count them, or keep them until they are old.
- If the handle has already been released, `Wait` never reaches position 0 and loops until its context ends (`line.go:150-157`). Return an error when `h.ticket.Seq` isn't among the live tickets.
- Line tests depend on timing: `notAdmitted` waits 150 ms of real time while polling every 1 ms. The injected `Sleep` seam exists but no test uses it to control interleaving (ARCH-ORDER).
- `NextHead` sorts `displaced` again (`order.go:96`), although the tickets are already in Seq order within the normal class. This is harmless.

**5. Test coverage notes**
- Plan Task 1 and Task 2 cases are covered: FIFO, overtaking, FIFO among fleet-blocking, expired, revoked, no grantor, over 24 h, latest entry wins, Authorize's four cases, a dead waiter, a dead head, no preemption, independent directories, and cancellation.
- Missing:
  - a future-dated grant;
  - a sweep that drops a revoke;
  - a malformed ticket JSON that is still held (the "never silently skipped" branch at `line.go:268` has no fixture);
  - DelayStats quantiles (deferred to `suiteQuantiles` in M2, so not testable yet).

**6. Architectural notes**

| Principle | Result | Notes |
|---|---|---|
| ARCH-DRY | pass | The flock helper is shared with reviewbox; test-lock shares only the design, and the atlas says so. |
| ARCH-PURE | pass | `order.go` is pure and its tests do no IO. |
| ARCH-PURPOSE | pass for M1 | — |
| ARCH-MOCK | pass | The line is real files in a temp directory; Git and GitHub come in M2. |
| ARCH-CONSTRAINTS | pass | A scan is O(waiters) every 2 s on local IO. |
| ARCH-SECURE | flag | The future-dated grant (and the minor `Seq` trust). |
| ARCH-ORDER | pass with minor | There are two states (ticket held, or head held), and all admission runs under `line.lock`. Tests can't inject an interleaving. |
| ARCH-FUNERAL | flag | Future-dated entries are never swept. Everything else has an end: tickets are pruned, the lock files are fixed, and the logs keep 30 days. |

M2 should:
- build `revoke` and the landed-path revoke on the fixed `AddGrant` stamp;
- record `refused` on every early-return path once `Wait` returns, so `Delays` stays meaningful.

**7. Plan revision recommendations**
- Add a `## Revisions` entry for 2026-10-16 (M1). It should:
  - change `NextHead` to `(tickets, grants)`, since `ValidGrants` already applies the clock;
  - add a pure `Order` row to the table;
  - remove `PR` from `Ticket`, since the `admit` event carries it;
  - replace `DelayStats` (`stats.go`) with `Delays` (`log.go`), feeding test-stats' `suiteQuantiles` in M2;
  - say the admitted head leaves the ticket set and records itself in `head.lock`.
- Update the Core concepts table and Task 3's file list to match.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: plan-table-drift
    title: |
      Plan Core concepts table contradicts the M1 code (DelayStats/stats.go, NextHead signature, Ticket.PR, missing Order)
    detail: |
      The plan table lists DelayStats in stats.go (absent; the code has Delays in log.go), NextHead taking now (it doesn't), and Ticket with PR (it doesn't), and omits the pure Order. The deviations are only in the issue Log. Add a plan Revisions entry and update the table and Task 3's file list.
  - id: new
    severity: Important
    kind: defect
    family: untrusted-log-entry-validation
    title: |
      ValidGrants accepts future-dated grants, and the sweep never removes them
    detail: |
      order.go:59-60 checks now.Before(Expires) and the 24 h span but not Granted <= now, so Granted=2099 confers priority until 2099. The sweep test now.Sub(t) <= Retention (log.go:122) is always true for future times, so the entry is never removed. A probe test confirmed it. Fix: require !Granted.After(now), drop future-dated entries in the sweep, and add a table case.
  - id: new
    severity: Important
    kind: defect
    family: untrusted-log-entry-validation
    title: |
      AddGrant doesn't stamp Granted, so a revoke with a zero Granted is swept and its grant becomes valid again
    detail: |
      Record stamps At when it is zero but AddGrant (log.go:52) doesn't stamp Granted, and the sweep ages grants by Granted. A revoke appended with a zero Granted is dropped by the next daily sweep, and the earlier grant, still within 24 h, becomes the latest entry and valid again (reproduced with a probe). Fix: stamp Granted from l.o.Now() when it is zero, and test grant, revoke, sweep.
  - id: new
    severity: Minor
    kind: defect
    family: untrusted-log-entry-validation
    title: |
      scan trusts the Seq in a ticket's JSON over its file name
    detail: |
      line.go:262-272 uses the parsed Ticket's Seq; a hand-edited ticket can claim another position. Use the file name as the authoritative Seq.
  - id: new
    severity: Minor
    kind: defect
    family: malformed-input-visibility
    title: |
      sweep silently drops malformed log lines instead of counting them
    detail: |
      log.go:122 discards unparsable lines without a count, which contradicts the plan's skipped-and-counted rule.
  - id: new
    severity: Minor
    kind: defect
    family: wait-termination
    title: |
      Wait on a handle whose ticket is gone polls forever until its context ends
    detail: |
      line.go:150-157: if h's Seq isn't among the live tickets, pos is never 0. Return an error instead.
  - id: new
    severity: Minor
    kind: defect
    family: ordering-test-seam
    title: |
      Line tests rely on wall-clock timing (150 ms notAdmitted) instead of the injected Sleep seam
    detail: |
      The interleaving isn't controllable or reproducible (ARCH-ORDER), and the tests may be flaky under load.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"44c551a2-f8fb-4d94-842e-f5cef1741ce2","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/565e06589b4d0944","reviewed_head":"17b06388290cedca6a6b94f80652583481b448df","slot":"ariadne:2","requested_at":"2026-10-16T07:05:18.296777Z"},"phase":"finished","pid":14769,"started_at":"2026-10-16T07:05:18.498697Z","finished_at":"2026-10-16T07:07:20.669233Z","session_id":"9218fbae-1512-4e98-9fae-cd38013d4370","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T00:13:11-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 655bda2e67d4fcb701adda3c88f3bc526fae1f05..45eea86e7073b7f34a82638e4293aba92d2e2d5d |
| command | sdlc milestone-close --issue 393 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T00:13:11-07:00 |
| verdict | REWORK |

## Review

Checking the round-1 fixes now. Next I'll run the tests and a probe for a suspected sweep bug.

Now the probe for live-grant deletion by the sweep.

```verdict
verdict: REWORK
confidence: high
```

The round-1 fixes mostly hold up. Each of the six code fixes has a regression test, and I checked two of them by reverting the fix: the test for BR-2 (`TestValidGrants`) and the test for BR-3 (`TestSweptRevokeKeepsTheGrantRevoked`) both fail without their fixes. The plan now has a Revisions entry, and the line tests step through `tryAdmit` with `Sleep` injected. One problem blocks SHIP, and the BR-2/BR-3 fixes caused it. The grants sweep now dates a grant by `max(Granted, Expires)` and also drops anything "dated in the future". A grant that hasn't expired yet always has `Expires > now`, so whenever the daily sweep runs on an `AddGrant` call, it deletes every grant that is still in force. I confirmed this with a probe test (below). The atlas describes the same broken rule.

**1. Strengths**
- `order.go:53-109`: the pure core is small and deterministic. `Order` is the single source for both the waiting position and `NextHead`, and `displaced` comes from the same order (ARCH-DRY).
- `line.go:18-29, 251-295`: the line's state lives in flocks the kernel owns. A dead waiter or a dead head leaves nothing inconsistent behind. Since the round-1 fix, the file name decides the ticket number, not the JSON inside it.
- `line_test.go`: each interleaving is now chosen by the test (`step`/`wantPos`, injected `Sleep`). The helper-process tests check readiness through stdout and wait for `cmd.Wait` after killing the helper, so they don't depend on wall-clock timing.
- `internal/flock` is lifted cleanly out of reviewbox, and reviewbox's tests still pass (ARCH-DRY).

**2. Critical**
- `log.go:62-71` together with `log.go:144`: any daily sweep run by `AddGrant` deletes every grant that is still in force (details in the findings block). This undoes the main feature: a renewal or a second grant silently removes the first issue's priority.

**3. Important**
- None beyond the Critical finding. The atlas text at `atlas/workflow/sdlc-binary.md` ("not in the future … the later of its grant and expiry times") states the broken rule and has to change with the code.

**4. Minor**
- Plan drift: the Task 1 code sketch still has `NextHead(..., now time.Time)` (plan:80), and Task 2 still gives `Status() ([]Ticket, headPID int)`. The code has `Status() (*Ticket, []Ticket, error)`.
- `line.go:303`: `nextSeq` ignores a parse error. A hand-edited `seq` file resets the counter to 1, and `os.Rename` can then overwrite a live ticket's file. This belongs in the rule enumeration of the Critical finding.

**5. Test coverage notes**
- All tests pass at the reviewed head (`landingline` and `reviewbox`).
- No test runs a sweep while a grant is still in force. That gap is how the regression shipped. The fix should be an invariant test across event sequences, not a single extra case.
- BR-6: only the "released" path is tested. The "ticket file removed by someone else" path (`line.go:167`) is never reached by a test. This is acceptable for a Minor finding.

**6. Architecture**
- **ARCH-DRY: pass.** flock is shared, and `Order` is the single ordering source.
- **ARCH-PURE: flag.** The sweep's decision about which entries to keep is buried in the IO routine (`log.go:144`) and depends on a closure, so no pure test touches it. That is where the bug sits. Extracting it as a pure function is part of the Critical fix.
- **ARCH-PURPOSE: pass,** apart from the Critical bug, which breaks the purpose of grants.
- **ARCH-MOCK: pass.** The only dependencies are the filesystem and flock, which run for real in temp directories. A real subprocess is the fake for a process dying.
- **ARCH-CONSTRAINTS: pass.** M1 sets the poll interval (2 s) and a 1 MiB scanner line cap. The logs are small.
- **ARCH-SECURE: flag.** The rule for which log entries to keep doesn't follow from the validity rule (see the Critical finding). The `seq` parse is also untrusted input (Minor).
- **ARCH-ORDER: pass.** Admission is one transition under the mutex, and the tests now pick the interleaving. The sweep-versus-live-grant ordering is the event sequence that was missed (it is the Critical finding).
- **ARCH-FUNERAL: pass with a caveat.** Both logs have a 30-day sweep, ticket files are pruned by liveness, and the `.swept`/`.tmp` sidecars are bounded. The caveat is that the sweep now removes too much.

**7. Plan revision recommendations**
- Add a Revisions entry: "Sweep retention is `keepGrant(g, now)` / `keepEvent(e, now)`, pure functions in order.go. Invariant: the valid grants are the same before and after a sweep (`ValidGrants(sweep(L, now), now) == ValidGrants(L, now)`)."
- Update the Task 1 sketch and the Task 2 `Status` signature to match the code, or mark them as superseded by the Revisions.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan table now lists Order, Delays in log.go, NextHead without now, Ticket without PR, plus a Revisions entry. Residual Task 1 sketch and Status signature drift is noted as Minor.
  - id: BR-2
    disposition: addressed
    note: |
      order.go:61 adds the Granted-not-after-now check; reverting it turns TestValidGrants red. The sweep drops future-dated entries, but that fix interacts with BR-3's Expires key and caused the new Critical finding.
  - id: BR-3
    disposition: addressed
    note: |
      AddGrant stamps Granted (log.go:59); reverting it turns TestSweptRevokeKeepsTheGrantRevoked red. The revoke-outlives-grant goal is met; the new Critical finding covers the keying regression.
  - id: BR-4
    disposition: addressed
    note: |
      line.go:283-291 takes Seq from the file name; TestLineTicketNumberComesFromItsName covers both a lying JSON body and garbage.
  - id: BR-5
    disposition: addressed
    note: |
      The sweep stamp records malformed=N; TestLogSweepDropsFutureDatedAndCountsMalformed checks it.
  - id: BR-6
    disposition: addressed
    note: |
      tryAdmit errors on a released or vanished ticket (line.go:156-169); TestLineReleasedTicketCantWait covers the released path.
  - id: BR-7
    disposition: addressed
    note: |
      Line tests step tryAdmit directly and inject Sleep; helper tests sync on stdout readiness and cmd.Wait, not timers.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: untrusted-log-entry-validation
    title: |
      The grants sweep deletes every grant still in force: retention keys on Expires, and the sweep drops any entry dated in the future
    detail: |
      log.go:67-70 dates a grant by Expires when Expires > Granted, and log.go:144 keeps an entry only when 0 <= now - t <= Retention. A grant still in force has Expires > now, so whenever the daily sweep runs on an AddGrant call (a renewal, another issue's grant, a revoke), all grants still in force are erased. Probe: AddGrant(issue 1) at t0, AddGrant(issue 391, 24 h) at t0+23h, AddGrant(issue 7) at t0+25h, then ValidGrants has no 391 (FAIL). The atlas states the same rule. This is the 4th finding in family untrusted-log-entry-validation (BR-2 and BR-3 fixed instances, and this one came from the interaction of those fixes). Do not patch this instance alone. Rule: an entry's removal is decided by a pure keep(entry, now) function in the core, beside ValidGrants, and must preserve the invariant ValidGrants(sweep(L, now), now) == ValidGrants(L, now) for every log L. For grants: drop when malformed, when Granted is after now (forged), or when max(Granted, Expires) + Retention is before now. Events keep the At-based rule. Test the invariant with a table or sequence test over a grant still in force, an expired grant, a revoke, a future-dated grant and a zero-Granted entry. Also apply the rule to the untrusted seq file (line.go:303, where a parse error is silently ignored), and fix the atlas paragraph.
  - id: new
    severity: Minor
    kind: defect
    family: plan-table-drift
    title: |
      Plan Task 1 sketch still passes now to NextHead, and Task 2 gives a Status signature the code doesn't have
    detail: |
      This is the 2nd finding in family plan-table-drift. Rule: the Core concepts table and the Revisions entry are authoritative, and every signature elsewhere in the plan either matches them or is marked superseded. Here that covers the sketch at plan:80 and Status() ([]Ticket, headPID int), against the code's Status() (*Ticket, []Ticket, error).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"76cf42fa-0e32-47a3-bf0e-8b9039b6f1de","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2ca82bc177b46e19","reviewed_head":"45eea86e7073b7f34a82638e4293aba92d2e2d5d","slot":"ariadne:2","requested_at":"2026-10-16T07:10:49.616945Z"},"phase":"finished","pid":68483,"started_at":"2026-10-16T07:10:49.915698Z","finished_at":"2026-10-16T07:13:09.394268Z","session_id":"99c7596e-ab46-4003-94ac-79e39e672cee","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T00:17:35-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 655bda2e67d4fcb701adda3c88f3bc526fae1f05..60c6724cea01c7460dff564a5adf2af2097788bb |
| command | sdlc milestone-close --issue 393 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T00:17:35-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking BR-8's fix now: running tests, then reverting the fix to confirm the regression test goes red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

BR-8, the blocking finding from round 2, is fixed, and I checked the fix by breaking it. Retention now lives in two pure functions in `retention.go` (`KeepEvents` and `KeepGrants`), and `sweep` only applies them. Grants are aged by `Granted` alone, and when an issue's latest entry is stale, its whole history goes with it. I put the old rule (age by the later of `Granted` and `Expires`) back into the checkout: `TestSweepKeepsGrantsInForce`, `TestKeepGrantsPreservesValidity` and `TestKeepGrantsCases` all failed. With the fix restored, the touched packages pass and `go vet` is clean. The seq counter is now treated as untrusted input, as BR-8 also asked, and the atlas paragraph is fixed. Nothing blocks the gate. What's left is plan text that still contradicts the code (BR-9, only partly fixed) and one Minor case where an unreadable `head.lock` is reported as "no head".

1. **Strengths**
   - `retention.go:31-42`: the keep rule is pure and stated as an invariant, and `retention_test.go:25-50` checks it over 2000 random logs that include old, out-of-order, revoked and future-dated entries. It tests the rule itself, not a copy of the code.
   - `log.go:113-150`: `sweep` is now only IO. It parses, asks the rule what to keep, then rewrites the file atomically and records the malformed-line count in its stamp.
   - `line.go:302-322` and `line.go:108-113`: a malformed `seq` file now refuses instead of resetting to 0, and a counter that went backwards can't overwrite a live ticket. `TestLineCounterIsUntrustedInput` covers both.
   - `scan` (`line.go:248-290`) removes orphaned `.tmp-` files before checking their prefix, so a crash in the middle of `Enqueue` leaves nothing behind (ARCH-FUNERAL).

2. **Critical:** none.

3. **Important:** none new. BR-9 stays open below.

4. **Minor**
   - `line.go:205-206` and `line.go:174-175`: when `head.lock` is held but its contents don't parse, `readTicket` returns nil, so `Status` reports "no head" while a head is actually running. `scan` already handles the same case correctly: a held ticket whose content won't parse still counts. This is the 2nd finding in family `malformed-input-visibility`.
   - Plan drift that is still open (BR-9, details in the findings block).

5. **Test coverage**
   - The regression tests for BR-8 go red when the fix is reverted (confirmed).
   - The invariant is only checked at the sweep time `now`. A future-dated latest entry is dropped along with its issue's history, even though `ValidGrants` would accept it once its time arrives. On one machine clock skew is about zero, so this is fine, but the scope should be stated.
   - No test holds `head.lock` with garbage content.

6. **Architecture**
   - ARCH-DRY passes: there is one generic `appendLog`/`sweep`/`readLog`, and the flock code was moved into `internal/flock` rather than copied.
   - ARCH-PURE passes: ordering, validity and retention are pure and tested without IO.
   - ARCH-PURPOSE passes for M1: it delivers the package only, as planned.
   - ARCH-MOCK passes: real temp files, with no external service in M1.
   - ARCH-CONSTRAINTS passes: an O(waiters) scan and a sweep at most once a day.
   - ARCH-SECURE: one Minor gap, the `head.lock` case above.
   - ARCH-ORDER passes: all writes happen under the mutex, `tryAdmit` steps admission so the tests choose the interleaving, and `OnAdmit` runs outside the mutex to avoid deadlock.
   - ARCH-FUNERAL passes: tickets are pruned, logs are kept by a pure rule, and the stamp and lock files are fixed in number.
   - For M2: `readLog` uses a 1 MB scanner buffer, but `sweep` has no line limit, so one oversized line would make every `GrantEntries` call fail. That failure would at least be visible.

7. **Plan revisions needed**
   - A Revisions entry saying the Core concepts `Log` (with `Append`/`Sweep`) is built as `Line.Record`/`AddGrant`/`Events`/`GrantEntries` plus the pure `KeepEvents`/`KeepGrants`.
   - Ticket files are named `tickets/<%020d seq>`, without `.json`.
   - `Open` takes `(dir, Options)`.
   - Task 3's "Sweep drops entries older than 30 days" should be marked superseded by the round-2 retention rule.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      retention.go KeepGrants ages by Granted and drops an issue's whole history when its latest entry is stale. Reverting to the old expiry-based rule makes TestSweepKeepsGrantsInForce, TestKeepGrantsPreservesValidity and TestKeepGrantsCases fail. nextSeq now refuses a malformed seq (TestLineCounterIsUntrustedInput), and the atlas is fixed.
  - id: BR-9
    disposition: not-addressed
    note: |
      Fixed: the plan:81 sketch is marked superseded and the Task 2 Status signature matches the code. Still contradicting the code: the Core concepts Integration row and paragraph for Log (Append(file, v) and Sweep(file, now) don't exist; the methods are on Line), tickets/<seq>.json (the code uses %020d with no extension), Open(dir) (the code takes Options), and Task 3's "Sweep drops entries older than 30 days" (superseded by KeepGrants). Apply the rule to the whole plan, not just the named sites.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: malformed-input-visibility
    title: |
      Status reports no head when head.lock is held but its contents don't parse
    detail: |
      This is the 2nd finding in this family. In line.go:205-206, readTicket returns nil for a held but unparseable head.lock, so Status (and OnWait's head, line.go:174) says "no head" while one is running. Rule: a held lock whose contents are unreadable is reported as present but unreadable, never as absent (scan already does this for tickets). Fix: return a non-nil Ticket with an unreadable marker. Test: hold head.lock with garbage and assert head != nil.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"605efb05-bcd4-4f96-bd0c-56e3c60e4cef","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e70d424df0778b09","reviewed_head":"60c6724cea01c7460dff564a5adf2af2097788bb","slot":"ariadne:2","requested_at":"2026-10-16T07:15:25.276429Z"},"phase":"finished","pid":26691,"started_at":"2026-10-16T07:15:25.486136Z","finished_at":"2026-10-16T07:17:24.822424Z","session_id":"2fb5275e-2a0c-426c-8c42-c1f2b73d72be","outcome":"returned"}] -->
