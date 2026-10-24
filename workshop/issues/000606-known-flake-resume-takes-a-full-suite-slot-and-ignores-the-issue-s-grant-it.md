---
id: 000606
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '0b97ccef5e65d6754b9d101b8b71e6ca2b88c980' # card fields mirrored from issue-cards; edit via sdlc
---

# Known-flake resume takes a full suite slot and ignores the issue's grant; it should use the light lane

## Problem

## Spec

**A known-flake resume that reruns only registered tests takes a full suite slot at the back of the queue and ignores
the issue's landing grant (TL observation, 10-24).** pair#873's resume (parent `9b1291df`, child `82d1856f`, lease
`138335625`) reruns one registered test (seconds of work), yet it queued as an ordinary full-weight suite ticket
(ticket 40, position 10, `granted=false`) behind ariadne's full suites, while the light lane was free with nobody
waiting, and while pair#873 held an active landing grant.

Fix:
1. A resume whose work is only registered-test reruns runs in the light lane (or takes no suite slot), since its cost
   is a handful of tests.
2. A resume ticket carries its issue, so the issue's grant applies to it like any of its other suite tickets.

## Done when

- A registered-tests-only resume is admitted through the light lane when it's free; a test pins it.
- A resume ticket for a granted issue shows `granted=true` and orders accordingly; a test pins it.

## Plan

- [ ]

## Log

### 2026-10-24
