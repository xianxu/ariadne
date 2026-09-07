
## Log

### 2026-09-06 — field evidence from pair (was pair#195)

`pair#195` was filed against this same failure in the consumer repo before it
was noticed that `sdlc` owns it (`cmd/sdlc/claim.go:601`). It is now closed as a
duplicate of this issue; its evidence is grafted here because it is the
strongest case either issue carries — this is no longer a hypothetical race.

**Three ID collisions have actually occurred**, each costing a renumber plus
reference-chasing:

| ids | collided files | resolved |
| --- | --- | --- |
| `000172` | `clickable-status-bar` vs `parallelize-zellij-session-snapshot` | latter → `pair#191` |
| `000173` | `wire-actor-description` vs `disposition-six-production-symbols` | latter → `pair#192` |
| `000179` | `layout2-terminal-toggle` (open) vs `reattach-a-detached-thread` (**done, archived**) | former → `pair#194` |

The third is the worst shape and worth designing against specifically: one side
was already closed and archived, so `#179` named both a live issue and a shipped
one, and `sdlc claim --issue 179` refused outright with "multiple issue files
match" — the collision **blocked the issue from being worked at all**, and the
archive meant the duplicate was invisible to anyone reading `workshop/issues/`.

**Why this repo hits it structurally, not incidentally.** `sdlc change-code`
branches **in place** by default rather than creating a worktree, so a repo
being actively worked has no worktree on `main` — the reservation mechanism is
unavailable exactly when the workflow's own default mode is in use.
`git worktree list` in `pair` on 2026-09-06 confirms: the only checkout is on
`000172-clickable-status-bar`.

**Current exposure, measured 2026-09-06.** Every issue filed in `pair` across
these sessions printed the warning and is unreserved right now: **`pair#185`
through `pair#204`**, which includes nine filed in a single session that day
(`#196`–`#204`). Any concurrent session allocating "the next free ID" collides
with one of them and does not find out until someone runs `sdlc claim`.

Note the asymmetry that makes this easy to miss: `ariadne`'s own issues
(`#215`, `#216`, filed the same day) published to `origin/main` without
complaint, because this checkout *is* on `main`. The failure is invisible from
the repo that owns the code.
