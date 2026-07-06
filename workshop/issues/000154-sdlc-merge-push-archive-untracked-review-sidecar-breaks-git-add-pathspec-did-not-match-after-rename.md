---
id: 000154
status: working
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-05
estimate_hours: 0.53
started: 2026-07-05T22:54:11-07:00
---

# sdlc merge/push archive: untracked review sidecar breaks git-add (pathspec did not match) after rename

## Problem

`sdlc merge`'s archive step fails its commit when the boundary-review **sidecar**
(`workshop/plans/NNNNNN-<slug>-close-review.md`) is **untracked** at merge time.
The PR itself merges (main advances), but the local archive-to-history commit
dies with:

```
==> Committing archived history in main...
Error: git -C <repo> add: exit status 128
fatal: pathspec 'workshop/plans/NNNNNN-<slug>-close-review.md' did not match any files
```

…leaving main in a half-archived state: the issue file is moved to
`workshop/history/` (untracked) and staged-deleted from `workshop/issues/`, the
sidecar is moved to `workshop/history/` (untracked), but the archive commit is
never made. Recovery is manual: `git add` the destination history paths (+ the
issue rename) and commit.

**Reproduced deterministically 3× this week** — parley.nvim #158, #159, #156.
(Earlier closes #152/#154/#155/#157 archived cleanly, so it is *state-dependent*,
not universal — see the tracked-vs-untracked analysis below for the likely
discriminator.)

## Root cause

The archive move+add is split across `cmd/sdlc/push.go` / `cmd/sdlc/merge.go`:

- `archivePlanArtifacts` (`push.go:255`) **renames** each plan/review sidecar from
  `workshop/plans/` to `workshop/history/` on disk, then records a
  `preparedArchiveMove{ IssuePath: <plansDir>/base, HistoryPath: <historyDir>/base }`
  (`push.go:276-277`) — i.e. `IssuePath` is the **pre-move** plans path.
- `archiveAddArgs` (`push.go:222-225`) then builds `git add -- <IssuePath> <HistoryPath>`
  for every move — so for the sidecar it runs
  `git add -- workshop/plans/<sidecar> workshop/history/<sidecar>`.

For the **issue file** this works: it was committed on the branch, so
`git add <old issues path>` stages a real *deletion* of a tracked file. For the
**sidecar** it breaks when the sidecar is **untracked**: after the rename, the
plans path has no file on disk *and* no tracked entry, so `git add
workshop/plans/<sidecar>` → `pathspec did not match` (exit 128), aborting the
whole `git add` (and the archive commit).

**Why the sidecar is often untracked:** `sdlc close` *creates* the sidecar after
the implementer's last commit; the implementer then typically makes a
FIX-THEN-SHIP fixup commit staging **explicit paths** (not the sidecar), and
`sdlc pr` pushes the branch without it. So the sidecar reaches `sdlc merge`
untracked and not on origin — exactly the case `git add <old plans path>` can't
handle after the rename.

## Spec

Make the archive git-add robust to an **untracked** (or already-moved) sidecar
source. Options (decide at design time):

- **Add the destination only for untracked sources.** For each move, `git add
  -- <HistoryPath>` always stages the moved file at its new location; add
  `<IssuePath>` (the source) to the add-list **only if the source is tracked**
  (so its deletion is staged). An untracked source needs no source-side add — its
  old path simply ceases to exist. This is the minimal, targeted fix and keeps the
  tracked issue-file path (delete-old + add-new) intact.
- **Or** stage before renaming (compute the add-args against the pre-move tree,
  then move), or use per-path `git add -A` guarded by an existence/tracked check.

Keep the existing precise-path discipline (`archiveAddArgs` deliberately avoids a
broad `git add <issuesDir>/ <historyDir>/`, `push.go:215-221`) — the fix should
stay path-scoped, just not add a vanished untracked source path.

## Done when

- `sdlc merge` (and `sdlc push`, same `archiveDoneIssues*` path) completes the
  archive commit when the review sidecar is untracked at merge time — no
  `pathspec did not match`, no half-archived main.
- A tracked sidecar (committed on the branch) still archives correctly (deletion
  of the old plans path staged, new history path added).
- The issue file (always tracked) archives exactly as today.
- Unit test on the pure seam: `preparedArchiveMove` → `archiveAddArgs` (or a new
  helper) omits an untracked/missing source path from the add-list while keeping
  the destination — cf. the existing table tests in `cmd/sdlc/push_test.go`
  (`archiveAddArgs` / `preparedArchiveMoves`). Fake the tracked-vs-untracked
  probe so the git IO is injected.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.1 impl=0.2
item: milestone-review     design=0.0 impl=0.2
design-buffer: 0.30
total: 0.53
```

## Plan

Design (decided at plan time — Option A from Spec, "add destination always; add
source only if tracked"):

- Add `SourceUntracked bool` to `preparedArchiveMove` (`push.go`). Default `false`
  = "add both paths" → **every existing move site is behavior-preserved with zero
  edits** (issue-file moves, recovery moves — all always carry a tracked source).
- `archiveAddArgs` stages `IssuePath` only when `!SourceUntracked`; `HistoryPath`
  is always staged (the moved file physically exists at its new path). Keeps the
  precise-path, no-broad-`add <dir>/` discipline (#80).
- Only plan sidecars can be untracked, so the probe is confined to
  `archivePlanArtifacts`: inject an `untracked func(recPath string) bool` (thin IO
  seam, ARCH-PURE) that runs `git ls-files -- <recPath>` in the caller's git dir
  (empty ⇒ untracked). Conservative: on git error → treat as tracked (preserve old
  behavior). Push passes a `pushRunner.Git`-backed probe; merge a
  `mergeRunner.GitInDir(mainPath, …)`-backed probe (each resolves the recorded,
  git-relative path in the right worktree).

Verified empirically: `git add -- <vanished-untracked-src> <dest>` → exit 128
"pathspec did not match" (the bug); `git add -- <dest>` alone → OK.

- [x] Add `SourceUntracked` field + probe injection; make `archiveAddArgs` omit an
      untracked source from the add-list. Pure-seam table test on `archiveAddArgs`
      (untracked move → dest only; tracked → both) with the git probe faked, plus a
      regression covering `archivePlanArtifacts` archiving an untracked sidecar.

## Log

### 2026-07-01

Filed from three consecutive parley.nvim closes (#158, #159, #156) that each hit
this at `sdlc merge`. The PR merge always succeeded; only the local archive commit
failed, recovered by hand each time. Suspect seam: `archivePlanArtifacts`
(`push.go:255`) records the pre-move plans path as `IssuePath`; `archiveAddArgs`
(`push.go:222`) adds it; `git add` of an untracked+vanished path is the fault.
Symmetric to #148 (another `sdlc merge` edge that "silently did the wrong thing"),
though this one fails loud (exit 128) rather than silent — still leaves main in a
half-state a form-gate should prevent.

### 2026-07-05 — implemented (Option A)

Shipped the minimal targeted fix:

- `preparedArchiveMove` gains `SourceUntracked bool` (default `false` = pre-#154
  "stage both halves" → every existing move site — issue files, durable plans,
  recovery moves — is behavior-preserved with zero edits).
- `archiveAddArgs` stages `IssuePath` only when `!SourceUntracked`; `HistoryPath`
  always (the moved file physically exists at its new path). Keeps the precise,
  no-`add <dir>/` discipline (#80).
- Only plan sidecars can be untracked, so the probe is confined to
  `archivePlanArtifacts`, which gains an injected `srcUntracked func(recPath) bool`
  (ARCH-PURE — the git IO stays a thin, faked-in-tests seam). Backed by the shared
  `gitSrcUntracked` builder: `git ls-files -- <recPath>` empty ⇒ untracked; on any
  git error ⇒ treat as tracked (conservative — never drop a real deletion). Push
  passes a `pushRunner.Git` probe (cwd); merge a `mergeRunner.GitInDir(mainPath,…)`
  probe (main worktree). The recovery path (`preparedArchiveMoves`) is untouched:
  it only forms a move on a *staged deletion*, so its sources are inherently
  tracked — confirmed by the plan-quality judge.

**Verification (behavior diff vs main, not just green tests):**

- `TestArchiveDoneIssues_UntrackedSidecar_RealRepo` — end-to-end in a *real* git
  repo (`hermeticRepo`): commits an issue + durable plan, leaves the review
  sidecar untracked, runs the real `archiveDoneIssues` → real `git ls-files`
  probe → real `git add`/`commit`. Passes; worktree ends clean, all three files
  tracked in `history/`.
- **Proved it catches the bug:** temporarily reverting the `archiveAddArgs` guard
  (always add `IssuePath`) makes that same test fail with the exact production
  symptom — `exit status 128 / fatal: pathspec 'workshop/plans/000154-x-close-review.md'
  did not match any files`. Restored; test green again.
- Pure-seam unit tests: `TestArchiveAddArgs` (untracked → dest only; mixed →
  tracked stages both, untracked dest only), `TestGitSrcUntracked` (empty→untracked,
  echoed→tracked, git-error→tracked), `TestArchivePlanArtifacts_UntrackedSidecarStagesDestOnly`
  (faked probe; asserts index-based, not `os.Stat`, classification).
- Full `go test ./cmd/sdlc/` green (18.9s); `go build ./...`, `go vet`, `gofmt`
  clean. Incidental: `gofmt -w` corrected a pre-existing whitespace realignment in
  `merge.go`'s `mergeAction` const block (latent drift on main, unrelated to #154).

Atlas: updated `atlas/workflow/sdlc-binary.md` "sdlc push archive recovery" with
the untracked-sidecar seam.
