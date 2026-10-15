---
id: 000384
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '612c50a39261ec91117a2127bca8de940658aa0f' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc: fixed-then-ship re-close after a post-close fix can't measure actuals, and correct-actual refuses (ambiguous close generation)

## Problem

pair#528's fixed-then-ship re-close (2026-10-14 23:39 PDT, pair:2) refused: `sdlc close --issue 528
--fixed-then-ship …` could not compute actuals ("can't auto-compute (retained measurement code head differs from the
completion's reviewed generation) — fall back to a judgment estimate"), `sdlc actual --issue 528` returned the same
error with no number, and the supported repair `sdlc issue correct-actual --issue 528 --dry-run` refused with
"ambiguous calibration close generation; repair target evidence before correcting". The retained measurement pins
faa20561 while the completion's reviewed head is d0862f78 (a post-close fix changed the head). The worker was stuck
for ~2 hours with no supported path; the TL unblocked it with `--no-actual` and the reason in `--verified`.

## Spec

- A fixed-then-ship re-close after a post-close fix must either re-measure against the new reviewed head or carry
  the first close's adopted actual forward explicitly (with provenance saying so), never refuse with no path.
- `correct-actual` must resolve the "close generation" for this case, or its error must name the exact next action.
- The refusal text must not suggest a judgment estimate where the policy forbids hand-typed hours.

## Done when

- A test reproduces: whole-issue close, post-close fix commit, fixed-then-ship re-close; it succeeds with measured or
  carried-forward actuals and correct provenance.
- `correct-actual` on that state either repairs it or names the precise next step; tests cover both.

## Plan

- [ ]

## Log

### 2026-10-15
