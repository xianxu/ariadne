---
id: 000554
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0d980b1c65e0804833165366101b7774694c80de' # card fields mirrored from issue-cards; edit via sdlc
---

# actual: a legacy-claimed issue with post-cutover commits still reports 'no post-cutover owned interval' (echo#5)

## Problem

## Spec

From ops-echo:0 (10-22 ~11:00). echo#5's close refuses at measurement on fleet sdlc 1c0595d (and in a read-only preview
at pushed head 350a62d): "can't auto-compute (no post-cutover owned interval yet: measurement has no verified owned
interval before its pinned right edge; reconcile and record post-cutover work before closing)". But echo#5 has
post-cutover work: tracker cutover 10-21 20:25:53; #5 commits 20:34:31, 20:35:14 and 21:00:04; a merge of main
10-22 09:17:50; "#5: merge origin/main (echo#3 contracts); add VocabFromProto" 10:59:11. Its code-complete hold was
recorded around 21:00 on 10-21, which may be the pinned right edge, so the post-cutover owned interval is computed
as empty or isn't recognised. ariadne#544 deliberately refused echo#4 and #5 while they had no post-cutover work; now
the work exists and it still refuses.

Likely causes to check: the owned tracker interval after cutover never opened for an issue claimed pre-cutover and
reconciled (the reconcile event isn't treated as an engagement start); or the right edge is pinned at the
pre-existing hold and excludes later work.

## Done when

- A legacy-claimed issue reconciled after cutover, with commits after cutover, measures its post-cutover owned
  interval (real-git fixture of echo#5's shape).
- echo#5's 'sdlc actual --issue 5' measures (owner-run live proof by ops-echo).

## Plan

- [ ]

## Log

### 2026-10-22
