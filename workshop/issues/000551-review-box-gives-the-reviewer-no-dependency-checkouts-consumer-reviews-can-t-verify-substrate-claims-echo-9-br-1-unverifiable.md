---
id: 000551
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 2.37
collaboration-mode: collaboration
card_mirror: '0048dd0d5a57ccd9c595e730379c6b54d70ad51c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T05:22:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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
