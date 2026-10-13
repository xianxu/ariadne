---
gate: boundary-review
issue: 337
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T15:04:03-07:00"
      dispatched: "2026-10-13T14:58:58-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Publication trusts stale checkout policy instead of current destination policy
          detail: 'cmd/sdlc/assetpolicy.go:23 reads only the working-tree policy; republish.go and assetbundlegit.go reuse it during publication. A temporary regression tightened main to text-only while leaving the caller allowed, and publication still pushed new asset bytes. ARCH-SECURE / ARCH-PURPOSE: resolve authorization from the pinned destination snapshot on every publication attempt and sweep all admission callers, including first-publication recovery, legacy publication, and abandonment.'
          family: destination-policy-authority
          round: 1
        - id: BR-2
          severity: Critical
          title: Reopening can commit converted asset bytes under their original content name
          detail: 'cmd/sdlc/abandon.go:464 merges and materializes archived assets before line 504 checks only active destinations. A temporary regression using archive-only text eol=crlf and active -text reopened successfully but changed LF evidence to CRLF under its unchanged hash filename. ARCH-PURPOSE: validate all materialization paths before effects and verify restored bytes against the pinned bundle; add a production reopen regression.'
          family: exact-bytes-at-materialization
          round: 1
        - id: BR-3
          severity: Critical
          title: The core-concepts table incorrectly declares archive proof PURE
          detail: 'workshop/plans/000337-issue-assets-plan.md:82 places landing archive plan and proof in the Pure entity table, while cmd/sdlc/landingarchive.go:556 confirmLandingArchive executes Git, loads snapshots, and reads time.Now; its proof tests use temporary repositories. ARCH-PURE: retain the planner as PURE, classify the proof as INTEGRATION, and append a Revisions entry correcting the table.'
          family: core-concept-classification
          round: 1
        - id: BR-4
          severity: Important
          title: Host visibility has no live conformance check
          detail: 'cmd/sdlc/assetvisibility_test.go:11 supplies canned JSON through a replacement gh executable, and assetpolicy_test.go models visibility with a state file; neither verifies the real RepoVisibility invocation against the host. ARCH-MOCK: add an opt-in read-only live conformance check through the production seam and document when it runs.'
          family: external-dependency-conformance
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#337 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T15:04:03-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `destination-policy-authority` Publication trusts stale checkout policy instead of current destination policy
  cmd/sdlc/assetpolicy.go:23 reads only the working-tree policy; republish.go and assetbundlegit.go reuse it during publication. A temporary regression tightened main to text-only while leaving the caller allowed, and publication still pushed new asset bytes. ARCH-SECURE / ARCH-PURPOSE: resolve authorization from the pinned destination snapshot on every publication attempt and sweep all admission callers, including first-publication recovery, legacy publication, and abandonment.
- **BR-2** [Critical] `exact-bytes-at-materialization` Reopening can commit converted asset bytes under their original content name
  cmd/sdlc/abandon.go:464 merges and materializes archived assets before line 504 checks only active destinations. A temporary regression using archive-only text eol=crlf and active -text reopened successfully but changed LF evidence to CRLF under its unchanged hash filename. ARCH-PURPOSE: validate all materialization paths before effects and verify restored bytes against the pinned bundle; add a production reopen regression.
- **BR-3** [Critical] `core-concept-classification` The core-concepts table incorrectly declares archive proof PURE
  workshop/plans/000337-issue-assets-plan.md:82 places landing archive plan and proof in the Pure entity table, while cmd/sdlc/landingarchive.go:556 confirmLandingArchive executes Git, loads snapshots, and reads time.Now; its proof tests use temporary repositories. ARCH-PURE: retain the planner as PURE, classify the proof as INTEGRATION, and append a Revisions entry correcting the table.
- **BR-4** [Important] `external-dependency-conformance` Host visibility has no live conformance check
  cmd/sdlc/assetvisibility_test.go:11 supplies canned JSON through a replacement gh executable, and assetpolicy_test.go models visibility with a state file; neither verifies the real RepoVisibility invocation against the host. ARCH-MOCK: add an opt-in read-only live conformance check through the production seam and document when it runs.

## Open findings

- **BR-1** [Critical] `destination-policy-authority` Publication trusts stale checkout policy instead of current destination policy
- **BR-2** [Critical] `exact-bytes-at-materialization` Reopening can commit converted asset bytes under their original content name
- **BR-3** [Critical] `core-concept-classification` The core-concepts table incorrectly declares archive proof PURE
- **BR-4** [Important] `external-dependency-conformance` Host visibility has no live conformance check
