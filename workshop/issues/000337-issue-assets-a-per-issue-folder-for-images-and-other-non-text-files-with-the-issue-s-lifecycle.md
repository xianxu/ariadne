---
id: 000337
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours:
card_mirror: '2fb12787270130defada498e4a08cfe3915b6676' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T13:03:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle

## Problem

Issues are text only. Couch's DOM submission (pair#483) and the terminal note submission (pair#455) produce screenshot crops that an agent may turn into a new ticket or attach to an existing one ("this demonstrates XXX#NNN, update that ticket"). Today there's nowhere in the issue format for those images to live with the ticket and follow it through its lifecycle. The operator, 10-12: extend the ticket format to hold non-text files, following parley.nvim's chat assets.

## Spec

Unshaped; a starting sketch from parley.nvim's `lua/parley/assets.lua` (#231):
- A per-issue sidecar folder keyed by the issue's **id**, not its slug, so a rename never orphans it (e.g. `workshop/issues/assets/<id>/`).
- **The issue body is the index:** the folder holds only bytes the body links (`![](assets/<id>/<file>)`); nothing scans the folder to discover content.
- **Same lifecycle as the issue:** archive to `workshop/history/` moves the folder too; `sdlc migrate` and peer moves carry it; one writer owns create/move/delete (ARCH-DRY).
- **Git implications (the operator, 10-12):** binary files change how the repos use git: repository size and history weight, diffs, encrypted repos (gcrypt brain), publication through `sdlc issue publish` and the issue-tracker branch (#252). Decide in the design: plain git with size limits, LFS, or keeping bytes outside git with a pointer.

### Approved design — 2026-10-13

The original sketch above is retained as history. The approved design is
[the durable plan](../plans/000337-issue-assets-plan.md): one body-indexed bundle
per issue, stored in plain Git with immutable content names, bounded at 1 MiB
per file, 8 MiB and 32 files per issue. The binary resolves operator-owned
allowed/text-only policy before attaching; unset policy derives from host
visibility, with public/unknown becoming text-only. Help prints policy and
source; every refusal gives its clearing action. Source sharing eligibility
remains independently required by the Couch rule from pair#526.

Assets move with the issue through all supported in-repo lifecycle paths,
including publish, archive, restore, abandon/reopen and slot move, using one
shared bundle model and existing exact-byte Git transactions. Ownership and
retry proofs cover asset-only changes. Cross-repo migration remains refused
and is explicitly deferred to #347. Agents never change policy to admit bytes.

## Done when

- An agent can attach an image or other bounded file through an sdlc verb; the body links it and the bundle follows publish, archive, restore, abandonment/reopen and slot move within the repository.
- The attach verb enforces operator-owned allowed/text-only repository policy, derived from host visibility when unset; public or unknown visibility defaults to text-only. Help prints the policy and source. Refusals write nothing.
- Publication and lifecycle retries preserve exact linked bytes and existing ownership protections, including asset-only changes and legacy repositories.
- Cross-repo issue migration remains safely refused; its asset support is tracked by dependent #347.
- The git storage decision is recorded with its reasons.

## Plan

Approved [implementation plan](../plans/000337-issue-assets-plan.md); next gate is change-code.

- [ ] Typed asset policy, bounded bundle model, attach verb and dynamic policy help.
- [ ] Bundle publication, ownership, recovery and restore.
- [ ] Archive, abandon/reopen, legacy lifecycle and slot-move verification.
- [ ] Documentation, targeted tests, full suite below load 40, close review and shipping on SHIP.

## Log

### 2026-10-12

Filed from pair#483's shaping (pair:2), at the operator's request. Not a dependency of #483: a submission is written to Couch's store first; attaching it to a ticket is step two.

### 2026-10-13 — design exploration

Claimed in ariadne:5 and entered planning. Existing cross-repo migrate explicitly
refuses issues because destination renumbering is deferred; hub narrowed this
issue to the in-repo lifecycle and requested unassigned follow-up #347.
Read pair origin/main's canonical couch skill: only explicit shareable submission
content is eligible, private/unknown yields sanitized text without crops, and
recordings remain local and expendable. The installed couch skill predates that
rule. No implementation or tests run yet. ARCH-DRY and ARCH-PURPOSE motivate one
bundle model across publication, all archive paths, abandon/reopen and restore.

## Revisions

- 2026-10-12 — the operator: issue assets are where submitted web-cast note crops (pair#483) land when notes become an issue
  (text first in the issue body; only the crops that matter as assets; recordings stay local, never S3 yet). Add a
  **privacy gate enforced by the binary**, not agent judgment: a per-repo asset policy (`allowed` | `text-only`), defaulting
  to `text-only` for a public repository (from its host visibility) and to `allowed` for a private one; the attach verb
  refuses images under `text-only` and says why; the policy and its source are printed by `sdlc issue --help`. The
  operator changes a repo's policy; agents never do.

- 2026-10-13T20:06Z — Hub scope direction (ops:0, delivery d0338359-494f-4951-838c-4e18ae676bf6): deliver attach/privacy and the complete in-repository bundle lifecycle; retain migrate's issue-family refusal. Cross-repo issue migration needs the renumbering capability deferred by #179 and is separated into #347, dependent on #337 and left unassigned. This revises the original migrate acceptance criterion explicitly.
- 2026-10-13T20:06Z — Proposed design, pending plan review: plain Git with 1 MiB/file, 8 MiB/issue and 32-file limits; body-indexed immutable content-named assets, shared lifecycle planner, strict operator-owned repository policy, and unknown visibility treated as text-only. Full details and alternatives are in the durable plan; limits remain proposed choices until approved.
- 2026-10-13T20:09Z — Ops:0 approved plan commit 399e38374dd9d769c9d21748e53b15757ee70b16 and the proposed limits, directing change-code. Added the operator's actionable-refusal requirement: each refusal names the action that clears it. The plan records this addition.
- 2026-10-13T20:12Z — Plan-quality round 1 refused with PQ-1 (binary subtree readers), PQ-2 (attachment rollback/restart), PQ-3 (function-level test strategy). Verified existing Files consumers, refined the plan across each class, and recorded prevention rules in lessons. No code changes; rerunning plan-quality.
