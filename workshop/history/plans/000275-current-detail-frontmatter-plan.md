# Keep Issue Detail Frontmatter Current — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refresh an issue's mirrored detail frontmatter from its card at close and at every archive path. The final archived copy then says `done`, with the finalized dates and hours. The tracker stays the sole authority.

**Architecture:** Every refresh is the existing pure projection `issue.RefreshMirror(details, baseline, card)`. It is one-way and preserves the body, and it refuses when a mirrored field was edited by hand. This issue adds refresh *points*; it adds no new projection:
1. **Close.** After codecomplete publishes, a narrow follow-up commit on the issue branch refreshes the details. It cannot go into the evidence commit, because the codecomplete card embeds that commit's SHA, which would be circular.
2. **Checkout archives** (merge, push, interrupted-archive recovery). Each refreshes the moved history file before it is staged.
3. **Slot-landing archive.** It projects the done card deterministically. Its retry proof recovers the pinned card from the archived file's own `card_mirror`, so the proof stays independent of the live card, which is the property #252 dropped the refresh to protect.

**Tech Stack:** Go (`cmd/sdlc`), real-git tests using the existing `closeReady` / `closedAndLanded` / `laArchive` / `laProof` harness.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `issue.RefreshMirror` | `cmd/sdlc/internal/issue/mirror.go` | unchanged (reused) |
| `archivedDetails` | `cmd/sdlc/landingarchive.go` | new |
| `landingOwnedIssue` (+ `card` bytes, injected `readCard`) | `cmd/sdlc/landingarchive.go` | modified |
| `trackedArchiveBytes` / `pinArchivedCard` | `cmd/sdlc/landingarchive.go` | new |
| `planLandingArchive` | `cmd/sdlc/landingarchive.go` | modified |

- **archivedDetails(content, baseline, card []byte) []byte** — the one archive projection for tracked details. It returns `RefreshMirror(content, baseline, card)`, or `content` unchanged when the refresh refuses. A refusal means a hand-edited mirrored field, or no or an invalid mirror. The outcome is a function of its inputs only, so the proof can recompute it. It is pure, and `landingarchive_test.go` tests it without IO.
  - **Relationships:** 1:1 per archived tracked issue. `planLandingArchive` calls it for proposal and proof alike.
  - **DRY rationale:** Proposal and proof share one function, so they cannot drift (ARCH-DRY).
- **landingOwnedIssue** gains `baseline` (the card blob named by the PR-head details' `card_mirror`) and `card` (the done card). Both are resolved during selection, which is the IO side, so the planner stays pure.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `commitCloseMirror` | `cmd/sdlc/closetracker.go` | new | git temp index + `update-ref` CAS, tracker snapshot |
| `refreshArchivedMirror` | `cmd/sdlc/claim.go` (beside `refreshLocalMirrorAt`) | new | tracker env open + `refreshLocalMirrorAt` |
| `selectTrackedLandingIssues` | `cmd/sdlc/landingarchive.go` | modified | tracker records + `ReadCardBlob` |
| `confirmLandingArchive` | `cmd/sdlc/landingarchive.go` | modified | `ReadCardBlob` of the archived `card_mirror` |

- **commitCloseMirror(env, stderr, id, detailRel)** — after codecomplete is confirmed, it reads `HEAD:<details>` and projects the current card with `refreshMirror`. It commits the result as `#N: mirror codecomplete card`, using the same temp-index plus compare-and-swap `update-ref` shape as `gitEvidence`, so staged unrelated work is untouched.
  - If the worktree file equals the old HEAD blob, it writes the new bytes to the worktree.
  - If the worktree is dirty, it runs `refreshLocalMirrorAt` on the worktree bytes, which preserves the operator's body edits. If that refuses, it leaves the file and warns.
  - Every failure is a **warning**: the close already published, and a stale mirror is cosmetic. Callers are `publishTrackerClose` after a successful `Drive`, and the `completion` case of `issue recovery reconcile` (FIX-THEN-SHIP).
  - On the rest branch it does nothing (a close always runs on a branch).
- **refreshArchivedMirror(ctx, stderr, root, historyAbs)** — used by the checkout archive paths A, B and C. It opens the tracker at `root`, returns quietly in a legacy repository, and otherwise calls `refreshLocalMirrorAt` and warns on refusal. The file is archived either way.

### Lifecycle notes (ARCH-FUNERAL / ARCH-STATE)

- The close adds one commit per close generation on the issue branch, and that commit lands with the branch. It creates nothing else durable.
- Archive refresh changes the bytes of an existing file only. The landing proof pins the card through the archived file's `card_mirror`; that card blob stays reachable on the tracker branch's history, which is never rewritten.
- **Ordering.**
  - The close refresh runs strictly after `Drive` reports `completion` finished. If it runs twice, the second pass is a no-op because `RefreshMirror` is idempotent: the same card yields the same bytes and no commit, since an unchanged tree is skipped.
  - The archive refresh runs after `completeOnCard` or `settleLandedCompletions` has made the card done. All three flows already order it that way.
  - A reopen *after* the archive does not break the proof, because the proof reads the pinned card, not the live one.

### Remaining stale cases (to document)

- Main's *active* copy (from `move-detail`) stays at its creation-time mirror until the branch lands, because the rest branch is never edited. A main-side refresh commit would conflict with the branch's own frontmatter refreshes.
- A checkout that has not run an SDLC verb since a card change (`git pull` is not a refresh).
- The issue branch between a setter on another clone and the next local verb.
- A hand-edited mirrored field: the details are archived as-is, with a warning.

## Task 1: Close refresh commit

**Files:** Modify `cmd/sdlc/closetracker.go`, `cmd/sdlc/issuerecovery.go`. Test `cmd/sdlc/closetracker_test.go`.

- [ ] Write the failing test `TestCloseRefreshesDetailsToCodecompleteCard`. Use `closeReady(t, 341)`, a SHIP stub and `sdlc close`. Assert:
  - HEAD's details contain `status: codecomplete` and `card_mirror: '<git hash-object of card>'`;
  - HEAD^ is the evidence commit named by the card binding;
  - `git status --porcelain` is empty;
  - the body is byte-equal to the evidence commit's details apart from frontmatter.
- [ ] Write the failing test `TestCloseMirrorKeepsDirtyDetailsBody`. Append a Log line to the worktree details after the evidence is pinned, by injecting through a FIX-THEN-SHIP flow: close FIX-THEN-SHIP, commit a fix, add an uncommitted Log line, then `issue recovery reconcile`. Assert that the committed HEAD details are refreshed without the Log line, and that the worktree keeps the Log line *and* the refreshed mirror.
- [ ] Run `go test ./cmd/sdlc -run 'TestCloseRefreshes|TestCloseMirror'`. Expect FAIL (status still `working`).
- [ ] Implement `commitCloseMirror` in closetracker.go, then call it after the `cok(... codecomplete ...)` line in `publishTrackerClose` and after `Drive` succeeds for `completion` in issuerecovery.go.
- [ ] Run the two tests, then `go test ./cmd/sdlc -run 'Close|Recovery|Publish|Landing|Merge'`. Expect PASS. Fix any existing test that asserted the HEAD after close *is* the evidence commit.
- [ ] Commit: `#275: close: mirror the codecomplete card onto the issue branch`.

## Task 2: Checkout archive refresh (merge / push / recovery)

**Files:** Modify `cmd/sdlc/claim.go` (helper), `merge.go:~691`, `push.go:~660`, `push.go:~415` (`recoverInterruptedArchive`, before `git add`, for each issue move's history path). Test `cmd/sdlc/trackercompletion_test.go`.

- [ ] Extend `TestPublishFlipCompletesLandedClosesAndArchivesByCard` to reproduce the pair#358 archive half. The archived file must contain `status: done`, the card's `updated`, `actual_hours: 1`, and `card_mirror` equal to the done card's blob OID. Its body must equal the landed details' body.
- [ ] Add `TestArchiveKeepsHandEditedMirrorWithWarning`. Hand-edit `status:` in the landed details, archive, and assert that the file is archived byte-identical and stderr names `mirror not refreshed`.
- [ ] Add `TestRecoverInterruptedArchiveRefreshesMirror`. Rename the details to history by hand, leaving the card done, run `recoverInterruptedArchive`, and assert the committed history file says `done`.
- [ ] Run the tests and verify they FAIL.
- [ ] Implement `refreshArchivedMirror` and call it right after each `os.Rename` (paths A and B) and before `git add` (path C).
- [ ] Run the tests plus `go test ./cmd/sdlc -run 'Archive|Push|Merge'`. Expect PASS.
- [ ] Commit: `#275: archive: project the done card into archived details`.

## Task 3: Deterministic landing archive

**Files:** Modify `cmd/sdlc/landingarchive.go`. Test `cmd/sdlc/trackercompletion_test.go`, `cmd/sdlc/landingarchive_test.go`, `cmd/sdlc/tracker_e2e_test.go`.

- [ ] Add a pure test `TestArchivedDetailsProjectsOrKeeps` (table). It covers:
  - a clean mirror plus a done card, giving refreshed bytes;
  - the same inputs again, giving identical bytes (determinism);
  - a hand-edited status, giving the input unchanged;
  - no `card_mirror`, giving the input unchanged.
- [ ] Change the assertion in `TestDurableLandingArchivesTrackedCloseByBinding` (:121). The archived file must now equal `RefreshMirror(head details, baseline, done card)` and say `status: done`. Keep the proof and no-op-retry assertions.
- [ ] Add `TestDurableLandingProofSurvivesCardChangeAfterArchive`. After the archive, run `sdlc issue set-title --issue N 'renamed'`. The card stays done under the same close token, so the issue is still selected, but its live blob no longer matches the archived bytes. `laProof` must still report complete with no error, because the proof reads the pinned card. A reopen would not test this, since a reopened card is no longer selected at all. If `set-title` refuses on a done card, rewrite the card with the tracker's `ChangeCard` directly.
- [ ] Extend `TestTrackerFullSlotCycle` (tracker_e2e_test.go:~160) to assert that the archived details on origin main say `status: done`. This is the full pair#358 repro.
- [ ] Run and verify FAIL.
- [ ] Implement:
  - `selectTrackedLandingIssues`: set `ref.card = oc.Card.Raw`, and set `ref.baseline` from `ReadCardBlob(MirrorBaselineOID(content))`. Leave `baseline` nil when there is no or an invalid mirror, so `archivedDetails` keeps the bytes.
  - `planLandingArchive`: `done := archivedDetails(b.content, owned.baseline, owned.card)` when `owned.tracked`, and only when the card is `done`. A codecomplete card keeps the bytes; that cannot happen after `completeLandingPR` succeeds, but the planner must stay total.
  - `confirmLandingArchive`: before re-planning, for each tracked selected issue, read the archived destination's `card_mirror`. If it differs from the source's own mirror, `ReadCardBlob` it, require the same issue ID, `status: done` and the same completion token as `owned.card`, and substitute it as `owned.card`. Otherwise keep `owned.card` as the source's baseline, so the projection is the identity. This works on a copy of `selected`.
- [ ] Run the tests, then `make test` from the repo root. Expect all green.
- [ ] Commit: `#275: landing: project the done card deterministically, proof pins it`.

## Task 4: Documentation

**Files:** `atlas/workflow/issue-tracker.md`, `cmd/sdlc` help for `close` (if it describes the commits it creates), and `workshop/issues/000275-…md` Log.

- [ ] In issue-tracker.md, replace "Tracked details are archived byte-for-byte…" with a **Mirror freshness** subsection containing:
  - a refresh-point table: claim, start-plan, change-code, setters, close follow-up commit, checkout archives, landing archive with its pinned proof;
  - the "Remaining stale cases" list above.
  Add a `close` row note to the verb table.
- [ ] Grep for stale "byte-for-byte" / "as they are" claims: `grep -rn "byte-for-byte\|as they are" atlas cmd/sdlc --include='*.go' --include='*.md'`. Update each one, including the comments at landingarchive.go:35 and :325, publishgate.go:333 and claim.go:170.
- [ ] Commit: `#275: docs: mirror freshness points and remaining stale copies`.

## Revisions

- 2026-09-30 — close review (BR-1) and round-1 fixes.
  - **Reason:** the first design resolved the mirror baseline from the PR-head
    details during selection, and an unreadable baseline failed the landing
    closed. The planner, however, projects main's copy, and a cosmetic refresh
    must never wedge a landing.
  - **Delta:** `landingOwnedIssue` carries the done `card` and an injected
    `readCard` (`tracker.Repository.ReadCardBlob`) instead of pre-resolved
    `baseline` bytes. `trackedArchiveBytes` resolves the baseline from the
    content it projects (main's copy); a read failure keeps the bytes.
    `planLandingArchive` stays deterministic given that reader; it is pure in
    the seam sense, with IO only through the injected function. `archivedDetails`
    remains the pure core and is unit-tested without IO.
  - **Delta:** `pinArchivedCard` compares against main's source copy.
  - **Delta:** `retryCloseMirror` in `issue recovery reconcile` re-attempts the
    mirror commit after a crash between codecomplete and that commit. It
    replaces the per-receipt call.

