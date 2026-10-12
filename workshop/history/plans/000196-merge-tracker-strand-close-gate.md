---
gate: boundary-review
issue: 196
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T08:47:52-07:00"
      dispatched: "2026-10-12T08:42:51-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Stranded/9b messages advise committing, which makes the resume path refuse
          detail: 'strandedMessage (merge.go:696) and the 9b die text (merge.go:522) suggest commit. A commit on the merged branch trips step 4 (ahead of origin); pushing it trips actionResumeBlocked (#148, origin/main..origin/branch > 0) and the #160 publish gate. Recommend only stash/restore/move-aside then re-run; sweep both sites and assert in TestRunMerge_StrandedAfterMerge_NamesRecovery.'
          family: recovery-advice-must-resume
          round: 1
        - id: BR-2
          severity: Minor
          title: Tracker warning says "they match main" in worktree topology where nothing was checked
          detail: 'merge.go:331 prints unconditionally; in a linked worktree mergeRefusal skips the diff. Sibling: the 9b header at merge.go:520 blames a judge/hook, but a peer push to main now triggers it too.'
          family: message-claim-unverified
          round: 1
        - id: BR-3
          severity: Minor
          title: Refuse() has no production callers and is no longer the merge decision (ARCH-DRY)
          detail: Only merge_test.go:138 calls it; mergeRefusal is the real decision. Remove it and assert on Blocking (or mergeRefusal) in TestAssessDirty.
          family: single-decision-source
          round: 1
        - id: BR-4
          severity: Minor
          title: strandingTracker passes root-relative porcelain paths as cwd-relative pathspecs
          detail: Run from a subdirectory, the diff matches nothing and the check passes the file through; glob characters are also treated as patterns. Use --literal-pathspecs with :(top) or run git from the repo root.
          family: pathspec-root-relative
          round: 1
      recipe: small-diff-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T09:16:25-07:00"
      dispatched: "2026-10-12T09:14:43-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: strandedMessage (merge.go:688) says do not commit and re-run merge --yes; 9b text (merge.go:520) says restore or move aside; TestRunMerge_StrandedAfterMerge_NamesRecovery pins "do not commit" for both switch and pull.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Tracker warning now gives a per-topology reason (merge.go:328); the 9b header names main moving as a cause.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Refuse() removed (no callers remain per grep); TestAssessDirty asserts len(Blocking); mergeRefusal is the single decision.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Pathspecs are ":(top,literal)"+p (merge.go:194); the diffRunner fake in TestMergeRefusal rejects any unpinned pathspec.
          round: 2
      recipe: small-diff-review
      isolation: checkout
      reviewed: 7c48c569cd650b3a698f495ac19caacad2a02757
      blocked: false
---

# Gate ledger — ariadne#196 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T08:47:52-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `recovery-advice-must-resume` Stranded/9b messages advise committing, which makes the resume path refuse
  strandedMessage (merge.go:696) and the 9b die text (merge.go:522) suggest commit. A commit on the merged branch trips step 4 (ahead of origin); pushing it trips actionResumeBlocked (#148, origin/main..origin/branch > 0) and the #160 publish gate. Recommend only stash/restore/move-aside then re-run; sweep both sites and assert in TestRunMerge_StrandedAfterMerge_NamesRecovery.
- **BR-2** [Minor] `message-claim-unverified` Tracker warning says "they match main" in worktree topology where nothing was checked
  merge.go:331 prints unconditionally; in a linked worktree mergeRefusal skips the diff. Sibling: the 9b header at merge.go:520 blames a judge/hook, but a peer push to main now triggers it too.
- **BR-3** [Minor] `single-decision-source` Refuse() has no production callers and is no longer the merge decision (ARCH-DRY)
  Only merge_test.go:138 calls it; mergeRefusal is the real decision. Remove it and assert on Blocking (or mergeRefusal) in TestAssessDirty.
- **BR-4** [Minor] `pathspec-root-relative` strandingTracker passes root-relative porcelain paths as cwd-relative pathspecs
  Run from a subdirectory, the diff matches nothing and the check passes the file through; glob characters are also treated as patterns. Use --literal-pathspecs with :(top) or run git from the repo root.

## Round 2 — 2026-10-12T09:16:25-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — strandedMessage (merge.go:688) says do not commit and re-run merge --yes; 9b text (merge.go:520) says restore or move aside; TestRunMerge_StrandedAfterMerge_NamesRecovery pins "do not commit" for both switch and pull.
- BR-2 — addressed — Tracker warning now gives a per-topology reason (merge.go:328); the 9b header names main moving as a cause.
- BR-3 — addressed — Refuse() removed (no callers remain per grep); TestAssessDirty asserts len(Blocking); mergeRefusal is the single decision.
- BR-4 — addressed — Pathspecs are ":(top,literal)"+p (merge.go:194); the diffRunner fake in TestMergeRefusal rejects any unpinned pathspec.

## Open findings

(none — every finding has been disposed)
