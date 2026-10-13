# Issue assets implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Use superpowers-executing-plans for session-warm integration; delegate bounded pure-core tasks when useful. Steps use checkboxes.

**Goal:** Attach bounded evidence to an issue with a binary-enforced repository privacy policy, preserving its bytes through every supported in-repository lifecycle operation.

**Architecture:** The body indexes an issue's asset bundle. One shared bundle planner produces exact writes/deletes for existing publication and lifecycle transactions. Repository policy is resolved before attachment IO; agents never change that policy.

**Tech Stack:** Go, Cobra, existing Git transaction and GitHub CLI seams, portable temporary Git repositories for integration tests.

## Approval and scope

Approved by ops:0 on 2026-10-13 (receipt a0e515fc-0b1b-4ed6-8d69-c82163b90578),
including the limits below; implementation remains subject to change-code.
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
  A preflight refusal leaves body, files, index, and refs unchanged. IO failure
  after an effect reports its retained state and recovery action as specified below.
- Every refusal names the specific blocked input and the action that clears it:
  ask the operator to change the named policy file, remove an asset to fit a
  budget, reconcile an unlinked file, repair a missing link target, or rerun the
  named reconciliation verb. Tests assert the actionable remedy, not just failure.
- Existing published assets may still archive/restore if policy later tightens;
  these move existing evidence, they do not admit new evidence. Publishing new
  asset bytes rechecks current destination policy and ownership. Exact byte and
  bundle checks apply to binary-only diffs as well as body edits.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Policy / Decision (`ParsePolicy`, `ResolvePolicy`) | cmd/sdlc/internal/issueassets/policy.go | new |
| Bundle / Delta (`ParseRefs`, `BuildBundle`, `BundleDelta`) | cmd/sdlc/internal/issueassets/bundle.go | new |
| AttachState / AttachStage (`PrepareAttach`, `NextAttach`) | cmd/sdlc/internal/issueassets/attach.go | new |
| Publication conflict decision (`republishDecision`) | cmd/sdlc/republish.go | modified |
| Landing archive planner (`planLandingArchive`) | cmd/sdlc/landingarchive.go | modified |

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
| Archive proof (`confirmLandingArchive`) | cmd/sdlc/landingarchive.go | modified | Git snapshots, tracker state and clock |
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

ARCH-ORDER: `NextAttach(state,event)` is the pure transition authority. States
are Prepared, InstalledNew, InstalledExisting, Linked, RolledBack, and Retained;
effects are InstallNoReplace, CompareBodyAndLink, RemoveOwnedBytes, and Report.
The shell may not advance state directly. Prepared is constructed only after all
policy/ownership/path/budget checks and records expected body bytes, source digest,
exact destination, and whether matching destination bytes already existed.

| State + event | Next state and effect |
|---|---|
| Prepared + install succeeds newly | InstalledNew; compare expected body and link |
| Prepared + matching file already exists | InstalledExisting; compare expected body and link |
| Prepared + install fails | RolledBack; report IO failure, no destination deletion |
| InstalledNew + body link succeeds | Linked; finish |
| InstalledExisting + body link succeeds | Linked; finish |
| InstalledNew + body link fails | request removal only if exact installed bytes remain and current body does not reference them |
| InstalledExisting + body link fails | Retained; never delete pre-existing bytes |
| removal confirmed | RolledBack; report failure and rerun action |
| removal refused/failed/unconfirmed | Retained; report exact path and reconciliation action |
| Linked + retry | Linked; no duplicate link or byte mutation |

No journal is needed: a restarted request revalidates the current policy, owner,
body, source digest and destination. An otherwise-unlinked file is adoptable only
when its canonical content name and bytes exactly equal THIS request's validated
source; that is sufficient to finish this requested attachment, not authority to
delete it. Every other unlinked file refuses, naming the path to reconcile. After
restart, adopted bytes count as pre-existing. A changed body is never overwritten;
the caller reruns to prepare from its new contents. Ordinary rollback never
removes bytes another writer linked or modified. A failed body rename that cannot
be confirmed is reconciled by rereading exact body and destination bytes before
any cleanup. Tests inject these outcome events through the production transition.

Use atomic no-replace install and atomic body replacement under the repo lock.
Temporary names are bounded, same-directory staging; ordinary exits remove owned
temporaries. A hard-kill leftover has no persisted proof of ownership: report its
exact path for explicit removal before retry, never silently collect arbitrary
files. No background task survives return. Publish/archive reuse existing atomic
Git commits and whole-bundle retry proofs; a lost response is not failure evidence.

ARCH-FUNERAL: attach creates bytes, issue readers are last consumers, lifecycle
archive/restore moves the bundle. Removing a body link removes the current-tree
asset through the bundle writer at publication; retained Git history is intentional
and costs at most 1 MiB per distinct attachment (no automatic per-event capture).
The 8 MiB live issue budget is not a total-history bound: replacement accumulates
history; document that cost and refuse oversized writes. Temporary files end at
success/failure/reconciliation. No attachment ledger or second persistent index.

## Function-level test strategy

| Function | Strategy and mechanical guard |
|---|---|
| `ParsePolicy`, `ResolvePolicy` | Fuzz malformed config/host data and use stateful visibility responses; only typed explicit permission admits bytes and every refusal carries a remedy. |
| `ParseRefs` | Fuzz Markdown and path encodings with parser seeds; destinations must be actual links within the exact issue prefix, never code examples or traversals. |
| `BuildBundle`, `BundleDelta` | Property tests over bounded byte maps; output links resolve exactly, unreferenced members never publish, budgets and immutable names hold. |
| `NextAttach` | Reproducible generated event sequences and injected IO failures; no pre-existing/changed bytes deleted, no stale body overwritten, same-request retries converge. |
| `republishDecision` / bundle decision | Property tests over local/base/main snapshots; conflict detection covers byte-only change and retry equality covers the whole bundle. |
| `planLandingArchive`, `confirmLandingArchive` | Generated complete/incomplete bundle snapshots; only an exact archived generation proves completion. |
| `FilesSelected` (gitx) | Aggregate-volume regression with >32 MiB of unselected asset blobs; filtering must precede content reads and existing limits remain unchanged. |
| attach/publish/archive/restore production commands | Temporary bare remotes plus injected failures at transaction seams; snapshot files/index/refs before refusal and verify exact content after recovery. |

## Task 1 — Policy, bundle and attachment

Files: new issueassets core files and colocated tests; new `issueassets.go`,
`issueattach.go`, `assetpolicy.go` and tests; modify `issue.go`, `ghclient.go`,
`fetch_test.go` (shared fake), `main.go` only if needed for dynamic help.

- [x] Write failing tests for the named policy, parser, bundle and attach functions
  using the function-level strategies above; run and verify RED.
- [x] Implement typed pure policy, bundle parsing, path mapping and `NextAttach`.
- [x] Write failing production-command tests and implement thin IO/CLI adapters,
  including policy help through the same resolver and actionable refusals.
- [x] Run `go test ./cmd/sdlc/internal/issueassets -count=1` and
  `go test ./cmd/sdlc -run 'Test(Asset|IssueAttach)' -count=1`; expect PASS.
- [x] Commit and reconcile this plan's entity/file names with the implementation.

## Task 2 — Publication, recovery, ownership and discovery

Files: `republish.go`, `issuemovedetail.go`, `internal/tracker/detailpublish.go`,
`issuepublish.go`, `issuerestore.go`, `transferguard.go`, `transferverdict.go`,
`externalmerge.go`, `issuemigrate.go`, `internal/gitx/snapshot.go`;
existing colocated tests and focused bundle integration tests.

The reader sweep found four `TrunkView.Files` consumers: issue discovery in
`finishLandedLeftovers` and `migrationInventory` must select direct Markdown
children before reading blobs; `archiveIssueOnMain` reads plans only and tracker
`Reader` reads the card branch only, so neither ingests issue assets. Add
`FilesSelected(prefix, predicate)` to gitx and derive `Files` from it; filter the
metadata list before batch cat-file. Preserve 1 MiB/32 MiB reader limits.
Audit raw subtree readers too; issue discovery reads documents, asset operations
read the explicitly indexed bounded bundle, never a fleet-wide byte inventory.

- [x] Write RED tests for the publication/ownership strategies and aggregate reader
  strategy above; include legacy publication and first-publication recovery adapters.
- [x] Implement selected metadata-first reads and shared bundle publication read sets,
  exact writes/deletes, ownership rechecks and whole-bundle reconciliation.
- [x] Extend restore and resting-branch reconciliation to the same bundle generation;
  preserve untracked/concurrent user edits rather than overwriting them.
- [x] Run `go test ./cmd/sdlc -run 'Test(Asset|IssuePublish|MoveDetail|Republish|Transfer|IssueRestore|Migration|FinishLanded)' -count=1`
  and `go test ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/tracker -run 'Test.*(Detail|FilesSelected|Snapshot)' -count=1`; expect PASS.
- [x] Commit the publication/discovery integration.

## Task 3 — Lifecycle sweep

Files: `landingarchive.go`, `archivemain.go`, `abandon.go`, `push.go`, `merge.go`,
shared archive helpers as identified during implementation; `landingarchive_test.go`,
`archiveartifacts_test.go`, `abandon_test.go`, `migrate_test.go`, slot move tests.

- [x] Write RED tests using archive generation and transaction strategies above.
- [x] Extend pinned archive snapshots/proofs and exact writes/deletes with bundles.
  Abandon reads its final assets from the kept branch; restore/reopen uses the
  same generation. Legacy prepared moves include asset members in recovery.
- [x] Prove slug rename and slot move preserve links/bytes without a special copy
  path. Preserve cross-repo issue migrate refusal before any mutation.
- [x] Run `go test ./cmd/sdlc -run 'Test.*(Asset|Archive|Abandon|Restore|Move|Migrate)' -count=1`; expect PASS.
- [x] Commit the lifecycle integration.

## Task 4 — Documentation and acceptance

Files: `cmd/sdlc/helptext/issue.md`, relevant publish/migrate/recovery help,
`atlas/workflow/issue-assets.md` (new), `atlas/index.md`, `README.md`, issue Log.

- [x] Document operator-owned policy, visibility source and unknown behavior, limits,
  immutable plain-Git history cost, attach/publish workflow, private note summaries,
  no recording dependency, archive paths, retry behavior and migration exclusion.
- [x] Add a temporary-repository CLI test: text-only refuses, operator fixture policy
  permits attachment, publish carries bytes, archive resolves the unchanged link,
  restore/slot move retains evidence. Assert exact bytes, paths and refs.
- [x] Run focused tests, `git diff --check`, and review final diff against main.
- [x] Check `uptime`; only with load <40 start `make test` in background. Keep the
  checkout unchanged until the full suite completes; save exact results.
- [ ] Run `sdlc close --issue 337 --verified '<actual evidence>'`; its boundary review
  is the one required fresh-eyes review. Resolve findings and log measured results.
- [ ] Send the hub PR and real `git rev-parse HEAD` hash. On SHIP, merge fresh main,
  satisfy any revalidation/CI gates, run `sdlc merge --yes`, report and stop.

## Revisions

- 2026-10-13 — Initial implementation draft after repository exploration and hub
  scope direction. Cross-repo issue migration is excluded explicitly (#347);
  intra-repo lifecycle includes abandonment and legacy paths to avoid losing bytes.
- 2026-10-13T20:09Z — Hub approved the draft and limits; added the operator's
  requirement that every refusal name its clearing action. No scope change.
- 2026-10-13T20:12Z — Plan-quality PQ-1/PQ-2/PQ-3: added metadata-first issue
  discovery with aggregate-volume verification, explicit NextAttach transitions
  and ownership-aware rollback/restart, and named function-level adversarial test
  strategies in place of prose case inventories. These refine the approved scope.

- 2026-10-13T20:43Z — Implementation evidence refined byte preservation:
  ordinary Git staging/checkout can apply clean/smudge, encoding or newline
  conversion. Reuse one conservative attribute preflight at attachment and
  materialization boundaries; refuse with a path-specific `.gitattributes`
  remedy rather than create new local commit machinery. Exact publication
  writes stay unchanged. Abandon also checks newly admitted kept-branch assets
  before pushing the archive ref and again inside the main publication retry.

- 2026-10-13T20:55Z — Targeted verification used the combined Asset/IssueAttach
  and bounded-reader group plus all gitx/tracker package tests, then narrower
  lifecycle regression groups; all passed. The earlier broad command run timed
  out during transcript discovery under extreme machine load, so it is not
  acceptance evidence. The upcoming full module suite covers the remaining
  existing lifecycle cases. Use six shards to limit machine contention.

- 2026-10-13T21:57Z — Acceptance suite passed at 35eccf5153f31923d7f2c371ad1a8a618350d8f4 using PAIR_AGENT=claude make test ARGS=--shards 6: six shards, 1155 command tests, 47 other packages, zero failures, 431.2s. Started below load40 and acquired both machine suite slots. Close review and hub SHIP remain pending.

### 2026-10-13T22:06Z — close review correction plan

The first close returned REWORK. These corrections preserve the approved scope:

- BR-1 / ARCH-SECURE: resolve new-byte admission from the pinned destination
  policy on every transaction attempt, retaining local restrictions. Sweep
  republish, first publication and receipt recovery, legacy commit publication,
  transfer validation and abandonment. Regressions must fail when only main's
  policy tightens and when policy changes during retry.
- BR-2 / ARCH-PURPOSE: check both source and destination paths before Git
  materializes lifecycle asset bytes, then compare restored files with the
  pinned bundle. Cover archive-only conversion during production reopening and
  inspect all other materialization paths for the same class.
- BR-3 / ARCH-PURE: corrected the table: planLandingArchive is a pure planner;
  confirmLandingArchive is integration because it reads Git snapshots and time.
  The existing integration tests correctly use temporary repositories.
- BR-4 / ARCH-MOCK: add an opt-in read-only live visibility conformance test
  through RepoVisibility and document its cadence. Keep the deterministic
  stateful fake tests for ordinary test runs.

After RED/GREEN evidence for behavioral fixes, rerun the full suite with the
same load/lock conditions and return to the binary-owned close review.

- BR-4 conformance cadence: TestLiveAssetVisibilityConformance runs for this correction, after RepoVisibility argv/decoder changes or gh upgrades, and optionally in periodic credentials-enabled CI. Normal suites retain the stateful fake and skip the explicit opt-in live query; atlas/workflow/issue-assets.md documents the command.

- 2026-10-13T22:56Z — Corrected acceptance suite passed at ff3e3811b841091430163f44ad2812066ebf0746: 1169 command tests, 47 other packages, six shards, zero failures, 482.1s. Invoked below load40 with both slots reserved. Returning to close for disposition of BR-1 through BR-4.

### 2026-10-13T23:04Z — second close review correction plan

Round 2 accepted BR-1, BR-3 and BR-4, but retained BR-2 and added BR-5/BR-6.
The repeated materialization finding shows the previous caller inventory was
incomplete; fix these invariants across the full enumerated class:

- BR-2: every asset checkout/restore followed by a working-tree commit must
  verify pinned bytes before that commit. Enumerate issue restore, reopen,
  slot moves, resting fast-forward, restoreToHead/set-aside publication cleanup,
  first-publication removal, legacy archive preparation and landing cleanup.
  Git-only ExactBytes transactions remain distinct from materialization.
  Add a production issue-restore regression with a post-checkout tampering hook.
- BR-5 / ARCH-SECURE: authorization retains the complete destination host and
  owner/repository identity through the visibility seam. A conflicting GH_HOST
  must never select a different repository for attach, publication or recovery.
  Add explicit-host fake regressions and live conformance under conflicting
  ambient host settings.
- BR-6 / ARCH-PURPOSE: the body is authoritative even when no asset folder has
  ever existed. Publication validation derives requirements from changed bodies
  as well as asset paths. Reproduce a legacy body-only link to a missing canonical
  file, then sweep legacy/tracker publication, guards and lifecycle validators.

Keep the acceptance checkbox open, record RED/GREEN evidence and rerun the full
suite before returning to close. No approved feature-scope expansion.

- Round 2 implementation refinement: first-detail creation/import and persisted creation recovery also validate the body index before card/publication effects; they cannot carry asset links without corresponding bytes. Merge cleanup uses fetch plus pinned fast-forward, consistent with the resting-branch-only-fast-forward constitution, so preflight and byte verification refer to the actual materialized tree.

- Materialization caller inventory (round 2): issue restore (`issuerestore.go`),
  publication restore/finalization (`republish.go`), abandon/reopen
  (`abandon.go`), slot moves (`move.go`), resting fast-forward (`claim.go`),
  initial-detail handoff (`issuemovedetail.go`), planning/takeover
  (`planningbranch.go`, `handoff.go`), change-code branch/worktree creation
  (`branchcreate.go`), legacy merge cleanup (`mergematerialization.go`),
  durable landing return (`landing.go`), and tracker cutover/reconcile
  (`issuemigrate.go`). Archive staging checks cover both legacy push and merge.
  Successful retry shortcuts must verify retained bytes too. Git-only
  ExactBytes transactions do not check out assets; the isolated review checkout
  is not an issue lifecycle publication path. Cross-repository migration remains
  refused under the approved scope cut.

- Main integration at fa4a79b51f8ee5df0581da3f104d799bb9b22177 adds
  completed-issue reopen (`reopendone.go`) and staging/clearing (`stage.go`) to
  the caller inventory. The shared unarchive helper retains the asset bundle
  proof. Both new entry points have conversion/hook regressions; retained-branch
  reopen checks the switched tree before any optional merge. Ephemeral local
  check worktrees, like review worktrees, do not publish issue lifecycle state.

- 2026-10-13T23:54Z — Full acceptance after round-2 corrections and current-main
  integration passed at fbd5498222087c10d2c7b6f6b0822bcf4bc37e00: 1226 command
  tests, 48 other packages, six shards, zero failures, 818.8s. Start load 39.22;
  two-slot lock active. All required implementation verification is complete;
  binary-owned close review and hub SHIP remain pending.
