# Boundary Review — ariadne#275 (whole-issue close)

| field | value |
|-------|-------|
| issue | 275 — Keep issue detail frontmatter current |
| repo | ariadne |
| issue file | workshop/issues/000275-current-detail-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | c5e040ca68f5d104e535a587fa4ed314f6c56703..ea80ae010f675ded833f5c3dd87fff5f05ca8b47 |
| command | sdlc close --issue 275 |
| reviewer | claude |
| timestamp | 2026-09-30T16:49:57-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-09-30T17:06:11-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 275 — Keep issue detail frontmatter current |
| repo | ariadne |
| issue file | workshop/issues/000275-current-detail-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | c5e040ca68f5d104e535a587fa4ed314f6c56703..5e5d097218a09d8ee054c680d2a25da488401918 |
| command | sdlc close --issue 275 |
| reviewer | claude |
| timestamp | 2026-09-30T17:06:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

**Summary.** The branch does what #275 set out to do. Close adds a narrow mirror commit after the evidence commit using a compare-and-swap `update-ref`. All three checkout archive paths (merge, push, interrupted-archive recovery) refresh the moved history file before staging it. The slot-landing archive projects the done card through the pure `archivedDetails`. Its retry proof reads the card pinned by the archived file's own `card_mirror` and checks it: same id, `done`, same completion token. That keeps the proof independent of the live card, which is why #252 had removed the refresh in the first place. The pair#358 repro is covered end to end (`tracker_e2e_test.go:171`), along with the hand-edited, unreadable-baseline and card-change-after-archive cases.

Nothing blocks. One item should be fixed first: the plan's Core-concepts claims about the landing planner no longer match the code, because the round-1 fix moved the baseline lookup into the planner through an injected `readCard`. That needs a `## Revisions` entry, or a refactor back to pre-resolved bytes.

I inspected the whole range with the stat, name-status and targeted diffs. I did not run the test suite.

1. **Strengths**
   - `archivedDetails` (`cmd/sdlc/landingarchive.go`) is one total, pure projection. The proposal and the proof both use it, so they cannot drift apart (ARCH-DRY). Its table test covers determinism and every refusal case.
   - `pinArchivedCard` treats the archived `card_mirror` as untrusted input. Before substituting the pinned card it requires the same issue id, `status: done`, and the same completion token (ARCH-SECURE).
   - `commitCloseMirror` reuses the evidence commit's temp-index + `commitIndexOnto` + CAS shape. It only moves the index and worktree where they still hold the old blob, and an uncommitted edit keeps its body (`closetracker.go`, `closeMirrorCommit`).
   - Every refresh after publication degrades to a warning, never a failure. The new entry in `workshop/lessons.md` writes that rule down.
   - The atlas section "Mirror freshness" lists each refresh point and is honest about the copies that can still lag, including the crash window between codecomplete and the mirror commit.

2. **Critical:** none.

3. **Important**
   - **Plan's Core-concepts table contradicts the landing planner (ARCH-PURE).**
     - The plan (Core concepts → Pure entities) says `landingOwnedIssue` gains `baseline` and `card` *bytes*, resolved during selection "so the planner stays pure".
     - The code has `card []byte` plus `readCard func(oid string) ([]byte, error)`, which is bound to `env.repo.ReadCardBlob`. `planLandingArchive` → `trackedArchiveBytes` calls it, so the planner now does IO through a closure. There is no `baseline` field.
     - The cause is the round-1 fix: the baseline has to come from main's copy, which only the planner sees.
     - Fix either way: add a `## Revisions` entry recording the injected reader and why it is needed, or pre-resolve a map from card OID to blob in `selectTrackedLandingIssues` and `confirmLandingArchive` so the planner gets bytes again.

4. **Minor**
   - `cmd/sdlc/helptext/merge.md:73` has a reflow artifact: a dangling `An` alone on its line.
   - The branch where the details are staged differently from HEAD (`staged != oldBlob` in `closeMirrorCommit`) has no test.
   - After a crash between codecomplete and the mirror commit, nothing retries the mirror commit. This is documented in the atlas. An idempotent `commitCloseMirror` call in `reconcile` when the completion is already finished would close the gap cheaply.
   - Plan test names differ from the shipped ones, e.g. `TestRecoverInterruptedArchiveRefreshesMirror` became `TestRecoverInterruptedArchiveMirrorsTheDoneCard`. Cosmetic.

5. **Test coverage**
   - Real-git tests cover:
     - the close mirror, both clean and with a dirty worktree;
     - the push and merge checkout archives, recovery, and hand-edited details;
     - the pure projection table, plus proof survival after a card change;
     - an unreadable baseline, and the full slot-cycle e2e.
   - The log records mutation-checking `pinArchivedCard` (with it disabled, the proof goes red).
   - Close tests were correctly rebased onto `evidenceRev = "HEAD^"`.

6. **Architecture (each principle)**
   - **ARCH-DRY:** pass. There is one projection, `RefreshMirror` and `archivedDetails`. `commitIndexOnto` was extracted and is shared with `gitEvidence`.
   - **ARCH-PURE:** flag. The planner now calls the injected `readCard` (see Important).
   - **ARCH-PURPOSE:** pass. Close, all checkout archives and the landing archive are each wired in. Main's active copy is a documented, reasoned exception.
   - **ARCH-MOCK:** pass. Tests use real git repos and the existing tracker harness.
   - **ARCH-CONSTRAINTS:** pass. The tracker is opened once per archive run, with one blob read per issue.
   - **ARCH-SECURE:** pass. The pinned card is validated, and an unreadable or invalid mirror degrades to unchanged bytes that the proof expects.
   - **ARCH-ORDER:** pass. The mirror runs strictly after `Drive` confirms the completion, uses CAS on the branch ref, and is idempotent. Recovery only runs it on the close's own source branch.
   - **ARCH-FUNERAL:** pass. Each close generation adds one commit that lands with the branch. Nothing new is created.
   - Note for later work: the docs-only mirror commit after the evidence anchor depends on the #174 docs-only allowance in `validatePublishAnchors`. If that allowance is ever tightened, the mirror commit must stay exempt.

7. **Plan revisions recommended**
   - A `## Revisions` entry: "Round-1 review: the mirror baseline is resolved from main's copy inside `planLandingArchive` through an injected `landingOwnedIssue.readCard`, not pre-resolved `baseline` bytes. The planner is pure given that reader. Core-concepts row updated."

```findings
findings:
  - id: new
    severity: Important
    family: plan-table-drift
    title: |
      Plan's Core-concepts table says landingOwnedIssue carries baseline bytes and the planner stays pure; code injects a readCard IO closure into planLandingArchive
    detail: |
      trackedArchiveBytes (landingarchive.go) calls owned.readCard (env.repo.ReadCardBlob) from inside planLandingArchive; there is no baseline field. Add a Revisions entry or pre-resolve an OID->blob map in selection/confirmation (ARCH-PURE).
  - id: new
    severity: Minor
    family: helptext-reflow
    title: |
      merge.md help text leaves a dangling "An" line after the reflow
  - id: new
    severity: Minor
    family: untested-branch
    title: |
      closeMirrorCommit staged-differs-from-HEAD index branch has no test
  - id: new
    severity: Minor
    family: crash-window-retry
    title: |
      Mirror commit is not retried after a crash between codecomplete publish and the mirror commit
    detail: |
      Documented in the atlas; reconcile could call the idempotent commitCloseMirror when the completion is already finished on the source branch.
```

---

## Re-review — 2026-09-30T17:21:00-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 275 — Keep issue detail frontmatter current |
| repo | ariadne |
| issue file | workshop/issues/000275-current-detail-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | c5e040ca68f5d104e535a587fa4ed314f6c56703..56677eb11b34c4317c827f329f9643a71af4243a |
| command | sdlc close --issue 275 |
| reviewer | claude |
| timestamp | 2026-09-30T17:21:00-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All four findings from the earlier rounds are fixed, and I found nothing new worth raising. The checks I ran:
- **BR-1:** the plan's Core-concepts table and `## Revisions` now describe the injected `readCard` seam instead of pre-resolved `baseline` bytes. That matches `landingarchive.go:41-42`, where `readCard` is wired in from `selectTrackedLandingIssues` (`:657`).
- **BR-2:** the merge help text reads correctly again.
- **BR-3:** a new test covers the branch in `closeMirrorCommit` where the staged copy differs from HEAD.
- **BR-4:** the old per-receipt mirror call is gone. It is replaced by `retryCloseMirror`, which runs in reconcile's deferred block and has guards for each case it should skip.

The targeted tests (`Mirror|Archive|TrackerFullSlotCycle|Reconcile`) pass at HEAD (173s).

1. **Strengths**
   - `archivedDetails` (`landingarchive.go:66`) is the one pure projection. Both the planner and the proof derive from it, so a retry doesn't depend on the live card.
   - `pinArchivedCard` (`:85`) re-reads the card that the archived `card_mirror` names. It checks that card's completion token, status, and id, so a reopened or re-closed card is refused rather than trusted.
   - `retryCloseMirror` (`closetracker.go:121`) only acts when all of these hold:
     - the checkout is not on the rest branch;
     - the card is `codecomplete`;
     - the card was completed in this repository;
     - the evidence commit is an ancestor of HEAD.

     So it never commits onto another branch's close. The test checks that a second run commits nothing.
   - The mirror commit is built in a temporary index (`commitIndexOnto`, which is shared with `Prepare`, so the logic isn't duplicated per ARCH-DRY), so unrelated staged work is left alone.
   - `newArchiveMirrors` opens the tracker once per archive run and does nothing for files without a mirror.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** one, raised in the text only.
   - When a staged edit keeps its old index entry (`closeMirrorCommit`, the staged-differs branch), the index still holds the old mirror header while HEAD and the worktree have the new one. Committing that staged copy as-is would put the old header back.
   - This is harmless, because the mirror is a one-way projection that the next refresh rewrites. It's worth one line in the atlas's list of copies that can lag.

5. **Test coverage**
   - The regression in `TestReconcileRetriesAnInterruptedCloseMirror` can only go green through `retryCloseMirror`, because the old per-receipt call is deleted and the test resets to `evidenceRev` first. That means removing the fix would turn it red.
   - `TestTrackerFullSlotCycle` reproduces the pair#358 archive symptom.
   - The pin mutation check is recorded in the issue's Log (with `pinArchivedCard` disabled, the proof fails with "archive generation differs").

6. **Architecture**, one marker at a time:
   - **ARCH-DRY: pass.** `RefreshMirror`, `commitIndexOnto` and `refreshLocalMirrorAt` are reused rather than copied.
   - **ARCH-PURE: pass.** `archivedDetails` is the pure core. The only IO is the injected reader, and the plan now records that.
   - **ARCH-PURPOSE: pass.** Every lifecycle point the Spec lists is covered or documented: close, the checkout archives (merge, push, recovery), the slot-landing archive, reconcile, and main's active copy.
   - **ARCH-MOCK: pass.** The tests drive real git repositories through the existing seam.
   - **ARCH-CONSTRAINTS: pass.** The tracker is opened once per run, and each refresh touches one file.
   - **ARCH-SECURE: pass.** The `card_mirror` read from the archived file is untrusted input, and it is checked against the token and id. If the baseline can't be read, the details keep their bytes rather than having a header made up.
   - **ARCH-ORDER: pass.** Both the crash between codecomplete and the mirror commit and the interrupted archive are covered: the operations can be repeated safely, and recovery retries them.
   - **ARCH-FUNERAL: pass.** The change creates nothing durable except one narrow commit per close.

7. **Plan revisions:** none needed. The Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan table and Revisions now name the injected readCard seam; matches landingarchive.go:41-42 and :657.
  - id: BR-2
    disposition: addressed
    note: |
      merge.md reflowed; no dangling line.
  - id: BR-3
    disposition: addressed
    note: |
      TestReconcileRetriesAnInterruptedCloseMirror stages an edit and asserts the index keeps it (staged != oldBlob branch).
  - id: BR-4
    disposition: addressed
    note: |
      retryCloseMirror runs in reconcile's defer; the test resets to evidenceRev and reconcile restores the mirror commit; a second run is a no-op.
```
