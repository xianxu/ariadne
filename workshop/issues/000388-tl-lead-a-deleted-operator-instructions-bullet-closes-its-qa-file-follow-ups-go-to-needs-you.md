---
id: 000388
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '3ed9ab16c2d43dcf57093109c4e27530a2d7d957' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T11:11:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [x] State the rule in `cmd/tl-lead/help.md` beside the bullet-deletion line; phrase guards in `TestLeadHelpCoversPlaybook`.

## Log

### 2026-10-15
- Rule added to the Operator-instructions paragraph of `tl-lead --help`; two phrase guards (closes, never into the closed file). `go test ./cmd/tl-lead/` green.
