---
id: 000255
status: done
created: 2026-09-27
updated: 2026-09-28
estimate_hours:
github_issue:
started: 2026-09-27T23:21:50-07:00
actual_hours: N/A
tracker:
    version: 1
    completion:
        token: close-18cdf54d0f9c
        repository: github.com/xianxu/ariadne
        reviewed_head: 3ce173c5d70da69010df6304187bc43085945252
        evidence_commit: 3092bf0c5293259da260b71afa1a4d4a0c8c97b0
        landed_commit: a82d29c5b0d1164a15c3b64b5788aff909594f34
---

# Fleet cutover to the issue tracker; delete the legacy writers

## Problem

#252 built the issue tracker (cards on `issue-tracker`, details with
`card_mirror`), the one-time migration (`sdlc issue migrate`), the cutover
guard, and a legacy mode that runs the pre-#252 workflow wherever a repository
has not cut over. It proved them through smoke tests (A1, A2), a legacy-mode
soak and a canary cutover of parley.nvim (A3, A4).

What #252 does not do is move the rest of the fleet. Its "existing issue files
are migrated" criterion moves here. The legacy writers, which the binary keeps
only for repositories not yet cut over, are deleted once every repository has
moved.
