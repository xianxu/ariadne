---
id: '000113'
status: done
created: 2026-06-17
updated: 2026-06-17
estimate_hours: 4
actual_hours: 0.52
---

# Decouple sdlc claim (lock) from the estimate gate; anchor active-time window to the claim commit

## Problem

Two coupled defects, surfaced while refining estimate/actual calibration (#112):

1. **`sdlc claim` bundles the estimate guard onto its original job.** Claim's
   real purpose is a *lock*: flip `open → working` + broadcast to origin/main so
   peer agents don't collide. But it also requires `estimate_hours` — and at claim
   time the operator often *can't* estimate yet (brainstorm/design hasn't
   happened). So claiming EARLY — which is exactly what you want — is blocked by a
   premature estimate demand.

2. **active-time's window starts too late.** `gitx.CommitWindow` anchors the
   window at the parent of the first `#N`-subject commit, so operator-attention-
   heavy DESIGN time (brainstorm / spec / plan / reviews) *before* the first code
   commit is excluded from `sdlc actual` — systematically under-measuring the
   scarce resource (operator attention). Observed live: #111 measured 0.35h though
   design + four reviews were the bulk of the attention.

These are linked: if claim becomes a cheap early lock (no estimate), you claim at
the *start* of engagement, and **the claim commit's git timestamp becomes a
precise, free window-start anchor** — no new frontmatter, design attention
captured. The scarce resource is operator attention (AI execution is flat-
subsidized, immaterial now); this makes `sdlc actual` measure it honestly.
