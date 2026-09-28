---
id: '000058'
status: done
created: 2026-06-01
updated: 2026-06-03
estimate_hours: 1.5
actual_hours: 1.0
---

# milestone-close — base review window on prev milestone close (close inter-milestone gap)

## Problem

`sdlc milestone-close` computes its fresh-eyes review window as **"first commit
referencing `#<issue> Mx`"^..HEAD** — the parent of the first commit tagged with
that specific milestone. So a commit that references the issue (`#<issue>`) but
**not** a milestone — a side-quest, a fix, a plain `#N` commit made between two
milestone closes — can fall in the gap: it post-dates the previous milestone's
window and is *excluded* as the base of the next one. It escapes both reviews.

Surfaced in #57: `c52f482` ("#57: dev-aliases — build to owner bin/") referenced
`#57` but not `M2`. M1's window ended at its close (`76a2828`); M2's window based
at `699635d^` = `c52f482`, excluding it. The M2 judge happened to review it as
compensation (it was the base), but that's luck, not guarantee — a commit two or
more before the milestone's first `Mx` commit would simply vanish from review.

This defeats the milestone-review's purpose (every change reviewed before the
next milestone) precisely for the unplanned commits (`side-quest:`, fixes) that
most warrant a fresh eye.
