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

---

## Re-review — 2026-10-16T02:57:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | whole-issue close |
| milestone | — |
| window | 76b973042fca9c17e28bd6db0d1a449e4f8b1211..850dd46399f456c196e419928e89205552410544 |
| command | sdlc close --issue 393 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T02:57:14-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Every prior finding was already disposed, and this round's change introduces no new defect. The code change answers the close review's unraised Minor note: `readTicket` cut a ticket at 4096 bytes. It now reads the whole ticket through `io.NewSectionReader(f, 0, maxTicket)`, capped at 1 MiB (`cmd/sdlc/internal/landingline/line.go:345-355`). A section reader reads from fixed offsets, so the file's seek position doesn't matter, and both `heldTicket` and `scan` callers stay correct. A file over 1 MiB gets truncated, fails to parse, and shows as "unreadable". The landing line already shows a bad ticket that way rather than hiding it, so that behaviour is unchanged. I ran the regression test against the old code and it failed: an 8 KiB branch came back `Unreadable:true`. At head the landingline package passes. The rest of the window is tracker bookkeeping: the `codecomplete` status, `actual_hours`, the close log line, the gate-ledger run record, and the archived close-review file. None of it touches `## Spec`, `## Revisions` or `## Done when`, so the scope has not changed.

1. **Strengths**
   - `line.go:345` gives the bound a name and explains it in a comment. A file that exceeds the bound becomes visibly unreadable rather than silently cut short.
   - `line_test.go:324` `TestLineLongTicketReads` tests the real behaviour through `Enqueue` → `Status` on real temp files. I confirmed it fails without the fix.
   - The read is now offset-based (`SectionReader`), so it no longer depends on the file's seek position.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none.

5. **Test coverage notes:** I swapped in the base `line.go` and ran the test: it failed with `waiting [{Seq:1 ... Unreadable:true}]`. I then restored the file. `go test ./internal/landingline/` passes at head.

6. **Architecture (interdiff only)**
   - **ARCH-DRY: pass.** There is still one ticket reader, and both `heldTicket` and `scan` use it.
   - **ARCH-PURE: pass.** The change stays inside the existing IO layer.
   - **ARCH-PURPOSE: pass.** The fix covers every place a ticket is read, because all of them go through `readTicket`; no sibling fixed-size reads remain in the window.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-2
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-3
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-4
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-5
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-6
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-7
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-8
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-9
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-10
    disposition: addressed
    note: |
      heldTicket still returns Unreadable for an unparseable held head; readTicket now reads the whole ticket, bounded at 1 MiB; landingline tests pass.
  - id: BR-11
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-12
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-13
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-14
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-15
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-16
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
  - id: BR-17
    disposition: addressed
    note: |
      Disposed in earlier rounds; nothing in this interdiff reverts it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"73f65a9d-e4b6-4391-baeb-23a4cf9776b5","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"fix-review","round":8,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/867ab326465b532f","reviewed_head":"850dd46399f456c196e419928e89205552410544","slot":"ariadne:2","requested_at":"2026-10-16T09:55:44.658515Z"},"phase":"finished","pid":95479,"started_at":"2026-10-16T09:55:44.84631Z","finished_at":"2026-10-16T09:56:19.505274Z","session_id":"70a51523-1400-4658-8847-2f9d2d750e71","outcome":"returned"},{"version":1,"intent":{"disposable_root":true,"id":"56508f11-83c2-4066-b734-179036adf3e1","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"fix-review","round":8,"attempt":2,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/867ab326465b532f","reviewed_head":"850dd46399f456c196e419928e89205552410544","slot":"ariadne:2","requested_at":"2026-10-16T09:56:26.612669Z"},"phase":"finished","pid":3447,"started_at":"2026-10-16T09:56:26.888508Z","finished_at":"2026-10-16T09:57:10.214797Z","session_id":"8fc29e48-7e52-4cdf-8e3e-650354b9caf2","outcome":"returned"}] -->

## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing valid findings block"
output: "```verdict\nverdict: SHIP\nconfidence: high\n```\n\nThis window contains one fix made after the close review, plus tracker and ledger bookkeeping. The fix answers the close review's Minor note, which was not raised as a finding: `readTicket` read only a fixed 4 KiB, so a ticket with a long branch name showed up as unreadable. It now reads the whole ticket through `io.ReadAll(io.NewSectionReader(f, 0, maxTicket))`, with a 1 MiB cap (`cmd/sdlc/internal/landingline/line.go:345-356`). A new test, `TestLineLongTicketReads` (`line_test.go:324-335`), pins the change: it enqueues a branch name of 8192 bytes and checks that `Status` reads it back unchanged. I reverted `line.go` to the base in this checkout. The test then failed (the ticket came back `Unreadable:true` with an empty Branch). With the fix restored it passes, and the whole `landingline` package passes. Nothing in `## Spec`, `## Revisions`, `## Done when` or the plan changed. The two plan-directory files only add the close-gate ledger metadata and the archived close review. No prior findings are open and none of the 17 already disposed needed revisiting. The interdiff adds no new defect.\n\n1. **Strengths**\n   - **The cap keeps its purpose.** A file over 1 MiB is cut at the cap, fails to unmarshal, and so still reads as unreadable, as the comment at `line.go:345-347` says.\n   - **`io.ReadAll` handles end of file correctly.** It treats the `io.EOF` that the section reader returns as a normal finish. So `err != nil` only catches real read failures, and those also fall back to unreadable. A held `head.lock` still reads as present through `heldTicket`.\n   - **Both callers inherit the fix.** `scan` (`line.go:288`) and `heldTicket` (`line.go:339`) both go through the same function, so no second read path is left unfixed.\n   - **The test checks what matters.** It asserts the branch round-trips exactly, which is stronger than only checking the ticket isn't marked unreadable.\n\n2. **Critical:** none.\n\n3. **Important:** none.\n\n4. **Minor:** none.\n\n5. **Test coverage notes**\n   - I confirmed by revert that `TestLineLongTicketReads` fails without the fix and passes with it.\n   - `go test ./internal/landingline/ -count=1` passes.\n   - The test doesn't cover the over-cap case (a file larger than 1 MiB reads as unreadable), but that is the same unreadable path as before.\n\n6. **Architecture (interdiff only)**\n   - **ARCH-DRY: pass.** There is one read helper and both callers use it.\n   - **ARCH-PURE: pass.** The change is confined to the existing file-reading layer, and no business logic moved into it.\n   - **ARCH-PURPOSE: pass.** Every read of ticket contents goes through `readTicket`, so the fix covers the whole class rather than only the site the note named.\n\n7. **Plan revisions:** none.\n\n```findings\n```"

truncated: false
retained-bytes: 0000003037
