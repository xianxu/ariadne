---
id: 000474
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# scratchdisk: accept TestScratch only if APFS (HFS+ broke pair suites: whole-second mtimes, no clonefile)

## Problem

2026-10-19: with ariadne#468, recorded suites run on `/Volumes/TestScratch`, which the ops login item had formatted HFS+.
HFS+ keeps whole-second timestamps and has no clonefile, so pair's suites failed on every landing: same-size rewrites with a
restored mtime went undetected (pair:6 A/B on one head: passes with TMPDIR=/private/tmp, fails on TestScratch), pid-mtime
bounds failed, and clone-based write budgets blew up (couchcore 1.9 GB vs a 128 MB cap). pair:7 measured 20/20 undetected
rewrites and parent churn on TestScratch vs 0/20 on /private/tmp. The ops script now formats APFS (probed: ns timestamps,
clonefile work); but sdlc accepted the HFS+ volume silently.
