---
id: 000437
status: working
deps: [pair#661]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7722a79244483b48331e0c92efab6aab97221746' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T04:55:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "d0e81704", done: "584d65ed"}
---

# xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike

## Problem

Slice 4 of the universal-agent-input-interface project (design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md`, sections 3, 5 and 8). The shared review skill tells agents to edit documents with `couch --session-sync write` and to hand-write proposal markers under `deletions: propose-strike`. S4 gives edits one path, `apply`, with per-edit reasons, and one knob, `presentation: propose`, for proposals.

## Spec

Design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md` (operator-reviewed, 2026-10-17). Depends on pair's M1 (`couch --session-sync apply`, the derived round, the propose converter).
- `xx-universal-review` makes every edit to an existing document with `apply`: answering markers, fold-ins and collaboration edits. Each record is `{old, occurrence, new, explain}`, and the `explain` is written for the operator.
- An exit 6 is retried by reading again and rebuilding the records. An exit 1 `edits-invalid` is fixed and retried. Nothing is ever partial.
- `write` stays for generated content (tl-lead status's sections) and new side docs; its changes highlight via the derived round.
- The agent never hand-writes proposal markers: Couch converts under `presentation: propose`. On `propose-too-large`, the agent writes the rewrite as a side doc proposal (S2 naming), links it in place with `→ [proposal](…)`, and applies it on `🤖[fold in]` by the link (S3 fold-in).
- The reply lists the round's edits.
- `deletions: propose-strike` is retired in favor of `presentation: propose`, with the skill's prose, examples and examples checker updated.
- The hosted Pair review-workbench path in `construct/local/fix/SKILL.md` uses `pair review apply EDITS FILE --if-hash H --json`, commits the returned `Body` immediately, and retires the agent's handoff/landed-file and reconciliation instructions. pair#661 M2 keeps the legacy pane path alive until this skill lands; pair#664 retires it afterwards.

## Done when

The skill's contract and examples use `apply` for edits, cover exit-1 invalid/propose refusals and the exit-6 retry, the side-doc proposal path and fold-in, and no longer mention `deletions: propose-strike` except as retired. The examples checker passes. An agent round on a real document with the skill loaded arrives through `apply` with reasons. The Pair workbench instructions use `pair review apply` and commit its returned body without a pane-applied poke.

## Plan

- [x] Design the two apply adapters, Pair workbench transition and v3 export compatibility in `workshop/plans/000437-universal-review-edits-through-apply-plan.md`; get hub approval.
- [x] Update the skill and exporter with focused red/green contract tests, regenerate the manifest, and verify the examples checker.
- [ ] Run a live tracked-document apply round with reasons using the landed Pair engine, run the recorded suite, close and land after parley#383's dual-version consumer is in place.

## Log

### 2026-10-18

- Claimed and started in ariadne:1. pair#661 M1's Couch apply and round are on main; pair:2 is building the M2 pair-level apply. Its exact JSON and exit-code contracts are recorded in the durable plan. The hub added the xx-fix hosted workbench migration to this issue, with pair#664 retiring the legacy pane only after #437 lands. The hub assigned parley#383 to accept both export v2 and v3 before #437 publishes v3.
- The first tests failed on the retired mode flag and old adapter path, then passed after the authored skill/export changes. A large-proposal fixture exposed that the shared `pkg/sidedoc` parser did not count `→ [proposal](…)` as a link; a red test reproduced it, then the parser, examples checker and fixture passed. `go test ./pkg/reviewexport -count=1`, targeted `pkg/sidedoc` checks, `go test ./cmd/weave/... -count=1` and `make -k build` passed.
- Pair confirmed that M1 would propose the structural link itself. The hub extended Pair #661 M2: a body record answering the operator's own ready marker applies directly under `presentation: propose`, including a large fold, and the matching `discussions:` entry rides atomically with that link. The skill and fixtures now describe that contract; live conformance waits for Pair M2. Parley #383 must land before this branch's v3 export.
- A read-only skill pressure scenario compared the old and edited adapters: the old path rewrites through `write --all` and has no refusal path for a large proposal; the edited path uses records, the engine's refusal, a checked link/index, and explicit exit-6 rebuild and exit-3 wait. The remaining threshold discovery is deliberately owned by Pair's `propose-too-large` refusal, so the skill does not duplicate a tunable limit.

## Revisions

- 2026-10-18, M2 scope coordination: extend the original skill-only issue to migrate the hosted Pair review-workbench protocol in `construct/local/fix/SKILL.md`; preserve the legacy pane until pair#664 follows. Remove mode-level `deletions` from the authored skill and v3 export only after parley#383 accepts both versions. Pair's propose converter requires `write --frontmatter` for frontmatter-only changes, so that existing dedicated verb remains a narrow exception to body edits through `apply`.
- 2026-10-18, proposal side docs: keep a checked, atomic direct-apply exception for a record answering the operator's own ready marker, including a link and its matching `discussions:` entry or a large fold under `presentation: propose`. Pair #661 M2 owns the engine check; #437 documents and verifies it. This prevents a proposed-but-not-linked side doc and a fold that cannot pass the proposal size limit.
