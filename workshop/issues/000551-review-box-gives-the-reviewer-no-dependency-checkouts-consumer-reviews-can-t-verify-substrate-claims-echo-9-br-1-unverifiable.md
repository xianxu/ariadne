---
id: 000551
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 2.37
collaboration-mode: collaboration
card_mirror: 'f438bc40c79e27b4eb10efa5a2d14e4e6051b354' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T05:22:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.09
---

# review box gives the reviewer no dependency checkouts; consumer reviews can't verify substrate claims (echo#9 BR-1 unverifiable)

## Problem

## Spec

From ops-echo:0 (10-22 ~05:30): echo#9's boundary review returned REWORK with one Critical, BR-1 "unverifiable",
because the reviewer couldn't check echo against ariadne's fixtures and templates: it looked for
`/Volumes/TestScratch/review-checkouts/ariadne`, which doesn't exist (`review-checkouts` is the review box's root for
the disposable clone of the repo **under review**, `cmd/sdlc/internal/reviewbox/checkout.go`). No product defect was
found. Every echo close (echo#9, #10, #3, #5) touches the ariadne-seeded workflow, so each would hit the same wall.

The review box gives the reviewer only the repo under review, never its declared dependencies, so a consumer repo's
review can't verify claims that depend on the substrate. Same class as ariadne#545 (suites couldn't find
dependencies).

Fix: the review box also provides read-only, disposable checkouts of the repo's `construct/deps` dependencies (the
substrate at the exact pinned revision or main, as the deps row says), placed beside the review clone, with their
paths told to the reviewer in the review prompt (mirroring #545's SDLC_DEP_* names). Same lifecycle as the review
clone (ARCH-FUNERAL), never writable, bounded size and time.

## Done when

- A consumer boundary review receives disposable dependency checkouts pinned to last-fetched origin/main. The prompt names SDLC_DEP_* paths, exact SHA, fetch age or unknown, and mode; the same bindings reach reviewer-run commands. Available OS fences make dependencies read-only; checkout-only mode labels writable disposable copies whose edits are discarded (test).
- Removed with the review clone; repos with no deps are unchanged (test).

## Plan

Implementation plan: [review dependency checkouts](../plans/000551-review-dependency-checkouts-plan.md).

- [x] Share bounded declaration resolution and dependency descriptors.
- [x] Create pinned disposable dependency clones with retry/cleanup ownership.
- [ ] Expose paths, SHA, freshness and isolation mode in boundary prompts; verify and land.

## Log

### 2026-10-22
- 2026-10-22: closed — Full branch review supersedes narrow round 4 for BR-3 (visible manifest addition). Context guard, production boundary/dependency tests and full judge/reviewbox/layergraph/golint packages passed; clean runtime 26b4be05 host fence raw evidence is committed. Recorded suite lease-2607030942 remains queued on code-identical 1151135bfd02; remaining plan checkbox is suite/live/landing, so review overlaps verification under tl-worker. TL owns project. Echo live proof waits on independent #553, then approved main integration and rebuilt binary; no landing until final suite and owner-run proof.; review verdict: SHIP
- 2026-10-22: closed — Boundary/command environment regressions, full judge/reviewbox/layergraph, golint, promisor no-fetch and real OS-fence conformance pass. Host fence raw output and immutable source identity are committed at workshop/plans/000551-host-fence-evidence.md; BR-2 dispute supplies the exact measured result for explicit reviewer disposition. BR-1 README documentation is addressed. Recorded suite lease-3344910554 still runs on equivalent implementation 0ae0016dbbe6634bfbf59132045e683bfa76862a; subsequent changes are docs/evidence only. Echo #10 live review is coordinated. --no-plan-check acknowledges pending suite/live/close/landing release steps under the allowed review overlap; both proofs precede landing. --no-project: hub owns project updates.; review verdict: SHIP


## Revisions

### 2026-10-22 — TL-approved dependency and isolation semantics

The current construct/deps grammar has no revision-pin field. Reuse it without extension, resolve each direct local dependency, and pin refs/remotes/origin/main as last fetched once per review. Do not fetch in setup or retry. The prompt names the exact SHA and evidenced fetch age (unknown when not provable).

The original never-writable condition is revised by TL instruction: with an available OS fence, dependency clones are read-only; without it, give the reviewer disposable dependency clones, permit writes only under the existing checkout isolation guarantee, label that mode, and discard their mutations on retry/cleanup. Do not refuse an otherwise supported review solely because the fence is unavailable. No live dependency checkout is substituted for its disposable copy.

### 2026-10-22 — design checkpoint

Claimed and started planning after #548 landed. #531 is parked by TL. Read #545 environment resolution, layergraph grammar/resolver, reviewbox checkout/fence/confinement and the shared boundary dispatch. Read-only delegated lifecycle audit identified that an unowned sibling directory would be swept during a live review; the plan therefore reuses individually locked Create clones retained by the primary owner. No source implementation started. Full plan includes direct substrate/data bindings, exact last-fetched pins, freshness uncertainty, bounded materialization, retry/death cleanup and real-Git/fence proof. ARCH-DRY reuses existing declaration and checkout owners; ARCH-FUNERAL keeps all artifacts within their current cleanup lifecycle.


### 2026-10-22 — approved plan and plan-quality refinement

TL approved the full design and Echo-owned #10 boundary review as live conformance. Initial change-code inferred quick while the diff was still empty; reran with --flow full to obtain the intended plan-quality gate before implementation. Round 1 found PQ-1 contextual checkout/retry propagation and PQ-2 repeated test inventories; refined both as recorded in the plan. Current Done-when now reflects the approved modes. TL approved disposable environment exports, including Claude shell restoration. Echo supplied exact BR-1 evidence at its commit 5a3ada9 and confirmed #10 readers honor SDLC_DEP_SUBSTRATE. No source implementation yet.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; calibration marked stale, so provisional. Familiar existing Go/reviewbox/layergraph stack, no vendor propagation. Existing parser, resolver, flock lifecycle, process-group runner and fence implementations supply the library shortcut; the new budget/descriptor unit uses 1.0h base design halved for reuse, then ×0.2 for the approved concrete plan. Smaller extensions use 0.2h base design ×0.2 and 0.4h base implementation ×0.4. Context refactor uses 0.5h base design ×0.2 and 0.5h implementation ×0.4. Issue shaping retains 0.5h design; docs/review design receive the concrete-plan discount. Implementation values below already include v3.1's ×0.4. Bounds, live Echo conformance and regression proof are included, not separate scope.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.50 impl=0.08
item: smaller-go-module design=0.04 impl=0.16
item: smaller-go-module design=0.04 impl=0.16
item: smaller-go-module design=0.04 impl=0.16
item: greenfield-go-module design=0.10 impl=0.32
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 2.37
```

The three smaller modules are shared resolution, child-clone ownership and prompt/environment integration. Greenfield covers the bounded descriptor/tree policy. Design 0.88h ×1.15 + implementation 1.36h = 2.372h.


### 2026-10-22 — implementation checkpoint before combined verification

Plan-quality passed with all findings disposed (round 2; round 3 after the estimate body also CLEAN). Estimate-quality INFO accepted the provisional 2.37h derivation and noted implementation allowances may be optimistic; context refactor owns contextual checkout/reset/retry and fence wiring, smaller modules own resolution/child handles/prompt-environment projection, and their tests plus the shared verification/review allowance cover targeted checks, the recorded suite and owner-run Echo conformance.

Initial production-boundary regression failed with the missing SDLC_DEP_SUBSTRATE prompt binding before implementation (0.818s command). Added the exact Echo #9 Critical capture as cmd/sdlc/testdata/review-dependency-echo9.txt; owner supplied its source at echo commit 5a3ada9. Echo #9 subsequently landed through its dispute, so Echo holds #10 for this branch's owner-run check.

Implemented shared direct resolution/environment projection, pinned dependency clones, contextual ownership/reset/cleanup and full prompt/process/Claude-shell bindings. Lifecycle tests found cancellation cause lost through a %v wrapper; fixed it to preserve errors.Is. Targeted main regression/lifecycle/control checks passed in 9.470s. Full judge/reviewbox/layergraph package checks passed (13.332s/18.074s/4.268s). Real disposable OS-fence conformance passed (0.668s command), proving clone/source/common-dir write denials and primary write permission.

Ad-hoc bounded audit found one Important gap: Git ls-tree may lazily fetch promisor objects. The real missing-blob regression reproduced upload-pack invocation and source-config mutation. Disabled lazy fetching and denied transports for dependency observations; dependency materialization only allows the explicit local clone transport. Missing-promisor and both boundary-context regressions then passed (6.189s). The combined final targeted/lint/fence checks are running; no recorded suite queued yet. No commits to other repositories, no source dependency fetches, and no :0 stage requested.


Final combined checks passed on this implementation: judge 14.352s, reviewbox 21.398s, layergraph 4.891s and golint 1.308s. Boundary/dependency/local-check controls plus real OS-fence conformance passed together in 46.492s. Checkpointing the tested code next, then building a clean branch binary for Echo #10 and queuing the required recorded suite.


### 2026-10-22 — close review round 1

Mandatory boundary review found no runtime correctness blocker and returned one Important BR-1, user-surface-documentation: README omitted the new reviewer/preparation environment contract while atlas documented it. Added README guidance for bindings, exact last-fetched pins, no fetch, repair/retry, modes, freshness and lifecycle, with an atlas link. Existing local-check environment guidance still applies to local checks. The recorded suite remains running on unchanged implementation code; only documentation and review records change for the next close. Reviewer could not nest the live OS fence in its own sandbox; the owner-run disposable fence conformance had already passed outside that nesting, with the exact test command/results logged above.


### 2026-10-22 — close review round 2 and host evidence

Round 2 addressed BR-1, found no implementation defect, but raised Important BR-2 (unverifiable, os-fence-conformance): its nested sandbox cannot run sandbox-exec, and it needs captured host output rather than the prose summary. Re-ran the exact opt-in TestReviewDependencyLiveFence outside that nested sandbox on HEAD 80e095147c447abf1a3c41799029a050eca5817f; PASS, exit 0, 1.391s command / 0.91s test. [Host fence evidence](../plans/000551-host-fence-evidence.md) captures raw output, UTC capture time, tested HEAD and immutable source blob IDs. Only generated Markdown review sidecars were modified at execution. The same code is covered by the still-running recorded suite; evidence commits introduce no executable changes. Submit measured host evidence through finding dispute so the next reviewer must answer it explicitly.


### 2026-10-22 — first full suite and live proof, fixes before release

The recorded suite at implementation 0ae0016dbbe6634bfbf59132045e683bfa76862a completed all 8 shards and 76 other packages: wall 879.3s, 2236 SDLC tests. It failed TestVerbContextsReachTrackerReads (our new compatibility wrappers created Background contexts) and TestDocRepositoryCancellation (empty PID-readiness file). Raw log: .git/sdlc/test-stats/failed/4638dd482de33746116a1788f13ba510.log; SHA256 1776aed53279d84d6ecf1877f0646d63862a09edc9a18b9acba3fe9eeba42fef. The context failure is reproducible source work, so local-check resume is not applicable to that mixed failure.

Classified the independent tl-lead failure on a clean detached worktree at the exact failed head: whole TestDocRepositoryCancellation PASS (0.23s test, 0.390s command), with unchanged clean Git state. Filed and registered unclaimed #552 with both observations; removed the disposable classification worktree afterward. No test-source change in that family is included here.

Close round 3 withdrew the host-fence unverifiable finding, raised no new findings, and published codecomplete with SHIP. Before landing, Echo's required live check then failed: Claude's command shell was zsh, and our Bash-only reserved-variable expansion caused every command to fail with zsh:2 bad substitution. Echo held the rerun. Captured that exact diagnostic in reviewbox testdata; real shell matrix reproduced it before the fix, and now passes Bash/zsh/sh with no eval and without executing hostile values. Dependency clones were not skipped in no-fence mode; the manifest was originally only in the reviewer input, so absence from stdout did not establish missing clones. The same descriptor manifest now prints in close output, verified by a red/green production-boundary test.

Removed context-free compatibility wrappers rather than weakening the existing guard; threaded caller context through confinement, checkout, reset and dry-run protection. Combined guard/boundary/dependency/local-check and real-fence tests now PASS (14.842s); full judge/reviewbox/layergraph/golint PASS (7.678s/8.397s/1.391s/0.681s). Updated shell-capture matrix PASS (0.338s command). This runtime revision will be rebuilt for Echo, run through a fresh recorded suite, and receive the post-close fix review. The original clean binary is held and must not be used for another live attempt.


Replacement clean binary built at 26b4be05e40542b96adeda844591ddee88fcf366 (vcs.modified=false) and relayed by ops-echo to echo:3 for the owner-run #10 rerun. Re-ran live OS-fence conformance on this clean runtime revision: exit 0, PASS, 1.39s test / 1.910s command. Appended raw output and updated source blob identities to the host-fence artifact. Queue a fresh recorded suite on the evidence-only successor and run the required narrow fix review; preserve the Echo binary while it is in use.

### 2026-10-22 — review scope correction and retained verification

Round 4's narrow review refused with Critical BR-3, post-close-scope-control: printing the existing dependency manifest is observable behavior outside the earlier findings. Its targeted verification passed. Accept the finding and run a full close, rather than disputing it or repeating the narrow review. The plan Revisions supersede the earlier narrow-review instruction. Recorded suite lease-2607030942 remains alive on pushed head 1151135bfd024f4b709f1af158a58ed52f7a2ce9, normal FIFO position 5 at observation; no cancellation, duplicate or restart.

The Echo-owned live review cannot launch until independent #553 fixes attribution of Git-default integrate-main merge heads; both fleet and branch binaries refuse before review. TL assigned #553 to ariadne:5 and revoked #551's grant so other landings proceed. Keep #551 out of the landing head while waiting, then integrate landed #553 from main and rebuild the combined binary for Echo. Answered ariadne:5's attribution question: exact and malformed train handling must precede structural ordinary-merge fallback; never union publication-main issues.
