---
id: 000409
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
---

# Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310

## Problem

ariadne#310 made the reviewer agent a static per-repo policy. The operator (2026-10-16): having different models
review is useful robustness; the policy should be dynamic, choosing the reviewer by which model family designed and
coded the change, and by condition. Evidence from 10-15/16: Codex-authored designs under safety requirements
over-engineered (pair#556/#584, #587, #571), while Claude's reviews and redesigns caught it; Codex reviews caught real
bugs in Claude-authored code too.
