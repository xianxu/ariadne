---
id: 000335
status: open
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-11
estimate_hours:
card_mirror: 'edb441d4f08c00d2d6c73b57c34ae82eb4d29629' # card fields mirrored from issue-cards; edit via sdlc
---

# actual: close adopted 0.05h for an issue sdlc actual measures at 2.57h (branch moved to :0 for a smoke test)

## Problem

pair#454 closed with `Close-Actual: 0.05` (commit 3d8523e9) against an estimate of 4.69h. Right after the merge, `sdlc actual --issue 454` from the same slot (pair:1) measured 2.57h (window 2026-10-11 20:08 → HEAD; build 2.57h; one dominant attribution segment 20:08 → 22:22). The close-time measurement was wrong by 50×, and close adopts its measurement when `--actual` is omitted, so the velocity record now holds 0.05h. Context: before close the branch was moved to pair:0 (`sdlc move :0`) for the operator's smoke test and the work after that ran there; close ran in pair:1 afterwards. The slot reported the ratio as likely a measurement-window artifact.

## Spec

Find why close's measurement differs from `sdlc actual` for the same issue minutes later: the window close used (base, HEAD, which checkout), the session transcripts it attributed (pair:1's vs pair:0's), and whether `sdlc move` breaks the attribution. Fix at the root so close and `sdlc actual` agree, including across a move to :0 and back. Add a refusal or a loud warning when close's measurement is implausible against the estimate (e.g. under a tenth) instead of adopting it silently. Correct pair#454's recorded actual (2.57h, or what the fixed measurement says) through the verb that owns it.

## Done when

- The cause is reproduced in a test (a branch moved to :0 and back, then closed).
- Close and `sdlc actual` measure the same window and agree.
- An implausibly low measurement against the estimate warns or refuses at close.
- pair#454's actual is corrected.

## Plan

- [ ]

## Log

### 2026-10-11
