# Durable slot landing implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Use bounded subagents for GitHub evidence and remote archive, and keep orchestration/integration in the main session. Steps use checkboxes for tracking.

**Goal:** Land a reviewed issue from :0 or :N without refreshing its resting branch, deleting its workspace, or touching sibling checkouts.

**Architecture:** Route addressable primary/slot checkouts through the existing PR and merge commands using shared workspace identity. Confirm exact GitHub integration, archive through the existing conditional trunk publisher, then switch and remove only the proven issue ref. Recovery re-observes Git/GitHub with `merge --branch`; there is no journal.

**Tech Stack:** Go, Git, gh CLI, existing workspace resolver and gitx.TrunkFile; stateful GitHub fake with real local bare Git remotes.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| landingTarget / landingObservation | cmd/sdlc/landing.go | new |
| landingAction / nextLandingAction | cmd/sdlc/landing.go | new |
| landingPR | cmd/sdlc/ghlanding.go | new |
| landingArchivePlan | cmd/sdlc/landingarchive.go | new |

`landingTarget` identifies one checkout/resting ref/configured main destination; an observation holds current evidence, not authority cached across invocations. One target has one selected PR. `nextLandingAction` is the pure phase decision: refuse, merge, archive, return, delete, complete. All IO outcomes feed the next decision; no effect is authorized solely by an earlier success flag. `landingPR` is a validated structured external record, never a branch-name boolean. An archive plan maps a bounded set of source artifacts to history destinations plus done bodies; one PR owns one archive commit. Reuse existing issue parsing, status vocabulary and archive naming (ARCH-DRY/PURE). Future extension is another GitHub merge policy, not multi-repository transactions.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| resolveLandingTarget | cmd/sdlc/landing.go | new | workspace.Resolve and Git config/remote identity |
| structured PR observation / expected-head merge | cmd/sdlc/ghlanding.go | new | existing realGH implementation, gh JSON and merge |
| archiveLandingPR | cmd/sdlc/landingarchive.go | new | TrunkFile.UpdateMany and immutable Git tree/history reads |
| runDurableMerge / durable PR preparation | cmd/sdlc/landing.go, pr.go, merge.go | new/modified | existing gates, Git switch/ref deletion, confirmation |
| landing fake | cmd/sdlc/landing_test.go | new | stateful PR records and local bare Git remote |

Use existing gitRunner and GH adapters; add narrowly typed methods instead of raw command escape hatches. Keep legacy methods intact for ordinary/dependency consumers. Git integration tests use real disposable repositories; the GitHub fake creates actual merge, squash or rebased commits on a bare main and maintains exact PR head/base/integration metadata, queue/query failures and effect counters. Read-only conformance uses the real merged PR130 metadata (head remains available after branch deletion); `TestLandingGHLiveConformance` consumes an explicitly selected existing PR and performs no live mutation. Run it at #246 acceptance and whenever the typed adapter, fake protocol or supported gh version changes; offline CI skips it unless the repository/PR environment variables are supplied. It compares real same-repo identity, retained head/base OIDs and merged integration metadata with the fake contract; merge argv is checked against installed gh help without a live write. Existing workspace fakes retain identity conformance.

## Execution and safety contract

- User approved the spec and dependency-first sequence: land Ariadne's independent clone normally, verify Pair against the merged dependency, then land Pair while preserving siblings.
- One atomic delivery, plain tasks and one close review; use full flow explicitly because expected runtime changes exceed the quick shell. No estimate before plan-quality accepts.
- :0 and :N preserve rest. Ordinary feature worktrees and private dependency clones retain legacy cleanup. Resolve before any legacy primary lookup or mutation; failures do not silently fall back.
- Named remote tracking main is mandatory for durable operations. Bind remote fetch/push and GitHub repository identity; reject missing/ambiguous config, local-dot and mismatched/fork PR identity. Reuse remote URL parsing where suitable; unsupported URL forms fail with guidance.
- PR creation pushes to that configured remote and uses the fetched target main for the branch window. It never creates a PR for a resting ref. Existing ordinary/dependency PR path remains compatible.
- Before initial merge, tracked dirt (including tracker edits) and active Git operations refuse. Untracked/ignored files are preserved; collision-protected switching may refuse after integration and is recoverable. Never stash/reset/auto-commit local work.
- Require local selected head = fresh remote issue head = PR head, normal publish/instance/duplicate-ID gates, and an expected-head merge request. Do not invoke gh local branch deletion. MERGED + integration commit reachable from freshly fetched configured main is the only integration confirmation, including queue/unknown outcomes.
- Keep current merge-commit submission policy. Already merged squash/rebase PRs use the exact original PR head plus reachable integration evidence, not original-head ancestry.
- PR-owned archives select codecomplete records with close anchors in `PR head --not PR base` ancestry. Read records at the pinned head; current resting files are irrelevant. Independently published issue copies must not hide an owning close, and unrelated codecomplete base records must not be swept.
- Archive fresh remote bodies and matching plan/review artifacts in one UpdateMany write/delete commit. Validate ordinary blob modes and safe repository-relative configured roots. Refuse missing/reopened/conflicting identities and occupied destinations. A concurrent remote edit is re-read on retry; no stale overwrite.
- Archive commit trailers bind GitHub repository + PR number + integrated head. Its diff supplies the complete expected move set. Confirm reachable provenance, complete moves/done generation, absent active paths and intact history blobs on retry. Wrong/missing provenance, partial completion, reopened records or later conflicting edits refuse. No generic file-existence success.
- After archive confirmation, revalidate identity, issue/rest heads, upstream and occupancy. Switch using `-c submodule.recurse=false switch --no-overwrite-ignore REST`; confirm rest SHA unchanged. Recheck issue ref equals integrated head and is unoccupied, then `update-ref -d refs/heads/ISSUE EXPECTED_HEAD` and remove only that branch's config. Never select main/main-slotN for deletion. No remote branch deletion added.
- `merge --branch NAME` is recovery from current issue or rest only. While at rest it can resume a merged PR, never merge an open one or apply gates to resting HEAD. Missing issue ref requires unique integrated PR and proven archive completion. Print recovery invocation before irreversible effects.
- Existing common-dir SDLC lock covers the transaction; plain external Git/editor writers remain outside it. Rechecks and expected-SHA mutations are bounded protection, not an atomic cross-process checkout lock.

### State/event decisions

| Observation/event | Next effect or result |
|---|---|
| Open matching PR, issue checked out, gates passed | expected-head server merge |
| Request failed or response queued/open | stop with recovery; next invocation queries before submitting |
| Exact merged PR, reachable integration, local issue unchanged | confirm/build remote archive |
| Archive push unknown or CAS retry exhausted | preserve all local state; next run confirms provenance |
| Complete matching archive, issue checked out | revalidate then safe switch to captured rest |
| At rest, same issue head, unoccupied ref | compare-and-delete ref |
| Ref already absent, matching integration/archive | remove any remaining selected branch configuration, then complete |
| Changed head/config/topology, dirt, ambiguous evidence | refuse with preserved state and explicit recovery |

Numeric bounds are conservative implementation limits, not capacity promises: at most 100 matching PR records, 10,000 commits inspected for close/archive provenance and 10,000 selected tree entries; command failures or exceeded limits refuse. gh calls use a two-minute context deadline; no background work outlives the invocation. Tests use tiny fixtures plus limit+1 cases. Existing trunk CAS retry budget remains authoritative. No durable new files; temporary index/blob files retain TrunkFile cleanup, archives use existing history lifecycle (ARCH-CONSTRAINTS/ORDER/SECURE/FUNERAL/MOCK).

## Shared rules and executable test strategies

Extract `archiveDestination(historyDir, kind, basename)` for existing `archiveDoneIssues`, `archiveDoneIssuesInDir`, `archivePlanArtifacts` and the landing plan, delegating to `vocab.ArchiveSubdir`. Extract `planArtifactBelongsToIssue(issueBase, artifactBase)` from the current ID-prefix rule and use it in both the filesystem plan mover and immutable tree selector. Extract `publishedIssueContent(frontmatter, body, date)` from `publishCodecompleteIssues` and use it there and in the remote planner. Existing terminal selection continues through vocabulary-backed helpers; PR-ancestry ownership is an additional selection predicate, not another lifecycle definition. Existing filesystem IO loops remain intact.

| Function under test | Adversarial strategy and mechanical guard |
|---|---|
| `parseLandingPRs` / `realGH.LandingPRs` | Fuzz malformed external JSON/identity/OIDs; bounded fake gh process must distinguish absence from failed/incomplete evidence. |
| `realGH.LandingMerge` | Capture subprocess argv and cancellation; require expected-head binding and no local cleanup capability. |
| `archiveDestination`, `planArtifactBelongsToIssue`, `publishedIssueContent` | Table/property tests over issue families and parsed bodies; both legacy and remote consumers derive the same names and lifecycle transform. |
| `selectLandingIssues` | Real Git divergent ancestry and independently published records; only immutable PR-owned close anchors select artifacts. |
| `planLandingArchive` | Pure snapshots with conflicting generations/paths/modes; complete atomic move set or refusal, never partial overwrite. |
| `landingArchiveComplete` / `archiveLandingPR` | Stateful remote changes and interrupted publication; exact reachable provenance and intact complete generation required before local cleanup. |
| `nextLandingAction` | Exhaustive phase/evidence combinations; no destructive action from uncertain, conflicting or merely queued evidence. |
| `resolveLandingTarget` | Real primary/slot/dependency topologies and malformed configuration; exact workspace identity binds named remote/main and GH repository without fallback. |
| `returnLandingToRest` / `deleteLandingBranch` | Deterministic ref/occupancy/file races at effect boundaries; snapshots and expected-SHA deletion preserve all unrelated state. |
| `runDurableMerge` / `runPR` | Stateful GH integration plus real bare Git, deterministic interruption hooks; assert retry convergence, gate enforcement and parent/dependency isolation through production entry points. |


## Chunk 1: implement the approved flow

### Task 1 — Structured GitHub evidence

Files: create `cmd/sdlc/ghlanding.go`, `cmd/sdlc/ghlanding_test.go`; extend the realGH seam without changing ordinary legacy behavior.

- [x] TDD the typed GH adapter and parser using the named strategy below; keep the legacy interface compatible.
- [x] Verify `go test ./cmd/sdlc -run 'TestLandingGH' -count=1` red then green, and commit explicit paths.

### Task 2 — Remote archive and retry proof

Files: create `cmd/sdlc/landingarchive.go`, `cmd/sdlc/landingarchive_test.go`; reuse/refactor the explicitly shared pure archive rules below. Read `construct/vocabulary/issue.cue` before lifecycle edits.

- [x] TDD the shared archive rules, owned selection, pure archive plan and publication/proof adapters using the named strategies below.
- [x] Implement `archiveLandingPR(root, remote, repo string, pr landingPR, issuesDir, plansDir, historyDir string) error` and read-only `landingArchiveComplete(root, remoteMainOID, repo string, pr landingPR, issuesDir, plansDir, historyDir string) (bool, error)` through existing TrunkFile; no checkout or new publisher. An empty owned set needs no archive commit.
- [x] Verify `go test ./cmd/sdlc -run 'TestLandingArchive|TestArchive' -count=1` red then green and commit explicit paths.

### Task 3 — Route PR/merge and safe cleanup

Files: create `cmd/sdlc/landing.go`, `cmd/sdlc/landing_test.go`; modify `pr.go`, `merge.go`; adapt existing PR/merge fixtures to explicitly represent legacy ordinary/dependency topology.

- [x] TDD the named phase/target/return/delete functions and routed commands below, with a stateful GH fake backed by a real bare Git remote.
- [x] Wire the durable path before legacy main lookup; preserve existing gates, confirmation and read-only dry-run. Add `merge --branch` recovery through the existing command.
- [x] Verify `go test ./cmd/sdlc -run 'TestLanding|TestMerge|TestPR|TestArchive' -count=1` red then green and commit explicit paths.

### Task 4 — Documentation and acceptance

Files: `README.md`, `atlas/workflow/workspace-branching.md`, `atlas/workflow/sdlc-binary.md`, `cmd/sdlc/helptext/merge.md`, `cmd/sdlc/helptext/pr.md`, issue246 and Pair's project record. Keep docs concise and link existing concepts.

- [x] Document :0/:N no-refresh landing, explicit branch recovery, remotely archived records vs an intentionally old local baseline, and Ariadne-first/Pair-second example. Explain dependency clone legacy behavior and no recursive publication/cleanup.
- [x] Run `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1 -timeout=15m -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'` (existing #210 missing historical fixture), `go vet ./cmd/sdlc/...`, and `git diff --check`. Validate affected help and build the SDLC binary used to close.
- [x] Mutation-check the important guards: wrong-head deletion, missing archive proof and implicit resting refresh must make their fixtures fail; restore and rerun affected cases.
- [ ] Update issue/project/atlas evidence and close through the one binary-owned boundary review. Fix findings before the close commit, open the PR, report readiness; merge only on operator instruction.

## Revisions

### 2026-09-23 — Approved spec to executable plan

The operator approved the reviewed design and confirmed dependency-first landing. This plan refines the approved behavior into bounded components and fixtures; it does not add another approval checkpoint or a new workflow framework.

### 2026-09-23 — Recover interrupted branch configuration cleanup

Fresh plan review identified the interruption between compare-and-delete of the issue ref and removal of its branch configuration. Proven absent-ref recovery must remove remaining configuration for that selected branch before completion. Test this exact interruption; never remove configuration for a recreated or occupied ref.

### 2026-09-23 — Plan-quality refinement

Addressed the three reported finding classes: replace test inventories with named risky-function strategies; define shared archive naming, membership and lifecycle helpers and their consumers; require read-only live GH conformance at acceptance and adapter/fake/version changes. No behavior or scope expansion.

### 2026-09-23 — Integration gate reuse

Read-only integration review found that legacy duplicate checking assumes origin and publish candidate enumeration uses issue-body diffs. Durable landing instead passes its pinned configured main to duplicate checking and its PR-owned close records to the shared publish rules. Effective fetch/push URLs bind GitHub identity; PR preparation pins and revalidates issue HEAD. These enforce the approved contract rather than widening scope. Corrected advisory PQ-1 architecture marker to ARCH-ORDER.

### 2026-09-23 — Preserve work created on rest during cleanup

Boundary finding BR-2 separates pre-integration/pre-switch cleanliness from
post-return ref cleanup. `landingCheckoutReady` still rejects Git operations in
every phase, but permits staged/unstaged/untracked resting work for an already
integrated PR. The three cleanup interruption fixtures assert unchanged files,
index bytes and resting SHA. No cleanup effect writes the resting files/index.
