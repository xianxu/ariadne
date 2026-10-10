---
id: 000317
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '0ae42c5431df6cd03b06f32d3085b4828b32aaf7' # card fields mirrored from issue-cards; edit via sdlc
---

# active-time: every ref in a commit subject claims an equal share, so citations and merge subjects take the issue's time

## Problem

`sdlc actual` gives every local `#N` in a commit subject an equal share of the
activity run that commit claims (`activetime.attributeRun`). A subject that
*cites* another issue therefore hands that issue a share of the measured
issue's time. Two cases, both found reproducing ariadne:4's report that #304
measured 2.23 h before merging main and 1.01 h after (wall clock about 2.4 h,
claim 20:01 → close 22:26).

1. **Integration merge (regression from #270).** #270 drops merge commits
   from a branch's own boundary set, but keeps any commit whose subject names
   the measured issue. The merge `#304: merge origin/main (#300, #270
   landed)` names #304, so it stays a boundary. Its run is then split three
   ways. Replay at the post-merge head, scoped as `sdlc actual` does:
   #304 60.8 m, #300 44.6 m, #270 44.6 m, so **1.01 h**. Without that one
   boundary it is 1.38 h, and the pre-merge head gives 1.37 h. The number
   moves on integration, which #270 exists to prevent.
2. **Citations (pre-existing, v3 rule).** `#304: log: --no-validate rationale
   for the landing (F1, #308)` cites #308, which then takes **66 min** of
   #304's time at both heads. The window's total active time is 2.48 h, and
   all of it was #304 work in slot :4, but #304 measures 1.37 h.

Replay method: a local clone at each head; Compute with the slot-4 and brain
transcript dirs, tracker ref, claim window 20:01:24 → the head's last #304
commit, and branch points b77e9718 (pre) / 3fddafdf (post, main before PR 174).

## Spec

- Only a subject's **leading** issue ref(s) claim time. That is the `#N:` /
  `#N Mx:` lead (and the `ariadne#N` self-qualified form), and any refs the
  convention puts at the lead (`close #N`, `#N, #M:`). Refs later in the
  subject are citations. They stay in the mention scope but are not
  claimants. One extractor in `issueref`, shared by boundaries and peers
  (ARCH-DRY with #254's merge-PR mask).
- A branch's own integration merge is never a claimant beyond its lead.
  Leading-only covers it (`#304: merge origin/main (…)` claims for #304 only).
  Decide whether such merges should be boundaries at all; with leading-only
  they at least cannot leak.
- Replay test from real subject shapes: the #304 merge subject, a
  parenthesised citation, `#174-#176`, `chore: bump (refs #1, #2)`, and a
  subject with no lead (it stays a neutral boundary, or claims its refs; pick
  one and record why).

## Done when

- The #304 replay measures the same before and after its main merge, and #308
  and #270/#300 get nothing from #304's subjects.
- A table test pins lead vs citation extraction over the shapes above.
- Re-measure two or three recent closes before/after and record the deltas
  for calibration consumers.

## Plan

- [ ]

## Log

### 2026-10-09

- Filed by ariadne:3 at the ariadne:1 TL's request, investigating ariadne:4's
  #304 hours report. Not fixed yet (TL: file, don't fix).
