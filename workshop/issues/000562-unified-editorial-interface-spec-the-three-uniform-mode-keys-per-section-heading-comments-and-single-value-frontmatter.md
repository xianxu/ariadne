---
id: 000562
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7b8dacc967fba65b2ca2b79b8ab8ca01e7b2bc7a' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-22
