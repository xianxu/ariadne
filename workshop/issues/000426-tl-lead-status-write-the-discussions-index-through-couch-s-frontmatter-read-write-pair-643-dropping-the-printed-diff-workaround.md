---
id: 000426
status: codecomplete
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ea121c1d06071fa6aee04e2beb56413480b50402' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T07:33:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "52cc7fee", done: "74c82410"}
actual_hours: 0.08
---

# tl-lead status: write the discussions index through couch's frontmatter read/write (pair#643), dropping the printed-diff workaround

## Problem

ariadne#424 (M2) writes the TL session file's `discussions:` index through couch's whole-file `read --all` / `write --all`, which reset the diff baseline for the whole file. To avoid consuming the operator's unread edits, `tl-lead`'s `syncIndex` decides no-ops by `Peek` and prints the pending diff (`readAllShowingUnread`) before every whole-file read: correct, but noisy. pair#643 (landed, pair 43313c59) added `couch --session-sync read --frontmatter` and `write --frontmatter … --if-hash H`, which move the baseline only for the frontmatter.

## Spec

- `SessionSync` gains `ReadFrontmatter(file) (text, hash, err)` and `WriteFrontmatter(file, bodyPath, hash)` over the new verbs; the fake models a frontmatter-only baseline (a frontmatter read or write resets only the frontmatter's shadow).
- `syncIndex` reads and writes only the frontmatter: `sidedoc.ReconcileIndex` needs the body's links too, so it still reads the body without moving the baseline (`Peek`), but writes only the frontmatter through the CAS; the printed-diff workaround (`readAllShowingUnread`, `windowChanges`, `syncIndexAfter`) is removed.
- The live `TestSessionSyncConformance` checks that a frontmatter read or write leaves an operator's section edit in the diff.
- Precondition: the operator's running Couch has restarted with pair#643 (the TL confirms); until then the verbs refuse.

## Done when

- `instructions link`, `status submit` and `status index` never consume or print the operator's unread section edits, with tests against the fake and the live conformance.
- The printed-diff workaround is gone, and help/atlas describe the frontmatter write.

## Plan

Quick flow: about 40 added code lines, and index.go shrinks.

- [x] **Seam:** `SessionSync.ReadAll`/`WriteAll` become `ReadFrontmatter(file) (block, hash, err)` and `WriteFrontmatter(file, bodyPath, hash)` over `read --frontmatter --json` and `write --frontmatter … --if-hash`. The reply decodes like a section read. No other caller used the whole-file pair, so it goes (ARCH-DRY). `Peek` and `MarkAllSeen` (status seen) stay.
- [x] **Pure split:** `sidedoc.FrontBlock(doc) (block, rest, ok)` has couch's projection: the closed leading `---` block without its final line ending, and the rest verbatim. `syncIndex` composes `block + rest` from the frontmatter read and the peeked body, reconciles it, and writes `FrontBlock(next)`. ReconcileIndex is unchanged. Tests cover FrontBlock (round trip; an unclosed block or none refuses; CRLF).
- [x] **syncIndex:** keeps the Peek no-op shortcut. Then, up to `writeTries` times: ReadFrontmatter, then Peek for the body's links, reconcile, and on a change `WriteFrontmatter` with the read's hash; stale means retry, and the lock means a refusal. `syncIndexAfter`, `readAllShowingUnread`, `windowChanges` and submit's `printed` plumbing are removed.
- [x] **Fake:** a frontmatter read or write resets only the frontmatter's shadow (`shadow[""]`). The CAS hash covers the frontmatter block. `editBeforeFront` simulates a racing frontmatter edit.
- [x] **Tests:**
  - TestIndexKeepsUnreadEdits becomes a write that leaves the operator's `## Needs you` edit in the diff and prints nothing about it.
  - A racing frontmatter edit goes stale, then is recomputed.
  - Under the lock, nothing is written.
  - The workaround tests (`TestIndexShowsEditsOnEveryRead`, `TestSubmitPrintsUnreadOnce`) are replaced by a check that submit prints the diff once and the index adds nothing.
  - Live `TestSessionSyncConformance`: read --frontmatter returns the block with its hash; an operator edit in `## B` stays in the diff after the frontmatter read and after the frontmatter write; a wrong hash gives exit 6; the write keeps every section byte for byte; the lock gives exit 3.
- [x] Help (`status index`, the syncIndex comment) and the atlas (collaboration-modes "The index") describe the frontmatter write.

Design notes:
- **Body staleness is benign:** the CAS covers only the frontmatter, so a body edit after the Peek can only delay a drop to the next reconcile. Entries are never added from links.
- **Not covered:** an unread frontmatter edit (for example, the operator removing an entry, which means reviewed) is acknowledged by the frontmatter read when the index changes. The index is the only frontmatter field the TL writes, and the reconcile honors the removal, so nothing is lost. The Done-when scopes the guarantee to section edits.
- ARCH-MOCK: the fake and the live conformance check move together. ARCH-ORDER: holds no state between events beyond couch's CAS; a lost race is a stale write, then a reread. ARCH-FUNERAL: creates nothing durable. ARCH-SECURE: the frontmatter block comes from couch and is parsed by sidedoc, which refuses what YAML reads differently.

## Log


- 2026-10-18: closed — Suite green on 1278a020 (sdlc-local/suite, 8m31s). go test ./cmd/tl-lead ./pkg/sidedoc pass: TestFrontBlock; TestIndexKeepsUnreadEdits (a no-op reconcile reads nothing through couch; after a link writes the index the diff is exactly the operator section edit, nothing printed); TestIndexCommand (a racing frontmatter edit goes stale, is recomputed and kept; nothing written under the lock); one-way sequence and link split unchanged. Live TL_LEAD_LIVE=1 TestSessionSyncConformance passed against the real couch (pair#643) on a t.TempDir scratch file with a dummy owner, as the hub required: read/write --frontmatter keep a section edit in the diff, keep every body byte, stale gives exit 6, lock gives exit 3. Printed-diff workaround removed; help and atlas updated.; review verdict: SHIP
### 2026-10-17
- 2026-10-18: the hub approved the plan, on condition that the live conformance runs only against a scratch session-sync file. `TestSessionSyncConformance` writes its file under `t.TempDir()` with a dummy owner, never a real session file.
- Implemented as planned. `unreadChanges` lost its only production caller, so it moved into conformance_test.go, its one user. `newFake` now starts with `type: session-sync` frontmatter, since couch serves only such files; three instruction tests had used a frontmatter-less file that real couch refuses.
- Verified: `go test ./cmd/tl-lead ./pkg/sidedoc` passes (TestFrontBlock, TestIndexKeepsUnreadEdits with an exact diff after the write, a racing frontmatter edit in TestIndexCommand recomputed and kept, the lock). `TL_LEAD_LIVE=1 go test ./cmd/tl-lead -run TestSessionSyncConformance` passed against the real couch (pair#643).
