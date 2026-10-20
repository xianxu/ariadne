---
id: 000483
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0541d20619c601abcd40c18b4d72de8ba9002883' # card fields mirrored from issue-cards; edit via sdlc
---

# Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case)

## Problem

pair#622 (10-19): the issue's code landed (PR 406), but its live acceptance stayed open and needed a little more
work. `sdlc change-code --issue 622` insists on the canonical branch, which already merged; `--name` and `--issue`
are mutually exclusive; so there is no supported way to continue an issue whose branch has landed. The worker
improvised a fresh branch outside the gates.

## Spec

Pick one, simplest first, and make the refusal say it:
1. **Refuse with the recovery:** "#N's branch already landed; close #N with what's verified and file a follow-up
   issue for the rest (`sdlc issue new … --deps N`)." This keeps one issue, one branch.
2. Only if (1) proves too rigid in practice: a supported continuation branch (`<canonical>-2`) that the gates treat
   as the same issue.

## Done when

- `change-code` and `start-plan` on an issue whose canonical branch is merged refuse with that next action
  (test), and `sdlc issue new --deps` is shown as the path.

## Plan

- [ ]

## Log

### 2026-10-19
