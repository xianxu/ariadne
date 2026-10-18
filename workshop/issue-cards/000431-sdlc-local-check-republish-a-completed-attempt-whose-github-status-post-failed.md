---
id: 000431
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
---

# sdlc local-check: republish a completed attempt whose GitHub status post failed

## Problem

A completed `sdlc local-check run --name suite` whose GitHub status POST fails (network error to api.github.com after the run) leaves the attempt record and log intact locally, but no `sdlc-local/suite` status on the commit. There is no verb that publishes only that saved result, so the slot must rerun the whole suite (12-15 min, and a machine suite slot) or hand-post the status with `gh api`, which bypasses the gate. Hit twice on 2026-10-17 within minutes: pair:7 (pair#621, attempt a1f1f773, integrated head 055cb82b) and pair:5 (pair#651, head 80063e6d).
