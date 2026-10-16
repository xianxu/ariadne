# Boundary Review — ariadne#393 (whole-issue close)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | whole-issue close |
| milestone | — |
| window | 655bda2e67d4fcb701adda3c88f3bc526fae1f05..36b7cb950a0ee92184939d5f88da33c4b484ec5a |
| command | sdlc close --issue 393 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T02:54:11-07:00 |
| verdict | SHIP |

## Review

Still running the touched packages' tests. I'll check the partial output and meanwhile go through the M2 code.

I've finished the review: both open findings are fixed and nothing new turned up, so the verdict is SHIP.

```verdict
verdict: SHIP
confidence: high
```

The two open Minor findings from earlier rounds were fixed in commit c0305bcc. I re-checked both against the code and the plan, and both hold. The M2 wiring is consistent with the stated rule that "a merge without a turn acts only on an already-merged PR":
- **Before the line:** `runDurableMerge` asks for confirmation first, so a prompt never holds the line (`cmd/sdlc/landing.go`). It then skips the line only when `resumeOnly(pre)` holds.
- **After selection:** it checks the selected PR again, with `turn == nil && pr.State != "MERGED"` returning an error.

The landingline, flock and reviewbox packages pass. The 13 new `cmd/sdlc` tests also pass (run with `-run`). I found no new defects.

1. **Strengths**
   - `heldTicket` (`cmd/sdlc/internal/landingline/line.go`) now reports a held `head.lock` with unreadable contents as present but unreadable. It is used both in `Status` and in `tryAdmit`'s waiting-head report, so the rule "a held lock is never absent" lives in one place.
   - `scan` takes a ticket's number from its file name and still counts an unreadable ticket. Temp files are excluded, and stale tickets are removed only when nobody holds their flock.
   - `OnAdmit` runs outside the line mutex, and a comment explains why: it avoids a self-deadlock, because flock is held per open file.
   - `lineLog.record` warns once on a write failure and lets the landing continue. That makes lost audit data visible without blocking a merge.
   - `Order`, `NextHead`, `ValidGrants` and `Authorize` are pure, with no IO, and are tested directly.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** `readTicket` reads at most 4096 bytes. A ticket larger than that, such as one with a very long branch name, would show as "unreadable". That is safe (it degrades visibly), but it is cosmetic noise. I am not raising it as a finding.

5. **Test coverage notes**
   - **BR-10:** `TestLineUnreadableHeadIsStillAHead` writes garbage into a held `head.lock`. It then checks both `Status` and `tryAdmit` report a non-nil, Unreadable head. Without the fix, `readTicket` returns nil and the `got == nil` check fails, so the test pins the fix.
   - **New M2 tests** (`landingqueue_test.go`, `landingcmd_test.go`): all 13 pass, in about 190s total.
   - **Full `cmd/sdlc` package:** a run of the whole package hit go test's default 10-minute timeout here and reported FAIL. Each landing test takes 10–30s on real temp repos, so the package is just slow; I could not confirm whether a test was actually stuck. The targeted run passed, so this is not a finding against this change. The package needs a larger `-timeout` or the slow tests split out.

6. **Architecture**
   - **ARCH-DRY: pass.** The flock helper was moved out of reviewbox into `internal/flock` rather than copied. The waiting position and `NextHead` both come from `Order`.
   - **ARCH-PURE: pass.** The ordering, grant and retention rules are pure; `line.go`, `log.go` and `landingqueue.go` are the thin IO layer.
   - **ARCH-PURPOSE: pass.** Merge holds the line for its whole run, and status, grant and revoke are delivered.
   - **ARCH-MOCK: pass.** It reuses the stateful `landingFakeGH`, and the line itself is real files in a temp directory.
   - **ARCH-CONSTRAINTS: pass.** Each waiter polls every 2s, scanning a handful of local files, with no deadline beyond its context.
   - **ARCH-SECURE: pass.** The `seq` counter, ticket JSON, `head.lock` contents and log lines are all treated as untrusted, and bad input shows up visibly.
   - **ARCH-ORDER: pass.** State is only live flocks. Admission is serialized by the mutex plus the `head.lock` token, and the tests drive timing through the injected Sleep.
   - **ARCH-FUNERAL: pass.** Tickets are removed on release or by the next scan; the two lock files are fixed in number; the logs are swept after 30 days.
   - **For future work:** batching (`NextBatch`) can be built on `Order` without touching the line's files.

7. **Plan revisions:** none. The "as built" Core concepts entries and the note marking the Task 1 sketch as superseded now match the code.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      Plan lines 47-59 are now marked "as built" and give Status() (head *Ticket, waiting []Ticket, err error), matching line.go. Task 1's sketch at plan:88 is marked superseded (NextHead takes no now), and Task 2 line 125 uses the as-built Status signature.
  - id: BR-10
    disposition: addressed
    note: |
      line.go heldTicket returns Ticket{Unreadable: true} for a held but unparseable head.lock and is used in both Status and tryAdmit. TestLineUnreadableHeadIsStillAHead fails without the fix (it checks got != nil and Unreadable) and passes at head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"df00763e-3916-4bd4-a9e7-9271a5976258","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":7,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4bffd9a6ba7776f9","reviewed_head":"36b7cb950a0ee92184939d5f88da33c4b484ec5a","slot":"ariadne:2","requested_at":"2026-10-16T09:38:51.974656Z"},"phase":"finished","pid":6354,"started_at":"2026-10-16T09:38:53.678844Z","finished_at":"2026-10-16T09:54:08.909256Z","session_id":"fd9fd84c-8c04-45ea-b801-6a93c05ff35d","outcome":"returned"}] -->
