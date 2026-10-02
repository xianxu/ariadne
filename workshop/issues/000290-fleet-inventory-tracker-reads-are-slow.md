---
id: 000290
status: open
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'ae2c762b0868fd8bb94e8ce5a4043c7dc6decb85' # card fields mirrored from issue-cards; edit via sdlc
---

# Fleet inventory tracker reads are slow

## Problem

`sdlc fleet inventory` went from ~5s to ~17s on this machine (46 worktrees, 27
repositories) after ariadne#288 added tracker claims. A `GIT_TRACE` of one run
(638 git processes, 18.7s) attributes 10.5s to 15 tracker reads, one per tracked
repository, run in sequence: each is an `ls-remote --refs` plus a `fetch
--quiet`, and each fetch triggers an automatic `git maintenance run`. Before
#288 only worktrees on an issue branch triggered a tracker read. Local git
(~13 commands per worktree, ~13ms each) is the other ~8s.

## Spec

Make the inventory's tracker reads cheap without changing what they report:

- run the per-repository tracker reads concurrently, with a bounded worker
  count, keeping one read per repository (the `repoRecords` cache) and
  deterministic output;
- drop network round trips the read does not need (the `ls-remote` before a
  fetch that already answers presence, if that holds for every state the
  presence check distinguishes);
- suppress automatic maintenance on this read-only path;
- bound each fetch in time (the read takes the command's context; a remote that
  does not answer degrades that repository's claims to stale/unknown instead of
  stalling the inventory).

Stale/unknown semantics, one-read-per-repository and the #288 contract are
unchanged. Prerequisite for ariadne#289, which adds dependency-clone rows.

## Done when

- `sdlc fleet inventory --json` on this machine runs in well under the pre-#288
  ~5s plus local work (target: under 9s here), measured before/after with
  `GIT_TRACE`, numbers in the Log.
- Concurrency is tested: N repositories are read with at most the bound in
  flight, each exactly once, output identical to a sequential run.
- An unresponsive remote yields that repository's rows `stale`/`unknown` within
  the per-fetch deadline, tested with a remote that hangs; other repositories
  are unaffected.
- No automatic maintenance runs during inventory (asserted by trace or by the
  command's git config).

## Plan

- [ ]

## Log

### 2026-10-02

Filed while designing ariadne#289 at the operator's direction ("it's already a
bit slow"); to land before #289 resumes. Trace: 5.35s `ls-remote`, 5.13s
`fetch`, rest local.
