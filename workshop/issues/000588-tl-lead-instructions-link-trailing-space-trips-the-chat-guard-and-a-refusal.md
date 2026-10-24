---
id: 000588
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '9b85da4221e8163c6eafd3b3dd343496b4b6d4b1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T21:18:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead instructions link: trailing space trips the chat guard, and a refusal leaves a half-done write

## Problem

## Spec

**`tl-lead status instructions link` breaks on an item ending in a space, and leaves a half-done write (ops-echo:0,
10-23).** On a `@@chat@@` numbered item whose text ends with a trailing space, `link` trims the space before appending
` → [answer](…)`. Couch's chat guard (pair#838's) reads that as editing the operator's ask and refuses:
`editorial-chat at bytes 808:808: preserve each chat ask; only append its answer link`. By then `link` had already
written the `discussions:` index entry, so the failure left the document half-updated. Likely from ariadne#576's move to
the shared `sidedoc.ChatItems` grammar.

Fix: (1) append the link without changing any byte of the ask (trailing whitespace included); (2) make the operation
all-or-nothing: compose the item edit and the index entry, then write them in one session-sync write (or roll back the
first on failure).

## Done when

- Linking an item that ends in a space (and one ending in a tab) succeeds, leaving the ask's bytes intact (test).
- A refused link leaves the document unchanged, including the `discussions:` index (test).

## Plan

Quick flow (inside the shell; no durable plan).

**Cause.** `linkBullet` (cmd/tl-lead/instructions.go:126) rewrites the item's last line with
`strings.TrimRightFunc(line, unicode.IsSpace)` before appending, so the ask loses its trailing bytes. `linkInstruction`
then writes the `discussions:` index first (`syncIndex` → `WriteFrontmatter`) and the section second (`WriteSection`):
the #424 order, which a refusal of the second write turns into a half-done document.

**Design.**
1. `linkBullet` appends ` → [answer](target)` to the last line as it stands (a trailing `\r` stays last, so a CRLF
   line keeps its ending); the ask's bytes are a prefix of the new line.
2. One write: compose the whole new file from one `Peek` (the frontmatter block with the reconciled index, and the item
   line with the link) and send it as two `apply` records against that text's `fileHash` (`SessionSync.Apply`, all or
   none, #455). The records are proved before sending (`applyRecords` must reproduce the composed text), as
   `lineRecords` does. A stale hash is the existing "list again" refusal (exit 6), never a retry, since positions come
   from the listed read.
   - The index composition is pulled out of `syncIndex` into one pure function both use (ARCH-DRY), keeping its
     refusal of a frontmatter couch would end early.
   - The item line is placed in the file by the section body's single occurrence in the `Peek` text (and the section
     must still equal what `ReadSection` returned for the listed hash); anything else refuses as "changed since list".
   - This retires #424's index-first order: with one compare-and-replace there is no split to order.

**ARCH notes.** PURE: link composition and record building are pure over strings; the IO stays `ReadSection`, `Peek`,
`Apply`. ORDER: one read, one compare-and-replace; the events it can't block (an operator edit, the lock, couch's chat
guard refusing) all leave the file as it was. EVOLVE, FUNERAL, SECURE: no format change, nothing durable, the target
is already validated by `ValidAnswerTarget`.

- [ ] `linkBullet` keeps the ask's bytes; tests for a trailing space, a tab and a CRLF line
- [ ] `linkInstruction` writes index and link in one `Apply`; the shared index composer; tests: one apply and no
      frontmatter or section write on success, and a refused apply (chat guard, stale, lock) leaves the document
      byte-identical, index included
- [ ] targeted tl-lead tests, close

## Log

### 2026-10-23
- Claimed and started from ariadne:5 after #579 landed; plan to the TL.
