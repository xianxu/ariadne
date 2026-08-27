# Compact Boundary Review Manifest Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace milestone-review's embedded unified diff with a compact, pinned, repository-inspection manifest without changing any other judge contract.

**Architecture:** A pure `ReviewWindowManifest` in the judge package owns the two review-window variants and renders shell-quoted, read-only Git recipes from resolved values. A thin main-package integration resolves repository/ref/path state through an injected Git runner, then both automatic close and manual `judge milestone-review` feed the same manifest into the existing prompt/orientation/verdict pipeline. Automatic dispatch preserves its literal `workshop/history` exclusion; manual dispatch preserves its effective directory flags.

**Tech Stack:** Go, Cobra, Git CLI, embedded Markdown prompt templates, Go unit/integration tests.

**Execution note:** The operator explicitly chose the live in-place Ariadne branch because `bin/sdlc` and derivative repositories consume this checkout through symlinks; do not create a second implementation worktree.

---

## Chunk 1: Manifest, resolution, and dispatch

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `ReviewWindowManifest` | `cmd/sdlc/internal/judge/reviewwindow.go` | new |
| `ReviewCommand` | `cmd/sdlc/internal/judge/reviewwindow.go` | new |
| `RenderReviewWindow` | `cmd/sdlc/internal/judge/reviewwindow.go` | new |

- **`ReviewWindowManifest`** — immutable resolved review range plus repository and exclusion paths; represents either a committed range (`BaseSHA` + `HeadSHA`) or working tree (`BaseSHA` + `AmbientHeadSHA`).
  - **Relationships:** one manifest belongs to one boundary-review prompt; one manifest owns four command recipes.
  - **DRY rationale:** automatic close and manual milestone review need one definition of the reviewed range and exclusions.
  - **Future extensions:** add another read-only recipe without changing dispatch or prompt assembly.
- **`ReviewCommand`** — structured argv plus a human label; no shell source is stored or executed.
  - **Relationships:** four commands belong to one manifest (stat, name-status, full, targeted).
  - **DRY rationale:** one shell-quoting renderer serves every recipe.
  - **Future extensions:** expose machine-readable argv separately from display text.
- **`RenderReviewWindow`** — validates already-resolved values and renders bounded manifest prose.
  - **Relationships:** consumes one manifest and produces one prompt fragment.
  - **DRY rationale:** prompt template and dispatch callers do not construct Git syntax independently.
  - **Future extensions:** alternate presentation formats can reuse the same entity.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `boundaryGitRunner` | `cmd/sdlc/reviewwindow.go` | new | Git CLI |
| `resolveReviewWindow` | `cmd/sdlc/reviewwindow.go` | new | repository root/ref/path resolution |
| `boundaryReviewDispatchOptions` | `cmd/sdlc/milestoneclose.go` | modified | automatic close review dispatch |
| `runJudge` milestone branch | `cmd/sdlc/judge.go` | modified | manual milestone-review dispatch |

- **`boundaryGitRunner`** — interface with `RunGit(args ...string) ([]byte, error)`; production adapter calls `gitx.RunGit`.
  - **Injected into:** `resolveReviewWindow`.
  - **Fake:** a stateful test runner holds a ref-to-object map, repository root, failures, and ordered argv history.
  - **Future extensions:** reviewer isolation (#204) can provide a runner rooted in a disposable checkout.
- **`resolveReviewWindow`** — invokes `git rev-parse --show-toplevel` and `git rev-parse --verify <ref>^{commit}`, distinguishes committed/working-tree modes, and validates issue/optional-plan paths before creating the pure manifest.
  - **Injected into:** automatic and manual boundary prompt builders.
  - **Future extensions:** alternative checkout roots without changing pure rendering.
- **`boundaryReviewDispatchOptions`** — replaces `collectDiff` only for automatic milestone-review and carries the effective issues directory, literal `workshop/history`, and captured optional plan path.
  - **Injected into:** existing fresh-process `judge.Dispatch` seam.
  - **Future extensions:** automatic history override only via a separate issue and flag contract.
- **`runJudge` milestone branch** — resolves manual flags before prompt construction while leaving dry/pure/plan/specs byte-compatible.
  - **Injected into:** existing Cobra inputs and dispatch seam.
  - **Future extensions:** none planned; YAGNI keeps other judges inline-diff based.

### Task 1: Pure manifest and bounded rendering

**Files:**
- Create: `cmd/sdlc/internal/judge/reviewwindow.go`
- Create: `cmd/sdlc/internal/judge/reviewwindow_test.go`

- [ ] **Step 1: Write failing committed-range renderer tests**

Construct a manifest with full 40-character `BaseSHA`/`HeadSHA`, a repository path containing a space/single quote, custom issue/history directories, issue/plan paths, and assert the exact four argv displays:

```text
git -C <root> diff --stat <base> <head> -- :!<issues>/ :!<history>/
git -C <root> diff --name-status <base> <head> -- :!<issues>/ :!<history>/
git -C <root> diff <base> <head> -- :!<issues>/ :!<history>/
git -C <root> diff <base> <head> -- <path-from-name-status> :!<issues>/ :!<history>/
```

Assert each argv element is POSIX-shell quoted for display and the output names committed mode, issue file, optional plan file, and pinned SHAs.

- [ ] **Step 2: Write failing working-tree and validation tests**

Assert working-tree recipes omit a head argument but display `AmbientHeadSHA`; assert untracked files are explicitly out of scope. Reject missing/non-full SHAs, inconsistent mode fields, relative repository root, and missing base/head combinations.

- [ ] **Step 3: Run the focused tests and confirm RED**

Run: `go test ./cmd/sdlc/internal/judge -run 'TestReviewWindow' -count=1`

Expected: FAIL because the entities do not exist.

- [ ] **Step 4: Implement the minimal pure entities**

Use structured `[]string` argv internally. Implement one `shellQuoteArg` helper for display only; never feed its output to a shell. Keep the targeted command's selected path as a documented literal substitution token, not user-interpolated executable text.

- [ ] **Step 5: Prove bounded output**

Add a multi-megabyte sentinel only to an unrelated synthetic diff input and assert `RenderReviewWindow` output contains neither sentinel bytes nor content proportional to it. Assert a conservative upper bound for the fixed fixture.

- [ ] **Step 6: Run focused tests and commit**

Run: `go test ./cmd/sdlc/internal/judge -run 'TestReviewWindow' -count=1`

Expected: PASS.

Commit: `git commit -am "#162: add compact review window manifest"` after adding new files explicitly.

### Task 2: Git resolution seam with fake and live conformance

**Files:**
- Create: `cmd/sdlc/reviewwindow.go`
- Create: `cmd/sdlc/reviewwindow_test.go`
- Modify: `cmd/sdlc/orientation.go`
- Test: `cmd/sdlc/orientation_test.go`

- [ ] **Step 1: Write the stateful fake and failing resolver tests**

The fake records ordered argv and returns configured values for:

```text
rev-parse --show-toplevel
rev-parse --verify <base>^{commit}
rev-parse --verify <head>^{commit}
```

Cover explicit committed head, omitted head resolving ambient `HEAD`, symbolic input becoming full object IDs, missing/non-commit refs, unavailable root, issue-file lookup failure, and optional canonical plan discovery.

- [ ] **Step 2: Run resolver tests and confirm RED**

Run: `go test ./cmd/sdlc -run 'TestResolveReviewWindow' -count=1`

Expected: FAIL because the resolver and runner interface do not exist.

- [ ] **Step 3: Implement the integration seam**

Add `boundaryGitRunner`, production adapter, and `resolveReviewWindow`. Keep object-ID format validation pure. Return errors with the failed ref/path and next action; do not fall back to `<unknown-repo>` for a boundary dispatch.

- [ ] **Step 4: Add real-Git conformance tests**

Using the existing temporary repository helpers, create base/reviewed/later commits plus staged, unstaged, and untracked files. Assert explicit head pins the reviewed commit; omitted head's rendered commands cover committed-after-base plus staged/unstaged tracked state and exclude untracked files when executed.

- [ ] **Step 5: Run focused tests and commit**

Run: `go test ./cmd/sdlc -run 'Test(ResolveReviewWindow|BoundaryOrientation)' -count=1`

Expected: PASS.

Commit: `git commit -am "#162: resolve pinned boundary review windows"` after adding new files explicitly.

### Task 3: Automatic boundary dispatch wiring

**Files:**
- Modify: `cmd/sdlc/milestoneclose.go`
- Modify: `cmd/sdlc/close.go`
- Modify: `cmd/sdlc/closereview_test.go`
- Modify: `cmd/sdlc/boundaryledger_test.go`
- Modify: `cmd/sdlc/internal/judge/prompts.go`
- Modify: `cmd/sdlc/internal/judge/prompts/milestone-review.md`
- Modify: `cmd/sdlc/internal/judge/testdata/golden/milestone-review.prompt`

- [ ] **Step 1: Write failing prompt and dispatch tests**

Update the pinned-head regression to assert `REVIEWED_MARKER`/`LATER_MARKER` patch bytes are both absent, while full reviewed SHA and the exact pinned Git recipes are present. Add a multi-megabyte tracked file and assert automatic `DispatchOptions.Prompt` remains bounded and sentinel-free. Assert automatic commands use the effective `IssuesDir` and literal `workshop/history`.

- [ ] **Step 2: Add failure-path tests**

Assert automatic dispatch returns `ok=false` before agent launch for empty/symbolic/unresolvable anchors, inaccessible repository root, or missing issue file. Assert the reason becomes `VerdictNotRun` through `dispatchBoundaryReview`.

- [ ] **Step 3: Run focused tests and confirm RED**

Run: `go test ./cmd/sdlc -run 'TestBoundaryReviewDispatchOptions|TestDispatchBoundaryReview' -count=1`

Expected: FAIL because the prompt still embeds diff bytes.

- [ ] **Step 4: Wire the manifest into automatic dispatch**

Add a `ReviewWindow` field to `judge.PromptInput`, replace milestone template's `Diff:` block with `Review window:` and the rendered manifest token, and stop calling `collectDiff` from `boundaryReviewDispatchOptions`. Preserve orientation fields, `PriorFindings`, agent/tool resolution, findings parsing, sidecar persistence, and concrete captured head. Capture optional plan path before `reviewThenFinalizeLocked` blanks `PlansDir`; carry it as immutable dispatch data rather than reading after the lock is released.

- [ ] **Step 5: Update only the intentional golden**

Regenerate or edit only `milestone-review.prompt`, inspect its diff, then run the full judge package golden test to prove every non-boundary category is byte-identical.

- [ ] **Step 6: Run focused suites and commit**

Run:

```text
go test ./cmd/sdlc/internal/judge -count=1
go test ./cmd/sdlc -run 'Test(BoundaryReview|DispatchBoundary|ReviewThenFinalize)' -count=1
```

Expected: PASS.

Commit: `git commit -am "#162: send manifests to automatic boundary reviews"`.

### Task 4: Manual milestone-review wiring and compatibility

**Files:**
- Modify: `cmd/sdlc/judge.go`
- Modify: `cmd/sdlc/judge_command_test.go`
- Modify: `cmd/sdlc/collectdiff_test.go`

- [ ] **Step 1: Write failing manual-mode tests**

For `milestone-review`, assert explicit `--base`/`--head` resolve to full commits and custom `IssuesDir`/`HistoryDir` appear in every recipe. With omitted head, assert working-tree mode, ambient `HEAD`, and no head argv. Assert invalid refs fail before dispatch with a precise message.

- [ ] **Step 2: Snapshot non-boundary prompts before changing code**

Build dry/pure/plan/specs prompt output from the existing golden input and retain the bytes in the current golden fixtures; do not route those categories through the new resolver.

- [ ] **Step 3: Run focused tests and confirm RED**

Run: `go test ./cmd/sdlc -run 'TestJudge.*Milestone|TestCollectDiff' -count=1`

Expected: new milestone tests FAIL; existing collectDiff tests PASS.

- [ ] **Step 4: Special-case milestone review before `collectDiff`**

Resolve its review window and build the shared manifest/orientation input. Leave the existing `collectDiff` path byte-for-byte for dry/pure/plan/specs. On omitted head, document that Git diff includes committed-after-base plus staged/unstaged tracked files and excludes untracked files.

- [ ] **Step 5: Run compatibility tests and commit**

Run:

```text
go test ./cmd/sdlc -run 'TestJudge|TestCollectDiff' -count=1
go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -count=1
```

Expected: PASS; only milestone-review's golden changed.

Commit: `git commit -am "#162: use pinned manifests in manual milestone review"`.

### Task 5: Documentation, complete verification, and live binary

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/workflow/pre-merge-checks.md`
- Modify: `workshop/issues/000162-milestone-close-window-base.md`
- Modify: `workshop/plans/000162-milestone-close-window-base-plan.md`

- [ ] **Step 1: Update atlas wording**

Document that boundary reviewers receive pinned repository commands instead of patch bytes, distinguish committed/working-tree scope, state untracked exclusion, and note automatic-versus-manual directory sources. Keep `atlas/index.md` unchanged because both existing pages are already linked.

- [ ] **Step 2: Run shadow-document and stale-contract sweeps**

Run:

```text
rg -n 'milestone-review|boundary review|Diff:|embedded diff|unified diff|stdin|temp file' atlas cmd/sdlc workshop/issues/000162-* workshop/plans/000162-*
git diff --check
```

Expected: no live documentation claims that boundary prompts embed unified diff bytes; historical issue revisions may retain superseded rationale.

- [ ] **Step 3: Run complete verification**

Run:

```text
go test ./cmd/sdlc/internal/judge -count=1
go test ./cmd/sdlc/... -count=1
go test ./... -count=1
git diff --check
```

Expected: all PASS.

- [ ] **Step 4: Build and smoke-test the live in-place binary**

Build using the repository's canonical target discovered from `Makefile`, then run a manual dry-run milestone review against a small local range. Assert the prompt contains the pinned manifest and no patch hunk.

- [ ] **Step 5: Update durable execution state and commit**

Check completed issue/plan boxes, append verification and architectural decisions to `## Log`, and commit with:

```text
git commit -am "#162: document compact boundary review transport"
```

- [ ] **Step 6: Cross the SDLC boundary**

Run `sdlc close --issue 162 --verified '<exact test and smoke evidence>'`, address the binary-owned fresh review, merge/push according to the SDLC next-action output, and confirm the live `bin/sdlc` used by Pair contains the fix.

## Revisions

### 2026-08-26 — resolve plan-review blockers before execution

**Reason:** fresh plan review found a collision with the existing close-window
anchor resolver, an omitted fail-closed prompt edit, an undecided manual API,
stale help/process-manual files outside the task lists, and a vacuous pure-layer
boundedness test.

**Delta (supersedes the conflicting steps above):**

1. **Preserve the existing `resolveReviewWindow` in
   `cmd/sdlc/milestoneclose.go`.** It computes the close/atlas/review anchor tuple
   and is already shared with `close.go`; renaming or replacing it would create
   churn at the ARCH-DRY boundary. The new integration is named
   `resolveBoundaryReviewManifest`. Automatic callers first use the existing
   resolver under the repo lock, then pass its concrete `BaseLong`/`Head` into
   the new manifest resolver after unlock. Add the new tests to
   `cmd/sdlc/reviewwindow_test.go`; retain and run
   `cmd/sdlc/milestonewindow_test.go` unchanged to prove anchor behavior has not
   moved.

2. **Make reviewer inspection failure explicit.** Task 3 also modifies
   `cmd/sdlc/internal/judge/code-review.md` and its tests. Replace “Read the
   diff” with wording equivalent to:

   ```text
   Use the Review window commands below to inspect the pinned repository range.
   Run at least the stat and name-status recipes before targeted/full patch
   inspection. If the repository, pinned objects, or a required read-only
   command is unavailable, do not infer from prose: return REWORK and identify
   the failed inspection.
   ```

   Add a prompt-contract test that pins `return REWORK`, unavailable inspection,
   and read-only command language. The structured verdict parser remains the
   mechanical fail-closed layer for missing/unknown verdicts (ARCH-PURPOSE).

3. **Keep manual invocation without `--issue` compatible.** Manual
   `sdlc judge milestone-review` continues to allow no issue number. In that
   mode, orientation uses the actual repo, `IssueRef` is `<unspecified>`,
   `IssueFile` and `PlanFile` are empty, and no issue/plan path validation runs;
   ref/root validation and the complete manifest still apply. When `--issue N`
   is supplied, the issue path is required and the optional plan path is
   discovered. Add `PlansDir string` to `judgeFlags` and a standard
   `--plans-dir` flag defaulting through `WF_PLANS_DIR` / `workshop/plans`, so a
   custom manual tracker can name its canonical plan without guessing. Test
   issue-less, issue-present, missing-issue, default-plan, and custom-plan-dir
   modes; update help text for the preserved optional `--issue` behavior.

4. **Own every live documentation consumer.** Task 5's file list additionally
   modifies:

   - `cmd/sdlc/helptext/judge.md`
   - `cmd/sdlc/helptext/milestone-close.md`
   - `atlas/process-manual.md`
   - `atlas/workflow/process-manual.md`

   The two process-manual files are generated consumers: update their source
   prompt/help files, regenerate them using the repository's documented
   process-manual target, and inspect the generated diff rather than editing
   their content independently. Replace “prompt + diff” with “prompt + pinned
   review manifest” only for milestone-review; retain inline-diff wording for
   dry/pure/plan/specs. “Diff window” may remain when it describes semantic Git
   range selection rather than prompt transport.

5. **Move boundedness proof entirely to dispatch integration.** Delete Task 1
   Step 5's unrelated synthetic sentinel. In Task 3, create and commit a file
   containing a multi-megabyte sentinel inside the exact pinned review range;
   first run Git diff in the fixture and assert the sentinel is present, then
   build `DispatchOptions` and assert its prompt is bounded and sentinel-free.
   This proves the old input would have been large before proving the new prompt
   excludes it.

**Revised focused verification additions:**

```text
go test ./cmd/sdlc -run 'Test(ResolveReviewWindow|ResolveBoundaryReviewManifest|BoundaryReviewDispatchOptions|Judge.*Milestone)' -count=1
go test ./cmd/sdlc/internal/judge -run 'Test(ReviewWindow|BoundaryReviewInspectionContract|BuildPrompt_Golden)' -count=1
```

### 2026-08-26 — dispose plan-gate findings PQ-1 and PQ-2

**Reason:** the binary-owned plan-quality gate found that the earlier review
focused on transport and accidentally preserved a first-milestone fallback that
still violates the issue's original branch-point contract. It also rejected the
procedural test-case inventories in Tasks 1–4.

**PQ-1 (`deliver-full-stated-purpose`) — addressed:** modify
`boundaryWindowBase` in `cmd/sdlc/milestoneclose.go`, not the anchor tuple shape
of `resolveReviewWindow`. When no prior milestone boundary exists, use
`gitx.MergeBaseWithMain()` first and fall back to `branchStartByIssue` only for
the existing direct-on-main/no-divergence case. This keeps atlas and review on
the one shared base source (ARCH-DRY) while making a feature branch's first
milestone start at its actual branch point (ARCH-PURPOSE). Update
`cmd/sdlc/milestonewindow_test.go` rather than preserving it unchanged.

**PQ-2 (`test-strategy-contract`) — addressed:** all case inventories and
procedural assertion lists earlier in this plan are superseded as test design;
the executor derives concrete cases in Go. The executable test contract is only
the following named risky-function strategies:

| Function | Strategy |
|----------|----------|
| `boundaryWindowBase` | adversarial issue history predating branch divergence → a real-Git fixture mechanically compares the selected base with `merge-base main HEAD`, while prior-boundary precedence and direct-on-main fallback remain invariant |
| `RenderReviewWindow` | arbitrary paths/SHAs/mode combinations → table/property-style unit tests mechanically enforce structural validation, argv ordering, and round-trip-safe POSIX display quoting without IO |
| `resolveBoundaryReviewManifest` | missing, symbolic, non-commit, and changing Git state → a stateful fake records every argv and mechanically permits output only after root/commit/path validation; a temporary Git repo checks fake conformance |
| `boundaryReviewDispatchOptions` | a pinned range containing arbitrarily large patch bytes and a later moving HEAD → integration tests first prove bytes exist in the range, then enforce bounded prompt size, sentinel absence, and concrete-anchor presence |
| `runJudge` milestone branch | optional issue identity and explicit/omitted refs under custom directories → command-level tests mechanically enforce the documented committed/working-tree variants while existing non-boundary goldens enforce byte compatibility |
| `BuildPrompt` milestone template | unavailable repository/tool inspection → prompt-contract tests mechanically require the REWORK instruction and the manifest token; category goldens prevent shadow drift |

**Revised TDD order:** make `boundaryWindowBase`'s adversarial real-Git test red
first and fix the shared base helper; then implement the pure manifest tests,
Git-resolution fake/conformance tests, automatic dispatch boundedness test,
manual command tests, and prompt-contract/golden tests in that order. Each named
surface must go red before its implementation change and green before its task
commit. No earlier prose enumeration is an additional acceptance contract.
