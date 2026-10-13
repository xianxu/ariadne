---
id: 000349
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '89e8a1af70d2799226ba503e61a35647c4565735' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:18:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Pin agent-specific review fixtures instead of inheriting session defaults

## Problem

During #337's full suite, three unchanged baseline tests failed under the
session's PAIR_AGENT=codex. TestCloseRecordsTheVerdictBeforeAPostscript and
TestCloseWithoutAVerdictKeepsBothRuns supply Claude JSONL but leave the review
agent implicit. TestBoundaryReviewFencedArgvAndStamp expects Claude argv while
also inheriting the session default. Production selection correctly chose Codex.

## Spec

Pin the intended agent in agent-specific fixtures; audit analogous assertions.
Keep production agent resolution and its explicit precedence tests unchanged.

## Done when

- The three named tests pass under PAIR_AGENT=codex and PAIR_AGENT=claude, including conflicting AGENT_CMD, without requiring a suite-wide override.
- Agent-specific fixtures explicitly declare their agent; production resolution precedence remains covered.

## Plan



## Log

### 2026-10-13
- 2026-10-13: unclaimed: Independent baseline test-fixture issue discovered during #337; left unassigned for another slot.

Filed independently from #337. The source and fixtures match origin/main at
19dafa47ec755f12dd5a22d3c83a769b76f65a81. All three pass with PAIR_AGENT=claude:
closestream pair 21.947s; fenced-argv test 4.558s. No fixture changes made here.
