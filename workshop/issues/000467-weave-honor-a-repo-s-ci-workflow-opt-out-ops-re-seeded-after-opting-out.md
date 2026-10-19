---
id: 000467
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5a3cb5ecc544dd9e1a2be4e601120238d211fcca' # card fields mirrored from issue-cards; edit via sdlc
---

# weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out)

## Problem

ops' AGENTS.local.md says "ops opts out of ariadne's seeded CI workflows", yet a weave refresh in ops modified the tracked .github/workflows/merge-check.yml and added .github/workflows/local-gates-guard.yml. On 2026-10-19 the operator removed ops' CI workflow; the next weave refresh would seed it again.

## Spec

Give the opt-out a machine-readable form (a repo-level weave setting, e.g. in the repo's weave configuration: seeded CI workflows off) and make weave honor it: no seeding, modifying or re-adding of .github/workflows/* in an opted-out repo, while other seeded files (skills, settings) still refresh. weave reports "CI workflows: opted out" so the choice is visible. Migrate ops to the setting, and drop the prose-only claim or point it at the setting.

## Done when

- weave refresh in an opted-out fixture repo leaves .github/workflows untouched (and absent stays absent); a normal repo still gets them.
- ops carries the setting, and a refresh there leaves its workflows absent.

## Plan

- [ ]

## Log

### 2026-10-19
