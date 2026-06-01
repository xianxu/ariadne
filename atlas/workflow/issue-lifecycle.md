# Issue Lifecycle

## Flow

```
Issue created (sdlc issue new "<title>", or sdlc issue new --from-github 42) → workshop/issues/NNNNNN-slug.md → sdlc claim → sdlc change-code (in-place branch by default) → work → sdlc pr → sdlc merge   [direct sdlc push on main still available, but not the default]
```

## States

| Status | Meaning |
|--------|---------|
| open | Active work |
| working | An agent is working on something |
| done | Completed, awaiting archive |
| wontfix | Declined |
| punt | Deferred |

## Transitions

1. **Create**: `sdlc issue new "<title>"` allocates the next ID and writes the canonical template (the no-GitHub entry path); `sdlc issue new --from-github <num>` (or the older `sdlc fetch`) seeds it from a GitHub issue. See `sdlc issue --help` for the canonical issue-file contract.
2. **Claim**: `sdlc claim --issue N` flips an open issue to `working` and publishes the issue-state claim to main in one step (`--no-start` to skip the flip)
3. **Work**: Agent works within the issue file — updates Plan, Log, Spec sections
4. **Default — branch + PR**: `sdlc change-code` creates an **in-place branch** (a branch in the current checkout) after the gates; `sdlc pr` opens the pull request; `sdlc merge` merges it server-side, archives done issues, and switches back to main. `--worktree=yes` gets an isolated worktree instead (parallel work).
5. **Shortcut — direct on main**: `sdlc push` (auto-commit, pre-merge checks, push, archive, close GH issues) still exists for quick one-liners, but is no longer the default (#51).

## Worktree layout

Worktrees are created at `../worktree/<repo-dir-name>/<branch-name>/`, keeping
worktrees from different repos separated. The `<repo-dir-name>` is the basename
of the current working directory (i.e., the repo folder name).

```
../worktree/
└── my-repo/
    ├── 000042-add-feature/    ← branch: 000042-add-feature
    └── 000051-fix-bug/        ← branch: 000051-fix-bug
```

**Branching decision** (#51): `sdlc change-code --issue N` runs structural checks
and the plan-quality judge, then branches. The default (no `--worktree` flag) is
**in-place** — a branch in the current checkout, no worktree dir; the common
case, chosen without prompting. `--worktree=yes` gets an isolated worktree (the
layout above); `--worktree=ask` restores the interactive prompt, or for a
non-interactive agent emits the `ASK_BRANCHING_STRATEGY` sentinel (exit 2) so the
agent can ask the operator and rerun with `--worktree=yes|no`.

**Navigation**: worktree creation writes the path to `.goto`; the shell `g`
alias reads it to `cd` you there. `sdlc merge` writes the main worktree path
back into `.goto` for the return trip.

## Issue file structure

```markdown
---
id: 000042
status: open
deps: []
github_issue: 42
created: 2026-04-20
updated: 2026-04-20
estimate_hours:    # optional at create; required when status=working
actual_hours:      # required when status=done
---

# Title

## Done when
- acceptance criteria

## Spec
- brainstorming results (if needed)

## Plan
- [ ] checklist of work

## Log
### 2026-04-20 — session summary
One paragraph: what was attempted, what landed, what got deferred.

### 2026-04-20
- individual decisions, discoveries

## Side quests
- (optional; recommended for multi-day issues) name + ~time + commit ref
```

For frontmatter and section conventions see the **xx-issues skill**
(`construct/local/issues/SKILL.md`). For the cross-artifact closing
sweep (actual_hours, project-file update, atlas update, validation
log) see **AGENTS.md §5 closing checklist**.

## Closing

Each `sdlc push` / `sdlc merge` archives done issues into `history/`. Before that, run the **closing checklist** from AGENTS.md §5:

1. Verify behavior.
2. Tick the milestone in `## Plan` and flip `status` to `done`.
3. **Record `actual_hours`** in the frontmatter (required at close).
4. Update the parent project file (if any).
5. Update `atlas/` for any new architectural surface.
6. Append validation-log entry if estimated under a versioned playbook (AGENTS.md §4).
