# universal-review: one collaborative editing mode — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the universal-review skill to the operator's single collaborative editing mode, with a per-section `@@chat@@` carve-out, scope rules s1–s3 and a legacy-key mapping. Ship export v4 and the fixtures that pair#838 and parley.nvim#419 pin.

**Architecture:** The skill under `construct/local/universal-review/` is the single source; `pkg/reviewexport` builds its `export.json` manifest and checks the examples. Prose changes go to the skill files. The manifest moves to v4: the `mode` role becomes a flagless `brief` role, and the new examples are exported as fixtures. The examples checker learns two things. A side doc a round creates carries no mode keys. An existing document's legacy keys survive byte for byte.

**Tech stack:** Markdown (the skill), Go (`pkg/reviewexport`: manifest builder and example checker).

**Out of scope (filed):** ariadne#576 removes ariadne's Go writers of the mode keys and the datatype `collaboration-mode` rows. The `superpowers-writing-plans` header template that writes `collaboration-mode: discussion` is added to #576's list. pair#838 and parley.nvim#419 adopt this spec in their own repos.

---

## The model (what the rewritten spec says)

Source: ops `session.side/2026-10-23-03-mode-simplification.md`, decided "go" 10-23.

1. **Direct edits by the operator are content**, never instructions. A submit with no marker and no new chat item gets "read your edits; no asks" (as today).
2. **A ready `🤖[…]` or `🤖<sel>[…]` is a question or a command; the agent decides.**
   - **Question:** open a side doc, and the marker becomes `→ [discussion](…)`. Every question branches, however small.
   - **Command:** a change to the document or a side effect, with no side doc. A command the agent needs clarified asks in a side doc, linked the same way.
   - **Unsure:** treat it as a question. A trailing `?` is a strong question signal, and an imperative is a command.
   - **Bridge:** `🤖[fold this discussion in]` (any fold-in wording) beside a `→ [discussion]` link is a command.
   - **Decisions:** an approval or decision ("ok", "go with 2", "first in the section") is a command. It is recorded as one `Decided (operator): …` line and carried out, unchanged from today.
   - **Instructions that only act** (file this, tell X) are commands too, as today.
3. **Presentation:** always highlight, for the agent's own edits only (the last machine edit, as the adapter shows it now). There is no direct or propose setting; the proposal machinery is retired from the spec. pair#838 retires it in Couch.
4. **Processing** is inferred from the instruction. The agent may load a named brief (`briefs/copy-editing.md`, …) when an instruction calls for it ("copy edit this section"). No document key selects one, and there is no standing instruction key.
5. **Scope of an edit:**
   - **s1:** the text the marker refers to: its anchor, or else the current or preceding paragraph (today's marker rule).
   - **s2:** wider only when the instruction says so ("edit this section", "copy edit the document").
   - **s3:** a marker before the first heading applies to the whole document unless it says otherwise.
   - The old `scope` and `frontier` flags are gone, and so is the settled region above the topmost marker.
6. **`discussion:` frontmatter** keeps listing side docs, with today's index rules: the operator removing the last link to one drops its entry.
7. **Chat carve-out:** a `@@chat@@` line directly above an ATX heading makes that section chat. Its numbered top-level items are operator inputs, and each always gets a side-doc reply: ` → [answer](…)` is appended to the item. Follow-ups, deletions and relinking work as in today's chat. A `🤖[…]` inside a chat section is an ordinary marker under rule 2. Side docs are ordinary collaborative documents.
8. **Legacy keys are read, never fatal, never written:**

   | Legacy key or form | Effect |
   |---|---|
   | `collaboration-mode: chat` (scalar) | the whole document is a chat section: every top-level list item, in any section |
   | legacy section map `collaboration-mode: {chat: [names], default: …}` | each named section is chat (today's heading-name matching) |
   | heading comment `<!-- collaboration-mode: chat -->` | that section is chat |
   | `collaboration-mode: discussion` / `collaboration` (scalar, map default or comment) | none: the single mode applies. One notice line names the retired value |
   | `presentation-mode`, `presentation`, `processing-mode`, `processing-instruction`, `process-instruction`, and the same keys in heading comments | ignored, with one notice line per document naming the keys |
   | an unknown or malformed value of any of these | ignored with a notice. The round no longer fails over a mode key |

   A legacy chat section keeps its old item grammar (`-`, `*` and `N.` items), so documents already in chat behave as before. A new `@@chat@@` section takes `N.` items only, because the operator's chat protocol is numbered. A round never deletes or rewrites a legacy key line; it reads around it.

9. **`@@chat@@` grammar:**
   - The tag is a line consisting of exactly `@@chat@@`, allowing surrounding spaces. It sits on the line directly above an ATX heading, outside fenced code.
   - The section runs to the next heading of the same or a higher level, including its subsections.
   - A tag separated from its heading by a blank line, or followed by a non-heading, is ordinary text. The round names it in the reply ("`@@chat@@` at line N is not directly above a heading").
   - Other `@@tag@@` lines are parley's (outline labels and file refs) and are untouched by review.
   - Precedence: a `@@chat@@` tag, a legacy heading comment and a legacy map all mean "chat", so there is no conflict between them. Their union is chat.

## Decision D (decided by the operator, 10-23): D2

The unified model applies everywhere, issues, plans and projects included. Legacy `collaboration-mode: collaboration` is retired like `discussion`, as the rule-8 table says: a question gets a side doc, an instruction acts, and "fold in" collapses one. The `collaboration` example's `after/` opens a side doc and prints the retired-key notice. The AGENTS §1 rewrite and a side-doc link guard go to #576 or a sibling, not #575.

## Consumers of `export.json` (fleet grep, 10-23)

| Consumer | Reads | Version gate | On v4 |
|---|---|---|---|
| parley.nvim `scripts/sync-review-protocol.py` (build-time sync) | round, grammar, collaboration-modes, adapter, mode (+ flags), fixture | allowlist `version in (1, 2, 3)` (l.31); a test asserts 4 is rejected | refuses before any read or write. parley.nvim:3 confirmed it on a v4-edited copy: the packaged bytes stay unchanged, so installs keep their bundled snapshot. That snapshot is **v2** on parley's origin/main (the v3 one was on an unlanded branch; narrowed #418 keeps v2). #419 migrates consumer and snapshot together from v2 to v4, never assuming v3 shipped |
| parley.nvim runtime (`lua/parley/skills/review/`) | its packaged copy only; `mode.lua` reads flags from vendored `modes/*.md` | none (`export_version` is informational) | unaffected: it never reads ariadne's export at runtime. Caveat from parley:3: hand-copying v4 files over the bundle bypasses the sync guard |
| parley.nvim tests (`review_move_fixtures.lua`, `review_protocol_export_test.py`) | the vendored snapshot's move fixtures | pin the vendored version | unaffected until parley re-syncs (#419) |
| pair, echo, brain, tools, metis, nous, ops | none (ops has a design-doc mention only) | none | none. pair#838's fixture tests are new and will pin v4 |

## Export v4 (exact member shape, sent to parley.nvim:3 and pair)

- `Schema` is unchanged, and `Version` becomes 4.
- The roles are, in order: `round`, `grammar`, `collaboration-modes`, `adapter`, `brief`, `fixture`.
- `brief` members are `{"role":"brief","name":"copy-editing","path":"briefs/copy-editing.md","bytes":…,"sha256":…}`. They carry no `order`, `scope` or `frontier`, and are sorted by name.
- Brief files have flat frontmatter `name:` (equal to the basename) plus a prose brief. Any other key, or a flag, refuses the build, so v3 flags can't slip back in.
- `fixture` members are every file of each exported case, as in v3. v4 exports the move cases plus every new case below (`chat-tag`, `chat-tag-misplaced`, `ask-question`, `ask-command`, `ask-command-unclear`, `ask-unsure`, `scope-paragraph`, `scope-section`, `scope-document`, `legacy-chat-scalar`, `legacy-chat-map`, `legacy-chat-comment`, `legacy-keys-ignored`). The exported set is listed in one place, `exportedCases` in `export.go`, so the checker and the builder agree.
- The `collaboration-modes` member keeps its role name and path. The file's title changes, but the path doesn't, because pkg/sidedoc, tl-lead and sdlc cite `collaboration-modes.md §4/§5`.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Manifest` / `Member` (v4, `brief` role) | `pkg/reviewexport/export.go` | modified |
| `Brief` / `ParseBrief` (was `Mode` / `ParseMode`) | `pkg/reviewexport/export.go` | modified |
| `exportedCases` | `pkg/reviewexport/export.go` | new |
| `checkCase` side-doc frontmatter rule | `pkg/reviewexport/examples_test.go` | modified |
| `chatSections` (tag, legacy scalar, map and comment → chat extents) | `pkg/reviewexport/chat.go` | new |

- **Brief / ParseBrief**: parses `briefs/*.md`, requiring flat frontmatter with exactly `name` equal to the basename. It replaces `Mode` and drops `order`, `scope`, `frontier` and the defaults map. Tests: `export_test.go` `TestParseBrief` covers an accepted brief, a name mismatch, a leftover `scope:` and a leftover `order:` (each refused).
- **exportedCases**: the one list of example cases exported as fixtures. Build and `TestExportCurrent` read it (ARCH-DRY, one source). Today that list is hard-coded as "move cases".
- **checkCase side-doc rule**: a side doc present in `after/` but not in `before/` (created by the round) has frontmatter keys exactly `type`, `parent`, `marker`. Any mode key there is an error ("new side doc writes retired mode key X"). A side doc present in both keeps its bytes: a legacy one with all three keys passes unchanged. This replaces today's "collaboration-mode must be discussion" and the canonical-defaults check (`month_examples_test.go` `TestMonthSideDocCanonicalDefaults` becomes `TestMonthSideDocNoModeKeys`).
- **chatSections**: given a document's text, returns the chat sections' line extents and any misplaced-tag lines. It covers a `@@chat@@` tag directly above an ATX heading outside fences, a legacy scalar `collaboration-mode: chat`, a legacy map's names (today's matching rule) and legacy heading comments. It is pure, with no IO, and is tested by `chat_test.go` table rows. The checker uses it so the chat fixtures are validated against the grammar the spec states; it is the reference oracle pair and parley can copy. tl-lead's `instructions.go` keeps its own heading-name lookup for the session file. Adopting `@@chat@@` there is #576's scope, so it isn't duplicated here.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `TestExportCurrent` (regenerates and compares `export.json`) | `pkg/reviewexport/export_test.go` | modified | the skill directory on disk |
| `TestExamples` | `pkg/reviewexport/examples_test.go` | modified | `examples/*` trees on disk |

Both read the real skill tree through `os.DirFS`, as today, and there's no new external dependency. ARCH-MOCK, ARCH-ORDER, ARCH-FUNERAL: N/A. The skill is static text plus a generated manifest. It holds no state between events and creates nothing durable beyond tracked files; the fixtures are its own tests. ARCH-EVOLVE: the manifest format changes incompatibly on purpose (v3 → v4), and the version gate makes old readers refuse, as the consumer table shows. Unknown fields: v4 readers (parley/pair) should ignore unknown member fields, which the spec states. ARCH-SECURE: `marker:` stays a quoted YAML string, as today; there's no new untrusted input. ARCH-CONSTRAINTS: N/A (static docs).

## Tasks

### Task 1: Briefs and export v4
**Files:** `construct/local/universal-review/modes/*.md` → `briefs/*.md` (git mv, frontmatter reduced to `name:`, and each brief's sentence about `presentation-mode` removed); `pkg/reviewexport/export.go`; `export_test.go`; `export.json` (regenerated).
- [ ] Failing tests: `TestParseBrief` (rows above). `TestBuildOrder` expects the `brief` role after `adapter`, sorted by name. `TestV4ExportHasNoModeFlags` checks that no member has order, scope or frontier and that no `mode` role remains. `TestBuildFixtures` expects the exported cases from `exportedCases`.
- [ ] Implement `Version = 4`, `Brief`/`ParseBrief`, the `brief` role, `exportedCases` and Build over `briefs/*.md`; update the package comment's version history line.
- [ ] Move the files. Regenerate `export.json` the way `TestExportCurrent` names (its failure prints the command).
- [ ] Run `go test ./pkg/reviewexport/ -run 'TestParseBrief|TestBuild|TestV4|TestExportCurrent'`, expect PASS, then commit `#575: export v4: flagless briefs`.

### Task 2: chatSections oracle
**Files:** create `pkg/reviewexport/chat.go` and `chat_test.go`.
- [ ] Failing table test with these rows:
  - a tag directly above `## Ops` → chat through the next `##`, including `###` subsections;
  - a tag, a blank line, then a heading → no chat, and the misplaced line reported;
  - a tag inside a fence → no chat;
  - a tag above a fenced "heading" → no chat;
  - `collaboration-mode: chat` scalar → the whole document;
  - a legacy map naming "Operator instructions" → that section, matched by today's normalization (case, whitespace, trailing parenthesis);
  - a legacy comment `<!-- collaboration-mode: chat -->` → that section;
  - a tag plus a comment on the same heading → one chat section, with no error;
  - `collaboration-mode: discussion` → none;
  - an unknown value `debate` → none, with a notice.
- [ ] Implement it as pure string processing, reusing the existing heading and fence scanning helpers in `examples_test.go` by moving them into `chat.go`. Run the tests, expect PASS, and commit.

### Task 3: Rewrite the spec prose
**Files:** `collaboration-modes.md` (retitled "Collaborative editing"), `round.md`, `SKILL.md`, `adapters/file-edit.md` (presentation and proposal passages only).
- [ ] `collaboration-modes.md`:
  - §1 The model: rules 1–6 above, with the question/command table and today's Decided and instructions-that-only-act tables, which carry over.
  - §2 Scope s1–s3.
  - §3 The chat carve-out: the grammar, items, follow-ups, deletions and relinking (today's chat §3 text, adapted to the tag and to `N.` items).
  - §4 Side docs and answer docs: today's §4, with the frontmatter block reduced to `type`, `parent`, `marker` and the "always written out" bullets removed.
  - §5 Moves: unchanged.
  - §6 States and actions: rows renumbered for the new model. The question/command rows replace the discussion/collaboration rows, the chat rows reference `@@chat@@`, and the legacy rows are added.
  - §7 Legacy keys: the table in rule 8, plus "nothing new writes them".
- [ ] `round.md`:
  - §1 Inputs: no settings to read beyond the chat tags and the legacy keys (notices).
  - §2 Briefs: load a named brief when the instruction calls for it.
  - §3 Scope: s1–s3. The scope/frontier flags, the presentation table and the propose paths are removed.
  - §4 Routing: question / command / chat item.
  - §5 Markers: keep it, minus the presentation references.
  - §6 Reply: add the notices, minus the mode-failure errors.
- [ ] `SKILL.md` describes the new model and drops "using its editorial settings". Its description still triggers on `submitted:` and on legacy `process-instruction:` documents.
- [ ] Commit `#575: spec: one collaborative mode with a @@chat@@ carve-out`.

### Task 4: Retire and convert the examples
**Files:** `examples/*`, `examples/README.md`, `pkg/reviewexport/examples_test.go`, `editorial_fill_test.go`, `month_examples_test.go`.
- [ ] Failing checker test in `TestCheckCaseCatches`: a new side doc carrying `collaboration-mode: discussion` → the error "new side doc writes retired mode key", and a pre-existing legacy side doc with three keys left unchanged → no error.
- [ ] Update `checkCase` and the month tests (the side-doc rule above).
- [ ] Mechanical fixture sweep (script in the Log): in every case, strip the mode-key lines from side docs that exist in `after/` but not in `before/`. Pre-existing documents keep their bytes.
- [ ] Retire the cases whose behaviour no longer exists: the `editorial-fill-*` cases (seven of them; fill no longer writes mode keys), `error-unknown-mode` and `error-conflict` (the round no longer fails over modes), and `propose-*` and `untracked-propose-refused` (no propose presentation). Delete them, and delete `editorial_fill_test.go`. `collaboration` (answers in place) becomes `ask-question`'s legacy twin: its `after/` now opens a side doc, with a notice that `collaboration` is retired.
- [ ] Convert `editorial-heading-chat` and `chat` into legacy-mapping cases (comment and map), with their replies gaining the notices.
- [ ] New cases (each with `case.md`, `before/` and `after/`, dated 2026-10-23, side docs in `YYYY-MM/`), listed in `exportedCases`:
  - `chat-tag`: a `@@chat@@` section with a new item, a linked item with a follow-up, and a `🤖[…]` question inside it.
  - `chat-tag-misplaced`: a blank line between the tag and the heading → no chat, and the reply names it.
  - `ask-question`: `🤖[why …?]` → a side doc and its link.
  - `ask-command`: `🤖[give an example]` → the paragraph edited, with no side doc.
  - `ask-command-unclear`: `🤖[add the numbers]` with two readings → a clarifying side doc.
  - `ask-unsure`: `🤖[the retry story feels off]` → a side doc.
  - `scope-paragraph`: an unanchored command edits only the preceding paragraph.
  - `scope-section`: `🤖[copy edit this section]` edits its section only, loading the copy-editing brief.
  - `scope-document`: a marker before the first heading edits the whole document.
  - `legacy-chat-scalar`: `collaboration-mode: chat` makes every top-level item an ask.
  - `legacy-chat-map` and `legacy-chat-comment`: as described above.
  - `legacy-keys-ignored`: all the retired keys plus an unknown value → the round proceeds, with one notice line, and every key line is kept byte for byte.
- [ ] Have `TestExamples` also run `chatSections` on every case's submitted document: each item that gained a `→ [answer]` link lies inside a chat extent, and no chat item lacks one.
- [ ] Update `examples/README.md`: which cases are legacy, what the new ones show, and which ones pair and parley pin (the exported set).
- [ ] Run `go test ./pkg/reviewexport/`, expect PASS, and commit.

### Task 5: Prose guards and the atlas
**Files:** `examples_test.go` `TestCollaborationModesProse`, `export_test.go` contract tests, `cmd/tl-lead/main_test.go` (help wording, only if its quoted phrase changes), `atlas/workflow/universal-review.md`, `README.md` (its `presentation-mode: propose` mention).
- [ ] `TestCollaborationModesProse` gets the new row count and new phrases: "Direct edits by the operator are content", "treat it as a question", "`@@chat@@`", "nothing new writes them". Its retired-phrase list forbids "Written defaults", "presentation-mode: propose", "processing-mode:" outside §7, and "write all three keys". That list is the Done-when guard that no spec text asks a creator or filler to write mode keys. It applies to `collaboration-modes.md`, `round.md`, `SKILL.md`, `adapters/file-edit.md` and the README.
- [ ] Update the contract tests' expected phrases where they quote presentation or propose.
- [ ] Atlas: rewrite the mode description in `universal-review.md` to the new model, with a pointer to #576 for the remaining writers.
- [ ] Run `go test ./pkg/reviewexport/ ./cmd/tl-lead/ ./cmd/weave/ -run 'Review|Example|Collaboration|Editorial|Month|Export|Skill'`, expect PASS, and commit.

### Task 6: Verify and hand off
- [ ] Run the recorded suite (`sdlc local-check run --name suite`), then close.
- [ ] Send parley.nvim:3 and the pair owner (via the hub) the v4 member shape and the exported case list, with the landed commit.

## Done-when mapping
- The spec states the model, the `@@chat@@` grammar, the scope rules and the legacy mapping, with exported fixtures → Tasks 3, 4 and 1 (the exported set).
- No spec text asks a creator or filler to write mode keys → the Task 5 retired-phrase guard, plus the checker rule that new side docs carry no mode keys (Task 4).
