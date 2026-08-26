# Fleet Thread Inventory Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add JSON-first `sdlc fleet inventory` and `sdlc fleet policy` queries that enumerate every Git working tree in the sibling-repo fleet, report measured facts and provenance-bearing issue metadata, validate repo policy capabilities, and resolve a prospective path for couch; then remove couch's temporary repo-name policy authority in `pair#149`.

**Architecture:** A CUE-authored `fleet-policy` vocabulary defines the portable `.sdlc/fleet.json` declaration. A pure `internal/fleet` core validates declarations, resolves canonical requested paths to admission keys, associates branches with issue metadata, and assembles typed result envelopes. Inventory reports declaration capability because a worktree row alone does not name a target subtree; only the prospective-path query returns a resolved key. Thin filesystem/Git adapters normalize the caller's fleet vantage, reuse the existing filtered sibling walk, enumerate real worktrees through one shared porcelain parser, and collect facts. Cobra and human output render from the same typed results. Pair's `PolicyResolver` adapter consumes the normalized query under `pair#149`, while its registry remains the owner of live occupancy.

**Tech Stack:** Go 1.26, Cobra, CUE/vocabulary export pipeline, stdlib `encoding/json`/`path/filepath`/`os/exec`, and real temporary Git repositories/worktrees for integration tests.

**ARCH alignment.** `ARCH-DRY`: extract the existing project fleet filter into `project.FleetRepoDirs`, move the existing worktree porcelain grammar into `internal/gitx`, and use one loader/resolver for inventory capability and prospective-path queries. `ARCH-PURE`: declaration validation/path resolution, branch association, and result assembly are deterministic functions; filesystem and Git remain injected adapters. `ARCH-PURPOSE`: sdlc owns measured fleet state and policy normalization, while couch derives its admission decision from that result and owns live occupancy/lifecycle; the temporary `PolicyTable` is removed in the same delivery. `ARCH-MOCK`: every collector runs through a stateful `GitReader` fake whose stored repos, refs, commits, worktrees, and dirty entries model the Git commands consumed; portable real-Git tests execute the same scenarios as a conformance check on every full test run.

---

## Decisions and acceptance assumptions

1. The repo-owned declaration path is `.sdlc/fleet.json`. It is tool-namespaced and does not overload `construct/config.json`, which is compiler-owned and absent from several fleet repos.
2. The initial schema version is `1`. Unknown fields and unsupported versions fail closed with a structured `invalid-policy` diagnostic.
3. Admission-key kinds are `repo`, `worktree`, and `declared-root`. A version-1 `declared-root` rule has the exact grammar `<safe literal repo-relative prefix>/*`: one terminal `*` matches exactly one non-empty path segment, and that prefix plus segment is the admission root. Thus `competition/a/x` resolves to `competition/a`, not the full nested path. Rules containing absolute paths, `.`, `..`, empty segments, non-terminal wildcards, or overlapping prefixes are invalid at declaration load time; no ordering/first-match behavior exists. Paths outside every rule return `outside-declared-scope`.
4. Repo identity is the canonical absolute Git common directory. Worktree identity is the canonical absolute worktree root. These are local machine identities, intentionally stable across nested paths and linked worktrees.
5. Divergence uses `origin/main`, then `main`, then an explicit unavailable result. Every row reports the selected `base_ref`; no hidden fallback or derived staleness label exists.
6. The named worktree-provisioned rollout example is `xianxu.dev`.
7. Capacity is tagged. Bounded capacity has `{kind:"bounded", limit:<positive int>}` plus an `onCapacity` action; unbounded capacity is `{kind:"unbounded"}` and omits the unreachable action. Brain is unbounded. Its expected occupancy of fewer than five is explanatory only—no warning or refusal.
8. Worktree creation/removal, live occupancy counting, and branch garbage collection remain outside #200.
9. Every implementation commit uses the repo convention, includes a why-focused body when needed, and ends with `Co-Authored-By: <authoring model> <noreply address>`; the subject-only examples below are shorthand, not permission to omit the trailer.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `FleetPolicyModel` | `pkg/vocab/fleetpolicy.go` | new, embedded CUE-derived declaration contract |
| `PolicyDiagnostic` / `PolicyCapability` / `PolicyResult` | `cmd/sdlc/internal/fleet/policy.go` | new |
| `ResolvePolicy` | `cmd/sdlc/internal/fleet/policy.go` | new, pure over validated policy + canonical paths |
| `IssueAssociation` / `AssociateBranchIssue` | `cmd/sdlc/internal/fleet/issues.go` | new |
| `MeasuredFacts` / `TreeRow` / `Inventory` | `cmd/sdlc/internal/fleet/types.go` | new |
| human renderers | `cmd/sdlc/internal/fleet/render.go` | new, derive only from typed JSON contract |

- **FleetPolicy** is the parsed, validated declaration: version, admission-key rule, and tagged capacity. The bounded variant requires a positive limit and closed on-capacity action; the unbounded variant forbids that unreachable action. The CUE source exports the legal versions/key kinds/capacity kinds/actions and validates the shared instance corpus; the embedded Go binding derives its membership checks from those exported sets. Structural cross-field checks run in the strict Go loader too, with the same valid/invalid fixtures exercised by CUE and Go so the two validators cannot silently drift.
- **PolicyCapability** is the inventory-safe tagged result: either the validated declaration `{policy_version, policy_digest, key_kind, roots, capacity, on_capacity?}` or one declaration diagnostic. The SHA-256 digest derives from the validated canonical declaration rather than raw JSON, so cosmetic rewrites are stable and semantic changes are visible. It contains no admission key because a tree row does not identify a prospective target subtree.
- **PolicyResult** is the prospective-query tagged total result: either a normalized `{policy_version, policy_digest, repo_identity, admission_key, capacity, on_capacity?}` value or one `PolicyDiagnostic`. `on_capacity` exists only for bounded capacity. Missing, malformed, unsupported-version, and outside-rule cases never become `null` and never infer from a repo name.
- **ResolvePolicy** canonicalizes no filesystem itself. It accepts already-canonical repo/common-dir/worktree/requested paths and a validated policy, then deterministically maps the path to a collision key. Inventory and the prospective query share the declaration loader and validator; only `fleet policy --path`, which actually has a requested target path, calls `ResolvePolicy`.
- **IssueAssociation** carries `ref`, `declared_status`, and `provenance: "branch-prefix"`. `AssociateBranchIssue` accepts a whole branch name plus an issue lookup; it rejects slash-prefixed lookalikes, main/detached branches, malformed prefixes, and missing same-repo issues.
- **MeasuredFacts** contains only Git evidence: HEAD SHA, commit timestamp, `base_ref`, ahead, behind, dirty count, and explicit availability/error fields. `TreeRow` adds repo/tree identity, branch state, issue metadata, and policy result without deriving `cold`, `drift`, or actor liveness.

Relationships: one declaration produces one inventory capability and resolves many prospective paths; one repo has one stable repo identity and one-or-more worktree rows; each tree has zero-or-more provenance-bearing issue associations. Future association sources append entries without changing row identity. Future policy key kinds widen the CUE union and pure resolver together; couch remains insulated because it consumes the normalized result.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `project.FleetRepoDirs` | `cmd/sdlc/internal/project/discover.go` | extracted | filtered sibling-directory enumeration |
| `gitx.ParseWorktrees` / `gitx.Worktree` | `cmd/sdlc/internal/gitx/worktree.go` | extracted + widened | `git worktree list --porcelain` grammar |
| `PolicyLoader` | `cmd/sdlc/internal/fleet/load.go` | new | `.sdlc/fleet.json` read + strict decode + vocab validation |
| `gitReader` / `execGitRunner` | `cmd/sdlc/runner.go` | narrowed existing seam | directory-scoped Git commands |
| `GitReader` + stateful `FakeGit` | `cmd/sdlc/internal/fleet/gitreader.go` / `fakegit_test.go` | new | same directory-scoped Git command seam in production/tests |
| fleet-root normalizer | `cmd/sdlc/internal/fleet/gitpaths.go` | new | cwd/nested/linked-worktree to primary-parent fleet root |
| inventory collector | `cmd/sdlc/internal/fleet/inventory.go` | new | fleet walk, Git facts, policy loader, issue reader |
| `sdlc fleet` Cobra group | `cmd/sdlc/fleet.go` | new | `inventory` and `policy` commands |
| help text | `cmd/sdlc/helptext/fleet*.md` | new | embedded CLI contract |

The shell never collapses a failed Git command to an empty string: it appends a structured top-level repo diagnostic and continues the remaining fleet. The production adapter reuses `execGitRunner.GitInDir`. `FakeGit` persists repos keyed by canonical common-dir, each with primary checkout, worktrees, refs, commit graph/timestamps, and dirty path entries; it implements the exact consumed `worktree list`, `rev-parse`, `show`, `status -z`, and `rev-list` behavior. Production flows run against this fake, while portable temporary Git repositories replay the contract scenarios on every full test run as live conformance.

### Function-level test strategy

The named strategies below are authoritative; task-local examples are fixture
seeds, not independent hand-enumerated procedures.

| Function | Adversarial class and mechanical guard |
|---|---|
| `LoadPolicy` / `ValidatePolicy` | Run one shared valid/invalid JSON corpus through CUE and Go; fuzz token streams for duplicate keys, unknown fields, truncation, and cross-field capacity violations; every input returns one bounded typed capability/diagnostic and both validators agree on the corpus. Canonical digest properties pin stability under whitespace/key-order changes and inequality under every semantic field change. |
| `ResolvePolicy` | Property-test canonical contained/outside paths across repo/worktree/declared-root policies; equivalent nested paths preserve keys, distinct declared roots do not collide, and tagged capacity/action is copied without inference. |
| `gitx.ParseWorktrees` | Fuzz arbitrary porcelain records (including missing final separators, optional attributes, spaces, and malformed lines); parsing is bounded, never panics, and valid records round-trip through a canonical fixture encoder. |
| `NormalizeVantage` | Replay primary/nested/linked/symlink vantages against the stateful fake and temporary real Git repos; all equivalent vantages yield the same common-dir/primary/fleet-root tuple. |
| `GitReader` / `FakeGit` | Run one mutable contract trace through the fake and portable real Git adapter; every consumed command form observes equivalent repo/worktree/ref/status state after each mutation. |
| `AssociateBranchIssue` | Property-test the anchored issue-branch grammar over arbitrary branch strings; only a whole valid prefix plus exactly one same-repo issue yields provenance, and all other inputs return a non-null empty association list. |
| `CollectFacts` | Replay arbitrary NUL-delimited porcelain-v1 status records and ref graphs through FakeGit and real Git; dirty entry cardinality, divergence, and explicit unavailable/error states agree without filename parsing or derived cold/drift state. |
| `Inventory` | Mutate a multi-repo FakeGit fleet across calls and inject per-repo failures; stable ordering and complete unaffected rows are invariant, with failures represented once as typed diagnostics. A portable real fleet checks the same interaction contract. |
| `project.FleetRepoDirs` | Property-test arbitrary sibling-name partitions; output is sorted, includes every ordinary sibling (Git filtering is downstream), and excludes exactly the existing fleet filter classes. |
| `JSONContract` | Golden-test the typed algebra rather than scenarios; fuzz marshal/unmarshal round trips and mechanically reject null collections, unknown envelope variants, non-snake-case keys, missing policy identity, and bounded/unbounded field leakage. |
| `NewFleetCmd` | Table-test the Cobra command grammar and run both commands through injected collectors; one routing invariant owns loader sharing, resolver exclusivity, stdout JSON, and refusal exit semantics. |
| `RenderInventory` | Snapshot generated combinations of tagged row facts and diagnostics; renderer output may label fields but a forbidden-token guard rejects newly derived `cold`/`drift` judgments or recollection. |
| declaration rollout | Drive every repo-local declaration through the built prospective query and compare its result to the declaration fixture's expected algebra; no production repo-name branch is allowed. |
| `FleetEndToEnd` | Metamorphically vary caller vantage and prospective nested path while holding the portable fleet constant, then inject declaration/repo failures; normalized inventory/query results remain equivalent and unaffected repos remain complete. |
| Pair `Admission.Decide` / `PolicyResolver` | Table/property-test occupancy at `0`, `limit-1`, `limit`, unbounded, unknown liveness, and provider diagnostics; no refusal forks a child, and stateful fake/provider conformance proves Pair decodes the provider rather than restating policy kinds. |

---

## Chunk 1: Vocabulary, declaration loader, and pure resolver

### Task 1.1: Define and embed the `fleet-policy` noun

**Files:**
- Create: `construct/vocabulary/fleet-policy.cue`
- Create: `construct/vocabulary/testdata/fleet_policy_*.json` shared corpus
- Modify: `construct/vocabulary/vet_test.sh`
- Create: `pkg/vocab/fleetpolicy.go`
- Generate: `pkg/vocab/fleetpolicy.json`
- Create: `pkg/vocab/fleetpolicy_test.go`

- [x] Write the shared CUE/Go corpus and fuzz/property harness named in the function-level strategy; no structural rule may exist in only one validator.
- [x] Run `bash construct/vocabulary/vet_test.sh`; verify the new checks fail because `fleet-policy.cue` is absent.
- [x] Define a closed atomic noun with these JSON field names:

  ```json
  {
    "version": 1,
    "admission": {
      "key": {"kind": "repo", "roots": []},
      "capacity": {"kind": "bounded", "limit": 1},
      "onCapacity": "reject"
    }
  }
  ```

  Brain uses `"capacity":{"kind":"unbounded"}` and omits `onCapacity`. `roots` is required and empty for `repo`/`worktree`; `declared-root` requires at least one safe relative rule and rejects absolute paths or `..` segments. Bounded actions are exactly `reject` or `provision-worktree`.
- [x] Export and embed concrete contract metadata (supported versions, key kinds, actions, declaration path) through `pkg/vocab/fleetpolicy.go`. Expose `FleetPolicy()` membership predicates so the loader never hardcodes the CUE-owned sets. Run the same valid/invalid instance fixtures through CUE and Go to pin the few structural checks that Go must repeat at runtime (the binary does not embed the CUE evaluator).
- [x] Run `make vocab-embed`, `bash construct/vocabulary/vet_test.sh`, and `go test ./pkg/vocab -run FleetPolicy -count=1`; expect PASS.
- [x] Stage the listed vocabulary files (including generated JSON), then commit: `git commit -m "#200: model fleet concurrency policy" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 1.2: Strict loader and diagnostic algebra

**Files:**
- Create: `cmd/sdlc/internal/fleet/types.go`
- Create: `cmd/sdlc/internal/fleet/load.go`
- Create: `cmd/sdlc/internal/fleet/load_test.go`

- [x] Write `LoadPolicy` corpus/fuzz tests from the function-level strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run 'TestLoadPolicy|TestPolicyDiagnosticJSON' -count=1`; verify FAIL on missing package/types.
- [x] Implement strict decoding with `DisallowUnknownFields`, a token-walk duplicate-key rejection helper (the stdlib decoder otherwise accepts duplicates), vocabulary-derived membership checks, and the cross-field rules pinned by the shared corpus. Canonically encode the validated declaration and expose its SHA-256 digest plus schema version on every success. Return tagged `PolicyCapability` values with codes `missing-policy` or `invalid-policy`; preserve the declaration path in the diagnostic. Prospective resolution wraps a successful capability as `PolicyResult` only after a requested path is supplied.
- [x] Re-run the targeted tests; expect PASS.
- [x] Stage the listed loader files, then commit: `git commit -m "#200: load fleet policy with structured diagnostics" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 1.3: Pure requested-path resolver

**Files:**
- Create: `cmd/sdlc/internal/fleet/policy.go`
- Create: `cmd/sdlc/internal/fleet/policy_test.go`

- [x] Write `ResolvePolicy` property tests from the function-level strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestResolvePolicy -count=1`; verify FAIL.
- [x] Implement `ResolvePolicy(policy, canonicalInputs)` without filesystem or Git calls. The IO shell canonicalizes a prospective path by resolving its deepest existing ancestor and appending the clean nonexistent suffix, then performs a symlink-aware containment check; the pure resolver receives that canonical result. Reject ambiguous rules during declaration validation rather than choosing by order or repo name.
- [x] Re-run package tests; expect PASS.
- [x] Stage the listed resolver files, then commit: `git commit -m "#200: resolve prospective paths to admission keys" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Chunk 1 verification and review boundary

- [x] Run `bash construct/vocabulary/vet_test.sh && make vocab-embed && go test ./pkg/vocab ./cmd/sdlc/internal/fleet -count=1 && git diff --check`.
- [x] Request a fresh-context plan/implementation review focused on schema closure, resolver totality, and duplicate validation. Fix Critical/Important findings before proceeding.

---

## Chunk 2: Shared fleet and worktree discovery primitives

### Task 2.1: Extract the filtered fleet sibling walk

**Files:**
- Modify: `cmd/sdlc/internal/project/discover.go`
- Modify: `cmd/sdlc/internal/project/discover_test.go`

- [x] Write `FleetRepoDirs` partition/property tests from the function-level strategy.
- [x] Run `go test ./cmd/sdlc/internal/project -run 'TestFleetRepoDirs|TestDiscoverByIssueRef' -count=1`; verify the new test fails.
- [x] Extract `FleetRepoDirs` as `SiblingRepoDirs` plus the existing `isFleetSibling` filter; route `walkFleetProjects` through it. Do not duplicate filtering in fleet inventory.
- [x] Re-run the project package tests; expect PASS and no change to project discovery behavior.
- [x] Stage the project discovery changes, then commit: `git commit -m "#200: expose the filtered fleet repo walk" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 2.2: Extract and widen worktree porcelain parsing

**Files:**
- Create: `cmd/sdlc/internal/gitx/worktree.go`
- Create: `cmd/sdlc/internal/gitx/worktree_test.go`
- Modify: `cmd/sdlc/state.go`
- Modify: `cmd/sdlc/branchcreate.go`
- Modify: `cmd/sdlc/claim.go`
- Modify: `cmd/sdlc/branchcreate_test.go`

- [x] Move the existing parser corpus into `internal/gitx` and implement the `ParseWorktrees` fuzz/round-trip strategy.
- [x] Run `go test ./cmd/sdlc/internal/gitx ./cmd/sdlc -run 'TestParseWorktrees|TestFindMainWorktree|TestWorktreeForBranch' -count=1`; verify FAIL before the exported parser exists.
- [x] Implement `gitx.ParseWorktrees([]byte) ([]Worktree, error)` as the only porcelain grammar. Keep `sdlc state` JSON backward-compatible by mapping the richer entity back to its existing `{path, branch}` shape.
- [x] Route `findMainWorktree` and `worktreeForBranch` through the shared entity; remove the old parser after all consumers migrate.
- [x] Re-run targeted and `go test ./cmd/sdlc/... -count=1`; expect PASS.
- [x] Stage the worktree parser and migrated consumers, then commit: `git commit -m "#200: share rich git worktree parsing" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 2.3: Normalize fleet vantage and stable Git identity

**Files:**
- Create: `cmd/sdlc/internal/fleet/gitpaths.go`
- Create: `cmd/sdlc/internal/fleet/gitpaths_test.go`

- [x] Implement the `NormalizeVantage` fake/real-Git equivalence strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestNormalizeVantage -count=1`; verify FAIL.
- [x] Implement directory-scoped Git calls through a narrow `gitReader { GitInDir(...) }`: find the containing worktree, parse `git worktree list --porcelain`, select its primary record, resolve `git rev-parse --git-common-dir`, and canonicalize all path comparisons.
- [x] Re-run the targeted tests; expect PASS.
- [x] Stage the fleet Git-path files, then commit: `git commit -m "#200: normalize fleet vantage across linked worktrees" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 2.4: Put stateful Git behavior behind the production seam

**Files:**
- Create: `cmd/sdlc/internal/fleet/gitreader.go`
- Create: `cmd/sdlc/internal/fleet/fakegit_test.go`
- Create: `cmd/sdlc/internal/fleet/git_conformance_test.go`

- [x] Define the narrow directory-scoped `GitReader` used by vantage, facts, and inventory. Adapt `execGitRunner.GitInDir` at the Cobra shell; no collector calls `exec.Command` or `gitx.Capture` directly.
- [x] Build `FakeGit` with persisted repo state: canonical common directory, primary checkout, worktree records, refs/HEADs, commit parent graph and timestamps, plus dirty path entries. Implement the exact command forms consumed by production: `worktree list --porcelain`, `rev-parse`, `show`, `status --porcelain=v1 -z`, and `rev-list --left-right --count`.
- [x] Implement the `GitReader` / `FakeGit` mutable contract strategy; `git_conformance_test.go` runs it against installed Git on every fleet package test run.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run 'TestFakeGit|TestGitConformance' -count=1`; expect PASS.
- [x] Stage the Git seam/fake/conformance files, then commit: `git commit -m "#200: model consumed git behavior with a stateful fake" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Chunk 2 verification and review boundary

- [x] Run `go test ./cmd/sdlc/internal/project ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/fleet ./cmd/sdlc -run 'Test(FleetRepoDirs|ParseWorktrees|FindMainWorktree|WorktreeForBranch|NormalizeVantage)' -count=1 && git diff --check`.
- [x] Request fresh-context review focused on DRY extraction, state JSON compatibility, porcelain edge cases, linked-worktree identity, and fake/real-Git conformance. Fix Critical/Important findings.

---

## Chunk 3: Inventory facts and issue association

### Task 3.1: Provenance-bearing branch-prefix associations

**Files:**
- Create: `cmd/sdlc/internal/fleet/issues.go`
- Create: `cmd/sdlc/internal/fleet/issues_test.go`
- Create: `cmd/sdlc/internal/issue/filename.go`
- Create: `cmd/sdlc/internal/issue/filename_test.go`
- Modify: `cmd/sdlc/issuefiles.go`
- Modify: `cmd/sdlc/issuefiles_test.go`

- [x] Write the `AssociateBranchIssue` grammar property test from the function-level strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestAssociateBranchIssue -count=1`; verify FAIL.
- [x] Extract the filename grammar into `internal/issue.ParseFilename` and keep the package-main helpers as compatibility wrappers for current consumers. Build a branch-specific helper on that shared grammar which anchors at the whole branch start and resolves exactly one file through the existing issue-home/parse logic. Never call `filepath.Base` on the branch before validation.
- [x] Emit `issues: []` on every no-association path and `{ref, declared_status, provenance:"branch-prefix"}` on success.
- [x] Re-run tests; expect PASS.
- [x] Stage the shared filename grammar and association files, then commit: `git commit -m "#200: associate tree issues by branch prefix" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 3.2: Collect measured Git facts per worktree

**Files:**
- Create: `cmd/sdlc/internal/fleet/facts.go`
- Create: `cmd/sdlc/internal/fleet/facts_test.go`

- [x] Implement the `CollectFacts` fake/real-Git conformance strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestCollectFacts -count=1`; verify FAIL.
- [x] Implement directory-scoped commands with errors preserved: `rev-parse HEAD`, `show -s --format=%cI HEAD`, `status --porcelain=v1 -z --untracked-files=all`, main-ref probing, and `rev-list --left-right --count <base>...HEAD`. Parse each porcelain status code: count one status entry, then consume the extra source-path field for rename/copy records. This preserves one dirty item for `R`/`C` while remaining safe for filenames containing newlines.
- [x] Represent unavailable base/error explicitly; do not translate it to zero divergence and do not emit `cold` or `drift`.
- [x] Re-run facts tests; expect PASS.
- [x] Stage the measured-facts files, then commit: `git commit -m "#200: measure worktree git facts" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 3.3: Assemble a failure-tolerant fleet inventory

**Files:**
- Create: `cmd/sdlc/internal/fleet/inventory.go`
- Create: `cmd/sdlc/internal/fleet/inventory_test.go`
- Create: `cmd/sdlc/fleet_integration_test.go`

- [x] Implement the `Inventory` mutation/fault-isolation strategy.
- [x] Run `go test ./cmd/sdlc/internal/fleet ./cmd/sdlc -run 'TestInventory|TestFleetInventory' -count=1`; verify FAIL.
- [x] Implement assembly from `project.FleetRepoDirs`, a Git-repo predicate, shared worktree parser, policy loader, measured facts, and issue association. Each row carries `PolicyCapability` (validated declaration or declaration diagnostic), never a resolved key. Use canonical `(repo_identity, tree_path)` ordering. Keep row and top-level diagnostic arrays non-null.
- [x] Re-run targeted tests; expect PASS.
- [x] Stage the inventory files, then commit: `git commit -m "#200: assemble measured fleet worktree inventory" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Chunk 3 verification and review boundary

- [x] Run `go test ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/project ./cmd/sdlc/internal/gitx ./cmd/sdlc -run 'Test(AssociateBranchIssue|CollectFacts|Inventory|FleetInventory)' -count=1 && git diff --check`.
- [x] Request fresh-context review focused on measured-vs-declared separation, error preservation, branch association provenance, and complete enumeration. Fix Critical/Important findings.

---

## Chunk 4: JSON-first CLI and human rendering

### Task 4.1: Pin the JSON contract

**Files:**
- Create: `cmd/sdlc/internal/fleet/json_test.go`
- Modify: `cmd/sdlc/internal/fleet/types.go`

- [x] Implement the `JSONContract` structural/golden strategy, deriving every
      envelope from the typed result algebra and pinning snake_case fields,
      tagged success/diagnostic variants, non-null arrays, policy version/digest,
      and bounded-only `on_capacity` omission.

- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestJSONContract -count=1`; verify FAIL until tags/envelopes are complete.
- [x] Make the typed contract pass without map-shaped ad hoc encoding.
- [x] Stage the JSON-contract files, then commit: `git commit -m "#200: pin fleet JSON contract" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 4.2: Add `sdlc fleet inventory` and `sdlc fleet policy`

**Files:**
- Create: `cmd/sdlc/fleet.go`
- Create: `cmd/sdlc/fleet_test.go`
- Modify: `cmd/sdlc/main.go`
- Create: `cmd/sdlc/helptext/fleet.md`
- Create: `cmd/sdlc/helptext/fleet-inventory.md`
- Create: `cmd/sdlc/helptext/fleet-policy.md`

- [x] Implement the `NewFleetCmd` command grammar and execution strategy from the function-level table.
- [x] Run `go test ./cmd/sdlc -run 'TestFleet(Command|Inventory|Policy|Help)' -count=1`; verify FAIL.
- [x] Register `NewFleetCmd()` after `state`; wire typed collectors, `json.Encoder`, and human renderers. Default `--path` to `.` only if help and tests make that behavior explicit.
- [x] Re-run tests; expect PASS.
- [ ] Stage the command and help files, then commit: `git commit -m "#200: expose fleet inventory and policy queries" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 4.3: Derive concise human output from typed results

**Files:**
- Create: `cmd/sdlc/internal/fleet/render.go`
- Create: `cmd/sdlc/internal/fleet/render_test.go`

- [x] Implement the `RenderInventory` semantic snapshot strategy from the function-level table.
- [x] Run `go test ./cmd/sdlc/internal/fleet -run TestRender -count=1`; verify FAIL.
- [x] Implement a stable table/detail renderer from `Inventory`; do not recollect or reinterpret data in the view.
- [x] Re-run renderer and CLI tests; expect PASS.
- [ ] Stage the renderer files, then commit: `git commit -m "#200: render fleet facts for humans" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Chunk 4 verification and review boundary

- [x] Run `go test ./cmd/sdlc/internal/fleet ./cmd/sdlc -run 'Test(JSONContract|Fleet|Render)' -count=1 && go test ./cmd/sdlc/... -count=1 && git diff --check`.
- [x] Request fresh-context review focused on JSON stability, refusal exit semantics, shared resolver use, and renderer non-interpretation. Fix Critical/Important findings.

---

## Chunk 5: Named repo declarations, documentation, and fleet smoke

### Task 5.1: Land validated declarations in the named repos

**Files in coordinated peer repos:**
- Create: `ariadne/.sdlc/fleet.json` — `repo`, bounded 1, reject
- Create: `pair/.sdlc/fleet.json` — `repo`, bounded 1, reject
- Create: `parley.nvim/.sdlc/fleet.json` — `repo`, bounded 1, reject
- Create: `brain/.sdlc/fleet.json` — `repo`, unbounded, no on-capacity action
- Create: `kbench/.sdlc/fleet.json` — `declared-root`, roots `competition/*`, bounded 1, reject
- Create: `xianxu.dev/.sdlc/fleet.json` — `worktree`, bounded 1, provision-worktree

- [x] Before touching each peer, verify its clean/dirty state and read `AGENTS.local.md` plus `MEMORY.md` when present. Preserve unrelated user changes.
- [x] Add each declaration with no repo-name branch in Go. Validate every instance through the built `sdlc fleet policy --path ... --json` path.
- [x] Run the declaration-rollout conformance strategy from the function-level table against every named repository.
- [x] Commit declarations separately in each owning repo using its issue/workflow conventions. Do not bundle unrelated local changes.

### Task 5.2: Update Ariadne's map and issue checklist

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/workflow/vocabulary.md`
- Modify: `atlas/index.md` only if a new atlas page is needed
- Modify: `workshop/issues/000200-sdlc-fleet-thread-inventory.md`

- [ ] Document command ownership, declaration location, resolver boundary, diagnostics, and measured-vs-declared output. Map details to code/tests rather than duplicating the JSON schema.
- [ ] Verify the issue's 2026-08-24 revision and reconciled checklist still match the shipped contract: measured facts are juxtaposed with provenance-bearing declared status; no drift/cold verdict is emitted.
- [ ] Tick every completed issue Plan item and log the peer declaration commits plus verification evidence.
- [ ] Stage the atlas and issue updates, then commit: `git commit -m "#200: document fleet inventory and policy" -m "Co-Authored-By: Codex <noreply@openai.com>"`.

### Task 5.3: End-to-end verification from all vantages

- [ ] Build the local binary: `go build -o bin/sdlc ./cmd/sdlc`.
- [ ] Run the `FleetEndToEnd` metamorphic/fault-isolation strategy from the function-level table against the built binary and test-only fleet.
- [ ] Run the full gates:

  ```sh
  bash construct/vocabulary/vet_test.sh
  make vocab-embed
  go test ./pkg/vocab -count=1
  go test ./cmd/sdlc/internal/project ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/fleet -count=1
  go test ./cmd/sdlc/... -count=1
  go test ./... -count=1
  git diff --check
  ```

- [ ] Record the provider verification in #200's Log, but do not close while couch's temporary `PolicyTable` remains authoritative.

---

## Chunk 6: Pair consumer integration (`pair#149`)

This is a coordinated cross-repo delivery, but Pair keeps its own SDLC artifact and review gates. Before Pair code changes, claim/start-plan `pair#149`, append the normalized admission algebra to its existing revision history, and author/review `pair/workshop/plans/000149-couch-opaque-tags-and-a-human-naming-layer-plan.md`. That plan names **M1 — normalized policy consumer** as a real review boundary. Provider Chunks 1–5 land first; Pair then implements M1 and runs `sdlc milestone-close --issue 149 --milestone M1`; Ariadne records that reviewed commit, closes and merges #200; Pair rebases/updates to the merged provider and completes its later identity/name milestones before issue-close. Pair #149 is therefore a coordinated consumer, not blocked on the entire #200 issue state, and its frontmatter carries no circular whole-issue dependency.

### Task 6.1: Replace the shadow policy source with an sdlc query adapter

**Expected Pair surfaces (the Pair plan confirms exact ownership):**
- Remove/replace: `cmd/internal/couchcore/policy.go` (`Mode`, `PolicyTable`)
- Modify: `cmd/internal/couchcore/store.go` (stop reading unscoped `policy.json`)
- Modify: `cmd/internal/couchcore/couch.go` (inject/query normalized policy before spawn)
- Modify: `cmd/internal/couchcore/registry.go` (capacity by admission key, not tree/repo name)
- Modify: `cmd/internal/couchcmd/run.go` (render `reject` vs `provision-worktree` from normalized action)
- Add: a typed `PolicyResolver` adapter + JSON decoder and tests

- [ ] Define Pair's pure admission entity as `{repo_identity, admission_key, capacity, on_capacity?}` plus current live occupancy. Bounded capacity admits below its limit and returns its typed action at the limit; unbounded always admits and carries no action. No Pair enum restates repo/worktree/declared-root key kinds because couch consumes the resolved key.
- [ ] Put the external `sdlc fleet policy --path <requested> --json` call behind `PolicyResolver`. Its stateful fake stores path-keyed normalized results/diagnostics and is injected through the same Couch composition boundary; a live conformance test runs against a temporary repo plus the built sdlc provider on every relevant Pair full-suite run (ARCH-MOCK).
- [ ] Query before child fork. Count only records whose PID/identity probe is live and whose persisted normalized admission key equals the prospective result. Prune dead records first; unknown liveness remains occupied conservatively.
- [ ] Persist policy version/digest plus normalized repo/admission identities on the thread/incarnation record. Before later occupancy checks, compare them with the current provider result and re-resolve stale live/unknown incumbents; a legacy or unresolved record conservatively blocks, never disappears.
- [ ] Remove `Store.policyPath`, `policy.json` loading, `PolicyTable.Mode(repo)`, `InPlaceSerial`, `WorktreeParallel`, and `HeavyLocalState`. Add a shadow-sweep test that fails if those symbols or the old policy file authority remain.
- [ ] Implement the Pair `Admission.Decide` / `PolicyResolver` strategy above.
- [ ] Keep human rendering actionable: `reject` offers switch/stop; `provision-worktree` offers the provision action (not a fabricated path—path allocation belongs to the later worktree lifecycle owner). Do not re-infer the action from repo name.

### Task 6.2: Cross-repo contract and close gates

- [ ] Build Ariadne's `sdlc`, point the Pair live-conformance test at it, and verify the Pair decoder accepts success plus every structured diagnostic fixture emitted by the provider.
- [ ] Run Pair's targeted policy/registry/store/couchcmd tests, full `go test ./...`, race target required by its local plan, and `git diff --check`; cross the exact `pair#149` M1 boundary with `sdlc milestone-close --issue 149 --milestone M1`.
- [ ] In Ariadne, add the Pair commit/review evidence to #200's Log and run the complete provider suite again.
- [ ] Run `sdlc close --issue 200 --verified '<provider tests, four-vantage smoke, declarations, Pair consumer/conformance commit and removal shadow-sweep>'`; let the close gate measure actual time and dispatch the mandatory fresh-context review. Fix Critical/Important findings and re-run verification before crossing the boundary.

---

## Out-of-scope follow-ons

- A worktree lifecycle issue provisions, tracks, retires, and garbage-collects branches/worktrees.
- A future association source may append project/space metadata to `issues` without changing tree identity or branch-prefix provenance.

## Revisions

### 2026-08-24 — represent brain capacity as unbounded

**Reason:** the operator expects normally fewer than five concurrent brain
threads but does not want that observation enforced as a limit.

**Delta:** replaced the proposed positive `liveCapacity` scalar with a tagged
bounded/unbounded capacity. Bounded capacity requires a positive limit and an
on-capacity action. Unbounded capacity omits that unreachable action and always
admits. Brain's rollout declaration is explicitly unbounded; Pair conformance
tests admit more than five live occupants without warning or refusal.

### 2026-08-25 — make tests strategic and the cross-repo gates acyclic

**Reason:** plan-quality round 1 found case-by-case test scripts obscuring the
risky functions, one sentence falsely implying inventory could resolve a key
without a target path, and a circular whole-issue dependency between the
provider and its motivating Pair consumer.

**Delta:** one function-level test-strategy table now owns adversarial classes
and mechanical guards; task test bullets reference those strategies rather than
restating cases. Inventory shares only loading/validation, while the
prospective-path query alone invokes `ResolvePolicy`. Pair #149 has no
whole-issue dependency on #200: reviewed M1 consumes the locally built provider,
then #200 closes/merges, then Pair completes its remaining milestones.

### 2026-08-25 — version persisted admission evidence

**Reason:** Pair #149's transactional store revealed that persisted normalized
keys need an authority-change detector rather than being trusted indefinitely.

**Delta:** successful capabilities/results now include schema version and a
canonical semantic declaration digest. Loader tests pin digest stability and
change detection; Pair M1 persists the evidence and re-resolves stale
live/unknown occupants before admitting another actor.
