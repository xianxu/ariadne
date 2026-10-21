---
id: 000532
status: open
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
---

# Archive preflight misses noncanonical side bundles

## Problem

pair#768 landed in PR521 (https://github.com/xianxu/pair/pull/521; reported merge 2026-10-21 14:37:41 PDT), but post-merge SDLC cleanup refuses with `active or conflicting archive artifact workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`. The card is done and the issue and parent plan are archived, while this tracked side document remains active. The documented retry `sdlc merge --branch 000768-darwin-pty-final-output --yes` repeats the refusal. Reported `sdlc issue recovery reconcile --issue 768` finds no unfinished operations and deleted the remote branch without archiving this leftover. Successful landing therefore leaves an incomplete artifact bundle and no successful documented cleanup retry.
