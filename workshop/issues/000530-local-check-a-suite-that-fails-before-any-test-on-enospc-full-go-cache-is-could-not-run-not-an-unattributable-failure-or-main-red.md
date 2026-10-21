---
id: 000530
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '16c35ec9bba4eb78d62e091177f40ccd99a42c12' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red

## Problem

## Spec

10-21 ~12:40 the shared Go build cache `/Volumes/AgentScratch/go-build` (182 GB) filled its volume (186 GB, 100 MiB
free). Every recorded local-check suite on the machine then failed before running a single test, writing the cache
(`no space left on device`, typically in pair's runtimebundle generation): pair#768 (train 238 baseline), #771, #774
(integration + diagnostic main), #775, ariadne#526, #528, among others. sdlc classified these as **unattributable
failures** on the exact head, not as the narrow "could not run" (infra) class it uses for TMPDIR problems. Slots now
expect the same-head unexplained-failure guard to resist a plain rerun once space is restored, and train 238's
baseline turned a main-red verdict out of it.

Fix:
1. Classify a suite that fails **before any test ran** with a recognised infrastructure error (ENOSPC/EDQUOT on the
   Go cache, TMPDIR or the build output; also missing toolchain) as **could-not-run**, same as the TMPDIR class: no
   failure evidence, no main-red verdict, no same-head guard; the message names the full volume and path.
2. A main-baseline diagnostic that could not run must not mark main red or hold/split a train.
3. A supported recovery for runs already recorded as unattributable under this signature (one explicit
   `local-check` reclassification by the owner or :0, logged), never a flake registration.

## Done when

- A test drives a fake ENOSPC in the Go cache before any test output and gets could-not-run, rerun allowed,
  no main-red; a real test failure after tests start is still a failure.
- The reclassification path exists and is logged; tests cover refusal when tests had already run.

## Plan

- [ ]

## Log

### 2026-10-21
