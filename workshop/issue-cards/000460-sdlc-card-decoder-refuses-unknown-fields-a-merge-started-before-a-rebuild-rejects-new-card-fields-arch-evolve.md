---
id: 000460
status: done
created: 2026-10-19
updated: 2026-10-19
estimate_hours: 2.84
github_issue:
started: 2026-10-19T00:51:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.25
tracker:
    version: 1
    completion:
        token: close-c0257a616fec
        repository: github.com/xianxu/ariadne
        reviewed_head: f6124b5ef7985ae6e3f91eb378c396fee0652228
        evidence_commit: 1c9fc70fd1f8b04fa1c5d15cea2a233dbff1dda5
        landed_commit: 2faabc739387be72915d6b47a69944261a1053b4
---

# sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE)

## Problem

ariadne#394 added `flake_tests` to the tracker card. sdlc's card decoder refuses unknown fields, so any sdlc process started before ariadne:0 was rebuilt rejects every card that carries the new field. On 2026-10-19 pair:3's long-running `sdlc merge` for pair#687 (started on the old binary) refused the card after the server-side merge and needed manual recovery; other merges in flight at the time are exposed the same way. This is exactly what ARCH-EVOLVE (ariadne#454, landed the same night) forbids: an older reader must tolerate and preserve unknown fields.
