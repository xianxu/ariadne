---
id: 000422
status: codecomplete
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours: 1.82
collaboration-mode: collaboration
card_mirror: 'da7df850bfe30200e77005c763ddd8f7067250bc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T01:34:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 0.67
---

# A downstream repo's committed generated vocabulary goes stale on every ariadne vocabulary change and turns its CI red

## Problem

## Spec

Found 2026-10-16 by parley.nvim:3. parley.nvim commits a copy of the generated `construct/generated/vocabulary/issue.json`. When ariadne's issue vocabulary changes, weave's apply refuses to overwrite the committed copy ("preserving authored replacement"). parley's main merge-check then goes red until someone refreshes it by hand. That has happened three times: parley #321, #326, and PR 285 after ariadne#415, which turned parley's main red from 22:32 to about 23:50 and blocked every parley landing.

Preserving a committed file that differs from weave's output is the right default for authored files. The gap is a generated file that a repo tracks on purpose. Options:
1. Repos stop committing generated vocabulary; it is built at bootstrap. Find out why parley committed it first.
2. A repo declares tracked generated paths, and apply refreshes those, showing the diff, instead of preserving them.
3. An ariadne vocabulary change ships with a fleet check that names each downstream repo whose committed copy it makes stale, so the refresh lands with the change.

## Done when

- An ariadne vocabulary change can no longer turn a downstream repo's main CI red silently: a test or check reproduces the parley case and shows the chosen fix handling it.
- parley's issue.json is handled by the chosen rule, and the reason is recorded.

## Plan

Approved by the TL (ops:0, 10-17), full flow. Durable plan: `workshop/plans/000422-a-downstream-repo-s-committed-generated-vocabulary-goes-stale-on-every-ariadne-vocabulary-change-and-turns-its-ci-red-plan.md`.

**Why parley commits issue.json (parley#208, 3e2daf45):** parley is a Neovim plugin that users install straight from
GitHub through a plugin manager, so they never run weave or bootstrap. It reads its issue vocabulary from the installed
tree ("shipped vocabulary available without CUE"). So the commit is a distribution artifact, and **option 1 is out**:
it would break every plugin install.

**Mechanism (parley#326's class answer, confirmed in `cmd/weave/internal/plan/ownership.go` `prepareManaged`):** weave
proves it wrote a file through its ownership inventory, `construct/generated/weave/`, which is gitignored. A cold CI
checkout has no inventory. An occupied staged output whose bytes differ from the new output, with no recorded identity,
is refused as "something weave did not write". So any upstream vocabulary change turns parley's main CI red until
someone regenerates by hand.

**Overlap:** ariadne#368 (filed from parley#326, 10-14; claimed by parley.nvim:2, never started) is this same class. Its
spec asks to keep authored-replacement protection intact (ARCH-SECURE) and to validate committed exports against
recorded provenance. Proposal: #422 implements it and #368 closes as superseded (the TL's call).

**Recommended: option 2 with committed provenance, not a blanket overwrite.** A repository declares the generated
outputs it ships in its own `construct/base.manifest`, as a new row `internal ship <path>` (parley:
`internal ship construct/generated/vocabulary/issue.json`). For each declared output, weave also writes a committed digest sidecar, `<path>.weave-sha256` (the sha256 of
what weave wrote). In `prepareManaged`, an occupied declared output whose bytes match its committed sidecar counts as
weave's own, even with no inventory, so a cold checkout regenerates it. Apply prints `refreshed tracked output <path>:
commit it` so the refreshed copy reaches the shipped plugin.
- A hand edit to issue.json changes its bytes without the sidecar, so it is still refused: protection stays intact.
- An undeclared output behaves exactly as today.
- Option 3 (a fleet check that names stale consumers) would only report the breakage, not prevent it, so it doesn't
  meet Done-when.

**parley's side** (cross-repo, routed through the TL): declare the path, un-ignore and commit the sidecar (one
`.gitignore` line beside the existing `!/construct/generated/vocabulary/issue.json`), and record the reason.

**Durable state (ARCH-CONSTRAINTS at-plan):** one sidecar per declared output: weave writes it, the consumer's git
tracks it, and it changes only when the output changes. It is removed when the declaration goes (weave's retirement).

- [x] `ship` manifest verb; sidecar write; ownership by committed digest in `prepareManaged`.
- [x] Tests (cmd/weave plan): the parley case — a cold checkout (no inventory), a declared output plus a matching
      sidecar, and a changed upstream output → regenerated with the report line; a hand-edited declared output → still
      refused; an undeclared output → refused as today; a warm checkout → unchanged.
- [x] Atlas: weave ownership and tracked outputs.
- [x] parley follow-up proposal to the TL (declaration and sidecar commit): filed as parley#371 (deps ariadne#422), for the TL to route after landing.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.6 impl=0.08
item: smaller-go-module    design=0.1 impl=0.14
item: smaller-go-module    design=0.2 impl=0.16
item: smaller-go-module    design=0.05 impl=0.12
item: atlas-docs           design=0.05 impl=0.05
item: milestone-review     design=0.0 impl=0.12
design-buffer: 0.15
total: 1.82
```

Items: the spec and investigation (parley#208, the cold-checkout mechanism, #368); the `ship` verb with `ShipStaged`;
the pair rule in `prepareManaged`; compile wiring with the e2e test; the atlas entry; the close review. Design is low
because the durable plan settles the decisions (+15% buffer); impl is at 40% of the v2.1 ranges.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log



- 2026-10-17: closed — Round-1 SHIP minors fixed: symlinked-sidecar row in TestShippedOutputOwnership (refused cold), warning names its reason (TestCompileRegeneratesShippedOutputCold updated), plan Revisions + ticks; go test ./cmd/weave/internal/plan and ./cmd/weave pass. Full suite green earlier at 18c4a983. (fixed-then-ship); review verdict: SHIP
- 2026-10-17: closed — make test green at 18c4a983 (12 shards, 1540 cmd/sdlc tests + 66 packages, 0 failed). TestShippedOutputOwnership (7-row matrix: cold parley pair regenerated; cold hand edit / no / malformed sidecar / undeclared refused; warm pulled pair regenerated with nothing displaced; warm hand edit displaced) and TestCompileRegeneratesShippedOutputCold (e2e) fail before the change. Live on a scratch copy of parley origin/main with #326-era issue.json, cold: without the ship row compile refuses (the parley CI failure); with row + committed sidecar it regenerates and prints commit-it lines; a hand edit is refused.; review verdict: SHIP
### 2026-10-16
- TL approved the plan (full flow). Closed ariadne#368 wontfix as superseded (pointer in its Log, published) and told
  parley.nvim:2. Filed parley#371 (unclaimed, deps ariadne#422) for parley's side: the manifest row, the committed
  sidecar, the `.gitignore` line and the reason.
- Code map (subagent): the refusal is `prepareManaged`'s `strict` case in `cmd/weave/internal/plan/ownership.go`;
  vocabulary outputs are `stagedOutput`s in ScopeArtifacts; the inventory is gitignored
  `construct/generated/weave/ownership.json`. The existing printManagedReport already prints "regenerated tracked
  output X: commit it".
- Implemented (c9bebdcb, 89bbdfbc, e1d79fbe, d727deed). `TestShippedOutputOwnership` covers the matrix (7 rows); the
  parley and pulled-copy rows failed before the pair rule, and the pulled-copy failure showed the sidecar being
  displaced, as plan-quality PQ-1 predicted. `TestCompileRegeneratesShippedOutputCold` (e2e) fails without the
  main.go wiring.
- Live, parley's real tree (a scratch `git archive` of parley origin/main, with issue.json from parley bdfbefa7, i.e.
  #326's older vocabulary, next to this branch's ariadne; no inventory = cold). Without the `ship` row: weave compile
  refuses issue.json, the exact parley CI failure. With the row and a committed sidecar: exit 0, issue.json
  regenerated (07911d2a → 8e951335), sidecar updated, and "regenerated tracked output …: commit it" printed for both.
  A hand edit then (cold, inventory removed): refused, bytes untouched. weave's block adds the sidecar to the ignore
  list, so parley needs its authored negation, as parley#371 specifies.
- Close review round 1: SHIP, 3 Minors, fixed in-round: a `TestShippedOutputOwnership` row for a symlinked sidecar
  (refused cold); the plan's Revisions record the `ShippedPaths` signature, and its checkboxes are ticked; the
  unmatched-row warning now names the reason. The reviewer's two `cmd/weave` failures came from its review checkout
  (`GIT_CEILING_DIRECTORIES` hides git under `.sdlc-review/`; one of them predates this change); both pass here and in
  `make test`.
