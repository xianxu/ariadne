---
id: 000527
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b1cfd65bd5d2d84389ad5ced1004d14160180c6d' # card fields mirrored from issue-cards; edit via sdlc
---

# landing split: don't hold a car for a registered flake or a failure in a package it doesn't touch

## Problem

## Spec

Pair train 234 (10-21): the combined-head suite failed only `broadcast.TestHubRandomInterleavings` (/30). The test
was then registered as a known flake (pair#771) after same-head targeted reruns passed. The train's automatic split
still held car 2 (#768) at 08:05:30 "based on the earlier coarsened broadcast failure", even though broadcast files are
identical across the prefix and the failing head and the parent test passed on the same head. #768 (scribe-only) lost
its landing to an unrelated, registered flake.

When a split or bisection assigns blame for a failure, it should re-classify that failure against the **current**
flake registry, and check whether the failing package's sources differ between the passing prefix and the failing
head. A failure in a registered flake, or in a package the suspect car didn't touch, must not by itself hold a car;
report "unattributed: registered flake / untouched package" and continue (or rerun just that test on the suspect
prefix).

## Done when

- A train whose only failure is a test registered as flaky (registered before or during the split) doesn't hold a car
  for it; a regression reproduces train 234's sequence.
- A failure in a package the blamed car doesn't change is reported as unattributed rather than pinned on the car.

## Plan

- [ ]

## Log

### 2026-10-21
