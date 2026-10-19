---
id: 000465
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0cdcc3cf03205719b1e5fe6aa8c8ce7a0df4808b' # card fields mirrored from issue-cards; edit via sdlc
---

# dev-aliases script exceeds its per-script budget under suite load

## Problem

The recorded ariadne suite on head da8c62d2e111ef7fe86cc0c351e00d6512a5a87a failed only `scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`: the per-script 2m deadline fired after about 40 passing assertions. The failed attempt is 4c96305468ac871ff567625384059cc8, log SHA-256 prefix 7714ac8e8e11, saved at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/4c96305468ac871ff567625384059cc8.log`. All 12 cmd/sdlc shards and 1,688 tests passed. An unchanged exact-head targeted run of that subtest passed 82/82 assertions in 33.6 seconds. #461 expanded the script with real Git layouts and sentinel binary cases; that cost crosses #448's fixed two-minute per-script ceiling under fleet suite load. The #458 branch changes no construct script or runner code.

## Spec

Keep the per-script ceiling. Reduce the cost of `construct/scripts/test/dev-aliases.test.sh` itself through shared fixtures, fewer repeated Git initializations or independent parallel subtests, while preserving its assertions. Register the exact test as a known flake during the repair so a same-head targeted rerun can provide landing evidence.

## Done when

- A counted/profiling fixture identifies the expensive setup or repeated work; the script's wall time falls materially under the existing two-minute ceiling under representative fleet load without simply increasing a sleep or deadline.
- Its existing 82 assertions remain covered, and a recorded full suite plus repeated targeted runs pass.

## Plan

- [ ]

## Log

2026-10-19: Classified while closing ariadne#458. Targeted `go test ./scripts/test -run '^TestConstructScripts/plain/dev-aliases.test.sh$' -count=1 -v` passed in 33.779 seconds after the recorded full-suite timeout. The terminal failure was `scripts_test.go:70: .../construct/scripts/test/dev-aliases.test.sh did not finish within 2m0s`; the partial output had passed the prior build and failure-preservation assertions.
