---
id: 000223
status: open
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
github_issue:
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
