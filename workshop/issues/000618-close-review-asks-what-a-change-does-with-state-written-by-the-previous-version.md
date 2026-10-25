---
id: 000618
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '48c98c7f13662b26136825af1275fb451131dc3e' # card fields mirrored from issue-cards; edit via sdlc
---

# Close review asks what a change does with state written by the previous version

## Problem

## Spec

**Close reviews ask about state created by the previous version (operator request, 10-24).** On 10-24 a new invariant
on existing data (pair#866's readiness marker on live Zellij sessions) shipped through review without anyone asking
what happens to sessions created before it; it broke the operator's warm start. The close-review prompt should ask, for
any new invariant, schema, marker, file format or registry rule over data that already exists: what does the change do
with state written by the previous version, and which test proves it (pair's upgrade fixture where it applies)?

## Done when

- The boundary-review prompt includes the previous-version-state question for changes that add or tighten invariants
  on existing data; a finding is raised when no upgrade path or test is shown.
- A review fixture with a pair#866-shaped change produces the finding.

## Plan

- [ ]

## Log

### 2026-10-24
