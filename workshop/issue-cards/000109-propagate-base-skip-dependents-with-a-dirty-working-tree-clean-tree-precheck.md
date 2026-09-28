---
id: '000109'
status: done
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 1
actual_hours: 0.11
---

# propagate-base — skip dependents with a dirty working tree (clean-tree precheck)

## Problem

`sdlc propagate-base`'s `commitConsumption` stages the consumption with `git add
-A`, which assumes ANY dirty state in a dependent is the re-weave's own output. It
isn't: pre-existing uncommitted/untracked work in a dependent (e.g. **a concurrent
agent session mid-edit in a sibling repo**) is indistinguishable, so it gets
swept into a mislabeled `<ref>: consume base-layer change` commit.

Hit live on the #107 atlas-prose propagation (2026-06-16): a concurrent Claude
session was editing `parley.nvim`'s `workshop/plans/000128-*-plan.md`. The re-weave
itself produced no tracked diff (the woven constitution is gitignored everywhere),
yet `git add -A` committed that session's in-flight plan work under the consumption
message. Operator caught it; it was undone with `git reset --mixed HEAD~1` and the
other session re-committed its work properly — but it *raced*, and only resolved
cleanly by luck. With agents working peer repos in parallel, this is a live hazard.

This is the deferred-follow-on hardening flagged in #106's Revisions
("branch-first per dependent … the manual loop lacked"); the clean-tree precheck
is the minimal, root-cause slice of it.
