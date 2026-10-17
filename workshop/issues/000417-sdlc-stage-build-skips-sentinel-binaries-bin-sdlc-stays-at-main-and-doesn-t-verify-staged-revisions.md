---
id: 000417
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'c0ceb06f0fc5ced72609ea9e1ab040acfb973b76' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions

## Problem

## Spec

Found staging #416 on 2026-10-16. `sdlc stage --build` in ariadne:0 built every binary except `sdlc`. The generic `make build` scanner skips `cmd/sdlc` because of its `.skip-make-build` sentinel (#31: sdlc builds through `make sdlc-build`). So `bin/sdlc` stayed at main (0e32b0eb) while the stage reported 00558776. A live check of an sdlc change would have run the old binary. The TL rebuilt it by hand.

In ariadne, `stage --build` should build each binary through its own build target: for a sentinel-skipped binary, the target its sentinel names. It should then verify that every staged `bin/*` carries the staged `vcs.revision`, and refuse when one doesn't.

## Done when

- `sdlc stage --build` in an ariadne-style repo leaves `bin/sdlc` at the staged revision.
- The stage checks each built binary's `vcs.revision` and refuses, naming the binary, when one doesn't match; a test covers that refusal.

## Plan

- [ ]

## Log

### 2026-10-16
