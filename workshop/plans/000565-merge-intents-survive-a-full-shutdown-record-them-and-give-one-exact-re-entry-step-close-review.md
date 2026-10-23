# Boundary Review — ariadne#565 (whole-issue close)

| field | value |
|-------|-------|
| issue | 565 — Merge intents survive a full shutdown: record them and give one exact re-entry step |
| repo | ariadne |
| issue file | workshop/issues/000565-merge-intents-survive-a-full-shutdown-record-them-and-give-one-exact-re-entry-step.md |
| boundary | whole-issue close |
| milestone | — |
| window | 668e7637556b4748178be10e69ee969922c01e29..f0cf2e42a200e4047a9d3c69150cfdb901fa0c8a |
| command | sdlc close --issue 565 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T11:49:45-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the diff now; next I'm checking signal handling and the pause path in the queue.

Tests pass. Now I'm writing a quick throwaway test to confirm a race I suspect between status readers and `Hold`.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix one race, then ship.** The change delivers what the issue asks for, and the Spec and both Done-when items are met. `sdlc merge` now writes and flock-holds a durable intent record for its branch. A terminal outcome removes the record; an interruption or a kill leaves it behind as an orphan. `sdlc landing status`, `sdlc state` and `sdlc state --json` (`merge_intents`) name the orphan and print `sdlc merge --branch B --yes`. Re-entry reuses the recorded suite on an unchanged head and reruns it when main moved. All the new tests pass (`go test ./internal/landingline -run Intent` and `go test . -run 'MergeIntent|OrphanedMerge|MergeReentry'`).

One Important defect needs fixing before the boundary, and it is cheap. Status readers and the sweep take the same exclusive try-lock that marks a merge as running. So a re-entry that runs at the same moment as any status read or any other slot's merge start is refused with "a merge of B is already running (pid P…)", where P is the dead runner. I confirmed this with a throwaway test against the pinned head. After a full shutdown, every slot re-enters at about the same time, so this race hits the exact case the issue exists for.

1. **Strengths**
   - `intent.go:51-73` reuses the landing line's strict, field-preserving codec (`decodeStrict` / `encodeWithExtra`). Unknown fields survive a rewrite, including on re-entry (`mergeExtra` at `intent.go:189`), and duplicate keys, wrong types and other versions are refused (ARCH-EVOLVE, ARCH-DRY).
   - Liveness comes from the flock alone. The kernel releases it when the process dies, so a SIGKILL or a reboot cannot leave a record that looks live. `TestIntentOfAKilledRunnerIsOrphaned` kills a real child process to prove it.
   - The re-entry command is built from the validated branch name, not copied from the file (`mergeintent.go:104`). `validBranch` rejects option-like and shell-syntax names (ARCH-SECURE).
   - `runDurableMerge` (`landing.go:343-351`) decides keep-or-remove from the merge's own returned error via `queueInterruptionReason`, so a graceful cancel keeps the record. This was the case the plan flagged as most likely to go wrong.
   - `TestMergeReentryReusesThePassedSuite` checks both cases end to end: one suite run on an unchanged head, two after main moved.

2. **Critical findings:** none.

3. **Important findings**
   - **Readers block re-entry and make it name a dead pid (ARCH-ORDER)** — `intent.go:172-180` and `intent.go:257-271`.
     - `observe` (used by both List and Sweep) takes the same exclusive non-blocking lock that `Hold` uses to detect a running merge.
     - `Sweep`'s remove callback runs a `git rev-parse` subprocess while it holds that lock (`mergeintent.go:68-70`). Every `sdlc merge` start and every `landing status` sweeps every intent, so each orphan is briefly locked by someone else many times over.
     - If `Hold` lands inside one of those windows, it refuses with `IntentHeldError`, naming the dead runner's pid and telling the operator to "wait for it or stop it". My throwaway test against the pinned head reproduced this refusal.
     - No test injects a reader between the orphan and the re-entry; every current test sees only one ordering.
     - Fix sketch:
       - In `Hold`, retry the try-lock for a short bounded time (for example 2 s) before refusing.
       - In `Sweep`, work out whether the branch exists before taking the lock, and only re-check the file under the lock.
       - Add a regression test that holds an observe lock while `Hold` runs.

4. **Minor findings**
   - **Printed fields are not quoted (ARCH-SECURE):** `formatMergeIntent` (`mergeintent.go:124`) prints `Root` with `%s`, and the pause note prints owner and reason unquoted. `Valid` only checks that root is absolute and clean, so a hand-edited root could inject newlines into status output. The plan promised every printed field would be `%q`-quoted.
   - **`Hold` overwrites unreadable records (ARCH-EVOLVE):** a same-branch record that does not parse (including a future v2 record) is silently replaced, and its extra fields are dropped. The atlas says unreadable intents are "never acted on" (`intent.go:181-185`).
   - **Repeated git lookup (ARCH-DRY):** the `rev-parse --verify --quiet refs/heads/<branch>` lookup appears 3 times in `mergeintent.go`; one `localBranchHead` helper would cover them.
   - **`sdlc state` hides unreadable intents:** its filter compares `r.Root`, which is empty for an unreadable record, so these only show in `landing status`.

5. **Test coverage notes**
   - The codec, classification, hold/release, sweep, a SIGKILLed child process, merge lifecycle (landed, refused, interrupted, second runner), both status surfaces plus JSON, and suite reuse are all covered with real files in temp dirs and the existing landing fakes.
   - Gaps:
     - Status/merge-start readers racing re-entry (above).
     - The pause note is only unit-tested, not tested through `runLandingStatus`.
     - The surface test simulates the shutdown with an injected `context.Canceled`, not a kill. The kill is proven only at the store level, which is enough given the flock semantics.

6. **Architecture, marker by marker**
   - **ARCH-DRY: pass**, apart from the minor git-lookup repetition.
   - **ARCH-PURE: pass.** `ClassifyIntent` and `reportMergeIntent` are pure; IO lives in `Intents` and `observeMergeIntents`.
   - **ARCH-PURPOSE: pass.** Both Done-when items are met. Couch reading `merge_intents` is a separable follow-up in another repo.
   - **ARCH-MOCK: pass.** The tests use the real filesystem at a temp root and the existing git/GitHub fixtures.
   - **ARCH-CONSTRAINTS: pass.** The work per call is small and bounded.
   - **ARCH-SECURE: pass with a Minor.** Input is parsed and validated at the boundary; only the output quoting is missing.
   - **ARCH-ORDER: flag.** The reader/owner lock interleaving is unmodelled and untested (the Important finding).
   - **ARCH-FUNERAL: pass.** Records end through the terminal release, the stale sweep, the 30-day bound, and in-place rewrite on re-entry.
   - **ARCH-EVOLVE: pass with a Minor.** The overwrite of unreadable records drops their extras.

   For upcoming work: the Couch restart plan should treat `IntentHeldError` as possibly transient until the reader-lock fix lands.

7. **Plan revision recommendations**
   - Add a Revisions entry recording that readers take the same exclusive try-lock as the owner, and that `Hold` therefore waits for a bounded time before refusing. The plan's lifecycle table ("a lock already held means refuse") currently treats any lock holder as a live merge.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: reader-lock-masquerades-as-owner
    title: |
      Status/sweep readers take the owner's exclusive try-lock, so a concurrent re-entry refuses as "already running" naming the dead pid
    detail: |
      Intents.observe (intent.go:257-271) try-locks LOCK_EX on every intent, and Sweep runs a git rev-parse subprocess while holding it (mergeintent.go:68-70); every sdlc merge start and landing status sweeps all intents. A Hold during that window gets IntentHeldError naming the dead runner's pid ("wait for it or stop it"). A throwaway test against the pinned head reproduced the refusal. A fleet re-entering after a full shutdown is exactly when these collide. Fix: Hold retries for a bounded time (about 2 s) before refusing; Sweep works out branch existence before locking; add a regression test that holds an observe lock while Hold runs.
  - id: new
    severity: Minor
    kind: defect
    family: printed-untrusted-fields-unquoted
    title: |
      formatMergeIntent prints Root and pause owner/reason unquoted, against the plan's every-field-%q promise
    detail: |
      Valid only requires an absolute, clean root, so a hand-edited intent root (or pause.json reason) can put newlines or control characters into landing status / sdlc state output (mergeintent.go:124, mergeintent.go:106-110).
  - id: new
    severity: Minor
    kind: defect
    family: tolerant-writer-drops-unknown-fields
    title: |
      Hold silently overwrites an unreadable same-branch intent, dropping a newer version's fields
    detail: |
      When readIntent fails (corrupt or IntentVersion 2), Hold rewrites the file without merging extras (intent.go:189-200), although the atlas says unreadable intents are never acted on.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-git-lookup
    title: |
      The local branch-head rev-parse lookup is repeated three times in mergeintent.go
    detail: |
      holdMergeIntent, sweepMergeIntents and observeMergeIntents each build the same rev-parse --verify --quiet refs/heads/B call; one helper would cover all three.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"081ccaa4-559f-458b-b02b-92bbb8ae6fe7","repository":"github.com/xianxu/ariadne","issue_id":"000565","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/05cafbac3a313b39","reviewed_head":"f0cf2e42a200e4047a9d3c69150cfdb901fa0c8a","slot":"ariadne:5","requested_at":"2026-10-23T18:46:25.899641Z"},"phase":"finished","pid":22195,"started_at":"2026-10-23T18:46:26.07681Z","finished_at":"2026-10-23T18:49:45.028308Z","session_id":"9ba7866f-9262-44b0-8440-8bbff34449c0","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T11:56:03-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 565 — Merge intents survive a full shutdown: record them and give one exact re-entry step |
| repo | ariadne |
| issue file | workshop/issues/000565-merge-intents-survive-a-full-shutdown-record-them-and-give-one-exact-re-entry-step.md |
| boundary | whole-issue close |
| milestone | — |
| window | 668e7637556b4748178be10e69ee969922c01e29..1a8a6b06a228bf8ae0bfb55ca8467de46e32b010 |
| command | sdlc close --issue 565 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T11:56:03-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass on the head. Next I'm reverting the BR-1 fixes to check that the regression test fails without them.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All four prior findings are fixed, and each fix has evidence. The BR-1 regression test (`TestIntentReadersAreNotOwners`) is real. I reverted each half of the fix in the disposable checkout. Going back to an exclusive reader lock fails with `a second reader saw the first as a running merge`. Disabling Hold's patience makes the test hang until the timeout. The sweep now decides with no lock held and removes a record only under a brief exclusive lock, after checking the record is unchanged. The test covers the race where a re-entry lands between the sweep's read and its removal. All landingline tests and the targeted `cmd/sdlc` intent tests pass on the head, and `go vet` is clean. Nothing blocks. Two Minor items remain: one more unquoted printed field from the shared directory (the 2nd finding in `printed-untrusted-fields-unquoted`), and the plan's Core concepts table still uses names the code no longer has.

**1. Strengths**
- `intent.go:297-310` / `:343-357`: the lock roles are clean. Readers take a shared lock. Only a running merge holds the exclusive lock for long, and the sweep takes it only for a moment. `removeUnchanged` checks branch, pid and start time, so a re-entry that happens meanwhile is never deleted.
- `intent.go:215-224`: Hold refuses rather than overwrite a corrupt or later-version record, and still merges extras on a valid takeover. This keeps ARCH-EVOLVE intact. `TestIntentHoldNeverOverwritesAnUnreadableRecord` checks the bytes are unchanged.
- `mergeintent.go:101-123`: the report is a pure function. The re-entry command is built from the validated branch, never copied from the file. `ClassifyIntent` is the single reading rule (ARCH-ORDER).
- `mergeintent.go:66-68`: `localBranchHead` is now the one source for the branch-existence lookup, so the whole class is covered, not just one site.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- `mergeintent.go:129`: `r.Path` is printed with `%s`. Path is the line dir plus a file name read from the shared `intents/` directory (`intent.go:290` accepts any regular, non-dot name), so a hand-placed file name containing a newline can forge a status line.
  - **This is the 2nd finding in family `printed-untrusted-fields-unquoted`.** The rule: every string field in `formatMergeIntent` that comes from the shared directory (file contents *or* directory entries) is printed with `%q`. Only values sdlc builds itself are printed raw: `Reentry` (from the validated branch), the abbreviated validated head, ints and times. Apply the rule across `formatMergeIntent`. Path is the only remaining violation.
- Plan Core concepts table vs code: the names differ.
  - `MergeIntent` is `Intent`, and `IntentStore` is `Intents`.
  - `intentReentry` is `reportMergeIntent`/`formatMergeIntent`.
  - `runDurableMergeAttempt` is `runDurableMerge`, which holds the intent once across all train-reformation retries.
  - `IntentStore.List` returns `[]ObservedIntent`, not `(observed, unreadable)`.
  - These are renames with the same behaviour, so I rate this Minor rather than Critical, but the Revisions section should record them.
- `intent.go:190-194`: Hold creates the file before it locks it. A reader that lands in that window sees an empty, unlocked record and reports `unreadable … "intent: empty record"` with "inspect it, then remove it". The window is narrow and transient, but an operator who removed the file then would unlink a live merge's intent. One option is to treat an empty, unlocked file as live or skip it in reports.
- `intent.go:349`: while a sweep holds its brief exclusive lock, a concurrent reader in another process classifies the orphan as live, so it briefly disappears from `landing status`. This is cosmetic.

**5. Test coverage notes**
- Revert-and-test confirmed both halves of BR-1, as described above.
- BR-3 is pinned by a byte-equality check on both the corrupt and the later-version records.
- BR-2's root quoting is pinned by the `"/w/evil\nforged line"` case in `TestReportMergeIntent`. Pause owner and reason are pinned by the quoted expectation.
- Gap: if Hold's patience is ever removed, the `<-waited` handshake hangs until the test timeout instead of failing fast. Adding a `select` with a timeout would turn that into a clear failure. This is a nit.

**6. Architectural notes (one line per ARCH marker)**
- **ARCH-DRY:** pass. The git lookup is consolidated, and `files()` is shared by `List` and `Sweep`.
- **ARCH-PURE:** pass. `ClassifyIntent` and `reportMergeIntent` are pure, and the store is a thin IO shell.
- **ARCH-PURPOSE:** pass. The record, both status surfaces, the `--json` field, the recovery catalog and the atlas all derive from one store.
- **ARCH-MOCK:** pass. The flock store runs on a real temp dir, a SIGKILLed child simulates a shutdown, and git goes through the `gitRunner` seam.
- **ARCH-CONSTRAINTS:** pass. Hold's 2 s patience is bounded and declared. Readers hold their lock only while reading.
- **ARCH-SECURE:** pass, apart from the Path quoting above. Records are parsed and validated at the boundary, and size is capped (`maxTicket`).
- **ARCH-ORDER:** pass. The reader/sweep/re-entry interleavings now have a seam (`onHoldWait`, plus re-entering from inside `branchExists`) that pins one ordering on purpose.
- **ARCH-FUNERAL:** pass. Records end by terminal release, by the stale-branch sweep or by the 30-day age bound. Unreadable records can only be removed by the operator, which is acceptable for a rare, hand-damaged artifact.
- **ARCH-EVOLVE:** pass. Extras are preserved on takeover. A record from a later version is refused, never overwritten.

**7. Plan revision recommendations**
- Add a `## Revisions` entry: "Core concepts names as built: `Intent` (was `MergeIntent`), `Intents` (was `IntentStore`; `List` returns `[]ObservedIntent` with per-record `Err`), `reportMergeIntent`/`formatMergeIntent` (was `intentReentry`). The hold wraps `runDurableMerge` once across train-reformation retries (was `runDurableMergeAttempt`)."

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      observeIntent takes flock.Shared, Sweep decides with no lock and removes via removeUnchanged, Hold waits IntentHoldPatience; TestIntentReadersAreNotOwners fails on revert of either half (verified in checkout).
  - id: BR-2
    disposition: addressed
    note: |
      Root and pause owner/reason are now %q (mergeintent.go:117,135), pinned by the forged-root and quoted pause tests; the remaining Path sibling is raised separately.
  - id: BR-3
    disposition: addressed
    note: |
      Hold refuses a non-empty unreadable or other-branch record (intent.go:215-223); TestIntentHoldNeverOverwritesAnUnreadableRecord asserts the bytes are unchanged.
  - id: BR-4
    disposition: addressed
    note: |
      localBranchHead (mergeintent.go:66) is the single lookup used by hold, sweep and observe.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: printed-untrusted-fields-unquoted
    title: |
      formatMergeIntent prints the unreadable record's Path unquoted; Path embeds a file name from the shared intents directory
    detail: |
      2nd in family. Rule: every string in formatMergeIntent that comes from the shared directory (contents or entry names) is printed with %q; only sdlc-built values (Reentry, abbreviated head, ints, times) are raw. Path at mergeintent.go:129 is the only remaining violation.
  - id: new
    severity: Minor
    kind: defect
    family: plan-table-drifts-from-code
    title: |
      Plan Core concepts table names MergeIntent, IntentStore, intentReentry and runDurableMergeAttempt, but the code has Intent, Intents, reportMergeIntent and runDurableMerge
    detail: |
      Same behaviour under renamed entities, and List returns []ObservedIntent rather than (observed, unreadable). Add a Revisions entry recording the names as built.
  - id: new
    severity: Minor
    kind: defect
    family: reader-lock-masquerades-as-owner
    title: |
      A reader in the window between Hold's O_CREATE and its lock reports a live merge's empty file as unreadable and tells the operator to remove it
    detail: |
      intent.go:190-194 creates the file before taking the exclusive lock, so observeIntent can see an empty, unlocked record ("intent: empty record"). Lock-state rule: no reader classification should advise an action on a record whose writer may hold it next; treat an empty, unlocked record as transient (skip it or report it live) rather than unreadable.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a9ee95a4-0d3a-447f-9f61-7f0b5998e917","repository":"github.com/xianxu/ariadne","issue_id":"000565","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/ce0e68e871df4120","reviewed_head":"1a8a6b06a228bf8ae0bfb55ca8467de46e32b010","slot":"ariadne:5","requested_at":"2026-10-23T18:53:33.223165Z"},"phase":"finished","pid":4312,"started_at":"2026-10-23T18:53:33.405341Z","finished_at":"2026-10-23T18:56:03.148402Z","session_id":"014f220f-5b96-4c2e-8786-93617c6ced6d","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T12:00:56-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 565 — Merge intents survive a full shutdown: record them and give one exact re-entry step |
| repo | ariadne |
| issue file | workshop/issues/000565-merge-intents-survive-a-full-shutdown-record-them-and-give-one-exact-re-entry-step.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5a654a298c960eb38361ea37eae6e5fb80ff81a2..5f60907996bc7a32ea2453e72e625043a4780a04 |
| command | sdlc close --issue 565 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T12:00:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass at head. Next I'll revert the Hold creation fix and check whether the new test catches it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This round fixes all three open findings, but two problems remain, both Minor. BR-5 and BR-6 are addressed. BR-7's fix is in the code, but no test catches the bug it fixes: I reverted it in this checkout and every intent test still passed. So BR-7 stays not-addressed under the rule that a behaviour fix needs a test that fails without it. The new temp-file creation path also adds one small race of its own: one merge's sweep can delete another merge's temp file before that merge has locked it. Neither blocks the gate. The scope has not changed: the new `## Revisions` entry only records the names as built and the BR-7 fix.

**1. Strengths**
- `Intents.create` (`intent.go:254-277`) writes and locks a temp file and then `os.Link`s it into place. The record therefore appears at its real path already locked and complete. An `ErrExist` from the link sends the loop back round to the normal take-over path, which handles two merges of the same branch racing correctly.
- `Hold` now refuses an empty record instead of quietly treating it as new. This matches the earlier rule that a record Hold can't read is never overwritten (BR-3).
- BR-5 was fixed for the whole class. Both `formatMergeIntent` and `Hold`'s refusal message now print the path with `%q`, and the tests (`mergeintent_test.go:147`, `intent_test.go:291`) now check for the quoted form.
- `files()` already skips dot-files, so `.tmp-*` files never show up in `List` or `Sweep`. A reader that finds a record locked but unreadable reports it as live, not unreadable (`mergeintent.go:104-107`).

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **`intent.go:379-393`, `sweepTemps`:** it can delete a running merge's temp file. `os.CreateTemp` makes the file visible before `flock.Exclusive` runs in `create` (`intent.go:259-263`). Every merge sweeps before its own `Hold` (`mergeintent.go:42`). So merge A's sweep can lock and remove merge B's brand-new, still-unlocked temp file. B then fails with EWOULDBLOCK, or its `os.Link` fails with ENOENT, and B's merge refuses to start with an unclear error.
  - This is the 3rd finding in family `reader-lock-masquerades-as-owner`. The rule that covers it: an actor must never decide whether a file's owner is alive from the lock state while the file can still be in its "created but not yet locked" window.
  - Fix: either have the sweep leave alone any temp file younger than `IntentHoldPatience`, or have `create` retry on a lock or link failure.
- **BR-7: no regression test.** I put back `O_CREATE` and the treat-empty-as-new branch in `Hold`, and `go test ./internal/landingline -run Intent` passed. The new test only checks the state after `Hold` returns, which the old code also produced.
  - Fix: add a test seam in `create` that runs between writing and linking. At that point, assert that `List()` does not return the record at `path` at all.

**5. Test coverage notes**
- At the head commit, `go test ./internal/landingline -run Intent` and `go test . -run MergeIntent` both pass.
- The test for collecting dead temp files and keeping live ones is real: it checks both outcomes.
- Nothing covers the window where a sweep meets a temp file that exists but isn't locked yet.

**6. Architecture notes**
- **ARCH-DRY: pass.** `sweepTemps` and `removeUnchanged` use the same open, lock, `samePath`, remove steps. That is acceptable with two uses; a third should become a shared helper.
- **ARCH-PURE: pass.** The changes are all in the store's file-system layer. `reportMergeIntent` and `formatMergeIntent` stay pure.
- **ARCH-PURPOSE: pass for BR-5**, since every printed string from the shared directory is now quoted. **Flag for the lock family:** the race above is a sibling of the case BR-7 named.
- **ARCH-EVOLVE: pass.** The first record has no earlier fields to carry over, and the take-over path still merges `Extra`. The format did not change.

**7. Plan revision recommendations**
- The new entry dated 2026-10-23 matches the code: `runDurableMerge` wraps `holdMergeIntent` at `landing.go:346`. If the age-gate fix is adopted, add one line saying the sweep only removes temp files older than the patience window.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      formatMergeIntent and Hold's refusal now print the path with %q; tests assert the quoted form (mergeintent_test.go:147, intent_test.go:291).
  - id: BR-6
    disposition: addressed
    note: |
      The 2026-10-23 round-2 Revisions entry records Intent, Intents, List returning []ObservedIntent, reportMergeIntent/formatMergeIntent and runDurableMerge, all checked against landing.go:346 and mergeintent.go.
  - id: BR-7
    disposition: not-addressed
    note: |
      The temp-then-link fix is present, but reverting it (restoring O_CREATE and the treat-empty-as-new branch) leaves every intent test passing, so no regression test pins the window.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: reader-lock-masquerades-as-owner
    title: |
      sweepTemps can remove a running merge's temp file before create has locked it, so a concurrent merge fails to start
    detail: |
      3rd in this family. Rule: never judge a file's owner dead from its lock state while the file can still be created but not yet locked. os.CreateTemp (intent.go:259) makes .tmp-* visible before flock.Exclusive (intent.go:263); another merge's Sweep (mergeintent.go:42, intent.go:379-393) can lock and unlink it, and create then fails with EWOULDBLOCK or ENOENT on Link. Fix: have sweepTemps skip temp files younger than IntentHoldPatience, or retry create on a lock or link failure; add a test using a temp file that is fresh and unlocked.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ac8be880-49e3-47a0-9c5c-3f95c127e975","repository":"github.com/xianxu/ariadne","issue_id":"000565","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/4d0f435b9867b717","reviewed_head":"5f60907996bc7a32ea2453e72e625043a4780a04","slot":"ariadne:5","requested_at":"2026-10-23T18:58:59.470679Z"},"phase":"finished","pid":12288,"started_at":"2026-10-23T18:58:59.630195Z","finished_at":"2026-10-23T19:00:56.167535Z","session_id":"66aae585-db32-4bee-9f27-2973edd3c148","outcome":"returned"}] -->
