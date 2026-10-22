---
id: 000558
status: open
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
github_issue:
---

# Distribute the typed flake Plenary producer through the base manifest with operator verification

## Problem

ariadne#556 adds the typed script/Lua flake implementation and the Plenary producer source `scripts/sdlc-plenary-events.lua`. Peer adoption needs that producer distributed through the base layer. The required `construct/base.manifest` row is protected by `local-gates-guard`, so ops:0 explicitly split it from #556 into this operator-verified issue (Couch delivery `8556cb2f-b354-48e4-ad56-4a99963da64e`).
