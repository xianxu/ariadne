---
id: '000137'
status: done
started: 2026-06-29T20:56:16-07:00
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.53
actual_hours: 0.27
---

# sdlc boundary review repo orientation

## Problem

Boundary review prompts can misorient a fresh reviewer about which repository is
under review. During pair#81 retro, review context for pair work was observed to
refer to an `ariadne#...` issue shape, even though the operating repository was
`pair`.

This is risky because the boundary reviewer is intentionally fresh-context. If
the prompt names or implies the wrong repo, the reviewer can inspect the wrong
tracker, apply ariadne base-repo assumptions to a downstream repo, or report
findings against the wrong issue surface.
