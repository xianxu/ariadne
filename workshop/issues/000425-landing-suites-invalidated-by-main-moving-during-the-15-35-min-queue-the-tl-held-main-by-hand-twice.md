---
id: 000425
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3e835ba942489b9bffdd63d9b2742628230c1b0b' # card fields mirrored from issue-cards; edit via sdlc
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
