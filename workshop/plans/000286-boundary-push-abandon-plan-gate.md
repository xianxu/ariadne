---
gate: plan-quality
issue: 286
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-08T15:50:06-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: abandon's rerun after a partial step 6 is refused by its own step-1 precondition
          detail: Step 6 switches to the resting branch before deleting local and remote branches; step 1 refuses on the resting branch with started work. Key rerun detection on the card's abandoned record (terminal, same record means resume from step 5 or 6) ahead of the step-1 branch check, and add an interruption inside step 6 to TestAbandonRerunResumes.
          family: convergent-sequence-resume
          round: 1
        - id: PQ-2
          severity: Important
          title: reopen restore (D9) has no rerun design, and deletes the archive ref before the card CAS clears the record
          detail: 'A crash between ref deletion and the card CAS leaves a record naming a missing ref, so the rerun''s fetch fails. Order it as: card CAS clears the record, then delete the ref. Specify how a rerun detects an already-created branch, a merge in progress or a restore commit already made, and test interruptions at each step.'
          family: convergent-sequence-resume
          round: 1
        - id: PQ-3
          severity: Important
          title: abandon's main archive hardcodes history/issues/X instead of extending archiveDestination and archivedDetails, and ignores plan artifacts
          detail: archivepolicy.go:13 owns the archive path and landingarchive.go:66 owns the details projection, which accepts done only. Generalize them to terminal statuses rather than adding a parallel archiver, and say whether a plan on main (planArtifactBelongsToIssue) moves with the details.
          family: reuse-archive-policy
          round: 1
        - id: PQ-4
          severity: Important
          title: no non-goals stated; boundary pushes are tracker-only but the Done-when says every boundary verb
          detail: 'startplan.go:279-283 returns early for untracked repos and pushIssueBranch takes a trackerEnv. State the non-goals: legacy repos and merge.go, un-archiving a reopened done issue, and abandon by a non-owner.'
          family: stated-non-goals
          round: 1
        - id: PQ-5
          severity: Minor
          title: Tasks 1 and 5 enumerate test cases in prose; compress to one strategy line per risky function
          family: test-prose-enumeration
          round: 1
        - id: PQ-6
          severity: Minor
          title: set-status becomes an effect-heavy restore verb (fetch, merge, commit, push), and who may reopen is unstated
          family: unstated-seam-change
          round: 1
        - id: PQ-7
          severity: Minor
          title: added network latency per boundary verb and offline behaviour are not stated
          family: operating-envelope
          round: 1
        - id: PQ-8
          severity: Minor
          title: Task 2's re-close after rebase assumes close can re-run on a codecomplete card; name the path or use milestone-close
          family: unbacked-existing-behavior
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-08T15:51:35-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: D7 detects a rerun from the card's abandoned record before step 1; step 6 deletes the remote branch before switching; the test interrupts inside step 6.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: D9 writes the card (step 6) before deleting the ref (step 7), has a rerun check per step including MERGE_HEAD, and abandon step 3 accepts the orphan ref.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Step 5 reuses archiveDestination/archivedDetails generalized to terminal statuses (Task 4a) and moves plan files via planArtifactBelongsToIssue.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: 'Non-goals section added: legacy repos and merge, done-reopen, non-owner abandon, archive-ref pruning.'
          round: 2
        - id: PQ-5
          disposition: addressed
          round: 2
        - id: PQ-6
          disposition: addressed
          round: 2
        - id: PQ-7
          disposition: addressed
          round: 2
        - id: PQ-8
          disposition: addressed
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-08T15:52:28-07:00"
      agent: claude
      findings:
        - id: PQ-9
          severity: Minor
          title: runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
          detail: The behavior claims are correct and only the line anchors have moved. Name the functions and drop the line numbers.
          family: stale-line-anchor
          round: 3
      blocked: false
content_hash: b191f4d3d07446a708b5e6f0ad405ed75177ab900a6dc500fc21a28017af5333
---

# Gate ledger — ariadne#286 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-08T15:50:06-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `convergent-sequence-resume` abandon's rerun after a partial step 6 is refused by its own step-1 precondition
  Step 6 switches to the resting branch before deleting local and remote branches; step 1 refuses on the resting branch with started work. Key rerun detection on the card's abandoned record (terminal, same record means resume from step 5 or 6) ahead of the step-1 branch check, and add an interruption inside step 6 to TestAbandonRerunResumes.
- **PQ-2** [Important] `convergent-sequence-resume` reopen restore (D9) has no rerun design, and deletes the archive ref before the card CAS clears the record
  A crash between ref deletion and the card CAS leaves a record naming a missing ref, so the rerun's fetch fails. Order it as: card CAS clears the record, then delete the ref. Specify how a rerun detects an already-created branch, a merge in progress or a restore commit already made, and test interruptions at each step.
- **PQ-3** [Important] `reuse-archive-policy` abandon's main archive hardcodes history/issues/X instead of extending archiveDestination and archivedDetails, and ignores plan artifacts
  archivepolicy.go:13 owns the archive path and landingarchive.go:66 owns the details projection, which accepts done only. Generalize them to terminal statuses rather than adding a parallel archiver, and say whether a plan on main (planArtifactBelongsToIssue) moves with the details.
- **PQ-4** [Important] `stated-non-goals` no non-goals stated; boundary pushes are tracker-only but the Done-when says every boundary verb
  startplan.go:279-283 returns early for untracked repos and pushIssueBranch takes a trackerEnv. State the non-goals: legacy repos and merge.go, un-archiving a reopened done issue, and abandon by a non-owner.
- **PQ-5** [Minor] `test-prose-enumeration` Tasks 1 and 5 enumerate test cases in prose; compress to one strategy line per risky function
- **PQ-6** [Minor] `unstated-seam-change` set-status becomes an effect-heavy restore verb (fetch, merge, commit, push), and who may reopen is unstated
- **PQ-7** [Minor] `operating-envelope` added network latency per boundary verb and offline behaviour are not stated
- **PQ-8** [Minor] `unbacked-existing-behavior` Task 2's re-close after rebase assumes close can re-run on a codecomplete card; name the path or use milestone-close

## Round 2 — 2026-10-08T15:51:35-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — D7 detects a rerun from the card's abandoned record before step 1; step 6 deletes the remote branch before switching; the test interrupts inside step 6.
- PQ-2 — addressed — D9 writes the card (step 6) before deleting the ref (step 7), has a rerun check per step including MERGE_HEAD, and abandon step 3 accepts the orphan ref.
- PQ-3 — addressed — Step 5 reuses archiveDestination/archivedDetails generalized to terminal statuses (Task 4a) and moves plan files via planArtifactBelongsToIssue.
- PQ-4 — addressed — Non-goals section added: legacy repos and merge, done-reopen, non-owner abandon, archive-ref pruning.
- PQ-5 — addressed
- PQ-6 — addressed
- PQ-7 — addressed
- PQ-8 — addressed

## Round 3 — 2026-10-08T15:52:28-07:00 (claude) — passed

### Raised

- **PQ-9** [Minor] `stale-line-anchor` runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
  The behavior claims are correct and only the line anchors have moved. Name the functions and drop the line numbers.

## Open findings

- **PQ-9** [Minor] `stale-line-anchor` runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
