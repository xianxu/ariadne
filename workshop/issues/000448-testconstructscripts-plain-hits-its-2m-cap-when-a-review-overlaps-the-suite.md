---
id: 000448
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: 'ec2a10740da6bb83f87590f574deb4be39fd4f32' # card fields mirrored from issue-cards; edit via sdlc
---

# TestConstructScripts/plain hits its 2m cap when a review overlaps the suite

## Problem

`TestConstructScripts/plain` (`scripts/test/scripts_test.go`) runs all of `construct/scripts/test` under one `runnerDeadline` of 2m per environment. Under load the plain subtest hits that cap and fails, unrelated to the change being tested.

Evidence (ariadne:2, #441, 2026-10-18):
- The recorded suite on 208ed53a (`sdlc local-check run --name suite`) failed with `construct/scripts/test did not finish within 2m0s`. The last script reported ok was `bootstrap-transitive.test.sh`. The whole run took 14m7s against a usual 8-12m, because a `sdlc milestone-close` boundary review ran alongside it, which the #442 overlap rule permits. The failed log is kept at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/dd86d2996958c8e19e84dd8b9c49ee61.log`.
- The branch touched nothing under `construct/scripts` or `scripts/test` (only prose in `construct/local/fix/SKILL.md`).
- On the same tree, `go test ./scripts/test/ -run TestConstructScripts -count=1` alone passed in 230.9s for three environments, about 77s each against the 120s cap, so the headroom is only about 1.5x.
- The next two recorded suites (b7c033ab, 389adfdc), run with nothing overlapping, passed.

## Spec

## Done when

- `TestConstructScripts` does not fail on a machine running a review (or a second suite) alongside it, which the #442 rule and the 2-suite machine budget both allow. The fix is either a per-script deadline in place of one wall-clock cap for the whole directory, or a cap derived from measured headroom under that load. The measurement is recorded.
- If the fix is a rule (for example, no whole-directory wall-clock cap in a sharded suite), `pkg/golint` gains it, citing this issue.

## Plan

- [ ]

## Log

### 2026-10-18
