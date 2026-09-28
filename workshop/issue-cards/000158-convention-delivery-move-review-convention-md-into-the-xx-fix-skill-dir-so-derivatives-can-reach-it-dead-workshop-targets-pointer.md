---
id: '000158'
status: done
created: 2026-07-01
updated: 2026-07-06
actual_hours: 0.15
---

# convention delivery: move review-convention.md into the xx-fix skill dir so derivatives can reach it (dead workshop/targets pointer)

## Problem

`AGENTS.base.md §1` pointed every repo at *"Full table in `workshop/targets/review-convention.md`"*, but that file lives **only in ariadne** — `workshop/` is scaffolded per-repo, not exported, and `workshop/targets/` isn't even created in derivatives. So the pointer is **dead in brain, metis, nous, …** (verified: the file is missing in all three), while the dead pointer itself *is* delivered to every derivative's composed `AGENTS.md`. The convention's canonical grammar was unreachable exactly where agents work on docs.
