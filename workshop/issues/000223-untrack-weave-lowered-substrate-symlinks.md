---
id: 000223
status: open
deps: []
github_issue:
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
---

# weave-lowered substrate symlinks are tracked in every derivative, so weave's own prune and refresh show up as git churn

## Problem

weave lowers the base layer's `symlink` manifest rows into each derivative and
owns ignoring what it produces (`cmd/weave/internal/plan/gitignore.go`: the
three faces, `.claude/skills/`, `.agents/skills/`, `.claude/settings.json`,
`/.colima/`). But **28 of those lowered symlinks are tracked in git** in every
derivative — pair, parley.nvim, tools, kbench (28 each), nous (29):

    Makefile, Makefile.workflow
    scripts/{lib,docflow,parallel-checks,pre-merge-checks,run-merge-checks,sdlc-install}.sh, scripts/close-issue.py
    construct/dev-aliases.sh
    construct/scripts/{apply-gitignore-entries,bootstrap-peers,clone-data-deps,lib-deps,list-peers}.sh
    .openshell/{Makefile,policy.yaml,sandbox.sh,ssh_wrapper.sh,dotfiles,overlay,ssh-bin}
    .tart/Makefile, .tart/scripts
    .codex/config.toml, .claude/settings.ariadne.json, atlas/workflow
    scripts/issue-sync.sh          <- no longer in base.manifest at all

27 are exactly the manifest's `symlink` rows; `weave compile --dry-run` in pair
plans every one of them. They are tracked because `#38 M4` (2026-05-27, "drop
construct/vendor/, refresh against symlink-only substrate") replaced tracked
vendored *copies* with symlinks in place — `Makefile.workflow` went from 630
lines to a link — and never untracked them. `/.colima/` links got the ignore;
these predate `gitignore.go` and were missed. Not a decision; a migration
artifact.

**Measured 2026-09-11**, `weave compile` in a scratch clone of pair with
ariadne beside it (both as a symlink to the primary and as a real checkout —
same result either way):

    weave: applied 99 action(s); pruned 1 orphaned lowered symlink(s): scripts/issue-sync.sh
    git status:
      D  scripts/issue-sync.sh                          <- weave pruned a TRACKED file
      M  .github/workflows/merge-check.yml              <- seed refreshed; the tracked copy had drifted
      ?? scripts/merge-checks.d/40-duplicate-issue-id.sh <- lowered, not ignored

So two owners disagree about the same paths: weave prunes and refreshes them
as its output, git tracks them as authored content. Every derivative is behind
weave's plan on all three of those today, which means nobody runs `weave
compile` in a derivative without producing a diff — which is the incentive not
to run it.

Not measured but implied: any manifest change (a row dropped, a target moved)
becomes a tracked-file edit in five repos instead of a no-op recompile.

Found while designing couch worktree slots (brain pensive
`2026-09-11-01-pensive-couch-slots.md`). The slot design does not depend on
this — a symlinked peer beside a worktree keeps the link text `../ariadne/…`
stable, because `layergraph/walk.go:113-117` physicalizes only the parent
directory — but the churn is real on its own.

## Spec

weave owns every path it lowers, in git as well as on disk.

1. **Untrack the lowered symlinks** in each derivative (`git rm --cached`),
   leaving them on disk. Delete `scripts/issue-sync.sh` outright; it is not
   in the manifest.
2. **Ignore them where weave ignores its other outputs**: add the manifest's
   `symlink` targets to `gitignore.go`'s entry set — derived from the manifest
   rather than listed by hand, so a new `symlink` row is ignored the moment it
   is lowered (ARCH-DRY: one owner for "this artifact is weave-produced").
   `construct/scripts/apply-gitignore-entries.sh`'s list stays what it is; it
   is not weave's.
3. **Seeds are not this issue.** `.github/workflows/merge-check.yml` is a
   `seed` (a tracked, content-tracking copy by design). Its drift is refreshed
   by weave and committed like any other change; that is the seed contract
   working. Note it in the Log; do not change it here.
4. **Keep fresh-clone bootstrap working.** `bootstrap.sh` clones peers, then
   `exec make bootstrap` — which today resolves through the *tracked*
   `Makefile` link once `../ariadne` exists, and `make bootstrap` is what
   builds weave. Untracking `Makefile` breaks that handoff. Either hand off as
   `make -f "$repo_root/../ariadne/Makefile" bootstrap` (the peer is present by
   then), or keep `Makefile` alone as a `seed`. Prefer the `make -f` handoff:
   one fewer tracked path, and the seed kind exists for content that must work
   *before* peers are present, which `Makefile` does not.

Out of scope: `construct/deps` (tracked by design; a per-tree override is a
separate question), the `worktree/<repo>/<peer>` symlink convention, anything
in couch.

## Done when

- `git ls-files -s | awk '$1=="120000"'` returns no manifest-lowered paths in
  pair, parley.nvim, tools, kbench, nous.
- `weave compile` followed by `git status --porcelain` in each derivative is
  empty apart from genuine seed refreshes.
- A fresh `git clone` of a derivative followed by `./bootstrap.sh` completes
  (peers cloned, weave built, links lowered) with no tracked `Makefile`.
- `gitignore.go` derives the ignored set from the manifest's `symlink` rows;
  a test adds a row and asserts it is ignored without editing the list.

## Plan

- [ ] `gitignore.go`: derive `symlink`-row targets from the walked manifests; test with a synthetic row
- [ ] `bootstrap.sh`: `make -f` handoff; test under `BOOTSTRAP_CLONE_ONLY` and a real fresh clone
- [ ] Untrack in the five derivatives (one commit each: `git rm --cached` + delete `scripts/issue-sync.sh` + `weave compile` to prove a clean status)
- [ ] Manual: fresh clone of pair → `./bootstrap.sh` → `weave compile` → `git status` empty

## Log

### 2026-09-11

- Filed from the brain advisor session. Inventory by `git ls-files -s | awk
  '$1=="120000"'` across the five derivatives; manifest match by basename
  against `construct/base.manifest` (27/28 match; `issue-sync.sh` does not).
- Reproduction in `$TMPDIR` scratch clones, two arms (peer as symlink to the
  primary; peer as real checkout). Both: `Makefile -> ../ariadne/Makefile`
  unchanged; `D scripts/issue-sync.sh`, `M .github/workflows/merge-check.yml`,
  `?? scripts/merge-checks.d/40-duplicate-issue-id.sh`. The primary pair is
  behind on all three; the other four derivatives still track `issue-sync.sh`.
- A prediction was refuted on the way: I expected the symlinked-peer arm to
  rewrite the 28 links to a physical-path relative form. It does not —
  `walk.go:113-117` physicalizes the parent and re-joins the logical name.
  Recorded so nobody re-derives the wrong mechanism.
