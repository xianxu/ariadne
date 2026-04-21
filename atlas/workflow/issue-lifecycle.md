# Issue Lifecycle

## Flow

```
GitHub Issue → make fetch 42 → workshop/issues/000042-slug.md → work → make push (or make worktree → make pull-request → make merge)
```

## States

| Status | Meaning |
|--------|---------|
| open | Active work |
| done | Completed, awaiting archive |
| wontfix | Declined |
| punt | Deferred |

## Transitions

1. **Fetch**: `make fetch <num>` creates a local issue file from GitHub with frontmatter (id, status, github_issue, dates)
2. **Work**: Agent works within the issue file — updates Plan, Log, Spec sections
3. **Small work on main**: `make push` auto-commits, runs pre-merge checks, pushes, archives done issues to `history/`, closes GitHub issues
4. **Large work on branch**: `make worktree` → isolated branch → `make pull-request` → `make merge` → archives and cleans up

## Issue file structure

```markdown
---
id: 000042
status: open
deps: []
github_issue: 42
created: 2026-04-20
updated: 2026-04-20
---

# Title

## Done when
- acceptance criteria

## Spec
- brainstorming results (if needed)

## Plan
- [ ] checklist of work

## Log
### 2026-04-20
- what happened
```
