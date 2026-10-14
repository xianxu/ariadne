---
id: 000372
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'ce92c37a17cd35d9cfb40763423617c770448bed' # card fields mirrored from issue-cards; edit via sdlc
---

# Boundary review cannot run live-browser tests (reviewer sandbox can't start Chrome's GPU process)

## Problem

The boundary reviewer's sandbox can't start Chrome (its GPU process fails), so changes whose evidence is a live-browser test can't be re-verified in review; pair#525 needed a dispute (10-13). ariadne#364 made such gaps answerable ("unverifiable" findings, a declared prepare step) but the reviewer still can't run the browser itself. Retro item (b).

## Spec

Decide and implement one of: (1) let the review fence permit a headless browser (narrow rules like #364's literal /bin/ps: the specific Chrome helper processes, a scratch profile dir, no network beyond localhost); or (2) a declared `review-checks` step the harness runs OUTSIDE the reviewer's sandbox on the exact head and hands the reviewer its recorded output. Measure first which one Chrome actually needs on macOS.

## Done when

- A pair change with a browser test gets that test run and its result in front of the reviewer without a dispute (live check on a pair branch).

## Plan

- [ ]

## Log

### 2026-10-14
