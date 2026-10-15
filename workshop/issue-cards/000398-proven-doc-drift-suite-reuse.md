---
id: 000398
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# Proven suite reuse after doc-only main drift

## Problem

A main movement during `sdlc merge` normally requires re-integration and a fresh full suite even when only non-test prose moved. #390 treats a doc-only reuse path as a later optimization, not a reason to weaken the exact-head landing guarantee without proof.
