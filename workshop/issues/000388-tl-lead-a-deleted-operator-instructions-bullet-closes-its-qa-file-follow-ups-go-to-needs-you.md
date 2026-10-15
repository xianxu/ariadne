---
id: 000388
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '9a3245e3b7585ff4d641957e03de842627829fc2' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: a deleted Operator-instructions bullet closes its QA file (follow-ups go to Needs you)

## Problem

The operator added a TL rule on 2026-10-15 that tl-lead doesn't have yet (it sits in ops `skills-draft/xx-tl/SKILL.md`
under "Draft rules not yet lifted"): **a deleted Operator-instructions bullet closes its QA file.** The operator wants
TL rules changed in tl-lead directly, now that it lives in ariadne.

## Spec

Add to tl-lead's Operator-instructions / QA-file rules: when the operator deletes a bullet from `## Operator
instructions`, they have read and acknowledged its answer, so that QA file is **closed**. Later follow-ups or questions
on that topic go into `## Needs you` (or into a new bullet's answer), never as an edit or new question inside the
closed QA file, which the operator won't reopen. Add a phrase guard in tl-lead's help tests, like the #377 rules.

## Done when

- `tl-lead --help` states the rule; its phrase-guard test passes.
- The TL removes the rule from ops' draft after this lands.

## Plan

- [ ]

## Log

### 2026-10-15
