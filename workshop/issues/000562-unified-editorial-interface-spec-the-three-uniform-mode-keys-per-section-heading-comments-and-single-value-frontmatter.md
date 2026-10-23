---
id: 000562
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8b2b057be97f57a9c8e7a730e86af252384b07cd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T22:48:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6359504e", done: "8c7cc7bf"}
---

# unified-editorial-interface: spec the three uniform mode keys, per-section heading comments and single-value frontmatter

## Problem

## Spec

**Settled with the operator (10-22, design chat in brain `session.side/2026-10/2026-10-22-01-editorial-modes.md`):**
- **Uniform names:** `collaboration-mode` (where an answer goes: discussion / collaboration / chat),
  `presentation-mode` (how edits are shown: highlight / direct / propose), `processing-mode` (how an edit is made:
  the review modes in `ariadne/construct/local/universal-review/modes/`: copy-editing, developmental, fact-check,
  free-form, line-editing, proofreading; unique prefixes allowed). The old keys `presentation` and
  `process-instruction` stay readable as aliases. `process-instruction`'s free text splits out into
  `processing-instruction`.
- **Per section:** all three can be set for one section by an HTML comment directly under its heading:
  `<!-- collaboration-mode: chat; processing-mode: copy-editing -->`. Couch protects that line as operator-owned;
  parley conceals it as a badge.
- **Frontmatter:** each key is one value covering the whole document; the per-section map form
  (`collaboration-mode: {chat: ["Operator instructions"]}`) is retired.
- **Precedence:** section comment, then document frontmatter, then the built-in default.
- **Typeahead (required):** on a line next to a heading, `<!-- ` offers the three keys, `key: ` offers that key's
  values, `; ` offers the keys not yet used, and accepting the first key inserts the closing ` -->`. Values come from
  the modules that enforce them (extending parley's `lua/parley/frontmatter_enum.lua`), never a second list.


Project `unified-editorial-interface` (ops). **Spec first:** update the universal-review spec to the settled model:
`construct/local/universal-review/collaboration-modes.md`, `round.md` (and wherever `presentation` /
`process-instruction` are defined): the three uniform keys and their values, the aliases and their deprecation note,
`processing-instruction`, the heading-comment grammar (exact line pattern, key order free, `; ` separator, one comment
per heading, line directly under the heading), single-value frontmatter, precedence, and the migration rule for
documents still using the per-section map (read as before with a deprecation warning until migrated; the migration
writes one heading comment per mapped section). This is the contract pair and parley implement.


**Side-doc layout (decided 10-22):** `<doc>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md`: a year-month subfolder, with the full date kept in the filename so a file still makes sense when linked, grepped or moved. **No migration:** existing flat names stay valid (readers accept both). Reason: `ops/session.side/` already holds about 270 entries, 261 from October alone. Applies to the naming rule in the universal-review spec (`collaboration-modes.md` §4) and to every side-doc creator (parley, Couch, and tl-lead's side-doc helpers).


**Operator addition (10-22, night):** whenever front matter is filled in for a document (a new session file or side doc, or the first Alt+Return adding missing front matter), it lists **all three mode keys with their default values written out** (`collaboration-mode: discussion`, `presentation-mode: highlight`, `processing-mode: free-form`), as a visible reminder that they can be changed. Merging into existing front matter adds any missing ones the same way and never changes values already set.

## Done when
- New or filled front matter lists all three mode keys with their defaults; existing values are untouched (test).

- The spec documents all of the above with one worked example per mode and the precedence table.
- A machine-readable value list per key exists in one place that pair and parley read (or the spec names each
  enforcing module as the single source), so typeahead and enforcement can't drift.

## Plan

- [ ] Write the canonical contract in collaboration-modes.md and round.md: the three single-value keys/defaults, processing-instruction, readable deprecated aliases, immediately-adjacent heading-comment grammar and operator ownership, section/document/default precedence, legacy-map warning/migration, and required typeahead behavior. Keep every settled Spec decision; no additional editing modes or runtime parser in this issue.
- [ ] Sweep SKILL.md, the file-edit adapter, reference examples, README and atlas pointers for contradictory instructions. New side docs use <doc>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md with correct relative parent/backlinks; old flat docs stay readable/reusable, without mass migration or nested .side directories. Name the existing enforcing value owners rather than adding a second enum list: collaboration parsers/vocab, Pair's sessionsync presentation resolver, and exported modes/*.md metadata; parley's frontmatter_enum consumes these owners.
- [ ] Add worked examples for each collaboration/presentation/processing value and the precedence table, plus new/partial/existing frontmatter fill examples that make all defaults visible while preserving existing choices. Retain deprecated map/alias/flat-path cases as compatibility fixtures. Exercise the contract with fresh-context RED/GREEN review scenarios, including wrong heading placement, override precedence, aliases and non-default values, and side-doc creation from old/new locations. Extend only the existing test-only example checker as needed for the new reference shape; Pair/parley's runtime enforcement and UI tests remain pair#823/parley.nvim#419.
- [ ] Regenerate export.json through pkg/reviewexport, run focused reviewexport/vocab/sidedoc and weave discovery checks, verify export hashes/links, then the recorded suite on the pushed head and SDLC's close review. Proposed smoke: private contract/examples only, no live fleet or :0 stage. Land and report only after these gates.

Sizing: quick-flow contract/docs/fixtures, no planned production Go change and no separate milestone boundaries. Keep the design within the shell; re-plan before adding runtime behavior. Preferred approach is one shared exported contract plus examples and named enforcing owners (ARCH-DRY); a new central configuration API or independent consumer implementation here would duplicate the separately assigned runtime work.

Architecture: ARCH-PURE — specification and test-only checks, no new IO shell. ARCH-PURPOSE — cover all settled axes, defaults, typeahead and side-doc layout, not only renamed keys. ARCH-EVOLVE — old aliases, section maps and flat side docs remain readable with explicit migration/deprecation rules; no automatic destructive rewrite. ARCH-MOCK — no new external service/binary integration. ARCH-CONSTRAINTS — bounded document/export authoring, no new runtime work or performance promises. ARCH-SECURE — mode comments/values remain operator-owned, existing lock/CAS semantics preserved. ARCH-ORDER — this issue adds no stateful runtime; contract examples retain existing submit/write ordering. ARCH-FUNERAL — creates only version-controlled reference artifacts; existing side-doc lifetime and archive/move ownership remain intact.

## Log

### 2026-10-22

TL approved the plan and legacy-fill clarification via f5771e69-00d4-4f32-98e7-60dd9177f766. Preserve legacy alias lines; when canonical fields are missing, materialize their existing effective legacy mode/instruction rather than shadowing them with defaults. Pin both presentation and processing cases in the scenarios. Proceeding to change-code; the shared skill-authoring RED/GREEN checks test the contract separately from the downstream runtime implementations.

Claimed and entered planning in ariadne:5 after #549 landed. Read the settled Spec and current round, collaboration modes, export/checker and value-owner seams; inspected Pair/parley read-only, no peer edits. Current source still teaches composite process-instruction, collaboration-mode maps and flat side-doc creation. This issue establishes their replacement contract; consumers implement it in pair#823/parley.nvim#419. Plan submitted to TL before change-code or contract edits. One compatibility clarification accompanies the plan: when filling frontmatter containing only a deprecated alias (e.g. presentation: propose), writing a canonical default must not silently shadow that existing choice; propose materializing the existing effective value under the canonical key and preserving the alias. Await TL confirmation of this edge and plan approval. No estimate derived and no implementation started.

## Revisions

2026-10-22: TL confirmed legacy-only frontmatter fill preserves effective values: copy presentation into presentation-mode, split the legacy process-instruction mode/text into processing-mode and processing-instruction, and retain old lines unchanged. Absent values receive the stated defaults; already-set canonical values stay untouched. This pins compatibility within the approved scope rather than silently overriding existing operator choices.
