# Slots v2 workspace identity Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tasks or superpowers-executing-plans for session-warm work. Steps use checkbox syntax for tracking.

**Goal:** Give SDLC and Couch one read-only, Git-verified repository/workspace identity contract, including durable numbered slots.

**Architecture:** Promote existing fleet Git normalization and worktree parsing to `pkg/workspace`. Keep repository identity, artifact checkout, and workspace address distinct. SDLC uses the shared package; Couch can consume its versioned JSON CLI without importing SDLC internals.

**Tech Stack:** Go, Cobra, local Git plumbing, temporary real Git repositories.

**Status:** Proposed for operator approval. No implementation authorized by this document alone. This exceeds the 100-code-line quick shell; use the full flow at change-code. One atomic issue-close review boundary; task numbers below are not milestone tags.

## Design and alternatives

Reuse fleet normalization in a public package (recommended). It already resolves nested cwd, symlinks, the common Git directory and the primary without assuming `main` is checked out. A separate basename parser would duplicate that behavior and mistake unrelated directories for slots. A persisted slot registry would add a second authority and recovery lifecycle before provisioning needs one; it is unnecessary for the agreed fixed directory convention.

Git's canonical common directory identifies the repository. The first non-bare main-worktree entry identifies its primary checkout; its basename is the fleet repo name. Current worktree root identifies where local artifacts and code are read/written. Never replace a current checkout path with the primary just to repair a display name.

Primary has canonical address `repo:0` (input alias `repo`), resting branch `main`. A numbered slot must be a registered linked worktree of that same common directory at exactly `<fleet-root>/worktree/<repo>-slotN`, with positive canonical decimal N (no leading zero, sign, or integer overflow). Its address is `repo:N`, resting branch `main-slotN`, regardless of the active issue branch. Ordinary worktrees get kind `worktree`, no numbered address or inferred resting branch, and retain their current checkout behavior.

Slot validation requires its resting local ref to exist and resolve to a commit. It must not be checked out in another worktree. A slot currently on another reserved resting branch (`main` or a different `main-slotN`) is mismatched and errors. Commit equality with main/upstream is NOT required: independently stale baselines are valid. No remote fetch or tracking-policy requirement belongs to identity. Primary identity remains usable in ordinary master-only/unborn repositories; `main` describes the v2 convention, not a new gate for every historical repo.

An address resolves only an existing valid workspace. `:N` uses the caller's repo; qualified names use exact fleet primary names (no prefix matching for the new address grammar). Existing artifact-ref prefix behavior remains. Resolving a requested slot verifies both its expected canonical path and Git membership against the requested primary. A different repository at that path, missing directory/ref, duplicate canonical membership, bare primary, ambiguous membership, or inaccessible Git evidence is an explicit error. A slot-looking ordinary path outside the canonical location gets no slot ownership; an address request for it fails. A symlink spelling that resolves to the legitimate canonical location succeeds; a canonical slot path redirected outside it is refused.

The package exposes `NormalizeVantage(reader, dir)` for repository topology independent of slot readiness, `Resolve(reader, dir, address)` for validated workspace identity, and pure address/path/classification functions. This separation lets fleet enumerate a broken slot without hiding the entire repo and lets provisioning discover the primary before a slot exists. Couch uses `sdlc workspace [address] --json` (cwd as context); absent address resolves the current worktree. It can read primary/fleet identity and apply the documented path convention before provisioning, then validate the completed slot through the same command. No creation, number allocation, roles, claims, dependency binding, refresh, or landing in this issue.

Success JSON is one object with `schema_version: 1`, `repo`, `repo_identity` (canonical common dir), `primary_root`, `fleet_root`, `worktree_root`, `kind` (`primary|slot|worktree`), `address` (null for ordinary worktree), `slot` (0 for primary, positive for slot, null otherwise), `branch` (null when detached), `head`, and `resting_branch` (null for ordinary). JSON errors use nonzero exit and stderr, never a partial success object. Existing state JSON fields remain; add a `workspace` object with this schema and human-readable address/resting branch lines.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Worktree / ParseWorktrees | `pkg/workspace/worktree.go` | modified: promoted from gitx |
| Vantage | `pkg/workspace/paths.go` | modified: promoted from fleet |
| Address / ParseAddress / SlotPath | `pkg/workspace/address.go` | new |
| Identity / Classify | `pkg/workspace/identity.go` | new |

`Vantage` separates N worktrees from one common-dir identity and primary. `Classify` takes canonical observed topology, selected worktree, and observed resting refs; it never reads disk. Address parsing owns the closed input grammar. Ordinary worktree null fields avoid inventing a lifecycle. These entities eliminate path heuristics across SDLC (ARCH-DRY); future provisioning and branch operations consume the contract rather than extending it with roles.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| GitReader / NormalizeVantage / Resolve | `pkg/workspace/paths.go`, `resolve.go` | modified/new | injected local Git reader and canonical filesystem paths |
| workspace CLI / SDLC identity adapter | `cmd/sdlc/workspace.go`, `workspacepaths.go` | new | Cobra, existing execGitRunner |
| Fleet adapters | `cmd/sdlc/internal/fleet/gitpaths.go` | modified | shared topology package |
| Artifact, project and calibration consumers | files enumerated below | modified | shared identity and existing artifact IO |
| Workspace fake Git | `pkg/workspace/workspacetest/fake.go` | new | portable temp-directory state behind GitReader |

Keep `gitx.Worktree` and `gitx.ParseWorktrees` compatibility aliases/delegates so existing consumers continue compiling, but move all parser logic/tests to the public owner. Fleet's normalization delegates to the shared implementation; its prospective-path policy remains in fleet. Canonical path helpers used by both must delegate rather than fork. Existing fleet tests remain conformance coverage.

The workspace fake stores repositories keyed by common dir, registered worktrees, current branch/HEAD, and local refs; checkout/ref/membership changes alter subsequent reads. It supports only this resolver's read commands, rejects unsupported commands, and exposes fault injection for missing/inaccessible/malformed observations. Keep the richer existing fleet fake for fleet history/policy tests; do not copy its unrelated history/dirty-state model. Run the same workspace scenarios against the fake and temporary real Git repos on every package test run (ARCH-MOCK).

## Consumer contract and inventory

Two path meanings are deliberately separate:

- Local issue, plan, code and current-repo project operations use `WorktreeRoot`.
- Repo names, peer discovery and default brain location use `Repo`, `FleetRoot`, and `<FleetRoot>/brain`. Qualified refs naming the current repository resolve in the current worktree, like unqualified refs; other repos use their primary checkout. Logical issue deduplication uses common-dir identity plus issue ID.

Inventory to migrate and verify (test files live beside each production file):

| Surface | Production locations | Change |
|---|---|---|
| fleet topology and inventory | `internal/fleet/gitpaths.go`, `inventory.go`, `fleet.go` | delegate existing topology normalization, preserve ordinary and prunable-worktree handling |
| state and new CLI | `state.go`, `workspace.go`, `main.go` | expose identity, anchor default state issue/history reads at current root for nested cwd |
| resolve/open/issue deduplication | `resolve.go`, `open.go` | peer search from fleet root, canonical current repo alias, common-dir logical issue key, GitHub display repo |
| review/orientation | `reviewsidecar.go`, `orientation.go`, `judge.go`, `milestoneclose.go` | correct repo name, including manifest-derived labels, while retaining current worktree root |
| actual attribution | `actual.go`, `close.go:resolveActualRoots` | canonical name for DiscoverWindowIssues; keep transcript/session root at current workspace; default brain from fleet |
| ordinary branch-worktree placement | `branchcreate.go:createWorktreeBranch` | use fleet root and canonical repo name for `<fleet>/worktree/<repo>/<issue-branch>`; `.goto` stays at invoking checkout |
| close and project discovery | `close.go`, `projectfind.go`, `internal/project/discover.go` | fleet root and canonical repo label; substitute current checkout for its primary during project discovery, dedupe common-dir identity |
| project board/forecast | `projectforecast.go`, `projectstatus.go` | derive repo/fleet from project checkout, preserve current-worktree project content, use resolver for referenced issues |
| migration | `migrate.go` | repo labels and inbound fleet scan; reject migration between worktrees of the same repository |
| propagation/planning | `propagatebase.go`, `startplan.go` | fleet discovery and labels derive from identity; retain actual checked-out source as dependency origin |
| brain/calibration/transcripts | `estimatesource.go`, `startplan.go`, `actual.go`, `close.go`, `milestoneclose.go`, `project.go`, `projectstatus.go`, `projectsetstatus.go`, `projectthroughput.go`, `projectclose.go` | one default-brain resolver at command IO boundary; default is fleet sibling, explicit flags/env overrides retain their existing cwd-relative/absolute meaning |

Do not broadly chdir the process or rewrite all user-provided paths. Use Cobra flag Changed information to distinguish omitted `--brain-dir` from an explicit `../brain`; programmatic callers pass resolved paths. Default issue/project paths touched above are anchored to the current worktree; explicit path arguments retain existing meaning. Preserve existing best-effort warning behavior for optional calibration in planning, but do not substitute a guessed path on failed identity resolution. Required identity failures propagate before mutation. Existing non-Git pure helper tests should inject identity, not force production to accept unverified paths.

Retain `claim.go:findMainWorktree` as a branch-location query: a checkout currently on main is not necessarily the primary. Merge return/cleanup policy belongs to #246 and is not changed here. Keep repo transaction locks keyed to the common Git directory. Fleet inventory still reads primary policy/issues, with its repo-name argument explicitly documented as primary-derived. Ordinary branch-worktree placement is a path correction only; explicit source/refresh semantics remain #245.

## Architectural constraints

- ARCH-PURE: porcelain/address parsing and classification are deterministic; Git/FS probes stay in Resolve/NormalizeVantage. IO consumers receive resolved identities.
- ARCH-CONSTRAINTS: command/startup workload, local disk only, linear in registered worktrees and fleet siblings. No network, background tasks, recursive filesystem search, or process-global cache. Sequential reads, one topology load per resolver invocation; measure real fixtures with 1, 10, and 100 worktrees and record timing/query count, without claiming a latency SLA. These sizes are test samples, not product caps. Existing Git process cancellation/timeout conventions remain.
- ARCH-SECURE: Git output, addresses, refs and symlink paths are untrusted; validate closed grammars and canonical membership, use argv slices, reject malformed/unknown porcelain as today. Tests clear user Git config/hooks and operate only under temp dirs. No credentials.
- ARCH-ORDER: resolver holds no state between invocations because it is a read-only observation. A caller may change branches/remove a worktree during a read: recheck selected top-level/common-dir/HEAD at the end, fail on conflicting observations, do not retry invisibly. The result is a snapshot, never authority for a later mutation; future #245/#246 must revalidate under their operation lock. Ref occupancy is observational, not a reservation.
- ARCH-FUNERAL: creates nothing durable because resolution only reads Git/FS and emits output. Fixture directories are owned and removed by tests; no registry/cache/lock family added.
- ARCH-PURPOSE: finish the entire consumer inventory above and shadow-sweep remaining basename/parent assumptions. #243 handles dependency policy, #244 concurrency semantics, #245 branch/refresh, #246 landing; none is used to defer identity fixes.

## Chunk 1: shared resolver and production consumers

### Task 1: Promote canonical topology without changing fleet behavior

Files: create `pkg/workspace/worktree.go`, `paths.go` and colocated tests; modify `cmd/sdlc/internal/gitx/worktree.go`, `worktree_test.go`, `cmd/sdlc/internal/fleet/gitpaths.go` and its tests.

- [ ] Move parser and canonical topology tests to the shared owner, first asserting shared API behavior for nested cwd, symlinks, relative common-dir responses, detached/ordinary/bare/prunable records and whitespace-bearing paths.
- [ ] Run `go test ./pkg/workspace/... ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/gitx -count=1`; new API tests must fail before implementation.
- [ ] Promote the existing implementations, retain delegates for old imports, share canonical helpers, and rerun the same command to PASS.
- [ ] Commit explicit paths with `#242: refactor: share Git workspace topology` and model coauthor trailer.

### Task 2: Address and validated identity contract

Files: create `pkg/workspace/address.go`, `identity.go`, `resolve.go`, their tests, `workspacetest/fake.go`, and `conformance_test.go`.

- [ ] Write table tests for `repo`, `repo:0`, `:0`, `repo:1`, `:2`, invalid/negative/leading-zero/overflow slots, missing repo context, path traversal/separators, colon ambiguity, and repo names containing dots/hyphens. Repo token must be one exact safe basename (not `.`/`..`, no separators/control bytes/colon).
- [ ] Add fixtures with primary on an issue branch, slots 1/2 on different issue branches and different resting commits, and an ordinary feature worktree. Assert correct identity from root/nested/symlink cwd and address aliases. Confirm ordinary worktrees never become `:0`.
- [ ] Add negative tests for missing resting ref, resting branch occupied elsewhere, wrong reserved branch, wrong common-dir membership, canonical-path impostor, duplicate membership, redirected slot symlink, failed Git read, malformed porcelain and changed final observation. Assert refs, status and worktree lists unchanged on success and refusal.
- [ ] Run `go test ./pkg/workspace/... -count=1` and confirm behavioral failures; implement pure classification plus thin probes, then rerun to PASS.
- [ ] Execute identical supported scenarios against the stateful fake and real Git; measure 1/10/100-worktree fixtures. Commit explicit package files as `#242: feat: resolve durable workspace identity` with model trailer.

### Task 3: Expose CLI and state

Files: create `cmd/sdlc/workspace.go`, `workspacepaths.go`, tests, `helptext/workspace.md`; modify `main.go`, `state.go`, `state_test.go`, `helptext/state.md`.

- [ ] Write production command tests through `buildRoot()` for JSON schema, exact aliases, nested cwd, ordinary worktree null fields, explicit error/nonzero status, and additive state output. Verify state reads the slot's issue files, not the primary's.
- [ ] Run `go test ./cmd/sdlc -run 'TestWorkspace|TestState' -count=1`; confirm failures, wire shared resolver via existing execGitRunner, render prose/JSON and rerun to PASS.
- [ ] Ensure the read-only workspace command acquires no transaction lock and mutates no Git/filesystem state. Commit explicit paths as `#242: feat: expose workspace identity to Couch` with model trailer.

### Task 4: Migrate the consumer inventory

Files: every production surface in the inventory plus colocated tests; create `cmd/sdlc/workspace_consumers_test.go` for shared real-Git fixtures.

- [ ] Seed one temporary fleet with primary/slots containing deliberately different local issue/project content, a peer project referencing the canonical repo, and brain velocity/transcript fixtures. Add regressions through production resolve/open, project-find/forecast, close preparation, migration, propagation discovery, review/orientation, planning and calibration entry points. Do not run a live external judge or publish during tests.
- [ ] Assert local and qualified-current refs read the slot, peer refs read the peer primary, fleet project discovery includes local slot content once, calibration uses fleet brain, and explicit relative brain paths/overrides stay explicit. Verify actual-time qualified self refs and transcript root separately; manifest-based review labels are canonical. Same-repository migration across worktrees must refuse before edits. Test ordinary branch-worktree placement from primary, slot and ordinary feature cwd; `.goto` remains at the invoking checkout. Ordinary feature worktrees exercise the same repository identity fix without numbered lifecycle assumptions.
- [ ] Run `go test ./cmd/sdlc/... -count=1` to capture failures; migrate consumers to the shared adapter, preserving injected pure-test seams and existing optional warning behavior.
- [ ] Re-run that suite to PASS. Shadow-sweep `filepath.Base`, `filepath.Dir`, `../brain`, `--show-toplevel`, `--git-common-dir` and `worktree list` in production SDLC; document every retained identity-looking calculation as checkout-local, delegated, or unrelated. Add any missed production consumer test before changing it.
- [ ] Commit explicit touched paths as `#242: fix: use repository identity across workspace consumers` with model trailer.

### Task 5: Document, verify and close

Files: new `atlas/workflow/workspace-identity.md`; modify `atlas/index.md`, `atlas/workflow/sdlc-binary.md`, issue log/checkboxes. Project completion is updated by the close gate.

- [ ] Document JSON v1, Go seam, primary/slot/ordinary distinction, validation errors, snapshot limitations and provisioning integration examples. Update atlas links and CLI help contract tests.
- [ ] Run `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1`, `go vet ./pkg/workspace/... ./cmd/sdlc/...`, and `git diff --check`. Use only temp fixtures for command smoke tests; preserve unrelated #230/#240 edits.
- [ ] Read verification-before-completion skill, record actual command evidence and timing measurements, tick completed steps, and sync issue/design changes.
- [ ] Commit docs; run `sdlc close --issue 242 --verified '<actual evidence>'` for the one mandatory fresh-context boundary review. Resolve blocking findings and log the verdict; no separate duplicate boundary reviewer.
- [ ] Publish via `sdlc pr` then `sdlc merge` when the authorized workflow reaches shipping. Follow gate errors; never bypass unrelated dirty work or overwrite it.

## Revisions

### 2026-09-22 — initial engineering proposal

Derived from the v2 project and live SDLC consumer audit after claim/start-plan. Reuses fleet's existing topology authority; this is the first durable engineering design, pending operator approval. Estimate follows accepted plan-quality review.
