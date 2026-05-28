# Issue Lifecycle

## Flow

```
GitHub Issue → sdlc fetch --github-issue 42 → workshop/issues/000042-slug.md → sdlc claim → work → sdlc push (main) or sdlc change-code → sdlc pr → sdlc merge (branch)
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

1. **Fetch**: `sdlc fetch --github-issue <num>` creates a local issue file from GitHub with frontmatter (id, status, github_issue, dates)
2. **Claim**: `sdlc set-status --issue N working` then `sdlc claim --issue N` publishes the issue-state claim to main
3. **Work**: Agent works within the issue file — updates Plan, Log, Spec sections
4. **Small work on main**: `sdlc push` auto-commits, runs pre-merge checks, pushes, archives done issues to `history/`, closes GitHub issues
5. **Large work on branch**: `sdlc change-code` chooses the branch/worktree path after planning; `sdlc pr` opens the pull request; `sdlc merge` archives and cleans up

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

**Branching decision**: `sdlc change-code --issue N` runs structural checks and
the plan-quality judge, then asks whether to use a worktree or an in-place
branch. Agents call it after planning; if stdin is non-interactive, the binary
emits an `ASK_BRANCHING_STRATEGY` sentinel so the agent can ask the operator and
rerun with `--worktree=yes` or `--worktree=no`.

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
