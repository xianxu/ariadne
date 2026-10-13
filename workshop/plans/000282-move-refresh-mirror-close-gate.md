---
gate: boundary-review
issue: 282
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T13:04:37-07:00"
      dispatched: "2026-10-13T13:00:25-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: move's worktree mirror refresh dirties the destination; the documented move-back then refuses
          detail: 'Reproduced: after sdlc move :0 the destination shows " M workshop/issues/<id>.md", and sdlc move :1 from there fails with "has uncommitted changes". Instances: cmd/sdlc/move.go:115 (the refresh), helptext/move.md OWNERSHIP section (silent on the mirror edit), atlas/workflow/workspace-branching.md move paragraph and its "Moving back" bullet. Fix: commit the refresh narrowly on the issue branch (as close''s mirror commit does, #275), or at minimum document that it must be committed before moving back, and log the decision.'
          family: verb-leaves-tree-dirty-against-clean-precondition
          round: 1
        - id: BR-2
          severity: Important
          title: '"A non-applicable relocation leaves details untouched" has no test'
          detail: TestMoveLeavesAnUnattributedIssueUnknown commits no details, so "untouched" holds trivially; the case of a card owned by another slot has no test at all. Add commitMovedDetails to the unattributed test and assert the details are unchanged and the tree is clean.
          family: done-when-clause-tested-one-side
          round: 1
        - id: BR-3
          severity: Minor
          title: refreshMovedMirror repeats moveRelocation's cutover check, tracker open and details-path derivation
          detail: 'move.go:130-137 vs move.go:148-153: the cutover stat is repeated with a mismatched check (err != nil vs errors.Is ErrNotExist), the tracker is opened and snapshotted again, and the details path comes from branch+".md" while every other caller uses path.Base(card.Path) (cardsetters.go:53, startplan.go:316, reclaim.go:163, claim.go:136). Refresh inside moveRelocation''s flow using refreshMirrorFrom on the card it already read.'
          family: arch-dry-restated-precondition
          round: 1
        - id: BR-4
          severity: Minor
          title: A move with no card (moveRelocation's ErrNoCard path returns nil) still tries the refresh and warns
          detail: move.go:159-160 returns nil when the card is missing, so refreshMovedMirror runs; if details exist, snap.Require fails and the move prints a spurious "mirror not refreshed" warning. Skip the refresh when no card exists.
          family: refresh-after-noop-relocation
          round: 1
      recipe: small-diff-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#282 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T13:04:37-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `verb-leaves-tree-dirty-against-clean-precondition` move's worktree mirror refresh dirties the destination; the documented move-back then refuses
  Reproduced: after sdlc move :0 the destination shows " M workshop/issues/<id>.md", and sdlc move :1 from there fails with "has uncommitted changes". Instances: cmd/sdlc/move.go:115 (the refresh), helptext/move.md OWNERSHIP section (silent on the mirror edit), atlas/workflow/workspace-branching.md move paragraph and its "Moving back" bullet. Fix: commit the refresh narrowly on the issue branch (as close's mirror commit does, #275), or at minimum document that it must be committed before moving back, and log the decision.
- **BR-2** [Important] `done-when-clause-tested-one-side` "A non-applicable relocation leaves details untouched" has no test
  TestMoveLeavesAnUnattributedIssueUnknown commits no details, so "untouched" holds trivially; the case of a card owned by another slot has no test at all. Add commitMovedDetails to the unattributed test and assert the details are unchanged and the tree is clean.
- **BR-3** [Minor] `arch-dry-restated-precondition` refreshMovedMirror repeats moveRelocation's cutover check, tracker open and details-path derivation
  move.go:130-137 vs move.go:148-153: the cutover stat is repeated with a mismatched check (err != nil vs errors.Is ErrNotExist), the tracker is opened and snapshotted again, and the details path comes from branch+".md" while every other caller uses path.Base(card.Path) (cardsetters.go:53, startplan.go:316, reclaim.go:163, claim.go:136). Refresh inside moveRelocation's flow using refreshMirrorFrom on the card it already read.
- **BR-4** [Minor] `refresh-after-noop-relocation` A move with no card (moveRelocation's ErrNoCard path returns nil) still tries the refresh and warns
  move.go:159-160 returns nil when the card is missing, so refreshMovedMirror runs; if details exist, snap.Require fails and the move prints a spurious "mirror not refreshed" warning. Skip the refresh when no card exists.

## Open findings

- **BR-1** [Important] `verb-leaves-tree-dirty-against-clean-precondition` move's worktree mirror refresh dirties the destination; the documented move-back then refuses
- **BR-2** [Important] `done-when-clause-tested-one-side` "A non-applicable relocation leaves details untouched" has no test
- **BR-3** [Minor] `arch-dry-restated-precondition` refreshMovedMirror repeats moveRelocation's cutover check, tracker open and details-path derivation
- **BR-4** [Minor] `refresh-after-noop-relocation` A move with no card (moveRelocation's ErrNoCard path returns nil) still tries the refresh and warns
