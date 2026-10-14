---
gate: plan-quality
issue: 332
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-13T22:16:33-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Default suite facts (has test target, weight) are probed on the branch checkout, not main
          detail: 'requiredChecks reads the declaration from main (cigate.go:318,441), but makeTestFacts(root) runs in the issue branch''s checkout, so a branch that drops or renames its test target or lowers WF_TEST_WEIGHT escapes or weakens the required sdlc-local/suite. Probe main''s Makefile (a worktree at base, or git show base:Makefile) and add a test where a branch deleting test: still owes the suite.'
          family: trust-source-of-enforcement-config
          round: 1
        - id: PQ-2
          severity: Minor
          title: The integration loop does not say it waits for the PR head to equal the pushed merge commit
          detail: The publish gate (landinggate.go:76) and the head-match check in runDurableMerge need pr.HeadOID == local head. After the lease push GitHub updates it asynchronously, so name the re-read (selectLandingPR) and a bounded wait.
          family: post-push-pr-head-sync
          round: 1
        - id: PQ-3
          severity: Minor
          title: The envelope gives no worst-case total wall time across 3 integrations
          detail: 3 x (about 13 min of suite plus up to 20 min of CI wait) is well over an hour in the foreground. State the bound and confirm it is acceptable.
          family: operating-envelope-total-bound
          round: 1
        - id: PQ-4
          severity: Minor
          title: The plan has no explicit non-goals
          detail: 'For example: no rebasing, no automatic conflict resolution, no re-review after a clean integration, no change to the server-side merge method.'
          family: missing-non-goals
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T22:17:13-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: R1 lists the enforcement inputs; default-suite facts now probed in a worktree at main's sha; TestDefaultSuiteFactsComeFromMain added.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Re-reads the PR via selectLandingPR until HeadOID matches (5s poll, 2 min cap, then refuses); lagging-fake test.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Worst case about 105 min plus a 13 min baseline is stated, accepted as batch work, and each limit is printed.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Explicit non-goals list added in R1.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-13T22:17:59-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: d22fafcf695ddbf8e0d4ebb5a12ede1a721b353a04efb9425b386f3b7b922f14
---

# Gate ledger — ariadne#332 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T22:16:33-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `trust-source-of-enforcement-config` Default suite facts (has test target, weight) are probed on the branch checkout, not main
  requiredChecks reads the declaration from main (cigate.go:318,441), but makeTestFacts(root) runs in the issue branch's checkout, so a branch that drops or renames its test target or lowers WF_TEST_WEIGHT escapes or weakens the required sdlc-local/suite. Probe main's Makefile (a worktree at base, or git show base:Makefile) and add a test where a branch deleting test: still owes the suite.
- **PQ-2** [Minor] `post-push-pr-head-sync` The integration loop does not say it waits for the PR head to equal the pushed merge commit
  The publish gate (landinggate.go:76) and the head-match check in runDurableMerge need pr.HeadOID == local head. After the lease push GitHub updates it asynchronously, so name the re-read (selectLandingPR) and a bounded wait.
- **PQ-3** [Minor] `operating-envelope-total-bound` The envelope gives no worst-case total wall time across 3 integrations
  3 x (about 13 min of suite plus up to 20 min of CI wait) is well over an hour in the foreground. State the bound and confirm it is acceptable.
- **PQ-4** [Minor] `missing-non-goals` The plan has no explicit non-goals
  For example: no rebasing, no automatic conflict resolution, no re-review after a clean integration, no change to the server-side merge method.

## Round 2 — 2026-10-13T22:17:13-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — R1 lists the enforcement inputs; default-suite facts now probed in a worktree at main's sha; TestDefaultSuiteFactsComeFromMain added.
- PQ-2 — addressed — Re-reads the PR via selectLandingPR until HeadOID matches (5s poll, 2 min cap, then refuses); lagging-fake test.
- PQ-3 — addressed — Worst case about 105 min plus a 13 min baseline is stated, accepted as batch work, and each limit is printed.
- PQ-4 — addressed — Explicit non-goals list added in R1.

## Round 3 — 2026-10-13T22:17:59-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
