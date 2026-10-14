---
id: 000366
status: codecomplete
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'b027496dd84ea31c0a74363c7a92f2c2d39ee7e7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T09:00:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.89
---

# sdlc local-check: a slot whose ariadne substrate predates the weighted lock fails with a raw usage error; detect the version skew and name 'weave refresh'

## Problem

After ariadne#332 landed (2026-10-14), the shared `sdlc` (`~/workspace/ariadne/bin/sdlc`) passes `--slots N` to the slot's `scripts/test-lock.py`, which comes from the slot's own ariadne substrate (`worktree/<repo>-slotN/ariadne`). Slots whose substrate predates ariadne#346 (pair:2, :4, :5, parley.nvim:2, :3 on 10-14) fail `sdlc local-check run --name suite` with a raw argparse usage error. pair:5 hit it re-closing pair#547. The fix is `weave refresh`, but nothing says so.

## Spec

Before invoking a substrate helper, `sdlc` checks the helper's capability (e.g. a version or capability line, or `--help` for `--slots`) and, on mismatch, refuses with one line naming the slot's substrate path and its commit, and the action `weave refresh` in the checkout. Same for other helpers sdlc calls with newer flags. No automatic refresh (it changes the slot's substrate under a running agent).

## Done when

- With a stale substrate fixture, `sdlc local-check run --name suite` refuses before queueing with the 'run weave refresh' message (test).
- A current compatible helper runs unchanged, including weighted admission and reentry.
- On a clean issue branch ahead of or divergent from main, `weave refresh --substrate` updates its declared private substrates and recompiles while preserving the consumer branch, HEAD, index and origin/main ref; other slots and shared suppliers remain untouched.
- Substrate-only refresh rejects `--rebase` and retains the existing ownership, cleanliness, declaration, race and partial-progress safeguards.
- A source-owned stale helper names integration of current main as its repair instead of claiming substrate refresh can update tracked source.
- Local-check and all weave regressions pass, followed by a recorded full suite before close and green PR CI before landing.

## Plan

- [x] Preflight required helper capabilities and refuse before acquisition with supplying checkout/HEAD and an executable repair.
- [x] Add substrate-only refresh on clean issue branches, preserving consumer Git state and existing dependency safeguards.
- [x] Verify regressions, update atlas, run full acceptance, then enter close review.

## Log

### 2026-10-14
- 2026-10-14: closed — README/atlas/help now document all new surface. BR-2 measured host evidence supplied by dispute: exact reviewed HEAD retry passes normally and with Git ceiling, all weave packages pass; recorded make test suite on code-identical 5784a8e0 passed 12 shards over 1342 SDLC tests and 54 other packages in 529.7s, sdlc-local/suite success, log sha256 prefix 538323fb56bd. Stale helper and divergent issue-branch guard mutations fail their production regressions; preservation/races/retry verified.; review verdict: SHIP
- 2026-10-14: flow upgraded quick → full — 238 added lines in code files (limit 100); an earlier round of this close already ran the full review

- 2026-10-14 — Hub approved plan f74dd4a3 (receipt f01d54c4-9201-4d44-af26-b074729c156e), including source-owned exception; change-code succeeded. The diff leaves the quick code shell, so close will run full review.
- 2026-10-14 — Helper RED reproduced stale acceptance, premature worktree mutations and misleading help prose; initial GREEN 22.673s. Caller inventory found five more legacy test doubles; a shared side-effect-free fake advertisement now replaces their acquisition-only scripts. Expanded caller verification is pending.
- 2026-10-14 — Weave CLI RED reached unknown --substrate after fixture setup corrections. Mode GREEN passed (11.959s), preserving consumer Git state and other-slot/primary sentinels while composing updated private substrate. Internal race, transitive order and retry cases passed; ignored-file fixture initially could not commit an ignored remote file, corrected with explicit force-add and then passed its guard test (2.949s). Combined all-weave verification next.

- 2026-10-14 — Expanded local-check, weighted admission, merge and landing targets GREEN (60.256s), plus cancelled-push resume (4.165s). All weave packages GREEN: go test ./cmd/weave/... -count=1 -timeout 15m; CLI 26.793s, refresh 54.518s. Full output: /tmp/ariadne366-weave-targeted.log.

- 2026-10-14 — Mutation controls on disposable detached HEAD 5784a8e0: removing helper preflight made TestStaleSuiteHelperRefusesBeforeAcquisition fail with stale helper accepted; removing root fetch exclusion made TestRefreshSubstrateIssueBranchKeepsConsumerAndOtherSlots fail by treating the divergent consumer as an update subject. Both expected RED; checkout restored and removed. Current origin/main 407e48a39057503fab7cdecaf6c69ce338ce7631 already integrated. Recorded full suite started on 5784a8e0fe358601faac2b5c737344e5f951b6aa through sdlc local-check; waiting in FIFO weighted gate. Log: /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/5784a8e0fe358601faac2b5c737344e5f951b6aa/suite.log. Uses the gate-owned default make test (all shards), rather than a separate unrecorded make -k run.

- 2026-10-14 — Full recorded suite GREEN on 5784a8e0fe358601faac2b5c737344e5f951b6aa: make test, 12 shards over 1342 cmd/sdlc tests and 54 other packages, all exit 0, wall 529.7s; sdlc-local/suite success, log sha256 prefix 538323fb56bd. Local acceptance complete; entering binary-owned close review.

- 2026-10-14 — Boundary round 1 REWORK: BR-1 missing README flag guidance; added complete usage, preservation, incompatible flags, private ownership, source-owned exception and no-dependency behavior. BR-2 reviewer could not initialize existing retry fixture under its Git ceiling. Host counter-measurements at reviewed HEAD 5d13f3bc5d5a3d703dbb6b2a6edb2794daf22f60: TestRefreshCLICompilesAndRetries passes both normally (13.043s, /tmp/ariadne366-retry-host.log) and with GIT_CEILING_DIRECTORIES set to the slot parent (14.085s, /tmp/ariadne366-retry-ceiling.log); fresh all-weave run passes (/tmp/ariadne366-review-all-weave.log). Recorded full suite also passed on code-identical ancestor 5784a8e0. Dispute will provide this measured host evidence as the finding explicitly requests.

## Revisions

- 2026-10-14 08:20 — add (TL): the named fix must work on an **issue branch**. `weave refresh` advances the
  repository's own branch too, so it refuses on an issue branch that's ahead of main (pair:5 on pair#547), and
  `--rebase` would rewrite the issue branch. Provide a substrate-only refresh (e.g. `weave refresh --substrate`):
  fast-forward the slot's private substrate checkouts and recompile, leaving the repository's branch alone; the
  refusal message names it.

- 2026-10-14 — Refreshed main-slot5 before claiming: f5b513d43608477ba059418ed9fd84e6068fc705 → 407e48a39057503fab7cdecaf6c69ce338ce7631; compile succeeded. Claimed #366 and entered planning. Read-only consumer inventory found acquireSuiteSlot as the single direct SDLC substrate-script caller; explicit and automatic local-check share it.
- 2026-10-14 — Proposed durable plan: workshop/plans/000366-substrate-helper-refresh-plan.md. Preserve the consumer as an observed read-only subject while private dependencies advance. Capability probes use existing argparse help rather than requiring a new helper protocol. Source-owned helper edge is explicitly submitted for hub review: a no-dependency source branch cannot adopt newer helper source through substrate-only refresh without moving that branch. Implementation awaits plan approval.

- 2026-10-14 — Current Done when now includes the 08:20 substrate-only revision and approved source-owned exception. Historical criteria are retained in the preceding revision record; the current contract explicitly separates local acceptance from PR CI/landing.
