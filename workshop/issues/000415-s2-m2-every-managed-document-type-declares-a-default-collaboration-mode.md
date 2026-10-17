---
id: 000415
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '39a5c20b280efc9f68e14d7abd2b4e446663d4fb' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-16
