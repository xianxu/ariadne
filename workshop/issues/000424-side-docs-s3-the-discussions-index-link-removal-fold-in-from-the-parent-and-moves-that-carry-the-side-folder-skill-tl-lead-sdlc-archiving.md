---
id: 000424
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours: 4.13
collaboration-mode: collaboration
card_mirror: '06f2f44606833961f144cddea840a623f3f1ded8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T03:02:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving)

## Problem

S2 of the universal agent input interface (ops `teams/devinfra/projects/universal-agent-input-interface/design.md`) built how a side doc starts. S3 covers the rest of its life, and the operator approved its design, now final: ops `design.side/07-s3-side-docs.md` (cf0f0ebf), with threads 08 and 09. ariadne holds three parts of it: the `xx-universal-review` skill, `tl-lead status`, and sdlc's archiving. parley's mover is parley.nvim#373.

## Spec

From 07, sections 3 to 7 and 10. Decisions (operator, 10-16):
- **One attachment, two references.** The attachment is the side doc's `parent:` and `marker:`; the references are the link in the parent's text (`→ [discussion](…)` or `→ [answer](…)`) and the parent's `discussions:` index entry.
- **The index rule is one-way.** Deleting a link drops its index entry on the agent's next round; removing an entry means reviewed and leaves the link. The agent adds an entry back only when it next writes something for the operator in that side doc. The agent never removes a link itself.
- **The index lives in the immediate parent**, recursively. The entry rides the parent's single compare-and-swap write, in the same round as the link.
- **No closing.** A side doc is never deleted, marked or locked. S2's reuse by `parent:` and `marker:` is unchanged.
- **Needs you stays hand-written** free text, not a view of the index.
- **Fold-in is triggered in the parent** by a `🤖[fold in]` at or beside the link. The fold marks the link "(folded <date>)" and adds a closing note to the side doc. An approval inside a side doc only records a `Decided (operator): …` line there; this replaces S2's approval-folds rule (row 7 and its row in section 3). One level per fold. Answer docs never fold into their chat item.
- **Moves the agent makes** follow one contract: move the file and its `.side/` folder together; rewrite the side docs' `parent:` and backlinks; on a name change, rewrite the parent's own links and index; run `couch --session-sync rename OLD NEW` when the file is registered. When the target folder name is taken, the orphan folder takes the suffix (`spec_1.side/`, then `_2`, …), and the moving document keeps `spec.side/`.
- **One fixture set:** the move fixtures are shared with parley's projection specs and travel in `export.json`.

The work, by place:
- **`xx-universal-review`:** `collaboration-modes.md` gains 07's sections 3 to 7 and rows 18 to 28. `examples/` gains: a deleted link dropping its entry, an entry added back after review, fold-in from the parent, a chain fold, and a name-change move. The examples checker gains two invariants: every index entry names a side doc whose `parent:` is that document, and the move fixtures match parley's.
- **tl-lead:** `tl-lead status` writes the session file's `discussions:` index and drops an entry when its item is deleted.
- **sdlc:** archiving an issue or plan to `workshop/history/` moves its `.side/` folder with it, under the same move contract.

## Done when

- The skill's `collaboration-modes.md` and examples carry sections 3 to 7 and rows 18 to 28, and the examples checker enforces both new invariants.
- `tl-lead status` writes and prunes the session file's index by the one-way rule, with a test.
- `sdlc` archiving an issue or plan that has a `.side/` folder moves the folder and keeps every link and index entry resolving, with a test.
- On a real document: a question opens a side doc and its index entry; the operator's removal of the entry sticks until the agent next writes there; deleting the link drops the entry; a `🤖[fold in]` by the link folds the outcome and marks the link.

## Plan

Durable plan: [workshop/plans/000424-…-plan.md](../plans/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving-plan.md).

- [x] M1 — the skill: `pkg/sidedoc` (links, index), `collaboration-modes.md` (index, no closing, fold-in from the parent, moves; rows 21 to 31), six new examples and the existing ones updated, checker invariants, move fixtures exported (version 2)
- [x] M2 — tl-lead: the session file's `discussions:` index (`instructions link` adds, `status submit` and `status index` reconcile) through `write --all`
- [ ] M3 — sdlc: the three archive paths move a document's `.side/` folder with it

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.3 impl=0.24
item: skill-or-dispatcher    design=0.3 impl=0.16
item: skill-or-dispatcher    design=0.2 impl=0.2
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.2 impl=0.2
item: greenfield-go-module   design=0.5 impl=0.32
item: atlas-docs             design=0.1 impl=0.08
item: milestone-review       design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 4.13
```

Items in order: M1 `pkg/sidedoc` (links, index; greenfield, design mostly settled by 07); M1 `collaboration-modes.md` contract; M1 the six examples and the updates to existing ones; M1 checker invariants and `projectMove`; M1 export version 2; M2 tl-lead index over the existing seam; M3 the `.side/` counterparts of the four asset archive stages across paths (a) to (f), the riskiest part; atlas and help; three milestone reviews. impl at 40% of the v2 table (v3.1); design buffer 15% for the thorough plan.

> *Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-17
- 2026-10-17: closed M2 — BR-8 fixed for the class: every whole-file read goes through readAllShowingUnread (diff, Peek, read, changed lines shown); TestIndexShowsEditsOnEveryRead (stale retry shows the racing edit; an edit between diff and read shown as lines), TestIndexKeepsUnreadEdits, the M2 sequence tests pass; live conformance passes. BR-10: plan Integration points/ARCH-ORDER/Revisions updated with the class rule.; review verdict: SHIP
- 2026-10-17: closed M1 — BR-7 fixed by rule: ReconcileIndex and Entries decode the frontmatter with yaml.v3 before and after the rewrite and refuse on any disagreement; TestReconcileIndexRefuses adds the column-0 comment, tab and extra-key inputs; FuzzReconcileIndex 3m/4.6M execs against the YAML oracle; pkg/reviewexport (18 examples, export current) passes.; review verdict: SHIP
- Filed from the S3 design (ops 07, final at cf0f0ebf); parley's half is parley.nvim#373 (filed by the TL from this slot's draft).
- 07's rows 18 to 28 become 21 to 31: the skill's states table already runs to 20.
- sdlc never renames an issue or plan file (`set-title` keeps the slug), and archiving keeps the base name, so sdlc's part is a same-name move of the `.side/` folder, with no link rewriting.
- parley's `scripts/sync-review-protocol.py` refuses a role it doesn't know and any version but 1, so exporting the move fixtures bumps the export to version 2, which parley.nvim#373 accepts. Parley's committed snapshot is unaffected until it syncs.
- Plan approved by the TL; plan-quality cleared in 2 rounds (PQ-1: M3 now follows the asset archive's four stages; the Minors fixed). Estimate 4.13h (v3.1). The estimate-quality judge flagged M2 (CAS and conformance tests) and M3 (six archive paths) as thin and the smoke round as unitemized; kept as derived, to compare at close.
- M1 built: `pkg/sidedoc` (links, the index; fuzzed: it found a leading-space path that didn't round-trip, now refused), the contract (07 sections 3 to 7), six new examples plus the five updated, checker invariants, export v2 with the move fixtures.
- Rows: 07's row 23 (fold in from the parent) and row 24 (an approval in a side doc) update the skill's rows 10 and 7 instead of duplicating them, so the new rows are 21 to 29, not 21 to 31.
- The fold's closing note has a fixed form (`Folded into the parent on <date>: …`, or `Proposed in the parent on …` under propose-strike) so the checker can tie it to the folded link.
- Two invariants I first wrote were wrong and are not in the checker: "every new side doc gets an entry" (an answer doc that only records an action adds none, 07 section 4), and "an entry is added only for a changed side doc" (a relink adds the entry back without changing the doc).
- Move cases: the per-tree check skips the orphan folder (no document by definition); a link from another document into a folder isn't a fixture, because the tree rules allow a link only from the side doc's parent (TestProjectMove covers non-rewriting).
- Real-document round (the TL's pick: ops `qa/xian-xu/2026-10-16-24-disk-writes-now.md` and its `01-go-flags.md`). The parent is Couch-registered to ops:0, so instead of writing it from this slot (an owner's ledger and diff baseline), the rounds ran on byte-identical copies in this slot's scratchpad. The originals were never written (sha256 7934f029…, 90834d64… before and after). Rounds, each checked with the examples checker:
  - **Index add:** the operator's `🤖[how much of that is duplicates from untrimmed builds?]` in the build-cache bullet became `→ [discussion](2026-10-16-24-disk-writes-now.side/02-untrimmed-duplicates.md)`, the new side doc (number 02, its frontmatter copying the parent's process-instruction), and `  - doc: 2026-10-16-24-disk-writes-now.side/02-untrimmed-duplicates.md` appended after the existing 01 entry.
  - **Fold-in from the parent:** `→ [discussion](…/01-go-flags.md) 🤖[fold in]` became `… (folded 2026-10-17)` plus `Decided (operator): roll the slot restarts, each slot relaunched at the earliest moment it is idle, …` at the link's spot. 01 gained `Folded into the parent on 2026-10-17: the rolling slot restarts (after **Next**).` No entry was added; 01's existing entry stayed.
  - **Link deletion:** with the folded link to 01 deleted by the operator, the round dropped `  - doc: …/01-go-flags.md`, leaving only 02's entry. 01 is unchanged and stays.
  - Found on the way: the real `01-go-flags.md` has no `_From […], section "…"` backlink (S2's contract), a gap in the hand-written side doc. It exposed a checker bug: a malformed side doc's number wasn't counted, so the checker asked for a new `01-…` that would collide. Numbering now counts every file in the folder, with a mutation test.
- M2 built: `ReadAll`/`WriteAll` on the `SessionSync` seam; the fake gained a whole-file view over the same state, exactly invertible, so an untouched section keeps its hash across a whole-file write, as in couch. `syncIndex` (`cmd/tl-lead/index.go`) wraps `sidedoc.ReconcileIndex` in a retried CAS write. `instructions link` writes the entry, then the link. `submit` reconciles (a refusal is reported, not fatal). `status index FILE [--add DOC]` is new. The live `TestSessionSyncConformance` (TL_LEAD_LIVE=1) confirms `read --all`/`write --all --if-hash` with stale and lock refusals.
- M2 review round 1 (REWORK): BR-8 (Critical) — couch's read --all and write --all mark every edit seen, so an index write would have swallowed the operator's unread edits from the TL's "changes since your last read". Fixed: a no-op reconcile decides by `Peek` (the disk file, baseline untouched), and a needed write first prints the pending diff. BR-9 — the fake now models couch's seen-baseline (whole-file read/write reset it, a section write only its section), with `TestIndexKeepsUnreadEdits`; the live conformance pins that `read --all` empties the diff and that `Peek` matches it. The full fix is a frontmatter-scoped couch read and write, proposed to the TL for pair.
- M2 review round 2: BR-8 still open on the retry path (a stale retry re-read without printing the diff). Fixed for the class: every whole-file read goes through `readAllShowingUnread` (diff, then Peek, then read, printing any lines changed in between); `TestIndexShowsEditsOnEveryRead` covers the stale retry and the window. BR-10: the plan now names `Peek` and couch's seen-baseline (Integration points, ARCH-ORDER, Revisions, with the class rule). The TL filed the couch side as pair#643; tl-lead switches to it in a follow-up.
- M2 closed SHIP (round 3); its two Minors fixed after (lineDiff reused; submit prints the unread diff once).
- M3 built: the `.side/` counterparts of the asset archive's stages in `cmd/sdlc/archiveside.go` (prepare, the shared `finishArchiveMemberMoves`, recovery from the moved documents, pre-commit verify), wired into merge, push, push's recovery and the trunk archive (first write, retry residue refusal, successor replacement). Beyond the plan's (a)–(f): the landing archive (g), reopen's `unarchiveIssue` (the reverse move), and abandon's kept-branch side folders (`mainArchive.Sides`). `archivePlanArtifacts` used to rename a plan's `.side` folder whole (its glob matched the folder); it now skips folders. The archive test that timed out in M1's review passes here in 0.7 s: an environment timeout.
- M3 review round 1 (FIX-THEN-SHIP), all six fixed. Verify now checks only that every member arrived (present at its destination, gone from its source): a `parent:` resolves after a same-name move exactly when it did before, so judging content after every move could only strand the archive halfway. The successor episode replaces every document's side folder (issue and plans), and the retry refuses plan side-doc residue too (`trunkSideResidue`). The help text is updated (push, merge, abandon). Side-doc path and frontmatter facts live in `pkg/sidedoc` (`Folder`, `Fields`), and the checker and landing archive call them. Swallowed IO errors are now returned. Reopen compares bytes before replacing a live side doc, as it does for assets. Found while testing the successor: `keptPlans` and `unarchiveIssue` listed the plans directory non-recursively and read a plan's `.side` folder as a plan; both now list files only (`lsTreeFiles`).
- Targeted re-run after the M3 round-1 fixes: all archive, landing, abandon, reopen, help and side-doc tests pass. One failure outside #424: `TestLocalCheckSignalCancelsQueuedHelper/terminated/cleanup` ("cancelled run retained its worktree", a signal-timing test under load). It passed 3 of 3 runs alone. No existing record; watch for it in the full suite.
