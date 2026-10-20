---
id: 000481
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '695b9162170d7c0612c40b1bd599b52948d802e1' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain

## Problem

A fleet drain before a machine or Couch restart (pair counterpart: `couch --drain [--fast]`) needs the machine test
queue and the landing lines to stop admitting work, and workers need a rule for the drain notice. Operator item 15
(10-19), discussion: ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-15-fleet-drain.md`.

## Spec

- `sdlc queue pause [--reason TEXT] [--cancel-running]` / `sdlc queue resume` (operator shell or a :0 hub only, the
  same rule as grants): while paused, the machine test queue and every landing line admit nothing new; waiters stay
  queued and print the pause and its reason. Running suites finish, or, with `--cancel-running` (used by
  `couch --drain --fast`), are cancelled and recorded as interrupted, rerunnable on the same head, never as failures.
- A merge interrupted by the pause resumes on `sdlc merge --yes` after resume, with no new suite needed when its
  head is unchanged and the suite record exists.
- tl-worker rule: on a drain notice, graceful means finish the current atomic step, push, write a Log line saying
  where you stopped, report "drained" to the hub and stay idle; fast means stop now, write the Log line, report.

## Done when

- Tests: pause blocks admission in both the queue and the landing line; resume releases in FIFO order;
  `--cancel-running` records interrupted (not failed) and a same-head rerun is accepted.
- `tl-worker --help` carries the drain rule.

## Plan

- [ ]

## Log

### 2026-10-19
