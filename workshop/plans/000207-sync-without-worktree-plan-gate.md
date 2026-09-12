---
gate: plan-quality
issue: 207
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-09T18:33:01-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: takenOnTrunk "any slug" renumbers every issue already published on the trunk
          detail: |-
            The transform sketch aborts with ErrIDTaken whenever any file at this id
            exists on trunk, but syncIssuesToMain is shared by `sdlc issue sync`
            (cmd/sdlc/issue.go:423) and `sdlc change-code` (cmd/sdlc/changecode.go:283),
            whose job is republishing a body the trunk already carries. Define the
            foreign-claimant test (differing basename), and say why it does not fire
            for the archived-elsewhere shape (history/issues/NNNNNN-same-slug.md) or a
            slug change, and whether reservation and update are one path or two.
          family: id-identity-test
          round: 1
        - id: PQ-2
          severity: Critical
          title: TrunkFile.Update writes one path per commit; the arm it replaces publishes N files in one commit
          detail: |-
            syncViaMainWorktree commits the whole changed set once (cmd/sdlc/claim.go:449-500),
            and changedIssueFiles returns every changed issue file when f.Issue == 0 — bare
            `sdlc claim`. The plan does not choose between N sequential Update calls
            (N commits, N pushes, partial publish on failure, N fetches on an interactive
            path — ARCH-CONSTRAINTS has no budget stated) and widening the gitx seam
            shipped in ariadne#209. That choice is a seam decision, not an implementation detail.
          family: publish-unit-mismatch
          round: 1
        - id: PQ-3
          severity: Important
          title: takenOnTrunk re-implements refIDSpace, the single-source trunk id-space reader
          detail: |-
            cmd/sdlc/issueids.go:143 documents refIDSpace as THE reader precisely because
            four separate ls-tree loops each grew their own failure policy. It scans
            issue.IDDirs — three dirs including the #181 history/issues/ archive subdir
            (internal/issue/scaffold.go:35) — via issue.PathsByID. The plan's "issue +
            history dirs" loop is a fifth one and as worded omits the archive subdir, the
            exact shape the pair field evidence calls the worst. Name refIDSpace as consumed.
          family: single-id-space-reader
          round: 1
        - id: PQ-4
          severity: Important
          title: '"re-allocate + rename + retry, bounded" under-specifies the riskiest step'
          detail: |-
            Unstated: the frontmatter id rewrite, behavior when the file is already
            committed on the feature branch by a prior no-push sync (the state ariadne#206
            creates), the numeric bound, and the state left by a crash between rename and
            push. ARCH-ORDER wants the interrupting events named and targeted — process
            death mid-re-allocation, a second publisher taking the new id, a push accepted
            with a lost response.
          family: reallocation-mechanics
          round: 1
        - id: PQ-5
          severity: Minor
          title: The transform signature cannot express a deleted issue file
          detail: |-
            changedIssueFiles can report a deletion (see syncPathspec's comment,
            cmd/sdlc/claim.go:275), and commitAndPush only does update-index --add.
            State that deletion is out of scope and that it fails loudly rather than
            publishing nothing.
          family: transform-cannot-delete
          round: 1
        - id: PQ-6
          severity: Minor
          title: Problem says syncInPlace's missing fetch is fixed incidentally; Scope says it is a separate fix
          family: scope-statement-contradiction
          round: 1
        - id: PQ-7
          severity: Minor
          title: The issue body triplicates one Spec paragraph, once inside the Plan section
          detail: |-
            workshop/issues/000207-sync-without-worktree.md lines 98, 108 and 165 carry the
            same paragraph; line 165 sits between the Plan header and the Log header.
            Residue of the 021e6bb consolidation.
          family: artifact-hygiene
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-09T18:39:42-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Caller-split table (lines 88-91) fixes the renumbering; basename keying makes the archived-elsewhere shape safe.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: UpdateMany is chosen explicitly as a seam widening rather than N sequential Updates.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Spec now names refIDSpace as consumed, with a Done-when shadow sweep against a second reader.
          round: 2
        - id: PQ-4
          disposition: not-addressed
          note: Frontmatter rewrite named; the bound, the crash state, and the already-committed-on-branch case are still unstated.
          round: 2
        - id: PQ-5
          disposition: not-addressed
          note: Deletion is still unmentioned anywhere in Spec, Done-when or Plan.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Problem section now explicitly says syncInPlace's fetch is NOT fixed here.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Verified no duplicated paragraph remains; the Plan/Log boundary is clean.
          round: 2
      findings:
        - id: PQ-8
          severity: Important
          title: The widened UpdateMany contract is written as an ellipsis at the parameter that decides whether the CAS retry re-lands a collision
          detail: |-
            2nd in family — the rule, not the instance: when this plan widens the gitx
            seam, the widened contract must be stated in full (signature, per-path
            unchanged semantics, and what is re-evaluated inside the CAS window),
            because that seam's documented value is that the interleaving policy is
            legible at the seam (trunkfile.go:335-342). Prevalence 2: PQ-1 left
            "reservation and update, one path or two" unanswered, and issue line 70
            now reads UpdateMany(files map string byte, msg string, transform …) with
            a literal ellipsis. Concretely: files-as-content contradicts a transform
            that receives old bytes (trunkfile.go:369); the unchanged-content early
            return (trunkfile.go:377) is per-file or whole-set and that choice
            preserves filesDifferingFrom's idempotence (claim.go:398-420); and the
            plan never says the collision decision runs INSIDE the retry loop, so a
            peer landing mid-window causes rejection then retry then a re-push of the
            same id, landing the duplicate as a clean fast-forward. The plan's own
            step-1 test passes on that defect unless it seeds the collision during the
            retry window rather than before the first check.
          family: publish-unit-mismatch
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-09T18:44:09-07:00"
      agent: claude
      dispose:
        - id: PQ-8
          disposition: addressed
          note: Full signature written out; per-attempt prepare puts the collision decision inside the CAS loop, and test (b) now seeds mid-retry and asserts prepare ran twice.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Five numbered steps; residual for close review — step 3 implies the local os.Rename happens after the push succeeds, so the crash window between the two is unnamed.
          round: 3
        - id: PQ-5
          disposition: not-addressed
          note: Plan still says nothing about a deleted issue file; omitting the path hits the whole-set early return and reports success where today's arm fails loudly (claim.go:459).
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-09-09T18:47:34-07:00"
      agent: claude
      dispose:
        - id: PQ-5
          disposition: addressed
          note: 'Went past the finding: Delete is representable in TrunkWrite, with the whole-set early return covering both halves and a Done-when test.'
          round: 4
      blocked: false
content_hash: 9acc0b775d18c21a8e47cfbe3689a8adf1b874761bc4a527d36773fb4442e466
---

# Gate ledger — ariadne#207 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-09T18:33:01-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] `id-identity-test` takenOnTrunk "any slug" renumbers every issue already published on the trunk
  The transform sketch aborts with ErrIDTaken whenever any file at this id
  exists on trunk, but syncIssuesToMain is shared by `sdlc issue sync`
  (cmd/sdlc/issue.go:423) and `sdlc change-code` (cmd/sdlc/changecode.go:283),
  whose job is republishing a body the trunk already carries. Define the
  foreign-claimant test (differing basename), and say why it does not fire
  for the archived-elsewhere shape (history/issues/NNNNNN-same-slug.md) or a
  slug change, and whether reservation and update are one path or two.
- **PQ-2** [Critical] `publish-unit-mismatch` TrunkFile.Update writes one path per commit; the arm it replaces publishes N files in one commit
  syncViaMainWorktree commits the whole changed set once (cmd/sdlc/claim.go:449-500),
  and changedIssueFiles returns every changed issue file when f.Issue == 0 — bare
  `sdlc claim`. The plan does not choose between N sequential Update calls
  (N commits, N pushes, partial publish on failure, N fetches on an interactive
  path — ARCH-CONSTRAINTS has no budget stated) and widening the gitx seam
  shipped in ariadne#209. That choice is a seam decision, not an implementation detail.
- **PQ-3** [Important] `single-id-space-reader` takenOnTrunk re-implements refIDSpace, the single-source trunk id-space reader
  cmd/sdlc/issueids.go:143 documents refIDSpace as THE reader precisely because
  four separate ls-tree loops each grew their own failure policy. It scans
  issue.IDDirs — three dirs including the #181 history/issues/ archive subdir
  (internal/issue/scaffold.go:35) — via issue.PathsByID. The plan's "issue +
  history dirs" loop is a fifth one and as worded omits the archive subdir, the
  exact shape the pair field evidence calls the worst. Name refIDSpace as consumed.
- **PQ-4** [Important] `reallocation-mechanics` "re-allocate + rename + retry, bounded" under-specifies the riskiest step
  Unstated: the frontmatter id rewrite, behavior when the file is already
  committed on the feature branch by a prior no-push sync (the state ariadne#206
  creates), the numeric bound, and the state left by a crash between rename and
  push. ARCH-ORDER wants the interrupting events named and targeted — process
  death mid-re-allocation, a second publisher taking the new id, a push accepted
  with a lost response.
- **PQ-5** [Minor] `transform-cannot-delete` The transform signature cannot express a deleted issue file
  changedIssueFiles can report a deletion (see syncPathspec's comment,
  cmd/sdlc/claim.go:275), and commitAndPush only does update-index --add.
  State that deletion is out of scope and that it fails loudly rather than
  publishing nothing.
- **PQ-6** [Minor] `scope-statement-contradiction` Problem says syncInPlace's missing fetch is fixed incidentally; Scope says it is a separate fix
- **PQ-7** [Minor] `artifact-hygiene` The issue body triplicates one Spec paragraph, once inside the Plan section
  workshop/issues/000207-sync-without-worktree.md lines 98, 108 and 165 carry the
  same paragraph; line 165 sits between the Plan header and the Log header.
  Residue of the 021e6bb consolidation.

## Round 2 — 2026-09-09T18:39:42-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Caller-split table (lines 88-91) fixes the renumbering; basename keying makes the archived-elsewhere shape safe.
- PQ-2 — addressed — UpdateMany is chosen explicitly as a seam widening rather than N sequential Updates.
- PQ-3 — addressed — Spec now names refIDSpace as consumed, with a Done-when shadow sweep against a second reader.
- PQ-4 — not-addressed — Frontmatter rewrite named; the bound, the crash state, and the already-committed-on-branch case are still unstated.
- PQ-5 — not-addressed — Deletion is still unmentioned anywhere in Spec, Done-when or Plan.
- PQ-6 — addressed — Problem section now explicitly says syncInPlace's fetch is NOT fixed here.
- PQ-7 — addressed — Verified no duplicated paragraph remains; the Plan/Log boundary is clean.

### Raised

- **PQ-8** [Important] `publish-unit-mismatch` The widened UpdateMany contract is written as an ellipsis at the parameter that decides whether the CAS retry re-lands a collision
  2nd in family — the rule, not the instance: when this plan widens the gitx
  seam, the widened contract must be stated in full (signature, per-path
  unchanged semantics, and what is re-evaluated inside the CAS window),
  because that seam's documented value is that the interleaving policy is
  legible at the seam (trunkfile.go:335-342). Prevalence 2: PQ-1 left
  "reservation and update, one path or two" unanswered, and issue line 70
  now reads UpdateMany(files map string byte, msg string, transform …) with
  a literal ellipsis. Concretely: files-as-content contradicts a transform
  that receives old bytes (trunkfile.go:369); the unchanged-content early
  return (trunkfile.go:377) is per-file or whole-set and that choice
  preserves filesDifferingFrom's idempotence (claim.go:398-420); and the
  plan never says the collision decision runs INSIDE the retry loop, so a
  peer landing mid-window causes rejection then retry then a re-push of the
  same id, landing the duplicate as a clean fast-forward. The plan's own
  step-1 test passes on that defect unless it seeds the collision during the
  retry window rather than before the first check.

## Round 3 — 2026-09-09T18:44:09-07:00 (claude) — passed

### Disposed

- PQ-8 — addressed — Full signature written out; per-attempt prepare puts the collision decision inside the CAS loop, and test (b) now seeds mid-retry and asserts prepare ran twice.
- PQ-4 — addressed — Five numbered steps; residual for close review — step 3 implies the local os.Rename happens after the push succeeds, so the crash window between the two is unnamed.
- PQ-5 — not-addressed — Plan still says nothing about a deleted issue file; omitting the path hits the whole-set early return and reports success where today's arm fails loudly (claim.go:459).

## Round 4 — 2026-09-09T18:47:34-07:00 (claude) — passed

### Disposed

- PQ-5 — addressed — Went past the finding: Delete is representable in TrunkWrite, with the whole-set early return covering both halves and a Done-when test.

## Open findings

(none — every finding has been disposed)
