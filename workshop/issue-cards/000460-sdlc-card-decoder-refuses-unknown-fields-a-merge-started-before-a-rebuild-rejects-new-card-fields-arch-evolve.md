---
id: 000460
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE)

## Problem

ariadne#394 added `flake_tests` to the tracker card. sdlc's card decoder refuses unknown fields, so any sdlc process started before ariadne:0 was rebuilt rejects every card that carries the new field. On 2026-10-19 pair:3's long-running `sdlc merge` for pair#687 (started on the old binary) refused the card after the server-side merge and needed manual recovery; other merges in flight at the time are exposed the same way. This is exactly what ARCH-EVOLVE (ariadne#454, landed the same night) forbids: an older reader must tolerate and preserve unknown fields.
