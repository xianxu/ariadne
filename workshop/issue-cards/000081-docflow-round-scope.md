---
id: '000081'
status: done
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.5
actual_hours: 0.5
---

# docflow round: stage only in-scope started docs, not git add -u

## Problem

`docflow round` (no explicit files) stages with `git add -u` — *every* tracked
file that's modified, not just the doc(s) under review. In a working tree that
holds unrelated tracked WIP (another post being edited, a config tweak, a peer's
change), a review round sweeps those into the journaled round commit, conflating
unrelated changes with the prose review. Surfaced as dogfooding feedback during
#79's first real use (the blog-post review) — harmless there only because the
post was the sole tracked-modified file. Same hazard family as `lessons.md`
"[[git add -A sweeps unrelated untracked WIP]]" — docflow shipped with a latent
instance of the very rule that lesson states.
