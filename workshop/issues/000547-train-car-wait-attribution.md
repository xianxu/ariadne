---
id: 000547
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd33a4ce682c42a4d7f621cf331881ae1b1fb94c8' # card fields mirrored from issue-cards; edit via sdlc
---

# Train car active-time counts conductor-owned queue wait and status polls as focused work

## Problem

pair#778 renewed close on 2026-10-21 adopted 4.41h and identified a 213.9-minute/80% dominant segment from20:18 to23:52. That segment includes the retained car waiter PID51323 from20:37 to23:44 (about187minutes) while train261 was owned by pair:2 and waiting/running its combined suite, main baseline and prefix suite. pair:5 only polled/reported retained state during that interval; branch changes resumed after its waiter exited. The timing report excluded66.57minutes across2 earlier own-suite waits, but did not exclude this conductor-owned waiting. Frequent status polling appears as engagement. Do not treat the full213.9m as pure wait: it also contains work before/after the retained waiter.

## Spec

Investigate active-time-v3 engagement/queue attribution for a train car waiting on another workspace conductor. Use proven car membership, retained landing ticket and conductor-owned suite/baseline/prefix wait evidence to distinguish waiting from focused work. Preserve actual code/review attention and fail conservatively for unproved identity/time intervals; no guessed manual replacement hours. Determine whether existing lease/car evidence already provides the necessary binding.

## Done when

- Reproduce the train-car attribution with a controlled timeline containing real issue work, an owned car waiter, another slot conductor and periodic read-only status polls.
- Proven wait-only intervals do not become focused hours merely because of polls, including main-baseline and split-prefix phases; actual work around the wait remains attributable.
- Ordinary solo suite waits, concurrent unrelated work, absent/unbound car identity and interrupted/ended waiters retain explicit conservative behavior.
- Record the correction or limitation against pair#778 evidence; regression proves waiting N versus2N does not grow focused work from polls alone.

## Plan

- [ ]

## Log

Filed unclaimed by pair:5 while its corrected pair#778 suite/review run. No active-time implementation or recorded actual was hand-edited. Evidence in pair workshop/evidence/000778/train261-compile-failure.* and the issue Log pins train, attempts and terminal state. Renewed close report on source b9ea92045914dc1017e4e59ab473c7c73c015fe8 printed4.41h, dominant213.9m segment, and only66.57m excluded suite queue. Earlier close before landing wait adopted1.11h. Both are tool measurements, not replacement recommendations.
