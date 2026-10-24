---
id: 000581
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
severity: medium
card_mirror: 'f0e9afe190e8805fb42d53b55845f7193dd39ab9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T18:02:32-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "901eacd9", done: "7891c8f2"}
---

# sidedoc ChatItems: an invalid answer-link destination reads as answered

## Problem

## Spec

Reported by pair:3 during pair#838's integration (2026-10-23). `pkg/sidedoc` `answeredItem`
(`→\s*(\[answer\]\([^)]*\)\s*$|done:)`, chat.go) treats a syntactically invalid Markdown destination as answered:
the legacy item `- Why? → [answer](broken path)` reads as answered, and so does an empty `→ [answer]()`. The item
is then skipped forever, so an operator question goes unanswered without any signal. The writer already refuses
such targets: tl-lead `linkBullet` rejects empty targets, whitespace and parentheses. Pair's earlier guard required a real
no-whitespace destination and let the first real answer be appended.

Fix: the reader accepts what the writer can produce, a nonempty destination with no whitespace or parentheses
(`\[answer\]\([^\s()]+\)`), and the rule is stated once, shared by `linkBullet`'s refusal and `answeredItem`
(ARCH-DRY). collaboration-modes.md §3 names the destination rule. pair keeps the shared oracle (no local fork) and
picks up the fix at its landed revision.

**Scope add (TL, 10-23, from parley.nvim:3; parley.nvim#419 and pair#838 pin it).** A closed but unreadable leading
header is projected as body: `ChatSections` parses headings and tags from line 1 when `frontmatter()` refuses the
block (a spaced closing fence, CR line endings, a document marker, YAML that does not parse). A `@@chat@@` tag above
an ATX heading inside that header then makes the following prose chat-owned. Rule: tag and heading declarations
inside any closed leading header (`FrontBlock`) are ignored; the unread-header notice stays; markers in the body
after it are ordinary commands and questions.

## Done when

- `TestChatItems` rows: `→ [answer](broken path)` and `→ [answer]()` are open; `→ [answer](a/b.md)` is answered;
  a link inside a fence still never answers.
- tl-lead's refusal and the reader use one definition of a valid answer target (a test pins that they agree).
- `TestChatSections` rows: a closed unreadable header holding `@@chat@@` and `## Notes` yields no section, no tag,
  no heading, the unread-header notice, and headings after the header keep their real line numbers.
- Two new exported fixture cases (`chat-answer-invalid`, `frontmatter-unread-tag`) pin both rules for pair and
  parley; `exportedCases` grows by two (an additive manifest change, no version bump).
- pair:3 and parley.nvim:3 are told the landed revision.

## Plan

- [ ] `sidedoc.ValidAnswerTarget(target string) bool` (nonempty, no `unicode.IsSpace`, no parens). `answeredItem`
      captures the destination (`→\s*\[answer\]\(([^)]*)\)\s*$`) and the function decides; `done:` unchanged.
      tl-lead `linkBullet` refuses with `!sidedoc.ValidAnswerTarget` (ARCH-DRY). Tests: ChatItems rows; a tl-lead
      test that every target `linkBullet` accepts reads back as answered and every refused one reads open.
- [ ] `chatSections`: when `FrontBlock` is closed and `frontmatter()` refused it, the body (for headings, tags and
      the code mask) starts after the block, so nothing inside is a declaration. ChatSections table rows.
- [ ] Fixture cases `chat-answer-invalid` (a `@@chat@@` item ending `→ [answer](broken path)` gets its first real
      link) and `frontmatter-unread-tag` (spaced closing fence over a tag and heading; the body marker opens a side
      doc; the reply carries the notice); add both to `exportedCases` and README's table; regenerate `export.json`.
- [ ] collaboration-modes.md §3 states the destination rule and the header rule.
- [ ] Targeted tests (`./pkg/... ./cmd/tl-lead/`), recorded suite, close, land; send the revision to pair:3 and
      parley.nvim:3.

ARCH-EVOLVE: no persisted format changes; the export manifest only gains members. ARCH-SECURE: the document is
untrusted; an unreadable header degrades to "no declarations" plus the notice, never chat ownership.
ARCH-CONSTRAINTS: still one linear pass; the counted N/2N tests keep guarding it. ARCH-ORDER/FUNERAL: pure parsing,
holds no state and creates nothing durable.

## Log

### 2026-10-23
