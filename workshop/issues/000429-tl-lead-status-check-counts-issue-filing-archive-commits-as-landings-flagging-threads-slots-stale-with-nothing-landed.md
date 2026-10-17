---
id: 000429
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3a97ea3b66cace39b614fbd5477d416c5dbde65c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T12:28:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ceedcb63", done: "71e20541"}
---

# tl-lead status check counts issue-filing/archive commits as landings, flagging Threads/Slots stale with nothing landed

## Problem

## Spec

`tl-lead status check` marks Threads and Slots stale when they're older than "the fleet's newest landing". It treats any commit on a fleet repo's origin/main as a landing. That includes `#N: issue: new` filings, issue publishes, "archive completed issues to history" and project ticks. On 2026-10-17 the check and pair#630's reminder flagged the session file stale several times an hour after issue filings alone, for example parley.nvim 12:17 "#378: issue: new" and pair 11:46 "#652: issue: new". That makes the hub re-stamp sections with nothing to report and dulls the signal.

A landing is a merged PR: a merge commit from `sdlc merge`, or a landing ledger entry. Prefer the landing ledger or merge-commit shape over commit-message heuristics if one is available.

## Done when

- Issue-filing, publish, archive and tick commits don't count as landings. A merged PR does. Tests cover both.
- The stale message names the landing it compared against: repo, PR or merge commit, and time.

## Plan

- [x] Reproduce direct-main false landings with real Git; cover first-parent merges and identity in stale output.
- [x] Select the latest first-parent merge, render its hash, and synchronize help/atlas.
- [x] Run focused checks and the recorded suite, close, and land on SHIP.

## Log

### 2026-10-17

- Root cause: gitNewestLanding reads the origin/main tip timestamp without filtering. Use first-parent merge shape, which sdlc merge emits via gh pr merge --merge; no subject heuristics. The landing-line ledger is local, has 30-day retention, and records completion rather than the immutable merge commit. Existing Git seam and stale renderer remain the owners (ARCH-DRY, ARCH-PURE). No new durable state or lifecycle.

- RED: TestGitNewestLanding selected the direct-only repo at 9003 after filing/publish/archive/tick commits. GREEN: go test ./cmd/tl-lead -count=1 passed (2.690s), including no-merge/message-spoof controls, nested merge with newer timestamp, newest fleet selection, skipped brain/non-SDLC/no-origin refs, command-rendered merge identity/time and help contracts. git diff --check passed. Requested a short :0 grant for the read-only native check; no :0 access before grant.

- PR #287 opened with a behavior-focused description. Recorded suite launched on pushed 83c6878d (attempt ef3ae307b3f44247e87c0c956dc32cd1), initially position 5 in the weighted machine queue; no custom queue or main hold. Native stage grant request receipt 15b8627e-54eb-4eaa-be44-ab2451613bb3 queued, waiting for the hub composer.

- Native acceptance (TL grant delivery 548a7fe0): staged d84f950421459df3f6181bc9864d2e4aa32af62b on :0, seven binaries verified at that revision with modified=false. Read-only native tl-lead status check on ops/teams/devinfra/tl-status-xian-xu.md returned fresh: every budgeted section; root and check help phrase contracts passed. Status-file SHA256 before/after identical (9f898ead261f5387bcbf257c639dbbb2fb2beb59e1c5e851ccce880ed64584c3). Stage cleared; all seven binaries rebuilt/verified on clean main 0c268805c1c547b42a45e0792c5493e5d0a79c4d; reported free. UTC targeted landing/freshness tests also passed (3.385s); server CI green.

- Recorded pre-close suite PASS on 83c6878dd3ea5b258004319d06822bb041236f97, attempt ef3ae307b3f44247e87c0c956dc32cd1: 12 shards over 1572 SDLC tests plus 67 other packages, all exit 0; shard wall 513.4s, recorded exit 0 in 8m34s, log SHA256 prefix e2f6e294f55d. Saved log: .git/sdlc/local-checks/83c6878dd3ea5b258004319d06822bb041236f97/suite.log. No code changed after that pushed head; later commits record evidence only. Proceeding to close and the owned landing line without manually chasing newer main.
