---
id: 000425
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f80593cdfac36565b5c66a6e8c5add334b4eacd7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T06:17:40-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Landing suites invalidated by main moving during the 15–35 min queue; the TL held main by hand twice

## Problem

## Spec

2026-10-16/17: twice the TL had to hold every other ariadne landing by hand. First for #417, whose integration retry was about to run out, then for #423, which ran its full suite three times because main moved under it. Each recorded suite waits 15–35 minutes in the machine test queue and runs 8–11 minutes. A landing that integrates main, then queues its suite, is therefore invalidated whenever another landing reaches main during that wait. With five ariadne slots landing in one night this happens repeatedly, and `sdlc merge` gives up after three integration attempts.

The landing line already serializes merges. The gap is that a landing's suite runs against a main that can still move.

Options to evaluate:
1. **The line holds main through the suite:** the line-holder's integrated head is the next thing allowed onto main. Other merges queue behind it, and nobody else's merge advances main while its suite runs. This is already true for `sdlc merge` landings; check the issue-publish path and any other main writers (`issue new`, publish, project ticks).
2. **Docs-only drift doesn't invalidate:** if main moved only in issue or plan prose (`workshop/`), integrate without rerunning the suite. The #417 landing already did this once.
3. **Measure first:** how many reruns over the last N landings came from code drift vs docs-only drift.

## Done when

- The measurement (option 3) is recorded, split by cause.
- With the chosen fix, a landing whose suite is running is not invalidated by another slot's landing; a test simulates two landings.
- Docs-only drift during a landing's suite does not force a rerun, if that option is chosen and justified by the measurement.

## Plan

- [ ]

## Log

### 2026-10-17

- Measurement scope: last 10 successful local landing-line events (PRs 272,273,275,276,277,278,274,280,282,281), 25 completed recorded suites, from 62 retained attempt records. No unreadable attempt records or missing Git objects. Group attempts by nearest issue-prefixed first-parent commit; pin each issue's main comparison to its PR merge's first parent (today's main would incorrectly contain the issue's own commits). Compare successive attempts, identifying main-base advancement separately from branch code edits; prior failures are a separate cause, and mixed branch/main changes are not attributed solely to main. Classifier matches publishGateHasCodeSurface: tests/config/embedded cmd markdown count as code; ordinary documentation/process paths do not.

| Issue | Completed suites | Integration-only code drift repeats | Docs-only drift repeats | Branch changes/mixed | Previous failure |
|---|---:|---:|---:|---:|---:|
| #413 | 1 | 0 | 0 | 0 | 0 |
| #414 | 1 | 0 | 0 | 0 | 0 |
| #415 | 2 | 0 | 0 | 1 | 0 |
| #418 | 3 | 1 | 0 | 0 | 1 |
| #417 | 3 | 1 | 0 | 1 | 0 |
| #419 | 1 | 0 | 0 | 0 | 0 |
| #416 | 5 | 1 | 0 | 2 | 1 |
| #421 | 3 | 0 | 0 | 1 | 1 |
| #422 | 1 | 0 | 0 | 0 | 0 |
| #423 | 5 | 2 | 0 | 2 | 0 |

- Totals: 15 repeat runs = 5 integration-only code drift, 0 docs-only drift, 7 branch-change/mixed, 3 after failures. The 5 integration-only repeats consumed 46.197 minutes execution plus 231.472 minutes queue waiting. These costs describe observed repeats, not guaranteed recoverable savings. In all 5, the preceding passed suite was a pre-close local-check run, outside landing admission. Newly incorporated main code landed during the prior queue in 3 cases (#418, #416, first #423), during execution in 1 (second #423), and after the suite in 1 (#417); some bases were already stale before queueing as well. Seven sampled suites ran after landing admission; none has a demonstrated docs-only integration rerun in this sample.
- #423 specifically had five passing suites: two repeated solely for main code (#421 then #422), and two for BR-1/BR-2 branch fixes. Its two integration-only repeats account for 18.095 execution minutes and 56.038 queue minutes. The assertion that all five were invalidated by main would be incorrect.
- Existing boundaries: local-check never takes a landing turn (cmd/sdlc/localcheck_cmd.go:76–144); merge holds the turn from admission through its suite, server merge and archive (cmd/sdlc/landing.go:380–644). Existing first-parent ancestor reuse already accepts docs-only cumulative differences through publishGateHasCodeSurface (cmd/sdlc/cigate.go:480–535). Thus extending only main-publication exclusion inside merge does not protect the pre-close suites that produced the measured repeats.
