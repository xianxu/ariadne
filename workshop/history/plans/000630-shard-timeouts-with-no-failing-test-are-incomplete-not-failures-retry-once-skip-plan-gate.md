---
reviewer_runs:
    - version: 1
      intent:
        id: 0ac52257-a6eb-4736-90c8-50b5b99f9f02
        repository: github.com/xianxu/ariadne
        issue_id: "000630"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        reviewed_head: cbf1db7e7e3f01b40aca81b8ca84f787c6695fc0
        slot: ariadne:5
        requested_at: "2026-10-25T18:23:22.19439Z"
      phase: finished
      pid: 48447
      started_at: "2026-10-25T18:23:22.447623Z"
      finished_at: "2026-10-25T18:23:47.512904Z"
      session_id: 869956c4-d7c5-479f-b3de-164774b4f7cd
      outcome: returned
gate: plan-quality
issue: 630
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: bc1e4310-d24a-4d7b-bed5-c26a15fcf333
            repository: github.com/xianxu/ariadne
            issue_id: "000630"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: 186107804c53d0de750741670d92b60cd46971ea
            slot: ariadne:5
            requested_at: "2026-10-25T18:15:51.094141Z"
          phase: finished
          pid: 75169
          started_at: "2026-10-25T18:15:51.183971Z"
          finished_at: "2026-10-25T18:17:16.510165Z"
          session_id: 8a27b2a3-0726-4068-9dc4-8843a90f86fa
          outcome: returned
      "n": 1
      timestamp: "2026-10-25T11:17:16-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: M2 says the resume composite is reused unchanged, but the resume child is a known-flake rerun that cannot run a package subset for a suite parent
          detail: resumeLinkVersion gives suite links v3, and validateResumeLink refuses a Selection off v5 (localcheck_resume_evidence.go:51,122). The child batch comes from flakeRerunCommands(Known) plus affectedRerunBatch, which drops GoPkgs (localcheck_resume.go:145-149, localcheck_affected_rerun.go:26). ResumePassed needs KnownFlakePassed (localcheck_resume_evidence.go:369-373). rerunEnvelope gives a groupless child 1 slot and a 15m timeout (localcheck_affected_rerun.go:36-42), which a cmd/sdlc retry under load would hit. Name the new child command, the link selection for suite, the composite pass rule and how guardSuitePass verifies it, and the child envelope.
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: A v5 incomplete parent and its child are refused by the resume and retention version allow-lists (1/2/4), and v5's typed-contract binding is unspecified
          detail: The parent check (localcheck_resume_evidence.go:285), the child check (:354) and retention (localcheck_resume_retention.go:190) admit only v1/2/4. v5 "carries exactly v2's fields", so a typed (v4) suite that times out loses RunnerContractSHA256, which :288 and :349-353 check. List every version gate the sweep must touch, and say whether v5 keeps the runner contract digest.
          family: version-allowlist-sweep
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: The classifyIncomplete text still selects cmd/sdlc for never-ran tests; the revision uses the marker-named package
          family: base-layer-repo-specific-literal
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: 'Tasks 3 and 7 list test cases in prose; compress them to one strategy line per risky function (marker parse and classifier: table plus fuzz test over malformed shard output)'
          family: prose-test-enumeration
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: b6241bf0-a7be-4785-a28f-c33005fcfe17
            repository: github.com/xianxu/ariadne
            issue_id: "000630"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: cbf1db7e7e3f01b40aca81b8ca84f787c6695fc0
            slot: ariadne:5
            requested_at: "2026-10-25T18:21:36.095085Z"
          phase: finished
          pid: 32333
          started_at: "2026-10-25T18:21:36.198604Z"
          finished_at: "2026-10-25T18:22:08.395125Z"
          session_id: aa5f2e71-9573-4a98-b5be-edfbcf946632
          outcome: returned
      "n": 2
      timestamp: "2026-10-25T11:22:08-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Revision items 1-4 name the v5 suite link, incompleteRetryCommands, the IncompleteRetried pass rule through resolveResumeProof/guardSuitePass, and a parent-kind-aware envelope.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Gates :285/:296 branch for incomplete parents; :354 and retention :190 unchanged (v2 child); v5 is written only for untyped runners (localcheck_run.go:407,418).
          round: 2
        - id: PQ-3
          disposition: addressed
          note: The selection is the marker-named packages plus named failures; no cmd/sdlc literal.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: One strategy line per risky function (table plus fuzz for the parser and classifier; pure table for the incomplete branch).
          round: 2
      findings:
        - kind: defect
          id: PQ-5
          severity: Minor
          title: In ariadne (no .sdlc/affected.json on origin/main) the retry reruns the whole suite, not "only the packages that didn't finish"
          detail: Revision item 2 falls back to the full declared suite command without an affected config, so the operator's Done-when holds only in fixtures. At close, show the cost equivalence or get the operator to confirm the Done-when revision (ARCH-PURPOSE).
          family: done-when-unmet-in-default-config
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 32a924be-0654-44a2-b888-fb04868ab11f
            repository: github.com/xianxu/ariadne
            issue_id: "000630"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: cbf1db7e7e3f01b40aca81b8ca84f787c6695fc0
            slot: ariadne:5
            requested_at: "2026-10-25T18:23:10.569144Z"
          phase: finished
          pid: 47137
          started_at: "2026-10-25T18:23:10.66379Z"
          finished_at: "2026-10-25T18:23:19.59071Z"
          session_id: 3a8a073b-aabd-4bfd-9082-ab317284a36f
          outcome: returned
      "n": 3
      timestamp: "2026-10-25T11:23:19-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-5
          disposition: not-addressed
          note: Revision item 2 still runs the whole suite when no affected config exists (ariadne today), justified only by an unmeasured cost claim; carried to close for a measurement or the operator's Done-when confirmation.
          round: 3
      blocked: false
content_hash: 357f6bbbe1abdc41889d14db64d66e07554399199462b063bda0fa0c26d59e15
---

# Gate ledger — ariadne#630 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-25T11:17:16-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` (kind: defect) M2 says the resume composite is reused unchanged, but the resume child is a known-flake rerun that cannot run a package subset for a suite parent
  resumeLinkVersion gives suite links v3, and validateResumeLink refuses a Selection off v5 (localcheck_resume_evidence.go:51,122). The child batch comes from flakeRerunCommands(Known) plus affectedRerunBatch, which drops GoPkgs (localcheck_resume.go:145-149, localcheck_affected_rerun.go:26). ResumePassed needs KnownFlakePassed (localcheck_resume_evidence.go:369-373). rerunEnvelope gives a groupless child 1 slot and a 15m timeout (localcheck_affected_rerun.go:36-42), which a cmd/sdlc retry under load would hit. Name the new child command, the link selection for suite, the composite pass rule and how guardSuitePass verifies it, and the child envelope.
- **PQ-2** [Important] `version-allowlist-sweep` (kind: defect) A v5 incomplete parent and its child are refused by the resume and retention version allow-lists (1/2/4), and v5's typed-contract binding is unspecified
  The parent check (localcheck_resume_evidence.go:285), the child check (:354) and retention (localcheck_resume_retention.go:190) admit only v1/2/4. v5 "carries exactly v2's fields", so a typed (v4) suite that times out loses RunnerContractSHA256, which :288 and :349-353 check. List every version gate the sweep must touch, and say whether v5 keeps the runner contract digest.
- **PQ-3** [Minor] `base-layer-repo-specific-literal` (kind: defect) The classifyIncomplete text still selects cmd/sdlc for never-ran tests; the revision uses the marker-named package
- **PQ-4** [Minor] `prose-test-enumeration` (kind: defect) Tasks 3 and 7 list test cases in prose; compress them to one strategy line per risky function (marker parse and classifier: table plus fuzz test over malformed shard output)

## Round 2 — 2026-10-25T11:22:08-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Revision items 1-4 name the v5 suite link, incompleteRetryCommands, the IncompleteRetried pass rule through resolveResumeProof/guardSuitePass, and a parent-kind-aware envelope.
- PQ-2 — addressed — Gates :285/:296 branch for incomplete parents; :354 and retention :190 unchanged (v2 child); v5 is written only for untyped runners (localcheck_run.go:407,418).
- PQ-3 — addressed — The selection is the marker-named packages plus named failures; no cmd/sdlc literal.
- PQ-4 — addressed — One strategy line per risky function (table plus fuzz for the parser and classifier; pure table for the incomplete branch).

### Raised

- **PQ-5** [Minor] `done-when-unmet-in-default-config` (kind: defect) In ariadne (no .sdlc/affected.json on origin/main) the retry reruns the whole suite, not "only the packages that didn't finish"
  Revision item 2 falls back to the full declared suite command without an affected config, so the operator's Done-when holds only in fixtures. At close, show the cost equivalence or get the operator to confirm the Done-when revision (ARCH-PURPOSE).

## Round 3 — 2026-10-25T11:23:19-07:00 (claude) — passed

### Disposed

- PQ-5 — not-addressed — Revision item 2 still runs the whole suite when no affected config exists (ariadne today), justified only by an unmeasured cost claim; carried to close for a measurement or the operator's Done-when confirmation.

## Open findings

- **PQ-5** [Minor] `done-when-unmet-in-default-config` (kind: defect) In ariadne (no .sdlc/affected.json on origin/main) the retry reruns the whole suite, not "only the packages that didn't finish"
