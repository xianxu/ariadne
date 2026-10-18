---
id: 000431
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '944913a623eb4d6cd71bf2d97a6580b3c08004ac' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc local-check: republish a completed attempt whose GitHub status post failed

## Problem

A completed `sdlc local-check run --name suite` whose GitHub status POST fails (network error to api.github.com after the run) leaves the attempt record and log intact locally, but no `sdlc-local/suite` status on the commit. There is no verb that publishes only that saved result, so the slot must rerun the whole suite (12-15 min, and a machine suite slot) or hand-post the status with `gh api`, which bypasses the gate. Hit twice on 2026-10-17 within minutes: pair:7 (pair#621, attempt a1f1f773, integrated head 055cb82b) and pair:5 (pair#651, head 80063e6d).

## Spec

A publication-only retry: republish the saved result of a completed attempt for its exact commit, verifying that the attempt record, its log hash and its commit all match, and that the attempt finished with exit 0 (or publish the failure as a failure). Refuse when the record is missing, incomplete or for another commit. The run path should also retry the POST a few times before giving up, and its failure message should name the republish command.

## Done when

- A completed attempt whose POST failed can be published without rerunning, and only for its own commit and result.
- Tests cover: matching record publishes; tampered log hash, other commit, incomplete or missing record refuse.
- The POST failure message names the recovery command.

## Plan

- [ ]

## Log

### 2026-10-17
