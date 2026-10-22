---
collaboration-mode: discussion
---

# New repository tracker initialization plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tests and superpowers-executing-plans for the warm-context orchestration. Steps use checkbox syntax for tracking.

**Goal:** Make tracker initialization the default documented final step of new non-brain repository setup, while existing legacy repositories keep their behavior and receive migration guidance.

**Architecture:** Add explicit `sdlc issue init` as a narrow empty-inventory entry to the existing migration/bootstrap owner. Reuse its remote CAS, exact-root verification and cutover marker; finish by fast-forwarding the clean main checkout. Do not infer newness during ordinary weave compilation or publish from CI/refresh.

**Tech Stack:** Go/Cobra, existing tracker migration and Git trunk APIs, real local Git/bare remotes, existing command fixtures.

**Status:** Proposed; TL must approve the explicit setup step and opt-out boundary before change-code. This is planned as full flow because it adds an initialization/publication action; actual size and estimate follow the gate.

## Design decision

There is no distinct repository creator in this tree. The current setup is `weave link` then `weave compile`; bootstrap.sh merely invokes compile, and refresh shares its implementation. Compile supports non-Git roots and CI; refresh explicitly preserves leaf HEAD/index/branch/origin-main. An unconditional hook there would unexpectedly publish and could auto-migrate legacy repos. Missing scaffold or a first managed-index write cannot prove a new repo, and loses eligibility after partial setup.

Recommend an explicit `sdlc issue init` in the canonical new-repository setup sequence, after committing scaffolding and configuring/pushing main. Existing adoption/restore instructions remain distinct. Opt-out is choosing the existing adoption/compile path without issue init (or declining issue scaffolding); no new preference format or pending-setup marker. A brain marker always suppresses initialization. This is an explicit-init interpretation of the issue's allowed equivalent init step; seek TL confirmation before implementation.

Alternatives rejected: auto-init during compile/refresh conflates composition with remote publication and breaks existing contracts. Auto-init on first issue new changes the legacy verb's authority and cannot distinguish an intentionally legacy repo. A new top-level repository creation orchestrator duplicates setup and expands this issue unnecessarily.

## Core concepts

### Pure entities

| Name | Lives in | Status |
| --- | --- | --- |
| EmptyInitEligible | cmd/sdlc/internal/tracker/initialization.go | new |
| InitState / InitEvent / InitEffect / InitTransition | cmd/sdlc/internal/tracker/initialization.go | new |
| LegacySetupHint | cmd/sdlc/internal/tracker/initialization.go | new |
| MigrationInput / MigrationManifest / PlanTrackerMigration | cmd/sdlc/internal/tracker/migration.go | reused unchanged |

Eligibility accepts only an empty migration input: no active/archived details, branch issue edits, or dirty issue paths and no migration refusals. It cannot bless a nonempty repository by checking only card count. The input belongs to one pinned repository/main inventory; callers supply already observed facts. Unit tests require no IO. Existing completed tracker verification precedes this check, so rerunning initialization after later issue creation is a no-op when local main already carries the matching cutover.

### Integration points

| Name | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| issue init command / runIssueInit | cmd/sdlc/issueinit.go | new | explicit setup action and checkout completion |
| issue subcommand registration | cmd/sdlc/issue.go | modified | Cobra command surface |
| Existing migration/bootstrap owner | cmd/sdlc/issuemigrate.go | reused or minimally shared | inventory, tracker bootstrap, main cutover |
| Existing tracker/trunk publication | cmd/sdlc/internal/tracker, cmd/sdlc/internal/gitx | reused unchanged | Git remote CAS and cutover identity |
| Legacy migration hint | cmd/sdlc/legacymode.go | modified | one shared hint writer, no mutation |
| Legacy dispatches | cmd/sdlc/issue.go, cmd/sdlc/startplan.go | modified | hint at actual legacy entry |

The shell owns all Git/fs IO. Use existing local bare-remote fixtures/stateful repositories; no GitHub credentials, network service or function-call-only fake needed. Git is exercised through existing Git seams against real temporary repositories. Existing repository transaction lock spans initialization; no new lock or daemon. No weave tool lookup or compile hook is introduced.

## Contract and ordering

1. Resolve the explicit calling repository. Use canonical gitx.IsBrainRepo before any tracker mutation; brain setup reports skipped and creates no tracker. Missing/declined workshop/issues similarly skips rather than creating unwanted scaffolding. Ordinary compile/refresh never invoke this command implicitly.
2. Require the repository's clean resting main checkout, committed HEAD, configured publication target and reachable pushed main. Non-Git, unborn, no remote, unpushed main, dirty or non-main branch refuse before writes with the specific prerequisite to fix. Do not initialize Git, create a remote, commit source files or change branch automatically.
3. Open the existing unguarded migrate environment. If main already carries a valid matching tracker root, validate existing state; return if local checkout already incorporates that marker, otherwise finish the safe fast-forward below. A malformed marker, unreachable remote or mismatched tracker root is an error, never evidence of legacy absence.
4. Inventory through migrationInventory and apply the pure empty-init predicate. Existing issues/history/branch work means refuse with `sdlc issue migrate`; do not bypass migration's reviewed digest path for nonempty repositories.
5. Derive the ordinary empty MigrationManifest; call applyTrackerBootstrap and applyMainCutover, sharing their expected-absence CAS and exact-root checks. No new manifest version, root convention, branch name or alternative card authority.
6. Obtain the confirmed main cutover commit, verify local main remains clean and is its ancestor, then `git merge --ff-only` that exact commit. Never reset, stash or overwrite edits. Failure retains both published phases and names the safe explicit retry. Init success means the calling checkout can immediately use tracker-backed issue new and start-plan.

An offline result cannot report successful initialization. The existing initial commit and Git registration stay owned by setup, not this command. Unknown push acknowledgment preserves known partial state; retry observes root/marker identity rather than inventing completion. Concurrent initializers serialize locally; cross-clone races retain the existing remote CAS. A concurrent later issue write must not be overwritten or make init re-create the tracker.

## States, actions and lifecycle

Actions: complete Git prerequisites; rerun issue init; use reviewed issue migrate for nonempty legacy inventory; opt out by continuing compile/adoption without init.

| Observation / interruption window | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Brain or declined issue scaffold | initialization skipped | continue setup without tracker | none |
| Missing Git/main/remote or dirty/non-main caller | prerequisite refusal, unchanged repository | complete named prerequisite, rerun init | none |
| Nonempty legacy inventory | needs reviewed migration | issue migrate | none |
| Valid initialized tracker and current checkout | already initialized | continue setup | none |
| Before tracker CAS | not initialized or outcome unconfirmed | rerun init | none; original evidence retained |
| Tracker published, main marker absent | initialization incomplete | rerun init, verify same empty root | none; remote branch retained |
| Main marker published, local checkout stale | initialized remotely; checkout completion pending | rerun init after clean main | none; no forced checkout mutation |
| Marker/remote root inconsistent | evidence refusal | restore/correct named authority, rerun | none; no deletion or inferred success |

The pure eligibility/phase decision declares permitted effects; the existing migration/trunk owner executes publication. Test every interruption boundary through observed state changes, including a lost push response and concurrent initialization; failed reads never become absence. Init introduces no new durable artifact family: existing tracker root, main marker, temporary Git index/trunk stages and recovery evidence retain their current lifetimes. Opt-out has no new persisted config. Tracker history is existing authoritative issue history, not an init-specific growing ledger.

## Operating envelope and architectural review

This is an explicit batch setup command, never per keystroke or ordinary compile. Reuse bounded Git calls/output limits and caller cancellation. One pinned main inventory plus existing branch/worktree migration scan; no extra scan per issue. Initial repo has zero issue files. Count 8/16 unrelated files/branches to verify init does not spawn a new process per file; preserve existing migration limits and refusals. Network latency is the existing trunk/publication budget; an unreachable origin refuses visibly. No new CPU, memory or concurrency worker pool.

ARCH-DRY/PURE: share migration publication and make only the stricter empty eligibility pure. ARCH-PURPOSE: real setup fixture must proceed from init through issue new and start-plan. ARCH-MOCK: real portable Git/bare remote models actual state transitions and failures. ARCH-CONSTRAINTS: explicit bounded batch action; no compile hot-path cost. ARCH-SECURE: no credentials created/logged; all remote/marker/root reads remain typed and validated, and brains/opt-out are checked before writes. ARCH-ORDER: retain partial publication, test before/after both CAS boundaries and checkout completion; no rollback or inferred absence. ARCH-FUNERAL: reuse existing artifact owners/lifetimes. ARCH-EVOLVE: no existing format change; reused cutover version/root schema and unknown-field preservation stay intact.

## Enforced transition owner

`tracker.InitTransition(state InitState, event InitEvent) (InitState, InitEffect, error)` is the only phase mutation. InitState's fields and phase enum are package-private; callers construct it through NewInitialization and cannot advance it directly. Events/effects are tagged types. The returned state records confirmed root/commit identity, not just a phase boolean. The shell may only execute returned effects and feed their outcomes back; it never chooses a publication from raw marker observations.

| State | Event | Next state / returned effect |
| --- | --- | --- |
| new | Begin | inspecting / InspectInitialization |
| inspecting | Inspected(skip or prerequisite/refusal/unknown) | terminal skipped/refused; no mutation |
| inspecting | Inspected(confirmed complete/current) | done; no mutation |
| inspecting | Inspected(verified empty/uninitialized) | bootstrapping / BootstrapTracker |
| inspecting | Inspected(matching tracker-only root) | publishing / PublishCutover(root, expected main) |
| inspecting | Inspected(matching remote cutover, stale clean checkout) | completing / FastForwardCheckout(commit) |
| bootstrapping | BootstrapConfirmed(root) | publishing / PublishCutover(root, expected main) |
| publishing | CutoverConfirmed(commit) | completing / FastForwardCheckout(commit) |
| completing | CheckoutConfirmed | done; no effect |
| any active phase | effect failed, canceled or unconfirmed | terminal retry/refused, preserve confirmed identities; no next effect |
| wrong phase or terminal | duplicate/out-of-order event | reject without effect |

The in-memory machine is one invocation; retries reconstruct observations from existing durable root/marker/checkout authority. No new on-disk state is introduced. The IO adapter `initializationEffects` owns Inspect/Bootstrap/Publish/FastForward; its methods wrap existing migration/Git operations. A stateful fake and the real bare-Git adapter use the same effect dispatcher, providing a deterministic barrier before and after each external effect. Lost acknowledgments stop this invocation; a new invocation observes whether publication occurred. Cross-clone CAS races and later card creation are sequenced at these barriers. As with migration, setup requires other legacy SDLC writers to remain quiescent until initialization completes; the command does not claim to fence uncooperative old binaries or ordinary editors in another clone. Pinned-main CAS, exact-root checks, clean-checkout revalidation and the existing local repo lock remain the mechanical guards.

## Advisory legacy inventory

`legacySetupAdvice` resolves configured issue/history roots and uses `issue.LocalPathsByID`; it does not call the migration scanner, fetch or rescan remote state. `tracker.LegacySetupHint` chooses text from known-local-empty, local-nonempty or unknown observations. Empty-local text says issue init and qualifies the evidence: init verifies the complete inventory before publication. Nonempty text says migrate. Unknown reads are never called empty, and missing/declined scaffolds and brains receive no advice. The helper runs before legacy issue creation, and in runStartPlan beside its existing repositoryTracked result so bare start-plan is included. Advice changes neither legacy effects nor issue-new stdout.

## Tasks and verification strategies

### 1. Implement initialization through the pure model

Files: new cmd/sdlc/issueinit.go, cmd/sdlc/issueinit_test.go, cmd/sdlc/internal/tracker/initialization.go and initialization_test.go; register in cmd/sdlc/issue.go; share only necessary migration helpers from cmd/sdlc/issuemigrate.go.

- [x] Write pure decision/sequence tests and actual command regression RED, then implement the model and thin effect adapter.
- [x] Run the production init -> issue new -> claim/start-plan fixture against a disposable bare remote; commit after GREEN.

### 2. Wire legacy advice, recovery contract and default setup procedure

Files: cmd/sdlc/legacymode.go, cmd/sdlc/issue.go, cmd/sdlc/startplan.go and their tests; cmd/sdlc/internal/recovery/catalog.go (existing recovery owner), command help; README.md and atlas/workflow/{setup-and-replication,sdlc-binary}.md. Sweep operator/TL-facing setup guidance as required by the approval.

- [x] Write legacy advice regressions RED, implement one shared advisory helper and the pure text decision, and pass tests.
- [x] Add issue init to the recovery catalog, with verified retry semantics, help and proof references; run TestRecoveryContractsAreProven.
- [x] Make every new-repository/bootstrap creation procedure end with issue init after pushed initial main; preserve adoption/restore/refresh semantics and document opt-out/brain handling.

### 3. Acceptance and landing (one whole-issue boundary)

- [ ] Run private production CLI conformance against a disposable bare remote (TL approved, no stage), then the recorded pushed-head suite and close review.
- [ ] On SHIP and completed branch pass enter sdlc merge --yes, retain its integration run, verify merge/card/archive and report to hub. TL owns project updates and natural fleet follow-through.

| Risky function | Adversarial strategy / mechanical guard |
| --- | --- |
| EmptyInitEligible | Table over arbitrary MigrationInput contents and refusals; any nonempty or unknown input cannot authorize init. Pure predicate, no IO. |
| InitTransition | Generated event sequences with duplicate/out-of-order/cancel/unconfirmed events; opaque state and returned effects prevent skipped publication phases and further effects after uncertainty. |
| InspectInitialization | Contradictory repo/marker/root observations in a stateful backend; typed observed disposition and existing cutover parser/root verification reject fabricated empty or complete state. |
| runIssueInit / initializationEffects | Real bare-Git fixture plus controlled before/after-effect failures and concurrent clones; compare refs, card blobs, index/HEAD and working bytes across retries, with remote CAS and ff-only as guards. |
| LegacySetupHint / legacySetupAdvice | Local discovery fakes/temporary roots with empty, nonempty and unreadable observations; conservative text policy, no remote calls, pre-creation timing and unchanged stdout/effects are asserted through both production verbs. |
| Initialization inventory cost | Count Git calls at 8/16 unrelated files/branches through the existing runner; retain linear branch scan and fixed metadata reads, reject new per-file processes. |
| Recovery contract / command help | TestRecoveryContractsAreProven derives the mutating command obligation and validates proof names plus generated help. |

Targeted command: `go test ./cmd/sdlc/internal/tracker ./cmd/sdlc -run 'Test(IssueInit|EmptyTrackerInit|Initialization|LegacySetupHint|RecoveryContractsAreProven)' -count=1 -timeout 10m`. Expected GREEN; each behavior change must first produce its intended RED. The private production-CLI scenario is the full new-repository procedure, followed by the first issue and planning branch, rather than a mocked bootstrap-only call. All individual cases belong in executable tests, not this plan.

## Revisions

### 2026-10-21 — TL-approved default setup and inventory-sensitive hints

TL approved the explicit command (Couch 87ebe800-5c28-4e19-b11b-5bf5c2c1453a), full flow, and private production CLI plus real bare-Git conformance with no stage. Ordinary weave link/compile/refresh must not publish. To make initialization the default, every documented new-repo/bootstrap procedure in README and atlas, including operator/TL-facing creation guidance discovered in the implementation sweep, ends with issue init. The legacy hint is inventory-sensitive: empty inventory says issue init, nonempty says issue migrate. Brain/declined scaffold skip unchanged. This replaces Task 2's unconditional migrate wording; tests must prove both hints without changing legacy effects or the issue-new stdout path contract. Reuse one empty-inventory policy for initialization and hint selection; hint evidence is advisory, never publication authority, and unreadable/unknown inventory must not be claimed empty. The first hint observes pre-creation inventory (before legacy issue new creates its first details file).

### 2026-10-21 — plan-quality round 1 correction

PQ-1: named the opaque InitState/InitEvent/InitEffect model and InitTransition owner, explicit state/event/effect table, stateful effect dispatcher and deterministic publication boundaries. This formalizes the approved phases without changing authority or adding persisted state. PQ-2: replaced procedural case inventories with one named function/adversarial strategy/mechanical guard row per risky function. PQ-3 (Minor): included the required mutating-command recovery catalog/help/proof integration and TestRecoveryContractsAreProven. Advisory hint discovery uses the existing local-only scanner and labels local-empty evidence; authoritative init retains full migration inventory. No code changes before the gate passes.

### 2026-10-21 — implementation checkpoint after approval and change-code

The proposed status above is historical: TL approved the design and the full-flow gate passed after clean plan-quality rounds 2/3 and estimate acceptance. Implementation and targeted/private-CLI acceptance are complete; the recorded suite and whole-issue review remain. The approved procedure sweep also found tl-lead team-new help and its printed operator next steps; both now give base adoption/compile, committed pushed main, then explicit init, with a regression over their order. No new publication authority or mechanism was added.

### 2026-10-21 — BR-1 authoritative remote inventory correction (pending TL)

Close round 1 reproduced a remote-only legacy issue branch created after the initializer's last fetch: migrationBranchRefs lists cached remote-tracking refs, while snapshots refresh only main/tracker. This is a real shared-inventory defect; no publication change or new authority is intended. ARCH-PURPOSE/SECURE/ORDER require fresh evidence before the empty predicate.

Proposed bounded correction: the shared migration inventory refreshes all publication-remote heads with one explicit heads-to-remote-tracking fetch (no tags, submodule recursion or FETCH_HEAD write; prune only that remote heads mapping). A failed observation refuses. Enumerate local and refreshed remote branch names and full commit OIDs together once; all issue/anchor reads use those pinned OIDs while diagnostics retain branch names. Require refreshed main to match the already pinned main snapshot, otherwise refuse with rerun guidance. This preserves the existing quiescent-writer setup condition; it does not claim distributed fencing against writers after the snapshot. No new refs namespace, state files or cleanup owner: existing remote-tracking refs remain Git's cache.

Test strategies: two-clone real Git cases for branches newly created or advanced since the last fetch; empty tracking configuration must not suppress observation; fetch/remote failure must refuse before tracker/main publication; main movement between snapshot and inventory must refuse; branch movement after enumeration must leave reads pinned. Exercise the shared boundary via both init and migration. Repeat counted N/2N inventory guard, private production CLI reproduction, targeted init/migration tests, recorded suite and close review. The old queued suite was explicitly canceled and is not passing evidence. Await TL before production edits.

### 2026-10-21 — BR-1 correction approved and verified

TL approved the preceding revision (Couch 258bb7aa-6771-4df9-afb1-9784ce0a3803). The shared owner now does the explicit all-heads fetch, pins names/OIDs once for issue and anchor consumers, and rejects changed main. Explicit no-prune-tags and empty refmap prevent configured tag-pruning/ref mappings from expanding the approved cache write set. Before-fix real-Git tests failed in all six init/migrate remote-create/advance/narrow-fetch cases, plus failed observation and moved main. After correction those passed (11.666s), and the broad init/migration/transition/recovery set passed (106.994s), including pinning and fetch-scope tests. Rebuilt production CLI passed the reviewer's remote-only-issue scenario through both verbs and the complete new-repo workflow. No design expansion; new queued suite and review required.

### 2026-10-21 — BR-2 whole-inventory rule re-plan (pending TL)

Round 2 verified BR-1 and exposed archive evidence omitted from branch/worktree scans (second authoritative-inventory-completeness finding). Stop patching individual locations. The original migration owner was designed for active branch edits, not proof that all legacy evidence is absent. Re-plan its complete rule before another correction.

One root owner: issue.IDDirs(vocabulary home, vocabulary archive) supplies active, flat legacy history and canonical history/issues. One pure top-level issue-path classifier is used by every scope; nested plans/assets/side bundles are not issue documents. An existing non-directory/symlink root or unreadable observation is uncertainty, never empty. No additional persisted format or publication authority.

| Root/scope | Pinned main | Every local branch and refreshed remote branch | Every live local worktree |
|---|---|---|---|
| Active | existing import/conversion | existing changed-active policy; also detect full-tip evidence absent from main, including inherited or unrelated history | pin HEAD (including detached), scan its evidence, then dirty/staged/untracked/ignored issue candidates |
| Flat history | existing archived-card import | unchanged same-path main bytes allowed; added/modified/deleted or otherwise unrepresented archive evidence refuses, requiring publish/resolve under legacy workflow |
| history/issues | same archived-card import | same conservative archive rule | same HEAD and dirty evidence rule |

The worktree column applies to both archive roots as well as active. Enumerate worktrees once through existing Git parser, skip only proven bare/prunable records, and include live worktree HEADs in the same pinned-tip set. Retain current quiescent-writer condition. Branch scans use full-tip presence plus actual base-to-tip changes, so unchanged old branches missing newly published main files do not manufacture deletions. Archive differences are not converted as active cards; they refuse explicitly. Preserve active branch deletion/rename protections, and do not let archive presence count as kept active details.

Integration remains one migrationInventory owner; init consumes its strict empty predicate and migrate uses the same complete input/refusals. Batch blob reads per root/tip through existing TrunkView, no Git process per issue. Cost drivers are roots R=3, pinned tips B and live worktrees W, plus selected issue bytes; expected O(R*B + R*W) metadata calls and linear selected-byte reads, with current output bounds. Count N/2N branches and files to reject extra per-file processes. Existing Git remote caches/objects keep their lifetimes; no new lock, namespace, sidecar or cleanup.

Test the root × scope enumeration as a generated real-Git table through both init and migrate, with main import/refusal distinguished from branch/worktree refusals. Cover create/modify/delete, renames crossing roots, old inherited and unrelated branch evidence, detached clean worktrees, dirty/untracked/ignored candidates, unreadable/unsupported roots, and exclusion of plans/assets/side bundles. Preserve close-anchor and active archive/rename regressions. Build private production CLI for both review reproductions, then targeted suite, recorded final-head suite and whole-issue review. The second queued suite was explicitly canceled before execution and supplies no pass. Await TL approval of this complete rule; if another inventory design gap appears after this re-plan, stop for TL/operator rather than another patch.

### 2026-10-21 — TL-directed narrowing: init is presence-only (supersedes prior BR-2 re-plan)

TL 5119a959-0128-4420-89ee-9a91fb0b08d9 stopped expansion of the importer. The preceding whole-rule importer proposal is superseded, not approved. Pre-existing migrate omissions are filed separately as unclaimed ariadne#546. Keep the already approved BR-1 fresh/pinned head observation; do not add archive importing, branch merge/difference rules, or importer refusal machinery in #543.

Initialization's only rule: before any new bootstrap/cutover, recognized issue evidence present anywhere means refuse with `this repo has issue history; use sdlc issue migrate`. It imports nothing, parses no issue body and resolves no differences. Use one issue.IDDirs-derived root list (active, flat legacy history, history/issues) and one pure top-level issue-path predicate for every scope. Existing completed root/marker idempotence is verified first as already designed.

Presence scopes: pinned main, every fresh local/publication-remote branch tip, every live local worktree HEAD (including detached), each worktree index, and direct filesystem entries in the three roots. Filesystem presence includes dirty/untracked/ignored candidates; index metadata catches a staged issue removed from disk. Root symlinks/non-directories and failed reads refuse as uncertain. Skip only Git-proven bare/prunable worktrees; pin worktrees through existing NUL-delimited parser. Full tip presence includes inherited/unrelated histories without merge-base logic. No difference/normalization/import decision exists in this path.

Implementation: new verifyInitializationEmpty IO helper in issueinit_inventory.go; existing fresh/pinned branch helper reused. Batch snapshot metadata reads select no blob bodies, and one ls-files index metadata query plus direct directory scans per live worktree. Known empty observations produce an ordinary empty MigrationInput/Manifest only after verification, then existing publication phases execute. Retain no new durable file family. Cost O(R*B + R*W + selected metadata entries), R=3, B=pinned unique tips, W=live worktrees; fixed Git process count per root/tip or worktree, never per file. Counted N/2N guard remains.

Verification: generated three-roots × main/local-branch/remote-branch/detached-clean/dirty-worktree/untracked-worktree presence table, plus ignored and staged-index-only candidates, inherited/unrelated tips, root uncertainty, and plans/assets exclusions. All assert no tracker/main/checkout publication. BR-2's two reproductions are init refusal tests; importer completeness acceptance belongs to #546. Preserve current BR-1/migration regressions, private production CLI, recorded suite and close review. TL explicitly approved this narrower scope and said proceed; no additional approval pending.

### 2026-10-21 — presence-only implementation checkpoint

Implemented only the authorized narrower gate. initializationIssuePath is the pure shared path predicate in issueinit_inventory.go; verifyInitializationEmpty is its IO owner, with pinned metadata, index and filesystem readers. Existing migration import/difference logic was not expanded. A verified empty observation supplies an empty manifest to the existing bootstrap/cutover phases. The presence matrix exposed the intended before-fix omissions; an unrelated-history fixture initially used redundant git rm after switch --orphan, was corrected, and then produced its intended RED separately. All presence tests passed (40.539s), including plans/assets exclusions. Private rebuilt production CLI passed both BR-2 reproductions with no mutation, brain skip, and the complete fresh-repo workflow. Broader verification, recorded suite and close review remain.

### 2026-10-22 — BR-2 bounded complete-observation correction

Round 3 verified the archive cases but found a local issue-tracker branch hidden by migrationBranchRefs's name filter. This is an implementation miss against the approved every-tip presence rule, not importer scope. TL a3dd77da-61e5-425e-a901-1c54708447e4 directed separation: publicationBranchTips observes and pins every local/remote tip with no name exemptions; migrationBranchRefs alone filters that complete result for its unchanged importer policy. Init consumes the complete result. Test formerly excluded categories over all three roots. TL requested tracker-already-initialized refusal messaging for a pre-existing tracker branch; confirmed successful and partial exact-root retries were called out for clarification before changing those recovery contracts. No new state or authority. The third queued suite was canceled before execution and is not evidence.

### 2026-10-22 — complete-observation correction verified

TL 7ca0ad18-0fe5-4bf5-99c8-9b5d26749593 confirmed both existing verified recovery cases: exact empty remote bootstrap resumes, and valid completed root/marker is an idempotent success. Only an unverified/pre-existing reserved branch refuses as tracker already initialized. Init never creates a local reserved branch, so local reserved presence is unverified; remote bootstrap recovery retains the existing exact-root/manifest proof.

publicationBranchTips now observes all names/OIDs; migrationBranchRefs alone applies the unchanged importer filter. Before correction, all 18 category/root observation cases failed through a temporary test-only adapter to the unchanged filtered production observer, and four local-reserved CLI cases wrongly initialized; the unverified-remote case lacked the required diagnostic. The adapter was removed before implementation. Focused categories/refusals/recovery/workflow tests passed (76.690s). Private rebuilt production CLI passed all local reserved-branch roots; a real rejected main push retained a published empty bootstrap, then retry reused its root and completed; completed no-op and first issue/claim/start-plan also passed. No new importer rule or state.

| Final additional core concept | Kind | Location | Status |
|---|---|---|---|
| initializationIssuePath | PURE | cmd/sdlc/issueinit_inventory.go | new |
| verifyInitializationEmpty and metadata/index/filesystem readers | INTEGRATION | cmd/sdlc/issueinit_inventory.go | new |
| publicationBranchTips | INTEGRATION | cmd/sdlc/issuemigrate.go | extracted complete observer |
| migrationBranchRefs | INTEGRATION | cmd/sdlc/issuemigrate.go | existing filter over complete observation |
