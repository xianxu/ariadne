---
id: '000213'
status: done
started: 2026-09-03T11:04:17-07:00
created: 2026-09-02
updated: 2026-09-05
actual_hours: 7.34
---

# Allocate issue IDs against origin/main

## Problem

`issue.NextID` scans **local files only** —
`cmd/sdlc/internal/issue/scaffold.go:31`, walking `issuesDir`, `historyDir` and
the archive subdir for the highest 6-digit prefix. A feature branch cut before
some issue landed on `main` has a `workshop/issues/` that never contained it, so
`sdlc issue new` on that branch reallocates ids that already exist.

**This is not a race.** Measured in `pair`, 2026-09-02:

```
12:10:10  branch 000170-… cut from main (88fe1de0)
12:42:03  #171, #172 land on origin/main
15:43:29  actor files its own #171 on the branch
17:07:08  actor files its own #172
```

Three hours after publication. Any `sdlc issue new` on that branch allocates
`171`, whether run then or next week — the branch simply does not contain the
newer files. The result is four files and two ids:

| id | on `origin/main` | on the branch |
|---|---|---|
| 171 | Always-on idle notification fallback | Reconcile stale incarnations left by a crashed couch |
| 172 | Clickable status bar switches to that actor | Parallelize the zellij session snapshot |

**The repo transaction lock is the wrong tool and demonstrably did not help.**
Both allocations ran in linked worktrees of the same repo, so they shared one
`.git/sdlc.lock` and *were* serialized. They still collided. The lock prevents
concurrent access to a shared view; it cannot reconcile two disjoint views.

**And nothing detects it, and never has.** Two colliding issues get different
slugs, so `000212-fleet-policy-json.md` and `000212-issue-status-dup.md` are
different *paths*. Git merges both cleanly. There is no point in the lifecycle at
which anything objects — not the push, not the merge, not `issue validate`.

An earlier draft of this section blamed `#206`'s `syncInPlace` for removing
push-rejection detection. **Measured, that is wrong**, and the correction matters
because it changes the age of the defect:

| repo | colliding ids |
| --- | --- |
| ariadne | #40, #96, #168, #212 |
| parley.nvim | #51, #66, #81, #90 |

`ariadne#40`'s two files were created 2026-05-28 and 2026-05-27; `#96`'s on
2026-06-14 — three months before `#206` landed (2026-09-02). Three of ariadne's
four are already merged and archived. A rejected push only ever caught a
*non-fast-forward race*, never an id collision, because the filenames differ.
This has been silently corrupting the id space since May.

**There is no double-locking anywhere.** `NextID` has one caller
(`cmd/sdlc/issue.go:261`) and does a plain `os.ReadDir`; no code path re-checks
the id at commit, push, or merge. The repo transaction lock is real but
orthogonal — it serializes access to one checkout and cannot reconcile two
disjoint views.
