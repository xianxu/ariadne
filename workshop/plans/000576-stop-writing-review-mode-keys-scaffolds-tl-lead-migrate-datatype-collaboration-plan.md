# Stop writing review mode keys — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove every ariadne writer of the review mode keys retired by #575. Old documents stay readable and round-trip byte for byte. Land the shared `ChatSections` corrections pair#838 and parley.nvim#419 pin.

**Architecture:** Each writer drops its mode-key line. The CUE scaffold field and the datatype rows that fed the writers go away, and with them the Go code that read them. tl-lead's session-file contract moves to `@@chat@@`. Its chat items come from a grammar shared in `pkg/sidedoc`, so the review checker and tl-lead parse items one way. Readers keep tolerating every legacy key.

**Tech stack:** Go (sdlc, tl-lead, datatype, pkg/vocab, pkg/sidedoc), CUE (construct/vocabulary), markdown (datatype prototypes, skills, help).

**Already on the branch (agreed with the TL and pair:3 before this plan):**
- e12f0fccc and 3d109d2c8: `ChatSections` is linear (a counted N/2N test catches a reverted prefix count at ratio 4.07).
- The unread-frontmatter notice needs a closed fence block (`FrontBlock`), so a leading thematic break no longer gets it.
- An empty `collaboration-mode:` gets the notice "empty or malformed".
- `ChatSection.Legacy` keeps legacy provenance on tagged sections, and `Innermost` picks the governing section; a randomized laminar test pins both.

## The surface (mapped 10-23)

| Writer | What it writes | Change |
|---|---|---|
| `cmd/sdlc/internal/issue/scaffold.go:283` | `collaboration-mode: <issue.cue default>` | drop the line |
| `cmd/sdlc/internal/issue/handoff.go:281` `DetailsFromCard` | the same | drop it |
| `cmd/sdlc/internal/project/scaffold.go:23` | `collaboration-mode: <project.cue default>` | drop it |
| `construct/vocabulary/issue.cue:148`, `project.cue:63` | `scaffold: collaboration_mode` | delete the field; regenerate `pkg/vocab/*.json` (`make vocab-embed`) |
| `cmd/tl-lead/migrate.go:328` `answerFrontmatter` | adds `collaboration-mode: discussion`, drops any existing one, reorders `process-instruction` | write `type`/`parent`/`marker`, keep every other existing line verbatim and in order; update the golden `after/` fixtures |
| `cmd/tl-lead/help.md:336-343, 440-452` | tells the TL to write all three keys, a heading comment, and `processing-instruction` | `@@chat@@` above `## Operator instructions`; answer docs carry `type`/`parent`/`marker` only |
| `construct/datatype/*.md` (11 prototypes) and `type.md:34` | a required `collaboration-mode` row in the Frontmatter shape table; `type.md` requires the row | delete the rows and the requirement; `workingdoc.md`'s own frontmatter key stays (it is a document, and legacy is readable) |
| `construct/adapted/superpowers-writing-plans/SKILL.md:112-118` | the plan header's `collaboration-mode: discussion` | drop it from the header and its paragraph |

| Reader | Change |
|---|---|
| `cmd/datatype/collabmode.go` (`collaborationModeDefault`, `lintTypes`; `datatype lint`, render warning) | delete: it only checks the rows being removed; its tests go too |
| `pkg/vocab` `Scaffold.CollaborationMode`, `(*IssueModel/ProjectModel).CollaborationMode()` | delete |
| `pkg/vocab` `CollaborationModes` | rename to `LegacyCollaborationModes`, documented as the read-only legacy set; `TestCollaborationModesAgreeAcrossSources` keeps pinning it against `pensive.cue`'s optional enum and §7 |
| `construct/vocabulary/pensive.cue:25` optional enum | keep: deleting it would reject legacy pensive files (the closed schema) |
| `construct/local/fix/SKILL.md:240` refusal row naming `presentation:`/`collaboration-mode:` as operator-owned | reword: legacy keys are operator lines a round never rewrites |

Rewrite verbs already preserve legacy keys: `issue.SetField`, `ReconcileIndex`, `RebaseLinks`, `instructions link`, `SplitCard`. `answerFrontmatter` is the one lossy rewriter, and it is fixed above.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `ChatSections` / `ChatSection.Legacy` / `ChatScan.Innermost` | `pkg/sidedoc/chat.go` | modified |
| `ChatItems(section string, legacy bool) []ChatItem` | `pkg/sidedoc/chat.go` | new |
| `instructionBullets` | `cmd/tl-lead/instructions.go` | modified |
| `answerFrontmatter` | `cmd/tl-lead/migrate.go` | modified |
| issue/project `Render`, `DetailsFromCard` | `cmd/sdlc/internal/issue/`, `project/` | modified |
| `LegacyCollaborationModes` (was `CollaborationModes`) | `pkg/vocab/vocab.go` | modified |
| `collaborationModeDefault`, `lintTypes` | `cmd/datatype/collabmode.go` | deleted |

- **ChatItems:** the chat item grammar of collaboration-modes.md §3. An item starts at the shallowest list indent: `N.` items always, plus `-` and `*` items when `legacy` is set. It runs over sub-items and continuation lines until the next item, or until a blank line followed by prose. Fences start no item, and a tab indents to four. An item is answered when one of its lines outside a fence ends with `→ [answer](…)` or `→ done:`. Today's `instructionBullets` body moves here (ARCH-DRY: one item grammar for the review checker, tl-lead and pair). `instructionBullets` becomes a thin call. Tests: `chat_test.go` table rows for both grammars, fences, the blank-then-prose end, and answered links. tl-lead's existing bullet tests stay as its regression net.
- **How tl-lead chooses `legacy`:** it already reads the section by heading and can read the frontmatter. `ChatSections(frontmatter + section)` sees the heading comment (inside the section) and the map or scalar (in the frontmatter). The tag sits above the heading, outside the section, so it never decides grammar. `legacy` is the found section's `Legacy`, and false when no legacy form names it. This adds one Couch frontmatter read per `list` and `link`. The section hash still pins positions. The frontmatter read is advisory: it picks the grammar, and a race with an edit to it is resolved by the next list. ARCH-ORDER: a changed frontmatter between list and link could change bullet positions, so `link` reuses the grammar `list` printed in its output (the hash already covers the section).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `sessionSync.ReadFrontmatter` (existing) | `cmd/tl-lead/sessionsync.go` | modified use | `couch --session-sync read --frontmatter` |

The existing stateful fake gains no new state: the frontmatter read is already modeled.

ARCH-MOCK: no new dependency.
ARCH-EVOLVE: issue, project and side-doc frontmatter lose a written key, an additive removal. Readers never required it. A test rewrites an old issue that carries `collaboration-mode: collaboration` through `SetField` and through the issue/project verbs, and checks it byte for byte.
ARCH-SECURE: the frontmatter tl-lead reads is untrusted. Unreadable YAML yields the notice and `legacy=false`, never a crash.
ARCH-FUNERAL: N/A, because the issue creates nothing durable. It deletes code and rows.
ARCH-CONSTRAINTS: `ChatSections` and `ChatItems` are linear and counted. tl-lead's extra read is one Couch call per command.

## Tasks

### Task 1: Shared item grammar
- [ ] Failing `TestChatItems` rows: a tag grammar with `N.` items that ignores `-` lines; a legacy grammar that takes all three; a nested sub-item; a fence; a blank line followed by prose ending the item; `→ [answer](…)` on a follow-up line; and `→ done:`. Add a counted N/2N test.
- [ ] Move `instructionBullets`' body into `ChatItems`, and make `instructionBullets(section)` call `ChatItems(section, legacy)`.
- [ ] Make tl-lead `list` and `link` compute `legacy` from `ChatSections(frontmatter + section)`. Add fake-backed tests for a heading-comment session file (legacy, bullets) and a `@@chat@@` session file (numbered only, `-` lines are prose).
- [ ] Have `checkChat` in `pkg/reviewexport` use `ChatItems` for its "every item answered" check instead of its private `numberedItem`.

### Task 2: sdlc scaffolds and vocab
- [ ] Flip `scaffold_test.go` (:105, :146, :352), `handoff_test.go:64` and `project/scaffold_test.go:19` to assert that no mode key is written (byte-stable renders updated).
- [ ] Drop the three writer lines, delete the CUE fields, run `make vocab-embed`, and delete the Go accessors and `TestScaffoldsDeclareACollaborationMode`. Rename `CollaborationModes` to `LegacyCollaborationModes`.
- [ ] Add the legacy round-trip test: an issue whose details carry `collaboration-mode: collaboration` goes through `SetField` and a card publish, and the key line survives byte for byte.

### Task 3: tl-lead migrate and help
- [ ] Failing `answerFrontmatter` test: an existing side doc with `collaboration-mode: discussion`, `presentation-mode: highlight` and `process-instruction` keeps all three lines in order, and a new one gets only `type`/`parent`/`marker`. Implement, then update the golden `after/` fixtures.
- [ ] Rewrite `help.md`'s session-file and answer-doc instructions. Flip `editorial_help_test.go` and the `main_test.go:113` phrase to require `@@chat@@` and forbid the three keys.

### Task 4: Datatypes and the plan header
- [ ] Delete `cmd/datatype/collabmode.go`, its tests, and its calls in `runLint`/render. Remove the 11 rows and `type.md:34`'s requirement. Drop the writing-plans header key.
- [ ] Extend `TestNoSpecWritesModeKeys`'s retired-phrase guard to `construct/datatype/*.md`, `construct/adapted/superpowers-writing-plans/SKILL.md`, `cmd/tl-lead/help.md` and `construct/local/fix/SKILL.md`, so it covers every creator's instructions.

### Task 5: Docs, verify, close
- [ ] Update `atlas/workflow/data-artifacts.md:41-46`, `universal-review.md:30-32` ("remaining writers" → none) and `tl-skills.md:146`, plus the code comments named in the map.
- [ ] Run `go test ./pkg/... ./cmd/sdlc/internal/... ./cmd/tl-lead/ ./cmd/datatype/ ./cmd/weave/`, then the recorded suite, then close.
- [ ] Send the landed revision to pair:3 and parley.nvim:3.

## Done-when mapping
- No mode keys in new issues, projects or tl-lead side docs, with absence asserted → Tasks 2 and 3.
- No datatype rows, nothing in Go reads them, the vocab list reduced to a legacy reader → Tasks 2 and 4.
- Legacy fixtures parse and survive the rewrite verbs byte for byte → the Task 2 round-trip test, the Task 3 migrate test, and the existing index/rebase/link tests.
- The TL's scope additions (the two notices, ChatSections in tl-lead, legacy provenance) → already on the branch, plus Task 1.
