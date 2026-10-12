---
gate: boundary-review
issue: 308
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T09:24:26-07:00"
      dispatched: "2026-10-12T09:14:59-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: NeedsSourceCheckout still treats creation stage 1 as a local effect; plan says false
          detail: store.go:129-132 returns Stage<=1 for creation, and the comment still says "materializing created details". creation.main needs no checkout, so reconcile from another branch or worktree refuses with ErrForeignCheckout (issuerecovery.go:131) and recovery list says "finish from X". Return false for creation (check legacy receipts before the guard) and add a cross-branch reconcile test.
          family: retired-stage-restated-in-model
          round: 1
        - id: BR-2
          severity: Important
          title: Legacy creation.detail receipt release in reconcile has no test
          detail: 'issuerecovery.go:135-142 deletes a pre-#308 stage-1 receipt and points at issue publish. Only the LegacyCreationDetail predicate is tested. Add the plan''s TestRecoveryLegacyCreationReceiptPointsAtPublish: receipt removed, stderr names publish, main unchanged.'
          family: untested-recovery-path
          round: 1
        - id: BR-3
          severity: Minor
          title: No adapter-level test of an uncertain creation.main push (plan's TestCreationMainUncertainIsRecoverable)
          family: untested-recovery-path
          round: 1
        - id: BR-4
          severity: Minor
          title: issueDirGuard refuses a fresh legacy repo without workshop/issues with a misleading "not the repository you mean"
          family: guard-message-precision
          round: 1
        - id: BR-5
          severity: Minor
          title: move-detail alias re-implements publish's argument and legacy checks and skips guardSpineRepo
          family: alias-duplicates-target-guards
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T09:40:17-07:00"
      dispatched: "2026-10-12T09:28:11-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: store.go:131 drops creation; TestIssueNewStoppedAtMainFinishesFromAnotherCheckout fails with ErrForeignCheckout when the old case is restored (verified).
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestRecoveryLegacyCreationReceiptPointsAtPublish goes red (blob-not-a-commit) when the issuerecovery.go:135 release is disabled (verified).
          round: 2
        - id: BR-3
          disposition: addressed
          note: The pre-receive refusal with an unchanged base goes through publishEvent's probe/re-push path, and the recovery tests resume the creation from it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: issue.go:398 message now names the fresh-repo action; TestLegacyIssueNewRefusesRepoWithoutIssuesDir pins it.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Alias calls guardSpineRepo then runTrackedPublish; its distinct legacy refusal is deliberate (publish's legacy path is --commit).
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: issuerecovery.go:131 keeps a creation exemption in the foreign-checkout guard that can no longer fire
          detail: '2nd in family. Rule: NeedsSourceCheckout alone says where a receipt must be finished; no caller adds per-operation exceptions. Delete the clause and grep issuerecovery.go for other creation-specific checkout exceptions.'
          family: retired-stage-restated-in-model
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: 8d6c22b9a9d0ac0d9aec23283c1c22891a5c1520
      blocked: false
    - "n": 3
      timestamp: "2026-10-12T10:23:44-07:00"
      dispatched: "2026-10-12T10:16:01-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: c929e244 removes the creation clause at issuerecovery.go:131; no other creation-specific checkout exception remains (store.go NeedsSourceCheckout has no creation case). Dead-clause removal, no behaviour test needed.
          round: 3
      findings:
        - id: BR-7
          severity: Important
          title: 'Close no longer ticks a brain repo''s legacy data/project project: it is classed as a peer, skipped locally, then refused as a brain'
          detail: close.go:851 marks every match outside repoTop as a peer, applyClose skips its local write (close.go:926) and publishPeerProject refuses IsBrainRepo (peerproject.go:81). The base wrote the brain file and left it for nous to sweep. The atlas line saying so was deleted, the close Long text still promises a brain-side update, and no test covers data/project. Either exclude brains from peer (keep the local write) or record the retirement in Revisions, atlas and help, and add a data/project fixture test.
          family: replaced-path-case-dropped
          round: 3
        - id: BR-8
          severity: Minor
          title: Peer publish is gated on the stale local copy (newPT != pt) rather than on origin's bytes
          detail: A stale peer checkout can die on a missing detail block that origin has, or skip a publish that origin still needs.
          family: replaced-path-case-dropped
          round: 3
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 4
      timestamp: "2026-10-12T10:38:35-07:00"
      dispatched: "2026-10-12T10:26:23-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: close.go:856 excludes m.Legacy from peer; the in-place write is kept. TestClose_BrainLegacyProjectIsWrittenInPlace goes red when the fix is reverted (verified). Atlas and close help updated.
          round: 4
        - id: BR-8
          disposition: not-addressed
          note: The || peer gate is in, but no test fails when it is reverted (verified). A stale local copy missing the detail block still dies at close.go:848 before origin is read.
          round: 4
      recipe: milestone-review
      isolation: checkout
      reviewed: ab8a8dfbbf6287ac56354d19ace0ae65958d29d6
      blocked: false
---

# Gate ledger — ariadne#308 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T09:24:26-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `retired-stage-restated-in-model` NeedsSourceCheckout still treats creation stage 1 as a local effect; plan says false
  store.go:129-132 returns Stage<=1 for creation, and the comment still says "materializing created details". creation.main needs no checkout, so reconcile from another branch or worktree refuses with ErrForeignCheckout (issuerecovery.go:131) and recovery list says "finish from X". Return false for creation (check legacy receipts before the guard) and add a cross-branch reconcile test.
- **BR-2** [Important] `untested-recovery-path` Legacy creation.detail receipt release in reconcile has no test
  issuerecovery.go:135-142 deletes a pre-#308 stage-1 receipt and points at issue publish. Only the LegacyCreationDetail predicate is tested. Add the plan's TestRecoveryLegacyCreationReceiptPointsAtPublish: receipt removed, stderr names publish, main unchanged.
- **BR-3** [Minor] `untested-recovery-path` No adapter-level test of an uncertain creation.main push (plan's TestCreationMainUncertainIsRecoverable)
- **BR-4** [Minor] `guard-message-precision` issueDirGuard refuses a fresh legacy repo without workshop/issues with a misleading "not the repository you mean"
- **BR-5** [Minor] `alias-duplicates-target-guards` move-detail alias re-implements publish's argument and legacy checks and skips guardSpineRepo

## Round 2 — 2026-10-12T09:40:17-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — store.go:131 drops creation; TestIssueNewStoppedAtMainFinishesFromAnotherCheckout fails with ErrForeignCheckout when the old case is restored (verified).
- BR-2 — addressed — TestRecoveryLegacyCreationReceiptPointsAtPublish goes red (blob-not-a-commit) when the issuerecovery.go:135 release is disabled (verified).
- BR-3 — addressed — The pre-receive refusal with an unchanged base goes through publishEvent's probe/re-push path, and the recovery tests resume the creation from it.
- BR-4 — addressed — issue.go:398 message now names the fresh-repo action; TestLegacyIssueNewRefusesRepoWithoutIssuesDir pins it.
- BR-5 — addressed — Alias calls guardSpineRepo then runTrackedPublish; its distinct legacy refusal is deliberate (publish's legacy path is --commit).

### Raised

- **BR-6** [Minor] `retired-stage-restated-in-model` issuerecovery.go:131 keeps a creation exemption in the foreign-checkout guard that can no longer fire
  2nd in family. Rule: NeedsSourceCheckout alone says where a receipt must be finished; no caller adds per-operation exceptions. Delete the clause and grep issuerecovery.go for other creation-specific checkout exceptions.

## Round 3 — 2026-10-12T10:23:44-07:00 (claude) — BLOCKED

### Disposed

- BR-6 — addressed — c929e244 removes the creation clause at issuerecovery.go:131; no other creation-specific checkout exception remains (store.go NeedsSourceCheckout has no creation case). Dead-clause removal, no behaviour test needed.

### Raised

- **BR-7** [Important] `replaced-path-case-dropped` Close no longer ticks a brain repo's legacy data/project project: it is classed as a peer, skipped locally, then refused as a brain
  close.go:851 marks every match outside repoTop as a peer, applyClose skips its local write (close.go:926) and publishPeerProject refuses IsBrainRepo (peerproject.go:81). The base wrote the brain file and left it for nous to sweep. The atlas line saying so was deleted, the close Long text still promises a brain-side update, and no test covers data/project. Either exclude brains from peer (keep the local write) or record the retirement in Revisions, atlas and help, and add a data/project fixture test.
- **BR-8** [Minor] `replaced-path-case-dropped` Peer publish is gated on the stale local copy (newPT != pt) rather than on origin's bytes
  A stale peer checkout can die on a missing detail block that origin has, or skip a publish that origin still needs.

## Round 4 — 2026-10-12T10:38:35-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — close.go:856 excludes m.Legacy from peer; the in-place write is kept. TestClose_BrainLegacyProjectIsWrittenInPlace goes red when the fix is reverted (verified). Atlas and close help updated.
- BR-8 — not-addressed — The || peer gate is in, but no test fails when it is reverted (verified). A stale local copy missing the detail block still dies at close.go:848 before origin is read.

## Open findings

- **BR-8** [Minor] `replaced-path-case-dropped` Peer publish is gated on the stale local copy (newPT != pt) rather than on origin's bytes
