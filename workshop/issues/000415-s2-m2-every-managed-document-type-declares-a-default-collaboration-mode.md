---
id: 000415
status: codecomplete
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '60569104a3fc8936b7972e75776bb0d60fab23e6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T18:46:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.55
---

# S2 M2: every managed document type declares a default collaboration-mode

## Problem

S2 M1 (ariadne#412) gave each section of a reviewed document a collaboration mode (discussion, collaboration,
chat), with discussion as the implicit default when a document has no `collaboration-mode:` key. The operator wants
every ariadne-managed document type to declare its default explicitly, so a document's mode is visible at the top
and nothing new relies on the implicit default. The design is ops
`teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md` §6 M2 and side doc 04.
Past documents are not migrated (operator decision, QA files included).

## Spec

1. List every ariadne-managed document type, first-class or through the datatype system.
2. Give each a default collaboration mode. The first cut:
   - design docs and plans → discussion;
   - issues and projects → collaboration (short, factual edits);
   - the TL session file → chat for Operator instructions, discussion for the rest;
   - side docs and answer docs → discussion.
3. Each type's template, or the code that writes it, puts its default into new documents' frontmatter.
4. A lint flags a managed type that declares no default.

## Done when

- Every managed type has a declared default collaboration mode, and a new document of each type carries it in its
  frontmatter.
- A test fails when a managed type has no declared default, or when a writer emits a new document without it.
- No past document is migrated.

## Plan

**One declaration per type, at the artifact that already defines the type** (ARCH-DRY). Each writer emits the
declared default into a new document. A lint enumerates the types and fails on one with no valid default. The three
values are discussion, collaboration and chat (xx-universal-review's collaboration-modes.md); Go holds the list once
as `vocab.CollaborationModes`.

| Type | Declared in | Written by | Default |
|---|---|---|---|
| issue | issue.cue `scaffold: collaboration_mode` | `issue.Render` (and `DetailsFromCard`) | collaboration |
| project | project.cue `scaffold: collaboration_mode` (and the project.md table row, drift-tested against it) | `project.RenderScaffold` | collaboration |
| continuation, meeting-notes, procedure, prose, roadmap, target, product, pensive, type | a `collaboration-mode` row in the prototype's Frontmatter shape table | the agent, through xx-datatype ("emit the fields it lists") | roadmap, continuation, meeting-notes, procedure: collaboration; target, product, prose, pensive, type: discussion |
| durable plan | the writing-plans skill's header (a new frontmatter block) | the agent | discussion |
| side doc, answer doc | collaboration-modes.md §4 (already) | the agent | discussion |
| TL session file | tl-lead help (already) | the TL | chat for Operator instructions, discussion elsewhere |

Out of scope:
- Machine-written records nobody reviews: gate ledgers, close-review sidecars and doc-review sidecars.
- lessons.md and atlas pages (no frontmatter).
- Parley chats: parley's writer, so it goes to the TL as a one-line cross-repo proposal.
- Past documents: no migration (operator decision).

- [x] 1. `vocab.CollaborationModes` and `Issue()/Project().CollaborationMode()`, read from `scaffold.collaboration_mode` in issue.cue and project.cue, `make vocab-embed`; a vocab test that each is one of the three.
- [x] 2. `issue.Render` and `DetailsFromCard` write `collaboration-mode: <default>` (details only; it never reaches the card, which `SplitCard` already filters). Update the two byte goldens. `project.RenderScaffold` writes it. Tests: a new issue and a new project carry the model's default.
- [x] 3. pensive.cue (closed) gains the optional `"collaboration-mode"?` with the three values, so the pensive corpus test keeps passing. #Project stays open.
- [x] 4. A `collaboration-mode` row in each datatype prototype's Frontmatter shape table. type.md also says every prototype's table carries one.
- [x] 5. The writing-plans skill header gains `---\ncollaboration-mode: discussion\n---`. Check that sdlc's plan readers (plan-quality judge, design-limit line count, resolve) tolerate a frontmatter block.
- [x] 6. The lint, a test in cmd/datatype:
  - every merged prototype over this repo (`mergeTypes`, so a new prototype is caught automatically) has a `collaboration-mode` row whose first backticked value is one of the three;
  - the writing-plans header declares one;
  - project.md's row equals project.cue's default.
- [x] 7. The atlas: universal-review.md (per-type defaults) and the datatype page. The full suite, then close.

## Log

### 2026-10-16
- 2026-10-16: closed — The last advisory is fixed: renderToOutput takes its warning writer, and TestRenderWarnsOnATypeWithoutADefault renders a fixture tree where one prototype lacks a default. It asserts one warning, for that type only, and that SKILL.md is still written. ./cmd/datatype green; merge reruns the full suite. (fixed-then-ship); review verdict: SHIP
- 2026-10-16: closed — After SHIP: the datatype render (run by every weave compile) warns on a type without a default; TestCollaborationModesAgreeAcrossSources binds pensive.cue and collaboration-modes.md to vocab.CollaborationModes; collaborationModeDefault requires the notes to start with the default and accepts CRLF (new test cases); the test guards the project lookup. ./pkg/vocab and ./cmd/datatype green on 21ba3b6c; merge reruns the full suite. (fixed-then-ship); review verdict: SHIP
- 2026-10-16: closed — Full suite green on 614b8f11 (12 shards, 0 failed; sdlc-local/suite recorded). Every managed type declares a default collaboration-mode at its defining artifact. Issue and project: scaffold.collaboration_mode in their CUE models, written by issue.Render, DetailsFromCard and project.RenderScaffold (byte goldens, TestDetailsFromCard, TestRenderScaffoldDerivesProjectModel, TestScaffoldsDeclareACollaborationMode). The ten datatype prototypes: a collaboration-mode row each. `datatype lint` passes (10 types), and TestEveryTypeDeclaresACollaborationMode runs it over the registry and checks project.md against project.cue; TestLintTypesNamesTheTypeWithoutADefault and TestCollaborationModeDefault cover the failure paths. Durable plans: the writing-plans header (TestPlanHeaderDeclaresACollaborationMode). pensive.cue accepts the key, and the corpus test passes. Side docs and the session file were already declared by #412. No past document migrated.; review verdict: SHIP
- 2026-10-16: flow upgraded quick → full — 114 added lines in code files (limit 100)
- Built tasks 1 to 7. The lint became a real `datatype lint` subcommand, not only a test, so a derived repository's own prototypes can be checked too (ARCH-PURPOSE); the test runs the same function over this repository. A stray `datatype` binary from go build at the repo root briefly shadowed the leaf datatype/ dir; removed. Targeted: ./cmd/datatype ./pkg/vocab ./cmd/vocabulary ./cmd/sdlc/internal/... green; vet_test.sh run.
- Task 5 evidence (as the review asked): sdlc's plan readers tolerate the frontmatter block. Plans are recognized by their plans-dir id prefix, not by frontmatter (push.go archive moves). The design-limit counter counts 4 more lines per durable plan. The plan-quality judge reads the plan as text.
- Close review: SHIP, with Minors fixed in the same round:
  - every repository's `weave compile` runs the datatype render, which now warns on a type without a default, so derived repositories are covered with no extra wiring;
  - TestCollaborationModesAgreeAcrossSources binds pensive.cue's enum and collaboration-modes.md's error text to vocab.CollaborationModes;
  - the parser requires the notes to start with the default and accepts CRLF files;
  - the test guards the project lookup.
- A noted design point, not changed: issues default to collaboration (the operator's first cut), while an issue's `## Spec` is design. A per-section `collaboration-mode: {discussion: ["Spec"]}` on an issue is the lever if the operator wants Spec questions to branch.
- Fixed-then-ship round: SHIP; its one advisory (the compile-time warning was untested) fixed: renderToOutput takes its warning writer, and TestRenderWarnsOnATypeWithoutADefault renders a fixture tree with a prototype lacking a default.
