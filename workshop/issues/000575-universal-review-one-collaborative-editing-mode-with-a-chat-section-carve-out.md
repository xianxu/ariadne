---
id: 000575
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4ef3e7eeffe1e4033062400878941320b13b2020' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T11:52:48-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "87b18cbb", done: "e3ecd833"}
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

Durable plan: [000575 plan](../plans/000575-universal-review-one-collaborative-editing-mode-with-a-chat-section-carve-out-plan.md) (approved by the TL 10-23, (a) to (f); decision D2 by the operator).

- [x] Task 1: Briefs and export v4 (`briefs/`, `brief` role, `exportedCases`)
- [x] Task 2: `chatSections` oracle
- [x] Task 3: Spec prose (collaboration-modes.md, round.md, SKILL.md, file-edit adapter)
- [x] Task 4: Examples: retire, convert, add the 13 new cases; checker side-doc rule
- [ ] Task 5: Prose guards (the Done-when retired-phrase test) and the atlas
- [ ] Task 6: Recorded suite, close, hand the v4 shape to parley.nvim:3 and pair

## Log

### 2026-10-23
- Claimed after #573 landed. The plan was drafted while #573's merge ran; the TL approved (a) to (f), and the operator decided D2: the unified model applies everywhere. Consumer grep: parley.nvim is the only reader of export.json. Its build-time sync accepts v1 to v3 and refuses v4 before writing anything, and its shipped bundle is v2, so #419 migrates v2 to v4. Filed #576 for the Go writers and datatype rows.
- Task 1: export v4 with five flagless briefs. `free-form` is retired: it only meant "follow processing-instruction". A leftover v3 flag or unknown key refuses the build (`TestParseBrief`). `TestExportCurrent` stays red until Task 4 adds the exported cases.
- Task 2, deviating from the plan on purpose (ARCH-DRY): the scanner lives in `pkg/sidedoc/chat.go` as `ChatSections`, not in `pkg/reviewexport`. pkg/sidedoc is the shared side-doc library tl-lead and sdlc already import, so #576's tl-lead adoption reuses it instead of growing a third heading scanner. Headings come from goldmark (a fenced "heading" is code). Legacy forms that meant chat (scalar, map names, heading comment) still make chat sections. Every other mode key or value is a notice, never an error.
- Task 3: rewrote collaboration-modes.md (title "Collaborative editing"; path kept for its citers), round.md and SKILL.md. The adapter keeps Couch's propose mechanics, marked as legacy until pair#838, because they describe what Couch does today.
- Task 4: the checker's rule is that a side doc a round creates has only type/parent/marker (plus a later index), and pre-existing side docs keep their legacy keys. 13 new cases are generated by a script (scratchpad gen/cases.py), and 16 cases retired (editorial-fill-*, error-*, propose-*, untracked-propose-refused). `collaboration` became `legacy-collaboration` under D2; `editorial-heading-chat` became `legacy-chat-comment`; `chat` stays as the legacy-map case, since tl-lead's tests read it. `checkChat` cross-checks every case with `ChatSections`. Mutations: dropping an answer link → "a @@chat@@ item left unanswered"; dropping the misplaced-tag reply line → "does not name the misplaced tag".

