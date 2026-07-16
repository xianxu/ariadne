---
id: 000005
status: done
deps: []
created: 2026-04-21
updated: 2026-04-21
---

# Add `make issue-sync` target

## Summary

Add a `make issue-sync` Makefile target that synchronizes `workshop/issues/` changes to main and pushes to origin. This enables using issues as a coordination/locking mechanism across branches and collaborators.

## Spec

### Behavior

**Case 1: On main branch**
- Stage all changes and untracked files in `workshop/issues/`
- Commit and push to origin

**Case 2: On feature branch (worktree)**
1. Identify changed + untracked files in `workshop/issues/` on the feature branch
2. Find the main worktree, verify it's on `main`
3. In main worktree: `git pull --rebase origin main`
4. Compute merge base between main and feature branch
5. Check if any of the feature branch's changed issue files were also changed on main (since merge base)
6. **No overlap**: copy files to main worktree, `git add`, commit, push
7. **Overlap detected**: stop, list conflicting files, print plain-English resolution steps for the user

### Conflict resolution message

When overlap is detected, print:
- Which files conflict
- Step-by-step instructions to manually merge in the main worktree (assuming user is not a git expert)

### Edge cases
- No changes in `workshop/issues/` → no-op with message
- Main worktree has uncommitted changes in `workshop/issues/` → warn and stop
- Push fails → show error, don't swallow it

## Plan

1. Write `scripts/issue-sync.sh` with the logic above
2. Add `issue-sync` target to Makefile
3. Test both cases (on main, on feature branch)

## Log
