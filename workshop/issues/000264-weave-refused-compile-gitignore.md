---
id: 000264
status: working
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: 'fa51f845627c8b76ee852750ed91bdd041382f1a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T19:13:03-07:00
flow: {kind: quick, provenance: inferred, spec: "88c7843a", done: "6eb13e58"}
---

# weave compile: a refused compile must not rewrite .gitignore

## Problem

`weave compile` reconciles two ownership scopes in sequence
(`cmd/weave/main.go`): data mounts per owner (`ApplyManaged(…, ScopeData)`),
then the leaf's artifacts (`ApplyManaged(…, ScopeArtifacts)`). Each call
rewrites the root `.gitignore` via `managedIgnoreText`, which also *migrates*:
it strips every line of the historical fixed list
(`GeneratedRuntimeGitignoreEntries`) and writes a managed block holding only
the outputs recorded so far.

When the artifacts pass then refuses (e.g. `preserving authored replacement`
on pre-inventory `construct/generated/vocabulary/.source-sha`), the data pass
has already published a `.gitignore` whose legacy list is gone and whose block
covers only `/construct/generated/weave/`. Every generated file (AGENTS.md,
CLAUDE.md, skill links, `.colima/`, …) surfaces as untracked. The operator sees
a failed compile plus an unexplained pile of "new" files.

Observed 2026-09-28 in parley.nvim (and earlier in metis, parli, xianxu.dev per
ariadne#241's log, each hand-restored). Named in #241's "Candidate fixes" as "a
compile that stages its `.gitignore` edit with the outputs it covers"; split
out here so #241 stays about publication.

## Spec

Invariant: **the legacy fixed list is retired only by the pass that records
its replacement.** Every legacy entry (`/AGENTS.md`, `/.claude/skills/`, …)
covers an artifacts-scope output, so only an artifacts-scope `ApplyManaged`
may strip it — in the same write that puts the recorded artifact paths into
the block. A data-scope pass keeps legacy lines verbatim (it still manages its
own block entries, per #263's ownership rule). A refused artifacts pass fails
in its preflight, before any write, so the legacy list survives and nothing
generated surfaces as untracked.

Second rule: a pass with nothing to reconcile — no intended outputs and no
prior identities in its scope — writes nothing (no inventory, no `.gitignore`).
This makes the common failing case (a repo without data mounts, like
parley.nvim) byte-identical, and stops dependency owners with no mounts from
publishing an empty inventory they don't need.

ARCH: root cause in the ownership rule (`ApplyManaged`/`splitIgnore`), not by
reordering `compilePrepared` — any other caller of a data pass would still
strip. ARCH-FUNERAL: creates nothing durable; the second rule removes an
empty-inventory write rather than adding one.

## Done when

- Regression (fails on main): legacy list outside the block, no inventory,
  data `ApplyManaged` with no mounts → `.gitignore` byte-identical and no
  inventory written.
- Regression (fails on main): same, with a data mount → the mount's entry is
  added but every legacy line survives; a following artifacts pass then
  migrates them away.
- Existing ownership/gitignore tests pass (incl. #263's
  `TestManagedPreservesCommittedBlockWithoutInventory` and legacy migration in
  `TestManagedIgnoreMigrationAndLocalNegations`); `go test ./cmd/weave/...`.
- Live: a parley.nvim-shaped scratch repo (pre-inventory generated file, legacy
  list) → `weave compile` refuses and `git status` shows no change.

## Plan

- [x] Tests first: the two regressions above (red on main).
- [x] `splitIgnore`/`managedIgnoreText` take a migrate flag; `ApplyManaged`
      passes `scope == ScopeArtifacts`; the plain-Apply `EnsureGitignore` path
      keeps migrating.
- [x] `ApplyManaged`: return early when the scope has no wanted and no
      previous identities.
- [x] `go test ./cmd/weave/...`; live scratch-repo check.

## Revisions

- 2026-09-28 — Spec replaced at design. Reason: the filing-time candidate
  ("the data scope does not write `.gitignore`") would leave data mounts
  unignored, and #263 (merged after filing) already fixed in-block clobbering;
  what remains is the outside-block legacy strip. Delta: Spec/Done-when now
  scope legacy migration to the artifacts pass and make empty passes no-ops.

## Log

### 2026-09-28

Filed from the parley.nvim investigation: `ownership.json` held
`"outputs": []` (written by the data pass) while the artifacts pass had refused
`vocabulary/.source-sha`; the `.gitignore` diff showed the legacy list removed
and a one-entry block.

Implementation (1bc6c5f4): both regressions red on main (the data pass wrote
`/construct/generated/weave/` into a new block and dropped `/AGENTS.md`,
`/.claude/skills/`, `/construct/generated/`), green after. `go test
./cmd/weave/...` passes. Live: a detached parley.nvim worktree at f30cdefc
with its real pre-inventory `construct/generated/` (from today's backup),
`AGENTS.md` and `.colima/` present — main's weave refuses on
`vocabulary/.source-sha` and leaves ` M .gitignore`, `?? .colima/`,
`?? AGENTS.md` plus an empty inventory; this branch's weave refuses identically
and `git status` stays clean. The test compile cloned a private ariadne
dependency beside the worktree; both removed.
