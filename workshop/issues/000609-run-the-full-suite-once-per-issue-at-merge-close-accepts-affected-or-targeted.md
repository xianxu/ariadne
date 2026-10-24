---
id: 000609
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'b2636f22fd93a8c99a5727c63aa08cd7a7331726' # card fields mirrored from issue-cards; edit via sdlc
---

# Run the full suite once per issue, at merge; close accepts affected or targeted proof

## Problem

## Spec

**Run the full suite once per issue, at merge: let close accept affected or targeted proof (operator approved, 10-24).**
Today a clean issue runs the full suite at least twice: the branch pre-close and the merge's integrated suite on the
merged head, which is the authoritative record. When the merge will integrate main anyway, the branch's full pre-close
duplicates it. Measured overnight 10-23/24 this was roughly 30–40% of machine test time. Source: ops session.side/2026-10-24-09-test-inventory-across-lifecycle.md (operator approved 10-24).

Change: when the merge will run an integrated suite, close accepts the branch's affected selection (or targeted proof
where affected falls back) as pre-close evidence; the full suite runs once, at merge. Repos can opt out.

## Done when

- close accepts an affected (or targeted-proof) pre-close record when the merge's integrated suite is required;
  tested.
- merge still requires the integrated full-suite pass; a branch failure caught only at merge is reported plainly.
- Help and atlas describe the new evidence rule.

## Plan

- [ ]

## Log

### 2026-10-24
