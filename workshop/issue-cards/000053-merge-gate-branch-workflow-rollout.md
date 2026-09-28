---
id: '000053'
status: done
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 1
actual_hours: 2
---

# Rollout conductor — merge-gate + branch-in-place workflow

## Problem

**Conductor only.** Sequences three existing tickets + a one-time cut-over so the
program doesn't get lost across sessions. Substantive work + actuals live in the
sub-tickets; this file is the portfolio view and the ordering.

Sub-tickets: **ariadne #52** (generic CI merge-check mechanism), **ariadne #51**
(in-place-branch replaces direct-on-main), **you-decide #4** (the review-gate).
