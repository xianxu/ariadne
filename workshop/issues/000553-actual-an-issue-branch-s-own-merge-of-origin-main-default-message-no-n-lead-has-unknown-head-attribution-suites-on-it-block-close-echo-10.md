---
id: 000553
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '550f9d063772567e1f83e4b2f2a65896f7bca016' # card fields mirrored from issue-cards; edit via sdlc
---

# actual: an issue branch's own merge of origin/main (default message, no #N lead) has unknown head attribution; suites on it block close (echo#10)

## Problem

## Spec

From ops-echo:0 (10-22 ~06:4x): echo#10's close refuses at measurement on both the fleet sdlc (with #548) and the
#551 binary 26b4be0, before any review:

    Measured actual unavailable for issue close: can't auto-compute (queue evidence incomplete; ...: suite attempt
    dffca4b701d5264e244d6f019055b21d SHA 885dc171dcfde7aaaae15d79f2b7c3863f207331 has unknown issue-head
    attribution: commit has no definite issue lead or recognized neutral authority; ...)

885dc171 is echo:3's ordinary integrate-main merge on the echo#10 branch: git's default message "Merge
remote-tracking branch 'origin/main' into 000010-...", with no `#10:` lead, and its recorded suite is green. Before that
suite ran, the same close measured 1.63h. #548 ("suite evidence must belong to the measured issue head") follows
first-parent neutral integration only when it's recognized, and a plain `git merge origin/main` on the issue branch
isn't. Every landing integrates main, so echo#3 and #5 (and any issue that merges main before its final suite) hit
this.

Fix (structural, no commit-message heuristics): a merge commit on the issue branch whose **first parent** is
attributable to the issue (issue-led, or itself attributable through this same rule) and whose other parents are all
reachable from the pinned origin/main is the issue's own integration, so it's attributed to the issue. Unknown or
foreign parents keep strict refusal.

## Done when

- echo#10's exact shape (issue commits, then a default-message merge of origin/main, then a suite on the merge)
  measures, with the suite attributed to the issue (real-git fixture).
- A merge bringing in a foreign issue branch (a second parent not on main) still refuses (test).
- echo#10's live 'sdlc actual --issue 10' measures (owner-run proof).

## Plan

- [ ]

## Log

### 2026-10-22
