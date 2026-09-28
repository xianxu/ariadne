---
id: 000212
status: open
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
github_issue:
---

# Add dup to the issue status vocabulary

## Problem

There is no way to close an issue as a duplicate. `categories.terminal` is
`["done", "wontfix", "punt"]` (`construct/vocabulary/issue.cue`), and none of
them is true of a duplicate:

- `wontfix` — "rejected; will not be done". False: the work *is* being done,
  under another id.
- `punt` — "deferred". False: nothing is deferred.
- `done` — false, and it would pollute velocity calibration with a close that
  measured no work.

So closing a duplicate records a claim about the work that is wrong. Concrete
instance from 2026-09-02: `pair#178` was closed `wontfix` while its entire
content lives on as `pair#165` — the file says "rejected" about work that is
scheduled.

**`dup` alone is not enough.** A duplicate that does not say *which* issue
supersedes it is barely better than `wontfix` plus prose: the reader still has
to search. The status needs a pointer beside it.
