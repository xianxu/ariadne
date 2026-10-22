---
id: 000546
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'fd7fed45a5dc40fbe906e93ea9c6c70caf094bc6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T03:02:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# issue migrate inventory omits archive-only branches and other-worktree history evidence

## Problem

Pre-existing issue migrate inventory omits archived issue evidence outside pinned main. Discovered by #543 close review round 2 (BR-2, authoritative-inventory-completeness); TL explicitly moved importer completeness out of new-repo init's scope.

Reproductions against real Git/bare origin: (1) clone B creates and pushes a branch containing workshop/history/issues/000001-old.md while main has no issues; clone A migrationInventory scans branch edits only under workshop/issues and plans an empty tracker. (2) another linked live worktree contains uncommitted workshop/history/issues/000001-old.md; dirtyIssuePaths scans only workshop/issues, again allowing an empty plan. The same roots/scope rule must examine flat legacy history and detached/inherited branch/worktree evidence, rather than patching only these two sites.

## Spec

Main remains the only source of imported cards. Use the direct-child issue filename rule across the three roots from `issue.IDDirs`: active `workshop/issues`, flat legacy `workshop/history`, and canonical `workshop/history/issues`. Plans, nested side bundles and assets are excluded. Observation is complete before any consumer policy; init retains its independent presence-only policy.

| Scope | Active root | Both archive roots |
|---|---|---|
| Pinned publication main | Import and mirror, using existing card/close proof rules | Import cards using existing duplicate selection; never rewrite archived details |
| Every pinned local/current publication-remote tip and live worktree HEAD, including detached | Refuse a present ID absent from main's imported population, even if inherited. For related tips retain existing merge-base changed-file policy for card fields, active rename/deletion and close anchors. For unrelated tips require exact published active records; otherwise refuse rather than infer ancestry. | A present archive must match a published archive's filename ID and blob bytes (path may differ); otherwise refuse, naming its carrier and path. Main remains authoritative for absences: missing archives on stale tips are not invented deletions. |
| Every live worktree index and filesystem | Refuse uncommitted issue edits, additions and deletions | Same rule, including ignored/untracked direct-child documents |
| Unknown, unreadable, malformed or non-directory root observation | Refuse before tracker/main publication | Same rule |

Every branch name is observed, including main, resting and reserved names. Deduplicate immutable reads by OID while retaining carrier names. Verified partial tracker bootstrap and completed migration recovery remain the existing exact-root/marker proofs; no new durable state or publication phase. Bare/prunable worktrees retain their explicit skip policy. The local worktree HEAD is pinned, including detached and unrelated histories. Root components and candidate file kinds cannot silently hide evidence.

The archive rule is intentionally conservative: even an old branch's differing archive copy refuses until the operator lands, aligns or drops that evidence under the legacy workflow. It does not import branch-only cards or decide which differing historical record should win. Related inherited active copies of already represented IDs keep their existing semantics; they are not fabricated edits. Archive observations never enter the active `Branches` rename/reconcile input.

For worktrees, retain Git's tracked-content comparison (including attribute normalization), use NUL-delimited worktree/index/status parsing, and inspect direct filesystem entries against the index so ignored issue documents cannot disappear. Conflicted indexes and issue entries whose assume-unchanged/skip-worktree flags prevent trustworthy observation refuse with the named remedy. No raw-byte filesystem-versus-index comparison that mistakes CRLF normalization for lost work.

The operator's existing freeze of all legacy SDLC writers remains a precondition during dry run/apply. A local lock cannot freeze other clones or editors; this work does not claim a distributed absence guarantee. Apply recollects the complete inventory before checking the reviewed digest. Existing tracker-root CAS, main cutover CAS, exact retry and post-cutover reconcile remain unchanged.

## Done when

- Real-Git regressions reproduce the remote archive-only branch and other-worktree dirty archive cases and prevent silent empty migration.
- A reviewed root-by-scope inventory rule states what is imported, what refuses, and how unknown observations fail closed.
- Existing active branch rename/archive and provable legacy-close migration behavior remains verified.

## Plan

Durable plan: [complete migration inventory](../plans/000546-migration-inventory-plan.md). Full flow, one atomic close boundary.

- [ ] Implement typed evidence and pure matching while preserving related active reconciliation and close proofs.
- [ ] Reuse the complete observer across init/migrate; cover all roots, tips and live worktree evidence with main-only imports.
- [ ] Verify counted growth, private production CLI/native Git conformance and unchanged migration recovery; document, suite, close and land.

## Log

Filed unclaimed per TL direction, Couch 5119a959-0128-4420-89ee-9a91fb0b08d9. Review reproduction was against #543 pinned head 96e8ff1cfaf539a80c5c174ce4ba7bbf83992bae; the omission predates #543 in issuemigrate.go (branch/worktree scopes use d.Home only). No importer expansion authorized in #543.

### 2026-10-22 — planning inventory completeness

Claimed and started on a fresh main-based issue branch after #543's confirmed landing/cleanup. Read the current importer and real-Git regression contracts. Main's archive scan omits the flat root; branch deltas omit archive and inherited/unrelated presence; worktree dirtiness omits both archive roots and detached committed evidence. A bounded read-only survey independently confirmed these gaps and the active rename trap.

Alternatives: (A, recommended) complete observation, main-only imports and conservative refusal of unrepresented non-main evidence; (B) import/merge records from every carrier, rejected because it needs new authority and conflict-selection rules; (C) patch the two reproduced scan sites, rejected because inherited/detached/name-excluded evidence remains invisible. ARCH-PURPOSE requires the root × scope class; ARCH-DRY keeps one path/root observer and separate init/migrate policy; ARCH-PURE keeps archive/presence decisions in the existing pure migration planner. No implementation begun; spec sent to TL for approval before the detailed plan.

### 2026-10-22 — spec approved; durable plan frozen for review

TL approved the complete root-by-scope spec and private production CLI/native bare-Git conformance with no stage (Couch 0fc506bf-9240-416e-b647-898eb9f309a1). A wording check explicitly preserved related-active merge-base/card-field reconciliation and allowed prose differences; byte matching applies to archive evidence and unrelated active snapshots only (7a42fef5-cd44-4876-b333-aff944ce8657). Detailed plan now names pure entities, shared observation extraction, finite refusal/recovery table, the exact preserved authority boundary, counted growth and adversarial guards. No implementation or recorded tests started. Await TL plan approval before change-code.
