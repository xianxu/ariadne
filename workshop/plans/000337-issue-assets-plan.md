# Issue assets implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Use superpowers-executing-plans for session-warm integration; delegate bounded pure-core tasks when useful. Steps use checkboxes.

**Goal:** Attach bounded evidence to an issue with a binary-enforced repository privacy policy, preserving its bytes through every supported in-repository lifecycle operation.

**Architecture:** The body indexes an issue's asset bundle. One shared bundle planner produces exact writes/deletes for existing publication and lifecycle transactions. Repository policy is resolved before attachment IO; agents never change that policy.

**Tech Stack:** Go, Cobra, existing Git transaction and GitHub CLI seams, portable temporary Git repositories for integration tests.

## Approval and scope

Draft for hub/operator review; no implementation authorized by this document alone.
The hub directed the in-repository scope on 2026-10-13. Cross-repository issue
migration remains refused, tracked separately by #347, dependent on #337 and on
implementing the issue-renumbering capability deferred by #179. Do not silently
reinterpret the existing migrate command as supporting issues.

## Design decisions

- Use plain Git. LFS adds a second publication/availability dependency, and
  external pointers break self-contained evidence and encrypted/local workflows.
  No recordings, remote uploads, or recording/replay links are asset dependencies.
- Proposed operating limits: 1 MiB per file (matches gitx.SnapshotBlobLimit),
  8 MiB total and 32 files per issue. Refuse before mutation; never silently
  compress, truncate, or relax existing Git reader limits. These are design
  choices for review, not measured crop-size requirements.
- Active paths: `<issues-root>/assets/NNNNNN/<sha256>.<safe-extension>`;
  archived paths: `<history-root>/issues/assets/NNNNNN/<same-name>`.
  The body uses `assets/NNNNNN/<name>` relative links, unchanged by archive or slug
  rename. Custom roots follow the existing vocabulary archive layout.
- Content names are immutable. Reattaching identical bytes is idempotent;
  conflicting bytes at an existing name refuse. The body is the only index:
  parse actual Markdown destinations, ignoring fences/code, not arbitrary regex
  substrings. No secondary manifest. Directory enumeration only detects invalid
  residue; it never chooses evidence to publish.
- `sdlc issue attach --issue N --file PATH --caption TEXT` copies a regular file
  and adds its link to an optional `## Assets` section. Images use image links;
  other files use ordinary links. Escape caption/path syntax. No automatic commit
  or publication; print the changed paths and existing next-step commands.
- `workshop/asset-policy.json` is operator-owned, versioned, repo-local, and is
  not propagated by the base layer. Proposed strict shape:
  `{"version":1,"policy":"allowed"}` or `"text-only"`. No setter or override
  flag. Existing invalid/unreadable policy refuses attachment; help reports it.
  Missing policy uses the effective publication host repository's visibility:
  private permits assets; public or unknown means text-only. Text-only refuses
  all file attachments, eliminating MIME/extension renaming bypasses; prose
  remains in the issue body. Help prints effective policy and its exact source.
- Resolve visibility through the shared gh seam with bounded context (5 seconds),
  strict response parsing, effective remote identity checks, and no persistent
  cache. Unknown/offline/unsupported host is text-only with a reason. Do not print
  credentials or raw subprocess stderr. Help outside a repo still works and
  explains that no repository policy was resolved.
- Pair submission privacy is a separate prerequisite. Its canonical couch skill
  requires explicit `privacy: shareable`; private/missing/unknown means sanitized
  text summary and no crops. An allowed repo policy never grants source sharing.
  Document this distinction and preserve the pair-owned rule; this issue does not
  parse arbitrary recording stores or infer privacy from a filename.
- Attach obeys existing issue ownership, mirror validation, and repository lock.
  Reject foreign issue IDs, traversal, symlinked sources/destinations/ancestors,
  nonregular files, missing linked assets, ambiguous links and destination clashes.
  A refusal leaves body, files, index, and refs unchanged.
- Existing published assets may still archive/restore if policy later tightens;
  these move existing evidence, they do not admit new evidence. Publishing new
  asset bytes rechecks current destination policy and ownership. Exact byte and
  bundle checks apply to binary-only diffs as well as body edits.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| AssetPolicy / PolicyDecision | cmd/sdlc/internal/issueassets/policy.go | new |
| AssetRef / Bundle / bundle validation and delta | cmd/sdlc/internal/issueassets/bundle.go | new |
| Attach transition (prepared, bytes-installed, body-linked) | cmd/sdlc/internal/issueassets/attach.go | new |
| Publication bundle equality / conflict decision | cmd/sdlc/republish.go | modified |
| Landing archive plan and proof | cmd/sdlc/landingarchive.go | modified |

An issue owns one bundle with zero to 32 immutable members. Pure functions parse
links/policy, calculate names/paths/limits, and plan exact changes from snapshots.
The same bundle and path rules serve attach, publish, restore, and archives.
Unit tests are colocated and use bytes, without IO mocks.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Asset filesystem reader/writer | cmd/sdlc/issueassets.go | new | bounded file IO and existing repo lock |
| Attach CLI and policy help | cmd/sdlc/issueattach.go; cmd/sdlc/issue.go | new / modified | Cobra, ownership and policy |
| Visibility query | cmd/sdlc/assetpolicy.go; cmd/sdlc/ghclient.go | new / modified | context-bounded gh query |
| Publication and ownership | cmd/sdlc/republish.go; cmd/sdlc/issuemovedetail.go; cmd/sdlc/transferguard.go | modified | pinned Git snapshots and TrunkWrite |
| Restore / abandon / archive | cmd/sdlc/issuerestore.go; cmd/sdlc/landingarchive.go; cmd/sdlc/archivemain.go; cmd/sdlc/abandon.go; cmd/sdlc/push.go; cmd/sdlc/merge.go | modified | existing lifecycle transactions |

Reuse gitx snapshots, ExactBytes, optimistic publication retries and recovery
receipts. Extend the existing GitHub fake with per-repository visibility/error
state; test visibility changes between commands. Use temporary bare Git remotes
for conformance of publication and retry; never production mutations. A read-only
live visibility query against this repo validates the new gh argv once, if
available. No new external storage service.

## Ordering, trust and lifetime

ARCH-DRY / ARCH-PURE: one link/bundle model and one writer adapter; commands supply
snapshots and execute plans. Avoid separate regex inventories in each lifecycle.
ARCH-PURPOSE: audit every consumer below, including legacy modes and abandon.
ARCH-MOCK: portable Git remotes and persisted fake host visibility exercise IO.
ARCH-CONSTRAINTS: interactive one-file operation, bounded bytes/count, one bounded
host query, no fan-out. Full suites run only after uptime load is below 40.
ARCH-SECURE: parse files/host responses as untrusted; no agent policy edits. Read
and validate all inputs before effects, recheck expected body and ownership under
lock; compare bytes rather than assume a file stayed unchanged after stat.

ARCH-ORDER: attach's pure transition owns the progression from validated snapshot
to installed immutable bytes to linked body. An interruption after bytes-install
leaves a recognizable content-named file; retry validates it and finishes the
body link, never duplicates a link or overwrites changed prose. Write temporary
files beside destinations, rename atomically, and remove temporary files on
ordinary failure. On a hard kill, bounded owned temporary names are reconciled
on the next attachment; unknown files refuse for inspection, never auto-delete.
Publish/archive use existing atomic Git commits: unknown push outcome is probed
against the whole bundle, not body equality. Later changes make stale operations
refuse. No background work survives command return.

ARCH-FUNERAL: attach creates bytes, issue readers are last consumers, lifecycle
archive/restore moves the bundle. Removing a body link removes the current-tree
asset through the bundle writer at publication; retained Git history is intentional
and costs at most 1 MiB per distinct attachment (no automatic per-event capture).
The 8 MiB live issue budget is not a total-history bound: replacement accumulates
history; document that cost and refuse oversized writes. Temporary files end at
success/failure/reconciliation. No attachment ledger or second persistent index.

## Task 1 — Policy, bundle and attachment

Files: new issueassets core files and colocated tests; new `issueassets.go`,
`issueattach.go`, `assetpolicy.go` and tests; modify `issue.go`, `ghclient.go`,
`fetch_test.go` (shared fake), `main.go` only if needed for dynamic help.

- [ ] Write table tests for strict policies; public/private/unknown/error visibility;
  fenced/example links, ordinary/image/reference-style destinations, duplicate
  links, traversal, wrong IDs, symlinks, caps and exact cap boundaries.
- [ ] Run `go test ./cmd/sdlc/internal/issueassets -count=1` and demonstrate RED.
- [ ] Implement typed pure policy, bundle parsing, path mapping and attach transition.
- [ ] Add CLI tests proving policy refusal happens before any source copy; renamed
  images still refuse under text-only; malformed config never falls back to allowed;
  ownership/refusal leaves all files/refs/index unchanged.
- [ ] Test permitted image and nonimage attach, identical retry, body race, injected
  failures before/after install and before body rename, missing bytes and collisions.
- [ ] Implement the thin IO/CLI adapters and dynamic help using one policy resolver.
- [ ] Run targeted core and `go test ./cmd/sdlc -run 'Test(Asset|IssueAttach)' -count=1`.
- [ ] Commit implementation and update this plan with actual file/entity names.

## Task 2 — Publication, recovery and ownership

Files: `republish.go`, `issuemovedetail.go`, `internal/tracker/detailpublish.go`,
`issuepublish.go`, `issuerestore.go`, `transferguard.go`, `transferverdict.go`;
their existing tests plus focused bundle integration tests.

- [ ] Write failing tests for tracker publication (including binary-only changes),
  legacy commit publication, first-publication recovery, narrow commits and restore.
- [ ] Add ownership/stale-base tests where body is unchanged and only bytes change.
  Body and asset admission must not be bypassable through a direct PR/push path.
- [ ] Extend publication read sets and writes to complete bundles, retaining existing
  ownership rechecks, filters bypass, receipt proofs and optimistic retry semantics.
- [ ] Exercise lost-response retry, concurrent source edits, concurrent main edits,
  untracked assets blocking fast-forward and a later policy tightening. No unrelated
  files enter the publish commit. Main never receives a body with missing bytes.
- [ ] Run `go test ./cmd/sdlc -run 'Test(Asset|IssuePublish|MoveDetail|Republish|Transfer|IssueRestore)' -count=1`
  and `go test ./cmd/sdlc/internal/tracker -run 'Test.*Detail' -count=1`.
- [ ] Commit the publication integration.

## Task 3 — Lifecycle sweep

Files: `landingarchive.go`, `archivemain.go`, `abandon.go`, `push.go`, `merge.go`,
shared archive helpers as identified during implementation; `landingarchive_test.go`,
`archiveartifacts_test.go`, `abandon_test.go`, `migrate_test.go`, slot move tests.

- [ ] Write failures for archive bytes/path preservation, configured roots, duplicate
  names across issues, missing member, changed member, incomplete archive proof,
  collision, and interrupted legacy archive recovery.
- [ ] Extend pinned archive snapshots/proofs and exact writes/deletes with bundles.
- [ ] Prove abandon uses kept-branch assets absent from main, restore/reopen preserves
  exact generation, and retries refuse new conflicts rather than losing evidence.
- [ ] Prove slug rename and slot move preserve links/bytes without a special copy
  path. Preserve and test cross-repo issue migrate refusal before any mutation.
- [ ] Run `go test ./cmd/sdlc -run 'Test.*(Asset|Archive|Abandon|Restore|Move|Migrate)' -count=1`.
- [ ] Commit the lifecycle integration.

## Task 4 — Documentation and acceptance

Files: `cmd/sdlc/helptext/issue.md`, relevant publish/migrate/recovery help,
`atlas/workflow/issue-assets.md` (new), `atlas/index.md`, `README.md`, issue Log.

- [ ] Document operator-owned policy, visibility source and unknown behavior, limits,
  immutable plain-Git history cost, attach/publish workflow, private note summaries,
  no recording dependency, archive paths, retry behavior and migration exclusion.
- [ ] Add a temporary-repository CLI test: text-only refuses, operator fixture policy
  permits attachment, publish carries bytes, archive resolves the unchanged link,
  restore/slot move retains evidence. Assert exact bytes, paths and refs.
- [ ] Run focused tests, `git diff --check`, and review final diff against main.
- [ ] Check `uptime`; only with load <40 start `make test` in background. Keep the
  checkout unchanged until the full suite completes; save exact results.
- [ ] Run `sdlc close --issue 337 --verified '<actual evidence>'`; its boundary review
  is the one required fresh-eyes review. Resolve findings and log measured results.
- [ ] Send the hub PR and real `git rev-parse HEAD` hash. On SHIP, merge fresh main,
  satisfy any revalidation/CI gates, run `sdlc merge --yes`, report and stop.

## Revisions

- 2026-10-13 — Initial implementation draft after repository exploration and hub
  scope direction. Cross-repo issue migration is excluded explicitly (#347);
  intra-repo lifecycle includes abandonment and legacy paths to avoid losing bytes.
