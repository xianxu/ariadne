# Boundary Review — ariadne#162 (whole-issue close)

| field | value |
|-------|-------|
| issue | 162 — sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD) |
| repo | ariadne |
| issue file | workshop/issues/000162-milestone-close-window-base.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a3e7d47247f7c9ca301495416410ca2bb15b01..658640a9d4af7b5264cf90a880639ab5bcd0feb2 |
| command | sdlc close --issue 162 |
| reviewer | codex |
| timestamp | 2026-08-26T22:24:07-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The branch-point correction and compact review-manifest architecture are sound, but the rendered commands fail their tracker-exclusion contract when directory overrides are absolute. The live-Git conformance test also does not exercise the generated commands, leaving exactly this failure undetected.

```findings
findings:
  - id: new
    severity: Critical
    family: repo-relative-pathspecs
    title: |
      Absolute tracker-directory overrides do not exclude tracker files
    detail: |
      cmd/sdlc/internal/judge/reviewwindow.go:115 inserts IssuesDir and HistoryDir directly into Git pathspecs. Git matches tracked paths relative to the repository, so absolute in-repository values from --issues-dir, --history-dir, or their environment variables fail to exclude those files. Running the rendered shape against this pinned range included workshop/issues/000162-milestone-close-window-base.md, while the relative-path recipe excluded it. Normalize in-repository exclusion directories relative to RepoRoot before rendering, and pin both directory fields with an executable real-Git regression. This violates ARCH-PURPOSE because the manifest does not preserve the promised exclusions.
  - id: new
    severity: Important
    family: external-command-conformance
    title: |
      Live-Git conformance does not execute the rendered review recipes
    detail: |
      cmd/sdlc/reviewwindow_test.go:138-167 checks only ref resolution and mode fields. It never executes the rendered committed or working-tree commands, nor creates the promised staged, unstaged, untracked, and custom-directory states. Consequently the checked Task 2 Step 4 and ARCH-MOCK conformance claim are not delivered. Execute the structured argv against a temporary repository and assert range contents and exclusions.
```

### Strengths

- `boundaryWindowBase` now uses the feature branch point for the first milestone, with an adversarial filed-early/unrelated-history regression.
- `ReviewWindowManifest` and `RenderReviewWindow` are genuinely pure; Git and filesystem work remain in a thin integration seam.
- The multi-megabyte sentinel test proves the old patch would be large and that patch bytes no longer enter the prompt.
- Automatic reviews reject symbolic anchors, preserving the previously established moving-HEAD protection.
- Atlas and generated process documentation cover the new transport and working-tree semantics.

### Critical findings

- `cmd/sdlc/internal/judge/reviewwindow.go:115`: normalize absolute in-repository tracker directories to repository-relative pathspecs. Add a regression that fails with the current implementation.

### Important findings

- `cmd/sdlc/reviewwindow_test.go:138`: extend live conformance beyond SHA resolution to execute every relevant generated command against real Git state.

### Minor findings

None.

### Test coverage notes

- `go test ./cmd/sdlc/internal/judge -count=1` passed.
- Focused `cmd/sdlc` boundary-window, manifest, dispatch, manual-judge, and golden tests passed.
- `git diff --check` passed.
- `go test ./... -count=1` could not complete cleanly in this reviewer sandbox: `TestClose_MilestoneRefusesWithRedirect` was denied permission to create `.git/sdlc.lock`. Other reported packages passed.
- The absolute-path failure was directly reproduced with read-only Git commands against the pinned range.

### Architectural notes

- `ARCH-DRY`: pass — automatic and manual review share one renderer and boundary-base computation.
- `ARCH-PURE`: pass — manifest construction is pure; Git/path resolution is isolated.
- `ARCH-PURPOSE`: flag — absolute overrides defeat the promised tracker exclusions.
- `ARCH-MOCK`: flag — the fake records ref-resolution calls, but live conformance does not cover the Git diff behavior the feature relies upon.

### Plan revision recommendations

Append a revision recording:

- In-repository absolute tracker directories are converted to Git-root-relative exclusion pathspecs.
- The live-Git conformance fixture executes committed and working-tree recipes, covering relative and absolute issue/history directories plus committed, staged, unstaged, and untracked content.
- Task 2 Step 4 was found incomplete at boundary review and is reopened until that executable coverage passes.

---

## Re-review — 2026-08-26T22:38:08-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 162 — sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD) |
| repo | ariadne |
| issue file | workshop/issues/000162-milestone-close-window-base.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a3e7d47247f7c9ca301495416410ca2bb15b01..388717cdbd08c441607ec495fbfcac94d66527ad |
| command | sdlc close --issue 162 |
| reviewer | codex |
| timestamp | 2026-08-26T22:38:08-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation itself is well-structured and the full test suite passes, but BR-2 remains unaddressed: the live-Git test executes all four recipes yet permits `stat` and `targeted` to return empty results. A scratch mutation replacing both recipes with empty `base..base` ranges still passed the claimed conformance test. The new manual `--plans-dir` surface is also absent from README.md.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Both tracker directories are normalized relative to the repository, absolute or escaping renderer inputs are rejected, and removing normalization makes the real-Git regression fail.
  - id: BR-2
    disposition: not-addressed
    note: |
      The test executes all recipes, but only names and full assert included paths; empty stat and targeted recipes still pass.
findings:
  - id: new
    severity: Important
    family: readme-cli-surface
    title: |
      README omits the new manual milestone-review flag
    detail: |
      cmd/sdlc/judge.go:64 adds the user-facing --plans-dir flag, but README.md is unchanged. Add the manual milestone-review invocation and its plan-directory behavior, or a direct README entry pointing readers to that documented surface.
```

1. Strengths

- `RenderReviewWindow` and `ReviewCommands` keep manifest construction deterministic and IO-free ([reviewwindow.go:45](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/reviewwindow.go:45)).
- Repository/ref/path resolution stays in a thin integration shell, with both exclusions normalized before manifest construction ([reviewwindow.go:30](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow.go:30)).
- Automatic and manual reviews share the same manifest path while preserving concrete automatic anchors.
- The multi-megabyte regression proves the original patch contains the sentinel and the resulting prompt remains bounded.
- Atlas and generated help/process-manual consumers accurately describe the new transport.

2. Critical findings

None.

3. Important findings

- BR-2 — [reviewwindow_test.go:229](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow_test.go:229): inclusion assertions are conditional on `names` or `full`. Assert meaningful output for `stat` and `targeted` too. A mutant changing those two recipes to `git diff base base` passed.
- README gate — [judge.go:64](/Users/xianxu/workspace/ariadne/cmd/sdlc/judge.go:64): document the new `--plans-dir` user-facing flag in README.md.

4. Minor findings

None.

5. Test coverage notes

- Passed: `go test ./cmd/sdlc/internal/judge -count=1`
- Passed: `go test ./cmd/sdlc/... -count=1`
- Passed: `go test ./... -count=1`
- Passed: pinned-range `git diff --check`
- BR-1 mutation went red as expected.
- BR-2 mutation remained green, demonstrating the coverage gap.

6. Architectural notes for upcoming work

- `ARCH-DRY`: pass—automatic and manual paths share the manifest and recipe builder.
- `ARCH-PURE`: pass—rendering/validation is pure; Git and filesystem work remain at the integration boundary.
- `ARCH-PURPOSE`: flag—conformance claims all four recipes, but only two have their positive semantics pinned.
- `ARCH-MOCK`: flag—the production seam and real Git fixture are appropriate, but live conformance must verify each consumed recipe’s behavior, not merely successful execution.

7. Plan revision recommendations

Append a `## Revisions` entry stating that the previous remediation overclaimed conformance: require positive inclusion and exclusion assertions for each of `stat`, `names`, `full`, and `targeted`, record the empty-range mutant, and add README.md to the user-facing documentation consumers.

---

## Re-review — 2026-08-26T22:47:41-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 162 — sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD) |
| repo | ariadne |
| issue file | workshop/issues/000162-milestone-close-window-base.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a3e7d47247f7c9ca301495416410ca2bb15b01..846b862e2b3da38107b5bec86373ddfcd7e4439e |
| command | sdlc close --issue 162 |
| reviewer | codex |
| timestamp | 2026-08-26T22:47:41-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range implements the compact review manifest correctly, and BR-2 is now backed by meaningful live-Git conformance coverage. BR-3 remains open under the supplied claimed-fix rule: README contains the documentation, but no test asserts it and therefore no test fails when that fix is removed.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      All four rendered recipes execute against real committed and working-tree states; an empty-range stat mutant makes the conformance test fail.
  - id: BR-3
    disposition: not-addressed
    note: |
      README.md documents the flag, but no automated test reads or asserts that documentation, so removing the claimed fix leaves the suite green.
```

1. Strengths

- The manifest renderer is pure and keeps structured argv separate from display quoting ([reviewwindow.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/reviewwindow.go:45)).
- Repository/ref/path validation is confined to a thin integration seam ([reviewwindow.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow.go:30)).
- Absolute tracker overrides are normalized to repository-relative exclusions, preserving BR-1 ([reviewwindow.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow.go:89)).
- The first-milestone test correctly excludes pre-branch issue and unrelated commits ([milestonewindow_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/milestonewindow_test.go:107)).

2. Critical findings

None.

3. Important findings

- BR-3 — [README.md](/Users/xianxu/workspace/ariadne/README.md:15): the documentation itself is correct, but no regression test reads README. Add a test asserting the committed-range form, working-tree form, `--plans-dir`, and its default behavior; demonstrate it fails when this block is removed.

4. Minor findings

None.

5. Test coverage notes

Relevant focused suites and `git diff --check` passed. The BR-2 scratch mutant failed with `stat recipe omitted "tracked.txt"`, confirming the test detects the defective recipe.

The broader `go test ./cmd/sdlc/...` run reached one unrelated sandbox failure because the read-only review environment cannot create `.git/sdlc.lock`; other reported packages passed.

6. Architectural notes for upcoming work

- `ARCH-DRY`: pass — automatic and manual review share one manifest and recipe renderer.
- `ARCH-PURE`: pass — rendering is IO-free; Git and filesystem operations remain in the integration shell.
- `ARCH-PURPOSE`: pass — branch-point windows and bounded repository-based review fulfill the issue’s stated purpose.
- `ARCH-MOCK`: pass — production and tests share the Git-runner seam, with real temporary-repository conformance and a verified failing mutant.

7. Plan revision recommendations

None. The appended revision explicitly supersedes the original `resolveReviewWindow` table entry with `resolveBoundaryReviewManifest`, matching the implementation.

---

## Re-review — 2026-08-26T23:01:57-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 162 — sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD) |
| repo | ariadne |
| issue file | workshop/issues/000162-milestone-close-window-base.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a3e7d47247f7c9ca301495416410ca2bb15b01..62215347cd55c7433dedcef70063afc2e3a31b5b |
| command | sdlc close --issue 162 |
| reviewer | codex |
| timestamp | 2026-08-26T23:01:57-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range implements the compact review manifest well, and all three prior findings are now demonstrably addressed. However, close can still finalize when the newly exposed durable plan changes during the unlocked review. This violates the boundary’s review-integrity contract and requires rework.

### 1. Strengths

- The stat and name-status recipes were run successfully against the exact pinned commits.
- `ReviewWindowManifest` cleanly separates pure rendering from Git/path resolution ([reviewwindow.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/reviewwindow.go:45)).
- Automatic and manual boundary reviews share the same manifest and pinned commit resolution.
- Live-Git conformance now executes every rendered recipe and checks positive and exclusion semantics ([reviewwindow_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow_test.go:143)).
- README and atlas changes document the new manual and architectural surface.

### 2. Critical findings

- [close.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/close.go:1253), [reviewwindow.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/reviewwindow.go:81) — The manifest tells the reviewer to read the optional durable plan, but `closeReviewSnapshot` records only the issue and project files. A scratch integration test paused the reviewer, changed the named plan, then released it; `sdlc close` returned success instead of `boundary review stale`.

  Fix the class: every mutable artifact supplied to an unlocked reviewer must have its presence and contents captured under the lock and revalidated before finalization. Extend the snapshot to cover the canonical plan, including modified, created, deleted, and replaced cases, and add an interleaving regression test that fails without this protection. `ARCH-PURPOSE`.

### 3. Important findings

None.

### 4. Minor findings

None.

### 5. Test coverage notes

- `go test ./... -count=1` passed on the pinned head.
- Focused manifest, boundary-dispatch, README, and renderer tests passed.
- `git diff --check` passed.
- Mutation checks confirmed:

  - BR-1’s test fails when tracker-path normalization is removed.
  - BR-2’s test fails when stat/targeted recipes are changed to empty ranges.
  - BR-3’s test fails when the README section is removed.
  - The new plan-mutation scratch test fails because close incorrectly returns success.

- One earlier overlapping test invocation hit two-second synchronization timeouts; isolated focused and full reruns passed.

### 6. Architectural notes for upcoming work

- `ARCH-DRY`: Pass. Automatic/manual reviews share the manifest renderer and window logic.
- `ARCH-PURE`: Pass. Rendering and validation remain pure; Git/filesystem operations stay in the boundary layer.
- `ARCH-PURPOSE`: Flag. The “pinned review input” purpose is incomplete while the named plan can change without invalidating finalization.
- `ARCH-MOCK`: Pass. Git uses the production runner seam, a stateful fake, and executable live-repository conformance.
- Every Core concepts entity exists at its stated path and matches its PURE/INTEGRATION classification.

### 7. Plan revision recommendations

Append a `## Revisions` entry recording:

- The invariant that every mutable artifact named to the unlocked reviewer must be snapshotted.
- The addition of canonical-plan presence/content to `closeReviewSnapshot`.
- Interleaving coverage for plan modification, creation, deletion, and replacement during review.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Removing tracker-directory normalization makes the live-Git absolute-override regression test fail.
  - id: BR-2
    disposition: addressed
    note: |
      Mutating stat and targeted recipes to empty ranges makes their positive-semantic assertions fail.
  - id: BR-3
    disposition: addressed
    note: |
      Removing the README manual boundary-review section makes the scoped repository contract test fail.
findings:
  - id: new
    severity: Critical
    family: unlocked-review-snapshot-completeness
    title: |
      Durable plan changes do not invalidate an in-flight boundary review
    detail: |
      The manifest names the optional plan as reviewer input, but closeReviewSnapshot captures only issue and project text. A blocked-review integration reproduction changed the plan and close still finalized; snapshot and revalidate every mutable reviewer input, including plan presence and contents.
```

---

## Re-review — 2026-08-26T23:18:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 162 — sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD) |
| repo | ariadne |
| issue file | workshop/issues/000162-milestone-close-window-base.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a3e7d47247f7c9ca301495416410ca2bb15b01..f3c0721f71261a8c9bbfca25332323c48fdff021 |
| command | sdlc close --issue 162 |
| reviewer | codex |
| timestamp | 2026-08-26T23:18:31-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills the issue’s Spec and effective Plan. Review manifests are bounded and immutable, first-milestone windows use the branch point, atlas and review gates share the same base, and BR-4 is fully addressed with deletion-sensitive coverage.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      Both close paths snapshot canonical-plan presence and contents; disabling plan capture makes all eight mutation cases fail.
```

## 1. Strengths

- `boundaryWindowBase` provides one window source for atlas and review gates and preserves prior-boundary precedence (`cmd/sdlc/milestoneclose.go:289`).
- Manifest rendering is pure and uses structured argv with display-only shell quoting (`cmd/sdlc/internal/judge/reviewwindow.go:43`).
- Automatic and manual boundary reviews share `resolveBoundaryReviewManifest` (`cmd/sdlc/reviewwindow.go:30`).
- BR-4’s shared artifact snapshot covers issue, project, and canonical-plan presence/content (`cmd/sdlc/close.go:1275`).

## 2. Critical findings

None.

## 3. Important findings

None.

## 4. Minor findings

None.

## 5. Test coverage notes

- `go test ./... -count=1` passed.
- Focused review-window, boundary-base, dispatch, manual-judge, and BR-4 tests passed.
- BR-4 mutation proof: disabling plan capture caused all eight close/milestone × modify/create/delete/replace cases to fail.
- `git diff --check` passed.
- Tests exercise all four rendered Git recipes against a real temporary repository, alongside the stateful Git fake.

## 6. Architectural notes for upcoming work

- `ARCH-DRY`: Pass — shared manifest and boundary-base helpers avoid parallel implementations.
- `ARCH-PURE`: Pass — validation/rendering remains deterministic and IO-free; Git and filesystem work stays in the integration shell.
- `ARCH-PURPOSE`: Pass — the complete windowing and unlocked-review snapshot contracts are delivered, not only the originally observed instances.
- `ARCH-MOCK`: Pass — production and tests share the Git-runner seam, with a stateful fake and live-repository conformance.

The Core concepts cross-check passes after applying the plan’s appended revision, which supersedes the original `resolveReviewWindow` row with `resolveBoundaryReviewManifest`.

## 7. Plan revision recommendations

None.
