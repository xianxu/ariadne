---
id: 000575
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '96500b38e049df6313439b4f2626b1e06e3ed7ce' # card fields mirrored from issue-cards; edit via sdlc
---

# universal-review: one collaborative editing mode with a @@chat@@ section carve-out; mode keys become read-only aliases

## Problem

## Spec

**Pivot (operator, 10-23): one collaborative editing mode, plus a per-section chat carve-out.** Replaces the three
mode keys of ariadne#562 (collaboration-, presentation-, processing-mode: 54 combinations, too many). Source:
ops `session.side/2026-10-23-03-mode-simplification.md`. Rewrite universal-review's `collaboration-modes.md` (and
`round.md` where it depends on modes) to this model:

- **Direct edits by the operator are content**, never instructions.
- **A `🤖[…]` (or `🤖<sel>[…]`) is a question or a command; the agent decides.** A question gets a side doc and the
  marker becomes `→ [discussion](…)`. A command changes the document (or has a side effect) with no side doc; if the
  agent needs clarification on a command, it asks in a side doc. When unsure, treat it as a question (a trailing `?`
  is a strong signal; an imperative is a command). `🤖[fold this discussion in]` is the bridge.
- **Presentation:** always highlight, only for the agent's edits (the last machine edit, as now).
- **Processing:** inferred from the instruction; the agent may load a named skill (copy edit, ghost write, …) when
  the instruction calls for it. No processing key and no standing `processing-instruction:`; the operator repeats
  instructions, and a per-user definition ("what copy edit means for this user") can be extracted later.
- **Scope of an edit:** (s1) the current or preceding paragraph; (s2) wider only when asked ("edit this section");
  (s3) a marker before the first heading applies to the whole document unless it says otherwise.
- **`discussion:` frontmatter** keeps listing side docs; removing the last link to one removes it from the list.
- **Chat carve-out:** `@@chat@@` on the line above a heading makes that section chat: its numbered top-level items
  are operator inputs and always get a side-doc reply (`→ [answer](…)`). Side docs themselves are ordinary
  collaborative documents.
- **Legacy keys stay readable, never break a document:** `collaboration-mode: chat` (document-wide or the old
  per-section map) maps to the chat carve-out for the named scope; `presentation-mode`/`presentation`,
  `processing-mode`, `processing-instruction` and `process-instruction` are ignored with a one-line deprecation
  notice. Nothing new writes them.

## Done when

- The spec states the model above, the `@@chat@@` grammar, the scope rules and the legacy mapping, with examples
  (fixtures) that pair and parley pin.
- No spec text asks a creator or filler to write mode keys.

## Plan

- [ ]

## Log

### 2026-10-23
