---
id: 000623
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours: 1.84
card_mirror: 'db602c0f49eb155e3b9b109ef4ff31952bc8d4d6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T12:00:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Instance-conformance gate validates issue .side/ side docs as issues; exempt and validate by type

## Problem

## Spec

**The instance-conformance gate validates side docs under an issue's `.side/` folder as issues (parley.nvim:1, 10-25).**
parley.nvim#440's merge (PR #354) was refused: the gate checks every `.md` under `workshop/issues/`, including
`workshop/issues/000440-….side/upstream-neovim-draft.md`, a `type: side-doc` file, and demands an issue's status,
Spec, Plan and Done when. Merge's own archive step already moves an issue's `.side/` folder with it, so side docs
there are an intended layout; the gate is the inconsistency.

Exempt files under `<issue>.side/` from Issue validation, and validate them by their own `type:` (side-doc
frontmatter: type, parent) instead. Use the same notion of "this file belongs to issue N's side folder" that the
archive step uses (one shared helper, not a second path rule).

## Done when

- A repo with an issue `.side/` side doc passes the gate; a malformed side doc (missing type or parent) fails with a
  side-doc message; a malformed issue still fails as before. Tests for all three.
- The archive step and the gate share the side-folder rule.
- **Plan side, same gap:** merge's archive step refuses a `workshop/plans/<issue>-…-plan.side/` folder whose plan document doesn't exist ('active or conflicting archive artifact'), leaving it on main after the issue archives (parley.nvim#440: its upstream-neovim-draft.md stayed in workshop/plans while the archived issue's link points under history/plans). A side folder archives with its issue whether or not a plan document exists; test both. Then archive parley.nvim's leftover file.
- **Duplicate-ID check, same gap (pair#858, 10-25):** `sdlc merge` refused `#000858 would be claimed by 2 files after merge: workshop/issues/000858-….md and workshop/issues/000858-….side`. `cmd/sdlc/issueids.go` (refIDSpace/renderClashes) feeds non-recursive `git ls-tree --name-only` output into `cmd/sdlc/internal/issue/scaffold.go` PathsByID, which parses the `.side` directory name as another issue; LocalPathsByID also admits directory names. ID scanning ignores `<issue>.side/` entries through the same shared side-folder rule as the gate and the archive step; test with an issue plus its `.side/` folder.

## Plan

- [x] Approve [the shared side-folder design](../plans/000623-side-folder-rule-plan.md), then enter change-code.
- [x] Implement shared ownership, typed conformance, ID-space exclusions and complete archive/recovery coverage with red-first tests.
- [ ] Verify serial targeted tests and owner-coordinated parley.nvim leftover repair; obtain hub authorization for recorded-check/landing phase.
- [ ] Complete required verification, close review and landing.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; estimate-source flags the calibration stale, so provisional. Familiar Go/CUE/Git territory gives familiarity1.0. The thorough approved plan applies0.2 design discount and15% design buffer. Owner/ID work maps to a smaller module; schema/gate wiring maps to one new concern with the existing CUE engine halving its raw1h design before the0.2 discount. Archive/recovery is a cross-cutting refactor. The narrow owner-run repair is the small cross-repo primitive, with no new API discovery. Implementation uses the upper v2 ranges, scaled once by0.40. Targeted/private conformance is included in each implementation primitive; one review primitive covers the single close boundary. Full recorded-check/landing remain budget-gated by the hub and may add elapsed ship-time beyond this provisional estimate.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.1 impl=0.08
item: smaller-go-module design=0.06 impl=0.2
item: greenfield-go-module design=0.1 impl=0.32
item: cross-cutting-refactor design=0.2 impl=0.2
item: cross-repo-refactor-small design=0.06 impl=0.12
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0 impl=0.2
design-buffer: 0.15
total: 1.84
```

Design0.56×1.15=0.644 plus implementation1.20 gives1.844h, rounded1.84h.

## Log

### 2026-10-25

### 2026-10-25 — claimed; design checkpoint

Recovered on clean resting main-slot1 after #560's completed landing; no old runner is resumed. Claimed #623 and started its independent branch from current main. Read the issue and all three consumer families; a read-only audit found existing sidedoc helpers but no side-doc vocabulary model, plus document-dependent folder discovery across archive variants. Full design is warranted rather than a <=100-line quick patch that leaves archive/recovery siblings inconsistent. Existing Folder excludes a directory path itself; preserve that contract and add shared owner projection. Plan is ready for TL approval; no implementation or tests started. Machine budget is targeted only, one process at a time, count1; a full recorded check/merge waits for hub authorization. No new estimate before plan-quality.

### 2026-10-25 — design approved; transitional inventory complete

TL approved shared Owner, side roots conferring no ID and archive ownership independent of a plan document, with an explicit legacy compatibility condition. Read-only inventory across all48 registered worktrees and pinned origin/main of ariadne/pair/parley.nvim/metis/echo found4 typed Markdown side docs, all with nonempty parent; pair has168 untyped records, including15 Markdown and pair#858's14 JSON/log evidence files. Persisted condensed complete path/content inventory at ../plans/000623-side-inventory.json; plan records source pins, scope and counts. Choose typed-Markdown-only validation, so no attachment backfill or new rejection is needed. Parent target existence remains outside conformance. No implementation or tests yet; next run full-flow change-code.

### 2026-10-25 — plan-quality accepted; estimate derivation

First plan-quality round found no blocking issues. PQ-1 (Minor) identified the state table's outdated missing-type failure row; corrected it to the approved untyped-attachment compatibility rule. Added paired vocabulary/sdlc rollout guidance so compound-noun support is available before the new gate is deployed. Estimate is now derived from the accepted scope; full change-code must admit it before code or tests.


### 2026-10-25 — implementation checkpoint

Change-code admitted the approved design. Shared Owner and ID exclusions pass their focused tests; the side-doc vocabulary and compound-noun selector pass native CUE tests. Root command regressions reproduced the gate misclassification and orphan-plan-side omissions in landing, filesystem archive, recovery and abandon. Gate dispatch and configured-plan-root wiring are written but not yet verified; archive production changes remain outstanding. No recorded check or merge is running. Next: verify gate changes, implement shared archive owner inventory across these paths, then run the command regressions serially with count1. TL dispatched the historical parley repair to parley.nvim:1; this branch makes no peer edits.


### 2026-10-25 — targeted implementation verified

Shared inventory now covers filesystem archive/recovery, immutable landing, trunk archive, kept abandon trees, reopen and successor replacement. New real-Git regressions first reproduced absent-plan omissions, then passed (38.208s), preserving exact blobs, selected membership and idempotent retry. Existing and new side-related command controls plus ordinary gate controls passed serially (297.200s): go test ./cmd/sdlc -run '^(Test.*Side.*|TestNounGatesDefaultIssueHomeComesFromVocabulary|TestValidateChangedInstancesDispatchesByNoun|TestValidateChangedIssues|TestIssueIDGateIgnoresSideDirectories)$' -count=1 -parallel=1 -timeout15m. Counted8/16-family indexing asserts one owner visit per path. Separate serial focused pkg/sidedoc, internal/issue and cmd/vocabulary tests passed (0.164s/0.175s/0.402s), including native CUE required types and additive metadata. Updated atlas and lessons; no full suite or PR has started. Next: clean paired vocabulary/sdlc build and private conformance, collect parley.nvim:1 repair evidence, then report to hub for recorded-check authorization. GitHub Actions outage is infrastructure; any later CI gate waits for authenticated recovery, without reruns or bypass.


### 2026-10-25 — clean paired-binary conformance; ready for next phase

Built both vocabulary and sdlc from pushed fb17b728a13c6ac2be80e5543d1abc4760a65fa5; go version -m reports that exact revision and vcs.modified=false for each. The vocabulary CLI accepted a typed side doc with a nonexistent parent target and nested additive metadata (exit0), refused a missing parent (exit1, parent required), and refused type: issue (exit1, want side-doc). The disposable build was /Volumes/TestScratch/ariadne623-conformance-llrjmf_j; that path is ephemeral, while source, outcomes and native CUE regressions are durable here. No :0 stage, peer mutation, full check or PR was started. Asked parley.nvim:1 for its owner-run #440 repair blob proof (Couch receiptc439765a-6fb3-4b29-b81e-8c613b6dbe28 submitted); repair remains that slot's work. Next: hub authorizes recorded-check/close phase under machine budget, then the required pre-close check and boundary review. Any CI wait must hold for hub recovery from the account-wide GitHub Actions billing outage.

## Revisions

### 2026-10-25 — approved typed-Markdown transition

Per TL approval deliveryeef33473-162a-49ca-b24a-3cd6b99ef4d2, replace the Done-when missing-type failure case for side-folder members: a Markdown member without a declared type is a valid unmodeled attachment and is never validated as an issue. Only explicitly typed Markdown side docs enter the new side-doc schema; missing/empty parent or a malformed declared type still fails with a side-doc diagnostic. JSON/log/text/patch and untyped Markdown evidence keeps its bytes and archives with the issue, without conferring issue identity. Inventory evidence is recorded in the durable plan before change-code.
