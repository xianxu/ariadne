---
id: 000594
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
---

# Actual attribution excludes witnessed local staging intervals

## Problem

After #591, pair:6 cannot close pair#830 with numeric actuals because an explicitly witnessed local staging branch is treated as unknown/detached/ambiguous during concurrent claims.
