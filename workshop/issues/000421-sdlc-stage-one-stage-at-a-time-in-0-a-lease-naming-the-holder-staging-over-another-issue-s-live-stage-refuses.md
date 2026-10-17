---
id: 000421
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'e79ef0f667edabf138acda28c1f21729ade8968d' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc stage: one stage at a time in :0, a lease naming the holder; staging over another issue's live stage refuses

## Problem

## Spec

2026-10-16: ariadne:0 (the checkout whose `bin/` every slot's sdlc runs) was staged with #418 for pair:2's live merge of pair#618. ariadne:5 then ran `sdlc stage --build` for #416 on top of it. `stage` replaced HEAD and rebuilt the bin tools without noticing that another issue's stage was live and in use. `bin/sdlc` only survived by accident, because #417's sentinel skip left it alone. The TL had been granting :0 to one slot at a time by message. A rule kept by messages fails.

What makes "one stage at a time" a single check is one resource held for the stage's lifetime: a stage lease in :0 naming the issue, branch, slot and time. Then:
- `sdlc stage` takes the lease, and refuses while another issue holds it, naming the holder and when it staged.
- Clearing the stage releases the lease. Re-staging the same issue refreshes it.
- A recovery verb, shown in the refusal, frees an abandoned lease with an explicit acknowledgement.

## Done when

- With :0 staged for issue A, staging issue B refuses, naming A, its slot and its time; a test covers it.
- Clearing A then staging B succeeds; re-staging A while A holds the lease succeeds.
- The refusal names the recovery verb for an abandoned lease.

## Plan

- [ ]

## Log

### 2026-10-16
