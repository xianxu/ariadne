# Boundary Review — ariadne#424 (milestone M2)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf075d009ee932e321f06a20d9a7a3041c9f8d3b..4567a9e0931ed7c4bed5886d738bdfc73aab5cb8 |
| command | sdlc milestone-close --issue 424 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T04:07:12-07:00 |
| verdict | REWORK |

## Review

Checking whether the new whole-file read has the "mark all seen" side effect that `MarkAllSeen` documents.

Tracing couch's `read --all` into its handler to confirm whether it moves the diff baseline.

```verdict
verdict: REWORK
confidence: high
```

M2 mostly delivers what Task 6 promises, and the targeted packages pass (`go test ./cmd/tl-lead/ ./pkg/sidedoc/ ./pkg/reviewexport/` → ok). `syncIndex` reconciles through `sidedoc.ReconcileIndex`. `instructions link` writes the entry first and the link second, and the tests cover that split both ways. A new `status index` subcommand is registered, and both the conformance test and the fake gain the `--all` round trip. One defect blocks SHIP. The new `ReadAll`/`WriteAll` seam calls run Couch's `read --all` and `write --all`. In Couch, both of those move the agent's diff baseline to the whole file, which marks every operator edit as seen. So every `instructions link`, `status index`, and even a `submit` that changes nothing now silently empties the "changes since your last read" feed. The help text (`cmd/tl-lead/help.md:183-184`) and the `submit` → resolve → `seen` protocol both depend on that feed. The fake doesn't model this side effect, so no test can catch it.

Evidence (pair source): `/Users/xianxu/workspace/pair/cmd/internal/sessionsync/sync.go:243-255` says "Reading the whole file is the agent seeing every edit in it, so it refreshes the shadow and `diff` goes empty (pair#448)". `shadowAfter` (`sync.go:361-364`) returns the whole new document as the shadow for a whole-file write (`section == ""`). tl-lead's own seam comment already says this about the same verb: `MarkAllSeen` "reads the whole file (`read --all`), which tells Couch the agent has now seen every edit" (`cmd/tl-lead/sessionsync.go:22-24`). Yet `ReadAll` (`sessionsync.go:161`) runs that same verb with no mention of it.

### 1. Strengths
- `linkInstruction` (`cmd/tl-lead/instructions.go:166-200`) orders the two writes deliberately: entry first, link second. `TestIndexLinkSplit` (`cmd/tl-lead/index_test.go:79`) checks both outcomes, a rerun that completes and a reconcile that drops the entry. This is ARCH-ORDER done properly.
- `TestIndexOneWaySequence` tests the one-way rule as a sequence of events (link, the operator removing an entry, a submit, a bullet deletion, a later link). It checks the invariants directly rather than restating the implementation.
- The fake's whole-file view is built from the same per-section state (`sessionsync_fake_test.go:117-171`). So a section write racing a whole-file write goes stale just as it does in Couch, and the race test pins the retry count (`allWriteCalls != 3`).
- tl-lead reuses `pkg/sidedoc.ReconcileIndex` instead of re-parsing the index (ARCH-DRY). Paths added with `--add` are checked by `checkEntry` before any write (ARCH-SECURE).
- The conformance test now runs `write --all` against a live Couch: a stale refusal, a successful CAS write, sections left intact, and the lock refusal. This is the evidence the plan promised.

### 2. Critical findings
- **`cmd/tl-lead/index.go:25` and `:34`, `cmd/tl-lead/sessionsync.go:161-179`: the index's `read --all` and `write --all` mark every operator edit as seen.**
  - **What breaks:**
    - The TL runs `instructions link` while the operator has unread edits in Needs you or Threads. `syncIndex`'s `ReadAll` resets Couch's diff baseline, and the next `submit` prints an empty diff. The operator's edits are now invisible to the TL.
    - The same happens on every `status index`, even when nothing changes, because the read alone resets the baseline.
    - It also happens in `submit` itself, for edits that land between its `Diff` call and its `syncIndex`. After any submit, the separate `status seen` step has nothing left to do.
  - **Contract it breaks:** `help.md:183-184` says "follow-ups on linked ones are in the changes since your last read".
  - **Fix sketch:** pick one:
    - (a) Add a Couch verb that writes only the frontmatter and moves the baseline only there, the way `shadowAfter` already handles a section write. Plus a way to read the frontmatter without resetting the baseline. This is a pair change, so the plan's "Couch needs nothing new" no longer holds.
    - (b) Minimum fix inside tl-lead: before `ReadAll`, `syncIndex` calls `d.ss.Diff(file)`. If the diff isn't empty, it prints it ("changes since your last read, now marked seen:") so the edits are shown rather than lost. Also make a no-op `status index` skip the baseline reset.
  - **Tests either way:** see the fake finding below.

### 3. Important findings
- **`cmd/tl-lead/sessionsync_fake_test.go:97` and `:131-134`: the fake doesn't model the seen-baseline side effect of `read --all` and `write --all` (ARCH-MOCK).**
  - `Diff` returns a fixed string, and `ReadAll`/`WriteAll` never touch it. The fake therefore cannot express the bug above.
  - **Fix:** give the fake a shadow. `ReadAll` and `WriteAll` set it to the whole text; `WriteSection` updates only its own section; `Diff` compares the file against the shadow. Add a sequence test: the operator edits Needs you, the TL links a bullet, then `submit` still shows the edit (or `link` prints it, under fix (b)).
  - **Fix the conformance test too:** assert that `diff` is non-empty after an outside edit and check what it shows after `read --all`, so this behavior is checked against the real Couch.

### 4. Minor findings
- `instructions link` now refuses when the index can't be parsed, or when TARGET fails `plainPathRE` (for example a `+` or `%` in the name). This happens before the link write, so a broken hand-edited index blocks every answer link. The refusal is reasonable, but neither the help text nor the `link` error message points to `status index` as the way to see what's wrong.
- The conformance test checks that section A's body survives `write --all`, but not that the section's earlier hash still works afterwards. `link` depends on exactly that: `list`'s hash, then `write --all`, then the section write with that hash. Today it holds because Couch hashes only the section body; one more assertion would pin it.
- `syncIndex` repeats `writeSection`'s read-recompute-write retry loop (`status.go:34`) with a different read/write pair. Acceptable at two call sites; extract a shared helper if a third appears (ARCH-DRY).
- The `status` command's Long text still says "The one write that never retries is instructions link". Its index half now retries.

### 5. Test coverage notes
- The one-way sequence, the split, the lock, the race and the unparseable-index refusal are all covered against the fake. Couch's `--all` CAS is covered by the opt-in conformance test. I did not run it live (`TL_LEAD_LIVE=1`): it would create Couch state on this host, and the pair source already settles the question above.
- The gap is any interplay between these writes and the diff/seen state. That is exactly the bug class this diff ships.

### 6. Architecture principles
- **ARCH-DRY: pass.** Uses the shared `sidedoc` package; the retry-loop repeat is minor.
- **ARCH-PURE: pass.** `ReconcileIndex` is pure; `reportIndex`'s `os.Stat` is thin glue.
- **ARCH-PURPOSE: pass,** apart from the Critical finding. The Done-when items for M2 (tests against the fake, the conformance check) are present.
- **ARCH-MOCK: flag.** The fake leaves out a side effect of the dependency that production relies on (Important finding above).
- **ARCH-CONSTRAINTS: pass.** Two extra couch process launches per `link` or `submit`, which is within the seconds budget.
- **ARCH-SECURE: pass.** Entries and `--add` paths are validated as relative paths that don't climb out; argv arrays, no shell.
- **ARCH-ORDER: flag,** together with the Critical finding. The two-write order is explicit and tested, but the plan's ordering section leaves out the shared diff baseline these writes reset.
- **ARCH-FUNERAL: pass.** No new durable files. Index entries are bounded by the number of linked side docs and are removed by link deletion or the operator.

### 7. Notes for M3
- M3 will call git through the archive paths. Before wiring them, check what each seam call does to state beyond its own purpose, the same way `read --all` turned out to mark edits seen here.

### 8. Plan revision recommendations
- Add to `## Revisions`: "M2 review: Couch's `read --all` and `write --all` reset the agent's diff baseline (pair#448). The index write therefore marks every operator edit as seen. Either Couch gains a frontmatter-scoped read/write (a pair change; the claim 'Couch needs nothing new' is withdrawn), or `syncIndex` prints the pending diff before its read. The fake models the shadow."
- Integration-points table: say on `SessionSync.ReadAll`/`WriteAll` that both move the diff baseline.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: seam-side-effect-unmodeled
    title: |
      The index's read --all and write --all mark every operator edit as seen, so link, index and submit empty the "changes since your last read" feed
    detail: |
      cmd/tl-lead/index.go:25 and :34 call ReadAll and WriteAll, which run couch's read --all and write --all. In pair, sessionsync/sync.go:243-255 says reading the whole file refreshes the shadow so diff goes empty, and shadowAfter (sync.go:361) makes the whole new document the shadow on a whole-file write. tl-lead's own MarkAllSeen comment (sessionsync.go:22) documents the same verb. Result: an instructions link, a no-op status index, or a submit (for edits landing after its Diff) silently consumes unread operator edits, which help.md:183-184 says arrive as changes since the last read. Fix: either a couch frontmatter-scoped read/write that moves the baseline only there (a pair change), or syncIndex prints d.ss.Diff before ReadAll and a no-op reconcile avoids the baseline reset. Add a regression test against a fake that models the shadow.
  - id: new
    severity: Important
    kind: defect
    family: seam-side-effect-unmodeled
    title: |
      fakeSessionSync does not model couch's diff baseline, so no test can see the seen side effect of ReadAll and WriteAll (ARCH-MOCK)
    detail: |
      sessionsync_fake_test.go:97 returns a fixed diff string, and ReadAll/WriteAll (lines 131-171) never touch it. Give the fake a shadow: ReadAll and WriteAll set it to the whole text, WriteSection updates only its section, Diff compares the file with the shadow. Add a sequence test (operator edits Needs you, TL links, the edit is still reported), and have TestSessionSyncConformance check diff after read --all against the live couch.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"109a1e37-ebda-422d-8589-66f9d474da4e","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2ecfa701c2823147","reviewed_head":"4567a9e0931ed7c4bed5886d738bdfc73aab5cb8","slot":"ariadne:2","requested_at":"2026-10-17T11:04:05.49281Z"},"phase":"finished","pid":28404,"started_at":"2026-10-17T11:04:05.723048Z","finished_at":"2026-10-17T11:07:11.769654Z","session_id":"704720e8-70fb-40a0-8a02-3c6d66d19937","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T04:12:53-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf075d009ee932e321f06a20d9a7a3041c9f8d3b..efbd03435ea08cc407535ea55c8ddec94fa031b7 |
| command | sdlc milestone-close --issue 424 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T04:12:53-07:00 |
| verdict | REWORK |

## Review

The fix holds for the no-op case. But a probe shows the stale-retry path still consumes an operator edit without printing it. Next I'll revert the fix and confirm the new regression test goes red.

Reverting the fix turns the regression test red, so BR-9's test is real. Last step: checking the plan's M2 checklist against the code.

```verdict
verdict: REWORK
confidence: high
```

Most of M2 works. Every Plan Task 6 step is in the code: the `status index` subcommand, `link` writing the index entry before the link, `submit` reconciling the index, the `ReadAll`/`WriteAll` seam, the fake's whole-file view, the conformance round trip, and the help.md and atlas updates. `go test ./cmd/tl-lead/ ./pkg/sidedoc/ ./pkg/reviewexport/` passes. BR-9 is addressed: the fake now models couch's seen-baseline. When I reverted the fix in the disposable checkout, `TestIndexKeepsUnreadEdits` went red with "a no-op reconcile consumed the operator's edit". **BR-8 is only partly fixed, and that is what blocks SHIP.** The no-op path and the first write attempt are now safe. The stale-retry path in `syncIndex` still marks an operator edit seen without ever printing it. That path runs exactly when the operator (or anyone) edits during the write, which is the case BR-8 is about. I confirmed it with a probe test against the fake: call `f.seen()`, have an operator edit land before the first `WriteAll`, then run `status index --add …`. The output was only `index: added …`, and afterwards `Diff` returned `""`. The edit was consumed and never shown. The plan also wasn't revised for the new `Peek` seam or for couch's baseline being state between events.

**1. Strengths**
- `cmd/tl-lead/index.go:28-33`: if `Peek` and a pure `ReconcileIndex` show nothing to change, the command returns before touching couch's baseline. That makes the common case (a submit with no index change) safe.
- `sessionsync_fake_test.go:509-566`: the fake's shadow is a real stateful model. Whole-file reads and writes reset all of it, a section write updates only its own section, and `Diff` compares the current state against it. That is the right fake for ARCH-MOCK.
- `conformance_test.go:23-63` checks against the live couch that `read --all` empties the diff and that `Peek` returns the same text as `read --all`, plus the stale and lock refusals of `write --all`. The fake's assumptions are now checked against the real dependency.
- `instructions.go:166-190`: writing the entry first and the link second matches the plan's ARCH-ORDER section, and `TestIndexLinkSplit` tests both outcomes (rerun and reconcile).
- `examples_test.go:296-313`: the numbering now counts malformed side docs, with a regression case for it.

**2. Critical**
- BR-8 is not addressed (see the dispositions below). In `cmd/tl-lead/index.go:37-67`, `Diff` is printed once, before the loop. On `ErrStale`, the loop `continue`s into another `ReadAll`, which marks seen the very edit that caused the stale refusal. Nothing prints that edit.
  - Fix: move the `Diff` + `unreadChanges` print inside the loop, so it runs before every `ReadAll`.
  - Document the remaining window between `Diff` and `ReadAll`, which only a couch-side frontmatter-scoped write can close.
  - Add a regression test: `f.seen()`, then `editAllBefore{1: …}`, then assert the output contains `## Needs you changed`. The existing race test in `TestIndexCommand` never calls `seen()`, so it can't see this.

**3. Important**
- The plan doesn't describe the code (2nd finding in family `plan-matches-code`):
  - The Integration points table has no `SessionSync.Peek` row. That method is a new seam method that reads the file with `os.ReadFile` outside couch.
  - The ARCH-ORDER section says the index "holds no state between events beyond the file itself". Couch's seen-baseline is exactly that kind of state, and the BR-8 fix depends on it.
  - The rule that covers this class: any seam method or state assumption added while fixing a review finding needs a `## Revisions` entry and a table row in the same fix commit. The BR-8/BR-9 fix commit added neither.

**4. Minor**
- `submit` prints `Diff` at `status.go:185-189`, and then `syncIndex` prints it again whenever the index needs a write. The operator's edits show up twice in one submit's output.
- The fake's `ReadSection` doesn't update the shadow. If couch's section read moves that section's baseline, the fake under-models it. One more step in the conformance test would settle it.
- ARCH-DRY: the retry loop in `syncIndex` repeats the stale → re-read → retry pattern from `writeSection`. A shared `retryStale(func() (bool, error))` helper would cover both.
- If `Peek` fails, the code silently falls through to the write path. That's harmless, but the error is dropped.

**5. Test coverage**
- The one-way sequence, the link split, the lock, the section-write race, an unparseable index and the no-op baseline are all tested against the fake.
- Missing: the sequence "operator edit, stale retry, output". The race test runs without a baseline (`shadow == nil`), so it can't observe whether an edit was consumed.
- Conformance only runs with `TL_LEAD_LIVE=1`. I didn't run it here.

**6. Architecture**

| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass, with a minor note | the retry loop above |
| ARCH-PURE | pass | `ReconcileIndex` is pure; `syncIndex` is the thin IO layer around it |
| ARCH-PURPOSE | flag | the purpose of BR-8 ("the index never consumes unread edits") holds on two paths out of three; the retry path is the one left |
| ARCH-MOCK | pass | the fake has a shadow and conformance pins `read --all` and `Peek` |
| ARCH-CONSTRAINTS | pass | tens of entries, at most 4 tries |
| ARCH-SECURE | pass | the index parse refuses rather than guessing; a missing file is named, not dropped |
| ARCH-ORDER | flag | the retry ordering isn't tested with a baseline, and the plan's claim that the index holds no state is wrong |
| ARCH-FUNERAL | pass | nothing new is left behind; body temp files go through the existing `writeBody` |

A couch frontmatter-scoped read and write (proposed to pair) would make the `Peek`/print workaround unnecessary and close the remaining window.

**7. Plan revisions**
- Add a `## Revisions` entry, "M2 as built (M2 review)", that:
  - adds `SessionSync.Peek` (`cmd/tl-lead/sessionsync.go`, reads the file on disk and leaves couch's baseline alone) to the Integration points table;
  - corrects the tl-lead line in ARCH-ORDER: couch's seen-baseline is state carried between events; whole-file read and write reset it; the index never resets it without first printing the pending diff, on every attempt including stale retries; the window between `Diff` and `ReadAll` remains until a couch frontmatter-scoped write exists.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      The no-op path is fixed (Peek) and the first attempt prints Diff, but on ErrStale syncIndex (index.go:47-67) loops into ReadAll again without re-printing Diff, so the edit that caused the stale refusal is marked seen and never shown. A probe against the fake (f.seen(), editAllBefore{1: Needs you edit}, status index --add) printed only "index: added" and Diff was "" afterwards. Fix: Diff and print before every ReadAll in the loop, plus a regression test with a baseline.
  - id: BR-9
    disposition: addressed
    note: |
      The fake has a shadow (sessionsync_fake_test.go:517-566), the sequence test TestIndexKeepsUnreadEdits exists and went red when I reverted the Peek and print fix, and conformance_test.go:32-46 pins diff-after-read-all and Peek equal to read --all against the live couch.
findings:
  - id: new
    severity: Important
    kind: defect
    family: plan-matches-code
    title: |
      The plan omits the new Peek seam method and still says the index holds no state between events, though couch's seen-baseline is state the fix depends on
    detail: |
      This is the 2nd finding in family plan-matches-code. The rule for the class: any seam method or state assumption introduced while fixing a review finding gets a Revisions entry and an Integration points row in the same fix commit. Here, add SessionSync.Peek (cmd/tl-lead/sessionsync.go:466, os.ReadFile outside couch) to the table, and revise the ARCH-ORDER tl-lead line to name couch's seen-baseline, the print-before-every-ReadAll rule, and the remaining window between Diff and ReadAll.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2ca6af37-7856-49dd-a7fd-be8fd016e958","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/3a4c3d609c6ec262","reviewed_head":"efbd03435ea08cc407535ea55c8ddec94fa031b7","slot":"ariadne:2","requested_at":"2026-10-17T11:10:24.577893Z"},"phase":"finished","pid":52469,"started_at":"2026-10-17T11:10:24.699057Z","finished_at":"2026-10-17T11:12:23.664892Z","session_id":"58a54723-7a0f-4212-84d5-39e617c121e0","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T04:16:02-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf075d009ee932e321f06a20d9a7a3041c9f8d3b..1b2e41e1e7ecffbf7a7759fd2e4287e917ff8087 |
| command | sdlc milestone-close --issue 424 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T04:16:02-07:00 |
| verdict | SHIP |

## Review

Code diff read; now running the tl-lead tests and reverting the BR-8 fix to confirm the regression tests go red.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and I confirmed it. For BR-8, I made two separate reverts in the disposable checkout. Reverting the no-op reconcile from `Peek` back to `ReadAll` turns `TestIndexKeepsUnreadEdits` red with "a no-op reconcile consumed the operator's edit". Turning off the unread-diff print and the window print turns `TestIndexShowsEditsOnEveryRead` red. So the fixes can be reached, and the tests guard them. For BR-10, the plan now has a `SessionSync.Peek` row under Integration points. Its ARCH-ORDER tl-lead line names couch's seen-baseline, the rule that the diff is printed before every whole-file read, and the race that remains inside couch (tracked as pair#643). There is also a Revisions entry stating the class rule. `go test ./cmd/tl-lead/ ./pkg/reviewexport/ ./pkg/sidedoc/` passes at the pinned head. Nothing blocks SHIP. The two new findings are Minor.

**1. Strengths**
- `syncIndex` (`cmd/tl-lead/index.go:28-57`) decides with `Peek` and never touches couch when there is nothing to change. Every whole-file read goes through one function, `readAllShowingUnread`, including on each stale retry. That makes "nothing marked seen without being shown" hold structurally, not just at each call site.
- The fake now models couch's seen-baseline (`shadow` in `sessionsync_fake_test.go`). A whole-file read or write resets all of it, and a section write resets only its own section. The ordering hooks `editAllBefore` and `editOnReadAll` let tests replay specific interleavings, which ARCH-ORDER asks for.
- `conformance_test.go:75-125` checks against the live couch that `read --all` empties the diff, that `Peek` matches `read --all` byte for byte, and how `write --all` behaves on a stale hash and under the lock. The fake's model is therefore checked against the real binary.
- `instructions link` writes the index entry first and the link second (`instructions.go:166-190`). `TestIndexLinkSplit` tests both branches: rerunning `link` completes the pair, and if nobody reruns it, the next reconcile drops the entry.
- The checker fix in `pkg/reviewexport/examples_test.go:299-316` counts side-doc numbers across every file, including malformed side docs, and a new mutation case catches the regression.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- ARCH-DRY: `lineChanges` (`cmd/tl-lead/index.go:92`) re-implements the existing `lineDiff` (`cmd/tl-lead/linediff.go:8`) as a multiset diff. It could be `lineDiff` with the shared `" "` lines filtered out.
- When `submit` needs an index write, it prints the unread diff twice: once at `status.go:185-189`, and again under the "(the index write marks them seen)" label from `readAllShowingUnread`. The output is redundant but nothing is lost.

**5. Test coverage notes**
- The BR-8 regressions are confirmed red when the fix is reverted (above).
- The fake's `Diff` still returns the legacy `f.diff` string when it is set, so the older tests bypass the shadow model. That's acceptable, but new tests should leave `f.diff` empty.
- The live conformance test doesn't check that the diff is empty after `write --all`. Only the fake models that. It would be cheap to add next to the existing `read --all` check.

**6. Architectural notes**
- ARCH-DRY: flagged (Minor, above).
- ARCH-PURE: pass. `ReconcileIndex` is pure in `pkg/sidedoc`, and `syncIndex` is thin glue over the seam.
- ARCH-PURPOSE: pass. `link`, `submit` and `index` all reconcile through one `syncIndex`, and the checker shares the rule.
- ARCH-MOCK: pass. The fake is stateful, including the baseline, behind the same seam, and the live conformance test covers `ReadAll`, `WriteAll`, `Peek` and the diff baseline.
- ARCH-CONSTRAINTS: pass. The inputs are status files of about a screen.
- ARCH-SECURE: pass. A malformed index refuses with the document unchanged, and entry paths are only `Stat`ed.
- ARCH-ORDER: pass. The order of the two writes is designed in; the race inside couch is named and assigned to pair#643; the interleavings can be replayed through the fake's hooks.
- ARCH-FUNERAL: pass. Index entries are removed by the one-way rule (when a link is deleted or the operator removes the entry), and nothing else durable is created.
- Upcoming: when pair#643 lands, switch to the frontmatter-scoped read and write, then remove `readAllShowingUnread` and `lineChanges` and narrow the ARCH-ORDER text.

**7. Plan revision recommendations**
- Optional: change M2 Step 2's text from "`ReadAll` → `ReconcileIndex` → `WriteAll`" to "`Peek` decides → `readAllShowingUnread` → …". The Revisions entry already covers this.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      Peek no-op path plus readAllShowingUnread on every read and retry; each half verified red when reverted (TestIndexKeepsUnreadEdits, TestIndexShowsEditsOnEveryRead); live conformance pins the read --all baseline reset.
  - id: BR-10
    disposition: addressed
    note: |
      The plan adds a SessionSync.Peek Integration points row, revises the ARCH-ORDER tl-lead line (seen-baseline, print-before-read rule, remaining couch-internal window, pair 643) and adds a Revisions entry stating the class rule.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: reuse-existing-helper
    title: |
      lineChanges in cmd/tl-lead/index.go re-implements the package's existing lineDiff (ARCH-DRY)
    detail: |
      cmd/tl-lead/linediff.go:8 already renders the -/+ line diff of two status-file texts; lineChanges (index.go:92) is a second multiset variant. Use lineDiff and drop its shared-line output instead.
  - id: new
    severity: Minor
    kind: defect
    family: duplicate-output
    title: |
      submit prints the unread diff twice when the index reconcile needs a write
    detail: |
      status.go:185-189 prints Diff, then syncIndex's readAllShowingUnread prints the same unread changes again under its own header. The output is redundant but nothing is lost; the second print could be skipped when submit has just shown the same diff.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1d42e8d9-30d0-46fe-a741-9033094c2540","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8ec7d4ba145a72ff","reviewed_head":"1b2e41e1e7ecffbf7a7759fd2e4287e917ff8087","slot":"ariadne:2","requested_at":"2026-10-17T11:14:40.607693Z"},"phase":"finished","pid":50037,"started_at":"2026-10-17T11:14:40.97461Z","finished_at":"2026-10-17T11:16:01.175086Z","session_id":"d6d30cbf-790d-408b-9203-0549fb3a15c7","outcome":"returned"}] -->
