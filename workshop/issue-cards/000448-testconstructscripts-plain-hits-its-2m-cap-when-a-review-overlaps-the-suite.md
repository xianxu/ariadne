---
id: 000448
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
severity: low
started: 2026-10-18T19:54:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# TestConstructScripts/plain hits its 2m cap when a review overlaps the suite

## Problem

`TestConstructScripts/plain` (`scripts/test/scripts_test.go`) runs all of `construct/scripts/test` under one `runnerDeadline` of 2m per environment. Under load the plain subtest hits that cap and fails, unrelated to the change being tested.

Evidence (ariadne:2, #441, 2026-10-18):
- The recorded suite on 208ed53a (`sdlc local-check run --name suite`) failed with `construct/scripts/test did not finish within 2m0s`. The last script reported ok was `bootstrap-transitive.test.sh`. The whole run took 14m7s against a usual 8-12m, because a `sdlc milestone-close` boundary review ran alongside it, which the #442 overlap rule permits. The failed log is kept at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/dd86d2996958c8e19e84dd8b9c49ee61.log`.
- The branch touched nothing under `construct/scripts` or `scripts/test` (only prose in `construct/local/fix/SKILL.md`).
- On the same tree, `go test ./scripts/test/ -run TestConstructScripts -count=1` alone passed in 230.9s for three environments, about 77s each against the 120s cap, so the headroom is only about 1.5x.
- The next two recorded suites (b7c033ab, 389adfdc), run with nothing overlapping, passed.
