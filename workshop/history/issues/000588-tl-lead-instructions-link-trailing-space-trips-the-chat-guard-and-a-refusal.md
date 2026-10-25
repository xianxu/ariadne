---
id: 000588
status: done
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-24
estimate_hours:
card_mirror: '9d8fea3df468dba7fd214d7872c68a6c6c355536' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T21:18:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "cad72865", done: "18cb1fb1"}
actual_hours: 0.89
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

- [x] `linkBullet` keeps the ask's bytes; tests for a trailing space, a tab and a CRLF line
- [x] ~~`linkInstruction` writes index and link in one `Apply`~~ (revised below): a refused link puts the
      frontmatter back; tests: a stale link and a lock between the writes, and the chat guard taking a trailing-space
      and a tab ask
- [x] the fake models apply advancing the whole-file baseline (TestFakeApplyAdvancesBaseline) and couch's chat guard
- [x] pair#862 filed unclaimed: a whole-file apply should advance the baseline only where it edited
- [x] targeted tl-lead tests, close

### Revisions

- 2026-10-24: **one Apply dropped; two scoped writes and an undo instead** (TL approved). In pair's
  sessionsync `prepareWrite`, an apply is a write with no section, and `shadowAfter` with no section
  returns the whole new file as the diff baseline. So one Apply would have wiped the operator's unread edits
  elsewhere from the TL's diff, breaking #426. tl-lead's fake `Apply` never touched the baseline, which hid it
  (ARCH-MOCK). The index entry is still written first and the link second (#424). When the link's write is refused,
  `unlist` puts the frontmatter back byte for byte while it still holds only link's own write. After any other
  frontmatter edit it reconciles the orphan entry away. A refusal of the undo too (the lock taken between the
  writes) names the lingering entry, which the next reconcile drops. The shared index-composer step is dropped
  with the single apply: nothing needs it. Design item 1 stands, checked against pair's
  `editorial/edits.go` (`selectChatItems`): the guard takes only a pure insertion at the item's end, before any
  trailing `\r`/`\n`.

## Log




- 2026-10-24: closed — Round-2 cosmetic Minor fixed: rewrapped the linkInstruction and test doc comments (comment-only). go vet and go test -run TestIndex|TestLink|TestInstructions|TestFake ./cmd/tl-lead/ pass. (fixed-then-ship); review verdict: SHIP
- 2026-10-24: closed — Close review Minors fixed: linkInstruction comment names the unlist byte restore with reconcile as fallback; new end-to-end case in TestIndexLinkRefusedLeavesFile drives a chat-guard refusal (fake refuseAt) through link and the undo, file byte-identical and output names the unlist; help.md explains the lingering-entry message. go test ./cmd/tl-lead/ passes. (fixed-then-ship); review verdict: SHIP
- 2026-10-24: closed — go test ./cmd/tl-lead/ passes: TestLinkBullet (trailing space, tab, CRLF keep the ask bytes, CR stays last), TestIndexLinkRefusedLeavesFile (stale link leaves the file byte-identical incl. index; frontmatter edit between writes reconciles the entry away; lock between writes names the lingering entry, next reconcile drops it), TestIndexLinkUnderChatGuard (fake models pair#838 guard; trailing space/tab links, trimmed form refused, refusal leaves file), TestFakeApplyAdvancesBaseline (fake apply advances the whole-file baseline as pair prepareWrite/shadowAfter do), TestIndexKeepsUnreadEdits unchanged. Guard rule checked against pair cmd/internal/editorial/edits.go selectChatItems.; review verdict: SHIP
### 2026-10-23
- Claimed and started from ariadne:5 after #579 landed; plan to the TL.
- 2026-10-24: the one-Apply design was found to wipe the TL's unread diff (see Revisions), and the compensation
  was approved. The fake now models the baseline and the chat guard. pair#862 is filed, unclaimed. The CRLF
  case is tested in `TestLinkBullet` only: `ReconcileIndex` refuses a file with CRLF lines (an existing limit),
  so a CRLF session file can't link through the index at all.
- 2026-10-24: close review round 1 returned SHIP with three Minors, fixed in this round. `linkInstruction`'s
  comment now names unlist's byte restore, with the reconcile as fallback. A chat-guard refusal now runs
  end to end through `link` and the undo (`refuseAt` on the fake; TestIndexLinkRefusedLeavesFile). help.md says
  what "stays listed without its link" means and what to do.
