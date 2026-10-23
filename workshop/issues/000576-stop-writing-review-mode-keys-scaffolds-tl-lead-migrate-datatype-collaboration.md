---
id: 000576
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd76d81c01de2e879a58c286f5d4fd4b473fa5184' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T14:24:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.54
---

# Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows

## Problem

ariadne#575 rewrites universal-review to one collaborative editing mode with a `@@chat@@` section carve-out. The old mode keys stay readable as aliases, and nothing new writes them. ariadne's own creators still write them:
- `sdlc issue new` and handoff (`cmd/sdlc/internal/issue/scaffold.go`, `handoff.go`) and the project scaffold (`cmd/sdlc/internal/project/scaffold.go`) write `collaboration-mode: <datatype default>` into every new issue and project.
- `tl-lead` migrate (`cmd/tl-lead/migrate.go`) writes `collaboration-mode: discussion` into side docs.
- The datatype definitions (`construct/datatype/*.md`) carry a `collaboration-mode` row in their Frontmatter shape table. `cmd/datatype/collabmode.go` parses that row, and `pkg/vocab` `CollaborationModes` backs it.

## Spec

Remove every ariadne writer of the mode keys (`collaboration-mode`, `presentation-mode`/`presentation`, `processing-mode`, `processing-instruction`, `process-instruction`), as well as the datatype rows and the vocab list that exist only to feed them. Old documents must stay readable. Issue, project and side-doc files that already carry the keys still parse and round-trip unchanged: no reader refuses a document because a mode key is present, and no rewrite drops one. The `@@chat@@` carve-out is the only routing signal new documents carry. Land right after #575.


**Scope additions (TL, 10-23, deferred from ariadne#575's round-4 review):**
- ChatSections notice: "frontmatter could not be read" fires on a document that merely opens with a `---` thematic
  break and no closing fence. Notice only when a frontmatter block exists (an opening fence and a matching close,
  using frontmatter()'s fence detection, not a prefix test) and couldn't be read.
- An empty `collaboration-mode:` value prints "collaboration-mode:  is retired and ignored" (double space); say empty
  or malformed instead, as the spec's section 7 classes it.
- Adopt `sidedoc.ChatSections` in tl-lead's instructions.go (the review's recommendation).

## Done when

- New issues, projects (`sdlc issue new`, handoff, the project scaffold) and tl-lead side docs carry no mode keys. Tests assert their absence.
- The datatype tables have no `collaboration-mode` row, and nothing in Go reads one; `vocab.CollaborationModes` is gone or reduced to a legacy alias reader.
- Fixtures with legacy keys (an issue with `collaboration-mode: collaboration`, a side doc with all three keys) still parse, and survive every rewrite verb byte for byte.

## Plan

Durable plan: [000576 plan](../plans/000576-stop-writing-review-mode-keys-scaffolds-tl-lead-migrate-datatype-collaboration-plan.md) (approved by the TL 10-23, including the tl-lead ChatItems design call).

- [x] ChatSections scope additions: a linear scan (counted N/2N), the fence-based unread notice, the empty-value notice, Legacy and Innermost (laminar), and Tags.
- [x] Task 1: shared item grammar `sidedoc.ChatItems`; tl-lead list/link pick `legacy` from ChatSections(frontmatter + section); checkChat uses it
- [x] Task 2: sdlc scaffolds, handoff, CUE field, vocab accessors; LegacyCollaborationModes; legacy round-trip test
- [x] Task 3: tl-lead migrate lossless; help.md moves to @@chat@@
- [x] Task 4: datatype rows and collabmode.go removed; writing-plans header; the guard extended to every creator's instructions
- [x] Task 5: atlas and comments; tests; recorded suite; close; landed revision to pair:3 and parley.nvim:3

## Log

### 2026-10-23
- 2026-10-23: closed — Round 2 head 72459336c: all round-1 findings fixed. BR-1: the list token carries the grammar, and link refuses a change (fake race test gives exit 6 with no write; the mutant writes the wrong bullet). BR-2: fake-backed tests for tagged and heading-comment files. BR-3: card-publish round trip. Minors: plan Revisions; guard narrowed to "legacy `key`" spans (the mutation is caught); Innermost for an enclosing section. Round 1 evidence holds: the writers are removed with absence asserted, the legacy round trips pass, the creator guard holds, and the linear, CommonMark-correct chat grammar is shared. The recorded suite is queued on 72459336c.; review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 382 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Claimed after #575 landed (PR #418). A mapping agent listed every writer and reader. Before change-code I committed the ChatSections scope additions agreed with the TL and pair:3: 3d109d2c8 makes it linear (a counted test catches a reverted prefix count at ratio 4.07) and adds the fence-based unread notice and the empty-value notice; e12f0fccc adds Legacy and Innermost, with a randomized laminar test; 9dbb69c87 adds Tags for pair#838. The TL acknowledged the order as agreed scope. The plan, including the ChatItems design call, was approved 10-23.
- Task 1: `sidedoc.ChatItems(section, legacy)` holds the item grammar, with a counted N/2N test. tl-lead's `Bullet` is now an alias of `sidedoc.ChatItem`, and its private splitter, regexps and `indentOf` are gone. `checkChat` uses ChatItems too, and the chat-tag mutation still fails on the same line. One deviation from the approved wording, same spirit and one read: `list` and `link` choose `legacy` with `ChatSections` over `Peek(file)`, the whole file read without moving Couch's diff baseline (the interface documents Peek for exactly this kind of decision), instead of over frontmatter plus section. ReadFrontmatter would also mark the frontmatter seen as a side effect, and Peek lets the tag itself count. The rule: a tag alone gives numbered items; any legacy form, or none, keeps `-`, `*` and `N.` (`TestLegacyGrammar`). I also piped `sdlc change-code` through `tail` once more, though it only had to confirm the branch.
- pair:3 then found, by reading the source, that the fence toggle I carried over from tl-lead flipped on any fence-like line. I measured it: a tilde line inside a backtick fence, a shorter fence inside a longer one, and a backtick line inside a tilde fence each exposed an item. Fix (f8b819032): `sidedoc.CodeLines` takes code lines from goldmark, the parser CodeLiteralMask already uses. A counted N/2N test catches a per-block prefix count at about 4x. tl-lead's edge test encoded the toggle (an unindented line inside an item's fence), so it now uses the CommonMark form (8902b587b).
- Task 2: the issue scaffold, `DetailsFromCard` and the project scaffold write no mode key; tests assert absence across every retired key. The CUE `scaffold.collaboration_mode` field is deleted and JSON regenerated (`make vocab-embed`; its final drift check exits 1 only because the regenerated JSON is not yet committed). `Scaffold.CollaborationMode` and the two accessors are deleted, and `CollaborationModes` is renamed `LegacyCollaborationModes`. `TestLegacyModeKeySurvivesSetField` pins legacy lines surviving a rewrite.
- Task 4, in part: `cmd/datatype/collabmode.go`, its tests, `datatype lint` and the render warning are deleted. Nothing outside cmd/datatype called `datatype lint`.
- Task 3: `answerFrontmatter` puts type/parent/marker first and keeps every other existing line verbatim and in order, legacy mode keys included, adding none. Regenerating the golden tree removed only the two lines the migration used to add; the doc that already carried the key keeps it. `TestAnswerFrontmatterWritesNoModeKeys` pins both. `help.md` now prescribes `@@chat@@` above `## Operator instructions` and no mode keys for session files and answer docs. The help tests flip: they require the tag and forbid the three keys, "explicit default mode keys" and "use processing-instruction".
- Task 4: removed the 11 datatype `collaboration-mode` rows and workingdoc.md's own key; type.md now says prototypes carry no mode key. The writing-plans header no longer writes `collaboration-mode: discussion`. In the fix skill and the grammar, the retired "written-defaults" and "effective collaboration mode" wording is gone; Pair's legacy propose behavior is labeled legacy (until pair#838) and its operator-owned row no longer asks the operator to set a key. `TestNoSpecWritesModeKeys` now also covers the datatypes, the writing-plans skill, tl-lead's help and the fix skill, matching a key in any form but not the file name collaboration-modes.md, and skipping lines that describe legacy handling. My first version missed shape-table rows (no colon), and the mutation caught that; the fixed guard catches it.
- Close round 1 returned FIX-THEN-SHIP, with three Important findings and three Minors, all fixed in this round.
  - BR-1: list's token is now `hash:grammar` and link refuses on a grammar change. A fake-backed race test (list, add a tag, link) gives exit 6 with no write; removing the check writes to the wrong bullet.
  - BR-2: fake-backed tests cover a tagged and a heading-comment session file.
  - BR-3: `TestLegacyModeKeysSurviveCardPublish`.
  - Minors: plan Revisions added. The guard excuses only `legacy `key`` spans, and an instruction like "on a legacy document, write `presentation-mode`" now fails it. legacyGrammar resolves enclosing sections through Innermost.
  - The landed-revision message to pair:3 and parley.nvim:3 follows the landing.

