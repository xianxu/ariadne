# Portable root Makefile seed implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute using superpowers-executing-plans; the SDLC gates own fresh-context review.

**Goal:** A consumer keeps working product targets in a standalone checkout and regains maintainer tooling through bootstrap without following a seed destination symlink into its ancestor.

**Architecture:** Keep root Makefile and CI upstream-owned seeds, product behavior in Makefile.local and executable scripts/ci-setup.sh. Resolve the optional workflow from local overlay or sibling ariadne; resolve pre-weave helpers from that workflow's source. Reuse the existing filesystem seam and regular-file materialization pattern.

**Tech Stack:** Go weave planner, GNU Make, Bash, GitHub Actions.

## Approval and scope

The operator approved the parley.nvim#208 plan including this prerequisite on 2026-09-13. This refines that approved narrow prerequisite; no ownership override API, real consumer propagation, or local restoration workaround is introduced.

## Core concepts

| Name | Lives in | Status | Purpose |
|---|---|---|---|
| applySeed | cmd/weave/internal/plan/apply.go | modified | Materialize upstream bytes without following destination links |
| removeDestinationSymlink | cmd/weave/internal/plan/apply.go | new integration | Shared Lstat/remove guard for seed and composed writes |
| workflow include selection | Makefile | modified integration | Prefer local overlay, otherwise sibling ariadne, otherwise no overlay |
| portable bootstrap chain | Makefile.workflow | modified integration | Resolve pre-weave helpers and order peer setup before weave before downstream tooling |
| CI setup and runner selection | .github/workflows/merge-check.yml | modified integration | Run optional consumer setup and locate generic runner after cloning |

No new domain noun. Filesystem seam remains weavefs.FS; tests use OSFS rooted at t.TempDir and fault-injecting wrapper over that real backing folder. Make/Bash integration uses a portable scratch sibling tree and real subprocesses, with fixture tools recording ordered calls and creating materialized helper state. Real weave compilation checks the fixture contract each change; no production credentials or HOME touched.

## Decisions and architecture

- ARCH-DRY: share symlink unlink guard with applyWriteFile; keep one root template and one CI workflow. Local help extensibility stays ordinary Make targets.
- ARCH-PURE: no new policy layer; workflow wildcard discovery reads the filesystem and is an integration seam; filesystem mutation remains in FS-injected apply functions.
- ARCH-PURPOSE: cover linked and already-seeded consumers, absent helper links, product targets, effective CI after repeat weave; not merely optional include parsing.
- ARCH-MOCK: fault FS retains real files; bootstrap fixture tools persist call log/materialized helpers. Conformance uses real make/git/weave in scratch; no network cloning is needed because the upstream scratch peer already exists.
- ARCH-CONSTRAINTS: operator-triggered batch setup, one synchronous process chain; no UI budget. Existing peer-depth limits remain. Tests bounded to temporary trees; standard GNU Make on macOS/Linux. Source sizes remain manifest-controlled small text files; no new cache or scan.
- ARCH-SECURE: repository manifests and executable hooks are trusted repository code, same authority as Makefile.local. Missing source retains historical nonfatal skip before destination mutation. Lstat errors other than absence fail closed, unlink failure prevents write/chmod. Seed source read completes before unlink; no target bytes or permissions change. Concurrent hostile filesystem mutation is outside the existing single-writer weave contract.
- ARCH-ORDER: absent destination → write; regular identical → no rewrite; regular drifted → rewrite; symlink (matching/differing/dangling) → unlink then write. Retry after failed write restores absent destination. Bootstrap peers completes before weave, which completes before tooling/install; failed phase prevents later phases. No background processes survive return.
- ARCH-FUNERAL: seeded files replace fixed manifest slots, removed by repository owner when retired; no per-run growth. Fixture folders use test cleanup/traps, tool binaries remain existing owner bin lifecycle.

## Tasks

### 1. Safe seed migration

Files: cmd/weave/internal/plan/apply.go, apply_test.go; construct/base.manifest.

- [x] TestApplySeedDestinationStates and TestMaterializationFailures exercise applySeed, applyWriteFile, and removeDestinationSymlink through filesystem-state transitions and injected failures; assert ancestor preservation and retry convergence. Run `go test ./cmd/weave/internal/plan -run 'Seed|WriteFile' -count=1`; require expected red before editing implementation.
- [x] Share fail-closed destination-link removal between composed writes and seeds; bypass identical-content early path for links, preserve existing executable mode behavior. Change manifest root Makefile from symlink to seed. Rerun tests green and commit.

### 2. Portable product and bootstrap boundary

Files: Makefile, Makefile.workflow, bootstrap.sh comments; construct/scripts/test/portable-makefile.test.sh.

- [x] portable-makefile.test.sh exercises the actual Makefile help/product and bootstrap entry points against a stateful scratch sibling tree; assert product independence, overlay precedence, ordered materialization and failure propagation. Confirm red.
- [x] Root chooses optional local/sibling workflow and conditional available help targets. Workflow uses source fallback for dev-aliases and bootstrap-peers before weave. Use a private bootstrap composition target that invokes recursive make phases sequentially; keep public bootstrap prerequisite-only so consumer extensions remain additive. Materialized helpers service remaining existing targets after weave. Update stale bootstrap comments. Run focused shell tests, ensure-go and bootstrap-transitive/peer-update regressions.
- [x] Run actual scratch parley-like leaf weave twice with real compiled weave and portable copied upstream. Assert root regular file, ancestor hash/mode stable, local product target works, no content churn. All generated output stays scratch. Commit.

### 3. Portable generic CI and documentation

Files: .github/workflows/merge-check.yml; scripts/test/portable-ci.test.sh; atlas/workflow/base-layer.md and relevant atlas CI page/index; README.md; workshop/lessons.md.

- [x] portable-ci.test.sh executes extracted actual workflow run blocks against a stateful scratch repository; assert runner precedence, effective setup before checks and failure propagation through seed refresh. Confirm red.
- [x] Invoke optional executable scripts/ci-setup.sh after peer clone and before checks; select local runner or ../ariadne/scripts/run-merge-checks.sh, fail clearly if neither. Make Go setup use a conditional go.mod lookup compatible with products lacking a root Go module (clone first, select local go.mod else bootstrapped ariadne/go.mod). Run workflow regression green.
- [x] Update source ownership/bootstrap/CI atlas and lessons with verified behavior. Run `go test ./cmd/weave/... -count=1`, touched shell suites, `git diff --check`; tick plan, commit, run sdlc close with measured evidence then sdlc pr/merge. Fix any mandatory review findings before shipping.

## Revisions

### 2026-09-13 — executable prerequisite refinement

Reason: approved #208 requires absence of all ignored helper links. Delta: explicit pre-weave helper resolution, ordered bootstrap phases, and CI Go module selection avoid depending on the very links the first weave must create.

### 2026-09-13 — PQ-1 test strategy contract

Reason: plan review requested named test surfaces rather than scenario inventories. Delta: task verification now identifies production entry points, stateful seams and invariant oracles; acceptance criteria above remain unchanged.

### 2026-09-13 — implementation conformance

Reason: full real bootstrap revealed an implicit Make rule. Delta: public
bootstrap is explicitly phony, covered with the real bootstrap.sh in the fixture.
Post-weave tools finish before installation to avoid overlapping binary builds.
Verification recorded in issue Log; close/ship remain SDLC gate actions.

### 2026-09-13 — BR-1/BR-2 contract sweep

Reason: boundary review found a classification contradiction and missing public
setup documentation. Delta: workflow/helper wildcard discovery is explicitly
INTEGRATION (not pure); inspected every concept row for the same mistake.
README now covers seeded ownership, product overrides, bootstrap, and CI hook
ordering/failure behavior alongside both atlas pages. No runtime changes.
